// Copyright (C) 2022 Creditor Corp. Group.
// See LICENSE for copying information.

package database

import (
	"context"
	"database/sql"
	"errors"

	"github.com/zeebo/errs"

	"tricorn/bridge"
	"tricorn/bridge/networks"
)

// ensures that networkNoncesDB implements networks.NetworkNonces.
var _ networks.Nonces = (*networkNoncesDB)(nil)

// ErrNetworkNonces indicates that there was an error in the database.
var ErrNetworkNonces = errs.Class("network nonces repository")

// networkNoncesDB provide access to network nonces DB.
//
// architecture: Database
type networkNoncesDB struct {
	conn *sql.DB
}

// Create inserts network nonce to database.
func (networkNoncesDB *networkNoncesDB) Create(ctx context.Context, networkNonce networks.NetworkNonce) error {
	query := "INSERT INTO network_nonces(network_id,nonce) VALUES($1,$2)"
	_, err := networkNoncesDB.conn.ExecContext(ctx, query, networkNonce.NetworkID, networkNonce.Nonce)
	return ErrNetworkNonces.Wrap(err)
}

// Get returns nonce by network id from database.
func (networkNoncesDB *networkNoncesDB) Get(ctx context.Context, networkID networks.ID) (int64, error) {
	var nonce int64
	query := "SELECT nonce FROM network_nonces WHERE network_id = $1"
	row := networkNoncesDB.conn.QueryRowContext(ctx, query, networkID)

	if err := row.Scan(&nonce); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nonce, ErrNetworkNonces.Wrap(bridge.ErrNoNetworkNonce)
		}

		return nonce, ErrNetworkNonces.Wrap(err)
	}

	return nonce, nil
}

// List returns list of network id from database.
func (networkNoncesDB *networkNoncesDB) List(ctx context.Context) (_ []networks.ID, err error) {
	networkIDs := make([]networks.ID, 0)

	query := "SELECT network_id FROM network_nonces"
	rows, err := networkNoncesDB.conn.QueryContext(ctx, query)
	if err != nil {
		return networkIDs, Error.Wrap(err)
	}

	defer func() {
		err = errs.Combine(err, rows.Close())
	}()

	for rows.Next() {
		var networkID networks.ID
		err := rows.Scan(&networkID)
		if err != nil {
			return networkIDs, Error.Wrap(err)
		}

		networkIDs = append(networkIDs, networkID)
	}

	return networkIDs, nil
}

// Update updates network nonce in database.
func (networkNoncesDB *networkNoncesDB) Update(ctx context.Context, networkNonce networks.NetworkNonce) error {
	query := "UPDATE network_nonces SET nonce = $1 WHERE network_id = $2"
	result, err := networkNoncesDB.conn.ExecContext(ctx, query, networkNonce.Nonce, networkNonce.NetworkID)
	if err != nil {
		return ErrNetworkNonces.Wrap(err)
	}

	rowNum, err := result.RowsAffected()
	if rowNum == 0 && err == nil {
		return ErrNetworkNonces.Wrap(bridge.ErrNoNetworkNonce)
	}

	return ErrNetworkNonces.Wrap(err)
}

// Increment increments nonce for specific network in the database.
func (networkNoncesDB *networkNoncesDB) Increment(ctx context.Context, networkID networks.ID) error {
	query := "UPDATE network_nonces SET nonce = nonce + 1 WHERE network_id = $1"
	result, err := networkNoncesDB.conn.ExecContext(ctx, query, networkID)
	if err != nil {
		return ErrNetworkNonces.Wrap(err)
	}

	rowNum, err := result.RowsAffected()
	if rowNum == 0 && err == nil {
		return ErrNetworkNonces.Wrap(bridge.ErrNoNetworkNonce)
	}

	return ErrNetworkNonces.Wrap(err)
}
