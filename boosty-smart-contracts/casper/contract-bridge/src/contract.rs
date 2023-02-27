use alloc::string::String;

use casper_contract::{
    contract_api::{
        runtime::{self, revert},
        storage,
    },
    unwrap_or_revert::UnwrapOrRevert,
};
use casper_types::{
    account::AccountHash, bytesrepr::Bytes, contracts::NamedKeys, system::CallStackElement,
    ContractPackageHash, EntryPoints, Key, U128, U256,
};
use contract_util::{
    caller_context, current_contract, erc20,
    event::fire,
    signatures::{check_public_key, cook_msg_bridge_in, cook_msg_transfer_out},
};

use crate::{
    commissions,
    constants::{COMMISSIONS_BY_TOKEN_KEY_NAME, NK_ACCESS_UREF, NK_CONTRACT, USED_NONCES_KEY_NAME},
    entry_points::{self, PARAM_SIGNER, PARAM_STABLE_COMMISSION_PERCENT},
    error::BridgeError,
    interface, uref, used_nonces,
};
use casper_common::event::BridgeEvent;

pub fn install(signer: String) {
    let mut named_keys = NamedKeys::new();
    let default_percent = storage::new_uref(U256::one() * 3);
    let default_signer = storage::new_uref(signer);

    let used_nonces_uref = storage::new_dictionary(USED_NONCES_KEY_NAME).unwrap_or_revert();

    let used_nonces_dictionary_key = {
        runtime::remove_key(USED_NONCES_KEY_NAME);
        Key::from(used_nonces_uref)
    };

    named_keys.insert(
        String::from(USED_NONCES_KEY_NAME),
        used_nonces_dictionary_key,
    );

    let commissions_by_tokens_uref =
        storage::new_dictionary(COMMISSIONS_BY_TOKEN_KEY_NAME).unwrap_or_revert();

    let commissions_by_tokens_dictionary_key = { 
        runtime::remove_key(COMMISSIONS_BY_TOKEN_KEY_NAME);
        Key::from(commissions_by_tokens_uref) 
    };
    named_keys.insert(
        String::from(COMMISSIONS_BY_TOKEN_KEY_NAME),
        commissions_by_tokens_dictionary_key,
    );

    let default_percent_key_name = String::from(PARAM_STABLE_COMMISSION_PERCENT);
    named_keys.insert(default_percent_key_name, Key::URef(default_percent));

    let signer_key = String::from(PARAM_SIGNER);
    named_keys.insert(signer_key, Key::URef(default_signer));

    let mut entry_points = EntryPoints::new();

    entry_points.add_entry_point(entry_points::bridge_in());
    entry_points.add_entry_point(entry_points::bridge_in_confirm());
    entry_points.add_entry_point(entry_points::check_params());
    entry_points.add_entry_point(entry_points::bridge_out());
    entry_points.add_entry_point(entry_points::transfer_out());
    entry_points.add_entry_point(entry_points::withdraw_commission());
    entry_points.add_entry_point(entry_points::set_stable_commission_percent());
    entry_points.add_entry_point(entry_points::set_signer());
    entry_points.add_entry_point(entry_points::get_signer());
    entry_points.add_entry_point(entry_points::get_stable_commission_percent());

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

fn verify_signature(bytes: Bytes, signature: [u8; 64]) {
    let signer: String = uref::read(PARAM_SIGNER);

    if signer.is_empty() {
        revert(BridgeError::SignerIsNotEstablished);
    }

    let res = contract_util::signatures::verify_signature(&signer, &signature, &bytes);

    if !res {
        revert(BridgeError::InvalidSignature);
    }
}

fn verify_deadline(deadline: U256) {
    let current_time: u64 = runtime::get_blocktime().into();
    let deadline = U256::as_u64(&deadline);

    if current_time > deadline {
        revert(BridgeError::ExpiredSignature)
    }
}

fn verify_nonce(nonce: U128) {
    if used_nonces::is_used_nonce(nonce) {
        revert(BridgeError::AlreadyUsedSignature)
    }
}

fn from_keys() -> (Key, AccountHash) {
    let context = caller_context();

    let res: Key = match context {
        CallStackElement::Session { account_hash }
        | CallStackElement::StoredSession { account_hash, .. } => (*account_hash).into(),
        CallStackElement::StoredContract {
            contract_package_hash,
            ..
        } => (*contract_package_hash).into(),
    };
    (res, res.into_account().unwrap_or_revert())
}

pub fn get_stable_commission_percent() -> U256 {
    uref::read(PARAM_STABLE_COMMISSION_PERCENT)
}

pub fn set_stable_commission_percent(value: U256) {
    if value > U256::one() * 100 {
        revert(BridgeError::InvalidCommissionPercent)
    }

    uref::write(PARAM_STABLE_COMMISSION_PERCENT, value)
}

pub fn get_signer() -> String {
    uref::read(PARAM_SIGNER)
}

/// value - ecrecover compatible public key
pub fn set_signer(value: String) {
    check_public_key(&value);
    uref::write(PARAM_SIGNER, value)
}

pub fn get_commission_by_token(token_contract: ContractPackageHash) -> U256 {
    commissions::read(token_contract)
}

pub fn get_total_commission(amount: U256, gas_commission: U256) -> U256 {
    let stable_commission = amount * get_stable_commission_percent() / 100;
    stable_commission + gas_commission
}

pub fn bridge_in(
    token_contract: ContractPackageHash,
    amount: U256,
    gas_commission: U256,
    deadline: U256,
    nonce: U128,
    destination_chain: String,
    destination_address: String,
    signature: [u8; 64],
) {
    verify_deadline(deadline);

    let (self_contract_package, self_contract_hash) = current_contract();
    let (from_key, signer) = from_keys();
    let bytes = cook_msg_bridge_in(
        *self_contract_hash,
        token_contract,
        signer,
        amount,
        gas_commission,
        deadline,
        nonce,
        &destination_chain,
        &destination_address,
    );

    interface::onchain::check_params(*self_contract_hash, bytes, signature, nonce);

    let self_contract_key: Key = (*self_contract_package).into();
    let balance_before = erc20::balance_of(token_contract, self_contract_key);
    // vvvq, how casper understand from where we transfer it?
    erc20::transfer(token_contract, self_contract_key, amount);
    let balance_after = erc20::balance_of(token_contract, self_contract_key);

    if balance_after.checked_sub(balance_before) != Some(amount) {
        revert(BridgeError::UnexpectedTransferAmount)
    }

    interface::onchain::bridge_in_confirm(
        *self_contract_hash,
        token_contract,
        amount,
        gas_commission,
        nonce,
        destination_chain,
        destination_address,
        from_key,
    );
}

pub fn bridge_in_confirm(
    token_contract: ContractPackageHash,
    amount: U256,
    gas_commission: U256,
    nonce: U128,
    destination_chain: String,
    destination_address: String,
    sender: Key,
) {
    verify_caller_is_self();

    let total_commission = get_total_commission(amount, gas_commission);
    if total_commission > amount {
        revert(BridgeError::CommissionBiggerThanTransferredAmount)
    }

    commissions::increase(token_contract, total_commission);

    let event = BridgeEvent::FundsIn {
        token_contract,
        destination_chain,
        destination_address,
        amount,
        gas_commission,
        stable_commission_percent: get_stable_commission_percent(),
        nonce,
        sender,
    };

    fire(event);
}

pub fn check_params(bytes: Bytes, signature: [u8; 64], nonce: U128) {
    verify_caller_is_self();

    verify_nonce(nonce);

    verify_signature(bytes, signature);

    used_nonces::use_nonce(nonce);
}

pub fn bridge_out(
    token_contract: ContractPackageHash,
    amount: U256,
    transaction_id: U256,
    source_chain: String,
    source_address: String,
    recipient: Key,
) {
    let (self_contract_package, _) = current_contract();
    let self_contract_key: Key = (*self_contract_package).into();

    let balance_before = erc20::balance_of(token_contract, self_contract_key);

    let allowed_balance = balance_before
        .checked_sub(commissions::read(token_contract))
        .unwrap_or_revert();

    if amount > allowed_balance {
        revert(BridgeError::AmountExceedBridgePool);
    }

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
        transaction_id,
        recipient,
    };

    fire(event);
}

