#![cfg_attr(not(std), no_std)]

#[cfg(all(not(wasm), onchain))]
compile_error!("`onchain` feature only supported on wasm target");

#[cfg(all(std, wasm))]
compile_error!("`std` feature is not supported on wasm target");

extern crate alloc;

use alloc::vec::Vec;
use casper_contract::{
    contract_api::runtime::{self, revert},
    unwrap_or_revert::UnwrapOrRevert,
};
use casper_types::{
    account::AccountHash, system::CallStackElement, ContractHash, ContractPackageHash,
};
use error::UtilError;
use once_cell::unsync::OnceCell;

pub mod erc20;
pub mod error;
pub mod event;

/// Create a non-Sync static variable accessible safely in a wasm target,
#[macro_export]
macro_rules! st_static {
    ($t:ty, $init:expr) => {
        #[cfg(onchain)]
        static mut GLOBAL: $t = $init;

        #[cfg(onchain)]
        pub(super) fn get() -> &'static $t {
            unsafe { &GLOBAL }
        }

        #[cfg(not(onchain))]
        pub(super) fn get() -> &'static $t {
            #[allow(unused)]
            const _: $t = $init;
            panic!("cannot access non-Sync static in a non-wasm environment");
        }
    };
}

#[derive(Default)]
struct UtilCache {
    call_stack: OnceCell<Vec<CallStackElement>>,
}

impl UtilCache {
    pub const fn new() -> Self {
        Self {
            call_stack: OnceCell::new(),
        }
    }
}

mod cache {
    use super::UtilCache;

    st_static!(UtilCache, UtilCache::new());
}

fn get_cache() -> &'static UtilCache {
    cache::get()
}

/// Equivalent of [`runtime::get_call_stack`], but cached across invocations.
pub fn call_stack() -> &'static [CallStackElement] {
    get_cache()
        .call_stack
        .get_or_init(|| runtime::get_call_stack())
}

/// Get a call stack element at `depth`
pub fn call_stack_elem(depth: usize) -> Option<&'static CallStackElement> {
    let call_stack = call_stack();
    if depth >= call_stack.len() {
        None
    } else {
        call_stack.get(call_stack.len() - depth - 1)
    }
}

/// Return the context of the immediate caller of the current context.
///
/// Reverts with [`UtilError::InvalidStackDepth`] if there is no immediate caller.
pub fn caller_context() -> &'static CallStackElement {
    call_stack_elem(1).unwrap_or_revert_with(UtilError::InvalidStackDepth)
}

/// Return the current context.
pub fn current_context() -> &'static CallStackElement {
    // this is infallible
    call_stack_elem(0).unwrap_or_revert()
}

/// Return the current contract's package hash and contract hash.
///
/// Reverts with [`UtilError::CurrentContextNotContract`] if the current context doesn't reference a contract.
pub fn current_contract() -> (&'static ContractPackageHash, &'static ContractHash) {
    match current_context() {
        CallStackElement::Session { .. } => revert(UtilError::CurrentContextNotContract),
        CallStackElement::StoredSession {
            contract_package_hash,
            contract_hash,
            ..
        } => (contract_package_hash, contract_hash),
        CallStackElement::StoredContract {
            contract_package_hash,
            contract_hash,
        } => (contract_package_hash, contract_hash),
    }
}

/// Return the current session's account hash.
///
/// Reverts with [`UtilError::CurrentContextNotSession`] if the current context doesn't reference a session.
pub fn current_session() -> &'static AccountHash {
    match current_context() {
        CallStackElement::Session { account_hash }
        | CallStackElement::StoredSession { account_hash, .. } => account_hash,
        CallStackElement::StoredContract { .. } => revert(UtilError::CurrentContextNotSession),
    }
}
