// Copyright (C) 2022 Creditor Corp. Group.
// See LICENSE for copying information.

package dbtesting

import (
	"context"
	"fmt"
	"testing"

	"tricorn/internal/postgres"
	"tricorn/internal/tempdb"
	"tricorn/signer"
	"tricorn/signer/database"
)

// Database describes a test database.
type Database struct {
	Name string
	URL  string
}

// tempMasterDB is a signer.DB-implementing type that cleans up after itself when closed.
type tempMasterDB struct {
	signer.DB
	tempDB *tempdb.TempDatabase
}

// DefaultTestConn default test conn string that is expected to work with postgres server.
const DefaultTestConn = "postgres://postgres:1212@localhost:6432/boosty_bridge_db?sslmode=disable"

// Run method will establish connection with db, create tables in random schema, run tests.
func Run(t *testing.T, test func(ctx context.Context, t *testing.T, db signer.DB)) {
	t.Run("Postgres", func(t *testing.T) {
		ctx := context.Background()

		options := Database{
			Name: "Postgres",
			URL:  DefaultTestConn,
		}

		masterDB, err := CreateMasterDB(ctx, t.Name(), "Test", 0, options)
		if err != nil {
			t.Fatal(err)
		}
		defer func() {
			err := masterDB.Close()
			if err != nil {
				t.Fatal(err)
			}
		}()
		err = masterDB.CreateSchema(ctx)
		if err != nil {
			t.Fatal(err)
		}

		test(ctx, t, masterDB)
	})
}

// CreateMasterDB creates a new signer.DB for testing.
func CreateMasterDB(ctx context.Context, name string, category string, index int, dbInfo Database) (db signer.DB, err error) {
	if dbInfo.URL == "" {
		return nil, fmt.Errorf("database %s connection string not provided", dbInfo.Name)
	}

	schemaSuffix, err := tempdb.CreateRandomTestingSchemaName(6)
	if err != nil {
		return nil, err
	}

	schema := postgres.SchemaName(name, category, index, schemaSuffix)

	tempDB, err := tempdb.OpenUnique(ctx, dbInfo.URL, schema)
	if err != nil {
		return nil, err
	}

	return CreateMasterDBOnTopOf(tempDB)
}

// CreateMasterDBOnTopOf creates a new signer.DB on top of an already existing
// temporary database.
func CreateMasterDBOnTopOf(tempDB *tempdb.TempDatabase) (db signer.DB, err error) {
	masterDB, err := database.New(tempDB.ConnStr)
	return &tempMasterDB{DB: masterDB, tempDB: tempDB}, err
}
