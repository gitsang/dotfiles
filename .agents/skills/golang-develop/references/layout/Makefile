SHELL := /usr/bin/env bash

# version
VERSION    ?= $(shell git describe --tags --always --match "v*" --dirty 2>/dev/null || git describe --tags --always --dirty 2>/dev/null || echo dev)
BUILD_DATE ?= $(shell date -u --iso-8601=seconds)
GIT_COMMIT := $(shell git rev-parse HEAD 2>/dev/null || echo unknown)

# go
GO          := go
GO_VERSION  := $(shell $(GO) version | awk '{print $$3}')
GOOS        := $(if $(GOOS),$(GOOS),$(shell $(GO) env GOOS))
GOARCH      := $(if $(GOARCH),$(GOARCH),$(shell $(GO) env GOARCH))
GOPATH      := $(shell $(GO) env GOPATH)

# proto
PROTO_DIR     ?= ./api/proto
PROTO_API_DIR ?= $(PROTO_DIR)/v1
PROTO_API_PATH ?= api/proto/v1

# dist
OS        ?= linux
ARCH      ?= amd64
DIST_ROOT ?= dist

# petstore
PETSTORE ?= petstore
# petstore:go
PETSTORE_LDFLAGS += -s -w
PETSTORE_LDFLAGS += -X "golang-develop-example/cmd/petstore.Version=$(VERSION)"
PETSTORE_LDFLAGS += -X "golang-develop-example/cmd/petstore.BuildDate=$(BUILD_DATE)"
PETSTORE_LDFLAGS += -X "golang-develop-example/cmd/petstore.GitCommit=$(GIT_COMMIT)"
PETSTORE_LDFLAGS += -X "golang-develop-example/cmd/petstore.GoVersion=$(GO_VERSION)"
PETSTORE_LDFLAGS += -X "golang-develop-example/cmd/petstore.GOOS=$(GOOS)"
PETSTORE_LDFLAGS += -X "golang-develop-example/cmd/petstore.GOARCH=$(GOARCH)"
# petstore:dist
PETSTORE_DIST      ?= $(DIST_ROOT)/$(PETSTORE)/$(OS)/$(ARCH)
PETSTORE_DIST_BIN  ?= $(PETSTORE_DIST)/bin
PETSTORE_DIST_CONFIG ?= $(PETSTORE_DIST)/configs

#------------------------------------------------------------------------------#
##@ Prerequisites
#------------------------------------------------------------------------------#

# dir
BUILD_DIR ?= $(PWD)/.builds
BUILD_BIN ?= $(BUILD_DIR)/bin
$(BUILD_BIN):
	mkdir -p $(BUILD_BIN)
export PATH := $(BUILD_BIN):$(PATH)

# dlv
# https://github.com/go-delve/delve
DELVE         = $(BUILD_BIN)/dlv
DELVE_VERSION = v1.24.0
$(DELVE): $(BUILD_BIN)
ifeq ($(wildcard $(DELVE)),)
	GOBIN=$(BUILD_BIN) CGO_ENABLED=0 $(GO) install github.com/go-delve/delve/cmd/dlv@$(DELVE_VERSION)
endif

# goimports
# https://pkg.go.dev/golang.org/x/tools/cmd/goimports
GOIMPORTS         = $(BUILD_BIN)/goimports
GOIMPORTS_VERSION = latest
$(GOIMPORTS): $(BUILD_BIN)
ifeq ($(wildcard $(GOIMPORTS)),)
	GOBIN=$(BUILD_BIN) CGO_ENABLED=0 $(GO) install golang.org/x/tools/cmd/goimports@$(GOIMPORTS_VERSION)
endif

# golangci-lint
# https://github.com/golangci/golangci-lint
GOLANGCI_LINT         = $(BUILD_BIN)/golangci-lint
GOLANGCI_LINT_VERSION = v2.12.2
$(GOLANGCI_LINT): $(BUILD_BIN)
ifeq ($(wildcard $(GOLANGCI_LINT)),)
	GOBIN=$(BUILD_BIN) CGO_ENABLED=0 $(GO) install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)
endif

# nilaway
# https://github.com/uber-go/nilaway
NILAWAY         = $(BUILD_BIN)/nilaway
NILAWAY_VERSION = latest
$(NILAWAY): $(BUILD_BIN)
ifeq ($(wildcard $(NILAWAY)),)
	GOBIN=$(BUILD_BIN) CGO_ENABLED=0 $(GO) install go.uber.org/nilaway/cmd/nilaway@$(NILAWAY_VERSION)
