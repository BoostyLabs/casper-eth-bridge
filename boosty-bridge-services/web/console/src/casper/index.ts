/** Describes arguments needed to deploy contract. */
export enum CasperRuntimeArgs {
    AMOUNT = 'amount',
    DESTINATION_ADDRESS = 'destination_address',
    DESTINATION_CHAIN = 'destination_chain',
    TOKEN_CONTRACT = 'token_contract',
};

/** Describes casper contract entry points needed to interact with app. */
export enum CasperEntryPoints {
    SEND_TRANSACTION = 'bridge_in',
};
