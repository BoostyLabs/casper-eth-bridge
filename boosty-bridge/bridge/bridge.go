package bridge

import (
	"context"
	"errors"
	"math/big"

	"github.com/google/uuid"

	"tricorn/bridge/networks"
	"tricorn/bridge/transactions"
	"tricorn/bridge/transfers"
	"tricorn/chains"
	"tricorn/currencyrates"
	"tricorn/signer"
)

// authenticationMsg defines message which client signs to authenticate.
const authenticationMsg = "Bridge Authentication Proof"

var (
	// ErrNoNetworkBlock indicates that network block does not exist.
	ErrNoNetworkBlock = errors.New("network block does not exist")
	// ErrNoNetworkNonce indicates that network nonce does not exist.
	ErrNoNetworkNonce = errors.New("network nonce does not exist")
	// ErrNoNetworkToken indicates that network token does not exist.
	ErrNoNetworkToken = errors.New("network token does not exist")
	// ErrNoTokenTransfer indicates that token transfer does not exist.
	ErrNoTokenTransfer = errors.New("token transfer does not exist")
	// ErrNoToken indicates that token does not exist.
	ErrNoToken = errors.New("token does not exist")
	// ErrNoTransaction indicates that transaction does not exist.
	ErrNoTransaction = errors.New("transaction does not exist")
	// ErrTransactionAlreadyExists indicates that the transaction already exists.
	ErrTransactionAlreadyExists = errors.New("transaction already exists")
	// ErrNotConnectedNetwork indicates that network is not connected.
	ErrNotConnectedNetwork = errors.New("network is not connected")
	// ErrInvalidAmount indicates than invalid amount was received.
	ErrInvalidAmount = errors.New("received invalid amount")
	// ErrInvalidTransferStatus indicates about invalid transfer status for cancel transfer request.
	ErrInvalidTransferStatus = errors.New("invalid transfer status")
)

// Connector exposes access to the connector methods.
type Connector interface {
	// Network returns supported by connector network.
	Network(ctx context.Context) (networks.Network, error)
	// KnownTokens returns tokens known by this connector.
	KnownTokens(context.Context) (chains.Tokens, error)
	// EventStream initiates event stream from the network.
	EventStream(ctx context.Context, fromBlock uint64) error
	// BridgeOut initiates outbound bridge transaction.
	BridgeOut(context.Context, chains.TokenOutRequest) (chains.TokenOutResponse, error)
	// EstimateTransfer estimates a potential transfer.
	EstimateTransfer(context.Context, transfers.EstimateTransfer) (chains.Estimation, error)
	// BridgeInSignature returns signature for user to send bridgeIn transaction.
	BridgeInSignature(context.Context, BridgeInSignatureRequest) (BridgeInSignatureResponse, error)
	// CancelSignature returns signature for user to return funds.
	CancelSignature(context.Context, chains.CancelSignatureRequest) (chains.CancelSignatureResponse, error)

	// AddEventSubscriber adds subscriber to event publisher.
	AddEventSubscriber() EventSubscriber
	// RemoveEventSubscriber removes subscriber.
	RemoveEventSubscriber(id uuid.UUID)
	// Notify notifies all subscribers with events.
	Notify(ctx context.Context, event chains.EventVariant)
}

// CurrencyRates exposes access to the currency rates methods.
type CurrencyRates interface {
	// PriceStream is real time events streaming for tokens price.
	PriceStream(ctx context.Context, tokenPriceChan chan currencyrates.TokenPrice) error
}

// Signer describes the communication between bridge and signer.
type Signer interface {
	// Sign signs data for specific network.
	Sign(ctx context.Context, networkType networks.Type, data []byte, dataType signer.Type) ([]byte, error)
	// PublicKey returns public key for specific network.
	PublicKey(ctx context.Context, networkID networks.Type) (networks.PublicKey, error)
}

// DB provides access to all databases and database related functionality.
//
// architecture: Master Database.
type DB interface {
	// NetworkBlocks provides access to network blocks db.
	NetworkBlocks() networks.NetworkBlocks

	// Nonces provides access to network nonces db.
	Nonces() networks.Nonces

	// NetworkTokens provides access to network tokens db.
	NetworkTokens() networks.NetworkTokens

	// TokenTransfers provides access to token transfers db.
	TokenTransfers() transfers.TokenTransfers

	// Tokens provides access to tokens db.
	Tokens() Tokens

	// Transactions provides access to transactions db.
	Transactions() transactions.DB

	// Close closes underlying db connection.
	Close() error

	// CreateSchema create tables.
	CreateSchema(ctx context.Context) error
}

// Tokens is exposing access to tokens db.
//
// architecture: DB
type Tokens interface {
	// Create inserts token to database.
	Create(ctx context.Context, token Token) error
	// Get returns token by id from database.
	Get(ctx context.Context, id int64) (Token, error)
	// List returns list of tokens, supported by network, from database.
	List(ctx context.Context, networkID networks.ID) ([]Token, error)
	// Update updates token in database.
	Update(ctx context.Context, token Token) error
}

// Token describes token that is transferred between networks.
type Token struct {
	ID        int64
	ShortName string
	LongName  string
}

// BridgeInSignatureRequest describes the values needed to generate bridge in signature.
type BridgeInSignatureRequest struct {
	User          []byte
	Nonce         *big.Int
	Token         []byte
	Amount        *big.Int
	Destination   networks.Address
	GasCommission *big.Int
}

// BridgeInSignatureResponse describes the values needed to send bridge in transaction.
type BridgeInSignatureResponse struct {
	Token         []byte
	Amount        string
	GasCommission string
	Destination   networks.Address
	Deadline      string
	Nonce         *big.Int
	Signature     []byte
}
