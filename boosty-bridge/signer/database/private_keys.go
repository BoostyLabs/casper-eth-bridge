// Copyright (C) 2022 Creditor Corp. Group.
// See LICENSE for copying information.

package database

import (
	"context"
	"database/sql"
	"errors"

	"github.com/zeebo/errs"

	"tricorn/bridge/networks"
	"tricorn/signer"
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
func (privateKeysDB *privateKeysDB) Create(ctx context.Context, privateKey signer.PrivateKey) error {
	query := "INSERT INTO private_keys(network_type,private_key,type) VALUES($1,$2,$3)"
	_, err := privateKeysDB.conn.ExecContext(ctx, query, privateKey.NetworkType, privateKey.Key, privateKey.Type)
	return ErrPrivateKeys.Wrap(err)
}

// Get returns private key by network type from database.
func (privateKeysDB *privateKeysDB) Get(ctx context.Context, networkType networks.Type, keyType signer.Type) (string, error) {
	var privateKey string
	query := "SELECT private_key FROM private_keys WHERE network_type = $1 AND type = $2"
	row := privateKeysDB.conn.QueryRowContext(ctx, query, networkType, keyType)

	if err := row.Scan(&privateKey); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return privateKey, signer.ErrNoPrivateKey
		}

		return privateKey, ErrPrivateKeys.Wrap(err)
	}

	return privateKey, nil
}

// Update updates private key in database.
func (privateKeysDB *privateKeysDB) Update(ctx context.Context, privateKey signer.PrivateKey) error {
	query := "UPDATE private_keys SET private_key = $1 WHERE network_type = $2 AND type = $3"
	result, err := privateKeysDB.conn.ExecContext(ctx, query, privateKey.Key, privateKey.NetworkType, privateKey.Type)
	if err != nil {
		return ErrPrivateKeys.Wrap(err)
	}

	rowNum, err := result.RowsAffected()
	if rowNum == 0 && err == nil {
		return signer.ErrNoPrivateKey
	}

	return ErrPrivateKeys.Wrap(err)
}
