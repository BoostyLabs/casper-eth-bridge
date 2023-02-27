import {config as dotenvConfig} from "dotenv";
import '@nomiclabs/hardhat-ethers'
import '@primitivefi/hardhat-dodoc';
import "@typechain/hardhat";
import "@nomiclabs/hardhat-etherscan";
import "hardhat-gas-reporter";
import "solidity-coverage";
import {resolve} from "path";
import "@nomicfoundation/hardhat-chai-matchers";
import chai from "chai";
import { solidity } from "ethereum-waffle";
chai.use(solidity);


// You need to export an object to set up your config
// Go to https://hardhat.org/config/ to learn more

/**
 * @type import('hardhat/config').HardhatUserConfig
 */

dotenvConfig({path: resolve(__dirname, "./.env")});


export default {
  defaultNetwork: "hardhat",
  solidity: {
    version: "0.8.17",
    settings: {
      optimizer: {
        enabled: true,
        runs: 200
      }
    }
  },
  networks: {
    hardhat: {},
    goerli: {
      url: process.env.GOERLI_URL,
      accounts: [process.env.DEPLOY_KEY],
      //gas: 2100000,
      gasPrice: 120000000000
    },
    ethereum: {
      url: process.env.MAIN_URL,
      accounts: [process.env.DEPLOY_KEY],
    },
    avax: {
      url: process.env.AVAX_URL,
      accounts: [process.env.DEPLOY_KEY],
      gas: 2100000,
    },
    mumbai: {
      url: process.env.MUMBAI_URL,
      accounts: [process.env.DEPLOY_KEY],
      gas: 2100000,
    },
    bsc: {
      url: process.env.BSC_URL,
      accounts: [process.env.DEPLOY_KEY],
    }
  },
  dodoc: {
    runOnCompile: true,
    debugMode: false,
  },
  etherscan: {
    apiKey: process.env.ETHERSCAN_API_KEY
  },
  gasReporter: {
    currency: "USD",
    enabled: !!process.env.REPORT_GAS
  }
};