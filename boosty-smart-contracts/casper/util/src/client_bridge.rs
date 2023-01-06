use casper_execution_engine::core::engine_state::ExecutableDeployItem;
use casper_types::{ContractHash, ContractPackageHash, DeployHash, Key, RuntimeArgs, U256, U512};

use crate::{client::CasperClient, error::Error};

impl CasperClient {
    pub async fn bridge_in(
        &self,
        bridge_contract: ContractHash,
        token_contract: Key,
        amount: U256,
        destination_chain: String,
        destination_address: String,
    ) -> Result<DeployHash, Error> {
        let token_contract = ContractPackageHash::new(token_contract.into_hash().unwrap());

        let deploy = self.make_simple_deploy(
            U512::one() * 1_000_000_000u64,
            ExecutableDeployItem::StoredContractByHash {
                hash: bridge_contract,
                entry_point: "bridge_in".into(),
                args: RuntimeArgs::try_new(|args| {
                    args.insert("token_contract", token_contract)?;
                    args.insert("amount", amount)?;
                    args.insert("destination_chain", destination_chain)?;
                    args.insert("destination_address", destination_address)?;
                    Ok(())
                })
                .expect("args"),
            },
        )?;

        Ok(self.put_deploy(deploy).await?)
    }

    pub async fn bridge_out(
        &self,
        bridge_contract: ContractHash,
        token_contract: Key,
        amount: U256,
        recipient: Key,
        source_chain: String,
        source_address: String,
    ) -> Result<DeployHash, Error> {
        let token_contract = ContractPackageHash::new(token_contract.into_hash().unwrap());

        let deploy = self.make_simple_deploy(
            U512::one() * 1_000_000_000u64,
            ExecutableDeployItem::StoredContractByHash {
                hash: bridge_contract,
                entry_point: "bridge_out".into(),
                args: RuntimeArgs::try_new(|args| {
                    args.insert("token_contract", token_contract)?;
                    args.insert("amount", amount)?;
                    args.insert("source_chain", source_chain)?;
                    args.insert("source_address", source_address)?;
                    args.insert("recipient", recipient)?;
                    Ok(())
                })
                .expect("args"),
            },
        )?;

        Ok(self.put_deploy(deploy).await?)
    }
}
