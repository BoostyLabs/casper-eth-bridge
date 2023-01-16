// Copyright (C) 2022 Creditor Corp. Group.
// See LICENSE for copying information.

package controllers_test

import (
	"context"
	"github.com/caarlos0/env/v6"
	"testing"

	connectorpb "github.com/BoostyLabs/casper-eth-bridge/boosty-communication/go-gen/connector"
	networkspb "github.com/BoostyLabs/casper-eth-bridge/boosty-communication/go-gen/networks"
	transferspb "github.com/BoostyLabs/casper-eth-bridge/boosty-communication/go-gen/transfers"
	"github.com/joho/godotenv"
	_ "github.com/joho/godotenv/autoload"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/chains/server/controllers/apitesting"
	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/networks"
	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/pkg/envparse"
)

func TestCasperConnector(t *testing.T) {
	err := godotenv.Overload("./apitesting/configs/.test.casper.env")
	if err != nil {
		t.Fatalf("could not load casper testing file: %v", err)
	}

	err = godotenv.Overload("./apitesting/configs/.test.env")
	if err != nil {
		t.Fatalf("could not load testing file: %v", err)
	}

	cfg := apitesting.CasperConfig{}
	envOpt := env.Options{RequiredIfNoDef: true}
	err = env.ParseWithFuncs(&cfg, envparse.EthParseOpts(), envOpt)
	if err != nil {
		t.Fatalf("could not parse config: %v", err)
	}

	apitesting.CasperRun(t, func(ctx context.Context, t *testing.T) {
		connectorClient, err := apitesting.ConnectToConnector(cfg.GrpcServerAddress)
		require.NoError(t, err, "can't create connector service client")

		t.Run("metadata", func(t *testing.T) {
			metadata, err := connectorClient.Metadata(ctx, &emptypb.Empty{})
			require.NoError(t, err)
			assert.Equal(t, networkspb.NetworkType_NT_CASPER, metadata.Ty)
			assert.EqualValues(t, networks.IDCasperTest, metadata.Id)
			assert.EqualValues(t, cfg.Config.ChainName, metadata.Name)
			assert.True(t, metadata.IsTestnet)
		})

		t.Run("known tokens", func(t *testing.T) {
			_, err := connectorClient.KnownTokens(ctx, &emptypb.Empty{})
			require.NoError(t, err)
		})

		t.Run("EventStream", func(t *testing.T) {
			_, err := connectorClient.EventStream(ctx, &connectorpb.EventsRequest{})
			require.NoError(t, err)
		})

		t.Run("estimate transfer", func(t *testing.T) {
			response, err := connectorClient.EstimateTransfer(ctx, &transferspb.EstimateTransferRequest{
				SenderNetwork:    "",
				RecipientNetwork: cfg.Config.ChainName,
				TokenId:          0,
				Amount:           "",
			})
			require.NoError(t, err)

			assert.NotNil(t, response)
			assert.Equal(t, cfg.Config.Fee, response.GetFee())
			assert.Equal(t, cfg.Config.FeePercentage, response.GetFeePercentage())
			assert.EqualValues(t, cfg.Config.EstimatedConfirmation, response.GetEstimatedConfirmation())
		})

		// for manual testing.
		// t.Run("BridgeOut", func(t *testing.T) {
		// 	token, err := hex.DecodeString("013c0c1847d1c410338ab9b4ee0919c181cf26085997ff9c797e8a1ae5b02ddf23")
		// 	require.NoError(t, err)

		// 	to, err := hex.DecodeString("009060c0820b5156b1620c8e3344d17f9fad5108f5dc2672f2308439e84363c88e")
		// 	require.NoError(t, err)

		// 	in := connectorpb.TokenOutRequest{
		// 		Amount: "1000000",
		// 		Token:  &connectorpb.Address{Address: token},
		// 		To:     &connectorpb.Address{Address: to},
		// 		From: &transferspb.StringNetworkAddress{
		// 			NetworkName: "GOERLI",
		// 			Address:     "3095f955da700b96215cffc9bc64ab2e69eb7dab",
		// 		},
		// 	}
		// 	resp, err := connectorClient.BridgeOut(ctx, &in)
		// 	require.NoError(t, err)
		// 	t.Logf("tx hash: %s", hex.EncodeToString(resp.Txhash))
		// }).
	})
}

func TestEthConnector(t *testing.T) {
	err := godotenv.Overload("./apitesting/configs/.test.eth.env")
	if err != nil {
		t.Fatalf("could not load eth testing file: %v", err)
	}

	err = godotenv.Overload("./apitesting/configs/.test.env")
	if err != nil {
		t.Fatalf("could not load testing file: %v", err)
	}

	cfg := apitesting.EthConfig{}
	envOpt := env.Options{RequiredIfNoDef: true}
	err = env.ParseWithFuncs(&cfg, envparse.EthParseOpts(), envOpt)
	if err != nil {
		t.Fatalf("could not parse config: %v", err)
	}

	apitesting.EthRun(t, func(ctx context.Context, t *testing.T) {
		connectorClient, err := apitesting.ConnectToConnector(cfg.GrpcServerAddress)
		require.NoError(t, err, "can't create connector service client")

		t.Run("metadata", func(t *testing.T) {
			metadata, err := connectorClient.Metadata(ctx, &emptypb.Empty{})
			require.NoError(t, err)
			assert.Equal(t, networkspb.NetworkType_NT_EVM, metadata.Ty)
			assert.EqualValues(t, networks.IDGoerli, metadata.Id)
			assert.EqualValues(t, cfg.Config.ChainName, metadata.Name)
			assert.True(t, metadata.IsTestnet)
		})

		t.Run("known tokens", func(t *testing.T) {
			_, err := connectorClient.KnownTokens(ctx, &emptypb.Empty{})
			require.NoError(t, err)
		})

		t.Run("estimate transfer", func(t *testing.T) {
			estimation, err := connectorClient.EstimateTransfer(ctx, &transferspb.EstimateTransferRequest{
				RecipientNetwork: "GOERLI",
			})
			require.NoError(t, err)

			require.NotNil(t, estimation)
			assert.EqualValues(t, 60, estimation.EstimatedConfirmation)
			assert.EqualValues(t, "0.4", estimation.FeePercentage)
		})

		// for manual testing with GRPC mode.
		// t.Run("bridge out", func(t *testing.T) {
		// 	token, err := hex.DecodeString("e5bfc49E60a62AB039189D14b148ABEb80403460")
		// 	require.NoError(t, err)

		// 	to, err := hex.DecodeString("7848d440a80868e4c6b03b7649832824dba5da87")
		// 	require.NoError(t, err)

		// 	_, err = connectorClient.BridgeOut(ctx, &connectorpb.TokenOutRequest{
		// 		Amount: "1",
		// 		Token:  &connectorpb.Address{Address: token},
		// 		To:     &connectorpb.Address{Address: to},
		// 		From: &transferspb.StringNetworkAddress{
		// 			NetworkName: "CASPER",
		// 			Address:     "010ad302bfc22c0e606d94d98a3baa2c8eeedd1e148d9a20a4453bb8cc5e530a19",
		// 		},
		// 	})
		// 	require.NoError(t, err)
		// }).

		// TODO: add remaining methods.
	})
}
