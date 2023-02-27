use alloc::{string::String, vec};
use casper_types::{
    CLType, CLTyped, ContractPackageHash, EntryPoint, EntryPointAccess,
    EntryPointType, Group, Key, Parameter, U128, U256,
};

use crate::constants::GROUP_OPERATOR;

pub const EP_BRIDGE_IN: &str = "bridge_in";
pub const EP_BRIDGE_IN_CONFIRM: &str = "bridge_in_confirm";
pub const EP_CHECK_PARAMS: &str = "check_params";
pub const EP_BRIDGE_OUT: &str = "bridge_out";
pub const EP_TRANSFER_OUT: &str = "transfer_out";
pub const EP_WITHDRAW_COMMISSION: &str = "withdraw_commission";
pub const EP_SET_STABLE_COMMISSION_PERCENT: &str = "set_stable_commission_percent";
pub const EP_GET_STABLE_COMMISSION_PERCENT: &str = "get_stable_commission_percent";
pub const EP_SET_SIGNER: &str = "set_signer";
pub const EP_GET_SIGNER: &str = "get_signer";

pub const PARAM_TOKEN_CONTRACT: &str = "token_contract";
pub const PARAM_AMOUNT: &str = "amount";
pub const PARAM_COMMISSION: &str = "commission";
pub const PARAM_GAS_COMMISSION: &str = "gas_commission";
pub const PARAM_DEADLINE: &str = "deadline";
pub const PARAM_STABLE_COMMISSION_PERCENT: &str = "stable_commission_percent";
pub const PARAM_NONCE: &str = "nonce";
pub const PARAM_TRANSACTION_ID: &str = "transaction_id";
pub const PARAM_DESTINATION_CHAIN: &str = "destination_chain";
pub const PARAM_DESTINATION_ADDRESS: &str = "destination_address";
pub const PARAM_SOURCE_CHAIN: &str = "source_chain";
pub const PARAM_SOURCE_ADDRESS: &str = "source_address";
pub const PARAM_SENDER: &str = "sender";
pub const PARAM_RECIPIENT: &str = "recipient";
pub const PARAM_SIGNER: &str = "signer";
pub const PARAM_SIGNATURE: &str = "signature";
pub const PARAM_BYTES: &str = "bytes";

fn operator_access() -> EntryPointAccess {
    EntryPointAccess::Groups(vec![Group::new(GROUP_OPERATOR)])
}

pub fn bridge_in() -> EntryPoint {
    EntryPoint::new(
        EP_BRIDGE_IN,
        vec![
            Parameter::new(PARAM_TOKEN_CONTRACT, ContractPackageHash::cl_type()),
            Parameter::new(PARAM_AMOUNT, U256::cl_type()),
            Parameter::new(PARAM_GAS_COMMISSION, U256::cl_type()),
            Parameter::new(PARAM_DEADLINE, U256::cl_type()),
            Parameter::new(PARAM_NONCE, U128::cl_type()),
            Parameter::new(PARAM_DESTINATION_CHAIN, String::cl_type()),
            Parameter::new(PARAM_DESTINATION_ADDRESS, String::cl_type()),
            Parameter::new(PARAM_SIGNATURE, CLType::ByteArray(64)),
        ],
        CLType::Unit,
        EntryPointAccess::Public,
        EntryPointType::Session,
    )
}

pub fn bridge_in_confirm() -> EntryPoint {
    EntryPoint::new(
        EP_BRIDGE_IN_CONFIRM,
        vec![
            Parameter::new(PARAM_TOKEN_CONTRACT, ContractPackageHash::cl_type()),
            Parameter::new(PARAM_AMOUNT, U256::cl_type()),
            Parameter::new(PARAM_GAS_COMMISSION, U256::cl_type()),
            Parameter::new(PARAM_NONCE, U128::cl_type()),
            Parameter::new(PARAM_DESTINATION_CHAIN, String::cl_type()),
            Parameter::new(PARAM_DESTINATION_ADDRESS, String::cl_type()),
            Parameter::new(PARAM_SENDER, Key::cl_type()),
            Parameter::new(PARAM_GAS_COMMISSION, U256::cl_type()),
            Parameter::new(PARAM_DEADLINE, U256::cl_type()),
            Parameter::new(PARAM_NONCE, U256::cl_type()),
            Parameter::new(PARAM_SIGNATURE, CLType::ByteArray(64)),
        ],
        CLType::Unit,
        EntryPointAccess::Public,
        EntryPointType::Contract,
    )
}

