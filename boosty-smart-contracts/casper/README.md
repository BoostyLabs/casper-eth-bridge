# Golden-Gate Casper

## Development

```sh
$ cd contract-bridge
$ just build-contract-test
```

Open `contract-bridge-tests/lib.rs` and run tests. If you use Rust Analyzer, you can do that directly from the file



## Release /  Deploy

```sh
$ cd contract-bridge
$ just build-contract-release
$ cd ../util
$ just run-release deploy-bridge-contract -c ../contract-bridge/bridge-contract.wasm
```
You'll receive a hash like: `c7e1bc80565e834ebf0ad24331a7b93dd820db6991d51e5b40eac5afe041680d`

Go to: `https://testnet.cspr.live/deploy/c7e1bc80565e834ebf0ad24331a7b93dd820db6991d51e5b40eac5afe041680d`
Go to: `https://testnet.cspr.live/` - > View Account -> Named Keys and found keys of Deployed contract


## Typical Errors:


#### `Out of Gas` 

Open: `util/src/bin/cli.rs -> deploy_bridge_contract method` and increase payment


#### `Opcode ... ` errors 

Open: `contract-bridge/justfile -> build-contract-release section` and change `wasm-opt -O4` to `cp`. It may help however not resolve the problem with gas costs.