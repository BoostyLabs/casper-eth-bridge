// Copyright (C) 2022 Creditor Corp. Group.
// See LICENSE for copying information.

package controllers_test

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"github.com/caarlos0/env/v6"
	"testing"

	"github.com/casper-ecosystem/casper-golang-sdk/sdk"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"

	signer "github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/cmd/signer/db"
	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/networks"
	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/pkg/envparse"
	signer_service "github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/signer"
	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/signer/server/controllers/apitesting"
	pb_networks "github.com/BoostyLabs/casper-eth-bridge/boosty-communication/go-gen/networks"
	pb_signer "github.com/BoostyLabs/casper-eth-bridge/boosty-communication/go-gen/signer"
)

func TestSigner(t *testing.T) {
	dataTxHex := "078d3b790000000000000000000000005847422de06b9de9750788ef7a089293cde896a5000000000000000000000000f8e4121744b0744de8624da495449594582344d00000000000000000000000000000000000000000000000000000000000000001"

	signatureEVM := "c165c06e5ba9c5e127fc304a79444066480c4ed3a52691cfb057660291593ac6539aafe31c84f705505d3f819fe28a69350f786712bb9f5b6cfef43eae11274600"
	signatureCasper := "58784687434a1e68268b6a01b1a7f8d3a2adf395691ee1eebab802ff83065328c82b5ba34594509563e92f67ec209d9765ce75d2f9be0e53ed50c7502c1fb20d"

	publicKeyEVM := "2807d9de22f235ccc562969628cda551d437afb799d3f8f3baaccbe8ea9379f6344d890efb763c07d503d89d05a1669d15c44e396110a18fde1fc435afb3ad82"
	publicKeyCasper := "d90cdb7e06d2f2e6a5a1e1999f4e0447003a941c6a039b7749e24a85052863ea"

	privateKeyEVM := signer_service.PrivateKey{
		NetworkType: networks.TypeEVM,
		Key:         "3b2c3b9eec999beb061fd5b9fc60ae7995e9a81504e4f1c0e852ffc532cd0649",
	}
	privateKeyCasper := signer_service.PrivateKey{
		NetworkType: networks.TypeCasper,
		Key:         "b0fb23ba3c5a3e327a5624a7a25aa612b4c799e0981a11123d98109d94974020d90cdb7e06d2f2e6a5a1e1999f4e0447003a941c6a039b7749e24a85052863ea",
	}

	err := godotenv.Overload("./apitesting/configs/.test.signer.env")
	if err != nil {
		t.Fatalf("could not load signer testing file: %v", err)
	}

	config := new(apitesting.SignerConfig)
	envOpt := env.Options{RequiredIfNoDef: true}
	err = env.ParseWithFuncs(config, envparse.EthParseOpts(), envOpt)
	if err != nil {
		t.Fatalf("could not parse config: %v", err)
	}

	apitesting.SignerRun(t, func(ctx context.Context, t *testing.T, db signer.DB) {
		repository := db.PrivateKeys()

		signerClient, err := apitesting.ConnectToSigner(config.GrpcServerAddress)
		require.NoError(t, err, "can't create signer service client")

		t.Run("Seed", func(t *testing.T) {
			err := repository.Create(ctx, privateKeyEVM)
			require.NoError(t, err)

			err = repository.Create(ctx, privateKeyCasper)
			require.NoError(t, err)
		})

		t.Run("Sign EVM", func(t *testing.T) {
			dataTx, err := hex.DecodeString(dataTxHex)
			require.NoError(t, err)

			dataHash := crypto.Keccak256Hash(dataTx)

			signatureResponse, err := signerClient.Sign(context.Background(), &pb_signer.SignRequest{
				NetworkId: pb_networks.NetworkType_NT_EVM,
				Data:      dataHash.Bytes(),
			})
			require.NoError(t, err)
			require.Equal(t, signatureEVM, hex.EncodeToString(signatureResponse.Signature))

		})

		t.Run("Sign Casper", func(t *testing.T) {
			deploy := new(sdk.Deploy)
			dataBytes, err := json.Marshal(deploy)
			require.NoError(t, err)

			signatureResponse, err := signerClient.Sign(context.Background(), &pb_signer.SignRequest{
				NetworkId: pb_networks.NetworkType_NT_CASPER,
				Data:      dataBytes,
			})
			require.NoError(t, err)
			require.Equal(t, signatureCasper, hex.EncodeToString(signatureResponse.Signature))
		})

		t.Run("PublicKey EVM", func(t *testing.T) {
			publicKeyResponse, err := signerClient.PublicKey(context.Background(), &pb_signer.PublicKeyRequest{
				NetworkId: pb_networks.NetworkType_NT_EVM,
			})
			require.NoError(t, err)
			require.Equal(t, publicKeyEVM, hex.EncodeToString(publicKeyResponse.PublicKey))
		})

		t.Run("PublicKey Casper", func(t *testing.T) {
			publicKeyResponse, err := signerClient.PublicKey(context.Background(), &pb_signer.PublicKeyRequest{
				NetworkId: pb_networks.NetworkType_NT_CASPER,
			})
			require.NoError(t, err)
			require.Equal(t, publicKeyCasper, hex.EncodeToString(publicKeyResponse.PublicKey))
		})
	})
}
