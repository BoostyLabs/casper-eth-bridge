# Golden-Gate

- Tests from the box

- TypeChain typescript types for Smart Contracts generation upon compilation by `typechain`


- Docs - generation upon compilation by `dodoc`


- Autoformatting - added `prettier` tool for auto format smart contracts


- Lint - added `solhint` tool for auto format smart contracts


- Gar Reporter - will be generated upon tests execution


- Test Coverage Reporter - will be generated upon tests execution


## Usage

### Pre Requisites

Before running any command, make sure to install dependencies:

```sh
$ npm i
```

Copy .env.example to .env and populate with your values (TEST_ACC1= private key of your wallet)

### Compile

Compile the smart contracts:

```sh
$ npm run compile
```

### Test

Run the tests:

```sh
$ npm run test
```
Coverage and Gas reports will be generated

### Coverage

Generate the code coverage report:

```sh
$ npm run coverage
```

### Clean

Delete the smart contract artifacts, the coverage reports and the Hardhat cache:

```sh
$ npm run clean
```

### TypeChain

Compile the smart contracts and generate TypeChain artifacts upon:

```sh
$ npm run compile
```

### Docs Generation

Will be generated according to comments in Smart Contracts upon:

```sh
$ npm run compile
```

### Lint Solidity

Lint the Solidity code:

```sh
$ npm run lint
```

### Deploy

```sh
WILL BE INTRODUCED
```

#### Custom commands:
```shell
npm run clean 
npm run compile 
npm run deploy 
npm run lint 
npm run lint:prettier 
npm run test 
```

#### HardHat commands:
```shell
npx hardhat accounts
npx hardhat compile
npx hardhat clean
npx hardhat test
npx hardhat node
node scripts/sample-script.js
npx hardhat help
```
