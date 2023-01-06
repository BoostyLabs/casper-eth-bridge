// Copyright (C) 2022 Creditor Corp. Group.
// See LICENSE for copying information.

package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	bridge_signerpb "github.com/BoostyLabs/casper-eth-bridge/boosty-communication/go-gen/bridge-signer"
	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
	_ "github.com/joho/godotenv/autoload"
	"github.com/spf13/cobra"
	"github.com/zeebo/errs"
	"google.golang.org/grpc"

	peer "github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services"
	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/database"
	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/pkg/envparse"
	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/pkg/logger/zaplog"
	signer_service "github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/signer"
	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/signer/server"
	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/signer/server/controllers"
)

// Error is a default error type for signer cli.
var Error = errs.Class("signer cli")

// Config contains configurable values for signer project.
type Config struct {
	Database          string `env:"DATABASE"`
	GrpcServerAddress string `env:"GRPC_SERVER_ADDRESS"`
	Signer            signer_service.Config
	ServerName        string `env:"SERVER_NAME"`
}

// commands.
var (
	rootCmd = &cobra.Command{
		Use:   "signer",
		Short: "cli for interacting with signer project",
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
	ctx, cancel := context.WithCancel(context.Background())
	onSigInt(func() {
		// starting graceful exit on context cancellation.
		cancel()
	})

	log := zaplog.NewLog()

	err = godotenv.Overload("./configs/.signer.env")
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

	db, err := database.New(config.Database)
	if err != nil {
		log.Error("Error starting master database on signer bank service", Error.Wrap(err))
		return Error.Wrap(err)
	}
	defer func() {
		err = errs.Combine(err, db.Close())
	}()

	// TODO: remove for production.
	err = db.CreateSchema(ctx)
	if err != nil {
		log.Error("Error creating schema", Error.Wrap(err))
	}

	service := signer_service.NewService(config.Signer, db.PrivateKeys())
	controller := controllers.NewSigner(log, service)

	registerServer := func(grpcServer *grpc.Server) {
		bridge_signerpb.RegisterBridgeSignerServer(grpcServer, controller)
	}
	server := server.NewServer(ctx, log, config.GrpcServerAddress, registerServer)

	peer := peer.New(log, nil, nil, server, config.ServerName)

	return peer.Run(ctx)
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
