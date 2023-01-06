pub mod utils;

#[cfg(test)]
mod tests {

    use casper_engine_test_support::ExecuteRequestBuilder;
    use casper_execution_engine::core::{engine_state, execution};
    use casper_types::{runtime_args, RuntimeArgs, U256};
    use casper_types::{ApiError, Key};
    use contract_bridge::error::BridgeError;
    //use casper_types::CLType::Key;
    use crate::utils::{
        deploy_bridge, deploy_erc20, fill_purse_on_token_contract, query_balance,
        read_contract_event, setup_context, simple_deploy_builder, UserAccount, ACCOUNT_BALANCE,
    };
    use casper_common::event::BridgeEvent;
    use contract_bridge::entry_points::{
        EP_BRIDGE_IN, EP_BRIDGE_IN_CONFIRM, EP_BRIDGE_OUT, EP_TRANSFER_OUT, PARAM_AMOUNT,
        PARAM_DESTINATION_ADDRESS, PARAM_DESTINATION_CHAIN, PARAM_RECIPIENT, PARAM_SENDER,
        PARAM_SOURCE_ADDRESS, PARAM_SOURCE_CHAIN, PARAM_TOKEN_CONTRACT,
    };

    const ERC20_INSUFFIENT_BALANCE_ERROR_CODE: u16 = u16::MAX - 1; // https://github.com/casper-ecosystem/erc20/blob/master/erc20/src/error.rs

    #[test]
    fn test_deploy_erc20() {
        let mut context = setup_context();

        deploy_erc20(&mut context.builder, context.account.address);
    }

    #[test]
    fn test_deploy_bridge() {
        let mut context = setup_context();

        deploy_bridge(&mut context.builder, context.account.address);
    }

    #[test]
    fn verify_bridge_entry_poitns() {
        let mut context = setup_context();

        let (bridge_address, _) = deploy_bridge(&mut context.builder, context.account.address);

        let contract = context.builder.get_contract(bridge_address).unwrap();
        let expected_entries = vec![
            EP_BRIDGE_IN,
            EP_BRIDGE_IN_CONFIRM,
            EP_BRIDGE_OUT,
            EP_TRANSFER_OUT,
        ];

        let mut count = 0;
        for entry in contract.entry_points().keys() {
            assert!(expected_entries.contains(&entry.as_str()), "You have introduced a new entry point please add it to the expected list and cover with a tests");
            count += 1;
        }
        assert_eq!(count, expected_entries.len());
    }

    #[test]
    fn bridge_in_happy_path() {
        /*
            Scenario:

            1. Call "bridge_in" entrypoint in bridge contract with the specified token
            2. Assert that bridge contract received the expected amount of tokens
            3. Verify expected event
        */

        let mut context = setup_context();

        let (bridge_hash, bridge_package_hash) =
            deploy_bridge(&mut context.builder, context.account.address);
        let (token_hash, token_package_hash) =
            deploy_erc20(&mut context.builder, context.account.address);

        let deploy_item = simple_deploy_builder(context.account.address)
            .with_stored_session_hash(
                bridge_hash,
                EP_BRIDGE_IN,
                runtime_args! {
                    PARAM_TOKEN_CONTRACT => token_package_hash,
                    PARAM_AMOUNT => U256::one() * 1_000_000_000_000u64,
                    PARAM_DESTINATION_CHAIN => "DEST".to_string(),
                    PARAM_DESTINATION_ADDRESS => "DESTADDR".to_string(),
                },
            )
            .build();

        context
            .builder
            .exec(ExecuteRequestBuilder::from_deploy_item(deploy_item).build())
            .commit()
            .expect_success();

        let funds_in_event = read_contract_event::<_, BridgeEvent>(
            &mut context.builder,
            bridge_hash,
            "event_trigger",
        );

        if let BridgeEvent::FundsIn {
            token_contract,
            destination_chain,
            destination_address,
            amount,
            sender,
        } = funds_in_event
        {
            assert_eq!(token_contract, token_package_hash);
            assert_eq!(destination_chain, "DEST");
            assert_eq!(destination_address, "DESTADDR");
            assert_eq!(amount, U256::one() * 1_000_000_000_000u64);
            assert_eq!(sender, Key::Account(context.account.address));
        } else {
            panic!("wrong bridge event kind");
        }

        let bridge_balance = query_balance(
            &mut context.builder,
            token_hash,
            &Key::Hash(bridge_package_hash.value()),
        );

        assert_eq!(bridge_balance, U256::one() * 1_000_000_000_000u64);
    }

