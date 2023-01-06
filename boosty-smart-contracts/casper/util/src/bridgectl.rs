use serde::{Deserialize, Serialize};

#[derive(Deserialize, Serialize)]
pub enum BridgeCommand {
    TransferOut {
        source_chain: String,
        source_address: String,
        recipient: String,
        amount: String,
    },
}
