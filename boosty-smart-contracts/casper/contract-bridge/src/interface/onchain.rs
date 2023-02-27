use alloc::{string::String, vec::Vec};
use casper_contract::{contract_api::runtime::call_contract, unwrap_or_revert::UnwrapOrRevert};
use casper_types::{
    bytesrepr::Bytes, ContractHash, ContractPackageHash, Key, RuntimeArgs, U128, U256,
};

use crate::entry_points::{
    EP_BRIDGE_IN_CONFIRM, EP_CHECK_PARAMS, PARAM_AMOUNT, PARAM_BYTES, PARAM_DESTINATION_ADDRESS,
    PARAM_DESTINATION_CHAIN, PARAM_GAS_COMMISSION, PARAM_NONCE, PARAM_SENDER, PARAM_SIGNATURE,
    PARAM_TOKEN_CONTRACT,
};

pub fn bridge_in_confirm(
    bridge_contract: ContractHash,
    token_contract: ContractPackageHash,
    amount: U256,
    gas_commission: U256,
    nonce: U128,
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
            args.insert(PARAM_GAS_COMMISSION, gas_commission)?;
            args.insert(PARAM_NONCE, nonce)?;
            args.insert(PARAM_DESTINATION_CHAIN, destination_chain)?;
            args.insert(PARAM_DESTINATION_ADDRESS, destination_address)?;
            args.insert(PARAM_SENDER, from)?;
            Ok(())
        })
        .unwrap_or_revert(),
    );
}

pub fn check_params(bridge_contract: ContractHash, bytes: Vec<u8>, signature: [u8; 64], nonce: U128) {
    call_contract::<()>(
        bridge_contract,
        EP_CHECK_PARAMS,
        RuntimeArgs::try_new(|args| {
            args.insert(PARAM_BYTES, bytes)?;
            args.insert(PARAM_SIGNATURE, signature)?;
            args.insert(PARAM_NONCE, nonce)?;
            Ok(())
        })
        .unwrap_or_revert(),
    );
}