    #[test]
    fn bridge_out_happy_path() {
        /*
            Scenario:

            1. Transfer tokens to bridge contract via ERC-20's "transfer" entrypoint
            2. Assert that bridge contract received the expected amount of tokens
            3. Call "bridge_out" entrypoint in bridge contract
            4. Assert that the contract lost the expected amount of tokens
            5. Assert that the recipient received the expected amount of tokens
            6. Verify all expected events have been emitted
        */

        let mut context = setup_context();

        // Creating recipient account
        let recipient = UserAccount::unique_account(&mut context, 0);
        let recipient_key = recipient.key();

        let (bridge_hash, bridge_package_hash) =
            deploy_bridge(&mut context.builder, context.account.address);
        let (token_hash, token_package_hash) =
            deploy_erc20(&mut context.builder, context.account.address);

        // Transfer tokens to bridge contract via ERC-20's "transfer" entrypoint (1, 2)
        fill_purse_on_token_contract(
            &mut context,
            token_hash,
            U256::one() * 1_000_000_000_000u64,
            Key::from(bridge_package_hash),
        );

        // Call "bridge_out" entrypoint in bridge contract (3)
        let deploy_item = simple_deploy_builder(context.account.address)
            .with_stored_session_hash(
                bridge_hash,
                EP_BRIDGE_OUT,
                runtime_args! {
                    PARAM_TOKEN_CONTRACT => token_package_hash,
                    PARAM_AMOUNT => U256::one() * 900_000_000_000u64,
                    PARAM_SOURCE_CHAIN => "SOUR".to_string(),
                    PARAM_SOURCE_ADDRESS => "SOURADDR".to_string(),
                    PARAM_RECIPIENT => recipient_key
                },
            )
            .build();

        context
            .builder
            .exec(ExecuteRequestBuilder::from_deploy_item(deploy_item).build())
            .commit()
            .expect_success();

        let bridge_balance = query_balance(
            &mut context.builder,
            token_hash,
            &Key::from(bridge_package_hash),
        );
        assert_eq!(bridge_balance, U256::one() * 100_000_000_000u64);
        let recipient_balance = query_balance(&mut context.builder, token_hash, &recipient_key);
        assert_eq!(recipient_balance, U256::one() * 900_000_000_000u64);

        let bridge_out_event = read_contract_event::<_, BridgeEvent>(
            &mut context.builder,
            bridge_hash,
            "event_trigger",
        );
        if let BridgeEvent::FundsOut {
            token_contract,
            source_chain,
            source_address,
            recipient,
            amount,
        } = bridge_out_event
        {
            assert_eq!(token_contract, token_package_hash);
            assert_eq!(source_chain, "SOUR");
            assert_eq!(source_address, "SOURADDR");
            assert_eq!(recipient, recipient_key);
            assert_eq!(amount, U256::one() * 900_000_000_000u64);
        } else {
            panic!("Expected bridge out event, but got {:?}", bridge_out_event);
        }
    }

