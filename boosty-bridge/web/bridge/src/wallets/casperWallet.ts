import { Buffer } from 'buffer';
import { CLPublicKey, CLValueBuilder, DeployUtil, RuntimeArgs, encodeBase16 } from 'casper-js-sdk';

import { CasperClient } from '@/api/casper';
import { NetworksClient } from '@/api/networks';
import { TransfersClient } from '@/api/transfers';
import { TransfersService } from '@/transfers/service';
import { NetworksService } from '@/networks/service';
import appConfig from '@/app/configs/appConfig.json';
import { META_TAGS_CONFIG, parseMetaTag } from '@app/internal/parseMetaTag';
import { CasperEntryPoints, CasperRuntimeArgs } from '@/casper';
import { NetworkNames, NetworkTypes } from '@/networks';
import { NetworkAddress, SignatureRequest } from '@/transfers';
import { Wallet } from '@/wallets';
import { LocalStorageKeys } from '@app/hooks/useLocalStorage';

// TODO: Need research how get this value from casper.
/** For native-transfers the payment price is fixed. */
const PAYMENT_AMOUNT: number = 40000000000;

const networksClient = new NetworksClient();
const networksService = new NetworksService(networksClient);
const transfersClient = new TransfersClient();
const transfersService = new TransfersService(transfersClient);

/** Exposes all Casper functionality. */
export class CasperWallet implements Wallet {
    private readonly casper: CasperClient = new CasperClient();

    private readonly PAYMENT_AMOUNT: number = PAYMENT_AMOUNT;
    /** Defines Casper contract hash. */
    private readonly CONTRACT_HASH: string = parseMetaTag(META_TAGS_CONFIG.CASPER_BRIDGE_CONTRACT);
    /** Defines RPC node address. */
    private readonly RPC_NODE_ADDRESS: string = parseMetaTag(META_TAGS_CONFIG.CASPER_NODE_ADDRESS);
    private readonly RECIPIENT_NETWORK_ID: number = Number(localStorage.getItem(LocalStorageKeys.recipientNetworkId));

    constructor(
        public provider = window.casperlabsHelper,
    ) { };

    /** Checks if the site is connected to Casper extension.
     * @returns {boolean}
     */
    public async isSiteConnected(): Promise<boolean> {
        return await this.provider.isConnected();
    };

    /** Requests CasperCasper account-hash.
     * @returns {string} - active account-hash.
    */
    public async address(): Promise<string> {
        const publicKey = await this.getPublicKey();
        return encodeBase16(publicKey.toAccountHash());
    };

    /** Gerenates casper public key from hex string.
     * @returns {CLPublicKey} - active Casper public key.
     */
    private async getPublicKey(): Promise<CLPublicKey> {
        const publicKeyHex = await this.provider.getActivePublicKey()
        return await CLPublicKey.fromHex(publicKeyHex);
    };

    /** Converts contract hash string to byte array.
     * @param {string} contractHash
     * @returns {string} - contract address hash in byte array type.
     */
    private hashToByteArray = (contractHash: string) =>
        Uint8Array.from(Buffer.from(contractHash, 'hex'));

    /** Represents a collection of arguments passed to a smart contract.
     * @param {string} amount - transaction amount.
     * @param {string} destination - receiver wallet address.
     * @returns {RuntimeArgs} - arguments collection.
     */
    private async getRuntimeArgs(amount: string, destination: string): Promise<RuntimeArgs | null> {
        const connectedNetworks = await networksService.connected();
        const recipientNetwork = connectedNetworks.find(network => network.id === this.RECIPIENT_NETWORK_ID);
        if (!recipientNetwork) {
            return null;
        }
        const senderNetwork = connectedNetworks.find(network => network.type === NetworkTypes.CASPER);
        if (!senderNetwork) {
            return null;
        }
        const supportedTokens = await networksService.supportedTokens(Number(senderNetwork.id));
        const tokenId = supportedTokens[appConfig.numbers.ZERO_NUMBER].id;
        const address: string = await this.address();
        const sender = new NetworkAddress(address, senderNetwork.name);
        const recipient = new NetworkAddress(destination, recipientNetwork.name);
        const signatureRequest = new SignatureRequest(
            sender,
            tokenId,
            amount,
            recipient,
        );
        const signature = await transfersService.signature(signatureRequest);

        return await RuntimeArgs.fromMap({
            [CasperRuntimeArgs.TOKEN_CONTRACT]: CLValueBuilder.byteArray(this.hashToByteArray(parseMetaTag(META_TAGS_CONFIG.CASPER_TOKEN_CONTRACT))),
            [CasperRuntimeArgs.AMOUNT]: CLValueBuilder.u256(signature.amount),
            [CasperRuntimeArgs.GAS_COMMISSION]: CLValueBuilder.u256(signature.gasComission),
            [CasperRuntimeArgs.DEADLINE]: CLValueBuilder.u256(signature.deadline),
            [CasperRuntimeArgs.NONCE]: CLValueBuilder.u128(signature.nonce),
            [CasperRuntimeArgs.DESTINATION_CHAIN]: CLValueBuilder.string(signature.destination.networkName),
            [CasperRuntimeArgs.DESTINATION_ADDRESS]: CLValueBuilder.string(signature.destination.address),
            [CasperRuntimeArgs.SIGNATURE]: CLValueBuilder.byteArray(this.hashToByteArray(signature.signature.slice(appConfig.numbers.TWO_NUMBER))),
        });
    };

