import {ethers} from "hardhat";
import {expect} from "chai";

import {SignerWithAddress} from "@nomiclabs/hardhat-ethers/src/signers";
import { Wallet } from "ethers";
import { Bridge, TestToken } from "../typechain-types";
import { signMessage, getCurrentTimeFromNetwork } from "./util";
// TODO:
// bridgeOut
// expired signature
// add contract to signature?

describe("Multimint contract", function () {

    let bridgeContract: Bridge;
    let tokenContract: TestToken;
    let tokenContract2: TestToken;
    let owner: SignerWithAddress;
    let user1: SignerWithAddress;
    let user2: SignerWithAddress;

    let systemWallet = new Wallet("ecff8b9c717a56b30f35a75db85342a1b42fcfe8540a733c73cc9ef38a165a56")
    const initialSupply = ethers.utils.parseEther("10000");
    const amountToTransfer = ethers.utils.parseEther("1000");
    const testGasCommission = ethers.utils.parseEther("100");
    const testGasIncorrectCommission = ethers.utils.parseEther("1000");
    const stableCommissionPercent = 400;

    const destinationChain = "Solana";
    const destinationAddress = "4zXwdbUDWo1S5AP2CEfv4zAPRds5PQUG1dyqLLvib2xu";
    const bridgeOutTransactionId = 1011;
    const TYPES_FOR_SIGNATURE_BRIDGE_IN = ["address", "address", "address", "uint256", "uint256", "string", "string", "uint256", "uint256"];
    const TYPES_FOR_SIGNATURE_TRANSFER_OUT = ["address", "address", "address", "uint256", "uint256", "uint256"];

    const getAmountToReturnAndTotalCommission = async () => {
        const totalCommission = await bridgeContract.getTotalCommission(amountToTransfer, testGasCommission);    
        const amountToReturn = amountToTransfer.sub(totalCommission);
        return [totalCommission, amountToReturn]
    };

    const bridgeInTokens = async (nonce: number, gasCommission = testGasCommission) => {
        
        await tokenContract.connect(user1).approve(bridgeContract.address, amountToTransfer);

        const deadline = await getCurrentTimeFromNetwork() + 84_000;
        const signatureBridgeIn = signMessage(
            TYPES_FOR_SIGNATURE_BRIDGE_IN,
            [user1.address, bridgeContract.address, tokenContract.address, amountToTransfer, gasCommission, destinationChain, destinationAddress, deadline, nonce], 
            systemWallet
        );

        return bridgeContract.connect(user1).bridgeIn(
            tokenContract.address,
            amountToTransfer, 
            gasCommission,
            destinationChain,
            destinationAddress,
            deadline,
            nonce,
            signatureBridgeIn
        )
    }

    this.beforeAll(async () => {
        // @ts-ignore
        [owner, user1, user2] = await ethers.getSigners() as SignerWithAddress;
        const BridgeContract = await ethers.getContractFactory("Bridge");
        const TestTokenContract = await ethers.getContractFactory("TestToken");
        const TestTokenContract2 = await ethers.getContractFactory("TestToken");
        bridgeContract = await BridgeContract.deploy(systemWallet.address) as Bridge;
        tokenContract = await TestTokenContract.deploy(initialSupply) as TestToken;
        tokenContract2 = await TestTokenContract2.deploy(initialSupply) as TestToken;
        await tokenContract.transfer(user1.address, amountToTransfer);
    });

    it("should be deployed with correct values", async function () {
        expect(await bridgeContract.getStableCommissionPercent()).to.equal(stableCommissionPercent);
        expect(await tokenContract.balanceOf(user1.address)).to.equal(amountToTransfer);
    });

    it("should allow user to transfer tokens via Bridge and withdraw them back", async function () {
        await expect(bridgeInTokens(1)).to.emit(bridgeContract, "BridgeFundsIn")
            .withArgs(
                user1.address, 
                1, 
                tokenContract.address, 
                amountToTransfer, 
                stableCommissionPercent,
                testGasCommission,
                destinationChain, 
                destinationAddress
            );
        // Tokens transferred and accounted to the Bridge contract
        expect( await tokenContract.balanceOf(user1.address)).to.equal(0);
        expect( await tokenContract.balanceOf(bridgeContract.address)).to.equal(amountToTransfer);
    });

    it("user transfer his tokens out", async ()=> {
        const nonce = 2;
        const [totalCommission, amountToReturn] = await getAmountToReturnAndTotalCommission();

        const signatureTransferOut = signMessage(
            TYPES_FOR_SIGNATURE_TRANSFER_OUT,
            [bridgeContract.address, tokenContract.address, user1.address, amountToReturn, totalCommission, nonce], 
            systemWallet
        );
        await expect(bridgeContract.connect(user1).transferOut(
            tokenContract.address,
            user1.address,
            amountToReturn, 
            totalCommission,
            2,
            signatureTransferOut
        )).to.emit(bridgeContract, "TransferOut");
        expect( await tokenContract.balanceOf(user1.address)).to.equal(amountToTransfer);
        expect( await tokenContract.balanceOf(bridgeContract.address)).to.equal(0);
    });

    it("owner bridge tokens out", async function () {
        const [, amountToReturn] = await getAmountToReturnAndTotalCommission();
        await expect(bridgeInTokens(3)).to.emit(bridgeContract, "BridgeFundsIn")
            .withArgs(
                user1.address, 
                3, 
                tokenContract.address, 
                amountToTransfer, 
                stableCommissionPercent,
                testGasCommission,
                destinationChain, 
                destinationAddress
            );
        await expect(bridgeContract.bridgeOut(tokenContract.address, user1.address, amountToReturn, bridgeOutTransactionId, 'anySourceChain', 'anySourceAddress'))
        .to.emit(bridgeContract, "BridgeFundsOut").withArgs(
            user1.address, 
            tokenContract.address,
            amountToReturn,
            bridgeOutTransactionId,
            'anySourceChain',
            'anySourceAddress'
        );
        expect( await tokenContract.balanceOf(user1.address)).to.equal(amountToReturn);
        expect( await tokenContract.balanceOf(tokenContract.address)).to.equal(0);

        const [totalCommission, ] = await getAmountToReturnAndTotalCommission();
    });

    it("should return correct commission in pool", async function () {
        const [totalCommission, ] = await getAmountToReturnAndTotalCommission();    
        const commissionInPool = await bridgeContract.getCommissionPoolAmount(tokenContract.address);
        expect(totalCommission).to.equal(commissionInPool);
        
    });

    it("should correctly calculate total commission", async function () {
        expect(await bridgeContract.getTotalCommission(amountToTransfer, testGasCommission))
        .to.equal(amountToTransfer.mul(stableCommissionPercent).div(10000).add(testGasCommission));

        const [totalCommission, ] = await getAmountToReturnAndTotalCommission();
    });


    it("should withdraw commission", async function () {
        const initialOwnerBalance = await tokenContract.balanceOf(owner.address);
        const [totalCommission, ] = await getAmountToReturnAndTotalCommission();  
        await expect(bridgeContract.withdrawCommission(
            tokenContract.address,
            totalCommission
        )).to.emit(bridgeContract, "WithdrawCommission").withArgs(
            tokenContract.address, 
            totalCommission
        );
        expect( await tokenContract.balanceOf(owner.address)).to.equal(initialOwnerBalance.add(totalCommission));
        expect( await tokenContract.balanceOf(bridgeContract.address)).to.equal(0);
        expect( await bridgeContract.getCommissionPoolAmount(bridgeContract.address)).to.equal(0);
    });

    /////////////////////////////////////////////////////// NEGATIVE CASES ///////////////////////////////////////////////////////


    it("should not allow transaction if total commission greater than transferred amount", async function () {
        await expect(bridgeInTokens(1, testGasIncorrectCommission))
            .to.be.revertedWith('CommissionGreaterThanAmount');;
    });

    it("Owner can not bridge out more tokens than available in pool", async function () {
        await tokenContract.transfer(user1.address, amountToTransfer);
        await bridgeInTokens(22);
        const [, amountToReturn] = await getAmountToReturnAndTotalCommission();  
        await expect(bridgeContract.bridgeOut(tokenContract.address, user1.address, amountToReturn.add(1), bridgeOutTransactionId, 'anySourceChain', 'anySourceAddress'))
            .to.be.revertedWith('AmountExceedBridgePool');
    });

    it("Owner can not withdraw more tokens than available in commission pool", async function () {
        const commissionInPool = await bridgeContract.getCommissionPoolAmount(tokenContract.address);    
        await expect(bridgeContract.withdrawCommission(
            tokenContract.address,
            commissionInPool.add(1)
        )).to.be.revertedWith('AmountExceedCommissionPool');
    });

    it("arbitrary user con not set commission percent", async function () {
        await expect(bridgeContract.connect(user1).setStableCommissionPercent(100)).to.be.revertedWith('Ownable: caller is not the owner');
    });

    it("arbitrary user can not withdraw commission", async function () {
        const [totalCommission, ] = await getAmountToReturnAndTotalCommission();  
        await expect(bridgeContract.connect(user1).withdrawCommission(
            tokenContract.address,
            totalCommission
        )).to.be.revertedWith('Ownable: caller is not the owner');
    });

    it("arbitrary user can not bridge tokens out", async function () {
        await expect(bridgeContract.connect(user1).bridgeOut(tokenContract.address, user1.address, amountToTransfer, bridgeOutTransactionId, destinationChain, destinationAddress))
            .to.be.revertedWith('Ownable: caller is not the owner');;
    });

    it("user can not cheat bridgeIn", async function () {
        const deadline = await getCurrentTimeFromNetwork() + 1000;
        const signatureBridgeIn = signMessage(
            TYPES_FOR_SIGNATURE_BRIDGE_IN,
            [user1.address, bridgeContract.address, tokenContract.address, amountToTransfer, testGasCommission, destinationChain, destinationAddress, deadline, 5], 
            systemWallet
        );

        const incorrectNonce = 1;    
        await expect(bridgeContract.connect(user1).bridgeIn(
            tokenContract.address,
            amountToTransfer, 
            testGasCommission,
            destinationChain,
            destinationAddress,
            deadline,
            incorrectNonce,
            signatureBridgeIn
        )).to.be.revertedWith('AlreadyUsedSignature');

        const incorrectContract = tokenContract2.address;
        await expect(bridgeContract.connect(user1).bridgeIn(
            incorrectContract,
            amountToTransfer, 
            testGasCommission,
            destinationChain,
            destinationAddress,
            deadline,
            5,
            signatureBridgeIn
        )).to.be.revertedWith('InvalidSignature');

        const incorrectSum = amountToTransfer.add(10000);
        await expect(bridgeContract.connect(user1).bridgeIn(
            tokenContract.address,
            incorrectSum, 
            testGasCommission,
            destinationChain,
            destinationAddress,
            deadline,
            5,
            signatureBridgeIn
        )).to.be.revertedWith('InvalidSignature');

        const incorrectCommission = testGasCommission.sub(5);
        await expect(bridgeContract.connect(user1).bridgeIn(
            tokenContract.address,
            amountToTransfer, 
            incorrectCommission,
            destinationChain,
            destinationAddress,
            deadline,
            5,
            signatureBridgeIn
        )).to.be.revertedWith('InvalidSignature');

        const incorrectNetwork = "Near";
        await expect(bridgeContract.connect(user1).bridgeIn(
            tokenContract.address,
            amountToTransfer, 
            testGasCommission,
            incorrectNetwork,
            destinationAddress,
            deadline,
            5,
            signatureBridgeIn
        )).to.be.revertedWith('InvalidSignature');

        const incorrectDestinationAddress = "Near";
        await expect(bridgeContract.connect(user1).bridgeIn(
            tokenContract.address,
            amountToTransfer, 
            testGasCommission,
            destinationChain,
            incorrectDestinationAddress,
            deadline,
            5,
            signatureBridgeIn
        )).to.be.revertedWith('InvalidSignature');

        const incorrectDeadline = deadline + 100;
        await expect(bridgeContract.connect(user1).bridgeIn(
            tokenContract.address,
            amountToTransfer, 
            testGasCommission,
            destinationChain,
            destinationAddress,
            incorrectDeadline,
            5,
            signatureBridgeIn
        )).to.be.revertedWith('InvalidSignature');


        // bridgeContract will be invalid in this context
        const realContract = '0x47761b7E9E203aF9853107FbC6d8D0353Cda7a0e'; 
        const InvalidSignatureBridgeIn = signMessage(
            TYPES_FOR_SIGNATURE_BRIDGE_IN,
            [user1.address, realContract, tokenContract.address, amountToTransfer, testGasCommission, destinationChain, destinationAddress, deadline, 5], 
            systemWallet
        );

        await expect(bridgeContract.connect(user1).bridgeIn(
            tokenContract.address,
            amountToTransfer, 
            testGasCommission,
            destinationChain,
            destinationAddress,
            deadline,
            5,
            InvalidSignatureBridgeIn
        )).to.be.revertedWith('InvalidSignature');
    });


    it("user can not cheat transferOut", async function () {
        
        const [totalCommission, amountToReturn] = await getAmountToReturnAndTotalCommission();

        const signatureTransferOut = signMessage(
            TYPES_FOR_SIGNATURE_TRANSFER_OUT,
            [bridgeContract.address, tokenContract.address, user1.address, amountToReturn, totalCommission, 5], 
            systemWallet
        );

        const incorrectToken = '0x3a69f406e56962Afa584F65c090aa781e8180241';
        await expect(bridgeContract.connect(user1).transferOut(
            incorrectToken,
            user1.address,
            amountToTransfer.sub(totalCommission), 
            totalCommission,
            5,
            signatureTransferOut
        )).to.be.revertedWith('InvalidSignature');

        const incorrecRecipient = user2.address;
        await expect(bridgeContract.connect(user1).transferOut(
            tokenContract.address,
            incorrecRecipient,
            amountToTransfer.sub(totalCommission), 
            totalCommission,
            5,
            signatureTransferOut
        )).to.be.revertedWith('InvalidSignature');

        const incorrectAmount = amountToTransfer.sub(totalCommission).add(2000);
        await expect(bridgeContract.connect(user1).transferOut(
            tokenContract.address,
            user1.address,
            incorrectAmount, 
            totalCommission,
            5,
            signatureTransferOut
        )).to.be.revertedWith('InvalidSignature');

        const incorrectCommission = totalCommission.sub(500);
        await expect(bridgeContract.connect(user1).transferOut(
            tokenContract.address,
            user1.address,
            amountToTransfer.sub(totalCommission), 
            incorrectCommission,
            5,
            signatureTransferOut
        )).to.be.revertedWith('InvalidSignature');

        const incorrecNonce = 2;
        await expect(bridgeContract.connect(user1).transferOut(
            tokenContract.address,
            user1.address,
            amountToTransfer.sub(totalCommission), 
            totalCommission,
            incorrecNonce,
            signatureTransferOut
        )).to.be.revertedWith('AlreadyUsedSignature');

        // Bridge contract will be invalid
        const realContractAddress = '0x47761b7E9E203aF9853107FbC6d8D0353Cda7a0e';    
        const invalidSignatureTransferOut = signMessage(
            TYPES_FOR_SIGNATURE_TRANSFER_OUT,
            [realContractAddress, tokenContract.address, user1.address, amountToReturn, totalCommission, 5], 
            systemWallet
        );

        await expect(bridgeContract.connect(user1).transferOut(
            tokenContract.address,
            user1.address,
            amountToTransfer.sub(totalCommission), 
            totalCommission,
            5,
            invalidSignatureTransferOut
        )).to.be.revertedWith('InvalidSignature');

    });

    it("should set commission percent", async function () {
        await bridgeContract.setStableCommissionPercent(1000);
        expect(await bridgeContract.getStableCommissionPercent()).to.equal(1000);
    });

});