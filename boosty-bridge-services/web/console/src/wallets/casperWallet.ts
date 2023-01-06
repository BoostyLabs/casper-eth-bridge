import { Buffer } from 'buffer';
import { CLPublicKey, CLValueBuilder, DeployUtil, RuntimeArgs } from 'casper-js-sdk';

import { CasperClient } from '@/api/casper';
import appConfig from '@/app/configs/appConfig.json';
import { META_TAGS_CONFIG, parseMetaTag } from '@app/internal/parseMetaTag';
import { CasperEntryPoints, CasperRuntimeArgs } from '@/casper';
import { NetworkNames } from '@/networks';
import { Wallet } from '@/wallets';

// TODO: Need research how get this value from casper.
/** For native-transfers the payment price is fixed. */
const PAYMENT_AMOUNT: number = 10000000000;

/** Exposes all Casper functionality. */
export class CasperWallet implements Wallet {
    private readonly casper: CasperClient = new CasperClient();

    private readonly PAYMENT_AMOUNT: number = PAYMENT_AMOUNT;
    /** Defines Casper contract hash. */
    private readonly CONTRACT_HASH: string = parseMetaTag(META_TAGS_CONFIG.CASPER_BRIDGE_CONTRACT);
    /** Defines RPC node address. */
    private readonly RPC_NODE_ADDRESS: string = parseMetaTag(META_TAGS_CONFIG.CASPER_NODE_ADDRESS);

    constructor(
        public provider = window.casperlabsHelper,
    ) { };

    /** Checks if the site is connected to Casper extension.
     * @returns {boolean}
     */
    public async isSiteConnected(): Promise<boolean> {
        return await this.provider.isConnected();
    };

    /** Requests Casper public key hex (wallet address).
     * @returns {string} - active Casper key hex.
    */
    public async address(): Promise<string> {
        return await this.provider.getActivePublicKey();
    };

    /** Gerenates casper public key from hex string.
     * @returns {CLPublicKey} - active Casper public key.
     */
    private async getPublicKey(): Promise<CLPublicKey> {
        const publicKeyHex = await this.address();
        return await CLPublicKey.fromHex(publicKeyHex);
    };

    /** Converts contract hash string to byte array.
     * @param {string} contractHash
     * @returns {string} - contract address hash in byte array type.
     */
    private contractHashToByteArray = (contractHash: string) =>
        Uint8Array.from(Buffer.from(contractHash, 'hex'));

    /** Represents a collection of arguments passed to a smart contract.
     * @param {string} amount - transaction amount.
     * @param {string} destination - receiver wallet address.
     * @returns {RuntimeArgs} - arguments collection.
     */
    private async getRuntimeArgs(amount: string, destination: string): Promise<RuntimeArgs> {
        return await RuntimeArgs.fromMap({
            [CasperRuntimeArgs.TOKEN_CONTRACT]: CLValueBuilder.byteArray(this.contractHashToByteArray(parseMetaTag(META_TAGS_CONFIG.CASPER_TOKEN_CONTRACT))),
            [CasperRuntimeArgs.AMOUNT]: CLValueBuilder.u256(amount),
            [CasperRuntimeArgs.DESTINATION_CHAIN]: CLValueBuilder.string(NetworkNames.GOERLI),
            [CasperRuntimeArgs.DESTINATION_ADDRESS]: CLValueBuilder.string(destination),
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
        const contractHashAsByteArray = this.contractHashToByteArray(this.CONTRACT_HASH);
        const deploy = await this.makeDeploy(
            NetworkNames.CASPER_TEST.toLowerCase(),
            contractHashAsByteArray,
            entryPoint,
            runtimeArgs,
            this.PAYMENT_AMOUNT,
        );
        const json = DeployUtil.deployToJson(deploy);
        const publicKeyHex = await this.address();
        const signature = await this.provider.sign(json, publicKeyHex);

        return signature;
    };

    /** Signs authenticated message and returns signature.
     * @param {string} message - authenticated message.
     * @returns {string} - signed signature.
     */
    public async sign(message: string): Promise<string> {
        const publicKey = await this.address();
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
        const runtimeArgs: RuntimeArgs = await this.getRuntimeArgs(amount, destination);
        const deploy = await this.contractCall(CasperEntryPoints.SEND_TRANSACTION, runtimeArgs);
        await this.casper.sendTransaction(JSON.stringify(deploy), this.RPC_NODE_ADDRESS);
    };
};
