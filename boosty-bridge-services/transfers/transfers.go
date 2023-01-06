package transfers

import (
	"context"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/networks"
)

// Bridge exposes access to the bridge back-end methods related to transfers.
type Bridge interface {
	// Estimate returns approximate information about transfer fee and time.
	Estimate(ctx context.Context, sender, recipient networks.Type, tokenID uint32, amount string) (Estimate, error)
	// Info returns list of transfers of triggering transaction.
	Info(ctx context.Context, txHash string) ([]Transfer, error)
	// Cancel cancels a transfer in the CONFIRMING status, returning the funds to the sender
	// after deducting the commission for issuing the transaction.
	Cancel(ctx context.Context, id ID, signature, pubKey []byte) error
	// History returns paginated list of transfers.
	History(ctx context.Context, offset, limit uint64, signature, pubKey []byte, networkID uint32) (Page, error)
	// BridgeInSignature returns signature for user to send bridgeIn transaction.
	BridgeInSignature(ctx context.Context, req BridgeInSignatureRequest) (BridgeInSignatureResponse, error)
}

// Transfer hold all information about transferring funds from one network to another.
type Transfer struct {
	ID           ID               `json:"id"`
	Amount       big.Int          `json:"amount"`
	Sender       networks.Address `json:"sender"`
	Recipient    networks.Address `json:"recipient"`
	Status       Status           `json:"status"`
	TriggeringTx StringTxHash     `json:"triggeringTx"`
	OutboundTx   StringTxHash     `json:"outboundTx"`
	CreatedAt    time.Time        `json:"createdAt"`
}

// Page holds operator page entity which is used to show listed page of operators.
type Page struct {
	Transfers  []Transfer `json:"transfers"`
	Offset     int64      `json:"offset"`
	Limit      int64      `json:"limit"`
	TotalCount int64      `json:"totalCount"`
}

// ID is a type-alias for transaction id.
type ID uint64

// Status type describes transfers status.
type Status string

const (
	// StatusConfirming indicates that transfer status is being confirmed.
	StatusConfirming Status = "CONFIRMING"
	// StatusCancelled indicates that transfer was cancelled.
	StatusCancelled Status = "CANCELLED"
	// StatusFinished indicates that transfer is finished.
	StatusFinished Status = "FINISHED"
)

// StringTxHash stores string representation of tx hash.
type StringTxHash struct {
	NetworkName string      `json:"networkName,omitempty"`
	Hash        common.Hash `json:"hash,omitempty"`
}

// Estimate holds approximate information about transfer fee and time.
type Estimate struct {
	Fee                       string `json:"fee"` // expected fee for this transfer.
	FeePercentage             string `json:"feePercentage"`
	EstimatedConfirmationTime string `json:"estimatedConfirmationTime"`
}

// BridgeInSignatureRequest describes the values needed to generate bridge in signature.
type BridgeInSignatureRequest struct {
	Sender      networks.Address
	TokenID     uint32
	Amount      string
	Destination networks.Address
}

// BridgeInSignatureResponse describes the values needed to send bridge in transaction.
type BridgeInSignatureResponse struct {
	Token        []byte
	Amount       string
	GasComission string
	Destination  networks.Address
	Deadline     string
	Nonce        uint64
	Signature    []byte
}
