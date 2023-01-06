# Variables
ENVIRONMENT ?=
IMAGE_TAG ?= latest
LATEST_COMMIT := $$(git rev-parse HEAD)
REGISTRY_HOST ?= ghcr.io/boostylabs
BRIDGE_IMAGE_NAME = golden-gate-bridge-core

IMAGE_BRIDGE_CORE_BACKUP = $(REGISTRY_HOST)/$(BRIDGE_IMAGE_NAME)$(ENVIRONMENT):$(LATEST_COMMIT)
IMAGE_BRIDGE_CORE_LATEST = $(REGISTRY_HOST)/$(BRIDGE_IMAGE_NAME)$(ENVIRONMENT):$(IMAGE_TAG)

help: ## Show this help
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
%:
	@:

build_target_bridge_core: ## Build target folder that needed for bridge_core service.
	cd poc/bridge-core && CARGO_NET_GIT_FETCH_WITH_CLI=true cargo build --release

build_bridge_core: ## Build bridge-core docker image.
	make build_target_bridge_core && DOCKER_BUILDKIT=1 docker build --no-cache --secret id=gitkey,src=$$HOME/.ssh/id_rsa -f ./deploy/bridge_core.Dockerfile -t $(IMAGE_BRIDGE_CORE_BACKUP) . && DOCKER_BUILDKIT=1 docker build --no-cache --secret id=gitkey,src=$$HOME/.ssh/id_rsa -f ./deploy/bridge_core.Dockerfile -t $(IMAGE_BRIDGE_CORE_LATEST) .

push_bridge_core: ## Push bridge-core docker image.
	docker push $(IMAGE_BRIDGE_CORE_BACKUP) && docker push $(IMAGE_BRIDGE_CORE_LATEST)

docker: ## Build and push all docker images.
	make build_bridge_core push_bridge_core
