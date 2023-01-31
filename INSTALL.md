#### **Installing**

*This guide expects that you have cloned repository. All pathes are relative to it*

#### **Golang**

Golang is our backend language.

We are using version 1.18 

You can download it from the official website [GoLang](https://go.dev/dl/), and install it according to the official [instructions](https://go.dev/doc/install).

#### **Rust**

```
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
rustup default nightly
rustup target add wasm32-unknown-unknown
cargo install just
sudo apt install -y protobuf-compiler
```

Also, you have to install https://github.com/WebAssembly/binaryen utils for wasm contract optimizations.

#### **Database**

For our project we use a relational database PostgreSQL, which you can download by following the link from the official [website](https://www.postgresql.org/download/) or you can run your database in a Docker container.

#### **Docker**

For isolated installation of databases and servers we need a Docker,  version 20.10.21 or higher, you can download it at official [website](https://docs.docker.com/engine/install/).

##### Setup database in docker

 `docker run --name=db -e POSTGRES_PASSWORD=‘$YOUR_PASSWORD’ -p $YOUR_EXTERNAL_PORT:$YOUR_INTERNAL_PORT -d --rm $YOUR_USERNAME` - create container with postgres.

 `docker exec -it db createdb -U postgres $YOUR_DB_NAME` - create db in container.

###### Change creds in env configs:
**bridge**:
```
PG_PORT=$YOUR_EXTERNAL_PORT
PG_HOST=localhost
PG_USER=$YOUR_USERNAME
PG_DBNAME=$YOUR_DB_NAME
PG_PASS=$YOUR_PASSWORD
```

**signer**:
```
postgres://$YOUR_USERNAME:$YOUR_PASSWORD@localhost:$YOUR_EXTERNAL_PORT/$YOUR_DB_NAME?sslmode=disable
```

###### For example:

`docker run --name=db -e POSTGRES_PASSWORD='1212' -p 6432:5432 -d --rm postgres`

`docker exec -it db createdb -U postgres boosty_bridge_db`

**bridge**:
```
PG_PORT=6432
PG_HOST=localhost
PG_USER=postgres
PG_DBNAME=boosty_bridge_db
PG_PASS=1212
```

**signer**:
```
postgres://postgres:1212@localhost:6432/boosty_bridge_db?sslmode=disable
```

### Smart contract deployment

You have to generate private keys for Casper and Ethereum.
We use secp256k1 for Ethereum and ed25519 key scheme for Casper network.

##### Ethereum

You have to deploy smart contract in the `boosty-smart-contracts/ethereum/contracts` directory. Also, you have to deploy test erc20.

##### Casper
```
cd boosty-smart-contracts/casper/contract-bridge
just build-contract-release
```

You have to deploy boosty-smart-contracts/casper/contract-bridge/bridge-contract.wasm to the casper network.

Also, you have to deploy test erc20 token. https://github.com/casper-ecosystem/erc20

#### Configs
You have to create the following configs in the boosty-bridge-services/configs directory.

Copy `${ROOT_PROJECT_FOLDER}/boosty-bridge-services/configs/env_examples/*` to `${ROOT_PROJECT_FOLDER}/boosty-bridge-services/configs/` 
and remove `.example` prefix. Fill the configs.

##### Setup Infura 
Please, create your own infura and store into below configs. Also, we require it to put into 
* boosty-bridge-services/chains/server/controllers/apitesting/configs/.test.eth.env for running tests
* boosty-bridge-services/internal/contracts/evm/client/client_test.go: 46 line for running tests
* boosty-smart-contracts/ethereum/.env.example for deployment 

.env
```
BRIDGE_ADDRESS=localhost:10003

PING_SERVER_TIME=10s
PING_SERVER_TIMEOUT=1s
COMMUNICATION_MODE=GRPC
```

.casper.env
```
export BRIDGE_CONTRACT_PACKAGE_HASH=YOUR PACKAGE HASH
export BRIDGE_IN_EVENT_HASH=YOUR CONTRACT EVENT_TRIGGER NAMED KEY UREF
export BRIDGE_OUT_EVENT_HASH=YOUR CONTRACT EVENT_TRIGGER NAMED KEY UREF
export RPC_NODE_ADDRESS=http://136.243.187.84:7777/rpc # You can replace with your node
export EVENT_NODE_ADDRESS=http://136.243.187.84:9999/events/main # You can replace with your node

export GRPC_SERVER_ADDRESS=localhost:10004
export CHAIN_NAME=CASPER-TEST
export STANDARD_PAYMENT_FOR_BRIDGE_OUT=10000000000 # 10 casp
export IS_TESTNET=true
export FEE=10 # 10 casp
export FEE_PERCENTAGE=0.4 # 0,4%
export ESTIMATED_CONFIRMATION=600 # 10 min
export SERVER_NAME=casper-connector
```

.eth.env
```
BRIDGE_CONTRACT_ADDRESS=YOUR CONTRACT ADDRESS
NODE_ADDRESS=YOUR INFURA RPC ADDRESS
WS_NODE_ADDRESS=YOUR INFURA WS ADDRESS

GRPC_SERVER_ADDRESS=localhost:10005
FUND_IN_EVENT_HASH=0x702f486f4d225a476770977c5edd22b4cb48bcd23d5b8969ce41f43a61bfc55f
FUND_OUT_EVENT_HASH=0x22558aff1ae914331914dc5357b1aa7c20bbfae260d698bb745092bca82d0655
CHAIN_ID=5
CHAIN_NAME=GOERLI
IS_TESTNET=true
BRIDGE_OUT_METHOD_NAME=bridgeOut
GAS_INCREASING_COEFFICIENT=1.1
CONFIRMATION_TIME=60 # 1min
FEE_PERCENTAGE=0.4
GAS_LIMIT=80000
NUM_OF_SUBSCRIBERS=5
SERVER_NAME=eth-connector
SIGNATURE_VALIDITY_TIME=86400 # 1d
```

.gateway.env
```
GATEWAY_ADDRESS=127.0.0.1:8088
WEP_APP_ADDRESS=http://localhost:8089
BRIDGE_ADDRESS=127.0.0.1:10003
PING_SERVER_TIME=10s
PING_SERVER_TIMEOUT=1s
COMMUNICATION_MODE=GRPC
SERVER_NAME=gateway
```

.signer.env
```
DATABASE="postgresql://postgres:123456@localhost:5432/postgres?sslmode=disable"
GRPC_SERVER_ADDRESS=localhost:10006
CHAIN_ID=5
SERVER_NAME=signer
```

.web.env
```
export STATIC_DIR=FULL PATH TO boosty-bridge-services/web/console
export CASPER_BRIDGE_CONTRACT=YOUR CASPER BRIDGE ADDRESS
export CASPER_TOKEN_CONTRACT=YOUR CASPER TOKEN ADDRESS
export ETH_BRIDGE_CONTRACT=YOUR ETH BRIDGE ADDRESS
export ETH_TOKEN_CONTRACT=YOUR ETH TOKEN ADDRESS
export CASPER_NODE_ADDRESS=http://136.243.187.84:7777/rpc ### You can replace with your own if you need

export ADDRESS=localhost:8089
export GATEWAY_ADDRESS=http://localhost:8088
export ETH_GAS_LIMIT=115000
export SERVER_NAME=web-app
```

Also you need to add config file `.env` to `${PATH_TO_PROJECT_FOLDER}/boosty-bridge/poc/bridge-core` with fields:
```
RUST_BACKTRACE=1
PG_PORT=
PG_HOST=
PG_USER=
PG_DBNAME=
PG_PASS=
```

#### Bridge

Before bridge running, you need to execute the command:

`export $(grep -v '^#' .env | xargs -d '\n')`

to write env values in your OS system.
Then execute this commands to build project, it can take 5 min.

```
cd boosty-bridge/poc/bridge-core
cargo build --release
```

**Important**:
If you redeployed contracts, please, update boosty-bridge/poc/bridge-core/src/bin/bridge.rs:170 EVM and CASPER token addresses to allowlist your TOKEN contracts.

PLEASE NOTE, THAT THE BRIDGE SUPPOSED TO START BEFORE CONNECTORS AND GATEWAY.

* --conectors - specify connector address
* --signer-addr - specify signer address
* --addr - specify bridge host address
* --init-tables -  should be called with this param only first time.

Server running for example:
`RUST_LOG=DEBUG ./target/release/bridge --connectors http://127.0.0.1:10004 --connectors http://127.0.0.1:10005 --addr 127.0.0.1:10003 --signer-addr http://127.0.0.1:10006 --init-table`

`RUST_LOG=DEBUG` - optional, for debuging.
`--connectors http://127.0.0.1:10004` - by analogy, you can add many connectors.
`--addr 127.0.0.1:10003` - address where bridge runs.
`--signer-addr http://127.0.0.1:10006` - address to signer.
`--init-table` - for tables initialization, execute first time with this flag. 

After running server sets basic tables.

On the contracts side we use nonces to prevent identical transactions from being sent.
You need to be sure that nonce which inserted to the database is unused by bridge contract because transactions will be reverted.
If there were already transactions before the bridge contract, then you need to write/edit records with nonces for a specific network in the database(in network_nonces table) using this commands(where 1st value - our internal network id, 2nd value - nonce):
```
insert into network_nonces values(4, 8000);
insert into network_nonces values(5, 8000);
insert into network_nonces values(7, 8000);
insert into network_nonces values(9, 8000);
insert into network_nonces values(11, 8000);
```
Please, specify large nonce to prevent this error.

#### Signer

Firstly, you need to run server:
```
cd boosty-bridge-services
go run cmd/signer/main.go run
```

Server inits table for private keys, then you need to add private keys into the database. Here is example how to do that:

```
psql -U postgres -h localhost -p 6432 postgres # Password is 1212

insert into private_keys values('CASPER', 'YOUR PRIVATE KEY IN HEX');
insert into private_keys values('EVM', 'YOUR PRIVATE KEY IN HEX');
```

And restart signer server.

#### Connectors
Casper
```
cd boosty-bridge-services
go run cmd/casper/main.go run
```

Ethereum
```
cd boosty-bridge-services
go run cmd/eth/main.go run
```

#### Gateway

Gateway is a rest api for bridge.
```
cd boosty-bridge-services
go run cmd/gateway/main.go run
```

#### Front-end
Install node 18.12.1.
```
cd boosty-bridge-services/web/console
npm ci
npm run build
```

Make sure that the `dist` folder is in the path specified in the `.web.env` config in `STATIC_DIR` field.

Running web 
```
cd boosty-bridge-services
go run cmd/web_app/main.go run
```

The application should run under localhost:8089
