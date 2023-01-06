import { APIClient } from '@/api';
import { BridgeInSignatureResponse, SignatureRequest, Transfer, TransferEstimate, TransferEstimateRequest, TransferPagination, TransfersHistory } from '@/transfers';

/**
 * TransfersClient is a http implementation of transfers API.
 * Exposes all transfers-related functionality.
 */
export class TransfersClient extends APIClient {
    /** Requests transfers list by signature. */
    public async history(transferPagination: TransferPagination): Promise<TransfersHistory> {
        const response = await this.http.get(`${this.ROOT_PATH}/transfers/history/${transferPagination.signature}/${transferPagination.pubKey}?network-id=${transferPagination.networkId}&offset=${transferPagination.offset}&limit=${transferPagination.limit}`);
        if (!response.ok) {
            await this.handleError(response);
        }

        const history = await response.json();
        const transfers = history.transfers || [];

        return new TransfersHistory(
            history.limit,
            history.offset,
            history.totalCount,
            transfers.map((transfer: Transfer) =>
                new Transfer(
                    transfer.amount,
                    transfer.createdAt,
                    transfer.id,
                    transfer.outboundTx,
                    transfer.recipient,
                    transfer.sender,
                    transfer.status,
                    transfer.triggeringTx,
                )
            )
        );
    };

    /** Requests transfer estimate.
     * @param {TransferEstimateRequest} transferEstimateRequest - Params to estimate tranfer token
     * @returns {TransferEstimate} - Returns fee, fee percentage, estimated confirmation time for transfer token
     */
    public async estimate(transferEstimateRequest: TransferEstimateRequest): Promise<TransferEstimate> {
        const response = await this.http.get(`${this.ROOT_PATH}/transfers/estimate/${transferEstimateRequest.senderNetwork}/${transferEstimateRequest.recipientNetwork}/${transferEstimateRequest.tokenId}/${transferEstimateRequest.amount}`);
        if (!response.ok) {
            await this.handleError(response);
        }

        const transferEstimate = await response.json();

        return new TransferEstimate(
            transferEstimate.fee,
            transferEstimate.feePercentage,
            transferEstimate.estimatedConfirmationTime,
        );
    };

    /** Canceles transfer.
     * @param {number} transferId
     * @param {string} signature
     * @param {string} pubKey - Public key hex for canceles transaction (needed only for Casper transaction)
     */
    public async cancel(transferId: number, signature: string, pubKey: string): Promise<void> {
        const response = await this.http.delete(`${this.ROOT_PATH}/transfers/${transferId}/${signature}/${pubKey}`);
        if (!response.ok) {
            await this.handleError(response);
        }
    };

    /** Requests transfers signature.
     * @param {SignatureRequest} signatureRequest - holds information to request transfer signature.
     * @returns {BridgeInSignatureResponse} - values needed to send bridge in transaction.
     */
    public async signature(signatureRequest: SignatureRequest): Promise<BridgeInSignatureResponse> {
        const response = await this.http.post(`${this.ROOT_PATH}/transfers/bridge-in-signature`, JSON.stringify(signatureRequest));
        if (!response.ok) {
            await this.handleError(response);
        }
        const signatureResponse = await response.json();

        return new BridgeInSignatureResponse(
            signatureResponse.token,
            signatureResponse.amount,
            signatureResponse.gasComission,
            signatureResponse.destination,
            signatureResponse.deadline,
            signatureResponse.nonce,
            signatureResponse.signature,
        );
    };
};
