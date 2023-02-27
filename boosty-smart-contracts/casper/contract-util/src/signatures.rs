use alloc::vec::Vec;
use casper_types::{
    account::AccountHash,
    bytesrepr::{Bytes, ToBytes},
    ContractPackageHash, Key, U128, U256, ContractHash,
};

use k256::{
    ecdsa::{
        signature::Signer,
        signature::{Signature as SignatureTrait, Verifier},
        Signature, SigningKey, VerifyingKey,
    },
    pkcs8::DecodePublicKey,
    SecretKey as SignatureSecretKey,
};

pub fn cook_msg_bridge_in(
    bridge_hash: ContractHash,
    token_package_hash: ContractPackageHash,
    account_address: AccountHash,
    amount: U256,
    gas_commission: U256,
    deadline: U256,
    nonce: U128,
    destination_chain: &str,
    destination_address: &str,

) -> Vec<u8> {
    let prefix = "TRICORN_BRIDGE_IN";
    let amount = amount.to_bytes().unwrap();
    let gas_commission = gas_commission.to_bytes().unwrap();
    let deadline = deadline.to_bytes().unwrap();
    let nonce = nonce.to_bytes().unwrap();
    let destination_chain = destination_chain.as_bytes();
    let destination_address = destination_address.as_bytes();

    let mut bytes = Vec::new();
    bytes.extend_from_slice(prefix.as_bytes());
    bytes.extend_from_slice(bridge_hash.as_bytes());
    bytes.extend_from_slice(token_package_hash.as_bytes());
    bytes.extend_from_slice(account_address.as_bytes());
    bytes.extend_from_slice(&amount);
    bytes.extend_from_slice(&gas_commission);
    bytes.extend_from_slice(&deadline);
    bytes.extend_from_slice(&nonce);
    bytes.extend_from_slice(destination_chain);
    bytes.extend_from_slice(&destination_address);
    bytes
}

pub fn cook_msg_transfer_out(
    bridge_hash: ContractHash,
    token_package_hash: ContractPackageHash,
    recipient: Key,
    amount_to_transfer: U256,
    commission: U256,
    nonce: U128,
) -> Vec<u8> {
    let prefix = "TRICORN_TRANSFER_OUT";
    let amount = amount_to_transfer.to_bytes().unwrap();
    let commission = commission.to_bytes().unwrap();
    let nonce = nonce.to_bytes().unwrap();
    let recipient = recipient.to_bytes().unwrap();

    let mut bytes = Vec::new();
    bytes.extend_from_slice(prefix.as_bytes());
    bytes.extend_from_slice(bridge_hash.as_bytes());
    bytes.extend_from_slice(token_package_hash.as_bytes());
    bytes.extend_from_slice(&recipient);
    bytes.extend_from_slice(&amount);
    bytes.extend_from_slice(&commission);
    bytes.extend_from_slice(&nonce);
    bytes
}

pub fn check_public_key(signer: &str) {
    VerifyingKey::from_public_key_pem(signer).unwrap_or_else(|e| panic!("{e}"));
}

pub fn sign_data(bytes: &[u8], signer: &str) -> Signature {
    let se = SignatureSecretKey::from_sec1_pem(signer).unwrap();
    SigningKey::from(se).sign(bytes)
}

pub fn get_signature_bytes(bytes: &[u8], signer: &str) -> [u8; 64] {
    sign_data(bytes, signer).as_bytes()[0..64].try_into().unwrap()
}

/// Return whether signatures is correct
pub fn verify_signature(signer: &str, signature_bytes: &[u8; 64], bytes: &Bytes) -> bool {
    let verify_key = VerifyingKey::from_public_key_pem(signer).unwrap();

    let signature_bytes = signature_bytes.to_vec();

    let signature = Signature::from_bytes(&signature_bytes).unwrap();

    let res = verify_key.verify(bytes, &signature);

    res.is_ok()
}
