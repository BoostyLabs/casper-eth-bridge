// Copyright (C) 2022 Creditor Corp. Group.
// See LICENSE for copying information.

package tempdb

import (
	"context"
	"database/sql"

	_ "github.com/lib/pq" // using postgres driver.
	"github.com/zeebo/errs"
)

// Error indicates about internal error in tembdb processing.
var Error = errs.Class("tempdb internal")

// Execer is for executing sql.
type Execer interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

// TempDatabase is a database (or something that works like an isolated database,
// such as a PostgreSQL schema) with a semi-unique name which will be cleaned up
// when closed. Mainly useful for testing purposes.
type TempDatabase struct {
	conn    *sql.DB
	ConnStr string
	Schema  string
	Driver  string
	Cleanup func(*sql.DB) error
}

// OpenUnique opens a postgres database with a temporary unique schema, which will be cleaned up when closed.
func OpenUnique(ctx context.Context, connstr string, schemaPrefix string) (*TempDatabase, error) {
	randomSchema, err := CreateRandomTestingSchemaName(8)
	if err != nil {
		return nil, err
	}
	schemaName := schemaPrefix + "-" + randomSchema
	connStrWithSchema := ConnstrWithSchema(connstr, schemaName)

	db, err := sql.Open("postgres", connstr)
	if err == nil {
		// check that connection actually worked before trying CreateSchema, to make
		// troubleshooting (lots) easier.
		err = db.PingContext(ctx)
	}
	if err != nil {
		return nil, Error.New("failed to connect to %q with driver postgres: %w", connStrWithSchema, err)
	}

	if err = CreateSchema(ctx, db, schemaName); err != nil {
		return nil, errs.Combine(err, db.Close())
	}

	cleanup := func(conn *sql.DB) error {
		return DropSchema(ctx, conn, schemaName)
	}

	// dbutil.Configure(db, "tmp_postgres", mon).
	return &TempDatabase{
		conn:    db,
		ConnStr: connStrWithSchema,
		Schema:  schemaName,
		Driver:  "postgres",
		Cleanup: cleanup,
	}, nil
}
