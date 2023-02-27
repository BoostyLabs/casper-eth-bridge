// Copyright (C) 2022 Creditor Corp. Group.
// See LICENSE for copying information.

package bridge_test

import (
	"context"
	"encoding/hex"
	"errors"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tricorn/bridge"
	"tricorn/bridge/database/dbtesting"
	"tricorn/bridge/networks"
	"tricorn/bridge/transactions"
	"tricorn/bridge/transfers"
)

func TestNetworkBlocksDB(t *testing.T) {
	networkBlock := networks.NetworkBlock{
		NetworkID:     networks.IDCasper,
		LastSeenBlock: 1,
	}

	dbtesting.Run(t, func(ctx context.Context, t *testing.T, db bridge.DB) {
		repository := db.NetworkBlocks()

		t.Run("Negative Get", func(t *testing.T) {
			_, err := repository.Get(ctx, 0)
			require.Error(t, err)
			require.True(t, errors.Is(err, bridge.ErrNoNetworkBlock))
		})

		t.Run("Negative Update", func(t *testing.T) {
			networkBlock.NetworkID = networks.IDCasperTest
			err := repository.Update(ctx, networkBlock)
			require.Error(t, err)
			require.True(t, errors.Is(err, bridge.ErrNoNetworkBlock))
		})

		t.Run("Create", func(t *testing.T) {
			err := repository.Create(ctx, networkBlock)
			require.NoError(t, err)
		})

		t.Run("Get", func(t *testing.T) {
			lastSeenBlock, err := repository.Get(ctx, networkBlock.NetworkID)
			require.NoError(t, err)
			assert.Equal(t, networkBlock.LastSeenBlock, lastSeenBlock)
		})

		t.Run("Update", func(t *testing.T) {
			networkBlock.LastSeenBlock = 5
			err := repository.Update(ctx, networkBlock)
			require.NoError(t, err)
		})
	})
}

func TestNetworkNoncesDB(t *testing.T) {
	networkNonce := networks.NetworkNonce{
		NetworkID: networks.IDCasper,
		Nonce:     1,
	}

	dbtesting.Run(t, func(ctx context.Context, t *testing.T, db bridge.DB) {
		repository := db.Nonces()

		t.Run("Negative Get", func(t *testing.T) {
			_, err := repository.Get(ctx, 0)
			require.Error(t, err)
			require.True(t, errors.Is(err, bridge.ErrNoNetworkNonce))
		})

		t.Run("Negative Update", func(t *testing.T) {
			networkNonce.NetworkID = networks.IDCasperTest
			err := repository.Update(ctx, networkNonce)
			require.Error(t, err)
			require.True(t, errors.Is(err, bridge.ErrNoNetworkNonce))
		})

		t.Run("Negative Increment", func(t *testing.T) {
			networkID := networks.IDCasperTest
			err := repository.Increment(ctx, networkID)
			require.Error(t, err)
			require.True(t, errors.Is(err, bridge.ErrNoNetworkNonce))
		})

		t.Run("Empty List", func(t *testing.T) {
			list, err := repository.List(ctx)
			require.NoError(t, err)
			assert.Empty(t, list)
		})

		t.Run("Create", func(t *testing.T) {
			err := repository.Create(ctx, networkNonce)
			require.NoError(t, err)
		})

		t.Run("Get", func(t *testing.T) {
			nonce, err := repository.Get(ctx, networkNonce.NetworkID)
			require.NoError(t, err)
			assert.Equal(t, networkNonce.Nonce, nonce)
		})

		t.Run("List", func(t *testing.T) {
			list, err := repository.List(ctx)
			require.NoError(t, err)
			assert.Len(t, list, 1)
			assert.EqualValues(t, networkNonce.NetworkID, list[0])
		})

		t.Run("Update", func(t *testing.T) {
			networkNonce.Nonce = 5
			err := repository.Update(ctx, networkNonce)
			require.NoError(t, err)
		})

		t.Run("Increment", func(t *testing.T) {
			networkID := networks.IDCasperTest
			err := repository.Increment(ctx, networkID)
			require.NoError(t, err)

			nonce, err := repository.Get(ctx, networkID)
			require.NoError(t, err)
			assert.EqualValues(t, nonce, 6)
		})
	})
}

