use alloc::{string::String, vec};
use casper_types::{
    CLType, CLTyped, ContractPackageHash, EntryPoint, EntryPointAccess, EntryPointType, Group, Key,
    Parameter, U256,
};

use crate::contract::GROUP_OPERATOR;

pub const EP_BRIDGE_IN: &str = "bridge_in";
pub const EP_BRIDGE_IN_CONFIRM: &str = "bridge_in_confirm";
pub const EP_BRIDGE_OUT: &str = "bridge_out";
pub const EP_TRANSFER_OUT: &str = "transfer_out";

pub const PARAM_TOKEN_CONTRACT: &str = "token_contract";
pub const PARAM_AMOUNT: &str = "amount";
pub const PARAM_DESTINATION_CHAIN: &str = "destination_chain";
pub const PARAM_DESTINATION_ADDRESS: &str = "destination_address";
pub const PARAM_SOURCE_CHAIN: &str = "source_chain";
pub const PARAM_SOURCE_ADDRESS: &str = "source_address";
pub const PARAM_SENDER: &str = "sender";
pub const PARAM_RECIPIENT: &str = "recipient";

fn operator_access() -> EntryPointAccess {
    EntryPointAccess::Groups(vec![Group::new(GROUP_OPERATOR)])
}

pub fn bridge_in() -> EntryPoint {
    EntryPoint::new(
        EP_BRIDGE_IN,
        vec![
            Parameter::new(PARAM_TOKEN_CONTRACT, ContractPackageHash::cl_type()),
            Parameter::new(PARAM_AMOUNT, U256::cl_type()),
            Parameter::new(PARAM_DESTINATION_CHAIN, String::cl_type()),
            Parameter::new(PARAM_DESTINATION_ADDRESS, String::cl_type()),
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
            Parameter::new(PARAM_DESTINATION_CHAIN, String::cl_type()),
            Parameter::new(PARAM_DESTINATION_ADDRESS, String::cl_type()),
            Parameter::new(PARAM_SENDER, Key::cl_type()),
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
            Parameter::new(PARAM_SOURCE_CHAIN, String::cl_type()),
            Parameter::new(PARAM_SOURCE_ADDRESS, String::cl_type()),
            Parameter::new(PARAM_RECIPIENT, Key::cl_type()),
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
            Parameter::new(PARAM_RECIPIENT, Key::cl_type()),
        ],
        CLType::Unit,
        operator_access(),
        EntryPointType::Contract,
    )
}
