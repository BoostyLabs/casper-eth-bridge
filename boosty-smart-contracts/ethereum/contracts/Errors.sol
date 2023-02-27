// SPDX-License-Identifier: Unlicense

pragma solidity 0.8.17;

library Errors {
    string public constant COMMISSION_GREATER_THAN_AMOUNT =
        "CommissionGreaterThanAmount";
    string public constant ALREADY_USED_SIGNATURE =
        "AlreadyUsedSignature";
    string public constant EXPIRED_SIGNATURE =
        "ExpiredSignature";
    string public constant AMOUNT_EXCEED_BRIDGE_POOL =
        "AmountExceedBridgePool";
    string public constant AMOUNT_EXCEED_COMMISSION_POOL = 
        "AmountExceedCommissionPool";
}
