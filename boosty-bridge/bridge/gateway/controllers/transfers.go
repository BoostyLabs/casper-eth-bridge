package controllers

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/zeebo/errs"

	"tricorn/bridge/networks"
	"tricorn/bridge/transfers"
	"tricorn/internal/logger"
)

// ErrTransfers is an internal error type for transfers controller.
var ErrTransfers = errs.Class("transfers controller")

// Transfers is an api controller that exposes all transactions related endpoints.
type Transfers struct {
	log logger.Logger

	transfers *transfers.Service
}

// NewTransfers is a constructor for transfers api controller.
func NewTransfers(log logger.Logger, transfers *transfers.Service) *Transfers {
	return &Transfers{
		log:       log,
		transfers: transfers,
	}
}

// Info returns list of transfers of triggering transaction.
func (controller *Transfers) Info(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	w.Header().Set("Content-Type", "application/json")

	params := mux.Vars(r)
	tx := params["tx"]
	// TODO: implement transaction validation.

	tokens, err := controller.transfers.Info(ctx, tx)
	if err != nil {
		controller.log.Error("could not return supported tokens", ErrTransfers.Wrap(err))
		controller.serveError(w, http.StatusInternalServerError, ErrTransfers.Wrap(err))
		return
	}

	if err = json.NewEncoder(w).Encode(tokens); err != nil {
		controller.log.Error("failed to write json error response", ErrTransfers.Wrap(err))
	}
}

// Estimate returns approximate information about transfer fee and time.
func (controller *Transfers) Estimate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	w.Header().Set("Content-Type", "application/json")

	params := mux.Vars(r)
	senderNetwork := networks.Name(params["sender-network"])
	if err := senderNetwork.Validate(); err != nil {
		controller.serveError(w, http.StatusBadRequest, ErrTransfers.Wrap(errs.Combine(errors.New("sender-network parameter invalid"), err)))
		return
	}
	recipientNetwork := networks.Name(params["recipient-network"])
	if err := recipientNetwork.Validate(); err != nil {
		controller.serveError(w, http.StatusBadRequest, ErrTransfers.Wrap(errs.Combine(errors.New("recipient-network parameter invalid"), err)))
		return
	}
	tokenID, err := strconv.Atoi(params["token-id"])
	if err != nil {
		controller.serveError(w, http.StatusBadRequest, ErrTransfers.New("invalid token id"))
		return
	}
	amount := params["amount"]
	// TODO: implemenmt amount validation.

	preview, err := controller.transfers.Estimate(ctx, senderNetwork, recipientNetwork, uint32(tokenID), amount)
	if err != nil {
		controller.log.Error("could not estimate pending transfer", ErrTransfers.Wrap(err))
		controller.serveError(w, http.StatusInternalServerError, ErrTransfers.Wrap(err))
		return
	}

	if err = json.NewEncoder(w).Encode(preview); err != nil {
		controller.log.Error("failed to write json error response", ErrTransfers.Wrap(err))
	}
}

// History returns paginated list of transfers.
func (controller *Transfers) History(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	w.Header().Set("Content-Type", "application/json")

	query := r.URL.Query()
	offset, err := strconv.Atoi(query.Get("offset"))
	if err != nil {
		controller.serveError(w, http.StatusBadRequest, ErrTransfers.Wrap(errs.Combine(errors.New("offset parameter invalid"), err)))
		return
	}

	limit, err := strconv.Atoi(query.Get("limit"))
	if err != nil {
		controller.serveError(w, http.StatusBadRequest, ErrTransfers.Wrap(errs.Combine(errors.New("limit parameter invalid"), err)))
		return
	}

	networkID, err := strconv.Atoi(query.Get("network-id"))
	if err != nil {
		controller.serveError(w, http.StatusBadRequest, ErrTransfers.Wrap(errs.Combine(errors.New("network-id parameter invalid"), err)))
		return
	}

	params := mux.Vars(r)

	signatureHex := params["signature-hex"]
	signature, err := hex.DecodeString(signatureHex)
	if err != nil {
		controller.serveError(w, http.StatusBadRequest, ErrTransfers.Wrap(errs.Combine(errors.New("signature parameter is not in hex format"), err)))
		return
	}

	pubKeyHex := params["pub-key-hex"]
	pubKey, err := hex.DecodeString(pubKeyHex)
	if err != nil {
		controller.serveError(w, http.StatusBadRequest, ErrTransfers.Wrap(errs.Combine(errors.New("public key parameter is not in hex format"), err)))
		return
	}

	history, err := controller.transfers.History(ctx, uint64(offset), uint64(limit), signature, pubKey, uint32(networkID))
	if err != nil {
		controller.log.Error("could not cancel pending transfer", ErrTransfers.Wrap(err))
		controller.serveError(w, http.StatusInternalServerError, ErrTransfers.Wrap(err))
		return
	}

	if err = json.NewEncoder(w).Encode(history); err != nil {
		controller.log.Error("failed to write json error response", ErrTransfers.Wrap(err))
	}
}

