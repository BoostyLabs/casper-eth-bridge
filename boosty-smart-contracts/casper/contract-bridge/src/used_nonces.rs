//! Implementation of used nonces.
use alloc::string::String;

use casper_contract::{contract_api::storage, unwrap_or_revert::UnwrapOrRevert};
use casper_types::{bytesrepr::ToBytes, URef, U128};

use crate::{constants::USED_NONCES_KEY_NAME, util};

/// Creates a dictionary item key for a dictionary item.
fn make_dictionary_item_key(nonce: U128) -> String {
    let preimage = nonce.to_bytes().unwrap_or_revert();
    base64::encode(preimage)
}

fn uref() -> URef {
    util::get_uref(USED_NONCES_KEY_NAME)
}

pub(crate) fn use_nonce(nonce: U128) {
    let dictionary_item_key = make_dictionary_item_key(nonce);
    storage::dictionary_put(uref(), &dictionary_item_key, true);
}

pub(crate) fn is_used_nonce(nonce: U128) -> bool {
    let dictionary_item_key = make_dictionary_item_key(nonce);

    storage::dictionary_get(uref(), &dictionary_item_key)
        .unwrap_or_revert()
        .unwrap_or_default()
}
