import { Dispatch } from 'redux';

import { NetworksClient } from '@/api/networks';
import { Network, Token } from '@/networks';
import { NetworksService } from '@/networks/service';

export const SET_CONNECTED_NETWORKS: string = 'SET_CONNECTED_NETWORKS';
export const SET_SUPPORTED_TOKENS: string = 'SET_SUPPORTED_TOKENS';
export const SET_ACTIVE_SUPPORTED_TOKEN: string = 'SET_ACTIVE_SUPPORTED_TOKEN';

/** An action setConnectedNetworks contains type and payload data for sets connected networks. */
export const setConnectedNetworks = (networks: Network[]) => ({
    type: SET_CONNECTED_NETWORKS,
    payload: networks,
});

/** An action setSupportedTokens contains type and payload data for sets supported tokens. */
export const setSupportedTokens = (tokens: Token[]) => ({
    type: SET_SUPPORTED_TOKENS,
    payload: tokens,
});

/** An action setActiveSupportedToken contains type and payload data for sets supported token. */
export const setActiveSupportedToken = (token: Token) => ({
    type: SET_ACTIVE_SUPPORTED_TOKEN,
    payload: token,
});

const networksClient = new NetworksClient();
const networksService = new NetworksService(networksClient);

/** Thunk middleware that requests connected networks list and sets into reducer. */
export const getConnectedNetworks = () => async function(dispatch: Dispatch) {
    const connectedNetworks = await networksService.connected();
    dispatch(setConnectedNetworks(connectedNetworks));
};

/** Thunk middleware that requests supported tokens list and sets into reducer. */
export const getSupportedTokens = (networkId: number) => async function(dispatch: Dispatch) {
    const supportedTokens = await networksService.supportedTokens(networkId);
    dispatch(setSupportedTokens(supportedTokens));
};