pub fn transfer_out(
    token_contract: ContractPackageHash,
    amount: U256,
    commission: U256,
    nonce: U128,
    recipient: Key,
    signature: [u8; 64],
) {
    let (_, self_contract_hash) = current_contract();

    let (_, signer) = from_keys();

    let bytes = cook_msg_transfer_out(
        *self_contract_hash,
        token_contract,
        recipient,
        amount,
        commission,
        nonce,
    );
    interface::onchain::check_params(*self_contract_hash, bytes, signature, nonce);

    used_nonces::use_nonce(nonce);

    let (self_contract_package, _) = current_contract();
    let self_contract_key: Key = (*self_contract_package).into();

    let balance_before = erc20::balance_of(token_contract, self_contract_key);

    commissions::decrease(token_contract, commission);

    let total_sum_for_transfer = amount.checked_add(commission).unwrap_or_revert();

    erc20::transfer(token_contract, recipient, total_sum_for_transfer);

    let balance_after = erc20::balance_of(token_contract, self_contract_key);

    let actually_transferred = balance_before.checked_sub(balance_after).unwrap();
    if actually_transferred != total_sum_for_transfer {
        revert(BridgeError::UnexpectedTransferAmount)
    }

    let event = BridgeEvent::TransferOut {
        token_contract,
        total_sum_for_transfer,
        nonce,
        recipient,
    };

    fire(event);
}

pub fn withdraw_commission(token_contract: ContractPackageHash, amount: U256, recipient: Key) {
    if commissions::read(token_contract) < amount {
        revert(BridgeError::AmountExceedCommissionPool)
    }
    commissions::decrease(token_contract, amount);
    erc20::transfer(token_contract, recipient, amount);

    let event = BridgeEvent::WithdrawCommission {
        token_contract,
        amount,
    };
    fire(event);
}
