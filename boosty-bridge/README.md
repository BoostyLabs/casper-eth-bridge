# Bridge Core

This repository contains the "Core" component of Boosty Bridge, which serves as an orchestrator for the bridging process.

The primary crate here is located in `poc/bridge-core`, which is the binary that sets up a server that connects to "Connectors" and provides a gRPC interface for querying.

## Bridge Architecture

The Core tracks the state of all connected networks and all current and completed transactions. It is responsible for processing network events, which it receives from the Connectors, and responds to these events by instructing corresponding Connectors to issue a transaction on the destination chain.

A "Connector" in this context is a stateless gRPC service that provides a unified, blockchain-agnostic interface to a particular network. This allows for Connectors to be written in whatever language is most convenient, and avoids dependency issues in the Core itself. This also decouples the Core process from Connector processes, meaning that if a Connector goes down, the Core continues to function and provide services on other networks, while the downed Connector recovers.

The gRPC API provided by Connectors is described here: https://github.com/BoostyLabs/golden-gate-communication/blob/master/proto/bridge-connector/bridge-connector.proto

The Connector exposes to the Core some basic metadata about the network it's connected to, including a unique network id (NID), a network type (currently, Casper or EVM) which dictates the address and signature formats, as well as the network name. The Core is provided with a list of Connectors to use upon startup, and it will internally register each network and then call the `EventStream` method to begin processing events from the network.

The Connector will monitor its associated network for smart contract events, convert them into the format expected by the Core, and report them back. It is expected that each event reported by a Connector has reached a high level of finality, either absolute finality or with a high enough probabililty that a rollback is unlikely. The Core trusts the Connectors to correctly monitor and report consensus state. Ideally, each Connector should use a private full node or light client to verify that proper consensus has been reached before reporting an event back to the Core.

Upon receiving the event, which will usually be a `FundsIn` for inbound transfers, the Core will parse it to determine which destination network to route it to. If the event is well-formed and passes validation checks, the Core will then find the Connector for the destination chain of the transfer, and call the `BridgeOut` method to issue a transaction. A `FundsOut` event is expected to be issued on the destination network when the transaction has been finalized, which signals to the Core that the transfer should be marked as completed.

The Core also exposes a gRPC service to query state, and perform some authorized actions on the behalf of a user, such as cancelling a transfer before it has been issued on the destination chain. This gRPC service is accessed by the frontend via a REST API gateway. The endpoints exposed by the Core are described here: https://github.com/BoostyLabs/golden-gate-communication/blob/master/proto/gateway-bridge/gateway-bridge.proto

The Connectors also provide some other auxiliary methods necessary for full operation of the bridge, such as gas cost estimation and others.

Signing transactions is handled via a separate Signer component. It is expected that the Signer service is shielded from external networks, and only accessible to the Core via an internal VPN. The Connectors can then sign transactions using the Core, which proxies the requests to the Signer.

For storage, a PostgreSQL database is used. It tracks the state of current transfers, as well as token metadata for each connected network and some other miscellaneous data. The schema for the database can be found in `poc/bridge-core/src/sql/create_tables.sql`

## Launching

For a simple setup:

```
cd poc/bridge-core
cargo run --bin bridge -- --embed-db --connectors http://127.0.0.1:10001 --connectors http://127.0.0.1:10002
```

Connectors to networks are expected to be running at local ports `10001` and `10002`. `--embed-db` spins up a simple local PostgreSQL node by itself, this is only for development purposes. The configuration for the database can be found in `poc/bridge-core/src/bin/bridge.rs`, as well as the predefined network metadata.

