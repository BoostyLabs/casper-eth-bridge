// Copyright (C) 2022 Creditor Corp. Group.
// See LICENSE for copying information.

package tempdb

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/url"
	"strings"

	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/internal/postgres"
)

// CreateSchema creates a schema if it doesn't exist.
func CreateSchema(ctx context.Context, db Execer, schema string) error {
	for try := 0; try < 5; try++ {
		_, err := db.ExecContext(ctx, `CREATE SCHEMA IF NOT EXISTS `+QuoteSchema(schema)+`;`)

		// Postgres `CREATE SCHEMA IF NOT EXISTS` may return "duplicate key value violates unique constraint".
		// In that case, we will retry rather than doing anything more complicated.
		//
		// See more in: https://stackoverflow.com/a/29908840/192220
		if postgres.IsConstraintError(err) {
			continue
		}
		return err
	}

	return nil
}

// DropSchema drops the named schema.
func DropSchema(ctx context.Context, db Execer, schema string) error {
	_, err := db.ExecContext(ctx, `DROP SCHEMA `+QuoteSchema(schema)+` CASCADE;`)
	return err
}

// CreateRandomTestingSchemaName creates a random schema name string.
func CreateRandomTestingSchemaName(n int) (string, error) {
	data := make([]byte, n)
	_, err := rand.Read(data)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(data), nil
}

// ConnstrWithSchema adds schema to a  connection string.
func ConnstrWithSchema(connstr, schema string) string {
	if strings.Contains(connstr, "?") {
		connstr += "&options="
	} else {
		connstr += "?options="
	}
	return connstr + url.QueryEscape("--search_path="+QuoteSchema(schema))
}

// QuoteSchema quotes a schema for use in an interpolated SQL string.
func QuoteSchema(ident string) string {
	s := strings.Replace(ident, string([]byte{0}), "", -1) // nolint
	return `"` + strings.Replace(s, `"`, `""`, -1) + `"`   // nolint
}
