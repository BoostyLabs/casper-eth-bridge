// Copyright (C) 2022 Creditor Corp. Group.
// See LICENSE for copying information.

package signature

import "crypto/ed25519"

// VerifySolana verifies that given public key signed message and as a result received given signature in Solana network.
func VerifySolana(publicKeyBytes []byte, msgData string, signature []byte) bool {
	publicKey := ed25519.PublicKey(publicKeyBytes)
	return ed25519.Verify(publicKey, []byte(msgData), signature)
}
