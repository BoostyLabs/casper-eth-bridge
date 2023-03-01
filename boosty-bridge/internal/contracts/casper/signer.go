package casper

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	casper_chain "tricorn/chains/casper"
	"tricorn/internal/reverse"
	"tricorn/pkg/hexutils"
	signature_lib "tricorn/pkg/signature"
	"tricorn/signer"
)

// Signer describes sign func to generate signature for transactions.
type Signer struct {
	sign func([]byte, signer.Type) ([]byte, error)
}

// NewSigner is constructor for Signer.
func NewSigner(sign func([]byte, signer.Type) ([]byte, error)) casper_chain.Signer {
	return &Signer{
		sign: sign,
	}
}

// GetBridgeInSignature generates signature for inbound bridge transaction.
func (s *Signer) GetBridgeInSignature(ctx context.Context, bridgeIn casper_chain.BridgeInSignature) ([]byte, error) {
	var data []byte

	prefix := hexutils.ToHexString(fmt.Sprintf("%x", bridgeIn.Prefix))
	prefixBytes, err := hex.DecodeString(prefix)
	if err != nil {
		return nil, err
	}

	destinationChain := hexutils.ToHexString(fmt.Sprintf("%x", bridgeIn.DestinationChain))
	destinationChainBytes, err := hex.DecodeString(destinationChain)
	if err != nil {
		return nil, err
	}

	destinationAddress := hexutils.ToHexString(fmt.Sprintf("%x", bridgeIn.DestinationAddress))
	destinationAddressBytes, err := hex.DecodeString(destinationAddress)
	if err != nil {
		return nil, err
	}

	data = append(data, prefixBytes...)
	data = append(data, bridgeIn.BridgeHash...)
	data = append(data, bridgeIn.TokenPackageHash...)
	data = append(data, bridgeIn.AccountAddress...)
	data = append(data, withLenBytes(reverse.Bytes(bridgeIn.Amount.Bytes()))...)
	data = append(data, withLenBytes(reverse.Bytes(bridgeIn.GasCommission.Bytes()))...)
	data = append(data, withLenBytes(reverse.Bytes(bridgeIn.Deadline.Bytes()))...)
	data = append(data, withLenBytes(reverse.Bytes(bridgeIn.Nonce.Bytes()))...)
	data = append(data, destinationChainBytes...)
	data = append(data, destinationAddressBytes...)

	hash := sha256.Sum256(data)
	signature, err := s.sign(hash[:], signer.TypeDTSignature)
	if err != nil {
		return nil, err
	}

	return signature_lib.WithoutV(signature), nil
}

// GetTransferOutSignature generates signature for outbound transfer transaction.
func (s *Signer) GetTransferOutSignature(ctx context.Context, transferOut casper_chain.TransferOutSignature) ([]byte, error) {
	var data []byte

	prefixChain := hexutils.ToHexString(fmt.Sprintf("%x", transferOut.Prefix))
	prefixChainBytes, err := hex.DecodeString(prefixChain)
	if err != nil {
		return nil, err
	}

	data = append(data, prefixChainBytes...)
	data = append(data, transferOut.BridgeHash...)
	data = append(data, transferOut.TokenPackageHash...)
	data = append(data, transferOut.AccountAddress...)
	transferOut.Recipient = append([]byte{0}, transferOut.Recipient...)
	data = append(data, transferOut.Recipient...)
	data = append(data, withLenBytes(reverse.Bytes(transferOut.Amount.Bytes()))...)
	data = append(data, withLenBytes(reverse.Bytes(transferOut.GasCommission.Bytes()))...)
	data = append(data, withLenBytes(reverse.Bytes(transferOut.Nonce.Bytes()))...)

	hash := sha256.Sum256(data)

	signature, err := s.sign(hash[:], signer.TypeDTSignature)
	if err != nil {
		return nil, err
	}

	return signature_lib.WithoutV(signature), nil
}
