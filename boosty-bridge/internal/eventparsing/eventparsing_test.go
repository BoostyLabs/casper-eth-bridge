// Copyright (C) 2022 Creditor Corp. Group.
// See LICENSE for copying information.

package eventparsing_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"tricorn/internal/eventparsing"
)

func TestEventParsing(t *testing.T) {
	eventData := eventparsing.EventData{
		Bytes: "7c000000013c0c1847d1c410338ab9b4ee0919c181cf26085997ff9c797e8a1ae5b02ddf2306000000474f45524c49280000003330393566393535646137303062393632313563666663396263363461623265363965623764616203a0860100daa2b596e0a496b04933e241e0567f2bcbecc829aa57d88cab096c28fd07dee2",
	}

	expectedEventTyte := 1
	expectedTokenContractAddress := "3c0c1847d1c410338ab9b4ee0919c181cf26085997ff9c797e8a1ae5b02ddf23"
	expectedChainName := "GOERLI"
	expectedChainAddress := "3095f955da700b96215cffc9bc64ab2e69eb7dab"
	expectedAmount := 100000
	expectedUserWalletAddress := "daa2b596e0a496b04933e241e0567f2bcbecc829aa57d88cab096c28fd07dee2"

	t.Run("GetEventType", func(t *testing.T) {
		actualEventTyte, err := eventData.GetEventType()
		require.NoError(t, err)
		require.Equal(t, expectedEventTyte, actualEventTyte)
	})

	t.Run("GetTokenContractAddress", func(t *testing.T) {
		actualTokenContractAddress := eventData.GetTokenContractAddress()
		require.Equal(t, expectedTokenContractAddress, actualTokenContractAddress)
	})

	t.Run("GetChainName", func(t *testing.T) {
		actualChainName, err := eventData.GetChainName()
		require.NoError(t, err)
		require.Equal(t, expectedChainName, actualChainName)
	})

	t.Run("GetChainAddress", func(t *testing.T) {
		actualChainAddress, err := eventData.GetChainAddress()
		require.NoError(t, err)
		require.Equal(t, expectedChainAddress, actualChainAddress)
	})

	t.Run("GetAmount", func(t *testing.T) {
		actualAmount, err := eventData.GetAmount()
		require.NoError(t, err)
		require.Equal(t, expectedAmount, actualAmount)
	})

	t.Run("GetUserWalletAddress", func(t *testing.T) {
		actualUserWalletAddress := eventData.GetUserWalletAddress()
		require.Equal(t, expectedUserWalletAddress, actualUserWalletAddress)
	})
}
