// Copyright (C) 2022 Creditor Corp. Group.
// See LICENSE for copying information.

package db

import (
	"context"

	signer_service "github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/signer"
)

// DB provides access to all databases and database related functionality.
//
// architecture: Master Database.
type DB interface {
	// PrivateKeys provides access to private keys db.
	PrivateKeys() signer_service.DB

	// Close closes underlying db connection.
	Close() error

	// CreateSchema create tables.
	CreateSchema(ctx context.Context) error
}
