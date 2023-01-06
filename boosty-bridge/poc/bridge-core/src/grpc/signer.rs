use std::str::FromStr;

use anyhow::Context;
use reqwest::Url;
use tonic::transport::{Channel, Endpoint};
use tonic_codegen::{
    bridge_signer_client::BridgeSignerClient, connector_bridge_server::ConnectorBridge,
    PublicKeyRequest, PublicKeyResponse, SignRequest, Signature,
};

/// Bridge works as a proxy between the connector and the signer.
pub struct SignerProxy {
    signer_client: BridgeSignerClient<Channel>,
}

impl SignerProxy {
    /// Lazely connects to the signer.
    pub(crate) async fn new(service_url: Url) -> anyhow::Result<SignerProxy> {
        let channel = Endpoint::from_str(service_url.as_ref())
            .context("couldn't parse endpoint")?
            .connect_lazy();
        let signer_client = BridgeSignerClient::new(channel);

        Ok(SignerProxy { signer_client })
    }
}

#[async_trait::async_trait]
impl ConnectorBridge for SignerProxy {
    /// Forwards the request to the signer and returns the response.
    #[tracing::instrument(skip(self), request, err)]
    async fn sign(
        &self,
        request: tonic::Request<SignRequest>,
    ) -> Result<tonic::Response<Signature>, tonic::Status> {
        let client = self.signer_client.clone();
        let request = request.into_inner();
        Ok(crate::grpc::retry_request(
            move || {
                let mut client = client.clone();
                let request = request.clone();
                async move { client.sign(request).await }
            },
            |error| tracing::warn!(error=%error, "error when trying to sign out, retrying"),
        )
        .await?)
    }

    /// Forwards the request to the signer and returns the response.
    #[tracing::instrument(skip(self), request, err)]
    async fn public_key(
        &self,
        request: tonic::Request<PublicKeyRequest>,
    ) -> Result<tonic::Response<PublicKeyResponse>, tonic::Status> {
        let client = self.signer_client.clone();
        let request = request.into_inner();
        Ok(crate::grpc::retry_request(
            move || {
                let mut client = client.clone();
                let request = request.clone();
                async move { client.public_key(request).await }
            },
            |error| tracing::warn!(error=%error, "error when trying to get public key, retrying"),
        )
        .await?)
    }
}
