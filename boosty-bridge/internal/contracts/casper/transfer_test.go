package casper_test

import (
	"context"
	"encoding/hex"
	"math/big"
	"os"
	"testing"
	"time"

	casper_ed25519 "github.com/casper-ecosystem/casper-golang-sdk/keypair/ed25519"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"

	"tricorn/bridge/networks"
	casper_chain "tricorn/chains/casper"
	"tricorn/internal/contracts/casper"
	"tricorn/pkg/casper-sdk/client"
	signer_service "tricorn/signer"
)

func TestCasper_BridgeIn(t *testing.T) {
	t.Skip("for manual testing")

	var (
		casperNodeAddress = "http://136.243.187.84:7777/rpc"

		privateKeySecp256k1ForSignature = "cc903a2179a5c47acef21d732c0693848c6c33e626fd6651b3773732bde6e127"
		// account-hash-daa2b596e0a496b04933e241e0567f2bcbecc829aa57d88cab096c28fd07dee2.
		privateKeyEd25519ForTransaction = "33e8798230a0ce481455b8b0affe5f96455fa157799095567ef220d55789a2e10ad302bfc22c0e606d94d98a3baa2c8eeedd1e148d9a20a4453bb8cc5e530a19"

		accountAddress            = "daa2b596e0a496b04933e241e0567f2bcbecc829aa57d88cab096c28fd07dee2"
		tokenContractAddress      = "3c0c1847d1c410338ab9b4ee0919c181cf26085997ff9c797e8a1ae5b02ddf23"
		bridgeContractPackageHash = "7225d70f3e197d78dd286a77bfa219e8573d62e6c4b62ade27ec14c69a88442f"
	)

	ctx := context.Background()
	casperClient := client.New(casperNodeAddress)

	privateKeyECDSA, err := crypto.HexToECDSA(privateKeySecp256k1ForSignature)
	require.NoError(t, err)

	privateKeyForTransferSigningBytes, err := hex.DecodeString(privateKeyEd25519ForTransaction)
	require.NoError(t, err)

	publicKey := make([]byte, casper.Ed25519PublicKeySize)
	copy(publicKey, privateKeyForTransferSigningBytes[casper.Ed25519PublicKeySize:])
	pair := casper_ed25519.ParseKeyPair(publicKey, privateKeyForTransferSigningBytes[:casper.Ed25519PublicKeySize])

	tokenContractAddressBytes, err := hex.DecodeString(tokenContractAddress)
	require.NoError(t, err)

	deadlineTime := time.Now().UTC().Add(time.Second * time.Duration(86400)).UnixMilli()
	deadline := big.NewInt(0).SetInt64(deadlineTime)

	var signature []byte
	t.Run("get bridgeIn signature", func(t *testing.T) {
		signer := casper.NewSigner(func(b []byte, _ signer_service.Type) ([]byte, error) {
			signature, err := crypto.Sign(b, privateKeyECDSA)
			return signature, err
		})

		bridgeHashBytes, err := hex.DecodeString(bridgeContractPackageHash)
		require.NoError(t, err)

		accountAddressBytes, err := hex.DecodeString(accountAddress)
		require.NoError(t, err)

		signature, err = signer.GetBridgeInSignature(ctx, casper_chain.BridgeInSignature{
			Prefix:             "TRICORN_BRIDGE_IN",
			BridgeHash:         bridgeHashBytes,
			TokenPackageHash:   tokenContractAddressBytes,
			AccountAddress:     accountAddressBytes,
			Amount:             big.NewInt(10000),
			GasCommission:      big.NewInt(1000),
			Deadline:           deadline,
			Nonce:              big.NewInt(112),
			DestinationChain:   "DEST",
			DestinationAddress: "DESTADDR",
		})
		require.NoError(t, err)
		require.NotEmpty(t, signature)
	})

	t.Run("bridgeIn", func(t *testing.T) {
		transfer := casper.NewTransfer(casperClient, func(b []byte) ([]byte, error) {
			casperSignature := pair.Sign(b)
			return casperSignature.SignatureData, nil
		})

		txHash, err := transfer.BridgeIn(ctx, casper.BridgeInRequest{
			PublicKey:                   pair.PublicKey(),
			ChainName:                   "CASPER-TEST",
			StandardPaymentForBridgeOut: 40000000000, // 40 CSPR.
			BridgeContractPackageHash:   bridgeContractPackageHash,
			TokenContractAddress:        tokenContractAddressBytes,
			Amount:                      big.NewInt(10000),
			GasCommission:               big.NewInt(1000),
			Deadline:                    deadline,
			Nonce:                       big.NewInt(112),
			DestinationChain:            "DEST",
			DestinationAddress:          "DESTADDR",
			Signature:                   signature,
		})
		require.NoError(t, err)
		require.NotEmpty(t, txHash)
	})
}

