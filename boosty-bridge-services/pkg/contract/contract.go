// Copyright (C) 2022 Creditor Corp. Group.
// See LICENSE for copying information.

package contract

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"math/big"
	"strings"

	"github.com/casper-ecosystem/casper-golang-sdk/keypair"
	"github.com/casper-ecosystem/casper-golang-sdk/sdk"
	"github.com/casper-ecosystem/casper-golang-sdk/serialization"
	"github.com/casper-ecosystem/casper-golang-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/zeebo/errs"

	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/chains"
	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/communication/rpc"
	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/internal/eventparsing"
	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/networks"
	"github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/pkg/logger/zaplog"
)

// ErrContract indicates that there was an error in the contract package.
var ErrContract = errs.Class("contract package")

// BridgeInRequest describes values to initiate inbound bridge transaction.
type BridgeInRequest struct {
	Deploy         string
	RpcNodeAddress string
}

// StringNetworkAddress describes an address for some network.
type StringNetworkAddress struct {
	NetworkName string
	Address     string
}

// BridgeInResponse describes bridge in tx hash.
type BridgeInResponse struct {
	Txhash string
}

// BridgeIn initiates inbound bridge transaction.
func BridgeIn(ctx context.Context, req BridgeInRequest) (BridgeInResponse, error) {
	request := struct {
		Deploy struct {
			Hash      sdk.Hash                  `json:"hash"`
			Header    *sdk.DeployHeader         `json:"header"`
			Payment   *sdk.ExecutableDeployItem `json:"payment"`
			Session   *sdk.ExecutableDeployItem `json:"session"`
			Approvals []struct {
				Signer    string `json:"signer"`
				Signature string `json:"signature"`
			} `json:"approvals"`
		}
	}{}

	err := json.Unmarshal([]byte(req.Deploy), &request)
	if err != nil {
		return BridgeInResponse{}, ErrContract.Wrap(err)
	}

	pubKeyData, err := hex.DecodeString(request.Deploy.Approvals[0].Signer[eventparsing.LengthSelectorTag:])
	if err != nil {
		return BridgeInResponse{}, ErrContract.Wrap(err)
	}

	signer := keypair.PublicKey{
		Tag:        request.Deploy.Header.Account.Tag,
		PubKeyData: pubKeyData,
	}

	signatureData, err := hex.DecodeString(request.Deploy.Approvals[0].Signature[eventparsing.LengthSelectorTag:])
	if err != nil {
		return BridgeInResponse{}, ErrContract.Wrap(err)
	}

	signature := keypair.Signature{
		Tag:           request.Deploy.Header.Account.Tag,
		SignatureData: signatureData,
	}

	approval := sdk.Approval{
		Signer:    signer,
		Signature: signature,
	}

	deploy := sdk.Deploy{
		Hash:      request.Deploy.Hash,
		Header:    request.Deploy.Header,
		Payment:   request.Deploy.Payment,
		Session:   request.Deploy.Session,
		Approvals: []sdk.Approval{approval},
	}

	casperClient := sdk.NewRpcClient(req.RpcNodeAddress)
	deployResp, err := casperClient.PutDeploy(deploy)
	if err != nil {
		return BridgeInResponse{}, ErrContract.Wrap(err)
	}

	resp := BridgeInResponse{
		Txhash: deployResp.Hash,
	}

	return resp, nil
}

// BridgeInRequestWithoutSignature describes values to initiate inbound bridge transaction.
type BridgeInRequestWithoutSignature struct {
	Amount                      *big.Int
	Token                       common.Hash
	From                        StringNetworkAddress
	ChainName                   string
	StandardPaymentForBridgeOut *big.Int
	BridgeContractPackageHash   string
	RpcNodeAddress              string
	CommunicationConfig         rpc.Config
}

// BridgeInResponseWithoutSignature describes bridge in tx hash.
type BridgeInResponseWithoutSignature struct {
	Txhash []byte
}

