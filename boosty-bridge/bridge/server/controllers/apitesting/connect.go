// Copyright (C) 2022 Creditor Corp. Group.
// See LICENSE for copying information.

package apitesting

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	connectorbridgepb "github.com/BoostyLabs/casper-eth-bridge/boosty-communication/go-gen/connector-bridge"
	gatewaybridgepb "github.com/BoostyLabs/casper-eth-bridge/boosty-communication/go-gen/gateway-bridge"
)

// ConnectToGateway initiates connection with gateway server.
func ConnectToGateway(addr string) (gatewaybridgepb.GatewayBridgeClient, error) {
	conn, err := grpc.Dial(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	return gatewaybridgepb.NewGatewayBridgeClient(conn), nil
}

// ConnectToBridge initiates connection with bridge server.
func ConnectToBridge(addr string) (connectorbridgepb.ConnectorBridgeClient, error) {
	conn, err := grpc.Dial(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	return connectorbridgepb.NewConnectorBridgeClient(conn), nil
}
