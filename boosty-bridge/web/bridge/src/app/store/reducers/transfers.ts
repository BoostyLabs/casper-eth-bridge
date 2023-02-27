import { Transfer, TransferEstimate, TransfersHistory } from '@/transfers';
import { SET_HISTORY, SET_TRANSFER_ESTIMATE } from '@/app/store/actions/transfers';

/** Exposes transfers state. Uses as default state for reducer. */
class TransfersState {
    constructor(
        public history: TransfersHistory = new TransfersHistory(),
        public transferEstimate: TransferEstimate = new TransferEstimate(),
    ) { };
};

/** TransfersReducerAction uses as action payload for reducer. */
class TransfersReducerAction {
    constructor(
        public type: string = '',
        public payload: any = '',
    ) { };
};

export const transfersReducer = (
    transfersState: TransfersState = new TransfersState(),
    action: TransfersReducerAction = new TransfersReducerAction(),
) => {
    switch (action.type) {
    case SET_HISTORY:
        transfersState.history = action.payload;
        break;
    case SET_TRANSFER_ESTIMATE:
        transfersState.transferEstimate = action.payload;
        break;
    default:
        return transfersState;
    };

    return { ...transfersState };
};
