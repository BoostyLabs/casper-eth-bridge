use alloc::{string::String, vec::Vec};
use casper_types::{
    bytesrepr::{self, FromBytes, ToBytes},
    ContractPackageHash, Key, U256,
};
use contract_util::event::ContractEvent;

pub const BRIDGE_EVENT_FUNDS_IN_TAG: u8 = 0;
pub const BRIDGE_EVENT_FUNDS_OUT_TAG: u8 = 1;

#[derive(Debug, PartialEq, Eq)]
pub enum BridgeEvent {
    FundsIn {
        token_contract: ContractPackageHash,
        destination_chain: String,
        destination_address: String,
        amount: U256,
        sender: Key,
    },
    FundsOut {
        token_contract: ContractPackageHash,
        source_chain: String,
        source_address: String,
        amount: U256,
        recipient: Key,
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
                sender,
            } => {
                buffer.push(BRIDGE_EVENT_FUNDS_IN_TAG);
                buffer.extend(token_contract.to_bytes()?);
                buffer.extend(destination_chain.to_bytes()?);
                buffer.extend(destination_address.to_bytes()?);
                buffer.extend(amount.to_bytes()?);
                buffer.extend(sender.to_bytes()?);
            }
            BridgeEvent::FundsOut {
                token_contract,
                source_chain,
                source_address,
                amount,
                recipient,
            } => {
                buffer.push(BRIDGE_EVENT_FUNDS_OUT_TAG);
                buffer.extend(token_contract.to_bytes()?);
                buffer.extend(source_chain.to_bytes()?);
                buffer.extend(source_address.to_bytes()?);
                buffer.extend(amount.to_bytes()?);
                buffer.extend(recipient.to_bytes()?);
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
                sender,
            } => {
                destination_chain.serialized_length()
                    + destination_address.serialized_length()
                    + amount.serialized_length()
                    + sender.serialized_length()
                    + token_contract.serialized_length()
            }
            BridgeEvent::FundsOut {
                token_contract,
                source_chain,
                source_address,
                amount,
                recipient,
            } => {
                source_chain.serialized_length()
                    + source_address.serialized_length()
                    + amount.serialized_length()
                    + recipient.serialized_length()
                    + token_contract.serialized_length()
            }
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
                let (sender, remainder) = Key::from_bytes(remainder)?;
                Ok((
                    BridgeEvent::FundsIn {
                        token_contract,
                        destination_chain,
                        destination_address,
                        amount,
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
                let (recipient, remainder) = Key::from_bytes(remainder)?;
                Ok((
                    BridgeEvent::FundsOut {
                        token_contract,
                        source_chain,
                        source_address,
                        amount,
                        recipient,
                    },
                    remainder,
                ))
            }
            _ => Err(bytesrepr::Error::Formatting),
        }
    }
}
