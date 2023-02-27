pub mod connector_config;
pub mod error;

#[cfg(feature = "casper")]
use casper_types::ContractHash;
#[cfg(feature = "ethereum")]
use primitive_types::H160;

#[cfg(feature = "ethereum")]
pub type EthereumConnectorConfig = connector_config::ConnectorConfig<H160>;
#[cfg(feature = "casper")]
pub type CasperConnectorConfig = connector_config::ConnectorConfig<ContractHash>;
