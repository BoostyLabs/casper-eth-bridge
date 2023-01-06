// Copyright (C) 2022 Creditor Corp. Group.
// See LICENSE for copying information.

package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/casper-ecosystem/casper-golang-sdk/sdk"
	"github.com/pkg/errors"
	"github.com/zeebo/errs"

	casper "github.com/BoostyLabs/casper-eth-bridge/boosty-bridge-services/chains/casper"
)

// ensures that rpcClient implement casper.Casper.
var _ casper.Casper = (*rpcClient)(nil)

// rpcClient is a implementation of connector_service.Casper.
type rpcClient struct {
	client *sdk.RpcClient

	rpcNodeAddress string
}

// New is constructor for rpcClient.
func New(rpcNodeAddress string) casper.Casper {
	client := sdk.NewRpcClient(rpcNodeAddress)
	return &rpcClient{
		client:         client,
		rpcNodeAddress: rpcNodeAddress,
	}
}

// PutDeploy deploys a contract or sends a transaction and returns deployment hash.
func (r *rpcClient) PutDeploy(deploy sdk.Deploy) (string, error) {
	deployResp, err := r.client.PutDeploy(deploy)
	return deployResp.Hash, err
}

// GetBlockNumberByHash returns block number by deploy hash.
func (r *rpcClient) GetBlockNumberByHash(hash string) (int, error) {
	blockResp, err := r.client.GetBlockByHash(hash)
	return blockResp.Header.Height, err
}

// GetEventsByBlockNumbers returns events for range of block numbers.
func (r *rpcClient) GetEventsByBlockNumbers(fromBlockNumber uint64, toBlockNumber uint64, bridgeInEventHash string,
	bridgeOutEventHash string) ([]casper.Event, error) {
	events := make([]casper.Event, 0)

	for blockNumber := fromBlockNumber; blockNumber <= toBlockNumber; blockNumber++ {
		blockResp, err := r.client.GetBlockByHeight(blockNumber)
		if err != nil {
			return nil, err
		}

		for _, hash := range blockResp.Body.DeployHashes {
			deploy, err := r.getDeploy(hash)
			if err != nil {
				return nil, err
			}

			for i, executionResult := range deploy.ExecutionResults {
				for _, transform := range executionResult.Result.Success.Effect.Transforms {
					if transform.Key != bridgeInEventHash && transform.Key != bridgeOutEventHash {
						continue
					}

					event := casper.Event{
						DeployProcessed: casper.DeployProcessed{
							DeployHash: deploy.Deploy.Hash,
							Account:    deploy.Deploy.Header.Account,
							BlockHash:  deploy.ExecutionResults[i].BlockHash,
							ExecutionResult: casper.ExecutionResult{
								Success: casper.Success{
									Effect: casper.Effect{
										Transforms: []casper.Transform{
											{
												Key:       transform.Key,
												Transform: transform.Transform,
											},
										},
									},
								},
							},
						},
					}

					events = append(events, event)
				}
			}
		}
	}

	return events, nil
}

// GetCurrentBlockNumber returns current block number.
func (r *rpcClient) GetCurrentBlockNumber() (uint64, error) {
	blockResp, err := r.client.GetLatestBlock()
	if err != nil {
		return 0, err
	}

	return uint64(blockResp.Header.Height), nil
}

func (r *rpcClient) getDeploy(hash string) (DeployResult, error) {
	var result DeployResult

	resp, err := r.rpcCall("info_get_deploy", map[string]string{
		"deploy_hash": hash,
	})
	if err != nil {
		return DeployResult{}, err
	}

	err = json.Unmarshal(resp.Result, &result)

	return result, err
}

func (r *rpcClient) rpcCall(method string, params interface{}) (_ sdk.RpcResponse, err error) {
	var rpcResponse sdk.RpcResponse

	body, err := json.Marshal(sdk.RpcRequest{
		Version: "2.0",
		Method:  method,
		Params:  params,
	})
	if err != nil {
		return sdk.RpcResponse{}, errors.Wrap(err, "failed to marshal json")
	}

	resp, err := http.Post(r.rpcNodeAddress, "application/json", bytes.NewReader(body))
	if err != nil {
		return sdk.RpcResponse{}, fmt.Errorf("failed to make request: %v", err)
	}

	defer func() {
		err = errs.Combine(err, resp.Body.Close())
	}()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return sdk.RpcResponse{}, fmt.Errorf("failed to get response body: %v", err)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return sdk.RpcResponse{}, fmt.Errorf("request failed, status code - %d, response - %s", resp.StatusCode, string(b))
	}

	err = json.Unmarshal(b, &rpcResponse)
	if err != nil {
		return sdk.RpcResponse{}, fmt.Errorf("failed to parse response body: %v", err)
	}

	if rpcResponse.Error != nil {
		return rpcResponse, fmt.Errorf("rpc call failed, code - %d, message - %s", rpcResponse.Error.Code, rpcResponse.Error.Message)
	}

	return rpcResponse, nil
}

// DeployResult and nested structures describes deploy structure in casper network.
type (
	DeployResult struct {
		Deploy           JsonDeploy            `json:"deploy"`
		ExecutionResults []JsonExecutionResult `json:"execution_results"`
	}

	JsonDeploy struct {
		Hash      string           `json:"hash"`
		Header    JsonDeployHeader `json:"header"`
		Approvals []JsonApproval   `json:"approvals"`
	}

	JsonDeployHeader struct {
		Account      string    `json:"account"`
		Timestamp    time.Time `json:"timestamp"`
		TTL          string    `json:"ttl"`
		GasPrice     int       `json:"gas_price"`
		BodyHash     string    `json:"body_hash"`
		Dependencies []string  `json:"dependencies"`
		ChainName    string    `json:"chain_name"`
	}

	JsonApproval struct {
		Signer    string `json:"signer"`
		Signature string `json:"signature"`
	}

	JsonExecutionResult struct {
		BlockHash string          `json:"block_hash"`
		Result    ExecutionResult `json:"result"`
	}

	ExecutionResult struct {
		Success      SuccessExecutionResult `json:"success"`
		ErrorMessage *string                `json:"error_message,omitempty"`
	}

	SuccessExecutionResult struct {
		Transfers []string `json:"transfers"`
		Effect    Effect   `json:"effect"`
		Cost      string   `json:"cost"`
	}

	Effect struct {
		Transforms []Transform `json:"transforms"`
	}

	Transform struct {
		Key       string      `json:"key"`
		Transform interface{} `json:"transform"`
	}
)
