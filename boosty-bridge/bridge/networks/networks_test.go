package networks_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tricorn/bridge/networks"
)

func TestNetworkIsValid(t *testing.T) {
	var tests = []struct {
		network networks.Name
		err     error
	}{
		{network: networks.NameCasper, err: nil},
		{network: networks.NameEth, err: nil},
		{network: networks.NameSolana, err: nil},
		{network: networks.NamePolygon, err: nil},
		{network: networks.NameCasperTest, err: nil},
		{network: networks.NameGoerli, err: nil},
		{network: networks.NameSolanaTest, err: nil},
		{network: networks.NameMumbai, err: nil},
		{network: networks.NameAvalanche, err: nil},
		{network: networks.NameAvalancheTest, err: nil},
		{network: networks.NameBNB, err: nil},
		{network: networks.NameBNBTest, err: nil},
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
		result      string
	}{
		{
			network:     networks.IDCasper,
			strToDecode: "013c0c1847d1c410338ab9b4ee0919c181cf26085997ff9c797e8a1ae5b02ddf23",
		},
		{
			network:     networks.IDCasper,
			strToDecode: "account-hash-013c0c1847d1c410338ab9b4ee0919c181cf26085997ff9c797e8a1ae5b02ddf23",
			result:      "013c0c1847d1c410338ab9b4ee0919c181cf26085997ff9c797e8a1ae5b02ddf23",
		},
		{
			network:     networks.IDCasper,
			strToDecode: "hash-013c0c1847d1c410338ab9b4ee0919c181cf26085997ff9c797e8a1ae5b02ddf23",
			result:      "013c0c1847d1c410338ab9b4ee0919c181cf26085997ff9c797e8a1ae5b02ddf23",
		},
		{
			network:     networks.IDEth,
			strToDecode: "e5bfc49E60a62AB039189D14b148ABEb80403460",
		},
		{
			network:     networks.IDSolana,
			strToDecode: "7j1NJDkeg35gtRt4nco9fN4wak6M6rD16okMYgvLJeHp",
		},
	}

	for _, test := range tests {
		decodedBytes, err := networks.StringToBytes(test.network, test.strToDecode)
		require.NoError(t, err)

		decodedStr := networks.BytesToString(test.network, decodedBytes)

		if test.result == "" {
			equal := strings.EqualFold(test.strToDecode, decodedStr)
			assert.True(t, equal)
		} else {
			equal := strings.EqualFold(test.result, decodedStr)
			assert.True(t, equal)
		}
	}
}

func TestIsAddressValid(t *testing.T) {
	var tests = []struct {
		network networks.Type
		address string
		valid   bool
	}{
		{
			network: networks.TypeEVM,
			address: "0x3095F955Da700b96215CFfC9Bc64AB2e69eB7DAB",
			valid:   true,
		},
		{
			network: networks.TypeEVM,
			address: "0x3095F955Da700b96215CFfC9Bc64AB2e69eB7DAB_invalid",
			valid:   false,
		},
		{
			network: networks.TypeCasper,
			address: "01783c4d47a3030add05472a685b20c2f8ec2fe64b309b601347be0968167c0d67",
			valid:   true,
		},
		{
			network: networks.TypeCasper,
			address: "01783c4d47a3030add05472a685b20c2f8ec2fe64b309b601347be0968167c0d67_invalid",
			valid:   false,
		},
		{
			network: networks.TypeSolana,
			address: "JARehRjGUkkEShpjzfuV4ERJS25j8XhamL776FAktNGm",
			valid:   true,
		},
		{
			network: networks.TypeSolana,
			address: "JARehRjGUkkEShpjzfuV4ERJS25j8XhamL776FAktNG_invalid",
			valid:   false,
		},
		{
			network: "",
			address: "",
			valid:   false,
		},
	}

	for _, test := range tests {
		valid := test.network.IsAddressValid(test.address)
		assert.Equal(t, test.valid, valid)
	}
}

