import {ethers} from "hardhat";
import {BigNumber, Wallet} from "ethers";


export async function addSecondsToNetwork(time: any){
    await setTimeToNetwork(await getCurrentTimeFromNetwork() + time);
}

export async function setTimeToNetwork(time: Number){
    await ethers.provider.send('evm_mine', [time]);
}

export async function getCurrentTimeFromNetwork(){
    const blockNumBefore = await ethers.provider.getBlockNumber();
    const blockBefore = await ethers.provider.getBlock(blockNumBefore);
    return blockBefore.timestamp;
}

export async function signMessage(types: ReadonlyArray<string>, values: ReadonlyArray<any>, wallet: Wallet) {
    const data = ethers.utils.solidityKeccak256(types, values);
    return wallet.signMessage(ethers.utils.arrayify(data));
}

export function ethToWei(wei: BigNumber) {
  return wei.mul(BigNumber.from(10).pow(BigNumber.from(18)));
}

export function getPercent(value: BigNumber, percent: any) {
  let bigPercent = BigNumber.from(33_00000).mul(BigNumber.from(percent)).div(BigNumber.from(33));
  return value.mul(bigPercent).div(BigNumber.from(100_00000));
}
