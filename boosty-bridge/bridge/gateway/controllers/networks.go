package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/zeebo/errs"

	"tricorn/bridge/networks"
	"tricorn/internal/logger"
)

// ErrNetworks is an internal error type for network controller.
var ErrNetworks = errs.Class("networks controller")

// Networks is an api controller that exposes all networks related endpoints.
type Networks struct {
	log logger.Logger

	networks *networks.Service
}

// NewNetworks is a constructor for networks api controller.
func NewNetworks(log logger.Logger, networks *networks.Service) *Networks {
	return &Networks{
		log:      log,
		networks: networks,
	}
}

// Connected returns list of supported networks.
func (controller *Networks) Connected(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	w.Header().Set("Content-Type", "application/json")

	connectedNetworks, err := controller.networks.Connected(ctx)
	if err != nil {
		controller.log.Error("could not return connected networks", ErrNetworks.Wrap(err))
		controller.serveError(w, http.StatusInternalServerError, ErrNetworks.Wrap(err))
		return
	}

	if err = json.NewEncoder(w).Encode(connectedNetworks); err != nil {
		controller.log.Error("failed to write json error response", ErrNetworks.Wrap(err))
	}
}

// SupportedTokens returns list of supported by the network tokens.
func (controller *Networks) SupportedTokens(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	w.Header().Set("Content-Type", "application/json")

	params := mux.Vars(r)
	networkID, err := strconv.Atoi(params["network-id"])
	if err != nil {
		controller.serveError(w, http.StatusBadRequest, ErrNetworks.New("invalid network id"))
		return
	}

	tokens, err := controller.networks.SupportedTokens(ctx, uint32(networkID))
	if err != nil {
		controller.log.Error("could not return supported tokens", ErrNetworks.Wrap(err))
		controller.serveError(w, http.StatusInternalServerError, ErrNetworks.Wrap(err))
		return
	}

	if err = json.NewEncoder(w).Encode(tokens); err != nil {
		controller.log.Error("failed to write json error response", ErrNetworks.Wrap(err))
	}
}

// serveError replies to the request with specific code and error message.
func (controller *Networks) serveError(w http.ResponseWriter, status int, err error) {
	w.WriteHeader(status)

	response := ErrorResponse{
		Error: err.Error(),
	}

	response.Error = err.Error()
	if err = json.NewEncoder(w).Encode(response); err != nil {
		controller.log.Error("failed to write json error response", err)
	}
}
