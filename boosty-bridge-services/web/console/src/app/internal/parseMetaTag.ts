import appConfig from '@/app/configs/appConfig.json';

/** Describes casper and eth related config data . */
export enum META_TAGS_CONFIG {
    CASPER_BRIDGE_CONTRACT = 'casper-bridge-contract',
    CASPER_NODE_ADDRESS = 'casper-node-address',
    CASPER_TOKEN_CONTRACT = 'casper-token-contract',
    ETH_BRIDGE_CONTRACT = 'eth-bridge-contract',
    ETH_GAS_LIMIT = 'eth-gas-limit',
    ETH_TOKEN_CONTRACT = 'eth-token-contract',
    GATEWAY_ADDRESS = 'gateway-address',
};

/** Parses HTML meta tag and returns content. */
export function parseMetaTag(metaTagName: string): string {
    const metas = document.getElementsByTagName('meta');
    for (let i = appConfig.numbers.ZERO_NUMBER; i < metas.length; i++) {
        if (metas[i].getAttribute('name') === metaTagName) {
            return metas[i].content;
        }
    }
    return '';
}
