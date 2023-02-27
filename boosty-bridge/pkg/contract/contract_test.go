// Copyright (C) 2022 Creditor Corp. Group.
// See LICENSE for copying information.

package contract_test

// for manual testing.
// func TestBridgeIn(t *testing.T) {
// 	ctx := context.Background()

// 	deploy := `{
// 		"deploy": {
// 			"hash": "b36664d7738ba0212fc4edb304bbeed36918a812e2a672a7a2491a16b0d2a9bb",
// 			"header": {
// 				"account": "010ad302bfc22c0e606d94d98a3baa2c8eeedd1e148d9a20a4453bb8cc5e530a19",
// 				"timestamp": "2022-10-31T21:19:48.718Z",
// 				"ttl": "30m0s",
// 				"gas_price": 1,
// 				"body_hash": "1d3d713c96d5d2d7425ecdc1fff3897bf02f23ec63aef98835520e947c344685",
// 				"dependencies": [],
// 				"chain_name": "casper-test"
// 			},
// 			"payment": {
// 				"ModuleBytes": {
// 					"args": [
// 						[
// 							"amount",
// 							{
// 								"bytes": "040084d717",
// 								"cl_type": "U512"
// 							}
// 						]
// 					],
// 					"module_bytes": ""
// 				}
// 			},
// 			"session": {
// 				"StoredContractByHash": {
// 					"args": [
// 						[
// 							"token_contract",
// 							{
// 								"bytes": "3c0c1847d1c410338ab9b4ee0919c181cf26085997ff9c797e8a1ae5b02ddf23",
// 								"cl_type": {
// 									"ByteArray": 32
// 								}
// 							}
// 						],
// 						[
// 							"amount",
// 							{
// 								"bytes": "0101",
// 								"cl_type": "U256"
// 							}
// 						],
// 						[
// 							"destination_chain",
// 							{
// 								"bytes": "06000000474f45524c49",
// 								"cl_type": "String"
// 							}
// 						],
// 						[
// 							"destination_address",
// 							{
// 								"bytes": "2a000000307833303935663935356461373030623936323135636666633962633634616232653639656237646162",
// 								"cl_type": "String"
// 							}
// 						]
// 					],
// 					"entry_point": "bridge_in",
// 					"hash": "8153c553e8339fb87224097e4a3a2d8e4d8f49fbacee5c192e039709bc4211ba"
// 				}
// 			},
// 			"approvals": [
// 				{
// 					"signer": "010ad302bfc22c0e606d94d98a3baa2c8eeedd1e148d9a20a4453bb8cc5e530a19",
// 					"signature": "01f8a94f9745a32557f696dcd0e0943621776fca786c04d2a53e002dd3b3aa565dfe366555b41bf81fb5454685057dd1dd4480a02309ac53a946444017c5d68806"
// 				}
// 			]
// 		}
// 	}`

// 	in := contract.BridgeInRequest{
// 		Deploy:         deploy,
// 		RpcNodeAddress: "http://136.243.187.84:7777/rpc",
// 	}
// 	t.Run("BridgeIn", func(t *testing.T) {
// 		_, err := contract.BridgeIn(ctx, in)
// 		require.NoError(t, err)
// 	})
// }.

// func TestBridgeInWithoutSignature(t *testing.T) {
// 	ctx := context.Background()

// 	inWithoutSignature := contract.BridgeInRequestWithoutSignature{
// 		Amount: big.NewInt(1),
// 		Token:  common.HexToHash("013c0c1847d1c410338ab9b4ee0919c181cf26085997ff9c797e8a1ae5b02ddf23"),
// 		From: contract.StringNetworkAddress{
// 			NetworkName: "GOERLI",
// 			Address:     "0x3095f955da700b96215cffc9bc64ab2e69eb7dab",
// 		},

// 		ChainName:                   "CASPER-TEST",
// 		StandardPaymentForBridgeOut: big.NewInt(400000000), // 0.4 casp.
// 		BridgeContractPackageHash:   "8153c553e8339fb87224097e4a3a2d8e4d8f49fbacee5c192e039709bc4211ba",
// 		RpcNodeAddress:              "http://136.243.187.84:7777/rpc",
// 		CommunicationConfig: rpc.Config{
// 			BridgeAddress:     "localhost:10006",
// 			PingServerTime:    time.Second * 10,
// 			PingServerTimeout: time.Second * 1,
// 		},
// 	}

// 	t.Run("BridgeInWithoutSignature", func(t *testing.T) {
// 		_, err := contract.BridgeInWithoutSignature(ctx, inWithoutSignature)
// 		require.NoError(t, err)
// 	})
// }.
