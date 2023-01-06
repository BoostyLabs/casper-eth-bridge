// Copyright (C) 2022 Creditor Corp. Group.
// See LICENSE for copying information.

package main

import (
	"context"
	"errors"
	"github.com/caarlos0/env/v6"
	"os"
	"os/signal"
	"syscall"

	bridge_connectorpb "github.com/BoostyLabs/casper-eth-bridge/boosty-communication/go-gen/bridge-connector"
	"github.com/joho/godotenv"
	_ "github.com/joho/godotenv/autoload"
	"github.com/spf13/cobra"
	"github.com/zeebo/errs"
	"google.golang.org/grpc"

	peer "github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services"
	casper_service "github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/chains/casper"
	casper_server "github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/chains/server"
	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/chains/server/controllers"
	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/communication"
	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/communication/mockcommunication"
	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/communication/rpc"
	casper_client "github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/internal/casper-sdk/client"
	casper_mock "github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/internal/casper-sdk/mock"
	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/pkg/envparse"
	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/pkg/logger/zaplog"
)

// Error is a default error type for casper connector cli.
var Error = errs.Class("casper connector cli")

// Config is the global configuration to run connector server.
type Config struct {
	GrpcServerAddress string `env:"GRPC_SERVER_ADDRESS"`
	Config            casper_service.Config
	Communication     rpc.Config
	CommunicationMode communication.Mode `env:"COMMUNICATION_MODE"`
	ServerName        string             `env:"SERVER_NAME"`
}

// commands.
var (
	rootCmd = &cobra.Command{
		Use:   "connector",
		Short: "cli for interacting with casper connector project",
	}
	runCmd = &cobra.Command{
		Use:         "run",
		Short:       "runs the program",
		RunE:        cmdRun,
		Annotations: map[string]string{"type": "run"},
	}
)

func init() {
	rootCmd.AddCommand(runCmd)
}

func main() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func cmdRun(cmd *cobra.Command, args []string) (err error) {
	var (
		comm    communication.Communication
		client  casper_service.Casper
		service *casper_service.Service
		server  *casper_server.Server
	)

	ctx, cancel := context.WithCancel(context.Background())
	onSigInt(func() {
		// starting graceful exit on context cancellation.
		cancel()
	})

	log := zaplog.NewLog()

	err = godotenv.Overload("./configs/.casper.env")
	if err != nil {
		log.Error("could not load casper config: %v", Error.Wrap(err))
		return Error.Wrap(err)
	}

	err = godotenv.Overload("./configs/.env")
	if err != nil {
		log.Error("could not load config: %v", Error.Wrap(err))
		return Error.Wrap(err)
	}

	config := new(Config)
	envOpt := env.Options{RequiredIfNoDef: true}
	err = env.ParseWithFuncs(config, envparse.EthParseOpts(), envOpt)
	if err != nil {
		log.Error("could not parse config: %v", Error.Wrap(err))
		return Error.Wrap(err)
	}

	{ // Communication setup.
		switch config.CommunicationMode {
		case communication.ModeGRPC:
			comm, err = rpc.New(config.Communication, log)
			if err != nil {
				return Error.Wrap(err)
			}

			client = casper_client.New(config.Config.RPCNodeAddress)
		default:
			comm = mockcommunication.New()
			client = casper_mock.New()
		}
	}

	// Casper server setup.
	{
		service = casper_service.NewService(ctx, config.Config, log, comm.Bridge(), client)
	}

	{ // Server setup.
		controller := controllers.NewConnector(ctx, log, service)

		registerServer := func(grpcServer *grpc.Server) {
			bridge_connectorpb.RegisterConnectorServer(grpcServer, controller)
		}

		server = casper_server.NewServer(ctx, log, registerServer, config.GrpcServerAddress, config.ServerName)
	}

	connector := peer.New(log, comm, service, server, config.ServerName)

	return ignoreContextCancellationError(errs.Combine(connector.Run(ctx), connector.Close()))
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