func TestNetworkTokensDB(t *testing.T) {
	casperContractAddress, err := networks.StringToBytes(networks.IDCasper, "0xF035b4f54B4A5Dd154cD378ce9d77a12A2c97764")
	require.NoError(t, err)

	networkTokenCasper := networks.NetworkToken{
		NetworkID:       networks.IDCasper,
		TokenID:         1,
		ContractAddress: casperContractAddress,
		Decimals:        18,
	}

	ethContractAddress, err := networks.StringToBytes(networks.IDEth, "0x0E26df2BaaFBC976a104EE3cbcf1B467ff1b7a69")
	require.NoError(t, err)

	networkTokenEth := networks.NetworkToken{
		NetworkID:       networks.IDEth,
		TokenID:         1,
		ContractAddress: ethContractAddress,
		Decimals:        18,
	}

	dbtesting.Run(t, func(ctx context.Context, t *testing.T, db bridge.DB) {
		repository := db.NetworkTokens()

		t.Run("Negative Get", func(t *testing.T) {
			_, err := repository.Get(ctx, 0, 0)
			require.Error(t, err)
			require.True(t, errors.Is(err, bridge.ErrNoNetworkToken))
		})

		t.Run("Negative Update", func(t *testing.T) {
			networkTokenCasper.NetworkID = networks.IDCasperTest
			err := repository.Update(ctx, networkTokenCasper)
			require.Error(t, err)
			require.True(t, errors.Is(err, bridge.ErrNoNetworkToken))
		})

		t.Run("Empty List", func(t *testing.T) {
			list, err := repository.List(ctx, networkTokenCasper.TokenID)
			require.NoError(t, err)
			assert.Empty(t, list)
		})

		t.Run("Create", func(t *testing.T) {
			err := repository.Create(ctx, networkTokenCasper)
			require.NoError(t, err)

			err = repository.Create(ctx, networkTokenEth)
			require.NoError(t, err)
		})

		t.Run("Get", func(t *testing.T) {
			token, err := repository.Get(ctx, networkTokenCasper.NetworkID, networkTokenCasper.TokenID)
			require.NoError(t, err)
			assert.Equal(t, networkTokenCasper, token)
		})

		t.Run("List", func(t *testing.T) {
			list, err := repository.List(ctx, networkTokenCasper.TokenID)
			require.NoError(t, err)
			assert.Len(t, list, 2)
			assert.ElementsMatch(t, []networks.NetworkToken{networkTokenCasper, networkTokenEth}, list)
		})

		t.Run("Update", func(t *testing.T) {
			networkTokenCasper.Decimals = 9
			err := repository.Update(ctx, networkTokenCasper)
			require.NoError(t, err)
		})
	})
}

