// Copyright (C) 2022 Creditor Corp. Group.
// See LICENSE for copying information.

package signer

import (
	"context"
	"errors"

	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/networks"
)

var (
	// ErrNoPrivateKey indicates that private key does not exist.
	ErrNoPrivateKey = errors.New("private key does not exist")
)

const (
	// PublicKeySize defines public key bytes size.
	PublicKeySize = 32
)

// DB is exposing access to private keys db.
//
// architecture: DB
type DB interface {
	// Create inserts private key to database.
	Create(ctx context.Context, privateKey PrivateKey) error
	// Get returns private key by network type from database.
	Get(ctx context.Context, networkType networks.Type) (string, error)
	// Update updates private key in database.
	Update(ctx context.Context, privateKey PrivateKey) error
}

// Config is configuration to sign transactions.
type Config struct {
	ChainID int64 `env:"CHAIN_ID"`
}

// PrivateKey contains private key for specific network.
type PrivateKey struct {
	NetworkType networks.Type
	Key         string
}
