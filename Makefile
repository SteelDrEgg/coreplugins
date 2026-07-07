GOBIN := $(shell go env GOPATH)/bin
PROTOC_GEN_GO_PLUGIN := $(GOBIN)/protoc-gen-go-plugin

PLUGIN_DIR := plugins
DIST_DIR := dist

CORE_PLUGIN_TARGETS := hello web-assets login navigator plugin-manager ssh

.PHONY: tools proto proto-grpc proto-wasm plugins $(CORE_PLUGIN_TARGETS) hello-wasm web-assets-wasm login-wasm navigator-wasm plugin-manager-wasm ssh-grpc clean

define WASM_BUILD_RULE
$(1)-wasm:
	mkdir -p $$(DIST_DIR)
	GOOS=wasip1 GOARCH=wasm go build -o $$(DIST_DIR)/$(2).wasm -buildmode=c-shared ./coreplugins/$(3)
endef

define PACKAGE_RULE
$(1): $(2)
	rm -rf $$(DIST_DIR)/$(3)_pkg
	mkdir -p $$(DIST_DIR)/$(3)_pkg/Content $$(addprefix $$(DIST_DIR)/$(3)_pkg/Content/,$(6)) $$(PLUGIN_DIR)
	cp $$(DIST_DIR)/$(4) $$(DIST_DIR)/$(3)_pkg/Content/$(4)
	cp coreplugins/$(5)/info.yaml $$(DIST_DIR)/$(3)_pkg/info.yaml
	$$(if $$($(1)_CONTENT),$$($(1)_CONTENT),:)
	cd $$(DIST_DIR)/$(3)_pkg && zip -qr ../../$$(PLUGIN_DIR)/$(1).plg .
	rm -rf $$(DIST_DIR)/$(3)_pkg
endef

web-assets_CONTENT = cp -R coreplugins/webassets/assets $(DIST_DIR)/web_assets_pkg/Content/
login_CONTENT = cp -R coreplugins/login/pages $(DIST_DIR)/login_pkg/Content/
navigator_CONTENT = cp -R coreplugins/navigator/pages $(DIST_DIR)/navigator_pkg/Content/
plugin-manager_CONTENT = cp -R coreplugins/pluginmanager/pages $(DIST_DIR)/plugin_manager_pkg/Content
ssh_CONTENT = cp -R coreplugins/ssh/pages $(DIST_DIR)/ssh_pkg/Content && cp -R coreplugins/ssh/assets $(DIST_DIR)/ssh_pkg/Content

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
plugins: $(CORE_PLUGIN_TARGETS)

$(eval $(call WASM_BUILD_RULE,hello,hello,hello))
$(eval $(call WASM_BUILD_RULE,web-assets,web_assets,webassets))
$(eval $(call WASM_BUILD_RULE,login,login,login))
$(eval $(call WASM_BUILD_RULE,navigator,navigator,navigator))
$(eval $(call WASM_BUILD_RULE,plugin-manager,plugin_manager,pluginmanager))

$(eval $(call PACKAGE_RULE,hello,hello-wasm,hello,hello.wasm,hello,))
$(eval $(call PACKAGE_RULE,web-assets,web-assets-wasm,web_assets,web_assets.wasm,webassets,assets))
$(eval $(call PACKAGE_RULE,login,login-wasm,login,login.wasm,login,pages))
$(eval $(call PACKAGE_RULE,navigator,navigator-wasm,navigator,navigator.wasm,navigator,pages))
$(eval $(call PACKAGE_RULE,plugin-manager,plugin-manager-wasm,plugin_manager,plugin_manager.wasm,pluginmanager,pages))
$(eval $(call PACKAGE_RULE,ssh,ssh-grpc,ssh,ssh-plugin,ssh,pages assets/terminal))

ssh-grpc:
	mkdir -p $(DIST_DIR)
	go build -o $(DIST_DIR)/ssh-plugin ./coreplugins/ssh

## clean: remove build artifacts
clean:
	rm -rf $(DIST_DIR) $(PLUGIN_DIR)/hello.plg $(PLUGIN_DIR)/web-assets.plg $(PLUGIN_DIR)/login.plg $(PLUGIN_DIR)/navigator.plg $(PLUGIN_DIR)/plugin-manager.plg $(PLUGIN_DIR)/ssh.plg tmp
