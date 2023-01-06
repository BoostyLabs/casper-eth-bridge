import { describe, expect, it } from '@jest/globals';

import { WalletAddressValidator } from "@app/internal/walletAddressValidator";

describe('Validates ETH wallet addresses and Casper account hashes.', () => {
    const VALID_ETH_WALLET_ADDRESSES: string[] = [
        '0x3095F955Da700b96215CFfC9Bc64AB2e69eB7DAB',
        '0x0d7dB08D75679bbE90b7B1eDfB8BE3a16897Ee77',
    ];

    const NOT_VALID_ETH_WALLET_ADDRESSES: string[] = [
        '0x3095F955Da700b96215CFfC9Bc64AB2e69eB7D',
        '0x0d7dB08D75679bbE90b7B1eDfB8BE3a16897Ee772',
        '',
        '110a71C12D57429bkE30c231cJfU1BE0a14887OeP1',
        '0x0a71C12D57429bkE30c231cJfC1BE0a14887OeR1'
    ];

    const VALID_CASPER_ACCOUNT_HASHES: string[] = [
        'daa2b596e0a496b04933e241e0567f2bcbecc829aa57d88cab096c28fd07dee2',
        'eacsd123eo1230lzx0982abc12lkxqla2bec12398123d0192ccka029123adc12',
    ];

    const NOT_VALID_CASPER_ACCOUNT_HASHES: string[] = [
        'daa2b596e0a496b04933e241e0567f2bcbecc829aa57d88cab096c28fd07de',
        '0daa2b596e0a496b04933e241e0567f2bcbecc829aa57d88cab096c28fd07de',
        'eacsd123eo1230lzx0982abc12lkxqla2bec12398123d0192ccka029123adc12ads',
    ];

    it('Should be valid ETH wallet addresses.', () => {
        VALID_ETH_WALLET_ADDRESSES.forEach((walletAddress: string) => {
            expect(WalletAddressValidator.isEthAddressValid(walletAddress)).toBe(true);
        });
    });

    it('Should be not valid ETH wallet addresses.', () => {
        NOT_VALID_ETH_WALLET_ADDRESSES.forEach((walletAddress: string) => {
            expect(WalletAddressValidator.isEthAddressValid(walletAddress)).toBe(false);
        });
    });

    it('Should be valid CASPER account hashes.', () => {
        VALID_CASPER_ACCOUNT_HASHES.forEach((accountHash: string) => {
            expect(WalletAddressValidator.isCasperAccountHashValid(accountHash)).toBe(true);
        });
    });

    it('Should be not valid CASPER account hashes.', () => {
        NOT_VALID_CASPER_ACCOUNT_HASHES.forEach((accountHash: string) => {
            expect(WalletAddressValidator.isCasperAccountHashValid(accountHash)).toBe(false);
        });
    });
});
