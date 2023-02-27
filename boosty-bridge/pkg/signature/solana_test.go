// Copyright (C) 2022 Creditor Corp. Group.
// See LICENSE for copying information.

package signature_test

import (
	"encoding/hex"
	"testing"

	"github.com/mr-tron/base58/base58"
	"github.com/stretchr/testify/require"

	"tricorn/pkg/signature"
)

func TestVerifySolana(t *testing.T) {
	sig, err := hex.DecodeString("8e7bda89472cab7b1974be22fd550b6527997bb3c9c6058dff281434a8ec21e08c11dab0d96a6f11a99039283ca3054a1d93fab5d77449b710ae685d135a560c")
	require.NoError(t, err)

	publicKey, err := base58.Decode("9PmF2t7Fm2oBxiQLC8mRapZy2yqobbGmaqEo3QCDtR9o")
	require.NoError(t, err)

	msg := "Bridge Authentication Proof"

	valid := signature.VerifySolana(publicKey, msg, sig)
	require.True(t, valid)
}
