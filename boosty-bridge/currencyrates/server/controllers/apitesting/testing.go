// Copyright (C) 2022 Creditor Corp. Group.
// See LICENSE for copying information.

package apitesting

import (
	"context"
	"testing"
	"time"
	"tricorn"

	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
	_ "github.com/joho/godotenv/autoload"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"

	bridge_oraclepb "github.com/BoostyLabs/casper-eth-bridge/boosty-communication/go-gen/bridge-oracle"

	"tricorn/currencyrates"
	"tricorn/currencyrates/chainlink"
	"tricorn/currencyrates/server/controllers"
	"tricorn/internal/logger/zaplog"
	grpc_server "tricorn/internal/server/grpc"
)

// Config contains configurable values for currencyrates project.
type Config struct {
	GrpcServerAddress string `env:"GRPC_SERVER_ADDRESS"`
	CurrencyRates     currencyrates.Config
	ServerName        string `env:"SERVER_NAME"`
}

func RatesRun(t *testing.T, test func(ctx context.Context, t *testing.T)) {
	ctx, cancel := context.WithCancel(context.Background())
	log := zaplog.NewLog()

	err := godotenv.Overload("./apitesting/configs/.test.currencyrates.env")
	if err != nil {
		t.Fatalf("could not load currencyrates testing file: %v", err)
	}

	config := new(Config)
	envOpt := env.Options{RequiredIfNoDef: true}
	err = env.Parse(config, envOpt)
	if err != nil {
		t.Fatalf("could not parse config: %v", err)
	}

	chainlinkClient := chainlink.New(config.CurrencyRates.CurrencyRateBaseURL)

	service := currencyrates.NewService(ctx, config.CurrencyRates, log, chainlinkClient)
	controller := controllers.NewCurrencyRates(ctx, log, service)

	registerServer := func(grpcServer *grpc.Server) {
		bridge_oraclepb.RegisterBridgeOracleServer(grpcServer, controller)
	}

	server := grpc_server.NewServer(log, registerServer, config.ServerName, config.GrpcServerAddress)
	peer := tricorn.New(log, nil, nil, server, config.ServerName)

	var group errgroup.Group
	group.Go(func() error {
		return peer.Run(ctx)
	})

	time.Sleep(time.Second)

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
}
