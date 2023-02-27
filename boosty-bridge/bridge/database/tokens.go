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

// ensures that tokensDB implements bridge.Tokens.
var _ bridge.Tokens = (*tokensDB)(nil)

// ErrTokens indicates that there was an error in the database.
var ErrTokens = errs.Class("tokens repository")

// tokensDB provide access to tokens DB.
//
// architecture: Database
type tokensDB struct {
	conn *sql.DB
}

// Create inserts token to database.
func (tokensDB *tokensDB) Create(ctx context.Context, token bridge.Token) error {
	query := "INSERT INTO tokens(short_name,long_name) VALUES($1,$2)"
	_, err := tokensDB.conn.ExecContext(ctx, query, token.ShortName, token.LongName)
	return ErrTokens.Wrap(err)
}

// Get returns token by id from database.
func (tokensDB *tokensDB) Get(ctx context.Context, id int64) (bridge.Token, error) {
	token := bridge.Token{
		ID: id,
	}

	query := "SELECT short_name, long_name FROM tokens WHERE id = $1"
	row := tokensDB.conn.QueryRowContext(ctx, query, id)

	if err := row.Scan(&token.ShortName, &token.LongName); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return token, ErrTokens.Wrap(bridge.ErrNoToken)
		}

		return token, ErrTokens.Wrap(err)
	}

	return token, nil
}

// List returns list of tokents, supported by network, from database.
func (tokensDB *tokensDB) List(ctx context.Context, networkID networks.ID) (_ []bridge.Token, err error) {
	tokens := make([]bridge.Token, 0)

	query := `SELECT id, short_name, long_name FROM tokens INNER JOIN network_tokens as nt ON nt.token_id = id
		WHERE nt.network_id = $1`
	rows, err := tokensDB.conn.QueryContext(ctx, query, networkID)
	if err != nil {
		return tokens, Error.Wrap(err)
	}

	defer func() {
		err = errs.Combine(err, rows.Close())
	}()

	for rows.Next() {
		var token bridge.Token
		err := rows.Scan(&token.ID, &token.ShortName, &token.LongName)
		if err != nil {
			return tokens, Error.Wrap(err)
		}

		tokens = append(tokens, token)
	}

	return tokens, nil
}

// Update updates token in database.
func (tokensDB *tokensDB) Update(ctx context.Context, token bridge.Token) error {
	query := "UPDATE tokens SET short_name = $1, long_name = $2 WHERE id = $3"
	result, err := tokensDB.conn.ExecContext(ctx, query, token.ShortName, token.LongName, token.ID)
	if err != nil {
		return ErrTokens.Wrap(err)
	}

	rowNum, err := result.RowsAffected()
	if rowNum == 0 && err == nil {
		return ErrTokens.Wrap(bridge.ErrNoToken)
	}

	return ErrTokens.Wrap(err)
}
