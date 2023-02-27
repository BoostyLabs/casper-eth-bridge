pub mod constants;
pub mod utils;

#[cfg(test)]
mod tests {
    use crate::constants::{
        ERC20_INSUFFIENT_BALANCE_ERROR_CODE, TEST_ACCOUNT_BALANCE, TEST_AMOUNT, TEST_BLOCK_TIME,
        TEST_COMMISSION_PERCENT, TEST_CORRECT_DEADLINE, TEST_DESTINATION_ADDRESS,
        TEST_DESTINATION_CHAIN, TEST_EXPIRED_DEADLINE, TEST_GAS_COMMISSION, TEST_NONCE,
        TEST_STABLE_COMMISSION_PERCENT,
    };
    use crate::utils::{
        arbitrary_user, arbitrary_user_key, bridge_in, bridge_out, deploy_bridge,
        deploy_bridge_and_erc20, deploy_erc20, execution_context, execution_error,
        fill_purse_on_token_contract, get_context, query_balance, query_commission_pool,
        read_contract_event, set_test_signer, setup_context, simple_deploy_builder,
        test_public_key, transfer_out, withdraw_commission, UserAccount,
    };
    use casper_engine_test_support::ExecuteRequestBuilder;
    use casper_execution_engine::core::{engine_state, execution};
    use casper_types::bytesrepr::Bytes;
    use casper_types::{runtime_args, RuntimeArgs, U128, U256};
    use casper_types::{ApiError, Key};
    use contract_bridge::entry_points::{EP_CHECK_PARAMS, PARAM_BYTES, PARAM_SIGNATURE};
    use contract_util::signatures::cook_msg_transfer_out;
    use contract_util::{error::Error::Contract as ContractError, signatures::cook_msg_bridge_in};

    use casper_common::event::BridgeEvent;
    use contract_bridge::{
        entry_points::{
            EP_BRIDGE_IN, EP_BRIDGE_IN_CONFIRM, EP_BRIDGE_OUT, EP_GET_SIGNER,
            EP_GET_STABLE_COMMISSION_PERCENT, EP_SET_SIGNER, EP_SET_STABLE_COMMISSION_PERCENT,
            EP_TRANSFER_OUT, EP_WITHDRAW_COMMISSION, PARAM_AMOUNT, PARAM_DESTINATION_ADDRESS,
            PARAM_DESTINATION_CHAIN, PARAM_GAS_COMMISSION, PARAM_NONCE, PARAM_RECIPIENT,
            PARAM_SENDER, PARAM_SIGNER, PARAM_STABLE_COMMISSION_PERCENT, PARAM_TOKEN_CONTRACT,
        },
        error::BridgeError,
    };

    fn expected_total_commission() -> U256 {
        TEST_AMOUNT() * TEST_STABLE_COMMISSION_PERCENT() / 100 + TEST_GAS_COMMISSION()
    }

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
            EP_CHECK_PARAMS,
            EP_BRIDGE_OUT,
            EP_TRANSFER_OUT,
            EP_WITHDRAW_COMMISSION,
            EP_SET_STABLE_COMMISSION_PERCENT,
            EP_GET_STABLE_COMMISSION_PERCENT,
            EP_SET_SIGNER,
            EP_GET_SIGNER,
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
            2. Assert that commission added in the pool
            3. Verify expected event
        */

        let mut context = setup_context();

        let (token_hash, token_package_hash, bridge_hash, bridge_package_hash) =
            deploy_bridge_and_erc20(&mut context.builder, context.account.address);

        let deploy_item = bridge_in(
            bridge_hash,
            token_package_hash,
            context.account.address,
            TEST_AMOUNT(),
            TEST_CORRECT_DEADLINE(),
            TEST_NONCE(),
            TEST_GAS_COMMISSION(),
            Vec::new(),
        );