    #[test]
    fn bridge_in_out_happy_path() {
        /*
            Scenario:

            1. Call "bridge_in" entrypoint in bridge contract with the specified token
            2. Assert that bridge contract received the expected amount of tokens
            3. Call "bridge_out" entrypoint in bridge contract
            4. Assert that the contract lost the expected amount of tokens
            5. Assert that the recipient received the expected amount of tokens
            6. Verify all expected events have been emitted
        */
        let mut context = setup_context();

        let (bridge_hash, bridge_package_hash) =
            deploy_bridge(&mut context.builder, context.account.address);
        let (token_hash, token_package_hash) =
            deploy_erc20(&mut context.builder, context.account.address);

        let deploy_item = simple_deploy_builder(context.account.address)
            .with_stored_session_hash(
                bridge_hash,
                EP_BRIDGE_IN,
                runtime_args! {
                    PARAM_TOKEN_CONTRACT => token_package_hash,
                    PARAM_AMOUNT => U256::one() * 1_000_000_000_000u64,
                    PARAM_DESTINATION_CHAIN => "DEST".to_string(),
                    PARAM_DESTINATION_ADDRESS => "DESTADDR".to_string(),
                },
            )
            .build();

        context
            .builder
            .exec(ExecuteRequestBuilder::from_deploy_item(deploy_item).build())
            .commit()
            .expect_success();

        let in_event = read_contract_event::<_, BridgeEvent>(
            &mut context.builder,
            bridge_hash,
            "event_trigger",
        );

        let bridge_balance = query_balance(
            &mut context.builder,
            token_hash,
            &Key::Hash(bridge_package_hash.value()),
        );
        assert_eq!(bridge_balance, U256::one() * 1_000_000_000_000u64);
        let recipient = UserAccount::unique_account(&mut context, 0);
        let recipient_key = recipient.key();

        let deploy_item = simple_deploy_builder(context.account.address)
            .with_stored_session_hash(
                bridge_hash,
                EP_BRIDGE_OUT,
                runtime_args! {
                    PARAM_TOKEN_CONTRACT => token_package_hash,
                    PARAM_AMOUNT => U256::one() * 500_000_000_000u64,
                    PARAM_SOURCE_CHAIN => "SOUR".to_string(),
                    PARAM_SOURCE_ADDRESS => "SOURADDR".to_string(),
                    PARAM_RECIPIENT => recipient_key,
                },
            )
            .build();

        context
            .builder
            .exec(ExecuteRequestBuilder::from_deploy_item(deploy_item).build())
            .commit()
            .expect_success();

        let bridge_balance = query_balance(
            &mut context.builder,
            token_hash,
            &Key::Hash(bridge_package_hash.value()),
        );
        assert_eq!(bridge_balance, U256::one() * 500_000_000_000u64);

        let recipient_balance = query_balance(&mut context.builder, token_hash, &recipient_key);
        assert_eq!(recipient_balance, U256::one() * 500_000_000_000u64);

        // Verify all expected events have been emitted.
        let out_event = read_contract_event::<_, BridgeEvent>(
            &mut context.builder,
            bridge_hash,
            "event_trigger",
        );

        if let BridgeEvent::FundsIn {
            token_contract,
            destination_chain,
            destination_address,
            amount,
            sender,
        } = in_event
        {
            assert_eq!(token_contract, token_package_hash);
            assert_eq!(destination_chain, "DEST");
            assert_eq!(destination_address, "DESTADDR");
            assert_eq!(amount, U256::one() * 1_000_000_000_000u64);
            assert_eq!(sender, Key::Account(context.account.address));
        } else {
            panic!("Expected BridgeEvent::FundsIn but got {:?}", in_event);
        }

        if let BridgeEvent::FundsOut {
            token_contract,
            source_chain,
            source_address,
            amount,
            recipient,
        } = out_event
        {
            assert_eq!(token_contract, token_package_hash);
            assert_eq!(source_chain, "SOUR");
            assert_eq!(source_address, "SOURADDR");
            assert_eq!(amount, U256::one() * 500_000_000_000u64);
            assert_eq!(recipient, recipient_key);
        } else {
            panic!("Expected BridgeEvent::FundsOut but got {:?}", out_event);
        }
    }

