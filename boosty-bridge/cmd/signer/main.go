// Copyright (C) 2022 Creditor Corp. Group.
// See LICENSE for copying information.

package main

import (
	"context"
	"os"

	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
	_ "github.com/joho/godotenv/autoload"
	"github.com/spf13/cobra"
	"github.com/zeebo/errs"
	"google.golang.org/grpc"

	bridge_signerpb "github.com/BoostyLabs/golden-gate-communication/go-gen/bridge-signer"

	"tricorn"
	"tricorn/internal/config/envparse"
	"tricorn/internal/logger/zaplog"
	"tricorn/internal/process"
	grpc_server "tricorn/internal/server/grpc"
	"tricorn/signer"
	"tricorn/signer/database"
	"tricorn/signer/server/controllers"
)

// Error is a default error type for signer cli.
var Error = errs.Class("signer cli")

// Config contains configurable values for signer project.
type Config struct {
	Database          string `env:"DATABASE"`
	GrpcServerAddress string `env:"GRPC_SERVER_ADDRESS"`
	Signer            signer.Config
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
	process.OnSigInt(func() {
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
	err = env.ParseWithFuncs(config, envparse.EvmParseOpts(), envOpt)
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

	service := signer.NewService(config.Signer, db.KeyStore())
	controller := controllers.NewSigner(log, service)

	registerServer := func(grpcServer *grpc.Server) {
		bridge_signerpb.RegisterBridgeSignerServer(grpcServer, controller)
	}

	server := grpc_server.NewServer(log, registerServer, config.ServerName, config.GrpcServerAddress)
	peer := tricorn.New(log, nil, nil, server, config.ServerName)

	return peer.Run(ctx)
}
