#![no_std]
#![no_main]

extern crate alloc;
extern crate contract_bridge;

use alloc::string::String;
use casper_contract::contract_api::runtime::{self, revert};
use casper_types::{ApiError, ContractPackageHash, Key, U256};
use contract_bridge::entry_points::{
    PARAM_AMOUNT, PARAM_DESTINATION_ADDRESS, PARAM_DESTINATION_CHAIN, PARAM_RECIPIENT,
    PARAM_SENDER, PARAM_SOURCE_ADDRESS, PARAM_SOURCE_CHAIN, PARAM_TOKEN_CONTRACT,
};

/// Transfers funds to the bridge, with metadata specifying the destination chain.
///
/// Call context: session
#[no_mangle]
pub extern "C" fn bridge_in() {
    let token_contract: ContractPackageHash = runtime::get_named_arg(PARAM_TOKEN_CONTRACT);
    let amount: U256 = runtime::get_named_arg(PARAM_AMOUNT);
    let destination_chain: String = runtime::get_named_arg(PARAM_DESTINATION_CHAIN);
    let destination_address: String = runtime::get_named_arg(PARAM_DESTINATION_ADDRESS);

    contract_bridge::contract::bridge_in(
        token_contract,
        amount,
        destination_chain,
        destination_address,
    );
}

#[no_mangle]
pub extern "C" fn bridge_in_confirm() {
    let token_contract: ContractPackageHash = runtime::get_named_arg(PARAM_TOKEN_CONTRACT);
    let amount: U256 = runtime::get_named_arg(PARAM_AMOUNT);
    let destination_chain: String = runtime::get_named_arg(PARAM_DESTINATION_CHAIN);
    let destination_address: String = runtime::get_named_arg(PARAM_DESTINATION_ADDRESS);
    let from: Key = runtime::get_named_arg(PARAM_SENDER);

    contract_bridge::contract::bridge_in_confirm(
        token_contract,
        amount,
        destination_chain,
        destination_address,
        from,
    );
}

/// Transfers funds from the bridge.
///
/// Call context: contract
#[no_mangle]
pub extern "C" fn bridge_out() {
    let token_contract: ContractPackageHash = runtime::get_named_arg(PARAM_TOKEN_CONTRACT);
    let amount: U256 = runtime::get_named_arg(PARAM_AMOUNT);
    let source_chain: String = runtime::get_named_arg(PARAM_SOURCE_CHAIN);
    let source_address: String = runtime::get_named_arg(PARAM_SOURCE_ADDRESS);
    let recipient: Key = runtime::get_named_arg(PARAM_RECIPIENT);

    contract_bridge::contract::bridge_out(
        token_contract,
        amount,
        source_chain,
        source_address,
        recipient,
    );
}

/// Manually transfer funds from the. Intended for cancellations and manual fund movement operations.
///
/// Call context: contract
#[no_mangle]
pub extern "C" fn transfer_out() {
    let token_contract: ContractPackageHash = runtime::get_named_arg(PARAM_TOKEN_CONTRACT);
    let amount: U256 = runtime::get_named_arg(PARAM_AMOUNT);
    let recipient: Key = runtime::get_named_arg(PARAM_RECIPIENT);

    contract_bridge::contract::transfer_out(token_contract, amount, recipient);
}

#[no_mangle]
pub extern "C" fn call() {
    contract_bridge::contract::install();
}
