// Copyright (C) 2022 Creditor Corp. Group.
// See LICENSE for copying information.

package signature

import (
	"errors"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// EthereumSignedMessage is prefix which used in eth signed messages.
const EthereumSignedMessage = "\x19Ethereum Signed Message:\n"

// RecoverEVMPublicKeyFrom recovers public key from signature by signature and data which was signer.
func RecoverEVMPublicKeyFrom(signature []byte, msgData string) ([]byte, error) {
	data := []byte(EthereumSignedMessage)
	data = append(data, []byte(strconv.Itoa(len([]byte(msgData))))...)
	data = append(data, []byte(msgData)...)
	dataHash := crypto.Keccak256Hash(data)

	if len(signature) != 65 {
		return nil, errors.New("invalid signature len, should be 65")
	}

	// transform yellow paper V from 27/28 to 0/1.
	if signature[64] == 27 || signature[64] == 28 {
		signature[64] -= 27
	}

	return crypto.Ecrecover(dataHash.Bytes(), signature)
}

// EVMPublicKeySecp256k1ToAddress retrieves wallet address from secp256k1 public key.
func EVMPublicKeySecp256k1ToAddress(publicKey []byte) (common.Address, error) {
	publicKeyECDSA, err := crypto.UnmarshalPubkey(publicKey)
	if err != nil {
		return common.Address{}, err
	}

	address := crypto.PubkeyToAddress(*publicKeyECDSA)
	return address, nil
}
