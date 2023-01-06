//! This module contains all the cryptographic primitives used by the bridge.

use crate::{error::CryptoError, types::NetworkType};

const AUTH_MESSAGE_BODY: &[u8] = b"Bridge Authentication Proof";

/// The module contains bridge authentification proof recovery for EVM networks.
pub mod eth {
    use k256::{ecdsa::recoverable::Id, elliptic_curve::sec1::ToEncodedPoint};
    use tiny_keccak::{Hasher, Keccak};

    use crate::crypto::AUTH_MESSAGE_BODY;
    use crate::error::CryptoError;

    const ETH_MESSAGE_PREFIX: &[u8] = b"\x19Ethereum Signed Message:\n";

    /// Recovers the public key from the signature.
    pub fn recover_key_from_auth_signature(
        signature: &[u8],
    ) -> Result<k256::PublicKey, CryptoError> {
        if signature.len() != 65 {
            return Err(CryptoError::InvalidSignatureFormat);
        }

        let recovery_id = match signature[64] {
            27 => 0,
            28 => 1,
            _ => return Err(CryptoError::InvalidSignatureFormat),
        };

        let recovery_id = Id::new(recovery_id).unwrap();

        let signature = k256::ecdsa::recoverable::Signature::new(
            &k256::ecdsa::Signature::try_from(&signature[..64])
                .map_err(|_| CryptoError::InvalidSignatureFormat)?,
            recovery_id,
        )
        .map_err(|_| CryptoError::InvalidSignatureFormat)?;

        let mut message = ETH_MESSAGE_PREFIX.to_vec();
        message.extend_from_slice(format!("{}", AUTH_MESSAGE_BODY.len()).as_bytes());
        message.extend_from_slice(AUTH_MESSAGE_BODY);
        let digest = keccak256(&message);

        let verifying_key = signature
            .recover_verify_key_from_digest_bytes(digest.as_ref().into())
            .map_err(|_| CryptoError::KeyRecoveryFailed)?;

        Ok(k256::PublicKey::from(verifying_key))
    }

    /// Hashes the message with keccak256.
    pub fn keccak256(bytes: &[u8]) -> [u8; 32] {
        let mut output = [0u8; 32];
        let mut hasher = Keccak::v256();
        hasher.update(bytes.as_ref());
        hasher.finalize(&mut output);
        output
    }

    /// Converts the public key to the address.
    pub fn address_from_public_key(public_key: k256::PublicKey) -> [u8; 20] {
        let data = public_key.to_encoded_point(false);
        let data = data.as_bytes();
        let hash = keccak256(&data[1..]);
        hash[12..].try_into().unwrap()
    }
}

/// The module contains bridge authentification proof recovery for Casper network.
/// The Casper network supports two types of public keys: Ed25519 and Secp256k1.
pub mod casper {
    use crate::crypto::AUTH_MESSAGE_BODY;
    use crate::{error::CryptoError, types::CASPER_TAG_ACCOUNT};
    use blake2::{
        digest::{Update, VariableOutput},
        VarBlake2b,
    };
    use k256::ecdsa::signature::Verifier;

    const CASPER_MESSAGE_PREFIX: &[u8] = b"Casper Message:\n";
    const MIN_PK_LENGTH: usize = 32;
    const MIN_SIG_LENGTH: usize = 64;

    const ED25519_TAG: u8 = 1;
    const SECP256K1_TAG: u8 = 2;

    #[derive(Clone, Copy)]
    pub enum PublicKey {
        Ed25519(ed25519_dalek::PublicKey),
        Secp256k1(k256::ecdsa::VerifyingKey),
    }

    #[derive(Clone, Copy)]
    pub enum Signature {
        Ed25519(ed25519_dalek::Signature),
        Secp256k1(k256::ecdsa::Signature),
    }

