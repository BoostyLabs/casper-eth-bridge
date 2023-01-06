import { Wallet } from '@/wallets';

/**
 * Exposes all wallets-related functionality.
 */
export class WalletsService {
    constructor(
        public wallet: Wallet,
    ) { };

    /** Requests connected to blockchain node wallet address.
    * @returns {string} - connected wallet address.
    */
    public async address(): Promise<string> {
        return await this.wallet.address();
    };

    /** Sends transaction.
    * @param {string} receiver - wallet address to receive transaction token amount.
    */
    public async sendTransaction(receiver: string, amount: string): Promise<void> {
        await this.wallet.sendTransaction(receiver, amount);
    };

    /** Requests connection to blockchain node. */
    public async connect(): Promise<void> {
        return await this.wallet.connect();
    };

    /** Signs message and creates message raw signature.
    * @param {message} - bridge authentication message to sign.
    * @returns {string} - signed raw signature.
    */
    public async sign(message: string): Promise<string> {
        return await this.wallet.sign(message);
    };
};
