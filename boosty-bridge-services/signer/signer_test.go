// Copyright (C) 2022 Creditor Corp. Group.
// See LICENSE for copying information.

package signer_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	signer "github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/cmd/signer/db"
	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/database/dbtesting"
	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/networks"
	signer_service "github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/signer"
)

func TestPrivateKeysDB(t *testing.T) {
	privateKey := signer_service.PrivateKey{
		NetworkType: networks.TypeEVM,
		Key:         "private_key",
	}

	dbtesting.Run(t, func(ctx context.Context, t *testing.T, db signer.DB) {
		repository := db.PrivateKeys()

		t.Run("Create", func(t *testing.T) {
			err := repository.Create(ctx, privateKey)
			require.NoError(t, err)
		})

		t.Run("Get", func(t *testing.T) {
			value, err := repository.Get(ctx, privateKey.NetworkType)
			require.NoError(t, err)
			assert.Equal(t, privateKey.Key, value)
		})

		t.Run("Negative Get", func(t *testing.T) {
			_, err := repository.Get(ctx, "")
			require.Error(t, err)
			require.True(t, errors.Is(err, signer_service.ErrNoPrivateKey))
		})

		t.Run("Update", func(t *testing.T) {
			privateKey.Key = "new_private_key"
			err := repository.Update(ctx, privateKey)
			require.NoError(t, err)
		})

		t.Run("Negative Update", func(t *testing.T) {
			privateKey.NetworkType = networks.TypeCasper
			err := repository.Update(ctx, privateKey)
			require.Error(t, err)
			require.True(t, errors.Is(err, signer_service.ErrNoPrivateKey))
		})
	})
}
