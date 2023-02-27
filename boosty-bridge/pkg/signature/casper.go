// Copyright (C) 2022 Creditor Corp. Group.
// See LICENSE for copying information.

package signature

import (
	"crypto/ed25519"
	"encoding/hex"
	"errors"

	"github.com/casper-ecosystem/casper-golang-sdk/keypair"
	casper_ed25519 "github.com/casper-ecosystem/casper-golang-sdk/keypair/ed25519"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	"golang.org/x/crypto/blake2b"
)

var (
	// ErrEmptyPublicKey indicates that public key is empty.
	ErrEmptyPublicKey = errors.New("empty public key")
	// ErrInvalidPublicKeyAlgorithm indicates that public key has been created using unsupported algorithm.
	ErrInvalidPublicKeyAlgorithm = errors.New("public key created using unsupported algorithm")
)

// casperMessage defines prefix in messages in Casper network.
const casperMessage = "Casper Message:\n"

// tag defines list of all possible public key tag(prefix) in Casper network.
type tag byte

const (
	// tagED25519 defines ED25519 algorithm tag.
	tagED25519 tag = 1
	// tagSECP256K1 defines SECP256K1 algorithm tag.
	tagSECP256K1 tag = 2
)

// algorithm defines list of all possible public key generation algorithms in Casper network.
type algorithm int

const (
	// algorithmED25519 defines ED25519 algorithm.
	algorithmED25519 algorithm = 1
	// algorithmSECP256K1 defines SECP256K1 algorithm.
	algorithmSECP256K1 algorithm = 2
)

// Byte returns byte representation of tag.
func (t tag) Byte() byte {
	return byte(t)
}

// VerifyCasper verifies that given public key signed message and as a result received given signature in Casper network.
func VerifyCasper(publicKey []byte, msg string, signature []byte) (bool, error) {
	algorithm, err := getPublicKeyAlgorithm(publicKey)
	if err != nil {
		return false, err
	}

	switch algorithm {
	case algorithmED25519:
		return verifyED25519(publicKey, msg, signature), nil
	case algorithmSECP256K1:
		return verifySECP256K1(publicKey, msg, signature), nil
	default:
		return false, ErrInvalidPublicKeyAlgorithm
	}
}

// getPublicKeyAlgorithm returns the algorithm used to create given public key.
func getPublicKeyAlgorithm(key []byte) (algorithm, error) {
	if len(key) == 0 {
		return 0, ErrEmptyPublicKey
	}

	switch key[0] {
	case tagED25519.Byte():
		return algorithmED25519, nil
	case tagSECP256K1.Byte():
		return algorithmSECP256K1, nil
	default:
		return 0, ErrInvalidPublicKeyAlgorithm
	}
}

// verifyED25519 verifies signature using ED25519 algorithm.
func verifyED25519(publicKeyBytes []byte, msg string, signature []byte) bool {
	msgData := []byte(casperMessage)
	msgData = append(msgData, []byte(msg)...)

	// cut off tag of public key.
	publicKey := ed25519.PublicKey(publicKeyBytes[1:])
	return ed25519.Verify(publicKey, msgData, signature)
}

// verifySECP256K1 verifies signature using SECP256K1 algorithm.
func verifySECP256K1(publicKeyBytes []byte, msg string, signature []byte) bool {
	msgData := []byte(casperMessage)
	msgData = append(msgData, []byte(msg)...)

	// cut off tag of public key.
	publicKey := secp256k1.PubKey(publicKeyBytes[1:])
	return publicKey.VerifySignature(msgData, signature)
}

// WithoutV removes fragment V from the signature.
func WithoutV(signature []byte) []byte {
	if len(signature) == 64 {
		return signature
	}

	return signature[:len(signature)-1]
}

// PublicKeyToAccountHash converts public key to account hash.
func PublicKeyToAccountHash(pubKey []byte) (string, error) {
	tag, err := getPublicKeyAlgorithm(pubKey)
	if err != nil {
		return "", err
	}

	switch tag {
	case algorithmED25519:
		accountHash := casper_ed25519.AccountHash(pubKey[1:])
		return accountHash, nil
	case algorithmSECP256K1:
		var buffer = append([]byte(keypair.StrKeyTagSecp256k1), keypair.Separator)
		buffer = append(buffer, pubKey[1:]...)
		hash := blake2b.Sum256(buffer)

		accountHash := hex.EncodeToString(hash[:])
		return accountHash, nil
	}

	return "", nil
}
