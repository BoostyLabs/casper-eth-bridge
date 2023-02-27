// Copyright (C) 2022 Creditor Corp. Group.
// See LICENSE for copying information.

package apitesting

import (
	"context"
	"testing"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
	_ "github.com/joho/godotenv/autoload"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"

	bridge_signerpb "github.com/BoostyLabs/casper-eth-bridge/boosty-communication/go-gen/bridge-signer"

	peer "tricorn"
	"tricorn/internal/config/envparse"
	"tricorn/internal/logger/zaplog"
	grpc_server "tricorn/internal/server/grpc"
	"tricorn/signer"
	"tricorn/signer/database/dbtesting"
	"tricorn/signer/server/controllers"
)

// SignerConfig contains configurable values for signer project.
type SignerConfig struct {
	Database          string `env:"DATABASE"`
	GrpcServerAddress string `env:"GRPC_SERVER_ADDRESS"`
	Signer            signer.Config
	ServerName        string `env:"SERVER_NAME"`
}

func SignerRun(t *testing.T, test func(ctx context.Context, t *testing.T, db signer.DB)) {
	ctx, cancel := context.WithCancel(context.Background())
	log := zaplog.NewLog()

	err := godotenv.Overload("./apitesting/configs/.test.signer.env")
	if err != nil {
		t.Fatalf("could not load signer testing file: %v", err)
	}

	config := new(SignerConfig)
	envOpt := env.Options{RequiredIfNoDef: true}
	err = env.ParseWithFuncs(config, envparse.EvmParseOpts(), envOpt)
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

	service := signer.NewService(config.Signer, db.KeyStore())
	controller := controllers.NewSigner(log, service)

	registerServer := func(grpcServer *grpc.Server) {
		bridge_signerpb.RegisterBridgeSignerServer(grpcServer, controller)
	}
	serverName := "signer"
	server := grpc_server.NewServer(log, registerServer, serverName, config.GrpcServerAddress)

	signer := peer.New(log, nil, nil, server, config.ServerName)

	var group errgroup.Group
	group.Go(func() error {
		return signer.Run(ctx)
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
