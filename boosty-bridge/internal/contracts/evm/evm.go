package evm

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"tricorn/signer"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/BoostyLabs/evmsignature"
)

const (
	// SignatureLength defines signature length in bytes.
	SignatureLength int = 65
)

// NewKeyedTransactorWithChainID is a utility method to easily create a transaction signer from a single private key.
func NewKeyedTransactorWithChainID(ctx context.Context, signerAddress common.Address, chainID *big.Int, sign func([]byte, signer.Type) ([]byte, error)) (*bind.TransactOpts, error) {
	if chainID == nil {
		return nil, bind.ErrNoChainID
	}

	latestSigner := types.LatestSignerForChainID(chainID)
	return &bind.TransactOpts{
		From: signerAddress,
		Signer: func(address common.Address, tx *types.Transaction) (*types.Transaction, error) {
			if address != signerAddress {
				return nil, bind.ErrNotAuthorized
			}

			signature, err := sign(latestSigner.Hash(tx).Bytes(), signer.TypeDTTransaction)
			if err != nil {
				return nil, err
			}

			return tx.WithSignature(latestSigner, signature)
		},
		Context: ctx,
	}, nil
}

// ToEthSignedMessageHash wraps the data into eth message.
func ToEthSignedMessageHash(hash []byte) []byte {
	return evmsignature.SignHash(hash)
}

// ToEVMSignature reforms last two byte of signature from 00, 01 to 1b, 1c.
func ToEVMSignature(signature []byte) ([]byte, error) {
	if len(signature) != SignatureLength {
		return nil, fmt.Errorf("signature length isn't %d bytes", SignatureLength)
	}

	if signature[SignatureLength-1] != byte(evmsignature.PrivateKeyVZero) &&
		signature[SignatureLength-1] != byte(evmsignature.PrivateKeyVOne) {
		return nil, errors.New("signature is wrong")
	}

	signature[SignatureLength-1] += byte(evmsignature.PrivateKeyVTwentySeven)

	return signature, nil
}
