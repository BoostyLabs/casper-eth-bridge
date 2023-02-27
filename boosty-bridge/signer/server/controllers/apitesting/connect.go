package apitesting

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	bridgesignerpb "github.com/BoostyLabs/golden-gate-communication/go-gen/bridge-signer"
)

// ConnectToSigner initiates connection with signer server.
func ConnectToSigner(addr string) (bridgesignerpb.BridgeSignerClient, error) {
	conn, err := grpc.Dial(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	return bridgesignerpb.NewBridgeSignerClient(conn), nil
}
