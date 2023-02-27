import { BigNumber, Contract, Signer, ethers, utils } from 'ethers';

import { NetworksClient } from '@/api/networks';
import { TransfersClient } from '@/api/transfers';
import appConfig from '@/app/configs/appConfig.json';
import { LocalStorageKeys } from '@app/hooks/useLocalStorage';
import { META_TAGS_CONFIG, parseMetaTag } from '@app/internal/parseMetaTag';
import { ABI, ERC20_ABI, EVMProvider, JsonRPCMethods } from '@/ethers';
import { NetworkNames, NetworkTypes, NetworksChains } from '@/networks';
import { NetworksService } from '@/networks/service';
import { CancelSignatureRequest, NetworkAddress, SignatureRequest } from '@/transfers';
import { TransfersService } from '@/transfers/service';
import { Wallet } from '@/wallets';

const networksClient = new NetworksClient();
const transfersClient = new TransfersClient();
const networksService = new NetworksService(networksClient);
const transfersService = new TransfersService(transfersClient);

/** Exposes all MetaMask ethers-related functionality. */
export class MetaMaskWallet implements Wallet {
    /** Max gas limit to ethers iteraction. */
    private MAX_GAS_LIMIT: number = Number(parseMetaTag(META_TAGS_CONFIG.ETH_GAS_LIMIT));
    private SENDER_NETWORK_ID: number = Number(localStorage.getItem(LocalStorageKeys.senderNetworkId));
    private RECIPIENT_NETWORK_ID: number = Number(localStorage.getItem(LocalStorageKeys.recipientNetworkId));
    private selectedProvider: any = null;

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

    /** Requests token contract address depends on network name.
    * @param {string} networkName - network name.
    * @returns {string} - contract address .
     */
    private getTokenContractAddress(networkName: NetworkNames): string {
        let address: string = '';
        // TODO: rework with API implementation.
        if (networkName === NetworkNames.GOERLI) {
            address = parseMetaTag(META_TAGS_CONFIG.ETH_TOKEN_CONTRACT);
        }
        if (networkName === NetworkNames.MUMBAI) {
            address = parseMetaTag(META_TAGS_CONFIG.POLYGON_TOKEN_CONTRACT);
        }
        if (networkName === NetworkNames.BNB_TEST) {
            address = parseMetaTag(META_TAGS_CONFIG.BNB_TOKEN_CONTRACT)
        }
        if (networkName === NetworkNames.AVALANCHE_TEST) {
            address = parseMetaTag(META_TAGS_CONFIG.AVALANCHE_TOKEN_CONTRACT)
        }

        return address;
    };

    /** Requests bridge contract address depends on network name.
    * @param {string} networkName - network name.
    * @returns {string} - contract address .
     */
    private getBridgeContractAddress(networkName: NetworkNames): string {
        let address: string = '';
        // TODO: rework with API implementation.
        if (networkName === NetworkNames.GOERLI) {
            address = parseMetaTag(META_TAGS_CONFIG.ETH_BRIDGE_CONTRACT);
        }
        if (networkName === NetworkNames.MUMBAI) {
            address = parseMetaTag(META_TAGS_CONFIG.POLYGON_BRIDGE_CONTRACT);
        }
        if (networkName === NetworkNames.BNB_TEST) {
            address = parseMetaTag(META_TAGS_CONFIG.BNB_BRIDGE_CONTRACT);
        }
        if (networkName === NetworkNames.AVALANCHE_TEST) {
            address = parseMetaTag(META_TAGS_CONFIG.AVALANCHE_BRIDGE_CONTRACT)
        }

        return address;
    };

