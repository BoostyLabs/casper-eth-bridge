//! Implementation of commissions.
use alloc::string::String;

use casper_contract::{contract_api::storage, unwrap_or_revert::UnwrapOrRevert};
use casper_types::{bytesrepr::ToBytes, ContractPackageHash, URef, U256};

use crate::{constants::COMMISSIONS_BY_TOKEN_KEY_NAME, error::BridgeError, util};

/// Creates a dictionary item key for a dictionary item.
fn make_dictionary_item_key(owner: ContractPackageHash) -> String {
    let preimage = owner.to_bytes().unwrap_or_revert();
    base64::encode(preimage)
}

fn uref() -> URef {
    util::get_uref(COMMISSIONS_BY_TOKEN_KEY_NAME)
}

/// Add commission for a specified token into a dictionary.
pub(crate) fn increase(token_contract_address: ContractPackageHash, amount: U256) {
    let new_commission = {
        let commission = read(token_contract_address);
        commission
            .checked_add(amount)
            .unwrap_or_revert_with(BridgeError::Overflow)
    };

    let dictionary_item_key = make_dictionary_item_key(token_contract_address);

    storage::dictionary_put(uref(), &dictionary_item_key, new_commission);
}

/// Substract commission from a specified token in a dictionary.
pub(crate) fn decrease(token_contract_address: ContractPackageHash, amount: U256) {
    let new_commission = {
        let commission = read(token_contract_address);
        commission
            .checked_sub(amount)
            .unwrap_or_revert_with(BridgeError::Underflow)
    };

    let dictionary_item_key = make_dictionary_item_key(token_contract_address);
    storage::dictionary_put(uref(), &dictionary_item_key, new_commission);
}

/// Reads token commission of a specified token contract.
///
/// If a given token does not have commissions, then a 0 is returned.
pub(crate) fn read(token_contract_address: ContractPackageHash) -> U256 {
    let dictionary_item_key = make_dictionary_item_key(token_contract_address);

    storage::dictionary_get(uref(), &dictionary_item_key)
        .unwrap_or_revert()
        .unwrap_or_default()
}
