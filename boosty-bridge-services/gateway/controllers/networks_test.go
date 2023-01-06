package controllers_test

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/caarlos0/env/v6"
	"net"
	"net/http"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/communication"
	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/communication/mockcommunication"
	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/communication/rpc"
	server "github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/gateway"
	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/gateway/controllers/apitesting"
	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/networks"
	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/pkg/envparse"
	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/pkg/logger/zaplog"
	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/transfers"
)

func TestNetworks(t *testing.T) {
	log := zaplog.NewLog()

	err := godotenv.Overload("./apitesting/configs/.test.gateway.env")
	if err != nil {
		t.Fatalf("could not load config: %v", err)
	}

	config := new(apitesting.Config)
	envOpt := env.Options{RequiredIfNoDef: true}
	err = env.ParseWithFuncs(config, envparse.EthParseOpts(), envOpt)
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
		server   *server.Server
	}{}

	{ // Communication setup.
		switch config.CommunicationMode {
		case communication.ModeGRPC:
			g.communication, err = rpc.New(config.Communication, log)
			require.NoError(t, err)
		default:
			g.communication = mockcommunication.New()
		}
	}

	apitesting.Run(t, func(ctx context.Context, t *testing.T) {
		baseURL := fmt.Sprintf("http://%s/api/v0/networks", config.Server.Address)

		t.Run("connected networks", func(t *testing.T) {
			resp, err := apitesting.HTTPDo(ctx, baseURL, http.MethodGet, nil)
			assert.NoError(t, err)
			assert.Equal(t, resp.StatusCode, http.StatusOK)
			defer func() {
				err = resp.Body.Close()
				require.NoError(t, err)
			}()

			var result []networks.Network
			err = json.NewDecoder(resp.Body).Decode(&result)
			require.NoError(t, err)

			expected, err := g.communication.Networks().ConnectedNetworks(ctx)
			require.NoError(t, err)

			for i := 0; i < len(expected); i++ {
				assert.Equal(t, expected[i].Type, result[i].Type)
				assert.Equal(t, expected[i].ID, result[i].ID)
				assert.Equal(t, expected[i].IsTestnet, result[i].IsTestnet)
				assert.Equal(t, expected[i].Name, result[i].Name)
			}
		})

		t.Run("supported tokens", func(t *testing.T) {
			resp, err := apitesting.HTTPDo(ctx, baseURL, http.MethodGet, nil)
			assert.NoError(t, err)
			assert.Equal(t, resp.StatusCode, http.StatusOK)
			defer func() {
				err = resp.Body.Close()
				require.NoError(t, err)
			}()

			var result []networks.Network
			err = json.NewDecoder(resp.Body).Decode(&result)
			require.NoError(t, err)

			expected, err := g.communication.Networks().ConnectedNetworks(ctx)
			require.NoError(t, err)

			for i := 0; i < len(expected); i++ {
				assert.Equal(t, expected[i].Type, result[i].Type)
				assert.Equal(t, expected[i].ID, result[i].ID)
				assert.Equal(t, expected[i].IsTestnet, result[i].IsTestnet)
				assert.Equal(t, expected[i].Name, result[i].Name)
			}
		})
	})
}
