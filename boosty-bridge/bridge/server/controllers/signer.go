// Copyright (C) 2022 Creditor Corp. Group.
// See LICENSE for copying information.

package controllers

import (
	"context"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	connectorbridgepb "github.com/BoostyLabs/golden-gate-communication/go-gen/connector-bridge"
	networkspb "github.com/BoostyLabs/golden-gate-communication/go-gen/networks"
	signerpb "github.com/BoostyLabs/golden-gate-communication/go-gen/signer"

	"tricorn/bridge"
	"tricorn/bridge/networks"
	"tricorn/signer"
)

// ensures that Server implements connectorbridgepb.ConnectorBridgeServer.
var _ connectorbridgepb.ConnectorBridgeServer = (*Signer)(nil)

// Signer is controller that handles calls by connectors.
type Signer struct {
	bridge *bridge.Service
}

// NewSigner is Signer constructor.
func NewSigner(bridge *bridge.Service) *Signer {
	return &Signer{
		bridge: bridge,
	}
}

// Sign signs given data in specific network.
func (s *Signer) Sign(ctx context.Context, request *signerpb.SignRequest) (*signerpb.Signature, error) {
	networkType, err := networkFromProto(request.NetworkId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if len(request.Data) == 0 {
		return nil, status.Error(codes.InvalidArgument, "empty data to sign")
	}

	signedData, err := s.bridge.Sign(ctx, networkType, request.Data, signer.Type(request.GetDataType().String()))

	return &signerpb.Signature{
		NetworkId: request.NetworkId,
		Signature: signedData,
	}, err
}

// PublicKey returns public key in specific network.
func (signer *Signer) PublicKey(ctx context.Context, request *signerpb.PublicKeyRequest) (*signerpb.PublicKeyResponse, error) {
	networkType, err := networkFromProto(request.NetworkId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	publicKey, err := signer.bridge.PublicKey(ctx, networkType)

	return &signerpb.PublicKeyResponse{
		PublicKey: publicKey,
	}, err
}

// networkFromProto casts proto network type to internal one.
func networkFromProto(networkType networkspb.NetworkType) (networks.Type, error) {
	switch networkType {
	case networkspb.NetworkType_NT_EVM:
		return networks.TypeEVM, nil
	case networkspb.NetworkType_NT_CASPER:
		return networks.TypeCasper, nil
	case networkspb.NetworkType_NT_SOLANA:
		return networks.TypeSolana, nil
	default:
		return "", fmt.Errorf("invalid network type %v", networkType)
	}
}
