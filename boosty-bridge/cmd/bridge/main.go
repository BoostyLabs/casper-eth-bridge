// Copyright (C) 2022 Creditor Corp. Group.
// See LICENSE for copying information.

package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"

	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"github.com/zeebo/errs"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"

	connectorbridgepb "github.com/BoostyLabs/casper-eth-bridge/boosty-communication/go-gen/connector-bridge"
	gatewaybridgepb "github.com/BoostyLabs/casper-eth-bridge/boosty-communication/go-gen/gateway-bridge"

	"tricorn/bridge"
	"tricorn/bridge/database"
	"tricorn/bridge/networks"
	"tricorn/bridge/server/controllers"
	"tricorn/communication"
	"tricorn/communication/mockcommunication"
	"tricorn/communication/rpc"
	"tricorn/internal/logger"
	"tricorn/internal/logger/zaplog"
	"tricorn/internal/server"
	grpc_server "tricorn/internal/server/grpc"
)

// Error is a default error type for bridge cli.
var Error = errs.Class("bridge cli")

type Config struct {
	DialConfig               rpc.Config
	SignerServerAddress      string             `env:"SIGNER_SERVER_ADDRESS"`
	EthServerAddress         string             `env:"ETH_SERVER_ADDRESS"`
	CasperServerAddress      string             `env:"CASPER_SERVER_ADDRESS"`
	Database                 string             `env:"DATABASE"`
	GatewayGrpcServerAddress string             `env:"GATEWAY_GRPC_SERVER_ADDRESS"`
	BridgeGrpcServerAddress  string             `env:"BRIDGE_GRPC_SERVER_ADDRESS"`
	CommunicationMode        communication.Mode `env:"COMMUNICATION_MODE"`
}

// commands.
var (
	rootCmd = &cobra.Command{
		Use:   "bridge",
		Short: "cli for interacting with bridge service",
	}
	runCmd = &cobra.Command{
		Use:         "run",
		Short:       "runs the program",
		RunE:        cmdRun,
		Annotations: map[string]string{"type": "run"},
	}
	seedCmd = &cobra.Command{
		Use:         "seed",
		Short:       "seeds test data to db",
		RunE:        cmdSeed,
		Annotations: map[string]string{"type": "run"},
	}
)

func init() {
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(seedCmd)
}

func main() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func cmdRun(cmd *cobra.Command, args []string) (err error) {
	var (
		connectorBridgeServer server.Server
		gatewayBridgeServer   server.Server
	)

	ctx, cancel := context.WithCancel(context.Background())
	onSigInt(func() {
		// starting graceful exit on context cancellation.
		cancel()
	})

	log := zaplog.NewLog()

	err = godotenv.Overload("./configs/.bridge.env")
	if err != nil {
		log.Error("could not load bridge config: %v", Error.Wrap(err))
		return Error.Wrap(err)
	}

	config := new(Config)
	err = env.Parse(config)
	if err != nil {
		log.Error("could not parse config: %v", Error.Wrap(err))
		return Error.Wrap(err)
	}

	db, err := database.New(config.Database)
	if err != nil {
		log.Error("Error starting master database on signer bank service", Error.Wrap(err))
		return Error.Wrap(err)
	}
	defer func() {
		err = errs.Combine(err, db.Close())
	}()

	// TODO: replace with migrations.
	err = db.CreateSchema(ctx)
	if err != nil {
		log.Error("Error creation bridge schema", Error.Wrap(err))
		return Error.Wrap(err)
	}

	var signer bridge.Signer
	{ // communication setup.
		switch config.CommunicationMode {
		case communication.ModeGRPC:
			config.DialConfig.ServerAddress = config.SignerServerAddress
			comm, err := rpc.New(config.DialConfig, log, true)
			if err != nil {
				return Error.Wrap(err)
			}

			signer = comm.Signer()
		default:
			comm := mockcommunication.New()
			signer = comm.Signer()
		}
	}

	service := bridge.New(
		log,
		signer,
		db.Nonces(),
		db.NetworkTokens(),
		db.Tokens(),
		db.Transactions(),
		db.TokenTransfers(),
		db.NetworkBlocks(),
	)

	// connects to connectors.
	go connectorsConnect(ctx, log, service, *config)

	{ // connector-bridge server initialization.
		controller := controllers.NewSigner(service)

		registerServer := func(grpcServer *grpc.Server) {
			connectorbridgepb.RegisterConnectorBridgeServer(grpcServer, controller)
		}

		const serverName = "connector-bridge server"
		connectorBridgeServer = grpc_server.NewServer(log, registerServer, serverName, config.BridgeGrpcServerAddress)
	}

	{ // gateway-bridge server initialization.
		controller := controllers.NewGateway(log, service)

		registerServer := func(grpcServer *grpc.Server) {
			gatewaybridgepb.RegisterGatewayBridgeServer(grpcServer, controller)
		}

		const serverName = "gateway-bridge server"
		gatewayBridgeServer = grpc_server.NewServer(log, registerServer, serverName, config.GatewayGrpcServerAddress)
	}

	group, ctx := errgroup.WithContext(ctx)

	group.Go(func() error {
		return connectorBridgeServer.Run(ctx)
	})
	group.Go(func() error {
		return gatewayBridgeServer.Run(ctx)
	})

	return ignoreContextCancellationError(
		errs.Combine(
			group.Wait(),
			connectorBridgeServer.Close(),
			gatewayBridgeServer.Close(),
		),
	)
}