    #[test]
    fn bridge_in_insufficient_tokens() {
        /*
            Scenario:
            1. Call "bridge_in" entrypoint in bridge contract with the specified token with an amount that exceeds current balance
            2. Assert that the call failed
        */

        let mut context = setup_context();

        // Create an accoutn for the test
        let sender = UserAccount::unique_account(&mut context, 0);

        // Deploy the bridge contract and the token contract
        let (bridge_hash, bridge_package_hash) =
            deploy_bridge(&mut context.builder, context.account.address);
        let (token_hash, token_package_hash) =
            deploy_erc20(&mut context.builder, context.account.address);

        // Create a purse to hold the tokens for the bridge contract to verify that the amount didn't change
        fill_purse_on_token_contract(
            &mut context,
            token_hash,
            U256::one(),
            Key::from(bridge_package_hash),
        );

        fill_purse_on_token_contract(&mut context, token_hash, U256::one() * 1000, sender.key());

        // Try to transfer token in bridge from account that doesn't have enough tokens
        let deploy_item = simple_deploy_builder(sender.address)
            .with_stored_session_hash(
                bridge_hash,
                EP_BRIDGE_IN,
                runtime_args! {
                    PARAM_TOKEN_CONTRACT => token_package_hash,
                    PARAM_AMOUNT => U256::one() * ACCOUNT_BALANCE * 2,
                    PARAM_DESTINATION_CHAIN => "DEST".to_string(),
                    PARAM_DESTINATION_ADDRESS => "DESTADDR".to_string(),
                },
            )
            .build();

        // Verify that the transaction fails
        let error = context
            .builder
            .exec(ExecuteRequestBuilder::from_deploy_item(deploy_item).build())
            .commit()
            .expect_failure()
            .get_error()
            .unwrap();

        let balance = query_balance(
            &mut context.builder,
            token_hash,
            &Key::Hash(bridge_package_hash.value()),
        );

        // The balance didn't change
        assert_eq!(balance, U256::one());

        // Verifies that error is expected one.
        let expected_error = engine_state::Error::Exec(execution::Error::Revert(ApiError::User(
            ERC20_INSUFFIENT_BALANCE_ERROR_CODE,
        )));
        assert!(
            matches!(
                error.clone(),
                engine_state::Error::Exec(execution::Error::Revert(ApiError::User(
                    ERC20_INSUFFIENT_BALANCE_ERROR_CODE,
                )))
            ),
            "Unexpected error message. Expected: {}, but got {}",
            expected_error,
            error
        );
    }

    #[test]
    fn bridge_out_insufficient_tokens() {
        /*
            Scenario:

            1. Transfer tokens to bridge contract via ERC-20's "transfer" entrypoint
            2. Call "bridge_out" entrypoint in bridge contract with the specified token with an amount that exceeds the bridge's current balance
            3. Assert that the call failed
        */

        let mut context = setup_context();

        // Creating accounts for the test.
        let recipient = UserAccount::unique_account(&mut context, 0);
        let recipient_key = recipient.key();

        // Deploy bridge and token contracts.
        let (bridge_hash, bridge_package_hash) =
            deploy_bridge(&mut context.builder, context.account.address);
        let (token_hash, token_package_hash) =
            deploy_erc20(&mut context.builder, context.account.address);

        // Transferings token to bridge token purse
        fill_purse_on_token_contract(
            &mut context,
            token_hash,
            U256::one() * 3_000_000_000_000u64,
            Key::from(bridge_package_hash),
        );

        // Transfer tokens from bridge to recipient that exceeds the bridge's current balance.
        let deploy_item = simple_deploy_builder(context.account.address)
            .with_stored_session_hash(
                bridge_hash,
                EP_BRIDGE_OUT,
                runtime_args! {
                    PARAM_TOKEN_CONTRACT => token_package_hash,
                    PARAM_AMOUNT => U256::one() * 5_000_000_000_000u64,
                    PARAM_SOURCE_CHAIN => "SOUR".to_string(),
                    PARAM_SOURCE_ADDRESS => "SOURADDR".to_string(),
                    PARAM_RECIPIENT => recipient_key,
                },
            )
            .build();

        // Verify that transaction fails.
        let error = context
            .builder
            .exec(ExecuteRequestBuilder::from_deploy_item(deploy_item).build())
            .commit()
            .expect_failure()
            .get_error()
            .unwrap();

        let bridge_balance = query_balance(
            &mut context.builder,
            token_hash,
            &Key::Hash(bridge_package_hash.value()),
        );

        // Verify that the balance of the bridge is still 3_000_000_000_000 tokens.
        assert_eq!(bridge_balance, U256::one() * 3_000_000_000_000u64);
        let expected_error = engine_state::Error::Exec(execution::Error::Revert(ApiError::User(
            ERC20_INSUFFIENT_BALANCE_ERROR_CODE,
        )));

        // Verify that the error message is the expected one.
        assert!(
            matches!(
                error.clone(),
                engine_state::Error::Exec(execution::Error::Revert(ApiError::User(
                    ERC20_INSUFFIENT_BALANCE_ERROR_CODE,
                )))
            ),
            "Unexpected error message. Expected: {}, but got {}",
            expected_error,
            error
        );
    }

