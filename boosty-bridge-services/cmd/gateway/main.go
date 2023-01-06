package main

import (
	"context"
	"errors"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
	_ "github.com/joho/godotenv/autoload"
	"github.com/spf13/cobra"
	"github.com/zeebo/errs"

	gateway "github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services"
	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/communication"
	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/communication/mockcommunication"
	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/communication/rpc"
	server "github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/gateway"
	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/networks"
	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/pkg/envparse"
	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/pkg/logger/zaplog"
	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/transfers"
)

// Error is a default error type for golden-gate gateway cli.
var Error = errs.Class("gateway cli")

// Config is the global configuration for golden-gate gateway.
type Config struct {
	Server            server.Config
	Communication     rpc.Config
	CommunicationMode communication.Mode `env:"COMMUNICATION_MODE"`
	ServerName        string             `env:"SERVER_NAME"`
}

// commands.
var (
	rootCmd = &cobra.Command{
		Use:   "golden-gate-gateway",
		Short: "cli for interacting with golden-gate gateway",
	}
	runCmd = &cobra.Command{
		Use:         "run",
		Short:       "runs the gateway",
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

	err = godotenv.Overload("./configs/.gateway.env")
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

	g := struct {
		communication communication.Communication

		// declares all networks specific modules.
		networks *networks.Service

		// declares all transfers specific modules.
		transfers *transfers.Service

		// declares all gateway server specific modules.
		listener net.Listener
		server   *server.Server
	}{}

	{ // Communication setup.
		switch config.CommunicationMode {
		case communication.ModeGRPC:
			g.communication, err = rpc.New(config.Communication, log)
			if err != nil {
				return err
			}
		default:
			g.communication = mockcommunication.New()
		}
	}

	{ // networks setup.
		g.networks = networks.NewService(
			g.communication.Networks(),
		)
	}

	{ // transfers setup.
		g.transfers = transfers.NewService(
			g.communication.Transfers(),
		)
	}

	{ // server setup.
		g.listener, err = net.Listen("tcp", config.Server.Address)
		if err != nil {
			return err
		}

		g.server = server.NewServer(
			config.Server,
			log,
			g.listener,
			g.networks,
			g.transfers,
		)
	}

	gateway := gateway.New(log, g.communication, nil, g.server, config.ServerName)

	return ignoreContextCancellationError(errs.Combine(gateway.Run(ctx), gateway.Close()))
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
