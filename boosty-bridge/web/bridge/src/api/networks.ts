import { APIClient } from '@/api';
import { Network, Token, WrappedIn } from '@/networks';

/**
 * NetworksClient is a http implementation of networks API.
 * Exposes all networks-related functionality.
 */
export class NetworksClient extends APIClient {
    /** Requests connected networks list.
     * @returns {Network[]} - Connected networks list
    */
    public async connected(): Promise<Network[]> {
        const response = await this.http.get(`${this.ROOT_PATH}/networks`);
        if (!response.ok) {
            await this.handleError(response);
        }

        const networks = await response.json() || [];

        return networks.map((network: Network) =>
            new Network(
                network.id,
                network.name,
                network.type,
                network.isTestnet,
            )
        );
    };

    /** Requests supported tokens list.
     * @param {number} networkId - Active network ID to send transaction
     * @returns {Token[]} - Supported tokens list for active network
     */
    public async supportedTokens(networkId: number): Promise<Token[]> {
        const response = await this.http.get(`${this.ROOT_PATH}/networks/${networkId}/supported-tokens`);
        if (!response.ok) {
            await this.handleError(response);
        }

        const tokens = await response.json() || [];

        return tokens.map((token: Token) =>
            new Token(
                token.id,
                token.shortName,
                token.longName,
                tokens.wrapp ? token.wraps.map((wrap: WrappedIn) =>
                    new WrappedIn(wrap.networkId, wrap.smartContractAddress)
                ) : []
            )
        );
    };
};