    #[test]
    fn bridge_out_called_by_non_owner() {
        /*
            Scenario:
            1. Transfer tokens to bridge contract via ERC-20's "transfer" entrypoint
            2. Call bridge out entrypoint from another account
        */

        let mut context = setup_context();

        // Creating account for the test.
        let recipient = UserAccount::unique_account(&mut context, 0);
        let recipient_key = recipient.key();

        // Deploy bridge and token contracts.
        let (bridge_hash, bridge_package_hash) =
            deploy_bridge(&mut context.builder, context.account.address);
        let (token_hash, token_package_hash) =
            deploy_erc20(&mut context.builder, context.account.address);

        fill_purse_on_token_contract(
            &mut context,
            token_hash,
            U256::one() * 1_000_000_000_000u64,
            Key::from(bridge_package_hash),
        );

        // Transfer tokens from bridge to recipient called by non-owner.
        let deploy_item = simple_deploy_builder(recipient.address)
            .with_stored_session_hash(
                bridge_hash,
                EP_BRIDGE_OUT,
                runtime_args! {
                    PARAM_TOKEN_CONTRACT => token_package_hash,
                    PARAM_AMOUNT => U256::one() * 1_000_000_000_000u64,
                    PARAM_SOURCE_CHAIN => "SOUR".to_string(),
                    PARAM_SOURCE_ADDRESS => "SOURADDR".to_string(),
                    PARAM_RECIPIENT => recipient_key,
                },
            )
            .build();

        // Verify that transaction fails.
        let error = context
            .builder
            .exec(ExecuteRequestBuilder::from_deploy_item(deploy_item).build())
            .commit()
            .expect_failure()
            .get_error()
            .unwrap();

        let balance = query_balance(
            &mut context.builder,
            token_hash,
            &Key::from(bridge_package_hash),
        );
        assert_eq!(balance, U256::one() * 1_000_000_000_000u64);

        let expected_error = engine_state::Error::Exec(execution::Error::InvalidContext);
        assert_eq!(error.to_string(), expected_error.to_string());
    }

    #[test]
    fn bridge_in_confirn_not_public_available() {
        /* Call to bridge in confign shouldn't be available except from contract context */
        let mut context = setup_context();

        // Creating account for the test.
        let test_subj = UserAccount::unique_account(&mut context, 0);

        // Deploy bridge.
        let (bridge_hash, _) = deploy_bridge(&mut context.builder, context.account.address);

        // Deploy token contract
        let (_, token_package_hash) = deploy_erc20(&mut context.builder, context.account.address);
        for user in vec![&context.account, &test_subj] {
            let args = runtime_args! {
                PARAM_TOKEN_CONTRACT => token_package_hash,
                PARAM_AMOUNT => U256::one() * 1_000_000_000_000u64,
                PARAM_DESTINATION_ADDRESS => "DESTADDR".to_string(),
                PARAM_DESTINATION_CHAIN => "DEST".to_string(),
                PARAM_SENDER => user.key(),
            };

            let deploy = simple_deploy_builder(user.address)
                .with_stored_session_hash(bridge_hash, EP_BRIDGE_IN_CONFIRM, args.clone())
                .build();
            // Verify that transaction fails.
            let error = context
                .builder
                .exec(ExecuteRequestBuilder::from_deploy_item(deploy).build())
                .commit()
                .expect_failure()
                .get_error()
                .unwrap();

            let api_error: ApiError = contract_util::error::Error::<BridgeError>::Contract(
                BridgeError::OnlyCallableBySelf,
            )
            .into();
            let expected_error = engine_state::Error::Exec(execution::Error::Revert(api_error));
            assert_eq!(error.to_string(), expected_error.to_string());
        }
    }