// BridgeInWithoutSignature initiates inbound bridge transaction.
func BridgeInWithoutSignature(ctx context.Context, req BridgeInRequestWithoutSignature) (BridgeInResponseWithoutSignature, error) {
	var resp BridgeInResponseWithoutSignature

	log := zaplog.NewLog()

	bridge, err := rpc.New(req.CommunicationConfig, log)
	if err != nil {
		return resp, ErrContract.Wrap(err)
	}

	respPubKey, err := bridge.Bridge().PublicKey(ctx, networks.TypeCasper)
	if err != nil {
		return resp, ErrContract.Wrap(err)
	}

	publicKey := keypair.PublicKey{
		Tag:        keypair.KeyTagEd25519,
		PubKeyData: respPubKey,
	}

	deployParams := sdk.NewDeployParams(publicKey, strings.ToLower(req.ChainName), nil, 0)
	payment := sdk.StandardPayment(req.StandardPaymentForBridgeOut)

	// token contract.
	tokenContractFixedBytes := types.FixedByteArray(req.Token.Bytes())
	tokenContract := types.CLValue{
		Type:      types.CLTypeByteArray,
		ByteArray: &tokenContractFixedBytes,
	}
	tokenContractBytes, err := serialization.Marshal(tokenContract)
	if err != nil {
		return resp, ErrContract.Wrap(err)
	}

	// amount.
	amount := types.CLValue{
		Type: types.CLTypeU256,
		U256: req.Amount,
	}
	amountBytes, err := serialization.Marshal(amount)
	if err != nil {
		return resp, ErrContract.Wrap(err)
	}

	//  destination chain.
	destinationChain := types.CLValue{
		Type:   types.CLTypeString,
		String: &req.From.NetworkName,
	}
	destinationChainBytes, err := serialization.Marshal(destinationChain)
	if err != nil {
		return resp, ErrContract.Wrap(err)
	}

	// destination address.
	destinationAddress := types.CLValue{
		Type:   types.CLTypeString,
		String: &req.From.Address,
	}
	destinationAddressBytes, err := serialization.Marshal(destinationAddress)
	if err != nil {
		return resp, ErrContract.Wrap(err)
	}

	args := map[string]sdk.Value{
		"token_contract": {
			IsOptional:  false,
			Tag:         types.CLTypeByteArray,
			StringBytes: hex.EncodeToString(tokenContractBytes),
		},
		"amount": {
			Tag:         types.CLTypeU256,
			IsOptional:  false,
			StringBytes: hex.EncodeToString(amountBytes),
		},
		"destination_chain": {
			Tag:         types.CLTypeString,
			IsOptional:  false,
			StringBytes: hex.EncodeToString(destinationChainBytes),
		},
		"destination_address": {
			Tag:         types.CLTypeString,
			IsOptional:  false,
			StringBytes: hex.EncodeToString(destinationAddressBytes),
		},
	}

	keyOrder := []string{
		"token_contract",
		"amount",
		"destination_chain",
		"destination_address",
	}
	runtimeArgs := sdk.NewRunTimeArgs(args, keyOrder)

	contractHexBytes, err := hex.DecodeString(req.BridgeContractPackageHash)
	if err != nil {
		return resp, ErrContract.Wrap(err)
	}

	var contractHashBytes [32]byte
	copy(contractHashBytes[:], contractHexBytes)
	session := sdk.NewStoredContractByHash(contractHashBytes, "bridge_in", *runtimeArgs)

	deploy := sdk.MakeDeploy(deployParams, payment, session)

	data, err := json.Marshal(*deploy)
	if err != nil {
		return resp, ErrContract.Wrap(err)
	}

	reqSign := chains.SignRequest{
		NetworkId: networks.TypeCasper,
		Data:      data,
	}

	signature, err := bridge.Bridge().Sign(ctx, reqSign)
	if err != nil {
		return resp, ErrContract.Wrap(err)
	}

	signatureKeypair := keypair.Signature{
		Tag:           keypair.KeyTagEd25519,
		SignatureData: signature,
	}

	approval := sdk.Approval{
		Signer:    publicKey,
		Signature: signatureKeypair,
	}

	deploy.Approvals = append(deploy.Approvals, approval)

	casperClient := sdk.NewRpcClient(req.RpcNodeAddress)
	deployResp, err := casperClient.PutDeploy(*deploy)
	if err != nil {
		return resp, ErrContract.Wrap(err)
	}

	resp.Txhash, err = hex.DecodeString(deployResp.Hash)

	return resp, err
}
