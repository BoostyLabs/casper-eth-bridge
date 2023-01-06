import appConfig from '@/app/configs/appConfig.json';
import { NetworkNames } from '@/networks';

/** Defines available transfers statuses. */
export enum TransferStatuses {
    'UNSPECIFIED',
    'CONFIRMING',
    'CANCELED',
    'FINISHED',
    'WAITING',
};

export class NetworkAddress {
    constructor(
        public address: string = '',
        public networkName: NetworkNames = NetworkNames.CASPER_TEST,
    ) { };
};

export class StringTxHash {
    constructor(
        public hash: string = '',
        public networkName: NetworkNames = NetworkNames.GOERLI,
    ) { };
};

/** Holds transfer domain entity about transferring funds from one network to another. */
export class Transfer {
    constructor(
        public amount: string = '',
        public createdAt: string = '',
        public id: number = appConfig.numbers.ZERO_NUMBER,
        public outboundTx: StringTxHash = new StringTxHash(),
        public recipient: NetworkAddress = new NetworkAddress(),
        public sender: NetworkAddress = new NetworkAddress(),
        public status: TransferStatuses = TransferStatuses.FINISHED,
        public triggeringTx: StringTxHash = new StringTxHash(),
    ) { };
};

/** Defines request fields to get transfer estimate. */
export class TransferEstimateRequest {
    constructor(
        public senderNetwork: string = '',
        public recipientNetwork: string = '',
        public tokenId: number = appConfig.numbers.ZERO_NUMBER,
        public amount: string = '',
    ) { };
};

/** Holds approximate information about transfer fee and time. */
export class TransferEstimate {
    constructor(
        public fee: string = '',
        public feePercentage: string = '',
        public estimatedConfirmationTime: string = '',
    ) { };
};

export class TransferPagination {
    constructor(
        public pubKey: string = '',
        public signature: string = '',
        public networkId: number = appConfig.numbers.ONE_NUMBER,
        public offset: number = appConfig.numbers.ZERO_NUMBER,
        public limit: number = appConfig.numbers.FIVE_NUMBER,
    ) {};
};

export class TransfersHistory {
    constructor(
        public limit: number = appConfig.numbers.FIVE_NUMBER,
        public offset: number = appConfig.numbers.ZERO_NUMBER,
        public totalCount: number = appConfig.numbers.ZERO_NUMBER,
        public transfers: Transfer[] = [],
    ) {};
};

/** Holds information to request transfer signature. */
export class SignatureRequest {
    constructor(
        public sender: NetworkAddress = new NetworkAddress(),
        public tokenId: number = appConfig.numbers.ONE_NUMBER,
        public amount: string = '',
        public destination: NetworkAddress = new NetworkAddress(),
    ) {};
};

/** BridgeInSignatureResponse holds the values needed to send bridge in transaction. */
export class BridgeInSignatureResponse {
    constructor(
        public token: string = '',
        public amount: string = '',
        public gasComission: string = '',
        public destination: NetworkAddress = new NetworkAddress(),
        public deadline: string = '',
        public nonce: number = appConfig.numbers.ONE_NUMBER,
        public signature: string = '',
    ) {};
};