pub fn check_params() -> EntryPoint {
    EntryPoint::new(
        EP_CHECK_PARAMS,
        vec![
            Parameter::new(PARAM_BYTES, String::cl_type()),
            Parameter::new(PARAM_SIGNER, String::cl_type()),
            Parameter::new(PARAM_SIGNATURE, CLType::ByteArray(64)),
            Parameter::new(PARAM_NONCE, U128::cl_type()),
        ],
        CLType::Unit,
        EntryPointAccess::Public,
        EntryPointType::Contract,
    )
}

pub fn bridge_out() -> EntryPoint {
    EntryPoint::new(
        EP_BRIDGE_OUT,
        vec![
            Parameter::new(PARAM_TOKEN_CONTRACT, ContractPackageHash::cl_type()),
            Parameter::new(PARAM_AMOUNT, U256::cl_type()),
            Parameter::new(PARAM_TRANSACTION_ID, U256::cl_type()),
            Parameter::new(PARAM_SOURCE_CHAIN, String::cl_type()),
            Parameter::new(PARAM_SOURCE_ADDRESS, String::cl_type()),
            Parameter::new(PARAM_RECIPIENT, Key::cl_type()),
            Parameter::new(PARAM_TRANSACTION_ID, U256::cl_type()),
        ],
        CLType::Unit,
        operator_access(),
        EntryPointType::Contract,
    )
}

pub fn transfer_out() -> EntryPoint {
    EntryPoint::new(
        EP_TRANSFER_OUT,
        vec![
            Parameter::new(PARAM_TOKEN_CONTRACT, ContractPackageHash::cl_type()),
            Parameter::new(PARAM_AMOUNT, U256::cl_type()),
            Parameter::new(PARAM_COMMISSION, U256::cl_type()),
            Parameter::new(PARAM_NONCE, U128::cl_type()),
            Parameter::new(PARAM_RECIPIENT, Key::cl_type()),
            Parameter::new(PARAM_SIGNATURE, CLType::ByteArray(64)),
        ],
        CLType::Unit,
        EntryPointAccess::Public,
        EntryPointType::Contract,
    )
}

pub fn withdraw_commission() -> EntryPoint {
    EntryPoint::new(
        EP_WITHDRAW_COMMISSION,
        vec![
            Parameter::new(PARAM_TOKEN_CONTRACT, ContractPackageHash::cl_type()),
            Parameter::new(PARAM_AMOUNT, U256::cl_type()),
            Parameter::new(PARAM_RECIPIENT, Key::cl_type()),
            Parameter::new(PARAM_COMMISSION, U256::cl_type()),
            Parameter::new(PARAM_NONCE, U256::cl_type()),
            Parameter::new(PARAM_SIGNATURE, CLType::ByteArray(64)),
        ],
        CLType::Unit,
        operator_access(),
        EntryPointType::Contract,
    )
}

pub fn set_stable_commission_percent() -> EntryPoint {
    EntryPoint::new(
        EP_SET_STABLE_COMMISSION_PERCENT,
        vec![Parameter::new(
            PARAM_STABLE_COMMISSION_PERCENT,
            U256::cl_type(),
        )],
        CLType::Unit,
        operator_access(),
        EntryPointType::Contract,
    )
}

pub fn get_stable_commission_percent() -> EntryPoint {
    EntryPoint::new(
        EP_GET_STABLE_COMMISSION_PERCENT,
        vec![],
        CLType::U256,
        EntryPointAccess::Public,
        EntryPointType::Contract,
    )
}

pub fn set_signer() -> EntryPoint {
    EntryPoint::new(
        EP_SET_SIGNER,
        vec![Parameter::new(PARAM_SIGNER, String::cl_type())],
        CLType::Unit,
        operator_access(),
        EntryPointType::Contract,
    )
}

pub fn get_signer() -> EntryPoint {
    EntryPoint::new(
        EP_GET_SIGNER,
        vec![],
        String::cl_type(),
        EntryPointAccess::Public,
        EntryPointType::Contract,
    )
}
