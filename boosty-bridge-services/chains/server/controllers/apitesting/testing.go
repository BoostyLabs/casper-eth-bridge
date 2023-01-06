// Copyright (C) 2022 Creditor Corp. Group.
// See LICENSE for copying information.

package apitesting

import (
	"context"
	"testing"
	"time"

	bridge_connectorpb "github.com/BoostyLabs/casper-eth-bridge/boosty-communication/go-gen/bridge-connector"
	"github.com/caarlos0/env/v6"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/joho/godotenv"
	_ "github.com/joho/godotenv/autoload"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"

	peer "github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services"
	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/chains"
	casper_service "github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/chains/casper"
	evm_service "github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/chains/evm"
	chains_server "github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/chains/server"
	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/chains/server/controllers"
	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/communication"
	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/communication/mockcommunication"
	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/communication/rpc"
	casper_mock "github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/internal/casper-sdk/mock"
	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/internal/contracts/evm/bridge"
	bridge_client "github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/internal/contracts/evm/client"
	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/networks"
	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/pkg/envparse"
	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/pkg/logger/zaplog"
)

// CasperConfig is the global configuration to run casper connector server.
type CasperConfig struct {
	GrpcServerAddress string `env:"GRPC_SERVER_ADDRESS"`
	Config            casper_service.Config
	Communication     rpc.Config
	CommunicationMode communication.Mode `env:"COMMUNICATION_MODE"`
	ServerName        string             `env:"SERVER_NAME"`
}

func CasperRun(t *testing.T, test func(ctx context.Context, t *testing.T)) {
	var (
		comm    communication.Communication
		client  casper_service.Casper
		service *casper_service.Service
		server  *chains_server.Server
	)

	ctx, cancel := context.WithCancel(context.Background())
	log := zaplog.NewLog()

	err := godotenv.Overload("./apitesting/configs/.test.casper.env")
	if err != nil {
		t.Fatalf("could not load casper testing file: %v", err)
	}

	err = godotenv.Overload("./apitesting/configs/.test.env")
	if err != nil {
		t.Fatalf("could not load testing file: %v", err)
	}

	config := new(CasperConfig)
	envOpt := env.Options{RequiredIfNoDef: true}
	err = env.ParseWithFuncs(config, envparse.EthParseOpts(), envOpt)
	if err != nil {
		t.Fatalf("could not parse config: %v", err)
	}

	// Communication setup.
	comm = mockcommunication.New()
	client = casper_mock.New()

	// Casper server setup.
	{
		service = casper_service.NewService(ctx, config.Config, log, comm.Bridge(), client)
	}

	{ // Server setup.
		controller := controllers.NewConnector(ctx, log, service)

		registerServer := func(grpcServer *grpc.Server) {
			bridge_connectorpb.RegisterConnectorServer(grpcServer, controller)
		}

		server = chains_server.NewServer(ctx, log, registerServer, config.GrpcServerAddress, config.ServerName)
	}

	connector := peer.New(log, comm, service, server, config.ServerName)

	var group errgroup.Group
	group.Go(func() error {
		return connector.Run(ctx)
	})

	time.Sleep(time.Second) // for connector initialization.

	group.Go(func() error {
		test(ctx, t)
		cancel()
		return nil
	})

	err = group.Wait()
	if err != nil {
		log.Error("could not run test/server", err)
		return
	}

	defer func() {
		cancel()
		err := connector.Close()
		if err != nil {
			log.Error("could not close connector", err)
		}
	}()
}

// EthConfig is the global configuration to run connector server.
type EthConfig struct {
	GrpcServerAddress string `env:"GRPC_SERVER_ADDRESS"`
	Config            evm_service.Config
	Communication     rpc.Config
	CommunicationMode communication.Mode `env:"COMMUNICATION_MODE"`
	ServerName        string             `env:"SERVER_NAME"`
	Bridge            bridge_client.Config
}

func EthRun(t *testing.T, test func(ctx context.Context, t *testing.T)) {
	var (
		comm    communication.Communication
		service *evm_service.Service
		server  *chains_server.Server
	)

	ctx, cancel := context.WithCancel(context.Background())
	log := zaplog.NewLog()

	err := godotenv.Overload("./apitesting/configs/.test.eth.env")
	if err != nil {
		t.Fatalf("could not load eth testing file: %v", err)
	}

	err = godotenv.Overload("./apitesting/configs/.test.env")
	if err != nil {
		t.Fatalf("could not load testing file: %v", err)
	}

	config := new(EthConfig)
	envOpt := env.Options{RequiredIfNoDef: true}
	err = env.ParseWithFuncs(config, envparse.EthParseOpts(), envOpt)
	if err != nil {
		t.Fatalf("could not parse config: %v", err)
	}

	{ // Communication setup.
		comm = mockcommunication.New()
	}

	{ // Eth server setup.
		// connect client to default http connection node.
		client, err := ethclient.Dial(config.Config.NodeAddress)
		if err != nil {
			t.Fatal(err)
		}

		instance, err := bridge.NewBridge(config.Config.BridgeContractAddress, client)
		if err != nil {
			t.Fatal(err)
		}

		signerAddress := common.Address{}

		sign := func(data []byte) ([]byte, error) {
			singIn := chains.SignRequest{
				// TODO: fix it.
				NetworkId: networks.TypeEVM,
				Data:      data,
			}

			return comm.Bridge().Sign(ctx, singIn)
		}

		transfer, err := bridge_client.NewClient(ctx, config.Bridge, signerAddress, sign)
		if err != nil {
			t.Fatal(err)
		}

		wsClient, err := ethclient.Dial(config.Config.WsNodeAddress)
		if err != nil {
			t.Fatal(err)
		}

		service = evm_service.New(
			ctx,
			config.Config,
			log,
			comm.Bridge(),
			instance,
			transfer,
			client,
			wsClient,
		)
	}

	{ // Server setup.
		controller := controllers.NewConnector(ctx, log, service)

		registerServer := func(grpcServer *grpc.Server) {
			bridge_connectorpb.RegisterConnectorServer(grpcServer, controller)
		}

		server = chains_server.NewServer(ctx, log, registerServer, config.GrpcServerAddress, config.ServerName)
	}

	connector := peer.New(log, comm, service, server, config.ServerName)

	var group errgroup.Group
	group.Go(func() error {
		return connector.Run(ctx)
	})
	group.Go(func() error {
		test(ctx, t)
		cancel()
		return nil
	})

	err = group.Wait()
	if err != nil {
		log.Error("could not run test/server", err)
		return
	}

	defer func() {
		cancel()
		err := connector.Close()
		if err != nil {
			log.Error("could not close connector", err)
		}
	}()
}
