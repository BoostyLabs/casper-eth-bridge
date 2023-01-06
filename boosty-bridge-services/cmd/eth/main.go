// Copyright (C) 2022 Creditor Corp. Group.
// See LICENSE for copying information.

package main

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"github.com/caarlos0/env/v6"
	"math/big"
	"os"
	"os/signal"
	"syscall"

	bridge_connectorpb "github.com/BoostyLabs/casper-eth-bridge/boosty-communication/go-gen/bridge-connector"
	"github.com/btcsuite/btcd/btcec"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/joho/godotenv"
	_ "github.com/joho/godotenv/autoload"
	"github.com/spf13/cobra"
	"github.com/zeebo/errs"
	"google.golang.org/grpc"

	peer "github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services"
	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/chains"
	evm_service "github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/chains/evm"
	eth_server "github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/chains/server"
	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/chains/server/controllers"
	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/communication"
	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/communication/mockcommunication"
	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/communication/rpc"
	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/internal/contracts/evm/bridge"
	bridge_client "github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/internal/contracts/evm/client"
	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/networks"
	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/pkg/envparse"
	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/pkg/logger/zaplog"
)

// Error is a default error type for eth connector cli.
var Error = errs.Class("eth connector cli")

// Config is the global configuration to run connector server.
type Config struct {
	GrpcServerAddress string `env:"GRPC_SERVER_ADDRESS"`
	Service           evm_service.Config
	Communication     rpc.Config
	CommunicationMode communication.Mode `env:"COMMUNICATION_MODE"`
	ServerName        string             `env:"SERVER_NAME"`
	Bridge            bridge_client.Config
}

// commands.
var (
	rootCmd = &cobra.Command{
		Use:   "connector",
		Short: "cli for interacting with eth connector project",
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
		comm          communication.Communication
		service       *evm_service.Service
		server        *eth_server.Server
		signerAddress common.Address
	)

	ctx, cancel := context.WithCancel(context.Background())
	onSigInt(func() {
		// starting graceful exit on context cancellation.
		cancel()
	})

	log := zaplog.NewLog()

	err = godotenv.Overload("./configs/.eth.env")
	if err != nil {
		log.Error("could not load eth config: %v", Error.Wrap(err))
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

			// TODO: fix it.
			publicKey, err := comm.Bridge().PublicKey(ctx, networks.TypeEVM)
			if err != nil {
				return Error.Wrap(err)
			}

			if len(publicKey) < evm_service.CurveCoordinatesSize {
				return Error.New("invalid public key curve coordinates")
			}

			x := big.NewInt(0).SetBytes(publicKey[:32])
			y := big.NewInt(0).SetBytes(publicKey[32:])
			publicKeyECDSA := ecdsa.PublicKey{
				Curve: btcec.S256(),
				X:     x,
				Y:     y,
			}

			signerAddress = crypto.PubkeyToAddress(publicKeyECDSA)
		default:
			comm = mockcommunication.New()
			signerAddress = common.Address{}
		}
	}

	{ // Eth server setup.
		// connect client to default http connection node.
		client, err := ethclient.Dial(config.Service.NodeAddress)
		if err != nil {
			return Error.Wrap(err)
		}

		instance, err := bridge.NewBridge(config.Service.BridgeContractAddress, client)
		if err != nil {
			return Error.Wrap(err)
		}

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
			return Error.Wrap(err)
		}

		wsClient, err := ethclient.Dial(config.Service.WsNodeAddress)
		if err != nil {
			return Error.Wrap(err)
		}

		service = evm_service.New(
			ctx,
			config.Service,
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

		server = eth_server.NewServer(ctx, log, registerServer, config.GrpcServerAddress, config.ServerName)
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
