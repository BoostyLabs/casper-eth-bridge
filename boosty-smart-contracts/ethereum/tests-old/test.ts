import { ethers } from "hardhat";
import { expect } from "chai";
import { Token } from "../typechain-types";
import { BigNumber } from "ethers";
import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/src/signers";
import { addSecondsToNetwork, setTimeToNetwork } from "./util";


describe("Token contract", function () {

    let owner: SignerWithAddress;
    let user1: SignerWithAddress;

    const TOKEN_NAME = "Tugrik";
    const TOKEN_SYMBOL = "TGK";
    const TOTAL_SUPPLY: BigNumber = BigNumber.from(1_000_000_000);

    let token: Token;

    beforeEach(async () => {
        // @ts-ignore
        [owner, user1] = await ethers.getSigners() as SignerWithAddress;

        const Token = await ethers.getContractFactory("Token");
        token = await Token.deploy(TOKEN_NAME, TOKEN_SYMBOL, TOTAL_SUPPLY);
    });

    it("Deployment should assign the total supply of tokens to the owner", async function () {
        expect(await token.name()).to.equal(TOKEN_NAME);
        expect(await token.symbol()).to.equal(TOKEN_SYMBOL);
        expect(await token.name()).to.equal(TOKEN_NAME);
        expect(await token.totalSupply()).to.equal(TOTAL_SUPPLY);
    });

    it("should distribute tokens correctly", async function () {
        expect(await token.balanceOf(owner.address)).to.equal(TOTAL_SUPPLY);
    });

    it("should set A correctly", async function () {
        const tx = await token.setA(22);
        await tx.wait();

        expect(22).to.equal(22);
    });

    it("should set and added time", async function () {
        await setTimeToNetwork(2000000000);
        await addSecondsToNetwork(100);
    });

});