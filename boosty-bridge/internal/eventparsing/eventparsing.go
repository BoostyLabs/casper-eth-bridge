// Copyright (C) 2022 Creditor Corp. Group.
// See LICENSE for copying information.

package eventparsing

import (
	"encoding/hex"
	"math/big"
	"strings"

	"tricorn/internal/reverse"
)

// LengthSelector defines list of all possible length selectors.
type LengthSelector int

const (
	// LengthSelectorString defines that length of string selector is 8.
	LengthSelectorString LengthSelector = 8
	// LengthSelectorAddress defines that length of address selector is 64.
	LengthSelectorAddress LengthSelector = 64
	// LengthSelectorU256 defines that length of uint256 selector is 2.
	LengthSelectorU256 LengthSelector = 2
	// LengthSelectorU128 defines that length of uint128 selector is 2.
	LengthSelectorU128 LengthSelector = 2
	// LengthSelectorTag defines that length of tag selector is 2.
	LengthSelectorTag LengthSelector = 2
	// LengthSelectorAddressInBytes defines that length of address in bytes selector is 1.
	LengthSelectorAddressInBytes LengthSelector = 1
)

// Int returns int value from LengthSelector type.
func (l LengthSelector) Int() int {
	return int(l)
}

// Tag defines list of all possible tags.
type Tag string

const (
	// TagAccount defines that tag belongs to the account.
	TagAccount Tag = "00"
	// TagHash defines that tag belongs to the hash.
	TagHash Tag = "01"
)

// String returns string value from Tag type.
func (a Tag) String() string {
	return string(a)
}

var (
	// SuffixOfSelectorForDynamicField defines that suffix of selector for dynamic field is 0.
	SuffixOfSelectorForDynamicField string = "0"
	// SymbolsInByte defines that there are 2 symbols in a byte.
	SymbolsInByte int = 2
)

// EventData defines event data with offset length for pre-use.
type EventData struct {
	Bytes  string
	offset int
}

// getNextParam returns next parameter for specified data length.
func (e *EventData) getNextParam(offset int, limit int) string {
	e.offset += offset
	param := e.Bytes[e.offset : e.offset+limit]
	e.offset += limit
	return param
}

// GetEventType returns event type from event data.
func (e *EventData) GetEventType() (int, error) {
	eventTypeHex := e.getNextParam(LengthSelectorString.Int(), LengthSelectorTag.Int())

	eventTypeBytes, err := hex.DecodeString(eventTypeHex)
	if err != nil {
		return 0, err
	}

	eventType := big.NewInt(0).SetBytes(eventTypeBytes)

	return int(eventType.Int64()), nil
}

// GetTokenContractAddress returns token contract address from event data.
func (e *EventData) GetTokenContractAddress() string {
	return e.getNextParam(0, LengthSelectorAddress.Int())
}

// GetChainName returns chain name from event data.
func (e *EventData) GetChainName() (string, error) {
	chainNameLengthHex := e.getNextParam(0, LengthSelectorString.Int())

	for i := 0; i < LengthSelectorString.Int(); i++ {
		chainNameLengthHex = strings.TrimSuffix(chainNameLengthHex, SuffixOfSelectorForDynamicField)
	}

	chainNameLengthBytes, err := hex.DecodeString(chainNameLengthHex)
	if err != nil {
		return "", err
	}

	chainNameLength := big.NewInt(0).SetBytes(chainNameLengthBytes)

	chainNameHex := e.getNextParam(0, int(chainNameLength.Int64())*SymbolsInByte)
	chainNameBytes, err := hex.DecodeString(chainNameHex)

	return string(chainNameBytes), err
}

// GetChainAddress returns chain address from event data.
func (e *EventData) GetChainAddress() (string, error) {
	chainAddressLengthHex := e.getNextParam(0, LengthSelectorString.Int())

	for i := 0; i < LengthSelectorString.Int(); i++ {
		chainAddressLengthHex = strings.TrimSuffix(chainAddressLengthHex, SuffixOfSelectorForDynamicField)
	}

	chainAddressLengthBytes, err := hex.DecodeString(chainAddressLengthHex)
	if err != nil {
		return "", err
	}

	chainAddressLength := big.NewInt(0).SetBytes(chainAddressLengthBytes)

	chainAddressHex := e.getNextParam(0, int(chainAddressLength.Int64())*SymbolsInByte)
	chainAddressBytes, err := hex.DecodeString(chainAddressHex)

	return string(chainAddressBytes), err
}

