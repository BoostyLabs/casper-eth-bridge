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

	bridge_oraclepb "github.com/BoostyLabs/casper-eth-bridge/boosty-communication/go-gen/bridge-oracle"

	"tricorn"
	"tricorn/currencyrates"
	"tricorn/currencyrates/chainlink"
	"tricorn/currencyrates/server/controllers"
	"tricorn/internal/logger/zaplog"
	"tricorn/internal/process"
	grpc_server "tricorn/internal/server/grpc"
)

// Error is a default error type for currencyrates cli.
var Error = errs.Class("currencyrates cli")

// Config contains configurable values for currencyrates project.
type Config struct {
	GrpcServerAddress string `env:"GRPC_SERVER_ADDRESS"`
	CurrencyRates     currencyrates.Config
	ServerName        string `env:"SERVER_NAME"`
}

// commands.
var (
	rootCmd = &cobra.Command{
		Use:   "currencyrates",
		Short: "cli for interacting with currencyrates project",
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

	err = godotenv.Overload("./configs/.currencyrates.env")
	if err != nil {
		log.Error("could not load config: %v", Error.Wrap(err))
		return Error.Wrap(err)
	}

	config := new(Config)
	envOpt := env.Options{RequiredIfNoDef: true}
	err = env.Parse(config, envOpt)
	if err != nil {
		log.Error("could not parse config: %v", Error.Wrap(err))
		return Error.Wrap(err)
	}

	chainlinkClient := chainlink.New(config.CurrencyRates.CurrencyRateBaseURL)

	service := currencyrates.NewService(ctx, config.CurrencyRates, log, chainlinkClient)
	controller := controllers.NewCurrencyRates(ctx, log, service)

	registerServer := func(grpcServer *grpc.Server) {
		bridge_oraclepb.RegisterBridgeOracleServer(grpcServer, controller)
	}

	server := grpc_server.NewServer(log, registerServer, config.ServerName, config.GrpcServerAddress)
	peer := tricorn.New(log, nil, nil, server, config.ServerName)

	return peer.Run(ctx)
}