    #[test]
    fn token_overflow() {
        /*
           Scenario:
           1. Mint tokens to bridge contract via ERC-20's "mint" entrypoint
           2. Transfer tokens to bridge contract via ERC-20's "transfer" entrypoint
           3. Verify that transaction fails because of overflow
        */

        let mut context = setup_context();

        // Deploy bridge and token contracts.
        let (bridge_hash, bridge_package_hash) =
            deploy_bridge(&mut context.builder, context.account.address);
        let (token_hash, token_package_hash) =
            deploy_erc20(&mut context.builder, context.account.address);

        // Transfer all tokens to bridge.
        fill_purse_on_token_contract(
            &mut context,
            token_hash,
            U256::max_value(),
            Key::from(bridge_package_hash),
        );

        let balance_before = query_balance(
            &mut context.builder,
            token_hash,
            &Key::from(bridge_package_hash),
        );

        assert_eq!(balance_before, U256::max_value());

        // Mint tokens to owner to send it more to bridge and overflow it.
        let deploy_item = simple_deploy_builder(context.account.address)
            .with_stored_session_hash(
                token_hash,
                "mint",
                runtime_args! {
                    PARAM_AMOUNT => U256::one() * 1_000_000_000_000u64,
                    PARAM_RECIPIENT => Key::from(context.account.address),
                },
            )
            .build();

        context
            .builder
            .exec(ExecuteRequestBuilder::from_deploy_item(deploy_item).build())
            .commit()
            .expect_success();

        let owner_balance = query_balance(
            &mut context.builder,
            token_hash,
            &Key::from(context.account.address),
        );
        assert_eq!(owner_balance, U256::one() * 1_000_000_000_000u64);

        // Transfer more tokens inside bridge and it supposed to overflow.
        let deploy_item = simple_deploy_builder(context.account.address)
            .with_stored_session_hash(
                bridge_hash,
                EP_BRIDGE_IN,
                runtime_args! {
                    PARAM_TOKEN_CONTRACT => token_package_hash,
                    PARAM_AMOUNT => U256::one() * 1000u64,
                    PARAM_DESTINATION_CHAIN => "DEST".to_string(),
                    PARAM_DESTINATION_ADDRESS => "DESTADDR".to_string(),
                },
            )
            .build();

        // Verify that transaction fails.
        let error = context
            .builder
            .exec(ExecuteRequestBuilder::from_deploy_item(deploy_item).build())
            .commit()
            .expect_failure()
            .get_error()
            .unwrap();

        let api_error = contract_util::error::Error::<BridgeError>::Contract(
            BridgeError::UnexpectedTransferAmount,
        );
        let expected_error = engine_state::Error::Exec(execution::Error::Revert(api_error.into()));

        assert_eq!(error.to_string(), expected_error.to_string());
    }

    #[test]
    fn transfer_out_funds() {
        /*
           Scenario:
           * Transfer tokens for bridge contract via ERC-20's "transfer" entrypoint
           * Verify that tokens are transferred to bridge contract
           * Transfer out to some account and verify that tokens are transferred to this account
        */

        let mut context = setup_context();

        let recipient = UserAccount::unique_account(&mut context, 0);

        // Deploy bridge and token contracts.
        let (bridge_hash, bridge_package_hash) =
            deploy_bridge(&mut context.builder, context.account.address);
        let (token_hash, token_package_hash) =
            deploy_erc20(&mut context.builder, context.account.address);

        fill_purse_on_token_contract(
            &mut context,
            token_hash,
            U256::one() * 1_000_000_000_000u64,
            Key::from(bridge_package_hash),
        );

        // Transfer out tokens from bridge (For example someone requested revert)
        let deploy_item = simple_deploy_builder(context.account.address)
            .with_stored_session_hash(
                bridge_hash,
                EP_TRANSFER_OUT,
                runtime_args! {
                    PARAM_TOKEN_CONTRACT => token_package_hash,
                    PARAM_AMOUNT => U256::one() * 1_000_000_000_000u64,
                    PARAM_RECIPIENT => recipient.key(),
                },
            )
            .build();

        context
            .builder
            .exec(ExecuteRequestBuilder::from_deploy_item(deploy_item).build())
            .commit()
            .expect_success();

        let balance_after = query_balance(
            &mut context.builder,
            token_hash,
            &Key::from(bridge_package_hash),
        );

        assert_eq!(balance_after, U256::zero());

        let recipient_balance = query_balance(
            &mut context.builder,
            token_hash,
            &Key::from(recipient.key()),
        );

        assert_eq!(recipient_balance, U256::one() * 1_000_000_000_000u64);
    }

