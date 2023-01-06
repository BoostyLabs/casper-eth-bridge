import { Dispatch } from 'redux';

import { TransfersClient } from '@/api/transfers';
import { TransferEstimate, TransferEstimateRequest, TransferPagination, TransfersHistory } from '@/transfers';
import { TransfersService } from '@/transfers/service';

export const SET_HISTORY: string = 'SET_HISTORY';
export const SET_TRANSFER_ESTIMATE: string = 'SET_TRANSFER_ESTIMATE';

/** An action setTransfers contains type and payload data for sets transfers list. */
export const setHistory = (history: TransfersHistory) => ({
    type: SET_HISTORY,
    payload: history,
});

/** An action setTransferEstimate contains type and payload data for sets transfer preview. */
export const setTransferEstimate = (transferEstimate: TransferEstimate) => ({
    type: SET_TRANSFER_ESTIMATE,
    payload: transferEstimate,
});

const transfersClient = new TransfersClient();
const transfersService = new TransfersService(transfersClient);

/** Thunk middleware that requests transfers list history and sets into reducer. */
export const getTransfersHistory = (transferPagination: TransferPagination) => async function(dispatch: Dispatch) {
    const history = await transfersService.history(transferPagination);
    dispatch(setHistory(history));
};

/** Thunk middleware that requests transfer estimate. */
export const estimateTransfer = (transferEstimateRequest: TransferEstimateRequest) => async function(dispatch: Dispatch) {
    const transferEstimate = await transfersService.estimate(transferEstimateRequest);
    dispatch(setTransferEstimate(transferEstimate));
};

/** Canceles transfer. */
export const cancelTransfer = async(transferId: number, signature: string, pubKey: string) => await transfersService.cancel(transferId, signature, pubKey);
