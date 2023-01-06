#![cfg_attr(not(feature = "std"), no_std)]

extern crate alloc;

pub mod contract;
pub mod entry_points;
pub mod error;
pub mod interface {
    pub mod offchain;
    pub mod onchain;
}
