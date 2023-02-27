// Copyright (C) 2022 Creditor Corp. Group.
// See LICENSE for copying information.

package signer

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"errors"
	"fmt"

	casper_ed25519 "github.com/casper-ecosystem/casper-golang-sdk/keypair/ed25519"
	"github.com/ethereum/go-ethereum/crypto"
	solana_types "github.com/portto/solana-go-sdk/types"
	"github.com/zeebo/errs"

	"tricorn/bridge/networks"
)

// ErrSigner indicates that there was an error in the service.
var ErrSigner = errs.Class("signer service")

// Service is handling signer related logic.
//
// architecture: Service
type Service struct {
	config   Config
	keyStore KeyStore
}

// Secp256k1PrivateKeyLength indicates that the length of private key is equal to 32 bytes.
const Secp256k1PrivateKeyLength = 32

// NewService is constructor for Service.
func NewService(config Config, keyStore KeyStore) *Service {
	return &Service{
		config:   config,
		keyStore: keyStore,
	}
}

// Sign creates and returns signature from data.
func (s *Service) Sign(ctx context.Context, networkType networks.Type, data []byte, dataType Type) ([]byte, error) {
	var signature []byte

	privateKeyHex, err := s.keyStore.Get(ctx, networkType, dataType)
	if err != nil {
		return signature, ErrSigner.Wrap(err)
	}

	switch networkType {
	case networks.TypeEVM:
		privateKeyECDSA, err := crypto.HexToECDSA(privateKeyHex)
		if err != nil {
			return signature, ErrSigner.Wrap(err)
		}

		signature, err = crypto.Sign(data, privateKeyECDSA)
		if err != nil {
			return nil, ErrSigner.Wrap(err)
		}
	case networks.TypeCasper:
		privateKeyBytes, err := hex.DecodeString(privateKeyHex)
		if err != nil {
			return signature, ErrSigner.Wrap(err)
		}

		switch {
		case len(privateKeyBytes) == ed25519.PrivateKeySize:
			publicKey := make([]byte, PublicKeySize)
			copy(publicKey, privateKeyBytes[PublicKeySize:])

			pair := casper_ed25519.ParseKeyPair(publicKey, privateKeyBytes[:PublicKeySize])

			casperSignature := pair.Sign(data)

			signature = casperSignature.SignatureData
		case len(privateKeyBytes) == Secp256k1PrivateKeyLength:
			privateKeyECDSA, err := crypto.HexToECDSA(privateKeyHex)
			if err != nil {
				return signature, ErrSigner.Wrap(err)
			}

			signature, err = crypto.Sign(data, privateKeyECDSA)
			if err != nil {
				return signature, ErrSigner.Wrap(err)
			}
		default:
			return signature, fmt.Errorf("invalid private key length: %d", len(privateKeyBytes))
		}
	case networks.TypeSolana:
		account, err := solana_types.AccountFromHex(privateKeyHex)
		if err != nil {
			return signature, ErrSigner.Wrap(err)
		}

		signature = account.Sign(data)
	default:
		return signature, errors.New("wrong network type")
	}

	return signature, nil
}

// PublicKey returns public key for specific network.
func (s *Service) PublicKey(ctx context.Context, networkType networks.Type) ([]byte, error) {
	var publicKey []byte

	privateKey, err := s.keyStore.Get(ctx, networkType, TypeDTTransaction)
	if err != nil {
		return publicKey, err
	}

	switch networkType {
	case networks.TypeEVM:
		privateKeyECDSA, err := crypto.HexToECDSA(privateKey)
		if err != nil {
			return publicKey, err
		}

		publicKey = append(publicKey, privateKeyECDSA.PublicKey.X.Bytes()...)
		publicKey = append(publicKey, privateKeyECDSA.PublicKey.Y.Bytes()...)
	case networks.TypeCasper:
		privateKeyHex, err := hex.DecodeString(privateKey)
		if err != nil {
			return publicKey, err
		}

		publicKey = ed25519.PrivateKey(privateKeyHex).Public().(ed25519.PublicKey)
	case networks.TypeSolana:
		account, err := solana_types.AccountFromHex(privateKey)
		if err != nil {
			return publicKey, err
		}

		publicKey = account.PublicKey.Bytes()
	default:
		return publicKey, errors.New("wrong network type")
	}

	return publicKey, nil
}