func TestCasper_SetSigner(t *testing.T) {
	t.Skip("for manual testing")

	var (
		casperNodeAddress = "http://136.243.187.84:7777/rpc"

		privateKeyEd25519ForTransaction = "1228fcc08c02bfe100543a2581f60b0ad0e09f4c53e81641f09a70850880a256c1a19239600bf8293462d99f2ec19d5d1b443c760f9bdb4d720554585b26139a"
		pathToPublicKeyFile             = "./public_key.in"

		bridgeContractPackageHash = "9299f58df67c2eff01e97f362996d35ab5393167e58c58360b1721cce95a7bbc"
	)

	ctx := context.Background()
	casperClient := client.New(casperNodeAddress)

	privateKeyForTransferSigningBytes, err := hex.DecodeString(privateKeyEd25519ForTransaction)
	require.NoError(t, err)

	publicKey := make([]byte, signer_service.PublicKeySize)
	copy(publicKey, privateKeyForTransferSigningBytes[signer_service.PublicKeySize:])
	pair := casper_ed25519.ParseKeyPair(publicKey, privateKeyForTransferSigningBytes[:signer_service.PublicKeySize])

	transfer := casper.NewTransfer(casperClient, func(b []byte) ([]byte, error) {
		casperSignature := pair.Sign(b)
		return casperSignature.SignatureData, nil
	})

	publicKeyFile, err := os.ReadFile(pathToPublicKeyFile)
	require.NoError(t, err)

	txHash, err := transfer.SetSigner(ctx, casper.SetSignerRequest{
		PublicKey:                   pair.PublicKey(),
		ChainName:                   "CASPER-TEST",
		StandardPaymentForBridgeOut: 2500000000, // 2.5 CSPR.
		BridgeContractPackageHash:   bridgeContractPackageHash,
		Value:                       string(publicKeyFile),
	})
	require.NoError(t, err)
	require.NotEmpty(t, txHash)
}

func TestCasper_SetStableCommissionPercent(t *testing.T) {
	t.Skip("for manual testing")

	var (
		casperNodeAddress = "http://136.243.187.84:7777/rpc"

		privateKeyEd25519ForTransaction = "1228fcc08c02bfe100543a2581f60b0ad0e09f4c53e81641f09a70850880a256c1a19239600bf8293462d99f2ec19d5d1b443c760f9bdb4d720554585b26139a"

		bridgeContractPackageHash = "9299f58df67c2eff01e97f362996d35ab5393167e58c58360b1721cce95a7bbc"
	)

	ctx := context.Background()
	casperClient := client.New(casperNodeAddress)

	privateKeyForTransferSigningBytes, err := hex.DecodeString(privateKeyEd25519ForTransaction)
	require.NoError(t, err)

	publicKey := make([]byte, signer_service.PublicKeySize)
	copy(publicKey, privateKeyForTransferSigningBytes[signer_service.PublicKeySize:])
	pair := casper_ed25519.ParseKeyPair(publicKey, privateKeyForTransferSigningBytes[:signer_service.PublicKeySize])

	transfer := casper.NewTransfer(casperClient, func(b []byte) ([]byte, error) {
		casperSignature := pair.Sign(b)
		return casperSignature.SignatureData, nil
	})

	txHash, err := transfer.SetStableCommissionPercent(ctx, casper.SetStableCommissionPercentRequest{
		PublicKey:                   pair.PublicKey(),
		ChainName:                   "CASPER-TEST",
		StandardPaymentForBridgeOut: 500000000, // 0.5 CSPR.
		BridgeContractPackageHash:   bridgeContractPackageHash,
		CommissionPercent:           big.NewInt(2),
	})
	require.NoError(t, err)
	require.Empty(t, txHash)
}

func TestCasper_BridgeOut(t *testing.T) {
	t.Skip("for manual testing")

	var (
		casperNodeAddress = "http://136.243.187.84:7777/rpc"

		privateKeyEd25519ForTransaction = "1228fcc08c02bfe100543a2581f60b0ad0e09f4c53e81641f09a70850880a256c1a19239600bf8293462d99f2ec19d5d1b443c760f9bdb4d720554585b26139a"

		toAccountAddress          = "daa2b596e0a496b04933e241e0567f2bcbecc829aa57d88cab096c28fd07dee2"
		tokenContractAddress      = "3c0c1847d1c410338ab9b4ee0919c181cf26085997ff9c797e8a1ae5b02ddf23"
		bridgeContractPackageHash = "9299f58df67c2eff01e97f362996d35ab5393167e58c58360b1721cce95a7bbc"
	)

	ctx := context.Background()
	casperClient := client.New(casperNodeAddress)

	privateKeyForTransferSigningBytes, err := hex.DecodeString(privateKeyEd25519ForTransaction)
	require.NoError(t, err)

	publicKey := make([]byte, signer_service.PublicKeySize)
	copy(publicKey, privateKeyForTransferSigningBytes[signer_service.PublicKeySize:])
	pair := casper_ed25519.ParseKeyPair(publicKey, privateKeyForTransferSigningBytes[:signer_service.PublicKeySize])

	transfer := casper.NewTransfer(casperClient, func(b []byte) ([]byte, error) {
		casperSignature := pair.Sign(b)
		return casperSignature.SignatureData, nil
	})

	txHash, err := transfer.BridgeOut(ctx, casper.BridgeOutRequest{
		PublicKey:                   pair.PublicKey(),
		ChainName:                   "CASPER-TEST",
		StandardPaymentForBridgeOut: 2700000000, // 2.7 CSPR.
		BridgeContractPackageHash:   bridgeContractPackageHash,
		Amount:                      big.NewInt(2),
		Token:                       common.HexToHash(tokenContractAddress),
		To:                          common.HexToHash(toAccountAddress),
		From: networks.Address{
			NetworkName: "GOERLI",
			Address:     "0x9032d7eb50b5b4a48c21035f34e0A84e54921D75",
		},
		TransactionID: big.NewInt(1),
	})
	require.NoError(t, err)
	require.Empty(t, txHash)
}

