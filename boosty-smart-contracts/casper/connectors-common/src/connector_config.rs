use serde::{Deserialize, Serialize};
use std::fs::read_to_string;

#[derive(Debug, Clone, Deserialize, Serialize)]
pub struct ConnectorConfig<BridgeHash> {
    pub url: url::Url,
    pub network_id: i32,
    pub network_name: String,
    pub is_testnet: bool,
    pub bridge_contract_hash: BridgeHash,
}

impl<BridgeHash> ConnectorConfig<BridgeHash>
where
    for<'de> BridgeHash: Deserialize<'de>,
{
    pub fn from_toml(filename: std::path::PathBuf) -> anyhow::Result<Self> {
        let file_content = read_to_string(&filename)?;
        toml::from_str(&file_content).map_err(|e| {
            anyhow::anyhow!(
                "Failed to parse an toml file({}): {}",
                filename.to_string_lossy(),
                e.to_string()
            )
        })
    }
}
