package networks

import (
	"context"
	"encoding/hex"
	"errors"
)

var (
	// ErrTransactionTypeInvalid indicates that transaction type is not valid.
	ErrTransactionTypeInvalid = errors.New("network is not supported or its type invalid")
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
)

// Name defines list of possible blockchain network interoperability names.
type Name string

const (
	// NameCasper describes Casper network name.
	NameCasper Name = "CASPER"
	// NameEth describes Eth network name.
	NameEth Name = "ETH"
	// NameCasperTest describes Casper test network name.
	NameCasperTest Name = "CASPER-TEST"
	// NameGoerli describes Goerli network name.
	NameGoerli Name = "GOERLI"
)

// String converts Name type to string.
func (n Name) String() string {
	return string(n)
}

// TypeID defines list of possible blockchain network interoperability type ids.
type TypeID int

const (
	// TypeIDEVM describes EVM compatible network id.
	TypeIDEVM TypeID = 0
	// TypeIDCasper describes Casper network id.
	TypeIDCasper TypeID = 1
)

// Validate validates supported network type.
func (network Type) Validate() error {
	if network == TypeEVM || network == TypeCasper {
		return nil
	}

	return ErrTransactionTypeInvalid
}

// Validate validates supported network name.
func (network Name) Validate() error {
	if network == NameCasper || network == NameEth || network == NameCasperTest ||
		network == NameGoerli {
		return nil
	}

	return ErrTransactionNameInvalid
}

// NetworkIDToNetworkType describes id-to-type ratio for network.
var NetworkIDToNetworkType = map[TypeID]Type{
	TypeIDEVM:    TypeEVM,
	TypeIDCasper: TypeCasper,
}

// NetworkTypeToNetworkID describes type-to-id ratio for network.
var NetworkTypeToNetworkID = map[Type]TypeID{
	TypeEVM:    TypeIDEVM,
	TypeCasper: TypeIDCasper,
}

// ID defines list of possible blockchain network interoperability ids.
type ID int

const (
	// IDCasper describes Casper network id.
	IDCasper ID = 0
	// IDEth describes Eth network id.
	IDEth ID = 1
	// IDCasperTest describes Casper test network id.
	IDCasperTest ID = 4
	// IDGoerli describes Goerli network id.
	IDGoerli ID = 5
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

// StringToBytes converts signature/public key to bytes depending on the given network.
func StringToBytes(networkID ID, signatureStr string) ([]byte, error) {
	switch networkID {
	case IDCasper, IDEth, IDCasperTest, IDGoerli:
		return hex.DecodeString(signatureStr)
	default:
		return nil, ErrTransactionNameInvalid
	}
}

// BytesToString converts signature/public key from bytes to string depending on the given network.
func BytesToString(networkID ID, signatureBytes []byte) string {
	switch networkID {
	case IDCasper, IDEth, IDCasperTest, IDGoerli:
		return hex.EncodeToString(signatureBytes)
	default:
		return ""
	}
}
