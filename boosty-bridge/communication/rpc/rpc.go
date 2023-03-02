// Copyright (C) 2022 Creditor Corp. Group.
// See LICENSE for copying information.

package rpc

import (
	"context"
	"time"

	"github.com/zeebo/errs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"

	bridgeconnectorpb "github.com/BoostyLabs/casper-eth-bridge/boosty-communication/go-gen/bridge-connector"
	bridgeoraclepb "github.com/BoostyLabs/casper-eth-bridge/boosty-communication/go-gen/bridge-oracle"
	bridgesignerpb "github.com/BoostyLabs/casper-eth-bridge/boosty-communication/go-gen/bridge-signer"
	connectorbridgepb "github.com/BoostyLabs/casper-eth-bridge/boosty-communication/go-gen/connector-bridge"
	gatewaybridgepb "github.com/BoostyLabs/casper-eth-bridge/boosty-communication/go-gen/gateway-bridge"

	"tricorn/bridge"
	"tricorn/bridge/networks"
	"tricorn/bridge/transfers"
	"tricorn/chains"
	"tricorn/communication"
	"tricorn/internal/logger"
)

// Error is default communication error type.
var Error = errs.Class("communication: rpc")

// Config hold configuration for GRPC implementation of communication.Communication protocol.
type Config struct {
	ServerAddress string `env:"SERVER_TO_CONNECT_ADDRESS"`

	PingServerTime    time.Duration `env:"PING_SERVER_TIME" help:"defines that we will ping server n seconds"`
	PingServerTimeout time.Duration `env:"PING_SERVER_TIMEOUT" help:"defines time for response from server after ping call."`
}

// ensures that rpc implements connector.Communication.
var _ communication.Communication = (*rpc)(nil)

// rpc combines access to different rpc modules methods.
type rpc struct {
	log logger.Logger

	cfg            Config
	connWithServer *grpc.ClientConn

	isConnected bool
}

// New is a constructor for grpc implementation of communication.Communication.
func New(cfg Config, log logger.Logger, withPing bool) (communication.Communication, error) {
	ctx, cancel := context.WithTimeout(context.Background(), cfg.PingServerTime)
	defer cancel()

	rpc := &rpc{
		cfg:            cfg,
		connWithServer: nil,
		log:            log,
		isConnected:    false,
	}

	if withPing {
		return rpc, rpc.ConnectWithPing(ctx)
	}

	return rpc, rpc.Connect(ctx)
}

// Networks provides access to the networks.Bridge rpc methods.
func (rpc *rpc) Networks() networks.Bridge {
	return &networksRPC{
		client:      gatewaybridgepb.NewGatewayBridgeClient(rpc.connWithServer),
		isConnected: rpc.isConnected,
	}
}

// Transfers provides access to the transfers.Bridge rpc methods.
func (rpc *rpc) Transfers() transfers.Bridge {
	return &transfersRPC{
		client:      gatewaybridgepb.NewGatewayBridgeClient(rpc.connWithServer),
		isConnected: rpc.isConnected,
	}
}

// Bridge provides access to the chains.Bridge rpc methods.
func (rpc *rpc) Bridge() chains.Bridge {
	return &bridgeRPC{
		client:      connectorbridgepb.NewConnectorBridgeClient(rpc.connWithServer),
		isConnected: rpc.isConnected,
	}
}

// Connector provides access to the bridge.Connector rpc methods.
func (rpc *rpc) Connector(ctx context.Context) bridge.Connector {
	return &connectorRPC{
		gctx:             ctx,
		client:           bridgeconnectorpb.NewConnectorClient(rpc.connWithServer),
		eventSubscribers: make([]bridge.EventSubscriber, 0),
	}
}

// CurrencyRates provides access to the bridge.CurrencyRates rpc methods.
func (rpc *rpc) CurrencyRates() bridge.CurrencyRates {
	return &currencyratesRPC{client: bridgeoraclepb.NewBridgeOracleClient(rpc.connWithServer)}
}

// Signer provides access to bridge.Signer rpc methods.
func (rpc *rpc) Signer() bridge.Signer {
	return &signerRPC{client: bridgesignerpb.NewBridgeSignerClient(rpc.connWithServer)}
}

// ConnectWithPing will try to establish connection which pings the server every interval.
func (rpc *rpc) ConnectWithPing(ctx context.Context) error {
	dialOpts := []grpc.DialOption{
		grpc.WithAuthority(rpc.cfg.ServerAddress),
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                rpc.cfg.PingServerTime,
			Timeout:             rpc.cfg.PingServerTimeout,
			PermitWithoutStream: true,
		}),
	}
	connWithServer, err := grpc.DialContext(ctx, rpc.cfg.ServerAddress, dialOpts...)
	if err != nil {
		rpc.log.Debug("could not establish grpc connection with server on " + rpc.cfg.ServerAddress)
		return Error.Wrap(err)
	}

	rpc.log.Debug("established grpc connection with server on " + rpc.cfg.ServerAddress)
	rpc.connWithServer = connWithServer
	rpc.isConnected = true
	return nil
}

// Connect will try to establish connection.
func (rpc *rpc) Connect(ctx context.Context) error {
	dialOpts := []grpc.DialOption{
		grpc.WithAuthority(rpc.cfg.ServerAddress),
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	connWithServer, err := grpc.DialContext(ctx, rpc.cfg.ServerAddress, dialOpts...)
	if err != nil {
		rpc.log.Debug("could not establish grpc connection with server on " + rpc.cfg.ServerAddress)
		return Error.Wrap(err)
	}

	rpc.log.Debug("established grpc connection with server on " + rpc.cfg.ServerAddress)
	rpc.connWithServer = connWithServer
	rpc.isConnected = true
	return nil
}

// Close closes underlying rpc connection.
func (rpc *rpc) Close() error {
	return Error.Wrap(rpc.connWithServer.Close())
}
