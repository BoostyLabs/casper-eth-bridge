pub mod abi;
pub mod util;

#[cfg(test)]
mod test {
    use ethers::prelude::U256;

    use crate::{
        abi::{BridgeContractEvents, BridgeFundsInFilter, BridgeFundsOutFilter},
        util::{deploy_contracts, fill_purse, get_bridge_event, guest_provider, AnvilContext},
    };

    #[tokio::test]
    async fn test_deploy_contracts() {
        let anvil = AnvilContext::default();
        let _ = deploy_contracts(&anvil.provider).await;
    }

    #[tokio::test]
    async fn bridge_in_happy_path() {
        /*
           Scenario:
           * Deploy contracts
           * Send funds to bridge
           * Check that funds were received
           * Check the event
        */
        let anvil = AnvilContext::default();
        let sender = anvil.instance.addresses()[0].clone();

        let contracts = deploy_contracts(&anvil.provider).await;
        let bridge_address = contracts.bridge_contract().address();

        contracts
            .token_contract()
            .approve(bridge_address, 1_000_000_000u64.into())
            .send()
            .await
            .expect("increase_allowance");

        contracts
            .bridge_contract()
            .bridge_in(
                contracts.token_contract().address(),
                sender,
                1_000_000_000u64.into(),
                "HLLO".into(),
                "0x1234".into(),
            )
            .send()
            .await
            .expect("bridge_in_call");

        let bridge_balance = contracts
            .token_contract()
            .balance_of(bridge_address)
            .call()
            .await
            .expect("balance_of bridge");

        assert_eq!(bridge_balance, 1_000_000_000u64.into());

        let in_event = get_bridge_event(&contracts).await;

        if let BridgeContractEvents::BridgeFundsInFilter(BridgeFundsInFilter {
            token,
            sender,
            amount,
            destination_chain,
            destination_address,
        }) = in_event
        {
            assert_eq!(token, contracts.token_contract().address());
            assert_eq!(sender, anvil.instance.addresses()[0]);
            assert_eq!(amount, U256::from(1_000_000_000));
            assert_eq!(destination_chain, "HLLO".to_string());
            assert_eq!(destination_address, "0x1234".to_string());
        } else {
            panic!("Expected BridgeFundsInFilter");
        }
    }

    #[tokio::test]
    async fn bridge_out_happy_path() {
        /*
           Scenario:
           * Deploy contracts
           * Transfer funds into the bridge purse using ERC20 token contract
           * Bridge out funds using bridge contract
           * Check that funds were sent
           * Check the event
        */

        let anvil = AnvilContext::default();
        let recipient = anvil.instance.addresses()[1].clone();

        let contracts = deploy_contracts(&anvil.provider).await;
        let bridge_address = contracts.bridge_contract().address();

        fill_purse(&contracts, 1_000_000_000u64.into(), bridge_address).await;
        contracts
            .bridge_contract()
            .bridge_out(
                contracts.token_contract().address(),
                recipient,
                1_000_000_000.into(),
                "HLLO".into(),
                "0x1234".into(),
            )
            .send()
            .await
            .expect("bridge_out_call");
        let contract_balance = contracts
            .token_contract()
            .balance_of(bridge_address)
            .call()
            .await
            .expect("balance_of bridge");
        assert_eq!(contract_balance, 0.into());

        let user_balance = contracts
            .token_contract()
            .balance_of(recipient)
            .call()
            .await
            .expect("balance_of user");
        assert_eq!(user_balance, 1_000_000_000.into());

        let out_event = get_bridge_event(&contracts).await;

        if let BridgeContractEvents::BridgeFundsOutFilter(BridgeFundsOutFilter {
            token,
            recipient,
            amount,
            source_chain,
            source_address,
        }) = out_event
        {
            assert_eq!(token, contracts.token_contract().address());
            assert_eq!(recipient, anvil.instance.addresses()[1]);
            assert_eq!(amount, U256::from(1_000_000_000));
            assert_eq!(source_chain, "HLLO".to_string());
            assert_eq!(source_address, "0x1234".to_string());
        } else {
            panic!("Expected BridgeFundsOutFilter");
        }
    }

