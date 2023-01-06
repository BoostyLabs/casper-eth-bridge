import { TransfersClient } from '@/api/transfers';
import { BridgeInSignatureResponse, SignatureRequest, TransferEstimate, TransferEstimateRequest, TransferPagination, TransfersHistory } from '@/transfers';

/**
 * Exposes all transfers domain entities related logic.
 */
export class TransfersService {
    protected readonly transfers: TransfersClient;

    public constructor(transfers: TransfersClient) {
        this.transfers = transfers;
    };

    /** Requests transfer estimate. */
    public async estimate(transferEstimateRequest: TransferEstimateRequest): Promise<TransferEstimate> {
        return await this.transfers.estimate(transferEstimateRequest);
    };

    /** Canceles transfer. */
    public async cancel(transferId: number, signature: string, pubKey: string): Promise<void> {
        await this.transfers.cancel(transferId, signature, pubKey);
    };

    /** Requests list of transfers by signature. */
    public async history(transferPagination: TransferPagination): Promise<TransfersHistory> {
        return await this.transfers.history(transferPagination);
    };

    /** Requests transfers signature. */
    public async signature(signatureRequest: SignatureRequest): Promise<BridgeInSignatureResponse> {
        return await this.transfers.signature(signatureRequest);
    };
};