endif

# buf
# https://github.com/bufbuild/buf
BUF         = $(BUILD_BIN)/buf
BUF_VERSION = v1.26.1
$(BUF): $(BUILD_BIN)
ifeq ($(wildcard $(BUF)),)
	GOBIN=$(BUILD_BIN) CGO_ENABLED=0 $(GO) install github.com/bufbuild/buf/cmd/buf@$(BUF_VERSION)
endif

# protoc-gen-go
# https://grpc.io/docs/protoc-installation/
PROTOC_GEN_GO         = $(BUILD_BIN)/protoc-gen-go
PROTOC_GEN_GO_VERSION = v1.31.0
$(PROTOC_GEN_GO): $(BUILD_BIN)
ifeq ($(wildcard $(PROTOC_GEN_GO)),)
	GOBIN=$(BUILD_BIN) CGO_ENABLED=0 $(GO) install google.golang.org/protobuf/cmd/protoc-gen-go@$(PROTOC_GEN_GO_VERSION)
endif

# protoc-gen-go-grpc
# https://grpc.io/docs/languages/go/quickstart/
PROTOC_GEN_GO_GRPC         = $(BUILD_BIN)/protoc-gen-go-grpc
PROTOC_GEN_GO_GRPC_VERSION = v1.3.0
$(PROTOC_GEN_GO_GRPC): $(BUILD_BIN)
ifeq ($(wildcard $(PROTOC_GEN_GO_GRPC)),)
	GOBIN=$(BUILD_BIN) CGO_ENABLED=0 $(GO) install google.golang.org/grpc/cmd/protoc-gen-go-grpc@$(PROTOC_GEN_GO_GRPC_VERSION)
endif

# protoc-gen-grpc-gateway
# https://github.com/grpc-ecosystem/grpc-gateway/tree/main/protoc-gen-grpc-gateway
PROTOC_GEN_GRPC_GATEWAY         = $(BUILD_BIN)/protoc-gen-grpc-gateway
PROTOC_GEN_GRPC_GATEWAY_VERSION = v2.24.0
$(PROTOC_GEN_GRPC_GATEWAY): $(BUILD_BIN)
ifeq ($(wildcard $(PROTOC_GEN_GRPC_GATEWAY)),)
	GOBIN=$(BUILD_BIN) CGO_ENABLED=0 $(GO) install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@$(PROTOC_GEN_GRPC_GATEWAY_VERSION)
endif

# protoc-gen-openapiv2
# https://github.com/grpc-ecosystem/grpc-gateway/tree/main/protoc-gen-openapiv2
PROTOC_GEN_OPENAPIV2         = $(BUILD_BIN)/protoc-gen-openapiv2
PROTOC_GEN_OPENAPIV2_VERSION = v2.24.0
$(PROTOC_GEN_OPENAPIV2): $(BUILD_BIN)
ifeq ($(wildcard $(PROTOC_GEN_OPENAPIV2)),)
	GOBIN=$(BUILD_BIN) CGO_ENABLED=0 $(GO) install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@$(PROTOC_GEN_OPENAPIV2_VERSION)
endif

## init development env
init: $(DELVE) \
			$(GOIMPORTS) \
			$(GOLANGCI_LINT) \
			$(NILAWAY) \
			$(BUF) \
			$(PROTOC_GEN_GO) \
			$(PROTOC_GEN_GO_GRPC) \
			$(PROTOC_GEN_GRPC_GATEWAY) \
			$(PROTOC_GEN_OPENAPIV2)
	@echo "All prerequisites installed."

#------------------------------------------------------------------------------#
##@ Generate
#------------------------------------------------------------------------------#

.PHONY: protoc-gen
## proto generate all
protoc-gen: $(BUF) \
						$(PROTOC_GEN_GO) \
						$(PROTOC_GEN_GO_GRPC) \
						$(PROTOC_GEN_GRPC_GATEWAY) \
						$(PROTOC_GEN_OPENAPIV2)
	mkdir -p pkg/api/v1 docs/api
	$(BUF) mod update
	$(BUF) generate . --template buf.gen.yaml --path $(PROTO_API_PATH)

.PHONY: generate
## generate all
generate: protoc-gen
	$(GO) generate ./...

#------------------------------------------------------------------------------#
##@ Develop
#------------------------------------------------------------------------------#

.PHONY: fmt
## format code
fmt: $(GOIMPORTS)
	$(GOIMPORTS) ./...

.PHONY: test
## run unit tests
test:
	$(GO) test ./...