func TestTokenTransfersDB(t *testing.T) {
	senderAddress, err := hex.DecodeString("4zXwdbUDWo1S5AP2CEfv4zAPRds5PQUG1dyqLLvib2xu")
	require.Error(t, err)

	recipientAddress, err := hex.DecodeString("0x9744bC7A2D91928017E1DEdf98Ff7d912d6Cd263")
	require.Error(t, err)

	tokenTransfer := transfers.TokenTransfer{
		ID:                 1,
		TriggeringTx:       1,
		OutboundTx:         0,
		TokenID:            1,
		Amount:             *new(big.Int).SetInt64(1),
		Status:             "waiting",
		SenderNetworkID:    1,
		SenderAddress:      senderAddress,
		RecipientNetworkID: 2,
		RecipientAddress:   recipientAddress,
	}

	dbtesting.Run(t, func(ctx context.Context, t *testing.T, db bridge.DB) {
		repository := db.TokenTransfers()

		t.Run("Negative Get", func(t *testing.T) {
			_, err := repository.Get(ctx, 0)
			require.Error(t, err)
			require.True(t, errors.Is(err, bridge.ErrNoTokenTransfer))
		})

		t.Run("Negative Update", func(t *testing.T) {
			tokenTransfer.OutboundTx = 2
			err := repository.Update(ctx, tokenTransfer)
			require.Error(t, err)
			require.True(t, errors.Is(err, bridge.ErrNoTokenTransfer))
		})

		t.Run("Create", func(t *testing.T) {
			err := repository.Create(ctx, tokenTransfer)
			require.NoError(t, err)
		})

		t.Run("Get", func(t *testing.T) {
			tokenTransferFromDB, err := repository.Get(ctx, tokenTransfer.ID)
			require.NoError(t, err)
			assert.Equal(t, tokenTransfer, tokenTransferFromDB)
		})

		t.Run("GetByAllParams", func(t *testing.T) {
			params := transfers.TokenTransfer{
				TokenID:          1, // todo: dynamically change.
				Amount:           *new(big.Int).SetInt64(1),
				SenderAddress:    senderAddress,
				RecipientAddress: recipientAddress,
			}

			tokenTransferFromDB, err := repository.GetByAllParams(ctx, params)
			require.NoError(t, err)
			assert.Equal(t, tokenTransfer, tokenTransferFromDB)
		})

		t.Run("Update", func(t *testing.T) {
			tokenTransfer.Status = "finished"
			err := repository.Update(ctx, tokenTransfer)
			require.NoError(t, err)
		})
	})

	testHash := []byte{1, 2, 3, 4, 5}
	token := bridge.Token{
		ID:        1,
		ShortName: "TESTONIUM",
		LongName:  "TESTONIUM TOKEN",
	}
	transaction1 := transactions.Transaction{
		ID:          1,
		NetworkID:   networks.IDCasper,
		TxHash:      testHash,
		Sender:      senderAddress,
		BlockNumber: 1,
		SeenAt:      time.Now(),
	}
	transaction2 := transactions.Transaction{
		ID:          2,
		NetworkID:   networks.IDSolana,
		TxHash:      []byte{},
		Sender:      []byte{},
		BlockNumber: 1,
		SeenAt:      time.Now(),
	}
	transaction3 := transactions.Transaction{
		ID:          3,
		NetworkID:   networks.IDCasper,
		TxHash:      []byte{},
		Sender:      senderAddress,
		BlockNumber: 2,
		SeenAt:      time.Now(),
	}
	transaction4 := transactions.Transaction{
		ID:          4,
		NetworkID:   networks.IDEth,
		TxHash:      []byte{},
		Sender:      []byte{},
		BlockNumber: 1,
		SeenAt:      time.Now(),
	}
	tokenTransfer1 := transfers.TokenTransfer{
		ID:                 1,
		TriggeringTx:       transaction1.ID,
		OutboundTx:         transaction2.ID,
		TokenID:            token.ID,
		Amount:             *new(big.Int).SetInt64(1),
		Status:             transfers.StatusFinished,
		SenderNetworkID:    int64(transaction1.NetworkID),
		SenderAddress:      senderAddress,
		RecipientNetworkID: int64(transaction2.NetworkID),
		RecipientAddress:   recipientAddress,
	}
	tokenTransfer2 := transfers.TokenTransfer{
		ID:                 2,
		TriggeringTx:       transaction3.ID,
		TokenID:            token.ID,
		Amount:             *new(big.Int).SetInt64(1),
		Status:             transfers.StatusWaiting,
		SenderNetworkID:    int64(transaction3.NetworkID),
		SenderAddress:      senderAddress,
		RecipientNetworkID: int64(transaction4.NetworkID),
		RecipientAddress:   recipientAddress,
	}

	dbtesting.Run(t, func(ctx context.Context, t *testing.T, db bridge.DB) {
		tokenTransfersRepository := db.TokenTransfers()
		tokensRepository := db.Tokens()
		transactionsRepository := db.Transactions()

		t.Run("Empty GetByNetworkAndTx", func(t *testing.T) {
			_, err := tokenTransfersRepository.GetByNetworkAndTx(ctx, networks.IDCasper, testHash)
			require.Error(t, err)
			require.True(t, errors.Is(err, bridge.ErrNoTokenTransfer))
		})

		t.Run("Empty ListByUser", func(t *testing.T) {
			list, err := tokenTransfersRepository.ListByUser(ctx, 0, 2, senderAddress, networks.IDCasper)
			require.NoError(t, err)
			assert.Empty(t, list)
		})

		t.Run("Zero CountByUser", func(t *testing.T) {
			amount, err := tokenTransfersRepository.CountByUser(ctx, networks.IDCasper, senderAddress)
			require.NoError(t, err)
			assert.EqualValues(t, 0, amount)
		})

		t.Run("Create", func(t *testing.T) {
			err := tokensRepository.Create(ctx, token)
			require.NoError(t, err)

			_, err = transactionsRepository.Create(ctx, transaction1)
			require.NoError(t, err)

			_, err = transactionsRepository.Create(ctx, transaction2)
			require.NoError(t, err)

			_, err = transactionsRepository.Create(ctx, transaction3)
			require.NoError(t, err)

			_, err = transactionsRepository.Create(ctx, transaction4)
			require.NoError(t, err)

			err = tokenTransfersRepository.Create(ctx, tokenTransfer1)
			require.NoError(t, err)

			err = tokenTransfersRepository.Create(ctx, tokenTransfer2)
			require.NoError(t, err)
		})

		t.Run("GetByNetworkAndTx", func(t *testing.T) {
			transfer, err := tokenTransfersRepository.GetByNetworkAndTx(ctx, networks.IDCasper, testHash)
			require.NoError(t, err)
			assert.EqualValues(t, tokenTransfer1, transfer)
		})

		t.Run("ListByUser", func(t *testing.T) {
			list, err := tokenTransfersRepository.ListByUser(ctx, 0, 3, senderAddress, networks.IDCasper)
			require.NoError(t, err)
			require.Len(t, list, 2)
			assert.ElementsMatch(t, []transfers.TokenTransfer{tokenTransfer1, tokenTransfer2}, list)
			for _, tokenTransfer := range list {
				if tokenTransfer.Status == transfers.StatusWaiting {
					assert.EqualValues(t, 0, tokenTransfer.OutboundTx)
				} else {
					assert.NotNil(t, tokenTransfer.OutboundTx)
				}
			}
		})

		t.Run("CountByUser", func(t *testing.T) {
			amount, err := tokenTransfersRepository.CountByUser(ctx, networks.IDCasper, senderAddress)
			require.NoError(t, err)
			assert.EqualValues(t, 2, amount)
		})
	})
}

