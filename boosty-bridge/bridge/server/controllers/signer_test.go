package controllers_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	networkspb "github.com/BoostyLabs/casper-eth-bridge/boosty-communication/go-gen/networks"
	signerpb "github.com/BoostyLabs/casper-eth-bridge/boosty-communication/go-gen/signer"

	"tricorn/bridge"
	"tricorn/bridge/server/controllers/apitesting"
)

// TestConnectorBridge tests calls which directed to bridge from connectors,
// all this calls proxies to signer microservice.
func TestConnectorBridge(t *testing.T) {
	err := godotenv.Overload("./apitesting/configs/.test.bridge.env")
	if err != nil {
		t.Fatalf("could not load gateway testing file: %v", err)
	}

	config := new(apitesting.Config)
	err = env.Parse(config)
	if err != nil {
		t.Fatalf("could not parse config: %v", err)
	}

	const (
		invalidNetworkID = 100
	)
	errMsgForInvalidNetworkID := fmt.Sprintf("invalid network type %v", invalidNetworkID)

	apitesting.BridgeRun(t, func(ctx context.Context, t *testing.T, db bridge.DB) {
		connectorBridge, err := apitesting.ConnectToBridge(config.GrpcServerAddress)
		require.NoError(t, err)

		t.Run("sign", func(t *testing.T) {
			_, err := connectorBridge.Sign(ctx, &signerpb.SignRequest{
				NetworkId: networkspb.NetworkType_NT_EVM,
				Data:      []byte("some_data_here"),
			})
			require.NoError(t, err)
		})

		t.Run("sign data for unknown network(negative)", func(t *testing.T) {
			_, err := connectorBridge.Sign(ctx, &signerpb.SignRequest{
				NetworkId: networkspb.NetworkType(invalidNetworkID),
				Data:      []byte("some_data_here"),
			})
			require.Error(t, err)
			assert.True(t, status.Code(err) == codes.InvalidArgument)
			assert.Contains(t, err.Error(), errMsgForInvalidNetworkID)
		})

		t.Run("sign empty data (negative)", func(t *testing.T) {
			_, err := connectorBridge.Sign(ctx, &signerpb.SignRequest{
				NetworkId: networkspb.NetworkType_NT_EVM,
				Data:      []byte{},
			})
			require.Error(t, err)
			assert.True(t, status.Code(err) == codes.InvalidArgument)
			assert.Contains(t, err.Error(), "empty data to sign")
		})

		t.Run("public key", func(t *testing.T) {
			_, err := connectorBridge.PublicKey(ctx, &signerpb.PublicKeyRequest{
				NetworkId: networkspb.NetworkType_NT_EVM,
			})
			require.NoError(t, err)
		})

		t.Run("public key for unknown network(negative)", func(t *testing.T) {
			_, err := connectorBridge.PublicKey(ctx, &signerpb.PublicKeyRequest{
				NetworkId: networkspb.NetworkType(invalidNetworkID),
			})
			require.Error(t, err)
			assert.True(t, status.Code(err) == codes.InvalidArgument)
			assert.Contains(t, err.Error(), errMsgForInvalidNetworkID)
		})
	})
}
