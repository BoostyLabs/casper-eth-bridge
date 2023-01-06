import { CasperLabsHelper } from 'casper-js-sdk/dist/@types/casperlabsSigner';

import { EVMProvider } from '@/ethers';
import { SolanaProvider } from '@/phantom';

/** Defines web3 provider to communicate with Blockchain Node via JSON-RPC. */
type Provider = typeof EVMProvider | CasperLabsHelper | typeof SolanaProvider;

// TODO: could be extended or changed depend on new providers.
export interface Wallet {
    provider: Provider;
    address: () => Promise<string>;
    sign: (message: string) => Promise<string>;
    connect: () => Promise<void>;
    sendTransaction: (receiver: string, amount: string) => Promise<void>;
};
