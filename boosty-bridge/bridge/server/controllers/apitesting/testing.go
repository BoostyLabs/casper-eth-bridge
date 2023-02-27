// Copyright (C) 2022 Creditor Corp. Group.
// See LICENSE for copying information.

package apitesting

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"

	connectorbridgepb "github.com/BoostyLabs/golden-gate-communication/go-gen/connector-bridge"
	pb_gateway_bridge "github.com/BoostyLabs/golden-gate-communication/go-gen/gateway-bridge"

	peer "tricorn"
	"tricorn/bridge"
	"tricorn/bridge/database/dbtesting"
	"tricorn/bridge/networks"
	"tricorn/bridge/server/controllers"
	"tricorn/bridge/transfers"
	"tricorn/chains"
	"tricorn/communication/mockcommunication"
	"tricorn/internal/logger/zaplog"
	grpc_server "tricorn/internal/server/grpc"
)

// Config contains configurable values for gateway and bridge microservices.
type Config struct {
	Database          string `env:"DATABASE"`
	GrpcServerAddress string `env:"GRPC_SERVER_ADDRESS"`
	ServerName        string `env:"SERVER_NAME"`
}

func GatewayRun(t *testing.T, test func(ctx context.Context, t *testing.T, db bridge.DB)) {
	ctx, cancel := context.WithCancel(context.Background())
	log := zaplog.NewLog()

	err := godotenv.Overload("./apitesting/configs/.test.gateway.env")
	if err != nil {
		t.Fatalf("could not load gateway testing file: %v", err)
	}

	config := new(Config)
	err = env.Parse(config)
	if err != nil {
		t.Fatalf("could not parse config: %v", err)
	}

	masterDB := dbtesting.Database{
		Name: "Postgres",
		URL:  config.Database,
	}

	db, err := dbtesting.CreateMasterDB(ctx, t.Name(), "Test", 0, masterDB)
	require.NoError(t, err)
	defer func() {
		err := db.Close()
		require.NoError(t, err)
	}()

	err = db.CreateSchema(ctx)
	require.NoError(t, err)

	service := bridge.New(
		log,
		mockSigner(),
		db.Nonces(),
		db.NetworkTokens(),
		db.Tokens(),
		db.Transactions(),
		db.TokenTransfers(),
		db.NetworkBlocks(),
	)

	casperConnector := getMockConnector()
	ethConnector := getMockConnector()

	connectors := map[networks.Name]bridge.Connector{
		networks.NameCasperTest: casperConnector,
		networks.NameGoerli:     ethConnector,
	}

	for name, connector := range connectors {
		service.AddConnector(context.Background(), name, connector)
	}

	controller := controllers.NewGateway(log, service)

	registerServer := func(grpcServer *grpc.Server) {
		pb_gateway_bridge.RegisterGatewayBridgeServer(grpcServer, controller)
	}

	server := grpc_server.NewServer(log, registerServer, config.ServerName, config.GrpcServerAddress)

	gateway := peer.New(log, nil, nil, server, config.ServerName)

	var group errgroup.Group
	group.Go(func() error {
		return gateway.Run(ctx)
	})

	time.Sleep(time.Second)

	group.Go(func() error {
		test(ctx, t, db)
		cancel()
		return nil
	})

	err = group.Wait()
	if err != nil {
		log.Error("could not run test/server", err)
		return
	}
}

func BridgeRun(t *testing.T, test func(ctx context.Context, t *testing.T, db bridge.DB)) {
	ctx, cancel := context.WithCancel(context.Background())
	log := zaplog.NewLog()

	err := godotenv.Overload("./apitesting/configs/.test.bridge.env")
	if err != nil {
		t.Fatalf("could not load gateway testing file: %v", err)
	}

	config := new(Config)
	err = env.Parse(config)
	if err != nil {
		t.Fatalf("could not parse config: %v", err)
	}

	masterDB := dbtesting.Database{
		Name: "Postgres",
		URL:  config.Database,
	}

	db, err := dbtesting.CreateMasterDB(ctx, t.Name(), "Test", 0, masterDB)
	require.NoError(t, err)
	defer func() {
		err := db.Close()
		require.NoError(t, err)
	}()

	err = db.CreateSchema(ctx)
	require.NoError(t, err)

	service := bridge.New(
		log, mockSigner(),
		db.Nonces(),
		db.NetworkTokens(),
		db.Tokens(),
		db.Transactions(),
		db.TokenTransfers(),
		db.NetworkBlocks(),
	)

	casperConnector := getMockConnector()
	ethConnector := getMockConnector()

	connectors := map[networks.Name]bridge.Connector{
		networks.NameCasperTest: casperConnector,
		networks.NameGoerli:     ethConnector,
	}

	for name, connector := range connectors {
		service.AddConnector(context.Background(), name, connector)
	}

	gateway := controllers.NewGateway(log, service)
	signer := controllers.NewSigner(service)

	registerServer := func(grpcServer *grpc.Server) {
		pb_gateway_bridge.RegisterGatewayBridgeServer(grpcServer, gateway)
		connectorbridgepb.RegisterConnectorBridgeServer(grpcServer, signer)
	}

	server := grpc_server.NewServer(log, registerServer, config.ServerName, config.GrpcServerAddress)

	var group errgroup.Group
	group.Go(func() error {
		return server.Run(ctx)
	})

	time.Sleep(time.Second)

	group.Go(func() error {
		test(ctx, t, db)
		cancel()
		return nil
	})

	err = group.Wait()
	if err != nil {
		log.Error("could not run test/server", err)
		return
	}
}

func getMockConnector() bridge.Connector {
	connector := new(mockcommunication.ConnectorMock)
	connector.SetEstimateTransfer(func(ctx context.Context, req transfers.EstimateTransfer) (chains.Estimation, error) {
		return chains.Estimation{Fee: "1000", FeePercentage: "12", EstimatedConfirmation: 123}, nil
	})
	connector.SetBridgeInSignature(func(ctx context.Context, req bridge.BridgeInSignatureRequest) (bridge.BridgeInSignatureResponse, error) {
		return bridge.BridgeInSignatureResponse{Nonce: new(big.Int)}, nil
	})
	connector.SetCancelSignature(func(ctx context.Context, req chains.CancelSignatureRequest) (chains.CancelSignatureResponse, error) {
		return chains.CancelSignatureResponse{}, nil
	})
	connector.SetAddEventSubscriber(func() bridge.EventSubscriber {
		return bridge.EventSubscriber{}
	})
	connector.SetEventStream(func(ctx context.Context, fromBlock uint64) error {
		return nil
	})
	connector.SetNetwork(func(ctx context.Context) (networks.Network, error) {
		return networks.Network{}, nil
	})

	return connector
}

func mockSigner() bridge.Signer {
	communication := mockcommunication.New()
	return communication.Signer()
}
