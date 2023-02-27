import { ethers } from 'ethers';

// @ts-ignore
export const EVMProvider = window.ethereum && new ethers.providers.Web3Provider(window.ethereum, 'any');

/** Defines supported JSON RPC methods to interact with ethereum node. */
export enum JsonRPCMethods {
    // TODO: could be extended with other methods.
    requestAccounts = 'eth_requestAccounts',
    switchNetwork = 'wallet_switchEthereumChain',
};

/* eslint-disable */
export const ABI = [
	{
		"inputs": [
			{
				"internalType": "address",
				"name": "signer",
				"type": "address"
			}
		],
		"stateMutability": "nonpayable",
		"type": "constructor"
	},
	{
		"inputs": [],
		"name": "AlreadyUsedSignature",
		"type": "error"
	},
	{
		"inputs": [],
		"name": "AmountExceedBridgePool",
		"type": "error"
	},
	{
		"inputs": [],
		"name": "AmountExceedCommissionPool",
		"type": "error"
	},
	{
		"inputs": [],
		"name": "ExpiredSignature",
		"type": "error"
	},
	{
		"inputs": [],
		"name": "InvalidSignature",
		"type": "error"
	},
	{
		"anonymous": false,
		"inputs": [
			{
				"indexed": true,
				"internalType": "address",
				"name": "sender",
				"type": "address"
			},
			{
				"indexed": true,
				"internalType": "uint256",
				"name": "nonce",
				"type": "uint256"
			},
			{
				"indexed": false,
				"internalType": "address",
				"name": "token",
				"type": "address"
			},
			{
				"indexed": false,
				"internalType": "uint256",
				"name": "amount",
				"type": "uint256"
			},
			{
				"indexed": false,
				"internalType": "uint256",
				"name": "stableCommissionPercent",
				"type": "uint256"
			},
			{
				"indexed": false,
				"internalType": "uint256",
				"name": "gasCommission",
				"type": "uint256"
			},
			{
				"indexed": false,
				"internalType": "string",
				"name": "destinationChain",
				"type": "string"
			},
			{
				"indexed": false,
				"internalType": "string",
				"name": "destinationAddress",
				"type": "string"
			}
		],
		"name": "BridgeFundsIn",
		"type": "event"
	},
	{
		"anonymous": false,
		"inputs": [
			{
				"indexed": true,
				"internalType": "address",
				"name": "recipient",
				"type": "address"
			},
			{
				"indexed": false,
				"internalType": "address",
				"name": "token",
				"type": "address"
			},
			{
				"indexed": false,
				"internalType": "uint256",
				"name": "amount",
				"type": "uint256"
			},
			{
				"indexed": false,
				"internalType": "uint256",
				"name": "transactionId",
				"type": "uint256"
			},
			{
				"indexed": false,
				"internalType": "string",
				"name": "sourceChain",
				"type": "string"
			},
			{
				"indexed": false,
				"internalType": "string",
				"name": "sourceAddress",
				"type": "string"
			}
		],
		"name": "BridgeFundsOut",
		"type": "event"
	},
	{
		"anonymous": false,
		"inputs": [
			{
				"indexed": true,
				"internalType": "address",
				"name": "previousOwner",
				"type": "address"
			},
			{
				"indexed": true,
				"internalType": "address",
				"name": "newOwner",
				"type": "address"
			}
		],
		"name": "OwnershipTransferred",
		"type": "event"
	},
	{
		"anonymous": false,
		"inputs": [
			{
				"indexed": false,
				"internalType": "address",
				"name": "account",
				"type": "address"
			}
		],
		"name": "Paused",
		"type": "event"
	},
	{
		"anonymous": false,
		"inputs": [
			{
				"indexed": true,
				"internalType": "address",
				"name": "recipient",
				"type": "address"
			},
			{
				"indexed": true,
				"internalType": "uint256",
				"name": "nonce",
				"type": "uint256"
			},
			{
				"indexed": false,
				"internalType": "address",
				"name": "token",
				"type": "address"
			},
			{
				"indexed": false,
				"internalType": "uint256",
				"name": "amount",
				"type": "uint256"
			}
		],
		"name": "TransferOut",
		"type": "event"
	},
	{
		"anonymous": false,
		"inputs": [
			{
				"indexed": false,
				"internalType": "address",
				"name": "account",
				"type": "address"
			}
		],
		"name": "Unpaused",
		"type": "event"
	},
	{
		"anonymous": false,
		"inputs": [
			{
				"indexed": true,
				"internalType": "address",
				"name": "token",
				"type": "address"
			},
			{
				"indexed": false,
				"internalType": "uint256",
				"name": "amount",
				"type": "uint256"
			}
		],
		"name": "WithdrawCommission",
		"type": "event"
	},
	{
		"inputs": [],
		"name": "HUNDRED_PERCENT",
		"outputs": [
			{
				"internalType": "uint16",
				"name": "",
				"type": "uint16"
			}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [
			{
				"internalType": "address",
				"name": "token",
				"type": "address"
			},
			{
				"internalType": "uint256",
				"name": "amount",
				"type": "uint256"
			},
			{
				"internalType": "uint256",
				"name": "gasCommission",
				"type": "uint256"
			},
			{
				"internalType": "string",
				"name": "destinationChain",
				"type": "string"
			},
			{
				"internalType": "string",
				"name": "destinationAddress",
				"type": "string"
			},
			{
				"internalType": "uint256",
				"name": "deadline",
				"type": "uint256"
			},
			{
				"internalType": "uint256",
				"name": "nonce",
				"type": "uint256"
			},
			{
				"internalType": "bytes",
				"name": "signature",
				"type": "bytes"
			}
		],
		"name": "bridgeIn",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [
			{
				"internalType": "address",
				"name": "token",
				"type": "address"
			},
			{
				"internalType": "address",
				"name": "recipient",
				"type": "address"
			},
			{
				"internalType": "uint256",
				"name": "amount",
				"type": "uint256"
			},
			{
				"internalType": "uint256",
				"name": "transactionId",
				"type": "uint256"
			},
			{
				"internalType": "string",
				"name": "sourceChain",
				"type": "string"
			},
			{
				"internalType": "string",
				"name": "sourceAddress",
				"type": "string"
			}
		],
		"name": "bridgeOut",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [
			{
				"internalType": "address",
				"name": "token",
				"type": "address"
			}
		],
		"name": "getCommissionPoolAmount",
		"outputs": [
			{
				"internalType": "uint256",
				"name": "",
				"type": "uint256"
			}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [],
		"name": "getStableCommissionPercent",
		"outputs": [
			{
				"internalType": "uint256",
				"name": "",
				"type": "uint256"
			}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [
			{
				"internalType": "uint256",
				"name": "amount",
				"type": "uint256"
			},
			{
				"internalType": "uint256",
				"name": "gasCommission",
				"type": "uint256"
			}
		],
		"name": "getTotalCommission",
		"outputs": [
			{
				"internalType": "uint256",
				"name": "",
				"type": "uint256"
			}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [],
		"name": "owner",
		"outputs": [
			{
				"internalType": "address",
				"name": "",
				"type": "address"
			}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [],
		"name": "pause",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [],
		"name": "paused",
		"outputs": [
			{
				"internalType": "bool",
				"name": "",
				"type": "bool"
			}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [],
		"name": "renounceOwnership",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [
			{
				"internalType": "uint256",
				"name": "stableCommissionPercent_",
				"type": "uint256"
			}
		],
		"name": "setStableCommissionPercent",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [
			{
				"internalType": "address",
				"name": "token",
				"type": "address"
			},
			{
				"internalType": "address",
				"name": "recipient",
				"type": "address"
			},
			{
				"internalType": "uint256",
				"name": "amount",
				"type": "uint256"
			},
			{
				"internalType": "uint256",
				"name": "commission",
				"type": "uint256"
			},
			{
				"internalType": "uint256",
				"name": "nonce",
				"type": "uint256"
			},
			{
				"internalType": "bytes",
				"name": "signature",
				"type": "bytes"
			}
		],
		"name": "transferOut",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [
			{
				"internalType": "address",
				"name": "newOwner",
				"type": "address"
			}
		],
		"name": "transferOwnership",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [],
		"name": "unpause",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [
			{
				"internalType": "address",
				"name": "token",
				"type": "address"
			},
			{
				"internalType": "uint256",
				"name": "amount",
				"type": "uint256"
			}
		],
		"name": "withdrawCommission",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	}
];
/* eslint-enable */

export const ERC20_ABI = [{ 'anonymous':false, 'inputs':[{ 'indexed':true, 'internalType':'address', 'name':'owner', 'type':'address' }, { 'indexed':true, 'internalType':'address', 'name':'spender', 'type':'address' }, { 'indexed':false, 'internalType':'uint256', 'name':'value', 'type':'uint256' }], 'name':'Approval', 'type':'event' }, { 'anonymous':false, 'inputs':[{ 'indexed':true, 'internalType':'address', 'name':'from', 'type':'address' }, { 'indexed':true, 'internalType':'address', 'name':'to', 'type':'address' }, { 'indexed':false, 'internalType':'uint256', 'name':'value', 'type':'uint256' }], 'name':'Transfer', 'type':'event' }, { 'inputs':[{ 'internalType':'address', 'name':'owner', 'type':'address' }, { 'internalType':'address', 'name':'spender', 'type':'address' }], 'name':'allowance', 'outputs':[{ 'internalType':'uint256', 'name':'', 'type':'uint256' }], 'stateMutability':'view', 'type':'function' }, { 'inputs':[{ 'internalType':'address', 'name':'spender', 'type':'address' }, { 'internalType':'uint256', 'name':'amount', 'type':'uint256' }], 'name':'approve', 'outputs':[{ 'internalType':'bool', 'name':'', 'type':'bool' }], 'stateMutability':'nonpayable', 'type':'function' }, { 'inputs':[{ 'internalType':'address', 'name':'account', 'type':'address' }], 'name':'balanceOf', 'outputs':[{ 'internalType':'uint256', 'name':'', 'type':'uint256' }], 'stateMutability':'view', 'type':'function' }, { 'inputs':[], 'name':'totalSupply', 'outputs':[{ 'internalType':'uint256', 'name':'', 'type':'uint256' }], 'stateMutability':'view', 'type':'function' }, { 'inputs':[{ 'internalType':'address', 'name':'to', 'type':'address' }, { 'internalType':'uint256', 'name':'amount', 'type':'uint256' }], 'name':'transfer', 'outputs':[{ 'internalType':'bool', 'name':'', 'type':'bool' }], 'stateMutability':'nonpayable', 'type':'function' }, { 'inputs':[{ 'internalType':'address', 'name':'from', 'type':'address' }, { 'internalType':'address', 'name':'to', 'type':'address' }, { 'internalType':'uint256', 'name':'amount', 'type':'uint256' }], 'name':'transferFrom', 'outputs':[{ 'internalType':'bool', 'name':'', 'type':'bool' }], 'stateMutability':'nonpayable', 'type':'function' }];
