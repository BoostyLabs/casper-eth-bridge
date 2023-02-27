use casper_execution_engine::core::engine_state::ExecutableDeployItem;
use casper_types::{
    ContractHash, ContractPackageHash, DeployHash, Key, RuntimeArgs, U128, U256, U512,
};

use crate::{client::CasperClient, error::Error};

impl CasperClient {
    pub async fn bridge_in(
        &self,
        bridge_contract: ContractHash,
        token_contract: Key,
        amount: U256,
        gas_commission: U256,
        deadline: U256,
        nonce: U128,
        destination_chain: String,
        destination_address: String,
        signature: String,
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
                    args.insert("gas_commission", gas_commission)?;
                    args.insert("deadline", deadline)?;
                    args.insert("nonce", nonce)?;
                    args.insert("destination_chain", destination_chain)?;
                    args.insert("destination_address", destination_address)?;
                    args.insert("signature", signature)?;
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
        transaction_id: U256,
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
                    args.insert("transaction_id", transaction_id)?;
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

    pub async fn set_stable_commission_percent(
        &self,
        bridge_contract: ContractHash,
        stable_commission_percent: U256,
    ) -> Result<DeployHash, Error> {
        let deploy = self.make_simple_deploy(
            U512::one() * 1_000_000_000u64,
            ExecutableDeployItem::StoredContractByHash {
                hash: bridge_contract,
                entry_point: "set_stable_commission_percent".into(),
                args: RuntimeArgs::try_new(|args| {
                    args.insert("stable_commission_percent", stable_commission_percent)?;
                    Ok(())
                })
                .expect("args"),
            },
        )?;

        Ok(self.put_deploy(deploy).await?)
    }

    pub async fn set_signer(
        &self,
        bridge_contract: ContractHash,
        signer: String,
    ) -> Result<DeployHash, Error> {
        let deploy = self.make_simple_deploy(
            U512::one() * 1_000_000_000u64,
            ExecutableDeployItem::StoredContractByHash {
                hash: bridge_contract,
                entry_point: "set_signer".into(),
                args: RuntimeArgs::try_new(|args| {
                    args.insert("signer", signer)?;
                    Ok(())
                })
                .expect("args"),
            },
        )?;

        Ok(self.put_deploy(deploy).await?)
    }

    pub async fn get_stable_commission_percent(
        &self,
        bridge_contract: ContractHash,
    ) -> Result<DeployHash, Error> {
        let deploy = self.make_simple_deploy(
            U512::one() * 1_000_000_000u64, // vvvq do we need payment?
            ExecutableDeployItem::StoredContractByHash {
                hash: bridge_contract,
                entry_point: "get_stable_commission_percent".into(), // vvvq do we need this way ofn runtime args declaration?
                args: RuntimeArgs::try_new(|args| Ok(())).expect("args"),
            },
        )?;

        Ok(self.put_deploy(deploy).await?)
    }

    pub async fn get_signer(&self, bridge_contract: ContractHash) -> Result<DeployHash, Error> {
        let deploy = self.make_simple_deploy(
            U512::one() * 1_000_000_000u64, // vvvq do we need payment?
            ExecutableDeployItem::StoredContractByHash {
                hash: bridge_contract,
                entry_point: "get_signer".into(), // vvvq do we need this way ofn runtime args declaration?
                args: RuntimeArgs::try_new(|args| Ok(())).expect("args"),
            },
        )?;

        Ok(self.put_deploy(deploy).await?)
    }
}
