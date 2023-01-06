// SPDX-License-Identifier: MIT

pragma solidity 0.8.17;

library Errors {
    error InvalidSignature();
    error AlreadyUsedSignature();
    error ExpiredSignature();
    error AmountExceedCommissionPool();
    error AmountExceedBridgePool();
}
