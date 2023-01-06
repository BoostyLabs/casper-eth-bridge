// Copyright (C) 2022 Creditor Corp. Group.
// See LICENSE for copying information.

package rpc

import (
	"context"
	"time"

	connectorbridgepb "github.com/BoostyLabs/casper-eth-bridge/boosty-communication/go-gen/connector-bridge"
	gatewaybridgepb "github.com/BoostyLabs/casper-eth-bridge/boosty-communication/go-gen/gateway-bridge"
	"github.com/zeebo/errs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"

	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/chains"
	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/communication"
	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/networks"
	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/pkg/logger"
	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/transfers"
)

// Error is default communication error type.
var Error = errs.Class("communication: rpc")

// Config hold configuration for GRPC implementation of communication.Communication protocol.
type Config struct {
	BridgeAddress string `env:"BRIDGE_ADDRESS"`

	PingServerTime    time.Duration `env:"PING_SERVER_TIME" help:"defines that we will ping server n seconds"`
	PingServerTimeout time.Duration `env:"PING_SERVER_TIMEOUT" help:"defines time for response from server after ping call."`
}

// ensures that rpc implements connector.Communication.
var _ communication.Communication = (*rpc)(nil)

// rpc combines access to different rpc modules methods.
type rpc struct {
	log logger.Logger

	bridgeConn *grpc.ClientConn
}

// New is a constructor for rpc.
func New(cfg Config, log logger.Logger) (communication.Communication, error) {
	ctx, cancel := context.WithTimeout(context.Background(), cfg.PingServerTime)
	defer cancel()

	bridgeOpts := []grpc.DialOption{
		grpc.WithAuthority(cfg.BridgeAddress),
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                cfg.PingServerTime,
			Timeout:             cfg.PingServerTimeout,
			PermitWithoutStream: true,
		}),
	}

	bridgeConn, err := grpc.DialContext(ctx, cfg.BridgeAddress, bridgeOpts...)
	if err != nil {
		return nil, Error.Wrap(err)
	}
	log.Debug("established grpc connection with bridge on " + cfg.BridgeAddress)

	return &rpc{
		bridgeConn: bridgeConn,
		log:        log,
	}, nil
}

// Networks provides access to the networks.Bridge rpc methods.
func (rpc *rpc) Networks() networks.Bridge {
	return &networksRPC{client: gatewaybridgepb.NewGatewayBridgeClient(rpc.bridgeConn)}
}

// Transfers provides access to the transfers.Bridge rpc methods.
func (rpc *rpc) Transfers() transfers.Bridge {
	return &transfersRPC{client: gatewaybridgepb.NewGatewayBridgeClient(rpc.bridgeConn)}
}

// Bridge provides access to the chains.Bridge rpc methods.
func (rpc *rpc) Bridge() chains.Bridge {
	return &bridgeRPC{client: connectorbridgepb.NewConnectorBridgeClient(rpc.bridgeConn)}
}

// Close closes underlying rpc connection.
func (rpc *rpc) Close() error {
	return Error.Wrap(rpc.bridgeConn.Close())
}
