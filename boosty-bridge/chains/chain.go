// Copyright (C) 2022 Creditor Corp. Group.
// See LICENSE for copying information.

package chains

import (
	"context"
	"math/big"

	"github.com/google/uuid"

	"tricorn/bridge/networks"
	"tricorn/signer"
)

// Bridge describes communication between Connector and Bridge.
type Bridge interface {
	// Sign returns signed data for specific network.
	Sign(ctx context.Context, req SignRequest) ([]byte, error)
	// PublicKey returns public key for specific network.
	PublicKey(ctx context.Context, networkId networks.Type) ([]byte, error)
}

// Connector describes behaviour of connector.
type Connector interface {
	// Network returns supported by connector network.
	Network(ctx context.Context) networks.Network
	// KnownTokens returns tokens known by this connector.
	KnownTokens(ctx context.Context) Tokens
	// BridgeOut initiates outbound bridge transaction.
	BridgeOut(ctx context.Context, req TokenOutRequest) ([]byte, error)
	// ReadEvents reads real-time events from node and old events from blocks and notifies subscribers.
	ReadEvents(ctx context.Context, fromBlock uint64) error
	// EstimateTransfer estimates a potential transfer.
	EstimateTransfer(ctx context.Context) (Estimation, error)
	// GetChainName returns chain name.
	GetChainName() networks.Name
	// BridgeInSignature returns signature for user to send bridgeIn transaction.
	BridgeInSignature(context.Context, BridgeInSignatureRequest) (BridgeInSignatureResponse, error)
	// CancelSignature returns signature for user to return funds.
	CancelSignature(context.Context, CancelSignatureRequest) (CancelSignatureResponse, error)

	// TODO: get rid of what is below.

	// AddEventSubscriber adds subscriber to event publisher.
	AddEventSubscriber() EventSubscriber
	// RemoveEventSubscriber removes publisher subscriber.
	RemoveEventSubscriber(id uuid.UUID)
	// Notify notifies all subscribers with events.
	Notify(ctx context.Context, event EventVariant)

	// CloseClient closes HTTP node client.
	CloseClient()
}

// SignRequest describes request for data signing.
type SignRequest struct {
	NetworkId networks.Type
	Data      []byte
	DataType  signer.Type
}

// EventVariant describes one out of two event variants.
type EventVariant struct {
	Type          EventType
	EventFundsIn  EventFundsIn
	EventFundsOut EventFundsOut
}

// Block returns block on which event occurred.
func (e EventVariant) Block() uint64 {
	switch e.Type {
	case EventTypeIn:
		return e.EventFundsIn.Tx.BlockNumber
	case EventTypeOut:
		return e.EventFundsOut.Tx.BlockNumber
	default:
		return 0
	}
}

// EventFundsIn describes event of bridge in method in format required by bridge.
type EventFundsIn struct {
	From   []byte
	To     networks.Address
	Amount string
	Token  []byte
	Tx     TransactionInfo
}

// EventFundsOut describes event of bridge out method in format required by bridge.
type EventFundsOut struct {
	From   networks.Address
	To     []byte
	Amount string
	Token  []byte
	Tx     TransactionInfo
}

// EventType defines list of possible event type for our connector.
type EventType int

const (
	// EventTypeIn defines that event type is 0. That is, this event arrived after calling the bridge in method in our contract.
	EventTypeIn EventType = 0
	// EventTypeOut defines that event type is 1. That is, this event arrived after calling the bridge out method in our contract.
	EventTypeOut EventType = 1
)

// Int returns int value from EventType type.
func (eventType EventType) Int() int {
	return int(eventType)
}

// TransactionInfo describes transaction details.
type TransactionInfo struct {
	Hash        []byte
	BlockNumber uint64
	Sender      []byte
}

// Tokens describes tokens supported by connector.
type Tokens struct {
	Tokens []Token
}

// Token describes token supported by connector.
type Token struct {
	ID      uint32
	Address []byte
}

// TokenOutRequest describes values to initiate outbound bridge transaction.
type TokenOutRequest struct {
	Amount        *big.Int
	Token         []byte
	To            []byte
	From          networks.Address
	TransactionID *big.Int
}

// TokenOutResponse describes hash of transaction after outbound bridge transaction was initiated.
type TokenOutResponse struct {
	Txhash []byte
}

// Transfer describes the values needed to estimate a transaction.
type Transfer struct {
	RecipientNetwork string
	TokenID          uint32
	Amount           string
}

// Estimation describes the values that are got after the estimation of the transaction.
type Estimation struct {
	Fee                   string
	FeePercentage         string
	EstimatedConfirmation uint32
}

// BridgeInSignatureRequest describes the values needed to generate bridge in signature.
type BridgeInSignatureRequest struct {
	User          []byte
	Nonce         *big.Int
	Token         string
	Amount        *big.Int
	Destination   networks.Address
	GasCommission *big.Int
}

// BridgeInSignatureResponse describes the values needed to send bridge in transaction.
type BridgeInSignatureResponse struct {
	Token         string
	Amount        *big.Int
	GasCommission string
	Destination   networks.Address
	Deadline      string
	Nonce         *big.Int
	Signature     []byte
}

// CancelSignatureRequest describes the values needed to generate cancel signature.
type CancelSignatureRequest struct {
	Nonce      *big.Int
	Token      []byte
	Recipient  []byte
	Commission *big.Int
	Amount     *big.Int
}

// CancelSignatureResponse describes the values needed to cancel transaction.
type CancelSignatureResponse struct {
	Signature []byte
}
