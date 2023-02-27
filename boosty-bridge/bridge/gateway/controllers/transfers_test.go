package controllers_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"testing"

	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	transferspb "github.com/BoostyLabs/casper-eth-bridge/boosty-communication/go-gen/transfers"

	"tricorn/bridge/gateway"
	"tricorn/bridge/gateway/controllers/apitesting"
	"tricorn/bridge/networks"
	"tricorn/bridge/transfers"
	"tricorn/communication"
	"tricorn/communication/mockcommunication"
	"tricorn/communication/rpc"
	"tricorn/internal/config/envparse"
	"tricorn/internal/logger/zaplog"
)

// TODO: test with err sent from mock service.

func TestTransfers(t *testing.T) {
	log := zaplog.NewLog()

	err := godotenv.Overload("./apitesting/configs/.test.gateway.env")
	if err != nil {
		t.Fatalf("could not load config: %v", err)
	}

	config := new(apitesting.Config)
	envOpt := env.Options{RequiredIfNoDef: true}
	err = env.ParseWithFuncs(config, envparse.EvmParseOpts(), envOpt)
	if err != nil {
		t.Fatalf("could not parse ENV config: %v", err)
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

	{ // transfers setup.
		g.transfers = transfers.NewService(
			g.communication.Transfers(),
		)
	}

	apitesting.Run(t, func(ctx context.Context, t *testing.T) {
		baseURL := fmt.Sprintf("http://%s/api/v0/transfers", config.Server.Address)

		t.Run("transfer info", func(t *testing.T) {
			url := baseURL + "/tx"
			resp, err := apitesting.HTTPDo(ctx, url, http.MethodGet, nil)
			assert.NoError(t, err)
			assert.Equal(t, resp.StatusCode, http.StatusOK)
			defer func() {
				err = resp.Body.Close()
				require.NoError(t, err)
			}()

			var result []transfers.Transfer
			err = json.NewDecoder(resp.Body).Decode(&result)
			require.NoError(t, err)

			expected, err := g.transfers.Info(ctx, "")
			require.NoError(t, err)

			for i := 0; i < len(expected); i++ {
				assert.Equal(t, expected[i].Amount, result[i].Amount)
			}
		})

		t.Run("estimate transfer wrong sender-network", func(t *testing.T) {
			url := baseURL + "/estimate/q/e/r/t"
			resp, err := apitesting.HTTPDo(ctx, url, http.MethodGet, nil)
			assert.NoError(t, err)
			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
			defer func() {
				err = resp.Body.Close()
				require.NoError(t, err)
			}()
		})

		t.Run("estimate transfer wrong recipient-network", func(t *testing.T) {
			url := baseURL + "/estimate/EVM/e/r/t"
			resp, err := apitesting.HTTPDo(ctx, url, http.MethodGet, nil)
			assert.NoError(t, err)
			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
			defer func() {
				err = resp.Body.Close()
				require.NoError(t, err)
			}()
		})

		t.Run("estimate transfer wrong token-id", func(t *testing.T) {
			url := baseURL + "/estimate/EVM/CASPER/r/t"
			resp, err := apitesting.HTTPDo(ctx, url, http.MethodGet, nil)
			assert.NoError(t, err)
			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
			defer func() {
				err = resp.Body.Close()
				require.NoError(t, err)
			}()
		})

		t.Run("estimate transfer", func(t *testing.T) {
			url := baseURL + "/estimate/ETH/CASPER/1/1"
			resp, err := apitesting.HTTPDo(ctx, url, http.MethodGet, nil)
			assert.NoError(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
			defer func() {
				err = resp.Body.Close()
				require.NoError(t, err)
			}()

			var result transfers.Estimate
			err = json.NewDecoder(resp.Body).Decode(&result)
			require.NoError(t, err)

			expected, err := g.transfers.Estimate(ctx, networks.NameEth, networks.NameCasper, 1, "1")
			require.NoError(t, err)

			assert.Equal(t, expected.Fee, result.Fee)
			assert.Equal(t, expected.FeePercentage, result.FeePercentage)
			assert.Equal(t, expected.EstimatedConfirmationTime, result.EstimatedConfirmationTime)
		})

		t.Run("transfer history offset missing", func(t *testing.T) {
			url := baseURL + "/history/wrong-sign/wrong-pubKey"
			resp, err := apitesting.HTTPDo(ctx, url, http.MethodGet, nil)
			assert.NoError(t, err)
			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
			defer func() {
				err = resp.Body.Close()
				require.NoError(t, err)
			}()
		})

		t.Run("transfer history wrong offset", func(t *testing.T) {
			url := baseURL + "/history/wrong-sign/wrong-pubKey?offset=w"
			resp, err := apitesting.HTTPDo(ctx, url, http.MethodGet, nil)
			assert.NoError(t, err)
			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
			defer func() {
				err = resp.Body.Close()
				require.NoError(t, err)
			}()
		})

		t.Run("transfer history limit missing", func(t *testing.T) {
			url := baseURL + "/history/wrong-sign/wrong-pubKey?offset=1"
			resp, err := apitesting.HTTPDo(ctx, url, http.MethodGet, nil)
			assert.NoError(t, err)
			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
			defer func() {
				err = resp.Body.Close()
				require.NoError(t, err)
			}()
		})

		t.Run("transfer history wrong limit", func(t *testing.T) {
			url := baseURL + "/history/wrong-sign/wrong-pubKey?offset=1&limit=w"
			resp, err := apitesting.HTTPDo(ctx, url, http.MethodGet, nil)
			assert.NoError(t, err)
			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
			defer func() {
				err = resp.Body.Close()
				require.NoError(t, err)
			}()
		})

		t.Run("transfer history network-id missing", func(t *testing.T) {
			url := baseURL + "/history/wrong-sign/wrong-pubKey?offset=1&limit=1"
			resp, err := apitesting.HTTPDo(ctx, url, http.MethodGet, nil)
			assert.NoError(t, err)
			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
			defer func() {
				err = resp.Body.Close()
				require.NoError(t, err)
			}()
		})

		t.Run("transfer history wrong network-id", func(t *testing.T) {
			url := baseURL + "/history/wrong-sign/wrong-pubKey?offset=1&limit=1&network-id=w"
			resp, err := apitesting.HTTPDo(ctx, url, http.MethodGet, nil)
			assert.NoError(t, err)
			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
			defer func() {
				err = resp.Body.Close()
				require.NoError(t, err)
			}()
		})

		t.Run("transfer history", func(t *testing.T) {
			url := baseURL + "/history/34850b7e36e635783df0563c7202c3ac776df59db5015d2b6f0add33955bb5c43ce35efb5ce695a243bc4c5dc4298db40cd765f3ea5612d2d57da1e4933b2f201b/34850b7e36e635783df0563c7202c3ac776df59db5015d2b6f0add33955bb5c43ce35efb5ce695a243bc4c5dc4298db40cd765f3ea5612d2d57da1e4933b2f201b?offset=1&limit=1&network-id=1"
			resp, err := apitesting.HTTPDo(ctx, url, http.MethodGet, nil)
			assert.NoError(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
			defer func() {
				err = resp.Body.Close()
				require.NoError(t, err)
			}()

			var result transfers.Page
			err = json.NewDecoder(resp.Body).Decode(&result)
			require.NoError(t, err)

			expected, err := g.transfers.History(ctx, 1, 1, []byte("signature"), []byte("pubKey"), 1)
			require.NoError(t, err)

			assert.Equal(t, expected.TotalCount, result.TotalCount)
			assert.Equal(t, expected.Limit, result.Limit)
			assert.Equal(t, expected.Offset, result.Offset)
			assert.Equal(t, len(expected.Transfers), len(result.Transfers))

			for i := 0; i < len(expected.Transfers); i++ {
				assert.Equal(t, expected.Transfers[i].Amount, result.Transfers[i].Amount)
				assert.Equal(t, expected.Transfers[i].Recipient, result.Transfers[i].Recipient)
				assert.Equal(t, expected.Transfers[i].Status, result.Transfers[i].Status)
				assert.Equal(t, expected.Transfers[i].OutboundTx, result.Transfers[i].OutboundTx)
				assert.Equal(t, expected.Transfers[i].TriggeringTx, result.Transfers[i].TriggeringTx)
				assert.Equal(t, expected.Transfers[i].Sender, result.Transfers[i].Sender)
				assert.Equal(t, expected.Transfers[i].ID, result.Transfers[i].ID)
			}
		})

		t.Run("get bridge in signature", func(t *testing.T) {
			url := baseURL + "/bridge-in-signature"
			request := transferspb.BridgeInSignatureRequest{
				Sender: &transferspb.StringNetworkAddress{
					NetworkName: "GOERLI",
					Address:     "0x3095F955Da700b96215CFfC9Bc64AB2e69eB7DAB",
				},
				TokenId: 1,
				Amount:  "1",
				Destination: &transferspb.StringNetworkAddress{
					NetworkName: "CASPER-TEST",
					Address:     "9060c0820b5156b1620c8e3344d17f9fad5108f5dc2672f2308439e84363c88e",
				},
			}

			body, err := json.Marshal(&request)
			assert.NoError(t, err)

			resp, err := apitesting.HTTPDo(ctx, url, http.MethodPost, strings.NewReader(string(body)))
			require.NoError(t, err)
			defer func() {
				err = resp.Body.Close()
				require.NoError(t, err)
			}()
		})

		t.Run("get cancel signature", func(t *testing.T) {
			url := baseURL + "/cancel-signature"
			segments := "/1/1/1/1"
			url += segments

			resp, err := apitesting.HTTPDo(ctx, url, http.MethodGet, nil)
			require.NoError(t, err)
			defer func() {
				err = resp.Body.Close()
				require.NoError(t, err)
			}()
		})
	})
}
