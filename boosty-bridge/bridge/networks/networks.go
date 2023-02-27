package networks

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/mr-tron/base58"

	"tricorn/pkg/hexutils"
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
// TODO: place TokenContract to Token type.
type Network struct {
	ID        ID   `json:"id"`
	Name      Name `json:"name"`
	Type      Type `json:"type"`
	IsTestnet bool `json:"isTestnet"`

	NodeAddress string `json:"nodeAddress"` // TODO: delete after casper wallet release 2.

	TokenContract  string `json:"tokenContract"`
	BridgeContract string `json:"bridgeContract"`

	GasLimit uint64 `json:"gasLimit"` // TODO: comment why do we need it.
}

// Type defines list of possible blockchain network interoperability types.
type Type string

const (
	// TypeEVM describes EVM compatible network.
	TypeEVM Type = "NT_EVM"
	// TypeCasper describes Casper network.
	TypeCasper Type = "NT_CASPER"
	// TypeSolana describes Solana network.
	TypeSolana Type = "NT_SOLANA"
)

// Validate validates supported network type.
func (network Type) Validate() error {
	if network == TypeEVM || network == TypeCasper || network == TypeSolana {
		return nil
	}

	return ErrTransactionTypeInvalid
}

// Name defines list of possible blockchain network interoperability names.
type Name string

const (
	// NameCasper describes Casper network name.
	NameCasper Name = "CASPER"
	// NameEth describes Eth network name.
	NameEth Name = "ETH"
	// NameSolana describes Solana network name.
	NameSolana Name = "SOLANA"
	// NamePolygon describes Polygon network name.
	NamePolygon Name = "POLYGON"
	// NameCasperTest describes Casper test network name.
	NameCasperTest Name = "CASPER-TEST"
	// NameGoerli describes Goerli network name.
	NameGoerli Name = "GOERLI"
	// NameSolanaTest describes Solana test network name.
	NameSolanaTest Name = "SOLANA-TEST"
	// NameMumbai describes Mumbai network name.
	NameMumbai Name = "MUMBAI"
	// NameBNB describes mainnet network in BNB smart chain.
	NameBNB Name = "BNB"
	// NameBNBTest describes test network in BNB smart chain.
	NameBNBTest Name = "BNB-TEST"
	// NameAvalanche describes Avalanche network name.
	NameAvalanche Name = "AVALANCHE"
	// NameAvalancheTest describes Avalanche test network name.
	NameAvalancheTest Name = "AVALANCHE-TEST"
)

// Validate validates supported network name.
func (name Name) Validate() error {
	if name == NameCasper || name == NameEth || name == NameSolana || name == NamePolygon || name == NameCasperTest ||
		name == NameGoerli || name == NameSolanaTest || name == NameMumbai || name == NameAvalanche ||
		name == NameAvalancheTest || name == NameBNB || name == NameBNBTest {
		return nil
	}

	return ErrTransactionNameInvalid
}

// String converts Name type to string.
func (name Name) String() string {
	return string(name)
}

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
	// IDPolygon describes Polygon network id.
	IDPolygon ID = 3
	// IDCasperTest describes Casper test network id.
	IDCasperTest ID = 4
	// IDGoerli describes Goerli network id.
	IDGoerli ID = 5
	// IDSolanaTest describes Solana test network id.
	IDSolanaTest ID = 6
	// IDMumbai describes Mumbai network id.
	IDMumbai ID = 7
	// IDBNB describes BNB smart chain network id.
	IDBNB ID = 8
	// IDBNBTest describes BNB smart chain test network id.
	IDBNBTest = 9
	// IDAvalanche describes Avalanche network id.
	IDAvalanche ID = 10
	// IDAvalancheTest describes Avalanche test network id.
	IDAvalancheTest ID = 11
)

// PublicKey represents public key of cross-chain account.
type PublicKey []byte

// IDToNetworkName describes id-to-name ratio for network.
var IDToNetworkName = map[ID]Name{
	IDCasper:        NameCasper,
	IDEth:           NameEth,
	IDSolana:        NameSolana,
	IDPolygon:       NamePolygon,
	IDCasperTest:    NameCasperTest,
	IDGoerli:        NameGoerli,
	IDSolanaTest:    NameSolanaTest,
	IDMumbai:        NameMumbai,
	IDBNB:           NameBNB,
	IDBNBTest:       NameBNBTest,
	IDAvalanche:     NameAvalanche,
	IDAvalancheTest: NameAvalancheTest,
}

// NetworkNameToID describes name-to-id ratio for network.
var NetworkNameToID = map[Name]ID{
	NameCasper:        IDCasper,
	NameEth:           IDEth,
	NameSolana:        IDSolana,
	NamePolygon:       IDPolygon,
	NameCasperTest:    IDCasperTest,
	NameGoerli:        IDGoerli,
	NameSolanaTest:    IDSolanaTest,
	NameMumbai:        IDMumbai,
	NameBNB:           IDBNB,
	NameBNBTest:       IDBNBTest,
	NameAvalanche:     IDAvalanche,
	NameAvalancheTest: IDAvalancheTest,
}

// Type returns network type by id.
func (networkID ID) Type() Type {
	switch networkID {
	case IDCasper, IDCasperTest:
		return TypeCasper
	case IDSolana, IDSolanaTest:
		return TypeSolana
	}

	return TypeEVM
}

