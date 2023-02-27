// Copyright (C) 2022 Creditor Corp. Group.
// See LICENSE for copying information.

package rpc

import (
	"context"

	connectorbridgepb "github.com/BoostyLabs/casper-eth-bridge/boosty-communication/go-gen/connector-bridge"
	networkspb "github.com/BoostyLabs/casper-eth-bridge/boosty-communication/go-gen/networks"
	signerpb "github.com/BoostyLabs/casper-eth-bridge/boosty-communication/go-gen/signer"

	"tricorn/bridge/networks"
	"tricorn/chains"
)

// bridgeRPC provides access to the Signer.
type bridgeRPC struct {
	isConnected bool
	client      connectorbridgepb.ConnectorBridgeClient
}

// Sign returns signed data for specific network.
func (bridgeRPC *bridgeRPC) Sign(ctx context.Context, req chains.SignRequest) ([]byte, error) {
	in := signerpb.SignRequest{
		NetworkId: networkspb.NetworkType(networks.NetworkTypeToNetworkID[req.NetworkId]),
		Data:      req.Data,
		DataType:  signerpb.DataType(signerpb.DataType_value[req.DataType.String()]),
	}
	singResponse, err := bridgeRPC.client.Sign(ctx, &in)
	if err != nil {
		return nil, err
	}

	return singResponse.Signature, nil
}

// PublicKey returns public key for specific network.
func (bridgeRPC *bridgeRPC) PublicKey(ctx context.Context, networkId networks.Type) ([]byte, error) {
	in := signerpb.PublicKeyRequest{
		NetworkId: networkspb.NetworkType(networks.NetworkTypeToNetworkID[networkId]),
	}
	grpcResponse, err := bridgeRPC.client.PublicKey(ctx, &in)
	if err != nil {
		return nil, err
	}

	return grpcResponse.PublicKey, nil
}
