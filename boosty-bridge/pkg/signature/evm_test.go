// Copyright (C) 2022 Creditor Corp. Group.
// See LICENSE for copying information.

package signature_test

import (
	"log"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	"tricorn/bridge/networks"
	"tricorn/pkg/signature"
)

func TestRecoverPublicKeyFromSignature(t *testing.T) {
	msgToSign := "Bridge Authentication Proof"

	sig, err := networks.StringToBytes(networks.IDEth, "d29bb47954dc2c0d67778507d9a96852bd0da75dce2337009fcce23a6dedb5625ad5541523ac3c2959c0d31b60b62b980a3c778fd903cedf9f17a99ba9d2152e1b")
	require.NoError(t, err)

	addr, err := signature.RecoverEVMPublicKeyFrom(sig, msgToSign)
	require.NoError(t, err)
	log.Println(addr)
}

func TestPublicKeyToAddress(t *testing.T) {
	msgToSign := "Bridge Authentication Proof"

	sig, err := networks.StringToBytes(networks.IDEth, "d29bb47954dc2c0d67778507d9a96852bd0da75dce2337009fcce23a6dedb5625ad5541523ac3c2959c0d31b60b62b980a3c778fd903cedf9f17a99ba9d2152e1b")
	require.NoError(t, err)

	pubKeyHex, err := signature.RecoverEVMPublicKeyFrom(sig, msgToSign)
	require.NoError(t, err)

	address, err := signature.EVMPublicKeySecp256k1ToAddress(pubKeyHex)
	require.NoError(t, err)

	log.Println(address.String())

	isValidAddress := common.IsHexAddress(address.String())
	require.True(t, isValidAddress)
}
