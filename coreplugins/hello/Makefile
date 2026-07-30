ROOT_DIR ?= $(abspath $(CURDIR)/../..)
DIST_DIR ?= $(ROOT_DIR)/dist
PLUGIN_DIR ?= $(ROOT_DIR)/plugins

PLUGIN_NAME := $(shell awk -F': *' '$$1 == "Name" { print $$2; exit }' info.yaml)
PLUGIN_VERSION := $(shell awk -F': *' '$$1 == "Version" { print $$2; exit }' info.yaml)
PLUGIN_BINARY := hello.wasm
PACKAGE_STEM := hello
PACKAGE_WORKDIR := $(abspath $(DIST_DIR))/$(PACKAGE_STEM)_pkg
DIST_DIR_ABS := $(abspath $(DIST_DIR))
PLUGIN_DIR_ABS := $(abspath $(PLUGIN_DIR))
LD_FLAGS := -X main.pluginVersion=$(PLUGIN_VERSION)

.PHONY: all validate build package clean

all: package

validate:
	@test -n "$(PLUGIN_VERSION)" || (echo "missing Version in info.yaml"; exit 1)
	@test -f info.yaml
	@test -d assets

build: validate
	mkdir -p "$(DIST_DIR_ABS)"
	CGO_ENABLED=0 GOOS=wasip1 GOARCH=wasm go build -o "$(DIST_DIR_ABS)/$(PLUGIN_BINARY)" -buildmode=c-shared -ldflags "$(LD_FLAGS)" .

package: build
	rm -rf "$(PACKAGE_WORKDIR)"
	mkdir -p "$(PACKAGE_WORKDIR)/Content/assets" "$(PLUGIN_DIR_ABS)"
	cp info.yaml "$(PACKAGE_WORKDIR)/info.yaml"
	cp -R assets/. "$(PACKAGE_WORKDIR)/Content/assets/"
	cp "$(DIST_DIR_ABS)/$(PLUGIN_BINARY)" "$(PACKAGE_WORKDIR)/Content/$(PLUGIN_BINARY)"
	rm -f "$(PLUGIN_DIR_ABS)/$(PLUGIN_NAME).plg"
	cd "$(PACKAGE_WORKDIR)" && zip -qr "$(PLUGIN_DIR_ABS)/$(PLUGIN_NAME).plg" .
	rm -rf "$(PACKAGE_WORKDIR)"

clean:
	rm -rf "$(DIST_DIR_ABS)/$(PLUGIN_BINARY)" "$(PACKAGE_WORKDIR)" "$(PLUGIN_DIR_ABS)/$(PLUGIN_NAME).plg"
