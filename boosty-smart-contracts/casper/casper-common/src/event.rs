use alloc::{string::String, vec::Vec};
use casper_types::{
    bytesrepr::{self, FromBytes, ToBytes},
    ContractPackageHash, Key, U128, U256,
};
use contract_util::event::ContractEvent;

pub const BRIDGE_EVENT_FUNDS_IN_TAG: u8 = 0;
pub const BRIDGE_EVENT_FUNDS_OUT_TAG: u8 = 1;
pub const BRIDGE_EVENT_TRANSFER_OUT: u8 = 2;
pub const BRIDGE_EVENT_WITHDRAW_COMMISSION: u8 = 3;

#[derive(Debug, PartialEq, Eq)]
pub enum BridgeEvent {
    FundsIn {
        token_contract: ContractPackageHash,
        destination_chain: String,
        destination_address: String,
        amount: U256,
        gas_commission: U256,
        stable_commission_percent: U256,
        nonce: U128,
        sender: Key,
    },
    FundsOut {
        token_contract: ContractPackageHash,
        source_chain: String,
        source_address: String,
        amount: U256,
        transaction_id: U256,
        recipient: Key,
    },
    TransferOut {
        token_contract: ContractPackageHash,
        total_sum_for_transfer: U256,
        nonce: U128,
        recipient: Key,
    },
    WithdrawCommission {
        token_contract: ContractPackageHash,
        amount: U256,
    },
}

impl ContractEvent for BridgeEvent {}

impl ToBytes for BridgeEvent {
    fn to_bytes(&self) -> Result<Vec<u8>, bytesrepr::Error> {
        let mut buffer = bytesrepr::allocate_buffer(self)?;

        match self {
            BridgeEvent::FundsIn {
                token_contract,
                destination_chain,
                destination_address,
                amount,
                gas_commission,
                stable_commission_percent,
                nonce,
                sender,
            } => {
                buffer.push(BRIDGE_EVENT_FUNDS_IN_TAG);
                buffer.extend(token_contract.to_bytes()?);
                buffer.extend(destination_chain.to_bytes()?);
                buffer.extend(destination_address.to_bytes()?);
                buffer.extend(amount.to_bytes()?);
                buffer.extend(gas_commission.to_bytes()?);
                buffer.extend(stable_commission_percent.to_bytes()?);
                buffer.extend(nonce.to_bytes()?);
                buffer.extend(sender.to_bytes()?);
            }
            BridgeEvent::FundsOut {
                token_contract,
                source_chain,
                source_address,
                amount,
                transaction_id,
                recipient,
            } => {
                buffer.push(BRIDGE_EVENT_FUNDS_OUT_TAG);
                buffer.extend(token_contract.to_bytes()?);
                buffer.extend(source_chain.to_bytes()?);
                buffer.extend(source_address.to_bytes()?);
                buffer.extend(amount.to_bytes()?);
                buffer.extend(transaction_id.to_bytes()?);
                buffer.extend(recipient.to_bytes()?);
            }
            BridgeEvent::TransferOut {
                token_contract,
                total_sum_for_transfer,
                nonce,
                recipient,
            } => {
                buffer.push(BRIDGE_EVENT_TRANSFER_OUT);
                buffer.extend(token_contract.to_bytes()?);
                buffer.extend(total_sum_for_transfer.to_bytes()?);
                buffer.extend(nonce.to_bytes()?);
                buffer.extend(recipient.to_bytes()?);
            }
            BridgeEvent::WithdrawCommission {
                token_contract,
                amount,
            } => {
                buffer.push(BRIDGE_EVENT_WITHDRAW_COMMISSION);
                buffer.extend(token_contract.to_bytes()?);
                buffer.extend(amount.to_bytes()?);
            }
        }

        Ok(buffer)
    }

