package apitesting

import (
	"context"
	"io"
	"net"
	"net/http"
	"testing"

	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"

	peer "tricorn"
	"tricorn/bridge/gateway"
	"tricorn/bridge/networks"
	"tricorn/bridge/transfers"
	"tricorn/communication"
	"tricorn/communication/mockcommunication"
	"tricorn/communication/rpc"
	"tricorn/internal/config/envparse"
	"tricorn/internal/logger/zaplog"
)

// Config is the global configuration for golden-gate gateway.
type Config struct {
	Server            gateway.Config
	Communication     rpc.Config
	CommunicationMode communication.Mode `env:"COMMUNICATION_MODE"`
	ServerName        string             `env:"SERVER_NAME"`
}

// Run method will run api server, establish connection and close it after test is passed.
func Run(t *testing.T, test func(ctx context.Context, t *testing.T)) {
	ctx, cancel := context.WithCancel(context.Background())

	log := zaplog.NewLog()

	err := godotenv.Overload("./apitesting/configs/.test.gateway.env")
	if err != nil {
		t.Fatalf("could not load config: %v", err)
	}

	config := new(Config)
	envOpt := env.Options{RequiredIfNoDef: true}
	err = env.ParseWithFuncs(config, envparse.EvmParseOpts(), envOpt)
	if err != nil {
		t.Fatalf("could not parse config: %v", err)
	}

	g := struct {
		communication communication.Communication

		// declares all networks specific modules.
		networks *networks.Service

		// declares all transfers specific modules.
		transfers *transfers.Service

		// declares all gateway server specific modules.
		listener net.Listener
		server   *gateway.Server
	}{}

	{ // Communication setup.
		switch config.CommunicationMode {
		case communication.ModeGRPC:
			g.communication, err = rpc.New(config.Communication, log, true)
			require.NoError(t, err)
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
		require.NoError(t, err)

		g.server = gateway.NewServer(
			config.Server,
			log,
			g.listener,
			g.networks,
			g.transfers,
		)
	}

	gateway := peer.New(log, g.communication, nil, g.server, config.ServerName)

	var group errgroup.Group
	group.Go(func() error {
		return gateway.Run(ctx)
	})

	test(ctx, t)

	cancel()
	err = gateway.Close()
	require.NoError(t, err)
}

// HTTPDo performs http request.
func HTTPDo(ctx context.Context, url, method string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}
	return http.DefaultClient.Do(req)
}
