// Copyright (C) 2022 Creditor Corp. Group.
// See LICENSE for copying information.

package rpc

import (
	"context"

	connectorbridgepb "github.com/BoostyLabs/casper-eth-bridge/boosty-communication/go-gen/connector-bridge"
	networkspb "github.com/BoostyLabs/casper-eth-bridge/boosty-communication/go-gen/networks"
	signerpb "github.com/BoostyLabs/casper-eth-bridge/boosty-communication/go-gen/signer"

	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/chains"
	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/networks"
)

// bridgeRPC provides access to the Signer.
type bridgeRPC struct {
	client connectorbridgepb.ConnectorBridgeClient
}

// Sign returns signed data for specific network.
func (bridgeRPC *bridgeRPC) Sign(ctx context.Context, req chains.SignRequest) ([]byte, error) {
	in := signerpb.SignRequest{
		NetworkId: networkspb.NetworkType(networks.NetworkTypeToNetworkID[req.NetworkId]),
		Data:      req.Data,
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