func TestDecodeAddress(t *testing.T) {
	var tests = []struct {
		network    networks.Type
		address    string
		err        error
		checkBytes bool
	}{
		{
			network:    networks.TypeEVM,
			address:    "0x3095F955Da700b96215CFfC9Bc64AB2e69eB7DAB",
			err:        nil,
			checkBytes: true,
		},
		{
			network: networks.TypeEVM,
			address: "0x3095F955Da700b96215CFfC9Bc64AB2e69eB7DAB_invalid",
			err:     errors.New("invalid EVM address"),
		},
		{
			network:    networks.TypeCasper,
			address:    "01783c4d47a3030add05472a685b20c2f8ec2fe64b309b601347be0968167c0d67",
			err:        nil,
			checkBytes: true,
		},
		{
			network: networks.TypeCasper,
			address: "01783c4d47a3030add05472a685b20c2f8ec2fe64b309b601347be0968167c0d67_invalid",
			err:     errors.New("invalid casper address"),
		},
		{
			network:    networks.TypeSolana,
			address:    "JARehRjGUkkEShpjzfuV4ERJS25j8XhamL776FAktNGm",
			err:        nil,
			checkBytes: true,
		},
		{
			network: networks.TypeSolana,
			address: "JARehRjGUkkEShpjzfuV4ERJS25j8XhamL776FAktNG_invalid",
			err:     errors.New("invalid base58 digit ('_')"),
		},
		{
			network: "",
			address: "",
			err:     errors.New("unsupported network type "),
		},
	}

	for _, test := range tests {
		addressBytes, err := test.network.DecodeAddress(test.address)
		assert.Equal(t, test.err, err)

		if test.checkBytes {
			assert.Greater(t, len(addressBytes), 0)
		}
	}
}

func TestDecodeHash(t *testing.T) {
	var tests = []struct {
		network    networks.Type
		txHash     string
		err        error
		checkBytes bool
	}{
		{
			network:    networks.TypeEVM,
			txHash:     "0x1468a0193e6123ee5977417b008296f98e0c78bb24d362384f6fd4e1c3094885",
			err:        nil,
			checkBytes: true,
		},
		{
			network: networks.TypeEVM,
			txHash:  "0x1468a0193e6123ee5977417b008296f98e0c78bb24d362384f6fd4e1c3094885_invalid",
			err:     errors.New("invalid EVM hash"),
		},
		{
			network:    networks.TypeCasper,
			txHash:     "7b734c6851cd1e5839bfa52f0467db36351f73f58396734a43bd8b91d4e0e109",
			err:        nil,
			checkBytes: true,
		},
		{
			network: networks.TypeCasper,
			txHash:  "7b734c6851cd1e5839bfa52f0467db36351f73f58396734a43bd8b91d4e0e109_invalid",
			err:     errors.New("invalid casper hash"),
		},
		{
			network:    networks.TypeSolana,
			txHash:     "4DFmkJaV6e71be9JfGfMZtd4eTjomKZE29ZNn8kV1UUkSFotYHTTq1H58fRkpjENN3xojCSXNdURHzgNVxHYZN2f",
			err:        nil,
			checkBytes: true,
		},
		{
			network: networks.TypeSolana,
			txHash:  "4DFmkJaV6e71be9JfGfMZtd4eTjomKZE29ZNn8kV1UUkSFotYHTTq1H58fRkpjENN3xojCSXNdURHzgNVxHYZN2f_invalid",
			err:     errors.New("invalid base58 digit ('_')"),
		},
		{
			network: "",
			txHash:  "",
			err:     errors.New("unsupported network type "),
		},
	}

	for _, test := range tests {
		addressBytes, err := test.network.DecodeHash(test.txHash)
		assert.Equal(t, test.err, err)

		if test.checkBytes {
			assert.Greater(t, len(addressBytes), 0)
		}
	}
}
