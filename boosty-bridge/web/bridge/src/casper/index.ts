/** Describes arguments needed to deploy contract. */
export enum CasperRuntimeArgs {
    AMOUNT = 'amount',
    GAS_COMMISSION = 'gas_commission',
    COMMISSION = 'commission',
    DEADLINE = 'deadline',
    DESTINATION_ADDRESS = 'destination_address',
    DESTINATION_CHAIN = 'destination_chain',
    NONCE = 'nonce',
    TOKEN_CONTRACT = 'token_contract',
    SIGNATURE = 'signature',
    RECIPIENT = 'recipient',
};

/** Describes casper contract entry points needed to interact with app. */
export enum CasperEntryPoints {
    SEND_TRANSACTION = 'bridge_in',
    CANCEL_TRANSACTION = 'transfer_out',
};
