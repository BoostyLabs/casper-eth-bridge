package client_test

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"math/big"
	"testing"
	"time"

	"github.com/btcsuite/btcd/btcec"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"

	"tricorn/chains/evm"
	evm_client "tricorn/internal/contracts/evm/client"
	"tricorn/signer"
)

func TestClient(t *testing.T) {
	ctx := context.Background()

	expectedSignature := "1311fd12637b428ec0bd0a26099169e3136adb553d38dc42500eeaed19105df014b3f1d25fbb3299e3cc5ff122097c0cb7fd362ff49b5c70c7571a86bfd8704e1b"

	privateKey := "ecff8b9c717a56b30f35a75db85342a1b42fcfe8540a733c73cc9ef38a165a56" // 0x7e0f5A592322Bc973DDE62dF3f91604D21d37446.
	privateKeyECDSA, err := crypto.HexToECDSA(privateKey)
	require.NoError(t, err)

	var publicKey []byte
	publicKey = append(publicKey, privateKeyECDSA.PublicKey.X.Bytes()...)
	publicKey = append(publicKey, privateKeyECDSA.PublicKey.Y.Bytes()...)

	x := big.NewInt(0).SetBytes(publicKey[:32])
	y := big.NewInt(0).SetBytes(publicKey[32:])
	publicKeyECDSA := ecdsa.PublicKey{
		Curve: btcec.S256(),
		X:     x,
		Y:     y,
	}

	signerAddress := crypto.PubkeyToAddress(publicKeyECDSA)

	transfer, err := evm_client.NewClient(
		ctx,
		evm_client.Config{
			NodeAddress:           "https://goerli.infura.io/v3/b6b83346b6274c39ab08e2b2a0260235",
			BridgeContractAddress: common.HexToAddress("0xA0E532456654bcC83F584e0EDf6ba065f87f528F"),
		},
		signerAddress,
		func(data []byte, _ signer.Type) ([]byte, error) {

			signature, err := crypto.Sign(data, privateKeyECDSA)
			if err != nil {
				return nil, err
			}
			return signature, nil
		},
	)
	require.NoError(t, err)

	var signature []byte
	t.Run("get signature for bridgeIn", func(t *testing.T) {
		amount, ok := big.NewInt(0).SetString("1", 10)
		require.True(t, ok)

		gasCommission, ok := big.NewInt(0).SetString("1", 10)
		require.True(t, ok)

		sig, err := transfer.GetBridgeInSignature(ctx, evm.GetBridgeInSignatureRequest{
			User:               signerAddress,
			Token:              common.HexToAddress("0x9fF6D0788066982c95D26F4A74d6C700F3Dc29ec"),
			Amount:             amount,
			GasCommission:      gasCommission,
			DestinationChain:   "Solana",
			DestinationAddress: "4zXwdbUDWo1S5AP2CEfv4zAPRds5PQUG1dyqLLvib2xu",
			Deadline:           big.NewInt(1732924800), // big.NewInt(time.Date(2025, 0, 0, 0, 0, 0, 0, time.UTC).Unix()).
			Nonce:              big.NewInt(0),          // increase after a successful transaction.
		})
		require.NoError(t, err)
		require.NotEmpty(t, sig)
		require.Equal(t, expectedSignature, hex.EncodeToString(sig))

		signature = sig

	})

	// for manual testing.
	t.Skip("you need enough allowance")
	t.Run("bridge in", func(t *testing.T) {
		amount, ok := big.NewInt(0).SetString("1", 10)
		require.True(t, ok)

		gasCommission, ok := big.NewInt(0).SetString("1", 10)
		require.True(t, ok)

		_, err := transfer.BridgeIn(ctx, evm.BridgeInRequest{
			Token:              common.HexToAddress("0x9fF6D0788066982c95D26F4A74d6C700F3Dc29ec"),
			Amount:             amount,
			GasCommission:      gasCommission,
			DestinationChain:   "Solana",
			DestinationAddress: "4zXwdbUDWo1S5AP2CEfv4zAPRds5PQUG1dyqLLvib2xu",
			Deadline:           big.NewInt(time.Now().AddDate(0, 0, 1).Unix()),
			Nonce:              big.NewInt(0),
			Signature:          signature,
		})
		require.NoError(t, err)
	})

	t.Run("transfer out", func(t *testing.T) {
		amount, ok := big.NewInt(0).SetString("1", 10)
		require.True(t, ok)

		err := transfer.TransferOut(ctx, evm.TransferOutRequest{
			Token:     common.HexToAddress("0x9fF6D0788066982c95D26F4A74d6C700F3Dc29ec"),
			Recipient: signerAddress,
			Amount:    amount,
			Nonce:     big.NewInt(1),
		})
		require.NoError(t, err)
	})
}
