/** WalletAddressValidator describes address vaidation to which the transaction will be sent. */
export class WalletAddressValidator {
    /** Validates ETH account address.
     * @param {string} address - ETH account address.
     * @returns {boolean}
    */
    static isEthAddressValid(address: string): boolean {
        const re = new RegExp(/^0x[a-fA-F0-9]{40}$/, 'i');

        return re.test(String(address).toLowerCase());
    };

    /** Validates Casper account hash.
     * @param {string} accountHash - Casper account hash.
     * @returns {boolean}
    */
    static isCasperAccountHashValid(accountHash: string): boolean {
        const re = new RegExp(/^[a-z0-9]{64}$/, 'i');

        return re.test(String(accountHash).toLowerCase());
    };
};