    /** Sets selected MetaMask provider. */
    private setSelectedProvider() {
        const { provider } = this.provider;
        if (!provider.providers) {
            this.selectedProvider = provider;
            return;
        }
        this.selectedProvider = provider.providers.find(({ isMetaMask }: { isMetaMask: boolean }) => isMetaMask);
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

    /** Switches metamask network depends on sender network name. */
    private async switchNetwork(senderNetworkName: NetworkNames) {
        /** Metadata about the chain that MetaMask will switch to. */
        const params = [];
        // TODO: rework with API implementation.
        switch (senderNetworkName) {
            case NetworkNames.MUMBAI:
                params.push({ chainId: NetworksChains.MUMBAI });
                break;
            case NetworkNames.BNB_TEST:
                params.push({ chainId: NetworksChains.BNB_TEST });
                break;
            case NetworkNames.AVALANCHE_TEST:
                params.push({ chainId: NetworksChains.AVALANCHE_TEST });
                break;
            case NetworkNames.GOERLI:
            default:
                params.push({ chainId: NetworksChains.GOERLI })
                break;
        }
        this.setSelectedProvider();
        await this.selectedProvider.request({ method: JsonRPCMethods.switchNetwork, params });
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
        this.setSelectedProvider();
        await this.selectedProvider.request({ method: JsonRPCMethods.requestAccounts });
    };

    /** Approves transaction request.
    * @param {string} amount - Transaction token amount
    */
    public async approve(amount: string, senderNetworkName: NetworkNames) {
        const tokenContractAddress: string = this.getTokenContractAddress(senderNetworkName);
        const bridgeContractAddress: string = this.getBridgeContractAddress(senderNetworkName);
        const contract = await this.getContract(tokenContractAddress, ERC20_ABI);
        const parsedAmount = await this.formatAmountToBigNumber(amount);
        if (!contract) {
            return;
        }
        await this.switchNetwork(senderNetworkName);
        await contract.approve(bridgeContractAddress, parsedAmount, { gasLimit: this.MAX_GAS_LIMIT });
    };

    /** Sends transaction.
    * @param {string} receiver - wallet address to receive transaction token amount
    */
    public async sendTransaction(receiver: string, amount: string): Promise<void> {
        const address = await this.address();
        const connectedNetworks = await networksService.connected();
        const senderNetwork = connectedNetworks.find(network => network.id === this.SENDER_NETWORK_ID);
        if (!senderNetwork) {
            return;
        }
        const recipientNetwork = connectedNetworks.find(network => network.id === this.RECIPIENT_NETWORK_ID);
        if (!recipientNetwork) {
            return;
        }
        await this.approve(amount, senderNetwork.name);
        const bridgeContractAddress: string = this.getBridgeContractAddress(senderNetwork.name);
        const contract = await this.getContract(bridgeContractAddress, ABI);
        const sender = new NetworkAddress(address, senderNetwork.name);
        const destination = new NetworkAddress(receiver, recipientNetwork.name);
        if (recipientNetwork.type === NetworkTypes.CASPER) {
            destination.address = `account-hash-${receiver}`
        }
        const supportedTokens = await networksService.supportedTokens(Number(senderNetwork.id));
        const tokenId = supportedTokens[appConfig.numbers.ZERO_NUMBER].id;
        const signatureRequest = new SignatureRequest(
            sender,
            tokenId,
            amount,
            destination,
        );
        const signature = await transfersService.signature(signatureRequest);
        const tokenContractAddress = this.getTokenContractAddress(senderNetwork.name);
        await contract.bridgeIn(
            tokenContractAddress,
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

    /** Canceles transaction.
    * @param {CancelSignatureRequest} cancelSignatureRequest - fields needed to generate signature to cancel transfer.
    * @param {string} amount
    */
    public async cancelTransaction(cancelSignatureRequest: CancelSignatureRequest): Promise<void> {
        const cancelTransferResponse = await transfersService.cancelSignature(cancelSignatureRequest);
        const connectedNetworks = await networksService.connected();
        const senderNetwork = connectedNetworks.find(network => network.id === this.SENDER_NETWORK_ID);
        if (!senderNetwork) {
            return;
        }
        const bridgeContractAddress: string = this.getBridgeContractAddress(senderNetwork.name);
        const contract = await this.getContract(bridgeContractAddress, ABI);
        await contract.transferOut(
            cancelTransferResponse.token,
            cancelTransferResponse.recipient,
            cancelTransferResponse.amount,
            cancelTransferResponse.commission,
            cancelTransferResponse.nonce,
            cancelTransferResponse.signature,
        );
    };
};
