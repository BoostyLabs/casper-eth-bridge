use std::time::Duration;

use casper_execution_engine::core::engine_state::ExecutableDeployItem;
use casper_node::types::{Deploy, TimeDiff, Timestamp};
use casper_types::{
    bytesrepr::Bytes, CLValue, ContractHash, DeployHash, ExecutionResult, Key, RuntimeArgs, U256,
    U512,
};
use once_cell::sync::Lazy;

use crate::error::Error;
use crate::{client::CasperClient, util::erc20_dictionary_key};

/// Cost of a standard ERC20 transfer deploy: ~0.15 CSPR
pub const ERC20_DEPLOY_TRANSFER_COST: Lazy<U512> = Lazy::new(|| U512::one() * 150_000_000u64);
/// Approx. cost of a ERC20 contract deploy: ~60 CSPR
pub const ERC20_DEPLOY_CONTRACT_COST: Lazy<U512> = Lazy::new(|| U512::one() * 60_000_000_000u64);

fn simple_payment(amount: U512) -> ExecutableDeployItem {
    let mut args = RuntimeArgs::new();
    args.insert_cl_value("amount", CLValue::from_t(amount).expect("infallible"));

    ExecutableDeployItem::ModuleBytes {
        module_bytes: Bytes::new(),
        args: args,
    }
}

impl CasperClient {
    pub fn make_simple_deploy(
        &self,
        payment: U512,
        session: ExecutableDeployItem,
    ) -> Result<Deploy, Error> {
        let config = self.config();

        Ok(Deploy::new(
            Timestamp::now(),
            TimeDiff::from_seconds(60 * 60),
            1,
            vec![],
            config.chain_name.clone(),
            simple_payment(payment),
            session,
            config.main_secret()?,
            Some(config.main_public()?),
        ))
    }

    pub fn key_from_str(&self, key: &str) -> Result<Key, Error> {
        if key == "self" {
            let config = self.config();

            config.main_key()
        } else {
            Key::from_formatted_str(key).map_err(|_| Error::InvalidKeyFormat { given: key.into() })
        }
    }

    pub async fn put_contract(
        &self,
        code: Vec<u8>,
        deploy_args: RuntimeArgs,
    ) -> Result<DeployHash, Error> {
        let deploy = self.make_simple_deploy(
            *ERC20_DEPLOY_CONTRACT_COST,
            ExecutableDeployItem::ModuleBytes {
                module_bytes: Bytes::from(code),
                args: deploy_args,
            },
        )?;

        self.put_deploy(deploy).await
    }

    pub async fn erc20_query_balance(&self, contract: Key, who: Key) -> Result<CLValue, Error> {
        let state_root_hash = self.get_state_root_hash().await?;

        let dictionary_item = self
            .get_dictionary_item(
                state_root_hash,
                casper_node::rpcs::state::DictionaryIdentifier::ContractNamedKey {
                    key: contract.to_formatted_string(),
                    dictionary_name: "balances".into(),
                    dictionary_item_key: erc20_dictionary_key(&who),
                },
            )
            .await?
            .stored_value;

        let cl_value = dictionary_item
            .as_cl_value()
            .ok_or_else(|| Error::UnexpectedStoredValueType {
                expected: "ClValue".into(),
                got: dictionary_item.type_name(),
            })?
            .clone();

        Ok(cl_value)
    }

    // assumes default key as source
    pub async fn erc20_deploy_transfer(
        &self,
        contract: Key,
        to: Key,
        amount: U256,
    ) -> Result<DeployHash, Error> {
        let contract_hash = ContractHash::new(contract.into_hash().expect("key must be hashaddr"));

        let deploy = self.make_simple_deploy(
            *ERC20_DEPLOY_TRANSFER_COST,
            ExecutableDeployItem::StoredContractByHash {
                hash: contract_hash,
                entry_point: "transfer".into(),
                args: RuntimeArgs::try_new(|args| {
                    args.insert("recipient", to)?;
                    args.insert("amount", amount)?;
                    Ok(())
                })
                .expect("infallible"),
            },
        )?;

        self.put_deploy(deploy).await
    }

    pub async fn erc20_deploy_get_balance(
        &self,
        contract: Key,
        who: Key,
    ) -> Result<DeployHash, Error> {
        let contract_hash = ContractHash::new(contract.into_hash().expect("key must be hashaddr"));

        let deploy = self.make_simple_deploy(
            *ERC20_DEPLOY_TRANSFER_COST,
            ExecutableDeployItem::StoredContractByHash {
                hash: contract_hash,
                entry_point: "balance_of".into(),
                args: RuntimeArgs::try_new(|args| {
                    args.insert("address", who)?;
                    Ok(())
                })
                .expect("infallible"),
            },
        )?;

        self.put_deploy(deploy).await
    }

    pub async fn confirm_deploy(
        &self,
        deploy_hash: DeployHash,
    ) -> Result<(Deploy, Vec<ExecutionResult>), Error> {
        loop {
            let (deploy, execution_results) = self.get_deploy(deploy_hash).await?;

            if execution_results.is_empty() {
                tokio::time::sleep(Duration::from_secs(2)).await;

                continue;
            }

            return Ok((deploy, execution_results));
        }
    }
}