    /** Makes deploy transaction.
     * @param {string} chainName - transaction chain name
     * @param {Uint8Array} contractHashAsByteArray
     * @param {string} entryPoint - contract method name
     * @param {RuntimeArgs} runtimeArgs - arguments needed to deploy contract
     * @param {string} paymentAmount - transaction payment amount
     * @returns {DeployUtil.Deploy} - JSON representation of a deploy - can be constructed using the `DeployUtil.deployToJSON()` method.
     */
    private makeDeploy = async(
        chainName: string,
        contractHashAsByteArray: Uint8Array,
        entryPoint: string,
        runtimeArgs: RuntimeArgs,
        paymentAmount: number,
    ): Promise<DeployUtil.Deploy> => {
        const publicKey = await this.getPublicKey();
        const deploy = await DeployUtil.makeDeploy(
            new DeployUtil.DeployParams(publicKey, chainName, appConfig.numbers.ONE_NUMBER),
            DeployUtil.ExecutableDeployItem.newStoredContractByHash(
                contractHashAsByteArray,
                entryPoint,
                runtimeArgs
            ),
            DeployUtil.standardPayment(paymentAmount)
        );

        return deploy
    }

    /** Call contract method with Casper transaction parameters and return signed signature.
     * @param {CasperEntryPoints} entryPoint - casper entry point.
     * @param {RuntimeArgs} runtimeArgs - arguments collection passed to a smart contract.
     * @returns {string} - Casper transaction signature.
    */
    private contractCall = async(entryPoint: CasperEntryPoints, runtimeArgs: RuntimeArgs) => {
        const contractHashAsByteArray = this.hashToByteArray(this.CONTRACT_HASH);
        const deploy = await this.makeDeploy(
            NetworkNames.CASPER_TEST.toLowerCase(),
            contractHashAsByteArray,
            entryPoint,
            runtimeArgs,
            this.PAYMENT_AMOUNT,
        );
        const json = DeployUtil.deployToJson(deploy);
        const publicKeyHex = await this.provider.getActivePublicKey();
        const signature = await this.provider.sign(json, publicKeyHex);

        return signature;
    };

    /** Signs authenticated message and returns signature.
     * @param {string} message - authenticated message.
     * @returns {string} - signed signature.
     */
    public async sign(message: string): Promise<string> {
        const publicKey = await this.provider.getActivePublicKey()
        return await this.provider.signMessage(message, publicKey);
    };

    /** Requests Casper extension for app connection. */
    public async connect(): Promise<void> {
        await this.provider.requestConnection();
    };

    /** Sends transaction via Casper Wallet.
     * @param {string} amount - transaction amount.
     * @param {string} destination - receiver wallet address.
     */
    public async sendTransaction(amount: string, destination: string): Promise<void> {
        const runtimeArgs: RuntimeArgs | null = await this.getRuntimeArgs(amount, destination);
        if (!runtimeArgs) {
            return;
        }
        const deploy = await this.contractCall(CasperEntryPoints.SEND_TRANSACTION, runtimeArgs);
        await this.casper.sendTransaction(JSON.stringify(deploy), this.RPC_NODE_ADDRESS);
    };

    public async cancelTransaction(): Promise<void> {
        // TODO: implement.
    }
};
