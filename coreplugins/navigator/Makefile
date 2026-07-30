ROOT_DIR ?= $(abspath $(CURDIR)/../..)
DIST_DIR ?= $(ROOT_DIR)/dist
PLUGIN_DIR ?= $(ROOT_DIR)/plugins

PLUGIN_NAME := $(shell awk -F': *' '$$1 == "Name" { print $$2; exit }' info.yaml)
PLUGIN_VERSION := $(shell awk -F': *' '$$1 == "Version" { print $$2; exit }' info.yaml)
PLUGIN_BINARY := navigator.wasm
PACKAGE_STEM := navigator
PACKAGE_WORKDIR := $(abspath $(DIST_DIR))/$(PACKAGE_STEM)_pkg
DIST_DIR_ABS := $(abspath $(DIST_DIR))
PLUGIN_DIR_ABS := $(abspath $(PLUGIN_DIR))
LD_FLAGS := -X main.pluginVersion=$(PLUGIN_VERSION)

WORKSPACE_ROOT := $(ROOT_DIR)
UI_DIR := ui
ROOT_PACKAGE_JSON := $(WORKSPACE_ROOT)/package.json
ROOT_PACKAGE_LOCK := $(WORKSPACE_ROOT)/package-lock.json
ROOT_NODE_MODULES_STAMP := $(WORKSPACE_ROOT)/node_modules/.package-lock.json
FRONTEND_SOURCES := $(UI_DIR)/package.json $(ROOT_PACKAGE_JSON) $(ROOT_PACKAGE_LOCK) $(UI_DIR)/svelte.config.js $(UI_DIR)/tsconfig.json $(UI_DIR)/vite.config.ts $(UI_DIR)/paraglide.config.js $(shell find $(UI_DIR)/src $(UI_DIR)/static $(UI_DIR)/locale $(UI_DIR)/project.inlang $(UI_DIR)/scripts -type f)

.PHONY: all validate build package frontend-build frontend-clean clean

all: package

validate:
	@test -n "$(PLUGIN_VERSION)" || (echo "missing Version in info.yaml"; exit 1)
	@test -f info.yaml
	@test -d "$(UI_DIR)"

build: validate frontend-build
	mkdir -p "$(DIST_DIR_ABS)"
	CGO_ENABLED=0 GOOS=wasip1 GOARCH=wasm go build -o "$(DIST_DIR_ABS)/$(PLUGIN_BINARY)" -buildmode=c-shared -ldflags "$(LD_FLAGS)" ./core

frontend-build: $(ROOT_NODE_MODULES_STAMP) $(FRONTEND_SOURCES)
	npm --prefix "$(WORKSPACE_ROOT)" run build --workspace @arupa/navigator

$(ROOT_NODE_MODULES_STAMP): $(ROOT_PACKAGE_JSON) $(ROOT_PACKAGE_LOCK) $(UI_DIR)/package.json
	npm --prefix "$(WORKSPACE_ROOT)" ci

package: build
	rm -rf "$(PACKAGE_WORKDIR)"
	mkdir -p "$(PACKAGE_WORKDIR)/Content/ui/build" "$(PLUGIN_DIR_ABS)"
	cp info.yaml "$(PACKAGE_WORKDIR)/info.yaml"
	cp -R "$(UI_DIR)/build/." "$(PACKAGE_WORKDIR)/Content/ui/build/"
	cp "$(DIST_DIR_ABS)/$(PLUGIN_BINARY)" "$(PACKAGE_WORKDIR)/Content/$(PLUGIN_BINARY)"
	rm -f "$(PLUGIN_DIR_ABS)/$(PLUGIN_NAME).plg"
	cd "$(PACKAGE_WORKDIR)" && zip -qr "$(PLUGIN_DIR_ABS)/$(PLUGIN_NAME).plg" .
	rm -rf "$(PACKAGE_WORKDIR)"

frontend-clean:
	rm -rf "$(UI_DIR)/build" "$(UI_DIR)/.svelte-kit"

clean: frontend-clean
	rm -rf "$(DIST_DIR_ABS)/$(PLUGIN_BINARY)" "$(PACKAGE_WORKDIR)" "$(PLUGIN_DIR_ABS)/$(PLUGIN_NAME).plg"