func cmdSeed(cmd *cobra.Command, args []string) (err error) {
	ctx, cancel := context.WithCancel(context.Background())
	onSigInt(func() {
		// starting graceful exit on context cancellation.
		cancel()
	})

	log := zaplog.NewLog()

	err = godotenv.Overload("./configs/.bridge.env")
	if err != nil {
		log.Error("could not load bridge config: %v", Error.Wrap(err))
		return Error.Wrap(err)
	}

	config := new(Config)
	err = env.Parse(config)
	if err != nil {
		log.Error("could not parse config: %v", Error.Wrap(err))
		return Error.Wrap(err)
	}

	db, err := database.New(config.Database)
	if err != nil {
		log.Error("Error starting master database on signer bank service", Error.Wrap(err))
		return Error.Wrap(err)
	}
	defer func() {
		err = errs.Combine(err, db.Close())
	}()

	err = db.Tokens().Create(ctx, bridge.Token{
		ID:        1,
		ShortName: "TST",
		LongName:  "TEST",
	})
	if err != nil {
		log.Error("could not create token in the database", Error.Wrap(err))
		return Error.Wrap(err)
	}

	casperContractAddress, err := networks.StringToBytes(networks.IDCasperTest, "hash-3c0c1847d1c410338ab9b4ee0919c181cf26085997ff9c797e8a1ae5b02ddf23")
	if err != nil {
		log.Error("could not decode casper contract address", Error.Wrap(err))
		return Error.Wrap(err)
	}

	ethContractAddress, err := networks.StringToBytes(networks.IDGoerli, "0E26df2BaaFBC976a104EE3cccf1B467ff1b7a68")
	if err != nil {
		log.Error("could not decode ethereum contract address", Error.Wrap(err))
		return Error.Wrap(err)
	}

	networkTokes := []networks.NetworkToken{
		{
			NetworkID:       networks.IDCasperTest,
			TokenID:         1,
			ContractAddress: casperContractAddress,
			Decimals:        18,
		},
		{
			NetworkID:       networks.IDGoerli,
			TokenID:         1,
			ContractAddress: ethContractAddress,
			Decimals:        18,
		},
	}

	for _, networkToken := range networkTokes {
		_ = db.NetworkTokens().Create(ctx, networkToken)
		if err != nil {
			log.Error("could not create network token records for networks", Error.Wrap(err))
			return Error.Wrap(err)
		}
	}

	return nil
}

// onSigInt fires in SIGINT or SIGTERM event (usually CTRL+C).
func onSigInt(onSigInt func()) {
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-done
		onSigInt()
	}()
}

// ignoreContextCancellationError ignores cancellation and stopping errors since they are expected.
func ignoreContextCancellationError(err error) error {
	if errors.Is(err, context.Canceled) {
		return nil
	}

	return err
}

func connectorsConnect(ctx context.Context, log logger.Logger, service *bridge.Service, config Config) {
	log.Debug("start reconnecting to connectors")
	for {
		select {
		case <-ctx.Done():
			log.Debug("reconnecting to connectors stopped")
			return
		default:
		}

		if !service.IsConnectorConnected(networks.NameGoerli) {
			var ethConnector bridge.Connector
			{ // communication setup.
				switch config.CommunicationMode {
				case communication.ModeGRPC:
					config.DialConfig.ServerAddress = config.EthServerAddress
					comm, err := rpc.New(config.DialConfig, log, false)
					if err != nil {
						continue
					}

					ethConnector = comm.Connector(ctx)
				default:
					comm := mockcommunication.New()
					ethConnector = comm.Connector(ctx)
				}
			}

			service.AddConnector(ctx, networks.NameGoerli, ethConnector)
		}

		if !service.IsConnectorConnected(networks.NameCasperTest) {
			var casperConnector bridge.Connector
			{ // communication setup.
				switch config.CommunicationMode {
				case communication.ModeGRPC:
					config.DialConfig.ServerAddress = config.CasperServerAddress
					comm, err := rpc.New(config.DialConfig, log, false)
					if err != nil {
						continue
					}

					casperConnector = comm.Connector(ctx)
				default:
					comm := mockcommunication.New()
					casperConnector = comm.Connector(ctx)
				}
			}
			service.AddConnector(ctx, networks.NameCasperTest, casperConnector)
		}
	}
}
