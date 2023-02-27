// Copyright (C) 2022 Creditor Corp. Group.
// See LICENSE for copying information.

package apitesting

import (
	"context"
	"testing"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/joho/godotenv"
	_ "github.com/joho/godotenv/autoload"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"

	bridge_connectorpb "github.com/BoostyLabs/golden-gate-communication/go-gen/bridge-connector"

	peer "tricorn"
	"tricorn/bridge/networks"
	"tricorn/chains"
	"tricorn/chains/casper"
	"tricorn/chains/communication/controllers"
	"tricorn/chains/evm"
	"tricorn/communication"
	"tricorn/communication/mockcommunication"
	"tricorn/communication/rpc"
	"tricorn/internal/config/envparse"
	signer_lib "tricorn/internal/contracts/casper"
	"tricorn/internal/contracts/evm/bridge"
	"tricorn/internal/contracts/evm/client"
	"tricorn/internal/logger/zaplog"
	"tricorn/internal/server"
	grpc_server "tricorn/internal/server/grpc"
	"tricorn/pkg/casper-sdk/mock"
	"tricorn/signer"
)

// CasperConfig is the global configuration to run casper connector server.
type CasperConfig struct {
	GrpcServerAddress string `env:"GRPC_SERVER_ADDRESS"`
	Config            casper.Config
	Communication     rpc.Config
	CommunicationMode communication.Mode `env:"COMMUNICATION_MODE"`
	ServerName        string             `env:"SERVER_NAME"`
}

func CasperRun(t *testing.T, test func(ctx context.Context, t *testing.T)) {
	var (
		comm    communication.Communication
		client  casper.Casper
		service *casper.Service
		server  server.Server
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
	err = env.ParseWithFuncs(config, envparse.EvmParseOpts(), envOpt)
	if err != nil {
		t.Fatalf("could not parse config: %v", err)
	}

	// Communication setup.
	comm = mockcommunication.New()
	client = mock.New()
	sign := func(data []byte, _ signer.Type) ([]byte, error) {
		return []byte{}, nil
	}

	signerClient := signer_lib.NewSigner(sign)

	// Casper server setup.
	{
		service = casper.NewService(ctx, config.Config, log, comm.Bridge(), client, signerClient)
	}

	{ // Server setup.
		controller := controllers.NewConnector(ctx, log, service)

		registerServer := func(grpcServer *grpc.Server) {
			bridge_connectorpb.RegisterConnectorServer(grpcServer, controller)
		}

		server = grpc_server.NewServer(log, registerServer, config.ServerName, config.GrpcServerAddress)
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
	Config            evm.Config
	Communication     rpc.Config
	CommunicationMode communication.Mode `env:"COMMUNICATION_MODE"`
	ServerName        string             `env:"SERVER_NAME"`
	Bridge            client.Config
}

func EthRun(t *testing.T, test func(ctx context.Context, t *testing.T)) {
	var (
		comm    communication.Communication
		service *evm.Service
		server  server.Server
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
	err = env.ParseWithFuncs(config, envparse.EvmParseOpts(), envOpt)
	if err != nil {
		t.Fatalf("could not parse config: %v", err)
	}

	{ // Communication setup.
		comm = mockcommunication.New()
	}

	{ // Eth server setup.
		// connect client to default http connection node.
		ethClient, err := ethclient.Dial(config.Config.NodeAddress)
		if err != nil {
			t.Fatal(err)
		}

		instance, err := bridge.NewBridge(config.Config.BridgeContractAddress, ethClient)
		if err != nil {
			t.Fatal(err)
		}

		signerAddress := common.Address{}

		sign := func(data []byte, dataType signer.Type) ([]byte, error) {
			singIn := chains.SignRequest{
				// TODO: fix it.
				NetworkId: networks.TypeEVM,
				Data:      data,
				DataType:  dataType,
			}

			return comm.Bridge().Sign(ctx, singIn)
		}

		transfer, err := client.NewClient(ctx, config.Bridge, signerAddress, sign)
		if err != nil {
			t.Fatal(err)
		}

		service = evm.New(
			ctx,
			config.Config,
			log,
			comm.Bridge(),
			instance,
			transfer,
			ethClient,
		)
	}

	{ // Server setup.
		controller := controllers.NewConnector(ctx, log, service)

		registerServer := func(grpcServer *grpc.Server) {
			bridge_connectorpb.RegisterConnectorServer(grpcServer, controller)
		}

		server = grpc_server.NewServer(log, registerServer, config.ServerName, config.GrpcServerAddress)
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
