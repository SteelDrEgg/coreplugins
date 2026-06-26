GOBIN := $(shell go env GOPATH)/bin
PROTOC_GEN_GO_PLUGIN := $(GOBIN)/protoc-gen-go-plugin

PLUGIN_DIR := plugins
DIST_DIR := dist

.PHONY: tools proto proto-grpc proto-wasm build run hello hello-wasm clean

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

## build: build the host server binary
build:
	mkdir -p $(DIST_DIR)
	go build -o $(DIST_DIR)/minimalpanel ./cmd

## run: run the host server
run:
	go run ./cmd

## hello: build and package the hello WASM plugin into plugins/hello.plg
hello: hello-wasm
	rm -rf $(DIST_DIR)/hello_pkg
	mkdir -p $(DIST_DIR)/hello_pkg/Content $(PLUGIN_DIR)
	cp $(DIST_DIR)/hello.wasm $(DIST_DIR)/hello_pkg/Content/hello.wasm
	cp coreplugins/hello/info.yaml $(DIST_DIR)/hello_pkg/info.yaml
	cd $(DIST_DIR)/hello_pkg && zip -qr ../../$(PLUGIN_DIR)/hello.plg .
	rm -rf $(DIST_DIR)/hello_pkg

hello-wasm:
	mkdir -p $(DIST_DIR)
	GOOS=wasip1 GOARCH=wasm go build -o $(DIST_DIR)/hello.wasm -buildmode=c-shared ./coreplugins/hello

## clean: remove build artifacts
clean:
	rm -rf $(DIST_DIR) $(PLUGIN_DIR)/hello.plg tmp