.PHONY: run-petstore
## run petstore app
run-petstore:
	CGO_ENABLED=0 GOOS=$(OS) GOARCH=$(ARCH) \
		$(GO) run \
			-trimpath -v -ldflags '$(PETSTORE_LDFLAGS)' \
			./cmd/$(PETSTORE) \
			$(filter-out $@, $(MAKECMDGOALS))

.PHONY: debug-petstore
## debug run petstore app
debug-petstore: $(DELVE)
	$(DELVE) debug \
		./cmd/$(PETSTORE) \
		--log \
		--output /tmp/__debug_$(PETSTORE) \
		-- $(filter-out $@, $(MAKECMDGOALS))

#------------------------------------------------------------------------------#
##@ Build
#------------------------------------------------------------------------------#

# petstore dist directories
$(PETSTORE_DIST_BIN):
	@rm -fr $@ && mkdir -p $@

$(PETSTORE_DIST_CONFIG):
	@rm -fr $@ && mkdir -p $@

$(PETSTORE_DIST): $(PETSTORE_DIST_BIN) $(PETSTORE_DIST_CONFIG)
	@mkdir -p $@

.PHONY: build-artifacts-petstore
## build petstore all local artifacts
build-artifacts-petstore: $(PETSTORE_DIST)
	CGO_ENABLED=0 GOOS=$(OS) GOARCH=$(ARCH) \
		$(GO) build \
			-trimpath -v -ldflags '$(PETSTORE_LDFLAGS)' \
			-o $(PETSTORE_DIST_BIN)/$(PETSTORE) \
			./cmd/$(PETSTORE)
	rsync -a --delete configs/config.example.yml $(PETSTORE_DIST_CONFIG)/config.yml 2>/dev/null || true

.PHONY: build-artifacts
## build all applications local artifacts
build-artifacts: build-artifacts-petstore

.PHONY: build-container-petstore
## build petstore container image
build-container-petstore:
	$(CONTAINER_CLI) build \
		--build-arg DOCKER_PACKAGE_PATH="$(PETSTORE_DIST)" \
		-t "$(CONTAINER_REPO)/$(PETSTORE):$(subst +,-,$(VERSION))" \
		-f "Containerfile.$(PETSTORE)" .

.PHONY: build-container
## build all container images
build-container: build-container-petstore

.PHONY: build
## build all
build: build-artifacts build-container

#------------------------------------------------------------------------------#
##@ Publish
#------------------------------------------------------------------------------#

.PHONY: publish-petstore
## publish petstore container
publish-petstore:
	$(CONTAINER_CLI) push $(CONTAINER_REPO)/$(PETSTORE):$(subst +,-,$(VERSION))

.PHONY: publish
## publish all containers
publish: publish-petstore

.PHONY: publish-release-petstore
## publish petstore container to release repo
publish-release-petstore:
	$(CONTAINER_CLI) tag $(CONTAINER_REPO)/$(PETSTORE):$(subst +,-,$(VERSION)) $(CONTAINER_REPO_RELEASE)/$(PETSTORE):$(subst +,-,$(VERSION))
	$(CONTAINER_CLI) push $(CONTAINER_REPO_RELEASE)/$(PETSTORE):$(subst +,-,$(VERSION))

.PHONY: publish-release
## publish all containers to release repo
publish-release: publish-release-petstore

#------------------------------------------------------------------------------#
##@ Clean
#------------------------------------------------------------------------------#

.PHONY: clean
## clean up git repo
clean:
	git clean -xfd
	rm -fr $(DIST_ROOT)

#------------------------------------------------------------------------------#
##@ Help
#------------------------------------------------------------------------------#

HELP_TAB_LEN := 36
## display help
help:
	@awk 'BEGIN \
	{ \
		FS = ":.*##"; \
		printf "\nUsage:\n  make \033[36m<target>\033[0m\n" \
	} \
	/^[0-9a-zA-Z\_\-\/]+:/ \
	{ \
		helpMessage = match(lastLine, /^## (.*)/); \
		if (helpMessage) { \
			helpCommand = substr($$1, 0, index($$1, ":")-1); \
			helpMessage = substr(lastLine, RSTART + 2, RLENGTH); \
			printf "  \033[36m%-$(HELP_TAB_LEN)s\033[0m %s\n", helpCommand,helpMessage; \
		} \
	} { lastLine = $$0 } \
	/^##@/ \
	{ \
		printf "\n\033[1m%s\033[0m\n", substr($$0, 5) \
	} ' $(MAKEFILE_LIST)

.DEFAULT_GOAL := help