// IsTestnet returns testnet value by network id.
func (networkID ID) IsTestnet() bool {
	return networkID == IDMumbai || networkID == IDCasperTest || networkID == IDGoerli || networkID == IDSolanaTest ||
		networkID == IDBNBTest || networkID == IDAvalancheTest
}

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
// TODO: place to signatures package after bridge will be rewritten.
func StringToBytes(networkID ID, signatureStr string) ([]byte, error) {
	switch networkID {
	case IDCasper, IDEth, IDPolygon, IDCasperTest, IDGoerli, IDMumbai, IDBNB, IDBNBTest, IDAvalanche, IDAvalancheTest:
		if hexutils.Has0xPrefix(signatureStr) {
			return hex.DecodeString(signatureStr[2:])
		}
		if hexutils.HasAccountHashPrefix(signatureStr) {
			return hex.DecodeString(signatureStr[13:])
		}
		if hexutils.HasHashPrefix(signatureStr) {
			return hex.DecodeString(signatureStr[5:])
		}
		return hex.DecodeString(signatureStr)
	case IDSolana, IDSolanaTest:
		return base58.Decode(signatureStr)
	default:
		return nil, ErrTransactionNameInvalid
	}
}

// BytesToString converts signature/public key from bytes to string depending on the given network.
// TODO: place to signatures package after bridge will be rewritten.
func BytesToString(networkID ID, signatureBytes []byte) string {
	switch networkID {
	case IDCasper, IDEth, IDPolygon, IDCasperTest, IDGoerli, IDMumbai, IDBNB, IDBNBTest, IDAvalanche, IDAvalancheTest:
		return hex.EncodeToString(signatureBytes)
	case IDSolana, IDSolanaTest:
		return base58.Encode(signatureBytes)
	default:
		return ""
	}
}

// IsAddressValid validates address/public key in given network.
func (network Type) IsAddressValid(address string) bool {
	switch network {
	case TypeEVM:
		return IsEVMHexAddress(address)
	case TypeCasper:
		return IsCasperHexAddress(address)
	case TypeSolana:
		_, err := base58.Decode(address)
		if err != nil {
			return false
		}

		return true
	default:
		return false
	}
}

// DecodeAddress decodes address/public key in given network to bytes.
func (network Type) DecodeAddress(address string) ([]byte, error) {
	switch network {
	case TypeEVM:
		if !IsEVMHexAddress(address) {
			return nil, errors.New("invalid EVM address")
		}

		if hexutils.Has0xPrefix(address) {
			address = address[2:]
		}

		return hex.DecodeString(address)
	case TypeCasper:
		if !IsCasperHexAddress(address) {
			return nil, errors.New("invalid casper address")
		}

		return hex.DecodeString(address)
	case TypeSolana:
		return base58.Decode(address)
	default:
		return nil, fmt.Errorf("unsupported network type %v", network)
	}
}

const (
	// EVMAddressLength defines length in bytes of wallet address in evm networks.
	EVMAddressLength = 20
	// CasperAddressLength defines length in bytes of public key in casper networks.
	CasperAddressLength = 32
	// EVMHashLength defines length in bytes of tx/block hash in evm networks.
	EVMHashLength = 32
	// CasperHashLength defines length in bytes of deploy/block hash in evm networks.
	CasperHashLength = 32
)

// IsEVMHexAddress validates evm address.
func IsEVMHexAddress(s string) bool {
	if hexutils.Has0xPrefix(s) {
		s = s[2:]
	}

	return len(s) == 2*EVMAddressLength && hexutils.IsHex(s)
}

// IsCasperHexAddress validates public key in casper network.
func IsCasperHexAddress(s string) bool {
	if !hexutils.HasTagPrefix(s) {
		return false
	}

	if len(s) < 3 {
		return false
	}

	// remove tag(first two items in string) which defines algorithm which used to create wallet.
	s = s[2:]
	return len(s) == 2*CasperAddressLength && hexutils.IsHex(s)
}

// IsEVMHash validates transaction/block hash in casper network.
func IsEVMHash(s string) bool {
	if hexutils.Has0xPrefix(s) {
		s = s[2:]
	}

	return len(s) == 2*EVMHashLength && hexutils.IsHex(s)
}

// IsCasperHash validates deploy/block hash in casper network.
func IsCasperHash(s string) bool {
	return len(s) == 2*CasperHashLength && hexutils.IsHex(s)
}

// DecodeHash decodes transaction/block hash in given network.
func (network Type) DecodeHash(txHash string) ([]byte, error) {
	switch network {
	case TypeEVM:
		// since txHash in evm networks starts with '0x' we should cut off this prefix(first two symbols).
		if !IsEVMHash(txHash) {
			return nil, errors.New("invalid EVM hash")
		}

		if hexutils.Has0xPrefix(txHash) {
			txHash = txHash[2:]
		}

		return hex.DecodeString(txHash)
	case TypeCasper:
		if !IsCasperHash(txHash) {
			return nil, errors.New("invalid casper hash")
		}

		return hex.DecodeString(txHash)
	case TypeSolana:
		return base58.Decode(txHash)
	default:
		return nil, fmt.Errorf("unsupported network type %v", network)
	}
}
