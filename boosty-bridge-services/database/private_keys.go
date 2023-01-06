// Copyright (C) 2022 Creditor Corp. Group.
// See LICENSE for copying information.

package database

import (
	"context"
	"database/sql"
	"errors"

	"github.com/zeebo/errs"

	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/networks"
	signer_service "github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/signer"
)

// ErrPrivateKeys indicates that there was an error in the database.
var ErrPrivateKeys = errs.Class("private keys repository")

// privateKeysDB provide access to private keys DB.
//
// architecture: Database
type privateKeysDB struct {
	conn *sql.DB
}

// Create inserts private key to database.
func (privateKeysDB *privateKeysDB) Create(ctx context.Context, privateKey signer_service.PrivateKey) error {
	query := "INSERT INTO private_keys(network_type,private_key)VALUES($1,$2)"
	_, err := privateKeysDB.conn.ExecContext(ctx, query, privateKey.NetworkType, privateKey.Key)
	return ErrPrivateKeys.Wrap(err)
}

// Get returns private key by network type from database.
func (privateKeysDB *privateKeysDB) Get(ctx context.Context, networkType networks.Type) (string, error) {
	var value string
	query := "SELECT private_key FROM private_keys WHERE network_type = $1"
	row := privateKeysDB.conn.QueryRowContext(ctx, query, networkType)

	if err := row.Scan(&value); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return value, signer_service.ErrNoPrivateKey
		}

		return value, ErrPrivateKeys.Wrap(err)
	}

	return value, nil
}

// Update updates private key in database.
func (privateKeysDB *privateKeysDB) Update(ctx context.Context, privateKey signer_service.PrivateKey) error {
	query := "UPDATE private_keys SET private_key = $1 WHERE network_type = $2"
	result, err := privateKeysDB.conn.ExecContext(ctx, query, privateKey.Key, privateKey.NetworkType)
	if err != nil {
		return ErrPrivateKeys.Wrap(err)
	}

	rowNum, err := result.RowsAffected()
	if rowNum == 0 && err == nil {
		return signer_service.ErrNoPrivateKey
	}

	return ErrPrivateKeys.Wrap(err)
}
