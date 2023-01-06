package envparse

import (
	"reflect"

	"github.com/caarlos0/env/v6"
	"github.com/ethereum/go-ethereum/common"
	"github.com/zeebo/errs"
)

// InvalidEthAddress indicates that given eth address is invalid.
var InvalidEthAddress = errs.New("invalid ethereum address")

// ethAddress returns env parsing options for ethereum address.
func ethAddress() map[reflect.Type]env.ParserFunc {
	parseOpt := make(map[reflect.Type]env.ParserFunc)

	parseAddrFunc := func(v string) (interface{}, error) {
		ok := common.IsHexAddress(v)
		if !ok {
			return nil, InvalidEthAddress
		}

		address := common.HexToAddress(v)

		return address, nil
	}

	parseOpt[reflect.TypeOf(common.Address{})] = parseAddrFunc

	return parseOpt
}

// ethHash returns env parsing options for hash.
func ethHash() map[reflect.Type]env.ParserFunc {
	parseOpt := make(map[reflect.Type]env.ParserFunc)

	parseHashFunc := func(v string) (interface{}, error) {
		hash := common.HexToHash(v)
		return hash, nil
	}

	parseOpt[reflect.TypeOf(common.Hash{})] = parseHashFunc

	return parseOpt
}

// EthParseOpts returns options for parsing ethereum address and hash.
func EthParseOpts() map[reflect.Type]env.ParserFunc {
	parseOpt := make(map[reflect.Type]env.ParserFunc)

	for key, val := range ethAddress() {
		parseOpt[key] = val
	}

	for key, val := range ethHash() {
		parseOpt[key] = val
	}

	return parseOpt
}
