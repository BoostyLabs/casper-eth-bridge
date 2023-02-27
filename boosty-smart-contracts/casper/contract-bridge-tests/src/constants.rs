#![allow(non_snake_case)]
use casper_types::{U128, U256};
pub const TEST_BLOCK_TIME: u64 = 1672071121;
pub(crate) fn TEST_AMOUNT() -> U256 {
    U256::one() * 1_000_000_000_000u64
}
pub(crate) fn TEST_GAS_COMMISSION() -> U256 {
    U256::one() * 1000
}
pub(crate) fn TEST_STABLE_COMMISSION_PERCENT() -> U256 {
    U256::one() * 3
}
pub(crate) fn TEST_NONCE() -> U128 {
    U128::one() * 555
}
pub(crate) fn TEST_COMMISSION_PERCENT() -> U256 {
    U256::one() * 54
}
pub(crate) fn TEST_CORRECT_DEADLINE() -> U256 {
    U256::one() * 1672943628
}
pub(crate) fn TEST_EXPIRED_DEADLINE() -> U256 {
    U256::one() * 1672051121
}
pub(crate) fn TEST_DESTINATION_CHAIN() -> String {
    "DEST".to_string()
}
pub(crate) fn TEST_DESTINATION_ADDRESS() -> String {
    "DESTADDR".to_string()
}

pub(crate) const TEST_ACCOUNT_BALANCE: u64 = 10_000_000_000_000u64;

pub(crate) const ERC20_INSUFFIENT_BALANCE_ERROR_CODE: u16 = u16::MAX - 1;
pub(crate) const TEST_ACCOUNT: [u8; 32] = [255u8; 32];

pub(crate) const TEST_PREFIX_BRIDGE_IN: &str = "TRICORN_BRIDGE_IN";
pub(crate) const TEST_PREFIX_TRANSFER_OUT: &str = "TRICORN_TRANSFER_OUT";
