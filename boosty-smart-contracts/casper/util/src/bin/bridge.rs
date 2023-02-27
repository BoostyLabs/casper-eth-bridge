use std::sync::Arc;

use anyhow::Context;
use casper_execution_engine::core::engine_state::executable_deploy_item::ContractIdentifier;
use casper_node::{
    event_stream_server::SseData, rpcs::state::GlobalStateIdentifier, types::Deploy,
};
use casper_types::{
    bytesrepr::{Bytes, FromBytes},
    ContractHash, DeployHash, ExecutionEffect, ExecutionResult, StoredValue, Transform, URef,
};
use casper_util::{
    client::CasperClient,
    util::{BridgeEnv, CommonEnv},
};
use event::BridgeEvent;
use futures::TryStreamExt;
use tokio;

#[tokio::main]
pub async fn main() -> anyhow::Result<()> {
    dotenv::dotenv().context("couldn't load .env file")?;
    let common_env: CommonEnv = envy::from_env().context("couldn't parse environment")?;
    let bridge_env: BridgeEnv = envy::from_env().context("couldn't parse environment")?;

    let client = Arc::new(common_env.make_client()?);

    event_watcher(client, bridge_env).await?;

    Ok(())
}

async fn event_watcher(client: Arc<CasperClient>, bridge_env: BridgeEnv) -> anyhow::Result<()> {
    let event_stream = client
        .event_stream_main()
        .await
        .context("couldn't setup event stream")?;

    let bridge_contract_hash = ContractHash::from_formatted_str(bridge_env.bridge_contract_hash())
        .expect("invalid bridge contract hash");

    let root_hash = client.get_state_root_hash().await?;
    let contract = client
        .query_global_state(
            GlobalStateIdentifier::StateRootHash(root_hash),
            bridge_contract_hash.into(),
            vec![],
        )
        .await?
        .stored_value;

    let event_trigger = if let StoredValue::Contract(contract) = contract {
        let key = contract
            .named_keys()
            .get("event_trigger")
            .expect("no event trigger found");
        *key.as_uref().expect("invalid event trigger type")
    } else {
        panic!("not a contract")
    }
    .remove_access_rights();

    println!("watching for events");

    event_stream
        .try_for_each_concurrent(None, |event| {
            let client = &client;
            async move {
                if let SseData::DeployProcessed { deploy_hash, .. } = event {
                    let deploy_hash = DeployHash::new(deploy_hash.inner().value());
                    let (deploy, execution_results) = client
                        .confirm_deploy(deploy_hash)
                        .await
                        .expect("failed to confirm deploy");

                    check_deploy_events(
                        deploy,
                        execution_results,
                        bridge_contract_hash,
                        event_trigger,
                    );
                }

                Ok(())
            }
        })
        .await
        .context("failed stream")?;

    Ok(())
}

fn check_deploy_events(
    deploy: Deploy,
    execution_results: Vec<ExecutionResult>,
    bridge_contract_hash: ContractHash,
    event_trigger: URef,
) {
    let mut valid_contract = false;
    if let Some(contract_identifier) = deploy.session().contract_identifier() {
        if let ContractIdentifier::Hash(hash) = contract_identifier {
            valid_contract = hash == bridge_contract_hash;
        }
    }

    if !valid_contract {
        return;
    }

    println!("caught event for contract {bridge_contract_hash}");

    for result in execution_results {
        match result {
            ExecutionResult::Failure { .. } => {}
            ExecutionResult::Success {
                effect: ExecutionEffect { transforms, .. },
                ..
            } => {
                for entry in transforms {
                    let key = entry.key;
                    let transform = entry.transform;
                    if event_trigger.to_formatted_string() == key {
                        if let Transform::WriteCLValue(value) = transform {
                            let data: Bytes = value.into_t().expect("deser failed");

                            let (event, _) = BridgeEvent::from_bytes(&data).expect("deser failed");
                            eprintln!("received event: {:#?}", event);
                        }
                    }
                }
            }
        }
    }
}

mod event {
    use casper_types::{
        bytesrepr::{self, FromBytes, ToBytes},
        ContractPackageHash, Key, U128, U256,
    };

    pub const BRIDGE_EVENT_FUNDS_IN_TAG: u8 = 0;
    pub const BRIDGE_EVENT_FUNDS_OUT_TAG: u8 = 1;

    #[derive(Debug)]
    pub enum BridgeEvent {
        FundsIn {
            token_contract: ContractPackageHash,
            destination_chain: String,
            destination_address: String,
            amount: U256,
            gas_commission: U256,
            stable_commission_percent: U256,
            nonce: U128,
            sender: Key,
        },
        FundsOut {
            token_contract: ContractPackageHash,
            source_chain: String,
            source_address: String,
            amount: U256,
            transaction_id: U256,
            recipient: Key,
        },
    }

