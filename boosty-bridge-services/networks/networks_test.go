package networks_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/networks"
)

func TestNetworkIsValid(t *testing.T) {
	var tests = []struct {
		network networks.Type
		err     error
	}{
		{network: networks.TypeEVM, err: nil},
		{network: networks.TypeCasper, err: nil},
		{network: networks.TypeCasperTest, err: nil},
		{network: networks.TypeSolana, err: nil},
		{network: "qwe", err: networks.ErrTransactionNameInvalid},
		{network: "", err: networks.ErrTransactionNameInvalid},
	}

	for _, test := range tests {
		err := test.network.Validate()
		assert.Equal(t, err, test.err)
	}
}
