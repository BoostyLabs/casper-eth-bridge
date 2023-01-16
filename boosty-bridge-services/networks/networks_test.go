package networks_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/networks"
)

func TestNetworkIsValid(t *testing.T) {
	var tests = []struct {
		network networks.Name
		err     error
	}{
		{network: networks.NameCasper, err: nil},
		{network: networks.NameEth, err: nil},
		{network: networks.NameCasperTest, err: nil},
		{network: networks.NameGoerli, err: nil},
		{network: "qwe", err: networks.ErrTransactionNameInvalid},
		{network: "", err: networks.ErrTransactionNameInvalid},
	}

	for _, test := range tests {
		err := test.network.Validate()
		assert.Equal(t, err, test.err)
	}
}

func TestConverting(t *testing.T) {
	var tests = []struct {
		network     networks.ID
		strToDecode string
	}{
		{
			network:     networks.IDCasper,
			strToDecode: "013c0c1847d1c410338ab9b4ee0919c181cf26085997ff9c797e8a1ae5b02ddf23",
		},
		{
			network:     networks.IDEth,
			strToDecode: "e5bfc49E60a62AB039189D14b148ABEb80403460",
		},
	}

	for _, test := range tests {
		decodedBytes, err := networks.StringToBytes(test.network, test.strToDecode)
		require.NoError(t, err)

		decodedStr := networks.BytesToString(test.network, decodedBytes)
		equal := strings.EqualFold(test.strToDecode, decodedStr)
		assert.True(t, equal)
	}
}
