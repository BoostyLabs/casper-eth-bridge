package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/zeebo/errs"

	casper_contract "github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/pkg/contract"
	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/pkg/logger"
)

var (
	// ErrContract is an internal error type for contract controller.
	ErrContract = errs.Class("contract controller")
)

// contract describes handlers for contract.
type contract struct {
	log logger.Logger
}

// New constructor for contract.
func New(log logger.Logger) *contract {
	return &contract{
		log: log,
	}
}

// BridgeIn sends transaction to bridgeIn method.
func (contract *contract) BridgeIn(w http.ResponseWriter, r *http.Request) {
	var req casper_contract.BridgeInRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		contract.serveError(w, http.StatusBadRequest, ErrContract.Wrap(err))
		return
	}

	resp, err := casper_contract.BridgeIn(r.Context(), req)
	if err != nil {
		contract.log.Error("couldn't send bridge in transaction", err)
		contract.serveError(w, http.StatusInternalServerError, ErrContract.Wrap(err))
		return
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		contract.log.Error("failed to write json response", err)
		return
	}
}

// serveError replies to the request with specific code and error message.
func (contract *contract) serveError(w http.ResponseWriter, status int, err error) {
	w.WriteHeader(status)
	var response struct {
		Error string `json:"error"`
	}

	response.Error = err.Error()
	if err = json.NewEncoder(w).Encode(response); err != nil {
		contract.log.Error("failed to write json error response", ErrContract.Wrap(err))
	}
}
