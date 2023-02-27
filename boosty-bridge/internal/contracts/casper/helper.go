package casper

import "math/big"

// withLenBytes appends the length of an existing byte slice to the front.
func withLenBytes(bytes []byte) []byte {
	return append(big.NewInt(int64(len(bytes))).Bytes(), bytes...)
}
