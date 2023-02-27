// Copyright (C) 2022 Creditor Corp. Group.
// See LICENSE for copying information.

package controllers_test

import (
	"context"
	"encoding/hex"
	"math/big"
	"testing"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/emptypb"

	pb_networks "github.com/BoostyLabs/golden-gate-communication/go-gen/networks"
	pb_transfers "github.com/BoostyLabs/golden-gate-communication/go-gen/transfers"

	"tricorn/bridge"
	"tricorn/bridge/networks"
	"tricorn/bridge/server/controllers/apitesting"
	"tricorn/bridge/transactions"
	"tricorn/bridge/transfers"
)

func TestGateway(t *testing.T) {
	casperNetworkID := networks.IDCasperTest
	networkNonceCasper := networks.NetworkNonce{
		NetworkID: casperNetworkID,
		Nonce:     1,
	}
	token := bridge.Token{
		ID:        1,
		ShortName: "TESTONIUM",
		LongName:  "TESTONIUM TOKEN",
	}
	casperContractAddress, err := networks.StringToBytes(networks.IDCasper, "0xF035b4f54B4A5Dd154cD378ce9d77a12A2c97764")
	require.NoError(t, err)
	networkTokenCasper := networks.NetworkToken{
		NetworkID:       casperNetworkID,
		TokenID:         token.ID,
		ContractAddress: casperContractAddress,
		Decimals:        18,
	}

	senderNetwork := networks.NameCasperTest
	senderAddress, err := hex.DecodeString("4zXwdbUDWo1S5AP2CEfv4zAPRds5PQUG1dyqLLvib2xu")
	require.Error(t, err)

	recipientNetwork := networks.NameGoerli
	recipientAddress, err := hex.DecodeString("0x9744bC7A2D91928017E1DEdf98Ff7d912d6Cd263")
	require.Error(t, err)

	casperHash, err := networks.StringToBytes(networks.IDCasperTest, "d92baa8981a59e0d9143d3b2d51775af65e626aa795229fb438f97e11fed8651")
	require.NoError(t, err)

	txTime1 := time.Now()
	transaction1 := transactions.Transaction{
		ID:          1,
		NetworkID:   casperNetworkID,
		TxHash:      casperHash,
		Sender:      senderAddress,
		BlockNumber: 1,
		SeenAt:      txTime1,
	}

	ethHash, err := networks.StringToBytes(networks.IDGoerli, "bf4d685afb739d609924b9c316c841a6a4d996e86b363b5cfb4386c9554144a6")
	require.NoError(t, err)

	txTime2 := txTime1.Add(5 * time.Second)
	transaction2 := transactions.Transaction{
		ID:          2,
		NetworkID:   networks.IDGoerli,
		TxHash:      ethHash,
		Sender:      []byte{},
		BlockNumber: 1,
		SeenAt:      txTime2,
	}
	tokenTransfer := transfers.TokenTransfer{
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

	err = godotenv.Overload("./apitesting/configs/.test.gateway.env")
	if err != nil {
		t.Fatalf("could not load gateway testing file: %v", err)
	}

	config := new(apitesting.Config)
	err = env.Parse(config)
	if err != nil {
		t.Fatalf("could not parse config: %v", err)
	}

	apitesting.GatewayRun(t, func(ctx context.Context, t *testing.T, db bridge.DB) {
		gatewayClient, err := apitesting.ConnectToGateway(config.GrpcServerAddress)
		require.NoError(t, err, "can't create gateway client")

		t.Run("Seed", func(t *testing.T) {
			err = db.Nonces().Create(ctx, networkNonceCasper)
			require.NoError(t, err)

			err = db.NetworkTokens().Create(ctx, networkTokenCasper)
			require.NoError(t, err)

			err = db.Tokens().Create(ctx, token)
			require.NoError(t, err)

			transaction1.ID, err = db.Transactions().Create(ctx, transaction1)
			require.NoError(t, err)

			transaction2.ID, err = db.Transactions().Create(ctx, transaction2)
			require.NoError(t, err)

			err = db.TokenTransfers().Create(ctx, tokenTransfer)
			require.NoError(t, err)
		})

		t.Run("ConnectedNetworks", func(t *testing.T) {
			connectedNetworksResponse, err := gatewayClient.ConnectedNetworks(ctx, new(emptypb.Empty))
			require.NoError(t, err)
			require.NotNil(t, connectedNetworksResponse)
			require.NotNil(t, connectedNetworksResponse.Networks)
		})

		t.Run("SupportedTokens", func(t *testing.T) {
			supportedTokensResponse, err := gatewayClient.SupportedTokens(ctx, &pb_networks.SupportedTokensRequest{NetworkId: uint32(casperNetworkID)})
			require.NoError(t, err)
			require.NotNil(t, supportedTokensResponse)
			require.NotNil(t, supportedTokensResponse.Tokens)
		})

		t.Run("EstimateTransfer", func(t *testing.T) {
			amount := "1000"
			estimateTransferResponse, err := gatewayClient.EstimateTransfer(ctx, &pb_transfers.EstimateTransferRequest{
				SenderNetwork:    senderNetwork.String(),
				RecipientNetwork: senderNetwork.String(),
				TokenId:          uint32(token.ID),
				Amount:           amount,
			})
			require.NoError(t, err)
			require.NotNil(t, estimateTransferResponse)
			assert.NotEmpty(t, estimateTransferResponse)
		})

		// TODO: Extent test with Solana network after txHash type change.
		t.Run("Transfer", func(t *testing.T) {
			transferResponse, err := gatewayClient.Transfer(ctx, &pb_transfers.TransferRequest{
				TxHash: &pb_transfers.StringTxHash{
					NetworkName: senderNetwork.String(),
					Hash:        networks.BytesToString(casperNetworkID, casperHash),
				},
			})
			require.NoError(t, err)
			assert.NotNil(t, transferResponse)
		})

		t.Run("TransferHistory", func(t *testing.T) {
			publicKey, err := hex.DecodeString("01eb6db16548f388fe35b542bccb2ba58284c99cb53d3fc8e8c596c7be1ba2146c")
			require.NoError(t, err)

			transferHistoryResponse, err := gatewayClient.TransferHistory(ctx, &pb_transfers.TransferHistoryRequest{
				Offset:        0,
				Limit:         3,
				UserSignature: nil,
				NetworkId:     uint32(transaction1.NetworkID),
				PublicKey:     publicKey,
			})
			require.NoError(t, err)
			require.NotNil(t, transferHistoryResponse)
		})

		t.Run("BridgeInSignature", func(t *testing.T) {
			amount := "1000"
			bridgeInSignatureResponse, err := gatewayClient.BridgeInSignature(ctx, &pb_transfers.BridgeInSignatureRequest{
				Sender: &pb_transfers.StringNetworkAddress{
					NetworkName: senderNetwork.String(),
					Address:     networks.BytesToString(transaction1.NetworkID, senderAddress),
				},
				TokenId: uint32(token.ID),
				Amount:  amount,
				Destination: &pb_transfers.StringNetworkAddress{
					NetworkName: recipientNetwork.String(),
					Address:     networks.BytesToString(transaction2.NetworkID, recipientAddress),
				},
			})
			require.NoError(t, err)
			assert.NotNil(t, bridgeInSignatureResponse)
			assert.NotEmpty(t, bridgeInSignatureResponse)
		})

		t.Run("Negative CancelTransfer", func(t *testing.T) {
			cancelTransferResponse, err := gatewayClient.CancelTransfer(ctx, &pb_transfers.CancelTransferRequest{
				TransferId: uint64(tokenTransfer.ID),
				Signature:  transaction1.TxHash,
				NetworkId:  uint32(tokenTransfer.SenderNetworkID),
				PublicKey:  tokenTransfer.SenderAddress,
			})
			require.Error(t, err)
			assert.ErrorContains(t, err, bridge.ErrInvalidTransferStatus.Error())
			assert.Empty(t, cancelTransferResponse)
		})

		t.Run("CancelTransfer", func(t *testing.T) {
			tokenTransfer.Status = transfers.StatusWaiting
			err := db.TokenTransfers().Update(ctx, tokenTransfer)
			require.NoError(t, err)

			cancelTransferResponse, err := gatewayClient.CancelTransfer(ctx, &pb_transfers.CancelTransferRequest{
				TransferId: uint64(tokenTransfer.ID),
				Signature:  transaction1.TxHash,
				NetworkId:  uint32(tokenTransfer.SenderNetworkID),
				PublicKey:  tokenTransfer.SenderAddress,
			})
			require.NoError(t, err)
			assert.NotNil(t, cancelTransferResponse)
			assert.NotEmpty(t, cancelTransferResponse)
		})
	})
}
