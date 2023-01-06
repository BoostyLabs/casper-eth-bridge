const hre = require("hardhat");
const {ethers} = require("hardhat");
async function main() {
    const [owner] = await ethers.getSigners();
    console.log('DEPLOYER ADDRESS : %s', owner.address);
    const Bridge = await ethers.getContractFactory("Bridge");
    const signer = "0x09111ca3BB247F5C531175DbC0F37e260c8f2a68";
    const bridge = await Bridge.deploy(signer);
    console.log('bridge.address', bridge.address);
}
main()
    .then(() => process.exit(0))
    .catch((error) => {
        console.error(error);
        process.exit(1);
    });