    fn serialized_length(&self) -> usize {
        match self {
            BridgeEvent::FundsIn {
                token_contract,
                destination_chain,
                destination_address,
                amount,
                gas_commission,
                stable_commission_percent,
                nonce,
                sender,
            } => {
                destination_chain.serialized_length()
                    + destination_address.serialized_length()
                    + amount.serialized_length()
                    + gas_commission.serialized_length()
                    + stable_commission_percent.serialized_length()
                    + nonce.serialized_length()
                    + sender.serialized_length()
                    + token_contract.serialized_length()
            }
            BridgeEvent::FundsOut {
                token_contract,
                source_chain,
                source_address,
                amount,
                transaction_id,
                recipient,
            } => {
                source_chain.serialized_length()
                    + source_address.serialized_length()
                    + amount.serialized_length()
                    + transaction_id.serialized_length()
                    + recipient.serialized_length()
                    + token_contract.serialized_length()
            }
            BridgeEvent::TransferOut {
                token_contract,
                total_sum_for_transfer,
                nonce,
                recipient,
            } => {
                token_contract.serialized_length()
                    + total_sum_for_transfer.serialized_length()
                    + nonce.serialized_length()
                    + recipient.serialized_length()
            }
            BridgeEvent::WithdrawCommission {
                token_contract,
                amount,
            } => token_contract.serialized_length() + amount.serialized_length(),
        }
    }
}

impl FromBytes for BridgeEvent {
    fn from_bytes(bytes: &[u8]) -> Result<(Self, &[u8]), bytesrepr::Error> {
        let (tag, remainder) = u8::from_bytes(bytes)?;
        match tag {
            BRIDGE_EVENT_FUNDS_IN_TAG => {
                let (token_contract, remainder) = ContractPackageHash::from_bytes(remainder)?;
                let (destination_chain, remainder) = String::from_bytes(remainder)?;
                let (destination_address, remainder) = String::from_bytes(remainder)?;
                let (amount, remainder) = U256::from_bytes(remainder)?;
                let (gas_commission, remainder) = U256::from_bytes(remainder)?;
                let (stable_commission_percent, remainder) = U256::from_bytes(remainder)?;
                let (nonce, remainder) = U128::from_bytes(remainder)?;
                let (sender, remainder) = Key::from_bytes(remainder)?;
                Ok((
                    BridgeEvent::FundsIn {
                        token_contract,
                        destination_chain,
                        destination_address,
                        amount,
                        gas_commission,
                        stable_commission_percent,
                        nonce,
                        sender,
                    },
                    remainder,
                ))
            }
            BRIDGE_EVENT_FUNDS_OUT_TAG => {
                let (token_contract, remainder) = ContractPackageHash::from_bytes(remainder)?;
                let (source_chain, remainder) = String::from_bytes(remainder)?;
                let (source_address, remainder) = String::from_bytes(remainder)?;
                let (amount, remainder) = U256::from_bytes(remainder)?;
                let (transaction_id, remainder) = U256::from_bytes(remainder)?;
                let (recipient, remainder) = Key::from_bytes(remainder)?;
                Ok((
                    BridgeEvent::FundsOut {
                        token_contract,
                        source_chain,
                        source_address,
                        amount,
                        transaction_id,
                        recipient,
                    },
                    remainder,
                ))
            }
            BRIDGE_EVENT_TRANSFER_OUT => {
                let (token_contract, remainder) = ContractPackageHash::from_bytes(remainder)?;
                let (total_sum_for_transfer, remainder) = U256::from_bytes(remainder)?;
                let (nonce, remainder) = U128::from_bytes(remainder)?;
                let (recipient, remainder) = Key::from_bytes(remainder)?;
                Ok((
                    BridgeEvent::TransferOut {
                        token_contract,
                        total_sum_for_transfer,
                        nonce,
                        recipient,
                    },
                    remainder,
                ))
            }
            BRIDGE_EVENT_WITHDRAW_COMMISSION => {
                let (token_contract, remainder) = ContractPackageHash::from_bytes(remainder)?;
                let (amount, remainder) = U256::from_bytes(remainder)?;
                Ok((
                    BridgeEvent::WithdrawCommission {
                        token_contract,
                        amount,
                    },
                    remainder,
                ))
            }
            _ => Err(bytesrepr::Error::Formatting),
        }
    }
}
