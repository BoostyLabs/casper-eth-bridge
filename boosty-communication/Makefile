PROTO_PATH=proto
DOCS_PATH=docs

help: ## Show this help
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
%:
	@:

.PHONY: golang_protobuf
golang_protobuf: ## Generates protobuf implementation for Golang.
	buf generate --path ${PROTO_PATH}/networks/networks.proto
	buf generate --path ${PROTO_PATH}/transfers/transfers.proto
	buf generate --path ${PROTO_PATH}/signer/signer.proto
	buf generate --path ${PROTO_PATH}/connector/connector.proto
	buf generate --path ${PROTO_PATH}/gateway-bridge/gateway-bridge.proto
	buf generate --path ${PROTO_PATH}/bridge-signer/bridge-signer.proto
	buf generate --path ${PROTO_PATH}/bridge-connector/bridge-connector.proto
	buf generate --path ${PROTO_PATH}/connector-bridge/connector-bridge.proto
	buf generate --path ${PROTO_PATH}/bridge-oracle/bridge-oracle.proto

.PHONY: docs
docs: ## Generates docs for proto-files.
	protoc -I. -I${PROTO_PATH} --openapiv2_out=${DOCS_PATH} --openapiv2_opt=logtostderr=true ${PROTO_PATH}/networks/networks.proto
	protoc -I. -I${PROTO_PATH} --openapiv2_out=${DOCS_PATH} --openapiv2_opt=logtostderr=true ${PROTO_PATH}/transfers/transfers.proto
	protoc -I. -I${PROTO_PATH} --openapiv2_out=${DOCS_PATH} --openapiv2_opt=logtostderr=true ${PROTO_PATH}/signer/signer.proto
	protoc -I. -I${PROTO_PATH} --openapiv2_out=${DOCS_PATH} --openapiv2_opt=logtostderr=true ${PROTO_PATH}/connector/connector.proto
	protoc -I. -I${PROTO_PATH} --openapiv2_out=${DOCS_PATH} --openapiv2_opt=logtostderr=true ${PROTO_PATH}/gateway-bridge/gateway-bridge.proto
	protoc -I. -I${PROTO_PATH} --openapiv2_out=${DOCS_PATH} --openapiv2_opt=logtostderr=true ${PROTO_PATH}/bridge-signer/bridge-signer.proto
	protoc -I. -I${PROTO_PATH} --openapiv2_out=${DOCS_PATH} --openapiv2_opt=logtostderr=true ${PROTO_PATH}/bridge-connector/bridge-connector.proto
	protoc -I. -I${PROTO_PATH} --openapiv2_out=${DOCS_PATH} --openapiv2_opt=logtostderr=true ${PROTO_PATH}/connector-bridge/connector-bridge.proto
	protoc -I. -I${PROTO_PATH} --openapiv2_out=${DOCS_PATH} --openapiv2_opt=logtostderr=true ${PROTO_PATH}/bridge-oracle/bridge-oracle.proto
