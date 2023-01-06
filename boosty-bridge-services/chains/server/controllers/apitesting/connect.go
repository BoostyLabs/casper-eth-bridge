package apitesting

import (
	bridgeconnectorpb "github.com/BoostyLabs/casper-eth-bridge/boosty-communication/go-gen/bridge-connector"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// ConnectToConnector initiates connection with connector server.
func ConnectToConnector(addr string) (bridgeconnectorpb.ConnectorClient, error) {
	conn, err := grpc.Dial(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	return bridgeconnectorpb.NewConnectorClient(conn), nil
}
