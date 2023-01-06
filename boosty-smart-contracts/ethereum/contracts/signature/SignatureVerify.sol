// SPDX-License-Identifier: MIT
// vvv do we need to return comission to the user?
pragma solidity 0.8.17;

import "@openzeppelin/contracts/utils/cryptography/ECDSA.sol";
import "../Errors.sol";

contract SignatureVerify {
    address private _signerAddress;

    constructor(address systemAddress_) {
        _signerAddress = systemAddress_;
    }

    function _checkBridgeInRequest(
        address senderAddress,
        address token,
        uint256 amount,
        uint256 commission,
        string memory destinationChain,
        string memory destinationAddress,
        uint256 deadline,
        uint256 nonce,
        bytes calldata signature
    ) internal view {
        if (
            !_verify(
                _signerAddress,
                _hashBridgeIn(
                    senderAddress, 
                    token,
                    amount, 
                    commission, 
                    destinationChain, 
                    destinationAddress, 
                    deadline, 
                    nonce
                ),
                signature
            )
        ) {
            revert Errors.InvalidSignature();
        }
    }

    function _checkTransferOutRequest(
        address token,
        address recipient,
        uint256 amount,
        uint256 commission,
        uint256 nonce,
        bytes calldata signature
    ) internal view {
        if (
            !_verify(
                _signerAddress,
                _hashTransferOut(token, recipient, amount, commission, nonce),
                signature
            )
        ) {
            revert Errors.InvalidSignature();
        }
    }

    function _verify(
        address singerAddress,
        bytes32 hash,
        bytes calldata signature
    ) private pure returns (bool) {
        return singerAddress == ECDSA.recover(hash, signature);
    }

    function _hashBridgeIn(
        address senderAddress,
        address token,
        uint256 amount,
        uint256 commission,
        string memory destinationChain,
        string memory destinationAddress,
        uint256 deadline,
        uint256 nonce
    ) private pure returns (bytes32) {
        return
            ECDSA.toEthSignedMessageHash(
                keccak256(
                    abi.encodePacked(
                        senderAddress, 
                        token,
                        amount, 
                        commission, 
                        destinationChain, 
                        destinationAddress, 
                        deadline, 
                        nonce
                    )
                )
            );
    }

    function _hashTransferOut(
        address token,
        address recipient,
        uint256 amount,
        uint256 commission,
        uint256 nonce
    ) private pure returns (bytes32) {
        return
            ECDSA.toEthSignedMessageHash(
                keccak256(abi.encodePacked(token, recipient, amount, commission, nonce))
            );
    }
}
