// Copyright (C) 2022 Creditor Corp. Group.
// See LICENSE for copying information.

package eventparsing

import (
	"encoding/hex"
	"math/big"
	"strings"

	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/internal/reverse"
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

// GetUserWalletAddress returns user wallet address from event data.
func (e *EventData) GetUserWalletAddress() string {
	return e.getNextParam(LengthSelectorTag.Int(), LengthSelectorAddress.Int())
}
