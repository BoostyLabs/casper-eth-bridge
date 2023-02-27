package apitesting

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	bridgeconnectorpb "github.com/BoostyLabs/golden-gate-communication/go-gen/bridge-connector"
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
