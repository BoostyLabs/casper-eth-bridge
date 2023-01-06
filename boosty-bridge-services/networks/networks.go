package networks

import (
	"context"
	"errors"
)

var (
	// ErrTransactionNameInvalid indicates that transaction type is not valid.
	ErrTransactionNameInvalid = errors.New("network is not supported or its name invalid")
)

// Bridge exposes access to the bridge back-end methods related to the networks.
type Bridge interface {
	// ConnectedNetworks returns all connected networks to the bridge.
	ConnectedNetworks(ctx context.Context) ([]Network, error)
	// SupportedTokens returns all supported tokens by certain network.
	SupportedTokens(ctx context.Context, networkID uint32) ([]Token, error)
}

// Network hold basic network characteristics.
type Network struct {
	ID uint32 `json:"id"`
	// internal network id string representation.
	Name      string `json:"name"`
	Type      Type   `json:"type"`
	IsTestnet bool   `json:"isTestnet"`
}

// Type defines list of possible blockchain network interoperability types.
type Type string

const (
	// TypeEVM describes EVM compatible network.
	TypeEVM Type = "EVM"
	// TypeCasper describes Casper network.
	TypeCasper Type = "CASPER"
	// TypeCasperTest describes test casper network.
	TypeCasperTest Type = "CASPER-TEST"
	// TypeSolana describes Solana network.
	TypeSolana Type = "SOLANA"
	// TypeGoerli describes Goerli network.
	TypeGoerli Type = "GOERLI"
)

// TypeID defines list of possible blockchain network interoperability type ids.
type TypeID int

const (
	// TypeIDEVM describes EVM compatible network id.
	TypeIDEVM TypeID = 0
	// TypeIDCasper describes Casper network id.
	TypeIDCasper TypeID = 1
	// TypeIDSolana describes Solana network id.
	TypeIDSolana TypeID = 2
)

// Validate validates supported network name.
func (network Type) Validate() error {
	if network == TypeEVM || network == TypeCasper || network == TypeCasperTest || network == TypeSolana || network == TypeGoerli {
		return nil
	}

	return ErrTransactionNameInvalid
}

// NetworkIDToNetworkType describes id-to-type ratio for network.
var NetworkIDToNetworkType = map[TypeID]Type{
	TypeIDEVM:    TypeEVM,
	TypeIDCasper: TypeCasper,
	TypeIDSolana: TypeSolana,
}

// NetworkTypeToNetworkID describes type-to-id ratio for network.
var NetworkTypeToNetworkID = map[Type]TypeID{
	TypeEVM:    TypeIDEVM,
	TypeCasper: TypeIDCasper,
	TypeSolana: TypeIDSolana,
}

// ID defines list of possible blockchain network interoperability ids.
type ID int

const (
	// IDCasper describes Casper network id.
	IDCasper ID = 0
	// IDEth describes Eth network id.
	IDEth ID = 1
	// IDSolana describes Solana network id.
	IDSolana ID = 2
)

// Token holds information about supported by golden-gate tokens.
type Token struct {
	ID        uint32      `json:"id"`
	ShortName string      `json:"shortName"`
	LongName  string      `json:"longName"`
	Wraps     []WrappedIn `json:"wraps"`
}

// WrappedIn holds information about wrapped version of the Token.
type WrappedIn struct {
	NetworkID            uint32 `json:"networkId"`
	SmartContractAddress string `json:"smartContractAddress"`
}

// Address stores network name with its address.
type Address struct {
	NetworkName string `json:"networkName,omitempty"`
	Address     string `json:"address,omitempty"`
}
