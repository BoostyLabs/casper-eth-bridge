import { NetworksClient } from '@/api/networks';
import { Network, Token } from '@/networks';

/**
 * Exposes all networks domain entities related logic.
 */
export class NetworksService {
    protected readonly networks: NetworksClient;

    public constructor(networks: NetworksClient) {
        this.networks = networks;
    };

    /** Requests connected networks list. */
    public async connected(): Promise<Network[]> {
        return await this.networks.connected();
    };

    /** Requests supported tokens list. */
    public async supportedTokens(networkId: number): Promise<Token[]> {
        return await this.networks.supportedTokens(networkId);
    };
};