// BridgeInSignature returns signature for user to send bridgeIn transaction.
func (controller *Transfers) BridgeInSignature(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	w.Header().Set("Content-Type", "application/json")

	request := struct {
		Sender      networks.Address `json:"sender"`
		TokenID     uint32           `json:"tokenId"`
		Amount      string           `json:"amount"`
		Destination networks.Address `json:"destination"`
	}{}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		controller.serveError(w, http.StatusBadRequest, ErrTransfers.Wrap(err))
		return
	}

	signature, err := controller.transfers.BridgeInSignature(ctx, transfers.BridgeInSignatureRequest{
		Sender:      request.Sender,
		TokenID:     request.TokenID,
		Amount:      request.Amount,
		Destination: request.Destination,
	})
	if err != nil {
		controller.log.Error("could not get bridge in signature", ErrTransfers.Wrap(err))
		controller.serveError(w, http.StatusInternalServerError, ErrTransfers.Wrap(err))
		return
	}

	response := struct {
		Token        string           `json:"token"`
		Amount       string           `json:"amount"`
		GasComission string           `json:"gasComission"`
		Destination  networks.Address `json:"destination"`
		Deadline     string           `json:"deadline"`
		Nonce        uint64           `json:"nonce"`
		Signature    string           `json:"signature"`
	}{
		Token:        hex.EncodeToString(signature.Token),
		Amount:       signature.Amount,
		GasComission: signature.GasCommission,
		Destination:  signature.Destination,
		Deadline:     signature.Deadline,
		Nonce:        signature.Nonce,
		Signature:    fmt.Sprintf("0x%s", hex.EncodeToString(signature.Signature)),
	}

	if err = json.NewEncoder(w).Encode(response); err != nil {
		controller.log.Error("failed to write json error response", ErrTransfers.Wrap(err))
	}
}

// CancelSignature returns signature for user to return funds.
func (controller *Transfers) CancelSignature(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	w.Header().Set("Content-Type", "application/json")

	params := mux.Vars(r)

	transferID, err := strconv.ParseUint(params["transfer-id"], 10, 64)
	if err != nil {
		controller.serveError(w, http.StatusBadRequest, ErrTransfers.New("invalid transfer id"))
		return
	}

	signatureParam := params["signature"]
	publicKeyParam := params["public-key"]

	networkID, err := strconv.Atoi(params["network-id"])
	if err != nil {
		controller.serveError(w, http.StatusBadRequest, ErrTransfers.New("invalid network id"))
		return
	}

	signature, err := networks.StringToBytes(networks.ID(networkID), signatureParam)
	if err != nil {
		controller.serveError(w, http.StatusBadRequest, ErrTransfers.Wrap(err))
		return
	}

	publicKey, err := networks.StringToBytes(networks.ID(networkID), publicKeyParam)
	if err != nil {
		controller.serveError(w, http.StatusBadRequest, ErrTransfers.Wrap(err))
		return
	}

	signatureResponse, err := controller.transfers.CancelSignature(ctx, transfers.CancelSignatureRequest{
		TransferID: transferID,
		Signature:  signature,
		NetworkID:  uint32(networkID),
		PublicKey:  publicKey,
	})
	if err != nil {
		controller.log.Error("could not get cancel signature", ErrTransfers.Wrap(err))
		controller.serveError(w, http.StatusInternalServerError, ErrTransfers.Wrap(err))
		return
	}

	response := struct {
		Status     string `json:"status"`
		Nonce      uint64 `json:"nonce"`
		Signature  string `json:"signature"`
		Token      string `json:"token"`
		Recipient  string `json:"recipient"`
		Commission string `json:"commission"`
		Amount     string `json:"amount"`
	}{
		Status:     signatureResponse.Status,
		Nonce:      signatureResponse.Nonce,
		Signature:  fmt.Sprintf("0x%s", networks.BytesToString(networks.ID(networkID), signatureResponse.Signature)),
		Token:      networks.BytesToString(networks.ID(networkID), signatureResponse.Token),
		Recipient:  networks.BytesToString(networks.ID(networkID), signatureResponse.Recipient),
		Commission: signatureResponse.Commission,
		Amount:     signatureResponse.Amount,
	}

	if err = json.NewEncoder(w).Encode(response); err != nil {
		controller.log.Error("failed to write json error response", ErrTransfers.Wrap(err))
	}
}

// serveError replies to the request with specific code and error message.
func (controller *Transfers) serveError(w http.ResponseWriter, status int, err error) {
	w.WriteHeader(status)

	response := ErrorResponse{
		Error: err.Error(),
	}

	if err = json.NewEncoder(w).Encode(response); err != nil {
		controller.log.Error("failed to write json error response", err)
	}
}
