# Variables
ENVIRONMENT ?=
CASPER_IMAGE_NAME = tricorn-connector-casper
ETHEREUM_IMAGE_NAME = tricorn-connector-ethereum
SIGNER_IMAGE_NAME = tricorn-signer
GATEWAY_IMAGE_NAME = tricorn-gateway
CURRENCY_RATES_IMAGE_NAME = tricorn-currencyrates
BRIDGE_IMAGE_NAME = tricorn-bridge
WEB_APP_IMAGE_NAME = tricorn-web-app
IMAGE_TAG ?= latest
LATEST_COMMIT := $$(git rev-parse HEAD)
REGISTRY_HOST ?= ghcr.io/boostylabs

IMAGE_CONNECTOR_CASPER_BACKUP = $(REGISTRY_HOST)/$(CASPER_IMAGE_NAME)$(ENVIRONMENT):$(LATEST_COMMIT)
IMAGE_CONNECTOR_CASPER_LATEST = $(REGISTRY_HOST)/$(CASPER_IMAGE_NAME)$(ENVIRONMENT):$(IMAGE_TAG)

IMAGE_CONNECTOR_ETHEREUM_BACKUP = $(REGISTRY_HOST)/$(ETHEREUM_IMAGE_NAME)$(ENVIRONMENT):$(LATEST_COMMIT)
IMAGE_CONNECTOR_ETHEREUM_LATEST = $(REGISTRY_HOST)/$(ETHEREUM_IMAGE_NAME)$(ENVIRONMENT):$(IMAGE_TAG)

IMAGE_SIGNER_BACKUP = $(REGISTRY_HOST)/$(SIGNER_IMAGE_NAME)$(ENVIRONMENT):$(LATEST_COMMIT)
IMAGE_SIGNER_LATEST = $(REGISTRY_HOST)/$(SIGNER_IMAGE_NAME)$(ENVIRONMENT):$(IMAGE_TAG)

IMAGE_GATEWAY_BACKUP = $(REGISTRY_HOST)/$(GATEWAY_IMAGE_NAME)$(ENVIRONMENT):$(LATEST_COMMIT)
IMAGE_GATEWAY_LATEST = $(REGISTRY_HOST)/$(GATEWAY_IMAGE_NAME)$(ENVIRONMENT):$(IMAGE_TAG)

IMAGE_BRIDGE_BACKUP = $(REGISTRY_HOST)/$(BRIDGE_IMAGE_NAME)$(ENVIRONMENT):$(LATEST_COMMIT)
IMAGE_BRIDGE_LATEST = $(REGISTRY_HOST)/$(BRIDGE_IMAGE_NAME)$(ENVIRONMENT):$(IMAGE_TAG)

IMAGE_WEB_APP_BACKUP = $(REGISTRY_HOST)/$(WEB_APP_IMAGE_NAME)$(ENVIRONMENT):$(LATEST_COMMIT)
IMAGE_WEB_APP_LATEST = $(REGISTRY_HOST)/$(WEB_APP_IMAGE_NAME)$(ENVIRONMENT):$(IMAGE_TAG)

help: ## Show this help
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
%:
	@:

build_connector_casper: ## Build casper connector docker image.
	DOCKER_BUILDKIT=1 docker build --ssh default -f ./deploy/casper.Dockerfile -t $(IMAGE_CONNECTOR_CASPER_BACKUP) . && DOCKER_BUILDKIT=1 docker build --ssh default -f ./deploy/casper.Dockerfile -t $(IMAGE_CONNECTOR_CASPER_LATEST) .

push_connector_casper: ## Push casper connector docker image.
	docker push $(IMAGE_CONNECTOR_CASPER_BACKUP) && docker push $(IMAGE_CONNECTOR_CASPER_LATEST)

build_connector_ethereum: ## Build ethereum connector docker image.
	DOCKER_BUILDKIT=1 docker build --ssh default -f ./deploy/ethereum.Dockerfile -t $(IMAGE_CONNECTOR_ETHEREUM_BACKUP) . && DOCKER_BUILDKIT=1 docker build --ssh default -f ./deploy/ethereum.Dockerfile -t $(IMAGE_CONNECTOR_ETHEREUM_LATEST) .

push_connector_ethereum: ## Push ethereum connector docker image.
	docker push $(IMAGE_CONNECTOR_ETHEREUM_BACKUP) && docker push $(IMAGE_CONNECTOR_ETHEREUM_LATEST)

build_signer: ## Build signer docker image.
	DOCKER_BUILDKIT=1 docker build --ssh default -f ./deploy/signer.Dockerfile -t $(IMAGE_SIGNER_BACKUP) . && DOCKER_BUILDKIT=1 docker build --ssh default -f ./deploy/signer.Dockerfile -t $(IMAGE_SIGNER_LATEST) .

push_signer: ## Push signer docker image.
	docker push $(IMAGE_SIGNER_BACKUP) && docker push $(IMAGE_SIGNER_LATEST)

build_gateway: ## Build gateway docker image.
	DOCKER_BUILDKIT=1 docker build --ssh default -f ./deploy/gateway.Dockerfile -t $(IMAGE_GATEWAY_BACKUP) . && DOCKER_BUILDKIT=1 docker build --ssh default -f ./deploy/gateway.Dockerfile -t $(IMAGE_GATEWAY_LATEST) .

push_gateway: ## Push gateway docker image.
	docker push $(IMAGE_GATEWAY_BACKUP) && docker push $(IMAGE_GATEWAY_LATEST)

build_bridge: ## Build bridge docker image.
	DOCKER_BUILDKIT=1 docker build --ssh default -f ./deploy/bridge.Dockerfile -t $(IMAGE_BRIDGE_BACKUP) . && DOCKER_BUILDKIT=1 docker build --ssh default -f ./deploy/bridge.Dockerfile -t $(IMAGE_BRIDGE_LATEST) .

push_bridge: ## Push bridge docker image.
	docker push $(IMAGE_BRIDGE_BACKUP) && docker push $(IMAGE_BRIDGE_LATEST)

build_dist: ## Build dist folder that needed for frontend.
	cd web/bridge && npm ci && npm run build

build_web_app: ## Build web-app docker image.
	make build_dist && DOCKER_BUILDKIT=1 docker build --ssh default -f ./deploy/web_app.Dockerfile -t $(IMAGE_WEB_APP_BACKUP) . && DOCKER_BUILDKIT=1 docker build --ssh default -f ./deploy/web_app.Dockerfile -t $(IMAGE_WEB_APP_LATEST) .

push_web_app: ## Push web-app docker image.
	docker push $(IMAGE_WEB_APP_BACKUP) && docker push $(IMAGE_WEB_APP_LATEST)

docker: ## Build and push all docker images.
	make build_connector_casper push_connector_casper build_connector_ethereum push_connector_ethereum build_signer push_signer build_gateway push_gateway build_bridge push_bridge build_web_app push_web_app