func TestTokensDB(t *testing.T) {
	token1 := bridge.Token{
		ID:        1,
		ShortName: "TEST",
		LongName:  "TEST TOKEN",
	}
	token2 := bridge.Token{
		ID:        2,
		ShortName: "TESTONIUM",
		LongName:  "TESTONIUM TOKEN",
	}

	casperContractAddress, err := networks.StringToBytes(networks.IDCasper, "0xF035b4f54B4A5Dd154cD378ce9d77a12A2c97764")
	require.NoError(t, err)
	eth1ContractAddress, err := networks.StringToBytes(networks.IDEth, "0x0E26df2BaaFBC976a104EE3cbcf1B467ff1b7a69")
	require.NoError(t, err)
	solana2ContractAddress, err := networks.StringToBytes(networks.IDSolana, "7S3P4HxJpyyigGzodYwHtCxZyUQe9JiBMHyRWXArAaKv")
	require.NoError(t, err)
	eth2ContractAddress, err := networks.StringToBytes(networks.IDEth, "0x5b4f54B4A5Dd1546a104EE3cbcf1B467ff1b7a69")
	require.NoError(t, err)

	networkTokenCasper1 := networks.NetworkToken{
		NetworkID:       networks.IDCasper,
		TokenID:         token1.ID,
		ContractAddress: casperContractAddress,
		Decimals:        18,
	}
	networkTokenEth1 := networks.NetworkToken{
		NetworkID:       networks.IDEth,
		TokenID:         token1.ID,
		ContractAddress: eth1ContractAddress,
		Decimals:        18,
	}
	networkTokenSolana2 := networks.NetworkToken{
		NetworkID:       networks.IDSolana,
		TokenID:         token2.ID,
		ContractAddress: solana2ContractAddress,
		Decimals:        16,
	}
	networkTokenEth2 := networks.NetworkToken{
		NetworkID:       networks.IDEth,
		TokenID:         token2.ID,
		ContractAddress: eth2ContractAddress,
		Decimals:        16,
	}

	dbtesting.Run(t, func(ctx context.Context, t *testing.T, db bridge.DB) {
		repository := db.Tokens()

		t.Run("Negative Get", func(t *testing.T) {
			_, err := repository.Get(ctx, 0)
			require.Error(t, err)
			require.True(t, errors.Is(err, bridge.ErrNoToken))
		})

		t.Run("Negative Update", func(t *testing.T) {
			token1.ShortName = "test"
			err := repository.Update(ctx, token1)
			require.Error(t, err)
			require.True(t, errors.Is(err, bridge.ErrNoToken))
		})

		t.Run("Create", func(t *testing.T) {
			err := repository.Create(ctx, token1)
			require.NoError(t, err)
		})

		t.Run("Get", func(t *testing.T) {
			tokenFromDB, err := repository.Get(ctx, token1.ID)
			require.NoError(t, err)
			assert.Equal(t, token1.ID, tokenFromDB.ID)
			assert.Equal(t, token1.ShortName, tokenFromDB.ShortName)
			assert.Equal(t, token1.LongName, tokenFromDB.LongName)
		})

		t.Run("Update", func(t *testing.T) {
			token1.ShortName = "test2"
			err := repository.Update(ctx, token1)
			require.NoError(t, err)
		})
	})

	dbtesting.Run(t, func(ctx context.Context, t *testing.T, db bridge.DB) {
		tokensRepository := db.Tokens()
		networkTokensRepository := db.NetworkTokens()

		t.Run("Empty List", func(t *testing.T) {
			list, err := tokensRepository.List(ctx, networks.IDEth)
			require.NoError(t, err)
			assert.Empty(t, list)

			list, err = tokensRepository.List(ctx, networks.IDSolana)
			require.NoError(t, err)
			assert.Empty(t, list)
		})

		t.Run("Create", func(t *testing.T) {
			err := tokensRepository.Create(ctx, token1)
			require.NoError(t, err)

			err = tokensRepository.Create(ctx, token2)
			require.NoError(t, err)

			err = networkTokensRepository.Create(ctx, networkTokenCasper1)
			require.NoError(t, err)

			err = networkTokensRepository.Create(ctx, networkTokenEth1)
			require.NoError(t, err)

			err = networkTokensRepository.Create(ctx, networkTokenSolana2)
			require.NoError(t, err)

			err = networkTokensRepository.Create(ctx, networkTokenEth2)
			require.NoError(t, err)
		})

		t.Run("List", func(t *testing.T) {
			list, err := tokensRepository.List(ctx, networks.IDEth)
			require.NoError(t, err)
			assert.Len(t, list, 2)
			assert.EqualValues(t, []bridge.Token{token1, token2}, list)

			list, err = tokensRepository.List(ctx, networks.IDSolana)
			require.NoError(t, err)
			assert.Len(t, list, 1)
			assert.EqualValues(t, token2, list[0])
		})
	})
}

