const hre = require("hardhat");
const {ethers} = require("hardhat");
async function main() {
    const [owner] = await ethers.getSigners();
    console.log('DEPLOYER ADDRESS : %s', owner.address);

    const TestToken = await ethers.getContractFactory("TestToken");
    const Bridge = await ethers.getContractFactory("Bridge");
    const emission = "100000000000000000000000000";
    const signer = "0x9032d7eb50b5b4a48c21035f34e0A84e54921D75";
    const bridge = await Bridge.deploy(signer);
    const erc20token = await TestToken.deploy(emission);
    console.log('ERC20 : %s', erc20token.address);
    console.log('Bridge : %s', bridge.address);
}
main()
    .then(() => process.exit(0))
    .catch((error) => {
        console.error(error);
        process.exit(1);
    });