        get_context(&mut context, deploy_item).expect_success();

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
            gas_commission,
            stable_commission_percent,
            nonce,
            sender,
        } = funds_in_event
        {
            assert_eq!(token_contract, token_package_hash);
            assert_eq!(destination_chain, TEST_DESTINATION_CHAIN());
            assert_eq!(destination_address, TEST_DESTINATION_ADDRESS());
            assert_eq!(amount, TEST_AMOUNT());
            assert_eq!(gas_commission, TEST_GAS_COMMISSION());
            assert_eq!(stable_commission_percent, TEST_STABLE_COMMISSION_PERCENT());
            assert_eq!(nonce, TEST_NONCE());
            assert_eq!(sender, Key::Account(context.account.address));
        } else {
            panic!("wrong bridge event kind");
        }

        let bridge_balance = query_balance(
            &mut context.builder,
            token_hash,
            &Key::Hash(bridge_package_hash.value()),
        );

        let commission_after =
            query_commission_pool(&mut context.builder, bridge_hash, token_package_hash);
        assert_eq!(commission_after, expected_total_commission());
        assert_eq!(bridge_balance, TEST_AMOUNT());
    }

    #[test]
    fn bridge_out_happy_path() {
        /*
            Scenario:

            1. Call "bridge_in" entrypoint in bridge contract with the specified token
            2. Assert that bridge contract received the expected amount of tokens
            3. Call "bridge_out" entrypoint in bridge contract
            4. Assert that the contract lost the expected amount of tokens
            5. Assert that the recipient received the expected amount of tokens
            6. Assert that commission added in the pool and not affected after bridge out
            7. Verify all expected events have been emitted
        */

        let mut context = setup_context();

        // Creating recipient account
        let recipient = arbitrary_user(&mut context);
        let recipient_key = recipient.key();

        let (token_hash, token_package_hash, bridge_hash, bridge_package_hash) =
            deploy_bridge_and_erc20(&mut context.builder, context.account.address);

        let deploy_item = bridge_in(
            bridge_hash,
            token_package_hash,
            context.account.address,
            TEST_AMOUNT(),
            TEST_CORRECT_DEADLINE(),
            TEST_NONCE(),
            TEST_GAS_COMMISSION(),
            Vec::new(),
        );
        get_context(&mut context, deploy_item).expect_success();

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
        assert_eq!(bridge_balance, TEST_AMOUNT());

        let deploy_item = bridge_out(
            bridge_hash,
            token_package_hash,
            context.account.address,
            recipient_key,
            U256::one() * 900_000_000_000u64,
        );
        get_context(&mut context, deploy_item).expect_success();

        let bridge_balance = query_balance(
            &mut context.builder,
            token_hash,
            &Key::from(bridge_package_hash),
        );
        let recipient_balance = query_balance(&mut context.builder, token_hash, &recipient_key);
        let commission_after =
            query_commission_pool(&mut context.builder, bridge_hash, token_package_hash);

        assert_eq!(bridge_balance, U256::one() * 100_000_000_000u64);
        assert_eq!(recipient_balance, U256::one() * 900_000_000_000u64);
        assert_eq!(commission_after, expected_total_commission());
        let bridge_out_event = read_contract_event::<_, BridgeEvent>(
            &mut context.builder,
            bridge_hash,
            "event_trigger",
        );

        if let BridgeEvent::FundsIn {
            token_contract,
            destination_chain,
            destination_address,
            amount,
            gas_commission,
            stable_commission_percent,
            nonce,
            sender,
        } = in_event
        {
            assert_eq!(token_contract, token_package_hash);
            assert_eq!(destination_chain, TEST_DESTINATION_CHAIN());
            assert_eq!(destination_address, TEST_DESTINATION_ADDRESS());
            assert_eq!(amount, TEST_AMOUNT());
            assert_eq!(gas_commission, TEST_GAS_COMMISSION());
            assert_eq!(stable_commission_percent, TEST_STABLE_COMMISSION_PERCENT());
            assert_eq!(nonce, TEST_NONCE());
            assert_eq!(sender, Key::Account(context.account.address));
        } else {
            panic!("Expected BridgeEvent::FundsIn but got {in_event:?}");
        }

        if let BridgeEvent::FundsOut {
            token_contract,
            source_chain,
            source_address,
            recipient,
            amount,
            transaction_id,
        } = bridge_out_event
        {
            assert_eq!(token_contract, token_package_hash);
            assert_eq!(source_chain, "SOUR");
            assert_eq!(source_address, "SOURADDR");
            assert_eq!(recipient, recipient_key);
            assert_eq!(amount, U256::one() * 900_000_000_000u64);
            assert_eq!(transaction_id, U256::one());
        } else {
            panic!("Expected bridge out event, but got {bridge_out_event:?}");
        }
    }

    #[test]
    fn bridge_in_insufficient_tokens() {
        // Timofei1 vvvfix - InvalidContext error because of account in a bridge_in is a sender.address which is not deployed the contract
        /*
            Scenario:
            1. Call "bridge_in" entrypoint in bridge contract with the specified token with an amount that exceeds current balance
            2. Assert that the call failed
        */

        let mut context = setup_context();

        // Create an accoutn for the test
        // Timofei1 - this happens because I use arbitrary account
        let sender = arbitrary_user(&mut context);

        // Deploy the bridge contract and the token contract
        let (token_hash, token_package_hash, bridge_hash, bridge_package_hash) =
            deploy_bridge_and_erc20(&mut context.builder, context.account.address);

        // Create a purse to hold the tokens for the bridge contract to verify that the amount didn't change
        fill_purse_on_token_contract(
            &mut context,
            token_hash,
            U256::one() * 10000,
            Key::from(bridge_package_hash),
        );

        fill_purse_on_token_contract(
            &mut context,
            token_hash,
            U256::one() * 10000000,
            sender.key(),
        );

        let deploy_item = bridge_in(
            bridge_hash,
            token_package_hash,
            sender.address,
            U256::one() * TEST_ACCOUNT_BALANCE * 2,
            TEST_CORRECT_DEADLINE(),
            TEST_NONCE(),
            TEST_GAS_COMMISSION(),
            Vec::new(),
        );

        // Verify that the transaction fails
        let error = execution_error(&mut context, deploy_item);

        let balance = query_balance(
            &mut context.builder,
            token_hash,
            &Key::Hash(bridge_package_hash.value()),
        );

        // The balance didn't change
        assert_eq!(balance, U256::one() * 10000);

        // Verifies that error is expected one.
        let expected_error = engine_state::Error::Exec(execution::Error::Revert(ApiError::User(
            ERC20_INSUFFIENT_BALANCE_ERROR_CODE,
        )));
        assert!(
            matches!(
                error,
                engine_state::Error::Exec(execution::Error::Revert(ApiError::User(
                    ERC20_INSUFFIENT_BALANCE_ERROR_CODE,
                )))
            ),
            "Unexpected error message. Expected: {expected_error}, but got {error}",
        );
    }

    #[test]
    fn bridge_in_incorrect_values_deadline() {
        /*
            Scenario:
            1. Call "bridge_in" entrypoint in bridge contract with incorrect deadline
            2. Assert that all calls failed
        */

        let mut context = setup_context();

        // Deploy the bridge contract and the token contract
        let (_, token_package_hash, bridge_hash, _) =
            deploy_bridge_and_erc20(&mut context.builder, context.account.address);

        let deploy_item = bridge_in(
            bridge_hash,
            token_package_hash,
            context.account.address,
            U256::one() * TEST_ACCOUNT_BALANCE,
            TEST_EXPIRED_DEADLINE(),
            TEST_NONCE(),
            TEST_GAS_COMMISSION(),
            Vec::new(),
        );

        // // Verify that the transaction fails
        let error = execution_error(&mut context, deploy_item);

        let expected_error: ApiError = ContractError(BridgeError::ExpiredSignature).into();
        assert_eq!(error.to_string(), expected_error.to_string());
    }

    #[test]
    fn bridge_in_incorrect_values_nonce_already_used() {
        /*
            Scenario:
            1. Call "bridge_in" entrypoint in bridge contract with incorrect nonce
            2. Assert that all calls failed
        */

        let mut context = setup_context();

        // Deploy the bridge contract and the token contract
        let (_, token_package_hash, bridge_hash, _) =
            deploy_bridge_and_erc20(&mut context.builder, context.account.address);

        let deploy_item_correct = bridge_in(
            bridge_hash,
            token_package_hash,
            context.account.address,
            U256::one() * TEST_ACCOUNT_BALANCE,
            TEST_CORRECT_DEADLINE(),
            TEST_NONCE(),
            TEST_GAS_COMMISSION(),
            Vec::new(),
        );
        get_context(&mut context, deploy_item_correct).expect_success();

        let deploy_item_incorrect = bridge_in(
            bridge_hash,
            token_package_hash,
            context.account.address,
            U256::one() * TEST_ACCOUNT_BALANCE,
            TEST_CORRECT_DEADLINE(),
            TEST_NONCE(),
            TEST_GAS_COMMISSION(),
            Vec::new(),
        );

        // Verify that the transaction fails
        let error = execution_error(&mut context, deploy_item_incorrect);

        let expected_error: ApiError = ContractError(BridgeError::AlreadyUsedSignature).into();
        assert_eq!(error.to_string(), expected_error.to_string());
    }

    #[test]
    fn bridge_in_commission_bigger_than_transferred_amount() {
        /*
            Scenario:
            1. Call "bridge_in" entrypoint in bridge contract with incorrect gas commission
            2. Assert that all calls failed
        */

        let mut context = setup_context();

        // Deploy the bridge contract and the token contract
        let (_, token_package_hash, bridge_hash, _) =
            deploy_bridge_and_erc20(&mut context.builder, context.account.address);

        let deploy_item = bridge_in(
            bridge_hash,
            token_package_hash,
            context.account.address,
            U256::one() * TEST_ACCOUNT_BALANCE,
            TEST_CORRECT_DEADLINE(),
            TEST_NONCE(),
            TEST_GAS_COMMISSION() + TEST_ACCOUNT_BALANCE,
            Vec::new(),
        );

        // // Verify that the transaction fails
        let error = execution_error(&mut context, deploy_item);

        let expected_error: ApiError =
            ContractError(BridgeError::CommissionBiggerThanTransferredAmount).into();
        assert_eq!(error.to_string(), expected_error.to_string());
    }

    #[test]
    fn bridge_in_incorrect_values_signature() {
        /*
            Scenario:
            1. Call "bridge_in" entrypoint in bridge contract with incorrect signature values
            2. Assert that all calls failed
        */

        let mut context = setup_context();
        let mut context2 = setup_context();

        // Deploy the bridge contract and the token contract
        let (_, token_package_hash, bridge_hash, _) =
            deploy_bridge_and_erc20(&mut context.builder, context.account.address);

        let (_, token_package_hash_incorrect, _, _) =
            deploy_bridge_and_erc20(&mut context2.builder, context2.account.address);

        // 1. Incorrect Token
        let bytes_incorrect_deadline = cook_msg_bridge_in(
            bridge_hash,
            token_package_hash_incorrect,
            context.account.address,
            U256::one() * TEST_ACCOUNT_BALANCE,
            TEST_GAS_COMMISSION(),
            TEST_CORRECT_DEADLINE(),
            TEST_NONCE() + 3,
            &TEST_DESTINATION_CHAIN(),
            &TEST_DESTINATION_ADDRESS(),
        );
        let deploy_item = bridge_in(
            bridge_hash,
            token_package_hash,
            context.account.address,
            U256::one() * TEST_ACCOUNT_BALANCE,
            TEST_CORRECT_DEADLINE(),
            TEST_NONCE(),
            TEST_GAS_COMMISSION(),
            bytes_incorrect_deadline,
        );
        let error = execution_error(&mut context, deploy_item);
        let expected_error: ApiError = ContractError(BridgeError::InvalidSignature).into();
        assert_eq!(error.to_string(), expected_error.to_string());

        // 2. Incorrect Amount
        let bytes_incorrect_amount = cook_msg_bridge_in(
            bridge_hash,
            token_package_hash,
            context.account.address,
            U256::one() * (TEST_ACCOUNT_BALANCE + 1),
            TEST_GAS_COMMISSION(),
            TEST_CORRECT_DEADLINE(),
            TEST_NONCE(),
            &TEST_DESTINATION_CHAIN(),
            &TEST_DESTINATION_ADDRESS(),
        );
        let deploy_item = bridge_in(
            bridge_hash,
            token_package_hash,
            context.account.address,
            U256::one() * TEST_ACCOUNT_BALANCE,
            TEST_CORRECT_DEADLINE(),
            TEST_NONCE(),
            TEST_GAS_COMMISSION(),
            bytes_incorrect_amount,
        );
        let error = execution_error(&mut context, deploy_item);
        let expected_error: ApiError = ContractError(BridgeError::InvalidSignature).into();
        assert_eq!(error.to_string(), expected_error.to_string());

        // 3. Incorrect Deadline
        let bytes_incorrect_deadline = cook_msg_bridge_in(
            bridge_hash,
            token_package_hash,
            context.account.address,
            U256::one() * TEST_ACCOUNT_BALANCE,
            TEST_GAS_COMMISSION(),
            TEST_CORRECT_DEADLINE() + 1,
            TEST_NONCE(),
            &TEST_DESTINATION_CHAIN(),
            &TEST_DESTINATION_ADDRESS(),
        );
        let deploy_item = bridge_in(
            bridge_hash,
            token_package_hash,
            context.account.address,
            U256::one() * TEST_ACCOUNT_BALANCE,
            TEST_CORRECT_DEADLINE(),
            TEST_NONCE(),
            TEST_GAS_COMMISSION(),
            bytes_incorrect_deadline,
        );
        let error = execution_error(&mut context, deploy_item);
        let expected_error: ApiError = ContractError(BridgeError::InvalidSignature).into();
        assert_eq!(error.to_string(), expected_error.to_string());

        // 4. Incorrect Nonce
        let bytes_incorrect_deadline = cook_msg_bridge_in(
            bridge_hash,
            token_package_hash,
            context.account.address,
            U256::one() * TEST_ACCOUNT_BALANCE,
            TEST_GAS_COMMISSION(),
            TEST_CORRECT_DEADLINE(),
            TEST_NONCE() + 3,
            &TEST_DESTINATION_CHAIN(),
            &TEST_DESTINATION_ADDRESS(),
        );
        let deploy_item = bridge_in(
            bridge_hash,
            token_package_hash,
            context.account.address,
            U256::one() * TEST_ACCOUNT_BALANCE,
            TEST_CORRECT_DEADLINE(),
            TEST_NONCE(),
            TEST_GAS_COMMISSION(),
            bytes_incorrect_deadline,
        );
        let error = execution_error(&mut context, deploy_item);
        let expected_error: ApiError = ContractError(BridgeError::InvalidSignature).into();
        assert_eq!(error.to_string(), expected_error.to_string());

        // 5. Incorrect User
        let sender = arbitrary_user(&mut context);
        let bytes_incorrect_deadline = cook_msg_bridge_in(
            bridge_hash,
            token_package_hash,
            sender.address,
            U256::one() * TEST_ACCOUNT_BALANCE,
            TEST_GAS_COMMISSION(),
            TEST_CORRECT_DEADLINE(),
            TEST_NONCE() + 3,
            &TEST_DESTINATION_CHAIN(),
            &TEST_DESTINATION_ADDRESS(),
        );
        let deploy_item = bridge_in(
            bridge_hash,
            token_package_hash,
            context.account.address,
            U256::one() * TEST_ACCOUNT_BALANCE,
            TEST_CORRECT_DEADLINE(),
            TEST_NONCE(),
            TEST_GAS_COMMISSION(),
            bytes_incorrect_deadline,
        );
        let error = execution_error(&mut context, deploy_item);
        let expected_error: ApiError = ContractError(BridgeError::InvalidSignature).into();
        assert_eq!(error.to_string(), expected_error.to_string());


        // 6. Incorrect Destination Chain
        let bytes_incorrect_deadline = cook_msg_bridge_in(
            bridge_hash,
            token_package_hash,
            context.account.address,
            U256::one() * TEST_ACCOUNT_BALANCE,
            TEST_GAS_COMMISSION(),
            TEST_CORRECT_DEADLINE(),
            TEST_NONCE(),
            "WRONG_CHAIN",
            &TEST_DESTINATION_ADDRESS(),
        );
        let deploy_item = bridge_in(
            bridge_hash,
            token_package_hash,
            context.account.address,
            U256::one() * TEST_ACCOUNT_BALANCE,
            TEST_CORRECT_DEADLINE(),
            TEST_NONCE(),
            TEST_GAS_COMMISSION(),
            bytes_incorrect_deadline,
        );
        let error = execution_error(&mut context, deploy_item);
        let expected_error: ApiError = ContractError(BridgeError::InvalidSignature).into();
        assert_eq!(error.to_string(), expected_error.to_string());

        // 6. Incorrect Destination Address
        let bytes_incorrect_deadline = cook_msg_bridge_in(
            bridge_hash,
            token_package_hash,
            context.account.address,
            U256::one() * TEST_ACCOUNT_BALANCE,
            TEST_GAS_COMMISSION(),
            TEST_CORRECT_DEADLINE(),
            TEST_NONCE(),
            &TEST_DESTINATION_CHAIN(),
            "WRONG_ADDRESS",
        );
        let deploy_item = bridge_in(
            bridge_hash,
            token_package_hash,
            context.account.address,
            U256::one() * TEST_ACCOUNT_BALANCE,
            TEST_CORRECT_DEADLINE(),
            TEST_NONCE(),
            TEST_GAS_COMMISSION(),
            bytes_incorrect_deadline,
        );
        let error = execution_error(&mut context, deploy_item);
        let expected_error: ApiError = ContractError(BridgeError::InvalidSignature).into();
        assert_eq!(error.to_string(), expected_error.to_string());


    }

    #[test]
    fn transfer_out_incorrect_values_signature() {
        /*
            Scenario:
            1. Call "transfer_out" entrypoint in bridge contract with incorrect signature values
            2. Assert that all calls failed
        */

        let mut context = setup_context();
        let mut context2 = setup_context();
        let (_, token_package_hash_incorrect, _, _) =
            deploy_bridge_and_erc20(&mut context2.builder, context2.account.address);
        let recipient_key = arbitrary_user_key(&mut context);

        let (_, token_package_hash, bridge_hash, _) =
            deploy_bridge_and_erc20(&mut context.builder, context.account.address);
        let deploy_item = bridge_in(
            bridge_hash,
            token_package_hash,
            context.account.address,
            TEST_AMOUNT(),
            TEST_CORRECT_DEADLINE(),
            TEST_NONCE(),
            TEST_GAS_COMMISSION(),
            Vec::new(),
        );
        get_context(&mut context, deploy_item).expect_success();

        let bytes = cook_msg_transfer_out(
            bridge_hash,
            token_package_hash,
            recipient_key,
            (TEST_AMOUNT()) - expected_total_commission(),
            expected_total_commission(),
            TEST_NONCE() + 1,
        );

        // 1. Incorrect Token
        let deploy_item = transfer_out(
            bridge_hash,
            token_package_hash_incorrect,
            context.account.address,
            recipient_key,
            (TEST_AMOUNT()) - expected_total_commission(),
            expected_total_commission(),
            TEST_NONCE() + 1,
            bytes.clone(),
        );

        let error = execution_error(&mut context, deploy_item);
        let expected_error: ApiError = ContractError(BridgeError::InvalidSignature).into();
        assert_eq!(error.to_string(), expected_error.to_string());

        // 2. Incorrect Amount
        let deploy_item = transfer_out(
            bridge_hash,
            token_package_hash,
            context.account.address,
            recipient_key,
            (TEST_AMOUNT()) - expected_total_commission() + 1,
            expected_total_commission(),
            TEST_NONCE() + 1,
            bytes.clone(),
        );

        let error = execution_error(&mut context, deploy_item);
        let expected_error: ApiError = ContractError(BridgeError::InvalidSignature).into();
        assert_eq!(error.to_string(), expected_error.to_string());

        // 3. Invalid commission
        let deploy_item = transfer_out(
            bridge_hash,
            token_package_hash,
            context.account.address,
            recipient_key,
            (TEST_AMOUNT()) - expected_total_commission(),
            expected_total_commission() - 10,
            TEST_NONCE() + 1,
            bytes.clone(),
        );
        let error = execution_error(&mut context, deploy_item);
        let expected_error: ApiError = ContractError(BridgeError::InvalidSignature).into();
        assert_eq!(error.to_string(), expected_error.to_string());

        // 4. Invalid recipient
        let recipient_key_invalid = UserAccount::unique_account(&mut context2, 1).key();

        let deploy_item = transfer_out(
            bridge_hash,
            token_package_hash,
            context.account.address,
            recipient_key_invalid,
            (TEST_AMOUNT()) - expected_total_commission(),
            expected_total_commission(),
            TEST_NONCE() + 1,
            bytes.clone(),
        );
        let error = execution_error(&mut context, deploy_item);
        let expected_error: ApiError = ContractError(BridgeError::InvalidSignature).into();
        assert_eq!(error.to_string(), expected_error.to_string());

    }

    #[test]
    fn bridge_out_insufficient_tokens() {
        /*
            Scenario:

            1. Call "bridge_in" entrypoint in bridge contract with the specified token
            2. Call "bridge_out" entrypoint in bridge contract with the specified token with an amount that exceeds the bridge's current balance MINUS commission
            3. Assert that the call failed because we did not consider locked commission and trying to withdraw the same amount
        */

        let mut context = setup_context();

        let recipient_key = arbitrary_user_key(&mut context);

        // Deploy bridge and token contracts.
        let (token_hash, token_package_hash, bridge_hash, bridge_package_hash) =
            deploy_bridge_and_erc20(&mut context.builder, context.account.address);

        let deploy_item = bridge_in(
            bridge_hash,
            token_package_hash,
            context.account.address,
            TEST_AMOUNT(),
            TEST_CORRECT_DEADLINE(),
            TEST_NONCE(),
            TEST_GAS_COMMISSION(),
            Vec::new(),
        );
        get_context(&mut context, deploy_item).expect_success();

        // Transfer tokens from bridge to recipient that exceeds the bridge's current balance.
        let deploy_item = bridge_out(
            bridge_hash,
            token_package_hash,
            context.account.address,
            recipient_key,
            TEST_AMOUNT(),
        );

        // Verify that transaction fails.
        let error = execution_error(&mut context, deploy_item);

        let bridge_balance = query_balance(
            &mut context.builder,
            token_hash,
            &Key::Hash(bridge_package_hash.value()),
        );

        // Verify that the balance of the bridge is still 3_000_000_000_000 tokens.
        assert_eq!(bridge_balance, TEST_AMOUNT());

        let expected_error: ApiError = ContractError(BridgeError::AmountExceedBridgePool).into();
        assert_eq!(
            error.to_string(),
            expected_error.to_string(),
            "Unexpected error message. Expected: {expected_error}, but got {error}"
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
        let (token_hash, token_package_hash, bridge_hash, bridge_package_hash) =
            deploy_bridge_and_erc20(&mut context.builder, context.account.address);

        fill_purse_on_token_contract(
            &mut context,
            token_hash,
            TEST_AMOUNT(),
            Key::from(bridge_package_hash),
        );

        // Transfer tokens from bridge to recipient called by non-owner.
        let deploy_item = bridge_out(
            bridge_hash,
            token_package_hash,
            recipient.address,
            recipient_key,
            TEST_AMOUNT(),
        );

        // Verify that transaction fails.
        let error = execution_error(&mut context, deploy_item);

        let balance = query_balance(
            &mut context.builder,
            token_hash,
            &Key::from(bridge_package_hash),
        );
        assert_eq!(balance, TEST_AMOUNT());

        let expected_error = engine_state::Error::Exec(execution::Error::InvalidContext);
        assert_eq!(error.to_string(), expected_error.to_string());
    }

    #[test]
    fn bridge_in_confirm_not_public_available() {
        /* Call to bridge in confign shouldn't be available except from contract context */
        let mut context = setup_context();

        // Creating account for the test.
        let test_subj = arbitrary_user(&mut context);

        // Deploy bridge.
        let (_, token_package_hash, bridge_hash, _) =
            deploy_bridge_and_erc20(&mut context.builder, context.account.address);

        for user in &[&context.account, &test_subj] {
            let args = runtime_args! {
                PARAM_TOKEN_CONTRACT => token_package_hash,
                PARAM_AMOUNT => TEST_AMOUNT(),
                PARAM_GAS_COMMISSION => TEST_GAS_COMMISSION(),
                PARAM_NONCE => TEST_NONCE(),
                PARAM_DESTINATION_ADDRESS => TEST_DESTINATION_ADDRESS(),
                PARAM_DESTINATION_CHAIN => TEST_DESTINATION_CHAIN(),
                PARAM_SENDER => user.key(),
            };

            let deploy_item = simple_deploy_builder(user.address)
                .with_stored_session_hash(bridge_hash, EP_BRIDGE_IN_CONFIRM, args.clone())
                .build();
            // Verify that transaction fails.
            let error = context
                .builder
                .exec(
                    ExecuteRequestBuilder::from_deploy_item(deploy_item)
                        .with_block_time(TEST_BLOCK_TIME)
                        .build(),
                )
                .commit()
                .expect_failure()
                .get_error()
                .unwrap();

            let api_error: ApiError = ContractError(BridgeError::OnlyCallableBySelf).into();
            let expected_error = engine_state::Error::Exec(execution::Error::Revert(api_error));
            assert_eq!(error.to_string(), expected_error.to_string());
        }
    }

    #[test]
    fn verify_signature_not_public_available() {
        /* Call to bridge in confign shouldn't be available except from contract context */
        let mut context = setup_context();

        // Creating account for the test.
        let test_subj = arbitrary_user(&mut context);

        // Deploy bridge.
        let (_, _, bridge_hash, _) =
            deploy_bridge_and_erc20(&mut context.builder, context.account.address);

        for user in &[&context.account, &test_subj] {
            let args = runtime_args! {
                PARAM_BYTES => Bytes::new(),
                PARAM_SIGNATURE => [0; 64],
                PARAM_SIGNER => test_public_key(),
                PARAM_NONCE => U128::one(),
            };

            let deploy_item = simple_deploy_builder(user.address)
                .with_stored_session_hash(bridge_hash, EP_CHECK_PARAMS, args.clone())
                .build();
            // Verify that transaction fails.
            let error = context
                .builder
                .exec(
                    ExecuteRequestBuilder::from_deploy_item(deploy_item)
                        .with_block_time(TEST_BLOCK_TIME)
                        .build(),
                )
                .commit()
                .expect_failure()
                .get_error()
                .unwrap();

            let api_error: ApiError = ContractError(BridgeError::OnlyCallableBySelf).into();
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
        let (token_hash, token_package_hash, bridge_hash, bridge_package_hash) =
            deploy_bridge_and_erc20(&mut context.builder, context.account.address);

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
                    PARAM_AMOUNT => TEST_AMOUNT(),
                    PARAM_RECIPIENT => Key::from(context.account.address),
                },
            )
            .build();

        get_context(&mut context, deploy_item).expect_success();

        let owner_balance = query_balance(
            &mut context.builder,
            token_hash,
            &Key::from(context.account.address),
        );
        assert_eq!(owner_balance, TEST_AMOUNT());

        let deploy_item = bridge_in(
            bridge_hash,
            token_package_hash,
            context.account.address,
            U256::one() * 1000,
            TEST_CORRECT_DEADLINE(),
            TEST_NONCE(),
            TEST_GAS_COMMISSION(),
            Vec::new(),
        );

        // Verify that transaction fails.
        let error = execution_error(&mut context, deploy_item);

        let api_error = ContractError(BridgeError::UnexpectedTransferAmount);
        let expected_error = engine_state::Error::Exec(execution::Error::Revert(api_error.into()));

        assert_eq!(error.to_string(), expected_error.to_string());
    }

    #[test]
    fn transfer_out_funds() {
        /*
           Scenario:
           * Transfer tokens for bridge contract via ERC-20's "transfer" entrypoint
           * Verify that tokens are transferred to bridge contract
           * Transfer out to some account
           * Verify that tokens are transferred together with comission and initial amount taken by this account.
           * Verify that contract balance and commission pool are empty
           * Verify event
        */

        let mut context = setup_context();

        let recipient_key = arbitrary_user_key(&mut context);
        // Deploy bridge and token contracts.
        let (token_hash, token_package_hash, bridge_hash, bridge_package_hash) =
            deploy_bridge_and_erc20(&mut context.builder, context.account.address);

        let deploy_item = bridge_in(
            bridge_hash,
            token_package_hash,
            context.account.address,
            TEST_AMOUNT(),
            TEST_CORRECT_DEADLINE(),
            TEST_NONCE(),
            TEST_GAS_COMMISSION(),
            Vec::new(),
        );
        get_context(&mut context, deploy_item).expect_success();

        let deploy_item = transfer_out(
            bridge_hash,
            token_package_hash,
            context.account.address,
            recipient_key,
            (TEST_AMOUNT()) - expected_total_commission(),
            expected_total_commission(),
            TEST_NONCE() + 1,
            Vec::new(),
        );
        get_context(&mut context, deploy_item).expect_success();

        let transfer_out_event = read_contract_event::<_, BridgeEvent>(
            &mut context.builder,
            bridge_hash,
            "event_trigger",
        );
        if let BridgeEvent::TransferOut {
            token_contract,
            total_sum_for_transfer,
            nonce,
            recipient,
        } = transfer_out_event
        {
            assert_eq!(token_contract, token_package_hash);
            assert_eq!(recipient, recipient_key);
            assert_eq!(total_sum_for_transfer, TEST_AMOUNT(), "HEH");
            assert_eq!(nonce, TEST_NONCE() + 1);
        } else {
            panic!("Expected bridge out event, but got {transfer_out_event:?}");
        }
        let balance_after = query_balance(
            &mut context.builder,
            token_hash,
            &Key::from(bridge_package_hash),
        );
        let recipient_balance = query_balance(&mut context.builder, token_hash, &recipient_key);
        let commission_after =
            query_commission_pool(&mut context.builder, bridge_hash, token_package_hash);

        assert_eq!(commission_after, U256::zero());
        assert_eq!(balance_after, U256::zero());
        assert_eq!(recipient_balance, TEST_AMOUNT());
    }

    #[test]
    fn transfer_out_insuffient_balance() {
        /*
           Scenario:
           * Call "bridge_in" entrypoint in bridge contract with the specified token
           * Try to use transfer out entrypoint from the bridge contract with request that will be higher than the balance of the bridge contract
           * Verify that transaction fails because of the balance of the bridge contract is not enough to transfer out
        */

        let mut context = setup_context();

        let recipient_key = arbitrary_user_key(&mut context);

        // Deploy bridge and token contracts.
        let (token_hash, token_package_hash, bridge_hash, bridge_package_hash) =
            deploy_bridge_and_erc20(&mut context.builder, context.account.address);

        let deploy_item = bridge_in(
            bridge_hash,
            token_package_hash,
            context.account.address,
            TEST_AMOUNT(),
            TEST_CORRECT_DEADLINE(),
            TEST_NONCE(),
            TEST_GAS_COMMISSION(),
            Vec::new(),
        );
        get_context(&mut context, deploy_item).expect_success();

        // Transfer out tokens from bridge
        let deploy_item = transfer_out(
            bridge_hash,
            token_package_hash,
            context.account.address,
            recipient_key,
            TEST_AMOUNT() * 2,
            expected_total_commission(),
            TEST_NONCE() + 1,
            Vec::new(),
        );

        let error = execution_error(&mut context, deploy_item);

        let expected_error = engine_state::Error::Exec(execution::Error::Revert(ApiError::User(
            ERC20_INSUFFIENT_BALANCE_ERROR_CODE,
        )));
        assert_eq!(error.to_string(), expected_error.to_string());

        let bridge_balance_after = query_balance(
            &mut context.builder,
            token_hash,
            &Key::from(bridge_package_hash),
        );
        let commission_after =
            query_commission_pool(&mut context.builder, bridge_hash, token_package_hash);

        assert_eq!(commission_after, expected_total_commission());
        assert_eq!(bridge_balance_after, TEST_AMOUNT());
    }

    #[test]
    fn set_stable_commission_percent() {
        /*
            Scenario:
            1. Call "set_stable_commission_percent" entrypoint in set percent
            2. Assert that the percent is established
        */

        let mut context = setup_context();

        // Deploy the bridge contract and the token contract
        let (bridge_hash, _) = deploy_bridge(&mut context.builder, context.account.address);

        let stable_commission_percent = TEST_COMMISSION_PERCENT();
        // Try to transfer token in bridge from account that doesn't have enough tokens
        let deploy_item = simple_deploy_builder(context.account.address)
            .with_stored_session_hash(
                bridge_hash,
                EP_SET_STABLE_COMMISSION_PERCENT,
                runtime_args! {
                    PARAM_STABLE_COMMISSION_PERCENT => stable_commission_percent,
                },
            )
            .build();

        let res: U256 = get_context(&mut context, deploy_item)
            .expect_success()
            .get_value(bridge_hash, PARAM_STABLE_COMMISSION_PERCENT);

        assert_eq!(res, stable_commission_percent);
    }

    #[test]
    fn set_signer_happy_path() {
        /*
            Scenario:
            1. Call "set_signer" entrypoint in set signer
            2. Assert that the signer is established
        */

        let mut context = setup_context();

        // Deploy the bridge contract and the token contract
        let (bridge_hash, _) = deploy_bridge(&mut context.builder, context.account.address);

        // Try to transfer token in bridge from account that doesn't have enough tokens
        let deploy_item = set_test_signer(bridge_hash, context.account.address, test_public_key());
        let res: String = get_context(&mut context, deploy_item)
            .expect_success()
            .get_value(bridge_hash, PARAM_SIGNER);

        assert_eq!(res, test_public_key());
    }

    #[test]
    fn set_signer_invalid() {
        /*
            Scenario:
            1. Call "set_signer" entrypoint with invalid value
            2. Assert fail
        */

        let mut context = setup_context();

        // Deploy the bridge contract and the token contract
        let (bridge_hash, _) = deploy_bridge(&mut context.builder, context.account.address);

        // Try to transfer token in bridge from account that doesn't have enough tokens
        let deploy_item = simple_deploy_builder(context.account.address)
            .with_stored_session_hash(
                bridge_hash,
                EP_SET_SIGNER,
                runtime_args! {
                    PARAM_SIGNER => "Any string",
                },
            )
            .build();

        execution_context(&mut context, deploy_item).expect_failure();
    }

    #[test]
    fn set_stable_commission_percent_invalid_value() {
        /*
            Scenario:
            1. Call "set_stable_commission_percent" entrypoint in set percent
            2. Assert fail
        */

        let mut context = setup_context();

        // Deploy the bridge contract
        let (bridge_hash, _) = deploy_bridge(&mut context.builder, context.account.address);

        let stable_commission_percent = U256::one() * 101;
        // Try to set percent
        let deploy_item = simple_deploy_builder(context.account.address)
            .with_stored_session_hash(
                bridge_hash,
                EP_SET_STABLE_COMMISSION_PERCENT,
                runtime_args! {
                    PARAM_STABLE_COMMISSION_PERCENT => stable_commission_percent,
                },
            )
            .build();

        let error = execution_error(&mut context, deploy_item);

        let expected_error: ApiError = ContractError(BridgeError::InvalidCommissionPercent).into();
        assert_eq!(error.to_string(), expected_error.to_string());
    }

    #[test]
    fn set_stable_commission_percent_called_by_non_owner() {
        /*
            Scenario:
            1. Call "set_stable_commission_percent" entrypoint to set percent
            2. Assert fail
        */

        let mut context = setup_context();

        // Deploy the bridge contract
        let (bridge_hash, _) = deploy_bridge(&mut context.builder, context.account.address);

        // Try to set percent
        let user = arbitrary_user(&mut context);
        let deploy_item = simple_deploy_builder(user.address)
            .with_stored_session_hash(
                bridge_hash,
                EP_SET_STABLE_COMMISSION_PERCENT,
                runtime_args! {
                    PARAM_STABLE_COMMISSION_PERCENT => TEST_STABLE_COMMISSION_PERCENT(),
                },
            )
            .build();

        let error = execution_error(&mut context, deploy_item);

        let expected_error = engine_state::Error::Exec(execution::Error::InvalidContext);
        assert_eq!(error.to_string(), expected_error.to_string());
    }

    #[test]
    fn withdraw_commission_test() {
        /*
            Scenario:
            1. Call "bridge_in" entrypoint to transfer tokens in contract.
            2. Call "withdraw_commission" entrypoint
            3. Assert received commission
            4. Assert commission in pool
            5. Assert contract balance
            6. Assert event
        */

        let mut context = setup_context();

        let (token_hash, token_package_hash, bridge_hash, bridge_package_hash) =
            deploy_bridge_and_erc20(&mut context.builder, context.account.address);

        let deploy_item = bridge_in(
            bridge_hash,
            token_package_hash,
            context.account.address,
            TEST_AMOUNT(),
            TEST_CORRECT_DEADLINE(),
            TEST_NONCE(),
            TEST_GAS_COMMISSION(),
            Vec::new(),
        );
        get_context(&mut context, deploy_item).expect_success();

        let recipient_key = arbitrary_user_key(&mut context);

        let commission_in_pool_before =
            query_commission_pool(&mut context.builder, bridge_hash, token_package_hash);
        assert_eq!(commission_in_pool_before, expected_total_commission());

        let deploy_item = withdraw_commission(
            bridge_hash,
            token_package_hash,
            context.account.address,
            recipient_key,
            expected_total_commission(),
        );
        get_context(&mut context, deploy_item).expect_success();
        let withdraw_commission_event = read_contract_event::<_, BridgeEvent>(
            &mut context.builder,
            bridge_hash,
            "event_trigger",
        );

        let commission_in_pool_after =
            query_commission_pool(&mut context.builder, bridge_hash, token_package_hash);

        let bridge_balance_after = query_balance(
            &mut context.builder,
            token_hash,
            &Key::from(bridge_package_hash),
        );

        let recipient_balance = query_balance(&mut context.builder, token_hash, &recipient_key);

        assert_eq!(recipient_balance, commission_in_pool_before);
        assert_eq!(commission_in_pool_after, U256::zero());
        assert_eq!(
            bridge_balance_after,
            TEST_AMOUNT() - commission_in_pool_before
        );

        if let BridgeEvent::WithdrawCommission {
            token_contract,
            amount,
        } = withdraw_commission_event
        {
            assert_eq!(token_contract, token_package_hash);
            assert_eq!(amount, commission_in_pool_before);
        } else {
            panic!("Expected bridge out event, but got {withdraw_commission_event:?}");
        }
    }

    #[test]
    fn withdraw_commission_insufficient_funds() {
        /*
            Scenario:
            1. Call "bridge_in" entrypoint to transfer tokens in contract.
            2. Call "withdraw_commission" entrypoint
            3. Assert fail
            4. Assert received commission
            5. Assert commission in pool
            6. Assert contract balance
        */

        let mut context = setup_context();

        let (_, token_package_hash, bridge_hash, _) =
            deploy_bridge_and_erc20(&mut context.builder, context.account.address);

        let deploy_item = bridge_in(
            bridge_hash,
            token_package_hash,
            context.account.address,
            TEST_AMOUNT(),
            TEST_CORRECT_DEADLINE(),
            TEST_NONCE(),
            TEST_GAS_COMMISSION(),
            Vec::new(),
        );
        get_context(&mut context, deploy_item).expect_success();

        let recipient_key = arbitrary_user_key(&mut context);

        let commission_in_pool_before =
            query_commission_pool(&mut context.builder, bridge_hash, token_package_hash);
        assert_eq!(commission_in_pool_before, expected_total_commission());

        let deploy_item = withdraw_commission(
            bridge_hash,
            token_package_hash,
            context.account.address,
            recipient_key,
            expected_total_commission() * 2,
        );
        let error = execution_error(&mut context, deploy_item);

        let expected_error: ApiError =
            ContractError(BridgeError::AmountExceedCommissionPool).into();
        assert_eq!(error.to_string(), expected_error.to_string());
    }

    #[test]
    fn withdraw_commission_called_by_non_owner() {
        /*
            Scenario:
            1. Call "withdraw_commission" entrypoint
            2. Assert fail
        */

        let mut context = setup_context();

        let (_, token_package_hash, bridge_hash, _) =
            deploy_bridge_and_erc20(&mut context.builder, context.account.address);

        let user = arbitrary_user(&mut context);
        let recipient_key = arbitrary_user_key(&mut context);

        let deploy_item = withdraw_commission(
            bridge_hash,
            token_package_hash,
            user.address,
            recipient_key,
            U256::one() * 1_000,
        );

        let error = execution_error(&mut context, deploy_item);

        let expected_error = engine_state::Error::Exec(execution::Error::InvalidContext);
        assert_eq!(error.to_string(), expected_error.to_string());
    }
}
