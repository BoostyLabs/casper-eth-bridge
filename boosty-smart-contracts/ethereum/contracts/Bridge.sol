// SPDX-License-Identifier: UNLICENSED

pragma solidity 0.8.17;

import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import "@openzeppelin/contracts/security/Pausable.sol";
import "./signature/SignatureVerify.sol";

contract Bridge is SignatureVerify, Ownable, Pausable {
    using SafeERC20 for IERC20;

    uint16 public constant HUNDRED_PERCENT = 10_000;
    uint256 private _stableCommissionPercent = 4_00;

    mapping(uint256 => bool) private _usedNonces;
    mapping(address => uint256) private _commissionPools;

    event BridgeFundsIn(
        address indexed sender,
        uint256 indexed nonce,
        address token,
        uint256 amount,
        uint256 stableCommissionPercent,
        uint256 gasCommission,
        string destinationChain,
        string destinationAddress
    );

    event BridgeFundsOut(
        address indexed recipient,
        address token,
        uint256 amount,
        uint256 transactionId,
        string sourceChain,
        string sourceAddress
    );

    event TransferOut(
        address indexed recipient,
        uint256 indexed nonce,
        address token,
        uint256 amount
    );

    event WithdrawCommission (
        address indexed token,
        uint256 amount
    );

    constructor(address signer) SignatureVerify(signer) {}

    /// @notice Deposit tokens to the bridge to transfer it to another chain
    /// @param token Token address
    /// @param amount Token amount
    /// @param gasCommission Commission which is calculated in transferred token.
    /// @param destinationChain Chain where we transfer tokens
    /// @param destinationAddress Address where we transfer tokens on the chain mentioned above
    /// @param deadline Timestamp until transaction is valid
    /// @param nonce Parameter to avoid double transaction attack
    /// @param signature Classic signature
    function bridgeIn(
        address token, // vvvcheck
        uint256 amount, // vvvcheck
        uint256 gasCommission, // 200
        string calldata destinationChain, // vvvcheck
        string calldata destinationAddress,
        uint256 deadline, 
        uint256 nonce,
        bytes calldata signature
    ) external whenNotPaused {

        if (_usedNonces[nonce]) {
            revert Errors.AlreadyUsedSignature();
        }

        if (block.timestamp > deadline) {
            revert Errors.ExpiredSignature();
        }

        _checkBridgeInRequest(
            _msgSender(),
            token,
            amount,
            gasCommission,
            destinationChain,
            destinationAddress,
            deadline,
            nonce,
            signature
        );

        _usedNonces[nonce] = true;

        uint256 totalCommission = getTotalCommission(amount, gasCommission);
       _commissionPools[token] += totalCommission;


        IERC20(token).safeTransferFrom(_msgSender(), address(this), amount);

        emit BridgeFundsIn(
            _msgSender(),
            nonce,
            token,
            amount,
            _stableCommissionPercent,
            gasCommission,
            destinationChain,
            destinationAddress
        );
    }

    /// @notice Withdraw tokens from the bridge. Can be initiated only by the owner
    /// @param token Token address
    /// @param recipient Recipient address
    /// @param amount Token amount
    /// @param sourceChain From which chain we transfer to the recipient
    /// @param sourceAddress From which address we transfer to the recipient
    function bridgeOut(
        address token,
        address recipient,
        uint256 amount,
        uint256 transactionId,
        string calldata sourceChain,
        string calldata sourceAddress
    ) external onlyOwner {
        uint256 balance = IERC20(token).balanceOf(address(this));
        uint256 allowedBalance = balance - _commissionPools[token];
        if (amount > allowedBalance) {
            revert Errors.AmountExceedBridgePool();
        }
        IERC20(token).safeTransfer(recipient, amount);
        emit BridgeFundsOut(
            recipient,
            token,
            amount,
            transactionId,
            sourceChain,
            sourceAddress
        );
    }

    /// @notice Withdraw tokens
    /// @param token Token address
    /// @param amount Token amount
    function withdrawCommission(
        address token,
        uint256 amount
    ) external onlyOwner {
        if (_commissionPools[token] < amount) {
            revert Errors.AmountExceedCommissionPool();
        }
        IERC20(token).safeTransfer(msg.sender, amount);
        emit WithdrawCommission(token, amount);
    }

    /// @notice Allow user to withdraw tokens back if the backend approved it by providing signature
    /// @param token Token address
    /// @param recipient Recipient address
    /// @param amount Token amount
    /// @param nonce Parameter to avoid double transaction attack
    /// @param signature Classic signature
    function transferOut(
        address token,
        address recipient,
        uint256 amount,
        uint256 commission,
        uint256 nonce,
        bytes calldata signature
    ) external whenNotPaused {
        if (_usedNonces[nonce]) {
            revert Errors.AlreadyUsedSignature();
        }

        _checkTransferOutRequest(token, recipient, amount, commission, nonce, signature);

        _usedNonces[nonce] = true;
        _commissionPools[token] -= commission;

        uint256 totalSumForTransfer = amount + commission;
        IERC20(token).safeTransfer(recipient,totalSumForTransfer);

        // TODO:Descrease pool commission
        emit TransferOut(recipient, nonce, token, totalSumForTransfer);
    }

    /// @notice Set stable commission percent. Allowed only for onwer
    /// @param stableCommissionPercent_ percent
    function setStableCommissionPercent(
        uint256 stableCommissionPercent_
    ) external onlyOwner {
        _stableCommissionPercent = stableCommissionPercent_;
    }

    /// @notice Stop all contract functionality allowed to the user
    function pause() external onlyOwner {
        _pause();
    }

    /// @notice Resume all contract functionality allowed to the user
    function unpause() external onlyOwner {
        _unpause();
    }

    /// @notice Get gasCommission percent. 
    /// @return stable commission percent
    function getStableCommissionPercent() external view returns(uint256) {
        return _stableCommissionPercent;
    }

    /// @notice Get amount in pool. 
    /// @return amount in commission pool
    function getCommissionPoolAmount(address token) external view onlyOwner returns(uint256) {
        return _commissionPools[token];
    }

    /// @notice Claculate total gasCommission
    /// @param amount Token amount
    /// @param gasCommission Manual gasCommission
    function getTotalCommission(
        uint256 amount,
        uint256 gasCommission
    ) public view returns (uint256) {
        uint256 stableCommission = (amount * _stableCommissionPercent) /
            HUNDRED_PERCENT;
        uint256 totalCommission = stableCommission + gasCommission;
        return totalCommission;
    }
}
