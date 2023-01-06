// Copyright (C) 2022 Creditor Corp. Group.
// See LICENSE for copying information.

package solana

// Config contains solana configurable values.
type Config struct {
	NodeAddress              string  `env:"NODE_ADDRESS"`
	ChainName                string  `env:"CHAIN_NAME"`
	IsTestnet                bool    `env:"IS_TESTNET"`
	BridgeContractAddress    string  `env:"BRIDGE_CONTRACT_ADDRESS"`
	BridgeOutMethodName      string  `env:"BRIDGE_OUT_METHOD_NAME"`
	EventsFundIn             string  `env:"FUND_IN_EVENT_HASH"`
	EventsFundOut            string  `env:"FUND_OUT_EVENT_HASH"`
	GasIncreasingCoefficient float64 `env:"GAS_INCREASING_COEFFICIENT"`
	ConfirmationTime         uint32  `env:"CONFIRMATION_TIME"`
	FeePercentage            string  `env:"FEE_PERCENTAGE"`
	GasLimit                 uint64  `env:"GAS_LIMIT"` // TODO: count by tx.
}
