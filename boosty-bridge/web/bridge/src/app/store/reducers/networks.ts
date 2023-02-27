import { Network, Token } from '@/networks';
import { SET_ACTIVE_SUPPORTED_TOKEN, SET_CONNECTED_NETWORKS, SET_SUPPORTED_TOKENS } from '@/app/store/actions/networks';

/** Exposes networks state. Uses as default state for reducer. */
class NetworksState {
    constructor(
        public networks: Network[] = [],
        public supportedTokens: Token[] = [],
        public activeSupportedToken: Token = new Token(),
    ) { };
};

/** NetworksReducerAction uses as action payload for reducer. */
class NetworksReducerAction {
    constructor(
        public type: string = '',
        public payload: any = '',
    ) { }
};

export const networksReducer = (
    networksState: NetworksState = new NetworksState(),
    action: NetworksReducerAction = new NetworksReducerAction(),
) => {
    switch (action.type) {
    case SET_CONNECTED_NETWORKS:
        networksState.networks = action.payload;
        break;
    case SET_SUPPORTED_TOKENS:
        networksState.supportedTokens = action.payload;
        break;
    case SET_ACTIVE_SUPPORTED_TOKEN:
        networksState.activeSupportedToken = action.payload;
        break;
    default:
        return networksState;
    };

    return { ...networksState };
};
