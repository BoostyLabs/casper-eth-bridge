use alloc::string::String;
use casper_contract::{contract_api::runtime::call_contract, unwrap_or_revert::UnwrapOrRevert};
use casper_types::{ContractHash, ContractPackageHash, Key, RuntimeArgs, U256};

use crate::entry_points::{
    EP_BRIDGE_IN_CONFIRM, PARAM_AMOUNT, PARAM_DESTINATION_ADDRESS, PARAM_DESTINATION_CHAIN,
    PARAM_SENDER, PARAM_TOKEN_CONTRACT,
};

pub fn bridge_in_confirm(
    bridge_contract: ContractHash,
    token_contract: ContractPackageHash,
    amount: U256,
    destination_chain: String,
    destination_address: String,
    from: Key,
) {
    call_contract::<()>(
        bridge_contract,
        EP_BRIDGE_IN_CONFIRM,
        RuntimeArgs::try_new(|args| {
            args.insert(PARAM_TOKEN_CONTRACT, token_contract)?;
            args.insert(PARAM_AMOUNT, amount)?;
            args.insert(PARAM_DESTINATION_CHAIN, destination_chain)?;
            args.insert(PARAM_DESTINATION_ADDRESS, destination_address)?;
            args.insert(PARAM_SENDER, from)?;
            Ok(())
        })
        .unwrap_or_revert(),
    );
}
