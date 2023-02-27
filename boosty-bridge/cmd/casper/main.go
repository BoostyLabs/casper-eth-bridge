// Copyright (C) 2022 Creditor Corp. Group.
// See LICENSE for copying information.

package main

import (
	"context"
	"errors"
	"os"

	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
	_ "github.com/joho/godotenv/autoload"
	"github.com/spf13/cobra"
	"github.com/zeebo/errs"
	"google.golang.org/grpc"

	bridge_connectorpb "github.com/BoostyLabs/golden-gate-communication/go-gen/bridge-connector"

	"tricorn"
	"tricorn/bridge/networks"
	"tricorn/chains"
	"tricorn/chains/casper"
	"tricorn/chains/communication/controllers"
	"tricorn/communication"
	"tricorn/communication/mockcommunication"
	"tricorn/communication/rpc"
	"tricorn/internal/config/envparse"
	signer_lib "tricorn/internal/contracts/casper"
	"tricorn/internal/logger/zaplog"
	"tricorn/internal/process"
	"tricorn/internal/server"
	grpc_server "tricorn/internal/server/grpc"
	"tricorn/pkg/casper-sdk/client"
	"tricorn/pkg/casper-sdk/mock"
	"tricorn/signer"
)

// Error is a default error type for casper connector cli.
var Error = errs.Class("casper connector cli")

// Config is the global configuration to run connector server.
type Config struct {
	GrpcServerAddress string `env:"GRPC_SERVER_ADDRESS"`
	Config            casper.Config
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
		comm         communication.Communication
		casperClient casper.Casper
		service      *casper.Service
		server       server.Server
	)

	ctx, cancel := context.WithCancel(context.Background())
	process.OnSigInt(func() {
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
	err = env.ParseWithFuncs(config, envparse.EvmParseOpts(), envOpt)
	if err != nil {
		log.Error("could not parse config: %v", Error.Wrap(err))
		return Error.Wrap(err)
	}

	{ // Communication setup.
		switch config.CommunicationMode {
		case communication.ModeGRPC:
			comm, err = rpc.New(config.Communication, log, true)
			if err != nil {
				return Error.Wrap(err)
			}

			casperClient = client.New(config.Config.RPCNodeAddress)
		default:
			comm = mockcommunication.New()
			casperClient = mock.New()
		}
	}

	sign := func(data []byte, dataType signer.Type) ([]byte, error) {
		singIn := chains.SignRequest{
			NetworkId: networks.TypeCasper,
			Data:      data,
			DataType:  dataType,
		}

		return comm.Bridge().Sign(ctx, singIn)
	}

	signerClient := signer_lib.NewSigner(sign)

	{ // Casper server setup.
		service = casper.NewService(ctx, config.Config, log, comm.Bridge(), casperClient, signerClient)
	}

	{ // Server setup.
		controller := controllers.NewConnector(ctx, log, service)

		registerServer := func(grpcServer *grpc.Server) {
			bridge_connectorpb.RegisterConnectorServer(grpcServer, controller)
		}

		server = grpc_server.NewServer(log, registerServer, config.ServerName, config.GrpcServerAddress)
	}

	connector := tricorn.New(log, comm, service, server, config.ServerName)

	return ignoreContextCancellationError(errs.Combine(connector.Run(ctx), connector.Close()))
}

// ignoreContextCancellationError ignores cancellation and stopping errors since they are expected.
func ignoreContextCancellationError(err error) error {
	if errors.Is(err, context.Canceled) {
		return nil
	}

	return err
}