    #[tokio::test]
    async fn transfer_out_success() {
        /*
           Scenario:
           * Deploy contracts
           * Transfer funds into the bridge purse using ERC20 token contract
           * Transfer out funds using bridge contract
           * Check that funds were sent
        */

        let anvil = AnvilContext::default();
        let recipient = anvil.instance.addresses()[1].clone();

        let contracts = deploy_contracts(&anvil.provider).await;
        let bridge_address = contracts.bridge_contract().address();

        fill_purse(&contracts, 1_000_000_000u64.into(), bridge_address).await;
        contracts
            .bridge_contract()
            .transfer_out(
                contracts.token_contract().address(),
                recipient,
                1_000_000_000.into(),
            )
            .send()
            .await
            .expect("bridge_out_call");

        let contract_balance = contracts
            .token_contract()
            .balance_of(bridge_address)
            .call()
            .await
            .expect("balance_of bridge");

        assert_eq!(contract_balance, 0.into());

        let user_balance = contracts
            .token_contract()
            .balance_of(recipient)
            .call()
            .await
            .expect("balance_of user");
        assert_eq!(user_balance, 1_000_000_000.into());
    }

    #[tokio::test]
    async fn bridge_in_out_happy_path() {
        let anvil = AnvilContext::default();
        let sender = anvil.instance.addresses()[0].clone();
        let recipient = anvil.instance.addresses()[1].clone();

        let contracts = deploy_contracts(&anvil.provider).await;
        let bridge_address = contracts.bridge_contract().address();

        contracts
            .token_contract()
            .approve(bridge_address, 1_000_000_000u64.into())
            .send()
            .await
            .expect("increase_allowance");

        contracts
            .bridge_contract()
            .bridge_in(
                contracts.token_contract().address(),
                sender,
                1_000_000_000u64.into(),
                "HLLO".into(),
                "0x1234".into(),
            )
            .send()
            .await
            .expect("bridge_in_call");

        let bridge_balance = contracts
            .token_contract()
            .balance_of(bridge_address)
            .call()
            .await
            .expect("balance_of bridge");

        assert_eq!(bridge_balance, 1_000_000_000u64.into());

        let in_event = get_bridge_event(&contracts).await;

        contracts
            .bridge_contract()
            .bridge_out(
                contracts.token_contract().address(),
                recipient,
                500_000_000u64.into(),
                "HLLO".into(),
                "0x1234".into(),
            )
            .send()
            .await
            .expect("bridge_out_call");

        let bridge_balance = contracts
            .token_contract()
            .balance_of(bridge_address)
            .call()
            .await
            .expect("balance_of bridge");

        assert_eq!(bridge_balance, U256::from(500_000_000));

        let user_balance = contracts
            .token_contract()
            .balance_of(recipient)
            .call()
            .await
            .expect("balance_of user");

        assert_eq!(user_balance, U256::from(500_000_000));
        let out_event = get_bridge_event(&contracts).await;

        if let BridgeContractEvents::BridgeFundsInFilter(BridgeFundsInFilter {
            token,
            sender,
            amount,
            destination_chain,
            destination_address,
        }) = in_event
        {
            assert_eq!(token, contracts.token_contract().address());
            assert_eq!(sender, anvil.instance.addresses()[0]);
            assert_eq!(amount, U256::from(1_000_000_000));
            assert_eq!(destination_chain, "HLLO".to_string());
            assert_eq!(destination_address, "0x1234".to_string());
        } else {
            panic!("Expected BridgeFundsInFilter");
        }

        if let BridgeContractEvents::BridgeFundsOutFilter(BridgeFundsOutFilter {
            token,
            recipient,
            amount,
            source_chain,
            source_address,
        }) = out_event
        {
            assert_eq!(token, contracts.token_contract().address());
            assert_eq!(recipient, anvil.instance.addresses()[1]);
            assert_eq!(amount, U256::from(500_000_000));
            assert_eq!(source_chain, "HLLO".to_string());
            assert_eq!(source_address, "0x1234".to_string());
        } else {
            panic!("Expected BridgeFundsOutFilter");
        }
    }

