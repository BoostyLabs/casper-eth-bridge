// Copyright (C) 2022 Creditor Corp. Group.
// See LICENSE for copying information.

package controllers

import (
	"context"
	"errors"
	"fmt"

	"github.com/zeebo/errs"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	bridgesignerpb "github.com/BoostyLabs/golden-gate-communication/go-gen/bridge-signer"
	signerpb "github.com/BoostyLabs/golden-gate-communication/go-gen/signer"

	"tricorn/bridge/networks"
	"tricorn/internal/logger"
	"tricorn/signer"
)

// ensures that Signer implements bridgesignerpb.SignerServer.
var _ bridgesignerpb.BridgeSignerServer = (*Signer)(nil)

// Error is an internal error type for signer controller.
var Error = errs.Class("signer controller")

// Signer is controller that handles all signer related routes.
type Signer struct {
	log logger.Logger

	signer *signer.Service
}

// NewSigner is a constructor for signer controller.
func NewSigner(log logger.Logger, signer *signer.Service) *Signer {
	signerController := &Signer{
		log:    log,
		signer: signer,
	}

	return signerController
}

// Sign returns signed data for specific network.
func (s *Signer) Sign(ctx context.Context, req *signerpb.SignRequest) (*signerpb.Signature, error) {
	var (
		resp = signerpb.Signature{
			NetworkId: req.NetworkId,
		}
		err error
	)

	networkType := networks.NetworkIDToNetworkType[networks.TypeID(req.GetNetworkId())]
	if err = networkType.Validate(); err != nil {
		return &resp, status.Error(codes.InvalidArgument, Error.Wrap(err).Error())
	}

	resp.Signature, err = s.signer.Sign(ctx, networkType, req.Data, signer.Type(req.GetDataType().String()))
	if err != nil {
		if errors.Is(err, signer.ErrNoPrivateKey) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		s.log.Error(fmt.Sprintf("couldn't sign data for %s network", networkType), err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &resp, nil
}

// PublicKey returns public key for specific network.
func (s *Signer) PublicKey(ctx context.Context, req *signerpb.PublicKeyRequest) (*signerpb.PublicKeyResponse, error) {
	var (
		resp signerpb.PublicKeyResponse
		err  error
	)

	networkType := networks.NetworkIDToNetworkType[networks.TypeID(req.NetworkId)]
	if err = networkType.Validate(); err != nil {
		return &resp, status.Error(codes.InvalidArgument, Error.Wrap(err).Error())
	}

	resp.PublicKey, err = s.signer.PublicKey(ctx, networkType)
	if err != nil {
		if errors.Is(err, signer.ErrNoPrivateKey) {
			return nil, status.Error(codes.NotFound, err.Error())
		}

		s.log.Error(fmt.Sprintf("couldn't get public key for %s network", networkType), err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &resp, nil
}