// GetAmount returns amount from event data.
func (e *EventData) GetAmount() (int, error) {
	amountLengthHex := e.getNextParam(0, LengthSelectorU256.Int())

	amountLengthBytes, err := hex.DecodeString(amountLengthHex)
	if err != nil {
		return 0, err
	}

	amountLength := big.NewInt(0).SetBytes(amountLengthBytes)

	amountHex := e.getNextParam(0, int(amountLength.Int64())*SymbolsInByte)
	amountBytes, err := hex.DecodeString(amountHex)
	if err != nil {
		return 0, err
	}

	amountBytes = reverse.Bytes(amountBytes)
	amount := big.NewInt(0).SetBytes(amountBytes)

	return int(amount.Int64()), err
}

// GetGasCommission returns gas commission from event data.
func (e *EventData) GetGasCommission() (int, error) {
	gasCommissionLengthHex := e.getNextParam(0, LengthSelectorU256.Int())

	gasCommissionLengthBytes, err := hex.DecodeString(gasCommissionLengthHex)
	if err != nil {
		return 0, err
	}

	gasCommissionLength := big.NewInt(0).SetBytes(gasCommissionLengthBytes)

	gasCommissionHex := e.getNextParam(0, int(gasCommissionLength.Int64())*SymbolsInByte)
	gasCommissionBytes, err := hex.DecodeString(gasCommissionHex)
	if err != nil {
		return 0, err
	}

	gasCommissionBytes = reverse.Bytes(gasCommissionBytes)
	gasCommission := big.NewInt(0).SetBytes(gasCommissionBytes)

	return int(gasCommission.Int64()), err
}

// GetStableCommissionPercent returns stable commission percent from event data.
func (e *EventData) GetStableCommissionPercent() (int, error) {
	stableCommissionPercentLengthHex := e.getNextParam(0, LengthSelectorU256.Int())

	stableCommissionPercentLengthBytes, err := hex.DecodeString(stableCommissionPercentLengthHex)
	if err != nil {
		return 0, err
	}

	stableCommissionPercentLength := big.NewInt(0).SetBytes(stableCommissionPercentLengthBytes)

	stableCommissionPercentHex := e.getNextParam(0, int(stableCommissionPercentLength.Int64())*SymbolsInByte)
	stableCommissionPercentBytes, err := hex.DecodeString(stableCommissionPercentHex)
	if err != nil {
		return 0, err
	}

	stableCommissionPercentBytes = reverse.Bytes(stableCommissionPercentBytes)
	stableCommissionPercent := big.NewInt(0).SetBytes(stableCommissionPercentBytes)

	return int(stableCommissionPercent.Int64()), err
}

// GetNonce returns nonce from event data.
func (e *EventData) GetNonce() (int, error) {
	nonceLengthHex := e.getNextParam(0, LengthSelectorU128.Int())

	nonceLengthBytes, err := hex.DecodeString(nonceLengthHex)
	if err != nil {
		return 0, err
	}

	nonceLength := big.NewInt(0).SetBytes(nonceLengthBytes)

	nonceHex := e.getNextParam(0, int(nonceLength.Int64())*SymbolsInByte)
	nonceBytes, err := hex.DecodeString(nonceHex)
	if err != nil {
		return 0, err
	}

	nonceBytes = reverse.Bytes(nonceBytes)
	nonce := big.NewInt(0).SetBytes(nonceBytes)

	return int(nonce.Int64()), err
}

// GetTransactionID returns transaction id from event data.
func (e *EventData) GetTransactionID() (int, error) {
	transactionIDLengthHex := e.getNextParam(0, LengthSelectorU256.Int())

	transactionIDLengthBytes, err := hex.DecodeString(transactionIDLengthHex)
	if err != nil {
		return 0, err
	}

	transactionIDLength := big.NewInt(0).SetBytes(transactionIDLengthBytes)

	transactionIDHex := e.getNextParam(0, int(transactionIDLength.Int64())*SymbolsInByte)
	transactionIDBytes, err := hex.DecodeString(transactionIDHex)
	if err != nil {
		return 0, err
	}

	transactionIDBytes = reverse.Bytes(transactionIDBytes)
	transactionID := big.NewInt(0).SetBytes(transactionIDBytes)

	return int(transactionID.Int64()), err
}

// GetUserWalletAddress returns user wallet address from event data.
func (e *EventData) GetUserWalletAddress() string {
	return e.getNextParam(LengthSelectorTag.Int(), LengthSelectorAddress.Int())
}
