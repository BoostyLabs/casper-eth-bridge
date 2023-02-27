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

// ensures that networkBlocksDB implements networks.NetworkBlocks.
var _ networks.NetworkBlocks = (*networkBlocksDB)(nil)

// ErrNetworkBlocks indicates that there was an error in the database.
var ErrNetworkBlocks = errs.Class("network blocks repository")

// networkBlocksDB provide access to network blocks DB.
//
// architecture: Database
type networkBlocksDB struct {
	conn *sql.DB
}

// Create inserts network block to database.
func (networkBlocksDB *networkBlocksDB) Create(ctx context.Context, networkBlock networks.NetworkBlock) error {
	query := "INSERT INTO network_blocks(network_id,last_seen_block)VALUES($1,$2)"
	_, err := networkBlocksDB.conn.ExecContext(ctx, query, networkBlock.NetworkID, networkBlock.LastSeenBlock)
	return ErrNetworkBlocks.Wrap(err)
}

// Get returns last seen block by network id from database.
func (networkBlocksDB *networkBlocksDB) Get(ctx context.Context, networkID networks.ID) (int64, error) {
	var lastSeenBlock int64
	query := "SELECT last_seen_block FROM network_blocks WHERE network_id = $1"
	row := networkBlocksDB.conn.QueryRowContext(ctx, query, networkID)

	if err := row.Scan(&lastSeenBlock); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return lastSeenBlock, ErrNetworkBlocks.Wrap(bridge.ErrNoNetworkBlock)
		}

		return lastSeenBlock, ErrNetworkBlocks.Wrap(err)
	}

	return lastSeenBlock, nil
}

// Update updates network block in database.
func (networkBlocksDB *networkBlocksDB) Update(ctx context.Context, networkBlock networks.NetworkBlock) error {
	query := "UPDATE network_blocks SET last_seen_block = $1 WHERE network_id = $2"
	result, err := networkBlocksDB.conn.ExecContext(ctx, query, networkBlock.LastSeenBlock, networkBlock.NetworkID)
	if err != nil {
		return ErrNetworkBlocks.Wrap(err)
	}

	rowNum, err := result.RowsAffected()
	if rowNum == 0 && err == nil {
		return ErrNetworkBlocks.Wrap(bridge.ErrNoNetworkBlock)
	}

	return ErrNetworkBlocks.Wrap(err)
}
