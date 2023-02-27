use casper_contract::{
    contract_api::storage::{self, new_uref},
    unwrap_or_revert::UnwrapOrRevert,
};
use casper_types::{
    bytesrepr::{Bytes, FromBytes, ToBytes},
    contracts::NamedKeys,
    Key, URef,
};

pub const EVENT_TRIGGER_UREF_NAME: &str = "event_trigger";

mod trigger_cache {
    use super::*;

    use crate::st_static;
    use once_cell::unsync::Lazy;

    st_static!(
        Lazy<casper_types::URef>,
        Lazy::new(
            || casper_contract::contract_api::runtime::get_key(EVENT_TRIGGER_UREF_NAME)
                .and_then(|key| key.into_uref())
                .unwrap_or_revert()
        )
    );
}

pub fn install(named_keys: &mut NamedKeys) {
    let event_trigger_uref = new_uref(Bytes::new());

    named_keys.insert(
        EVENT_TRIGGER_UREF_NAME.into(),
        Key::URef(event_trigger_uref),
    );
}

pub fn trigger_uref() -> URef {
    **trigger_cache::get()
}

pub fn fire<T: ContractEvent>(event: T) {
    storage::write(trigger_uref(), event.to_bytes().unwrap_or_revert())
}

pub trait ContractEvent: ToBytes + FromBytes {}
