import { BigNumber, Contract, Signer, ethers, utils } from 'ethers';

import { NetworksClient } from '@/api/networks';
import { TransfersClient } from '@/api/transfers';
import appConfig from '@/app/configs/appConfig.json';
import { META_TAGS_CONFIG, parseMetaTag } from '@app/internal/parseMetaTag';
import { ABI, ERC20_ABI, EVMProvider, JsonRPCMethods } from '@/ethers';
import { NetworkNames } from '@/networks';
import { NetworksService } from '@/networks/service';
import { NetworkAddress, SignatureRequest } from '@/transfers';
import { TransfersService } from '@/transfers/service';
import { Wallet } from '@/wallets';

const networksClient = new NetworksClient();
const transfersClient = new TransfersClient();
const networksService = new NetworksService(networksClient);
const transfersService = new TransfersService(transfersClient);

/** Exposes all MetaMask ethers-related functionality. */
export class MetaMaskWallet implements Wallet {
    // @ts-ignore
    private readonly ethereum = window?.ethereum;

    /** Max gas limit to ethers iteraction. */
    private MAX_GAS_LIMIT: number = Number(parseMetaTag(META_TAGS_CONFIG.ETH_GAS_LIMIT));

    constructor(
        public provider = EVMProvider,
    ) { };

    /** Requests ethers signer.
     * @returns {Signer} - an abstraction of an Ethereum Account
     * to sign messages and transactions and send signed transactions
     * to the Ethereum Network to execute state changing operations.
     */
    private async getSigner(): Promise<Signer> {
        return await this.provider.getSigner();
    };

    /** Formats amount to BigNumber.
    * @param {string} amount - Transaction token amount
    * @returns {BigNumber}
    */
    private async formatAmountToBigNumber(amount: string): Promise<BigNumber> {
        return await utils.parseEther(amount);
    };

    /** Requests connected contract instance.
     * @param {string} address
     * @param {any} ABI - Application binary interface
     * @returns {Contract} - Contact instance
     */
    private async getContract(address: string, ABI: any): Promise<Contract> {
        const signer: Signer = await this.getSigner();
        const contract: Contract = new ethers.Contract(address, ABI);

        return await contract.connect(signer);
    };

    /** Requests connected MetaMask wallet address.
     * @returns {string} - connected MetaMask wallet address.
     */
    public async address(): Promise<string> {
        const signer: Signer = await this.getSigner();
        return await signer.getAddress();
    };

    /** Signs message and creates message raw signature.
    * @param {message} - bridge authentication message to sign.
    * @returns {string} - signed raw signature.
    */
    public async sign(message: string): Promise<string> {
        const signer: Signer = await this.getSigner();
        return await signer.signMessage(message);
    };

    /** Requests connection to ethereum node. */
    public async connect(): Promise<void> {
        if (!this.ethereum.providers) {
            await this.ethereum.request({ method: JsonRPCMethods.requestAccounts });
            return;
        }
        const provider = this.ethereum.providers.find(({ isMetaMask }: { isMetaMask: boolean }) => isMetaMask);
        await provider.request({ method: JsonRPCMethods.requestAccounts });
    };

    /** Approves transaction request.
    * @param {string} amount - Transaction token amount
    */
    public async approve(amount: string) {
        const contract = await this.getContract(parseMetaTag(META_TAGS_CONFIG.ETH_TOKEN_CONTRACT), ERC20_ABI);
        const parsedAmount = await this.formatAmountToBigNumber(amount);
        if (!contract) {
            return;
        }
        await contract.approve(parseMetaTag(META_TAGS_CONFIG.ETH_BRIDGE_CONTRACT), parsedAmount, { gasLimit: this.MAX_GAS_LIMIT });
    };

    /** Sends transaction.
    * @param {string} receiver - wallet address to receive transaction token amount
    */
    public async sendTransaction(receiver: string, amount: string): Promise<void> {
        await this.approve(amount);
        const address = await this.address();
        const contract = await this.getContract(parseMetaTag(META_TAGS_CONFIG.ETH_BRIDGE_CONTRACT), ABI);
        const sender = new NetworkAddress(address, NetworkNames.GOERLI);
        const destination = new NetworkAddress(`account-hash-${receiver}`, NetworkNames.CASPER_TEST);
        const supportedTokens = await networksService.supportedTokens(appConfig.numbers.ONE_NUMBER);
        const tokenId = supportedTokens[appConfig.numbers.ZERO_NUMBER].id;
        const signatureRequest = new SignatureRequest(
            sender,
            tokenId,
            amount,
            destination,
        );
        const signature = await transfersService.signature(signatureRequest);
        await contract.bridgeIn(
            parseMetaTag(META_TAGS_CONFIG.ETH_TOKEN_CONTRACT),
            signature.amount,
            signature.gasComission,
            signature.destination.networkName,
            signature.destination.address,
            signature.deadline,
            signature.nonce,
            signature.signature,
            { gasLimit: this.MAX_GAS_LIMIT },
        );
    };
};