func TestTransactionsDB(t *testing.T) {
	transaction := transactions.Transaction{
		ID:          1,
		NetworkID:   networks.IDCasper,
		TxHash:      []byte{},
		Sender:      []byte{},
		BlockNumber: 1,
		SeenAt:      time.Now(),
	}

	dbtesting.Run(t, func(ctx context.Context, t *testing.T, db bridge.DB) {
		repository := db.Transactions()

		t.Run("Negative Get", func(t *testing.T) {
			_, err := repository.Get(ctx, 0)
			require.Error(t, err)
			require.True(t, errors.Is(err, bridge.ErrNoTransaction))
		})

		t.Run("Negative Exists", func(t *testing.T) {
			err := repository.Exists(ctx, transaction.NetworkID, transaction.TxHash)
			require.NoError(t, err)
		})

		t.Run("Create", func(t *testing.T) {
			_, err := repository.Create(ctx, transaction)
			require.NoError(t, err)
		})

		t.Run("Positive Exists", func(t *testing.T) {
			err := repository.Exists(ctx, transaction.NetworkID, transaction.TxHash)
			require.Error(t, err)
			assert.True(t, errors.Is(err, bridge.ErrTransactionAlreadyExists))
		})

		t.Run("Get", func(t *testing.T) {
			transactionFromDB, err := repository.Get(ctx, transaction.ID)
			require.NoError(t, err)
			assert.Equal(t, transaction.NetworkID, transactionFromDB.NetworkID)
			assert.Equal(t, transaction.TxHash, transactionFromDB.TxHash)
			assert.Equal(t, transaction.Sender, transactionFromDB.Sender)
			assert.Equal(t, transaction.BlockNumber, transactionFromDB.BlockNumber)
			assert.NotEmpty(t, transaction.SeenAt)
		})
	})
}
