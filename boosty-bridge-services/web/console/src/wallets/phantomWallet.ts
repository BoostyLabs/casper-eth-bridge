import { Buffer } from 'buffer';

import { Wallet } from '@/wallets';

/** Exposes all Phantom functionality. */
export class PhantomWallet implements Wallet {
    constructor(
        // @ts-ignore
        public provider = window.phantom?.solana,
    ) { };

    /** Requests Phantom wallet public key.
     * @returns {string} - Phantom public key.
    */
    public async address(): Promise<string> {
        const connectedAccount = await this.provider.connect();
        return connectedAccount.publicKey.toString();
    };

    /** Signs message and creates message raw signature.
     * @param {string} message - Authenticated message.
     * @returns {string} - Signed signature.
     */
    public async sign(message: string): Promise<string> {
        const encodedMessage = new TextEncoder().encode(message);
        const signedMessage = await this.provider.signMessage(encodedMessage);
        return Buffer.from(signedMessage.signature).toString('hex');
    };

    /** Requests connected phantom wallet public keys only for Phantom provider. */
    public async connect(): Promise<void> {
        await this.provider.connect();
    };

    /** Sends transaction.
    * @param {string} receiver - wallet address to receive transaction token amount
    */
    public async sendTransaction(receiver: string, amount: string): Promise<void> {
        // TODO: Will be added after backend implementation.
    };
};
