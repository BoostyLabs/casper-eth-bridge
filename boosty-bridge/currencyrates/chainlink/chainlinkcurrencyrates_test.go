// Copyright (C) 2023 Creditor Corp. Group.
// See LICENSE for copying information.

package chainlink_test

import (
	"context"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"tricorn/currencyrates/chainlink"
)

func TestChainlinkClient(t *testing.T) {
	ctx := context.Background()
	baseURL := "https://min-api.cryptocompare.com/data/price"

	client := chainlink.New(baseURL)

	t.Run("get rate", func(t *testing.T) {
		currency, err := client.GetPrice(ctx, "ETH", "BTC")
		require.NoError(t, err)
		require.NotNil(t, currency)
		require.Equal(t, "BTC", currency.Symbol)
		require.NotEmpty(t, "BTC", currency.Price)
	})

	t.Run("negative get price", func(t *testing.T) {
		currency, err := client.GetPrice(ctx, "ETHHH", "BTC")
		require.Error(t, err)
		require.Equal(t, "cccagg_or_exchange market does not exist for this coin pair (ETHHH-BTC)", err.Error())
		require.NotNil(t, currency)
		require.Empty(t, currency.Symbol)
		require.Empty(t, currency.Price)
	})

	t.Run("negative get price", func(t *testing.T) {
		currency, err := client.GetPrice(ctx, "ETH", "")
		require.Error(t, err)
		require.Equal(t, "tsyms param is empty or null.", err.Error())
		require.NotNil(t, currency)
		require.Empty(t, currency.Symbol)
		require.Empty(t, currency.Price)
	})

	t.Run("convert", func(t *testing.T) {
		amount, err := client.Convert(ctx, "ETH", "USD", big.NewFloat(0.8))
		require.NoError(t, err)
		require.NotEmpty(t, amount)
	})

	t.Run("negative convert", func(t *testing.T) {
		amount, err := client.Convert(ctx, "ETH", "USD", big.NewFloat(-0.8))
		require.Error(t, err)
		require.Equal(t, "amount -0.8 is less than or equal to 0", err.Error())
		require.Nil(t, amount)
	})
}