func TestCasperTransferOut(t *testing.T) {
	t.Skip("for manual testing")

	var (
		casperNodeAddress = "http://136.243.187.84:7777/rpc"

		privateKeySecp256k1ForSignature = "cc903a2179a5c47acef21d732c0693848c6c33e626fd6651b3773732bde6e127"
		// account-hash-daa2b596e0a496b04933e241e0567f2bcbecc829aa57d88cab096c28fd07dee2.
		privateKeyEd25519ForTransaction = "33e8798230a0ce481455b8b0affe5f96455fa157799095567ef220d55789a2e10ad302bfc22c0e606d94d98a3baa2c8eeedd1e148d9a20a4453bb8cc5e530a19"

		accountAddress            = "daa2b596e0a496b04933e241e0567f2bcbecc829aa57d88cab096c28fd07dee2"
		tokenContractAddress      = "3c0c1847d1c410338ab9b4ee0919c181cf26085997ff9c797e8a1ae5b02ddf23"
		bridgeContractPackageHash = "7225d70f3e197d78dd286a77bfa219e8573d62e6c4b62ade27ec14c69a88442f"
		recipientAddress          = "daa2b596e0a496b04933e241e0567f2bcbecc829aa57d88cab096c28fd07dee2"
	)

	ctx := context.Background()
	casperClient := client.New(casperNodeAddress)

	privateKeyECDSA, err := crypto.HexToECDSA(privateKeySecp256k1ForSignature)
	require.NoError(t, err)

	privateKeyForTransferSigningBytes, err := hex.DecodeString(privateKeyEd25519ForTransaction)
	require.NoError(t, err)

	publicKey := make([]byte, signer_service.PublicKeySize)
	copy(publicKey, privateKeyForTransferSigningBytes[signer_service.PublicKeySize:])
	pair := casper_ed25519.ParseKeyPair(publicKey, privateKeyForTransferSigningBytes[:signer_service.PublicKeySize])

	tokenContractAddressBytes, err := hex.DecodeString(tokenContractAddress)
	require.NoError(t, err)

	accountAddressBytes, err := hex.DecodeString(accountAddress)
	require.NoError(t, err)

	recipientAddressBytes, err := hex.DecodeString(recipientAddress)
	require.NoError(t, err)

	var signature []byte
	t.Run("get transferOut signature", func(t *testing.T) {
		signer := casper.NewSigner(func(b []byte, _ signer_service.Type) ([]byte, error) {
			signature, err := crypto.Sign(b, privateKeyECDSA)
			return signature, err
		})

		bridgeHash, err := hex.DecodeString(bridgeContractPackageHash)
		require.NoError(t, err)

		signature, err = signer.GetTransferOutSignature(ctx, casper_chain.TransferOutSignature{
			Prefix:           "TRICORN_TRANSFER_OUT",
			BridgeHash:       bridgeHash,
			TokenPackageHash: tokenContractAddressBytes,
			AccountAddress:   accountAddressBytes,
			Recipient:        recipientAddressBytes,
			Amount:           big.NewInt(2),
			GasCommission:    big.NewInt(1),
			Nonce:            big.NewInt(113),
		})
		require.NoError(t, err)
		require.NotEmpty(t, signature)
	})

	t.Run("transferOut", func(t *testing.T) {
		transfer := casper.NewTransfer(casperClient, func(b []byte) ([]byte, error) {
			casperSignature := pair.Sign(b)
			return casperSignature.SignatureData, nil
		})

		txHash, err := transfer.TransferOut(ctx, casper.TransferOutRequest{
			PublicKey:                   pair.PublicKey(),
			ChainName:                   "CASPER-TEST",
			StandardPaymentForBridgeOut: 40000000000, // 40 CSPR.
			BridgeContractPackageHash:   bridgeContractPackageHash,
			TokenContractAddress:        tokenContractAddressBytes,
			Amount:                      big.NewInt(2),
			GasCommission:               big.NewInt(1),
			Nonce:                       big.NewInt(113),
			Recipient:                   recipientAddressBytes,
			Signature:                   signature,
		})
		require.NoError(t, err)
		require.NotEmpty(t, txHash)
	})
}