    impl PublicKey {
        /// Parsing the public key from bytes.
        pub fn from_bytes(bytes: &[u8]) -> Result<PublicKey, CryptoError> {
            if bytes.len() < MIN_PK_LENGTH {
                return Err(CryptoError::InvalidKeyFormat);
            }

            let tag = bytes[0];

            match tag {
                ED25519_TAG => Ok(PublicKey::Ed25519(
                    ed25519_dalek::PublicKey::from_bytes(&bytes[1..])
                        .map_err(|_| CryptoError::InvalidKeyFormat)?,
                )),
                SECP256K1_TAG => Ok(PublicKey::Secp256k1(
                    k256::ecdsa::VerifyingKey::from_sec1_bytes(&bytes[1..])
                        .map_err(|_| CryptoError::InvalidKeyFormat)?,
                )),
                _ => Err(CryptoError::InvalidKeyFormat),
            }
        }

        /// Converts the public key to the account hash.
        pub fn to_account_hash(&self) -> [u8; 32] {
            const ED25519_LOWERCASE: &str = "ed25519";
            const SECP256K1_LOWERCASE: &str = "secp256k1";

            let (algorithm_name, public_key_bytes): (_, Vec<u8>) = match self {
                PublicKey::Ed25519(pk) => (ED25519_LOWERCASE, pk.to_bytes().into()),
                PublicKey::Secp256k1(pk) => (SECP256K1_LOWERCASE, pk.to_bytes().to_vec()),
            };

            let preimage = {
                let mut data =
                    Vec::with_capacity(algorithm_name.len() + public_key_bytes.len() + 1);
                data.extend(algorithm_name.as_bytes());
                data.push(0);
                data.extend(public_key_bytes);
                data
            };

            blake2b(&preimage)
        }

        /// Converts the public key to the `casper_types::Key` bytes.
        pub fn to_key(&self) -> Vec<u8> {
            let mut data = vec![CASPER_TAG_ACCOUNT];
            data.extend_from_slice(&self.to_account_hash());
            data
        }
    }

    impl Signature {
        /// Parsing ed25519 signature from bytes.
        pub fn ed25519(bytes: &[u8]) -> Result<Signature, CryptoError> {
            Ok(Signature::Ed25519(
                ed25519_dalek::Signature::from_bytes(bytes)
                    .map_err(|_| CryptoError::InvalidSignatureFormat)?,
            ))
        }

        /// Parsing secp256k1 signature from bytes.
        pub fn secp256k1(bytes: &[u8]) -> Result<Signature, CryptoError> {
            Ok(Signature::Secp256k1(
                k256::ecdsa::Signature::try_from(bytes)
                    .map_err(|_| CryptoError::InvalidSignatureFormat)?,
            ))
        }

        /// Parsing the signature from bytes distinguishing the type of signature by the first byte.
        pub fn from_bytes(bytes: &[u8]) -> Result<Signature, CryptoError> {
            if bytes.len() < MIN_SIG_LENGTH {
                return Err(CryptoError::InvalidSignatureFormat);
            }

            let tag = bytes[0];

            match tag {
                ED25519_TAG => Self::ed25519(&bytes[1..]),
                SECP256K1_TAG => Self::secp256k1(&bytes[1..]),
                _ => Err(CryptoError::InvalidSignatureFormat),
            }
        }
    }

    /// Verifies the message signature. The public key and signature types must match.
    pub fn verify_message(
        message: &[u8],
        public_key: PublicKey,
        signature: Signature,
    ) -> Result<(), CryptoError> {
        match (public_key, signature) {
            (PublicKey::Ed25519(public_key), Signature::Ed25519(signature)) => public_key
                .verify_strict(message, &signature)
                .map_err(|_| CryptoError::VerificationFailed),
            (PublicKey::Secp256k1(public_key), Signature::Secp256k1(signature)) => public_key
                .verify(message, &signature)
                .map_err(|_| CryptoError::VerificationFailed),
            _ => Err(CryptoError::AlgorithmMismatch),
        }
    }

    /// Verifies bridge authentication proof message signature.
    pub fn verify_authentication(
        public_key: PublicKey,
        signature: Signature,
    ) -> Result<(), CryptoError> {
        let mut message = CASPER_MESSAGE_PREFIX.to_vec();
        message.extend_from_slice(AUTH_MESSAGE_BODY);
        verify_message(&message, public_key, signature)
    }

    /// Hashes the message using blake2b.
    fn blake2b(data: &[u8]) -> [u8; 32] {
        let mut result = [0; 32];
        // NOTE: Assumed safe as `BLAKE2B_DIGEST_LENGTH` is a valid value for a hasher
        let mut hasher = VarBlake2b::new(32).expect("should create hasher");

        hasher.update(data);
        hasher.finalize_variable(|slice| {
            result.copy_from_slice(slice);
        });
        result
    }
}

