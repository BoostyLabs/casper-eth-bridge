#![no_std]
#![no_main]

extern crate alloc;
extern crate contract_bridge;

use alloc::string::String;
use casper_contract::{contract_api::runtime, unwrap_or_revert::UnwrapOrRevert};
use casper_types::{bytesrepr::Bytes, CLValue, ContractPackageHash, Key, U128, U256};
use contract_bridge::entry_points::{
    PARAM_AMOUNT, PARAM_BYTES, PARAM_COMMISSION, PARAM_DEADLINE, PARAM_DESTINATION_ADDRESS,
    PARAM_DESTINATION_CHAIN, PARAM_GAS_COMMISSION, PARAM_NONCE, PARAM_RECIPIENT, PARAM_SENDER,
    PARAM_SIGNATURE, PARAM_SIGNER, PARAM_SOURCE_ADDRESS, PARAM_SOURCE_CHAIN,
    PARAM_STABLE_COMMISSION_PERCENT, PARAM_TOKEN_CONTRACT, PARAM_TRANSACTION_ID,
};

/// Transfers funds to the bridge, with metadata specifying the destination chain.
///
/// Call context: session
#[no_mangle]
pub extern "C" fn bridge_in() {
    let token_contract: ContractPackageHash = runtime::get_named_arg(PARAM_TOKEN_CONTRACT);
    let amount: U256 = runtime::get_named_arg(PARAM_AMOUNT);
    let gas_commission: U256 = runtime::get_named_arg(PARAM_GAS_COMMISSION);
    let deadline: U256 = runtime::get_named_arg(PARAM_DEADLINE);
    let nonce: U128 = runtime::get_named_arg(PARAM_NONCE);
    let destination_chain: String = runtime::get_named_arg(PARAM_DESTINATION_CHAIN);
    let destination_address: String = runtime::get_named_arg(PARAM_DESTINATION_ADDRESS);

    let signature: [u8; 64] = runtime::get_named_arg(PARAM_SIGNATURE);

    contract_bridge::contract::bridge_in(
        token_contract,
        amount,
        gas_commission,
        deadline,
        nonce,
        destination_chain,
        destination_address,
        signature,
    );
}

#[no_mangle]
pub extern "C" fn bridge_in_confirm() {
    let token_contract: ContractPackageHash = runtime::get_named_arg(PARAM_TOKEN_CONTRACT);
    let amount: U256 = runtime::get_named_arg(PARAM_AMOUNT);
    let gas_commission: U256 = runtime::get_named_arg(PARAM_GAS_COMMISSION);
    let nonce: U128 = runtime::get_named_arg(PARAM_NONCE);
    let destination_chain: String = runtime::get_named_arg(PARAM_DESTINATION_CHAIN);
    let destination_address: String = runtime::get_named_arg(PARAM_DESTINATION_ADDRESS);
    let from: Key = runtime::get_named_arg(PARAM_SENDER);

    contract_bridge::contract::bridge_in_confirm(
        token_contract,
        amount,
        gas_commission,
        nonce,
        destination_chain,
        destination_address,
        from,
    );
}

#[no_mangle]
pub extern "C" fn check_params() {
    let bytes: Bytes = runtime::get_named_arg(PARAM_BYTES);
    let signature: [u8; 64] = runtime::get_named_arg(PARAM_SIGNATURE);
    let nonce: U128 = runtime::get_named_arg(PARAM_NONCE);
    contract_bridge::contract::check_params(bytes, signature, nonce);
}

/// Transfers funds from the bridge.
///
/// Call context: contract
#[no_mangle]
pub extern "C" fn bridge_out() {
    let token_contract: ContractPackageHash = runtime::get_named_arg(PARAM_TOKEN_CONTRACT);
    let amount: U256 = runtime::get_named_arg(PARAM_AMOUNT);
    let transaction_id: U256 = runtime::get_named_arg(PARAM_TRANSACTION_ID);
    let source_chain: String = runtime::get_named_arg(PARAM_SOURCE_CHAIN);
    let source_address: String = runtime::get_named_arg(PARAM_SOURCE_ADDRESS);
    let recipient: Key = runtime::get_named_arg(PARAM_RECIPIENT);

    contract_bridge::contract::bridge_out(
        token_contract,
        amount,
        transaction_id,
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
    let commission: U256 = runtime::get_named_arg(PARAM_COMMISSION);
    let nonce: U128 = runtime::get_named_arg(PARAM_NONCE);
    let recipient: Key = runtime::get_named_arg(PARAM_RECIPIENT);
    let signature: [u8; 64] = runtime::get_named_arg(PARAM_SIGNATURE);

    contract_bridge::contract::transfer_out(
        token_contract,
        amount,
        commission,
        nonce,
        recipient,
        signature,
    );
}

/// Withdraw commission to the specified owner account
///
/// Call context: contract
#[no_mangle]
pub extern "C" fn withdraw_commission() {
    let token_contract: ContractPackageHash = runtime::get_named_arg(PARAM_TOKEN_CONTRACT);
    let amount: U256 = runtime::get_named_arg(PARAM_AMOUNT);
    let recipient: Key = runtime::get_named_arg(PARAM_RECIPIENT);
    contract_bridge::contract::withdraw_commission(token_contract, amount, recipient);
}

/// Manually set commission percent
///
/// Call context:
#[no_mangle]
pub extern "C" fn set_stable_commission_percent() {
    let stable_commission_percent: U256 = runtime::get_named_arg(PARAM_STABLE_COMMISSION_PERCENT);
    contract_bridge::contract::set_stable_commission_percent(stable_commission_percent);
}

/// Get commission percent
///
/// Call context:
#[no_mangle]
pub extern "C" fn get_stable_commission_percent() {
    let res = contract_bridge::contract::get_stable_commission_percent();
    runtime::ret(CLValue::from_t(res).unwrap_or_revert());
}

/// Manually set signer
///
/// Call context:
#[no_mangle]
pub extern "C" fn set_signer() {
    let signer: String = runtime::get_named_arg(PARAM_SIGNER);
    contract_bridge::contract::set_signer(signer);
}

/// Get signer
///
/// Call context:
#[no_mangle]
pub extern "C" fn get_signer() {
    let res = contract_bridge::contract::get_signer();
    runtime::ret(CLValue::from_t(res).unwrap_or_revert());
}

/// Get total commission
///
/// Call context:
#[no_mangle]
pub extern "C" fn get_total_commission() {
    let amount: U256 = runtime::get_named_arg(PARAM_AMOUNT);
    let gas_commission: U256 = runtime::get_named_arg(PARAM_GAS_COMMISSION);
    let res = contract_bridge::contract::get_total_commission(amount, gas_commission);
    runtime::ret(CLValue::from_t(res).unwrap_or_revert());
}

/// Get commission pool by token
///
/// Call context:
#[no_mangle]
pub extern "C" fn get_commission_by_token() {
    let token_contract: ContractPackageHash = runtime::get_named_arg(PARAM_TOKEN_CONTRACT);
    let res = contract_bridge::contract::get_commission_by_token(token_contract);
    runtime::ret(CLValue::from_t(res).unwrap_or_revert());
}

#[no_mangle]
pub extern "C" fn call() {
    let signer: String = runtime::get_named_arg(PARAM_SIGNER);
    contract_bridge::contract::install(signer);
}
