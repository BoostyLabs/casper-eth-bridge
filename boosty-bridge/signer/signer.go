// Copyright (C) 2022 Creditor Corp. Group.
// See LICENSE for copying information.

package signer

import (
	"context"
	"errors"

	"tricorn/bridge/networks"
)

var (
	// ErrNoPrivateKey indicates that private key does not exist.
	ErrNoPrivateKey = errors.New("private key does not exist")
)

const (
	// PublicKeySize defines public key bytes size.
	PublicKeySize = 32
)

// DB provides access to all databases and database related functionality.
//
// architecture: Master Database.
type DB interface {
	// KeyStore provides access to private keys db.
	KeyStore() KeyStore

	// Close closes underlying db connection.
	Close() error

	// CreateSchema create tables.
	CreateSchema(ctx context.Context) error
}

// KeyStore is exposing access to private keys db.
//
// architecture: DB
type KeyStore interface {
	// Create inserts private key to database.
	Create(ctx context.Context, privateKey PrivateKey) error
	// Get returns private key by network type from database.
	Get(ctx context.Context, networkType networks.Type, keyType Type) (string, error)
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
	Type        Type
}

// Type defines list of possible private key types.
type Type string

const (
	// TypeDT_TRANSACTION describes transaction type of private key.
	TypeDTTransaction Type = "DT_TRANSACTION"
	// TypeDTSignature describes signature type of private key.
	TypeDTSignature Type = "DT_SIGNATURE"
)

// String returns string for Type type.
func (t Type) String() string {
	return string(t)
}