    #[test]
    fn transfer_out_called_by_non_owner() {
        /*
         * Transfer tokens to the bridge contract via ERC-20's "transfer" entrypoint
         * Create a user account and use transfer_out entrypoint from the bridge contract with this account as a sender
         * Verify that transaction fails because of the sender is not the owner of the bridge contract
         * Verify that balance of the bridge contract is not changed
         */

        let mut context = setup_context();
        let user = UserAccount::unique_account(&mut context, 0);

        // Deploy bridge and token contracts.
        let (bridge_hash, bridge_package_hash) =
            deploy_bridge(&mut context.builder, context.account.address);
        let (token_hash, token_package_hash) =
            deploy_erc20(&mut context.builder, context.account.address);

        fill_purse_on_token_contract(
            &mut context,
            token_hash,
            U256::one() * 1_000_000_000_000u64,
            Key::from(bridge_package_hash),
        );

        // Transfer out tokens from bridge
        let deploy_item = simple_deploy_builder(user.address)
            .with_stored_session_hash(
                bridge_hash,
                EP_TRANSFER_OUT,
                runtime_args! {
                    PARAM_TOKEN_CONTRACT => token_package_hash,
                    PARAM_AMOUNT => U256::one() * 1_000_000_000_000u64,
                    PARAM_RECIPIENT => user.key(),
                },
            )
            .build();
        let error = context
            .builder
            .exec(ExecuteRequestBuilder::from_deploy_item(deploy_item).build())
            .commit()
            .expect_failure()
            .get_error()
            .unwrap();

        let expected_error = engine_state::Error::Exec(execution::Error::InvalidContext);
        assert_eq!(error.to_string(), expected_error.to_string());
    }

    #[test]
    fn transfer_out_insuffient_balance() {
        /*
           Scenario:
           * Transfer tokens for bridge contract via ERC-20's "transfer" entrypoint
           * Try to use transfer out entrypoint from the bridge contract with request that will be higher than the balance of the bridge contract
           * Verify that transaction fails because of the balance of the bridge contract is not enough to transfer out
        */

        let mut context = setup_context();

        // Deploy bridge and token contracts.
        let (bridge_hash, bridge_package_hash) =
            deploy_bridge(&mut context.builder, context.account.address);
        let (token_hash, token_package_hash) =
            deploy_erc20(&mut context.builder, context.account.address);

        fill_purse_on_token_contract(
            &mut context,
            token_hash,
            U256::one() * 1_000_000_000_000u64,
            Key::from(bridge_package_hash),
        );

        let accout_balance =
            query_balance(&mut context.builder, token_hash, &context.account.key());

        // Transfer out tokens from bridge
        let deploy_item = simple_deploy_builder(context.account.address)
            .with_stored_session_hash(
                bridge_hash,
                EP_TRANSFER_OUT,
                runtime_args! {
                    PARAM_TOKEN_CONTRACT => token_package_hash,
                    PARAM_AMOUNT => U256::one() * 2_000_000_000_000u64,
                    PARAM_RECIPIENT => context.account.key(),
                },
            )
            .build();

        let error = context
            .builder
            .exec(ExecuteRequestBuilder::from_deploy_item(deploy_item).build())
            .commit()
            .expect_failure()
            .get_error()
            .unwrap();

        let expected_error = engine_state::Error::Exec(execution::Error::Revert(ApiError::User(
            ERC20_INSUFFIENT_BALANCE_ERROR_CODE,
        )));
        assert_eq!(error.to_string(), expected_error.to_string());

        let account_balance_after =
            query_balance(&mut context.builder, token_hash, &context.account.key());
        assert_eq!(account_balance_after, accout_balance);

        let bridge_balance_after = query_balance(
            &mut context.builder,
            token_hash,
            &Key::from(bridge_package_hash),
        );
        assert_eq!(bridge_balance_after, U256::one() * 1_000_000_000_000u64);
    }
}
