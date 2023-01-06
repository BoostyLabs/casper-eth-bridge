use alloc::string::String;
use casper_contract::{
    contract_api::{
        runtime::{self, revert},
        storage,
    },
    unwrap_or_revert::UnwrapOrRevert,
};
use casper_types::{
    contracts::NamedKeys, system::CallStackElement, ContractPackageHash, EntryPoints, Key, U256,
};
use contract_util::{caller_context, current_contract, erc20, event::fire};

use crate::{entry_points, error::BridgeError, interface};
use casper_common::event::BridgeEvent;

pub const NK_ACCESS_UREF: &str = "bridge_contract_uref";
pub const NK_CONTRACT: &str = "bridge_contract";

pub const GROUP_OPERATOR: &str = "operator";

pub fn install() {
    let mut named_keys = NamedKeys::new();
    let mut entry_points = EntryPoints::new();

    entry_points.add_entry_point(entry_points::bridge_in());
    entry_points.add_entry_point(entry_points::bridge_in_confirm());
    entry_points.add_entry_point(entry_points::bridge_out());
    entry_points.add_entry_point(entry_points::transfer_out());

    contract_util::event::install(&mut named_keys);

    let (contract_package_hash, access_uref) = storage::create_contract_package_at_hash();
    let (contract_hash, _) =
        storage::add_contract_version(contract_package_hash, entry_points, named_keys);

    storage::create_contract_user_group(contract_package_hash, "operator", 0, [access_uref].into())
        .unwrap_or_revert();

    runtime::put_key(NK_ACCESS_UREF, access_uref.into());
    runtime::put_key(NK_CONTRACT, contract_hash.into());
}

fn verify_caller_is_self() {
    let caller_context = caller_context();
    let (_, current_contract_hash) = current_contract();

    match caller_context {
        CallStackElement::Session { .. } => revert(BridgeError::OnlyCallableBySelf),
        CallStackElement::StoredSession { contract_hash, .. }
        | CallStackElement::StoredContract { contract_hash, .. } => {
            if contract_hash != current_contract_hash {
                revert(BridgeError::OnlyCallableBySelf)
            }
        }
    }
}

pub fn bridge_in(
    token_contract: ContractPackageHash,
    amount: U256,
    destination_chain: String,
    destination_address: String,
) {
    let (self_contract_package, self_contract_hash) = current_contract();
    let context = caller_context();

    let self_contract_key: Key = (*self_contract_package).into();

    let from_key: Key = match context {
        CallStackElement::Session { account_hash }
        | CallStackElement::StoredSession { account_hash, .. } => (*account_hash).into(),
        CallStackElement::StoredContract {
            contract_package_hash,
            ..
        } => (*contract_package_hash).into(),
    };

    let balance_before = erc20::balance_of(token_contract, self_contract_key);
    erc20::transfer(token_contract, self_contract_key, amount);
    let balance_after = erc20::balance_of(token_contract, self_contract_key);

    if balance_after.checked_sub(balance_before) != Some(amount) {
        revert(BridgeError::UnexpectedTransferAmount)
    }

    interface::onchain::bridge_in_confirm(
        *self_contract_hash,
        token_contract,
        amount,
        destination_chain,
        destination_address,
        from_key,
    );
}

pub fn bridge_in_confirm(
    token_contract: ContractPackageHash,
    amount: U256,
    destination_chain: String,
    destination_address: String,
    sender: Key,
) {
    verify_caller_is_self();

    let event = BridgeEvent::FundsIn {
        token_contract,
        destination_chain,
        destination_address,
        amount,
        sender,
    };

    fire(event);
}

pub fn bridge_out(
    token_contract: ContractPackageHash,
    amount: U256,
    source_chain: String,
    source_address: String,
    recipient: Key,
) {
    let (self_contract_package, _) = current_contract();
    let self_contract_key: Key = (*self_contract_package).into();

    let balance_before = erc20::balance_of(token_contract, self_contract_key);
    erc20::transfer(token_contract, recipient, amount);
    let balance_after = erc20::balance_of(token_contract, self_contract_key);

    if balance_before.checked_sub(balance_after) != Some(amount) {
        revert(BridgeError::UnexpectedTransferAmount)
    }

    let event = BridgeEvent::FundsOut {
        token_contract,
        source_chain,
        source_address,
        amount,
        recipient,
    };

    fire(event);
}

pub fn transfer_out(token_contract: ContractPackageHash, amount: U256, recipient: Key) {
    let (self_contract_package, _) = current_contract();
    let self_contract_key: Key = (*self_contract_package).into();

    let balance_before = erc20::balance_of(token_contract, self_contract_key);
    erc20::transfer(token_contract, recipient, amount);
    let balance_after = erc20::balance_of(token_contract, self_contract_key);

    if balance_before.checked_sub(balance_after) != Some(amount) {
        revert(BridgeError::UnexpectedTransferAmount)
    }
}
