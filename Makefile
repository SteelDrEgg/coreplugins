GOBIN := $(shell go env GOPATH)/bin
PROTOC_GEN_GO_PLUGIN := $(GOBIN)/protoc-gen-go-plugin

PLUGIN_DIR ?= plugins
DIST_DIR ?= dist

CORE_PLUGIN_TARGETS := hello web-assets login navigator plugin-manager ssh web-sdk

.PHONY: tools proto proto-grpc proto-wasm plugins $(CORE_PLUGIN_TARGETS) clean

## tools: install the protobuf generators used by `make proto`
tools:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/knqyf263/go-plugin/cmd/protoc-gen-go-plugin@v0.9.0

## proto: regenerate gRPC and WASM Go code from proto/panel.proto
proto: proto-grpc proto-wasm

proto-grpc:
	mkdir -p pluginsdk/grpc
	PATH="$(GOBIN):$(PATH)" protoc -I. \
		--go_out=./pluginsdk/grpc --go_opt=paths=source_relative \
		--go-grpc_out=./pluginsdk/grpc --go-grpc_opt=paths=source_relative \
		./proto/panel.proto

proto-wasm:
	mkdir -p pluginsdk/wasm
	protoc --plugin=protoc-gen-go-plugin=$(PROTOC_GEN_GO_PLUGIN) -I. \
		--go-plugin_out=./pluginsdk/wasm --go-plugin_opt=paths=source_relative \
		./proto/panel.proto

## plugins: build and package every core plugin into plugins/*.plg
plugins: proto $(CORE_PLUGIN_TARGETS)

$(CORE_PLUGIN_TARGETS): proto

hello:
	$(MAKE) -C coreplugins/hello package ROOT_DIR=$(CURDIR) DIST_DIR=$(CURDIR)/$(DIST_DIR) PLUGIN_DIR=$(CURDIR)/$(PLUGIN_DIR)

web-sdk:
	$(MAKE) -C coreplugins/websdk package ROOT_DIR=$(CURDIR) DIST_DIR=$(CURDIR)/$(DIST_DIR) PLUGIN_DIR=$(CURDIR)/$(PLUGIN_DIR)

web-assets:
	$(MAKE) -C coreplugins/webassets package ROOT_DIR=$(CURDIR) DIST_DIR=$(CURDIR)/$(DIST_DIR) PLUGIN_DIR=$(CURDIR)/$(PLUGIN_DIR)

login:
	$(MAKE) -C coreplugins/login package ROOT_DIR=$(CURDIR) DIST_DIR=$(CURDIR)/$(DIST_DIR) PLUGIN_DIR=$(CURDIR)/$(PLUGIN_DIR)

navigator:
	$(MAKE) -C coreplugins/navigator package ROOT_DIR=$(CURDIR) DIST_DIR=$(CURDIR)/$(DIST_DIR) PLUGIN_DIR=$(CURDIR)/$(PLUGIN_DIR)

plugin-manager:
	$(MAKE) -C coreplugins/pluginmanager package ROOT_DIR=$(CURDIR) DIST_DIR=$(CURDIR)/$(DIST_DIR) PLUGIN_DIR=$(CURDIR)/$(PLUGIN_DIR)

ssh:
	$(MAKE) -C coreplugins/ssh package ROOT_DIR=$(CURDIR) DIST_DIR=$(CURDIR)/$(DIST_DIR) PLUGIN_DIR=$(CURDIR)/$(PLUGIN_DIR)

## clean: remove build artifacts
clean:
	rm -rf $(DIST_DIR) tmp
	rm -f $(PLUGIN_DIR)/hello.plg $(PLUGIN_DIR)/web-assets.plg $(PLUGIN_DIR)/login.plg $(PLUGIN_DIR)/navigator.plg $(PLUGIN_DIR)/plugin-manager.plg $(PLUGIN_DIR)/ssh.plg
