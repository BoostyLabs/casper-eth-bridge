use std::fs::create_dir_all;
use std::path::Path;

fn main() -> Result<(), Box<dyn std::error::Error>> {
    create_dir_all(Path::new("src/proto"))?;
    tonic_build::configure().out_dir("src/proto").compile(
        &[
            "../proto/bridge-connector/bridge-connector.proto",
            "../proto/bridge-signer/bridge-signer.proto",
            "../proto/connector-bridge/connector-bridge.proto",
            "../proto/gateway-bridge/gateway-bridge.proto",
            "../proto/bridge-oracle/bridge-oracle.proto",
        ],
        &["../proto/"],
    )?;
    Ok(())
}
