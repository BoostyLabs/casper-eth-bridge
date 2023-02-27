import appConfig from '@/app/configs/appConfig.json';

/** Defines list of possible blockchain network interoperability types. */
export enum NetworkTypes {
    CASPER = 'NT_CASPER',
    EVM = 'NT_EVM',
};

/** Defines available networks */
export enum NetworkNames {
    CASPER_TEST = 'CASPER-TEST',
    GOERLI = 'GOERLI',
    MUMBAI = 'MUMBAI',
    BNB_TEST = 'BNB-TEST',
    AVALANCHE_TEST = 'AVALANCHE-TEST',
};

// TODO: delete after API implementation.
/** Defines all available EVM networks chains. */
export enum NetworksChains {
    GOERLI = '0x5',
    MUMBAI = '0x13881',
    BNB_TEST = '0x61',
    AVALANCHE_TEST = '0xA869'
};

/** Holds basic network characteristics. */
export class Network {
    constructor(
        public id: number = appConfig.numbers.ZERO_NUMBER,
        public name: NetworkNames = NetworkNames.GOERLI,
        public type: NetworkTypes = NetworkTypes.EVM,
        public isTestnet: boolean = true,
    ) { };
};

/** Holds wrapped version of the Token. */
export class WrappedIn {
    constructor(
        public networkId: string = '',
        public smartContractAddress: string = '',
    ) { };
};

/** Holds information about supported by golden-gate tokens. */
export class Token {
    constructor(
        public id: number = appConfig.numbers.ONE_NUMBER,
        public shortName: string = '',
        public longName: string = '',
        public wraps: WrappedIn[] = [],
    ) { };
};

export enum Networks {
    'CASPER',
    'EVM',
};
