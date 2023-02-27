// Copyright (C) 2022 Creditor Corp. Group.
// See LICENSE for copying information.

package signature_test

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tricorn/bridge/networks"
	"tricorn/pkg/signature"
)

func TestVerifyCasper(t *testing.T) {
	sig, err := hex.DecodeString("7088ef7cd32d4ff72a9877cdbdc11f91ea700f774e312e3a27359bd8a15e438200940aa680ea7bc673092721fdff5af689888c18be2128f1fa2da9d572035f83")
	require.NoError(t, err)

	msg := "Bridge Authentication Proof"

	t.Run("successful signature validation via algorithm secp256k1", func(t *testing.T) {
		publicKey, err := hex.DecodeString("02026144f73f26ad533465d48d7dfebf69edb4996e07fb05cd9e61b840540e7992fe")
		require.NoError(t, err)

		valid, err := signature.VerifyCasper(publicKey, msg, sig)
		require.NoError(t, err)
		require.True(t, valid)
	})

	t.Run("successful signature validation via algorithm ed25519", func(t *testing.T) {
		publicKey, err := hex.DecodeString("01eb6db16548f388fe35b542bccb2ba58284c99cb53d3fc8e8c596c7be1ba2146c")
		require.NoError(t, err)

		sig, err := hex.DecodeString("ff9534702440b8bb3224dde34f9f5b974a57fc206a2be76bb73d4e23e44ab1bef19c097e6609b94ea84c9856863c07efbfc052d862254b5a822d467fef63d806")
		require.NoError(t, err)

		valid, err := signature.VerifyCasper(publicKey, msg, sig)
		require.NoError(t, err)
		require.True(t, valid)
	})

	t.Run("empty public key error", func(t *testing.T) {
		var publicKey []byte
		_, err = signature.VerifyCasper(publicKey, msg, sig)
		require.Error(t, err)
		require.Equal(t, err, signature.ErrEmptyPublicKey)
	})

	t.Run("successful signature validation", func(t *testing.T) {
		publicKey, err := hex.DecodeString("32026144f73f26ad533465d48d7dfebf69edb4996e07fb05cd9e61b840540e7992fe")
		require.NoError(t, err)

		_, err = signature.VerifyCasper(publicKey, msg, sig)
		require.Error(t, err)
		require.Equal(t, err, signature.ErrInvalidPublicKeyAlgorithm)
	})
}

func TestAccountHashByPublicKey(t *testing.T) {
	t.Run("public key secp356k1 to account hash", func(t *testing.T) {
		publicKey, err := hex.DecodeString("0203c1253298f0617081edb618917c4109466b6cb734bae4bbb9b716b4c957f26e57")
		require.NoError(t, err)

		accountHash, err := signature.PublicKeyToAccountHash(publicKey)
		require.NoError(t, err)

		actual, err := networks.StringToBytes(networks.IDCasperTest, accountHash)
		require.NoError(t, err)

		expected, err := hex.DecodeString("463a154e75e6a5ba06e372e21f4ec12a6d6d89685286b160da4ab98d5a557fc4")
		require.NoError(t, err)

		assert.Equal(t, expected, actual)
	})

	t.Run("public key ed25519 to account hash", func(t *testing.T) {
		publicKey, err := hex.DecodeString("01eb6db16548f388fe35b542bccb2ba58284c99cb53d3fc8e8c596c7be1ba2146c")
		require.NoError(t, err)

		accountHash, err := signature.PublicKeyToAccountHash(publicKey)
		require.NoError(t, err)
		assert.NotEmpty(t, accountHash)

		actual, err := networks.StringToBytes(networks.IDCasperTest, accountHash)
		require.NoError(t, err)

		expected, err := hex.DecodeString("131b3843c4e5a2526229d158c253c5217adbfe6007a5b800be488e37a0ae9d58")
		require.NoError(t, err)

		assert.Equal(t, expected, actual)
	})
}