    impl ToBytes for BridgeEvent {
        fn to_bytes(&self) -> Result<Vec<u8>, bytesrepr::Error> {
            let mut buffer = bytesrepr::allocate_buffer(self)?;

            match self {
                BridgeEvent::FundsIn {
                    token_contract,
                    destination_chain,
                    destination_address,
                    amount,
                    gas_commission,
                    stable_commission_percent,
                    nonce,
                    sender,
                } => {
                    buffer.push(BRIDGE_EVENT_FUNDS_IN_TAG);
                    buffer.extend(token_contract.to_bytes()?);
                    buffer.extend(destination_chain.to_bytes()?);
                    buffer.extend(destination_address.to_bytes()?);
                    buffer.extend(amount.to_bytes()?);
                    buffer.extend(gas_commission.to_bytes()?);
                    buffer.extend(stable_commission_percent.to_bytes()?);
                    buffer.extend(nonce.to_bytes()?);
                    buffer.extend(sender.to_bytes()?);
                }
                BridgeEvent::FundsOut {
                    token_contract,
                    source_chain,
                    source_address,
                    amount,
                    transaction_id,
                    recipient,
                } => {
                    buffer.push(BRIDGE_EVENT_FUNDS_OUT_TAG);
                    buffer.extend(token_contract.to_bytes()?);
                    buffer.extend(source_chain.to_bytes()?);
                    buffer.extend(source_address.to_bytes()?);
                    buffer.extend(amount.to_bytes()?);
                    buffer.extend(transaction_id.to_bytes()?);
                    buffer.extend(recipient.to_bytes()?);
                }
            }

            Ok(buffer)
        }

        fn serialized_length(&self) -> usize {
            match self {
                BridgeEvent::FundsIn {
                    token_contract,
                    destination_chain,
                    destination_address,
                    amount,
                    gas_commission,
                    stable_commission_percent,
                    nonce,
                    sender,
                } => {
                    destination_chain.serialized_length()
                        + destination_address.serialized_length()
                        + amount.serialized_length()
                        + gas_commission.serialized_length()
                        + stable_commission_percent.serialized_length()
                        + nonce.serialized_length()
                        + sender.serialized_length()
                        + token_contract.serialized_length()
                }
                BridgeEvent::FundsOut {
                    token_contract,
                    source_chain,
                    source_address,
                    amount,
                    transaction_id,
                    recipient,
                } => {
                    source_chain.serialized_length()
                        + source_address.serialized_length()
                        + amount.serialized_length()
                        + transaction_id.serialized_length()
                        + recipient.serialized_length()
                        + token_contract.serialized_length()
                }
            }
        }
    }

    impl FromBytes for BridgeEvent {
        fn from_bytes(bytes: &[u8]) -> Result<(Self, &[u8]), bytesrepr::Error> {
            let (tag, remainder) = u8::from_bytes(bytes)?;
            match tag {
                BRIDGE_EVENT_FUNDS_IN_TAG => {
                    let (token_contract, remainder) = ContractPackageHash::from_bytes(remainder)?;
                    let (destination_chain, remainder) = String::from_bytes(remainder)?;
                    let (destination_address, remainder) = String::from_bytes(remainder)?;
                    let (amount, remainder) = U256::from_bytes(remainder)?;
                    let (gas_commission, remainder) = U256::from_bytes(remainder)?;
                    let (stable_commission_percent, remainder) = U256::from_bytes(remainder)?;
                    let (nonce, remainder) = U128::from_bytes(remainder)?;
                    let (sender, remainder) = Key::from_bytes(remainder)?;
                    Ok((
                        BridgeEvent::FundsIn {
                            token_contract,
                            destination_chain,
                            destination_address,
                            amount,
                            gas_commission,
                            stable_commission_percent,
                            nonce,
                            sender,
                        },
                        remainder,
                    ))
                }
                BRIDGE_EVENT_FUNDS_OUT_TAG => {
                    let (token_contract, remainder) = ContractPackageHash::from_bytes(remainder)?;
                    let (source_chain, remainder) = String::from_bytes(remainder)?;
                    let (source_address, remainder) = String::from_bytes(remainder)?;
                    let (amount, remainder) = U256::from_bytes(remainder)?;
                    let (transaction_id, remainder) = U256::from_bytes(remainder)?;
                    let (recipient, remainder) = Key::from_bytes(remainder)?;
                    Ok((
                        BridgeEvent::FundsOut {
                            token_contract,
                            source_chain,
                            source_address,
                            amount,
                            transaction_id,
                            recipient,
                        },
                        remainder,
                    ))
                }
                _ => Err(bytesrepr::Error::Formatting),
            }
        }
    }
}
