// Copyright (C) 2022 Creditor Corp. Group.
// See LICENSE for copying information.

package signer

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/networks"
	casper_ed25519 "github.com/casper-ecosystem/casper-golang-sdk/keypair/ed25519"
	"github.com/casper-ecosystem/casper-golang-sdk/sdk"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/zeebo/errs"
)

// ErrSigner indicates that there was an error in the service.
var ErrSigner = errs.Class("signer service")

// Service is handling signer related logic.
//
// architecture: Service
type Service struct {
	config Config
	signer DB
}

// NewService is constructor for Service.
func NewService(config Config, signer DB) *Service {
	return &Service{
		config: config,
		signer: signer,
	}
}

// Sign creates and returns signature from data.
func (s *Service) Sign(ctx context.Context, networkType networks.Type, data []byte) ([]byte, error) {
	var signature []byte

	privateKey, err := s.signer.Get(ctx, networkType)
	if err != nil {
		return signature, ErrSigner.Wrap(err)
	}

	switch networkType {
	case networks.TypeEVM:
		privateKeyECDSA, err := crypto.HexToECDSA(privateKey)
		if err != nil {
			return signature, ErrSigner.Wrap(err)
		}

		signature, err = crypto.Sign(data, privateKeyECDSA)
		if err != nil {
			return nil, ErrSigner.Wrap(err)
		}
	case networks.TypeCasper:
		deploy := new(sdk.Deploy)
		err := json.Unmarshal(data, deploy)
		if err != nil {
			return signature, ErrSigner.Wrap(err)
		}

		privateKeyHex, err := hex.DecodeString(privateKey)
		if err != nil {
			return signature, ErrSigner.Wrap(err)
		}

		if len(privateKeyHex) != ed25519.PrivateKeySize {
			return signature, fmt.Errorf("invalid private key length: %d", len(privateKeyHex))
		}

		publicKey := make([]byte, PublicKeySize)
		copy(publicKey, privateKeyHex[PublicKeySize:])

		pair := casper_ed25519.ParseKeyPair(publicKey, privateKeyHex[:PublicKeySize])

		deploy.SignDeploy(pair)

		signature = deploy.Approvals[0].Signature.SignatureData
	default:
		return signature, errors.New("wrong network type")
	}

	return signature, nil
}

// PublicKey returns public key for specific network.
func (s *Service) PublicKey(ctx context.Context, networkType networks.Type) ([]byte, error) {
	var publicKey []byte

	privateKey, err := s.signer.Get(ctx, networkType)
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
	default:
		return publicKey, errors.New("wrong network type")
	}

	return publicKey, nil
}