///
pub mod solana {
    use crate::crypto::AUTH_MESSAGE_BODY;
    use crate::error::CryptoError;

    /// Verifies bridge authentication proof message signature.
    pub fn verify_authentication(
        public_key: ed25519_dalek::PublicKey,
        signature: ed25519_dalek::Signature,
    ) -> Result<(), CryptoError> {
        public_key
            .verify_strict(AUTH_MESSAGE_BODY, &signature)
            .map_err(|_| CryptoError::VerificationFailed)
    }
}

/// Verify a signature (and optional public key) and return an address
/// The public key is optional for EVM signatures, but required for Casper signatures
pub fn verify_auth_signature(
    network_ty: NetworkType,
    signature: &[u8],
    public_key: Option<&[u8]>,
) -> Result<Vec<u8>, CryptoError> {
    match network_ty {
        NetworkType::Casper => {
            let public_key = if let Some(public_key) = public_key {
                casper::PublicKey::from_bytes(public_key)?
            } else {
                return Err(CryptoError::MissingPublicKey);
            };

            let signature = match public_key {
                casper::PublicKey::Ed25519(_) => casper::Signature::ed25519(signature)?,
                casper::PublicKey::Secp256k1(_) => casper::Signature::secp256k1(signature)?,
            };

            casper::verify_authentication(public_key, signature)?;

            Ok(public_key.to_key())
        }
        NetworkType::Evm => {
            let public_key = eth::recover_key_from_auth_signature(signature)?;

            Ok(eth::address_from_public_key(public_key).to_vec())
        }
        NetworkType::Solana => {
            let public_key_bytes = public_key.ok_or(CryptoError::MissingPublicKey)?;
            let public_key = ed25519_dalek::PublicKey::from_bytes(public_key_bytes)
                .map_err(|_| CryptoError::InvalidKeyFormat)?;
            let signature = ed25519_dalek::Signature::from_bytes(signature)
                .map_err(|_| CryptoError::InvalidSignatureFormat)?;
            solana::verify_authentication(public_key, signature)?;
            Ok(public_key_bytes.to_owned())
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn eth_sig() {
        let signature = base16::decode("d29bb47954dc2c0d67778507d9a96852bd0da75dce2337009fcce23a6dedb5625ad5541523ac3c2959c0d31b60b62b980a3c778fd903cedf9f17a99ba9d2152e1b").unwrap();

        let address = verify_auth_signature(NetworkType::Evm, &signature, None)
            .expect("signature must be valid");

        assert_eq!(
            "3095F955Da700b96215CFfC9Bc64AB2e69eB7DAB".to_lowercase(),
            base16::encode_lower(&address)
        )
    }

    #[test]
    fn casper_sig() {
        let signature = base16::decode("7088ef7cd32d4ff72a9877cdbdc11f91ea700f774e312e3a27359bd8a15e438200940aa680ea7bc673092721fdff5af689888c18be2128f1fa2da9d572035f83").unwrap();
        let pk =
            base16::decode("02026144f73f26ad533465d48d7dfebf69edb4996e07fb05cd9e61b840540e7992fe")
                .unwrap();

        let address = verify_auth_signature(NetworkType::Casper, &signature, Some(&pk))
            .expect("signature must be valid");

        assert_eq!(
            "002a58a625b26a456672b6e49c7468dab678c36dad115654a8d1676f5d18f019ee",
            base16::encode_lower(&address)
        )
    }

    #[test]
    fn solana_sig() {
        let signature = base16::decode("8e7bda89472cab7b1974be22fd550b6527997bb3c9c6058dff281434a8ec21e08c11dab0d96a6f11a99039283ca3054a1d93fab5d77449b710ae685d135a560c").unwrap();
        let pk = bs58::decode("9PmF2t7Fm2oBxiQLC8mRapZy2yqobbGmaqEo3QCDtR9o")
            .into_vec()
            .unwrap();

        let address = verify_auth_signature(NetworkType::Solana, &signature, Some(&pk))
            .expect("signature must be valid");

        assert_eq!(pk, address)
    }
}