    #[tokio::test]
    async fn bridge_out_insuffient_tokens() {
        /*
           Scenario:
           * Transfer some funds to the bridge purse
           * Request bridge out that exceeds balance
           * Verify the call fails
        */

        let anvil = AnvilContext::default();
        let recipient = anvil.instance.addresses()[1].clone();

        let contracts = deploy_contracts(&anvil.provider).await;
        let bridge_address = contracts.bridge_contract().address();

        fill_purse(&contracts, 1_000_000_000u64.into(), bridge_address).await;
        contracts
            .bridge_contract()
            .bridge_out(
                contracts.token_contract().address(),
                recipient,
                2_000_000_000.into(),
                "HLLO".into(),
                "0x1234".into(),
            )
            .send()
            .await
            .expect_err("Expected to fail");
    }

    #[tokio::test]
    async fn transfer_out_insuffient_tokens() {
        /*
           Scenario:
           * Transfer some funds to the bridge purse
           * Request bridge out that exceeds balance
           * Verify the call fails
        */

        let anvil = AnvilContext::default();
        let recipient = anvil.instance.addresses()[1].clone();

        let contracts = deploy_contracts(&anvil.provider).await;
        let bridge_address = contracts.bridge_contract().address();

        fill_purse(&contracts, 1_000_000_000u64.into(), bridge_address).await;
        contracts
            .bridge_contract()
            .transfer_out(
                contracts.token_contract().address(),
                recipient,
                2_000_000_000.into(),
            )
            .send()
            .await
            .expect_err("Expected to fail");
    }

    #[tokio::test]
    async fn bridge_in_insuffient_tokens() {
        /*
           Scenario:
           * Transfer some funds to the bridge purse
           * Request bridge out that exceeds balance
           * Verify the call fails
        */

        let anvil = AnvilContext::default();

        let contracts = deploy_contracts(&anvil.provider).await;
        let sender_provider = guest_provider(&anvil).await;
        let sender = sender_provider.address();

        fill_purse(&contracts, 1_000_000_000u64.into(), sender).await;
        contracts
            .token_contract_by(sender_provider.clone())
            .approve(contracts.bridge_contract().address(), 2_000_000_000.into())
            .send()
            .await
            .expect("approve");

        contracts
            .bridge_contract_by(sender_provider)
            .bridge_in(
                contracts.token_contract().address(),
                sender,
                2_000_000_000.into(),
                "HLLO".into(),
                "0x1234".into(),
            )
            .send()
            .await
            .expect_err("Expected to fail");
    }

    #[tokio::test]
    async fn bridge_out_not_operator() {
        let anvil = AnvilContext::default();

        let contracts = deploy_contracts(&anvil.provider).await;
        let sender_provider = guest_provider(&anvil).await;
        let sender = sender_provider.address();

        fill_purse(
            &contracts,
            1_000_000_000u64.into(),
            contracts.bridge_contract().address(),
        )
        .await;

        contracts
            .bridge_contract_by(sender_provider)
            .bridge_out(
                contracts.token_contract().address(),
                sender,
                500_000_000.into(),
                "HLLO".into(),
                "0x1234".into(),
            )
            .send()
            .await
            .expect_err("Expected to fail because the user is not an operator");
    }

    #[tokio::test]
    async fn transfer_out_not_operator() {
        /*
           Scenario:
           * Transfer some funds to the bridge purse
           * Request bridge out using not operator
           * Verify the call fails
        */

        let anvil = AnvilContext::default();

        let contracts = deploy_contracts(&anvil.provider).await;
        let sender_provider = guest_provider(&anvil).await;
        let sender = sender_provider.address();

        fill_purse(
            &contracts,
            1_000_000_000u64.into(),
            contracts.bridge_contract().address(),
        )
        .await;

        contracts
            .bridge_contract_by(sender_provider)
            .transfer_out(
                contracts.token_contract().address(),
                sender,
                500_000_000.into(),
            )
            .send()
            .await
            .expect_err("Expected to fail because the user is not an operator");
    }
}
