// Copyright (C) 2022 Creditor Corp. Group.
// See LICENSE for copying information.

package rpc

import (
	"context"
	"fmt"

	bridgesignerpb "github.com/BoostyLabs/casper-eth-bridge/boosty-communication/go-gen/bridge-signer"
	networkspb "github.com/BoostyLabs/casper-eth-bridge/boosty-communication/go-gen/networks"
	signerpb "github.com/BoostyLabs/casper-eth-bridge/boosty-communication/go-gen/signer"

	"tricorn/bridge/networks"
	"tricorn/signer"
)

// signerRPC provides access to the Signer.
type signerRPC struct {
	client bridgesignerpb.BridgeSignerClient
}

// Sign returns signed data in specific network.
func (signerRPC *signerRPC) Sign(ctx context.Context, networkType networks.Type, data []byte, dataType signer.Type) ([]byte, error) {
	pbNetworkType, err := networksToProto(networkType)
	if err != nil {
		return nil, err
	}

	resp, err := signerRPC.client.Sign(ctx, &signerpb.SignRequest{
		NetworkId: pbNetworkType,
		Data:      data,
		DataType:  signerpb.DataType(signerpb.DataType_value[dataType.String()]),
	})
	if err != nil {
		return nil, err
	}

	return resp.GetSignature(), nil
}

// PublicKey returns public key in specific network.
func (signerRPC *signerRPC) PublicKey(ctx context.Context, networkType networks.Type) (networks.PublicKey, error) {
	pbNetworkType, err := networksToProto(networkType)
	if err != nil {
		return nil, err
	}

	resp, err := signerRPC.client.PublicKey(ctx, &signerpb.PublicKeyRequest{
		NetworkId: pbNetworkType,
	})
	if err != nil {
		return nil, err
	}

	return resp.GetPublicKey(), nil
}

// networksToProto casts internal network type to proto one.
func networksToProto(networkType networks.Type) (networkspb.NetworkType, error) {
	switch networkType {
	case networks.TypeEVM:
		return networkspb.NetworkType_NT_EVM, nil
	case networks.TypeCasper:
		return networkspb.NetworkType_NT_CASPER, nil
	case networks.TypeSolana:
		return networkspb.NetworkType_NT_SOLANA, nil
	default:
		return 0, fmt.Errorf("invalid network type %v", networkType)
	}
}
