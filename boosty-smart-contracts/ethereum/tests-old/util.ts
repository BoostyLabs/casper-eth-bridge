import {ethers} from "hardhat";
import {Wallet} from "ethers";

export async function addSecondsToNetwork(time:Number){
    await ethers.provider.send('evm_increaseTime', [time]);
}

export async function setTimeToNetwork(time:Number){
    await ethers.provider.send('evm_mine', [time]);
}

export async function getCurrentTimeFromNetwork(){
    const blockNumBefore = await ethers.provider.getBlockNumber();
    const blockBefore = await ethers.provider.getBlock(blockNumBefore);
    return blockBefore.timestamp;
}

export async function createSignature(signatureAbi: any, sender: any, contract: any, prKey: any) {
    let ABI = ["function Signature(address sender, address contract)"];
    let iFace = new ethers.utils.Interface(ABI);
    let encodeFunctionCall = iFace.encodeFunctionData("Signature", [sender.address, contract.address])

    let keccak256EncodeFunctionCall = ethers.utils.keccak256(encodeFunctionCall);
    let wallet = new Wallet(prKey)
    return wallet.signMessage(ethers.utils.arrayify(keccak256EncodeFunctionCall));
}