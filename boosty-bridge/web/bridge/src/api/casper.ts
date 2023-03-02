import { APIClient } from '@/api';

/**
 * CasperClient is a http implementation of casper API.
 */
export class CasperClient extends APIClient {
    /** Sends transaction via Casper wallet.
     * @param {string} deploy - JSON representation of a deploy - can be constructed using the `DeployUtil.deployToJSON()` method.
     * @param {string} rpcNodeAddress - RPC node address
    */
    public async sendTransaction(deploy: string, rpcNodeAddress: string): Promise<void> {
        const path = '/bridge-in';
        const response = await this.http.post(path, JSON.stringify({ deploy, rpcNodeAddress }));
        if (!response.ok) {
            await this.handleError(response);
        }
    };

    /** Canceles transaction via Casper wallet.
     * @param {string} deploy - JSON representation of a deploy - can be constructed using the `DeployUtil.deployToJSON()` method.
     * @param {string} rpcNodeAddress - RPC node address
     */
    public async cancelTransction(deploy: string, rpcNodeAddress: string): Promise<void> {
        const path = '/transfer-out';
        const response = await this.http.post(path, JSON.stringify({ deploy, rpcNodeAddress }));
        if (!response.ok) {
            await this.handleError(response);
        }
    };
};
