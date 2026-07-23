ROOT_DIR ?= ../..
DIST_DIR ?= $(ROOT_DIR)/dist
PLUGIN_DIR ?= $(ROOT_DIR)/services

DIST_DIR_ABS := $(abspath $(DIST_DIR))
PLUGIN_DIR_ABS := $(abspath $(PLUGIN_DIR))
PACKAGE_WORKDIR := $(DIST_DIR_ABS)/$(PACKAGE_STEM)_pkg

PLUGIN_NAME := $(shell awk -F': *' '$$1 == "Name" { print $$2; exit }' info.yaml)
PLUGIN_VERSION := $(shell awk -F': *' '$$1 == "Version" { print $$2; exit }' info.yaml)
LD_FLAGS := -X main.pluginVersion=$(PLUGIN_VERSION)
CGO_ENABLED ?= 0

.PHONY: build package clean

build:
	$(if $(PLUGIN_NAME),,$(error missing Name in info.yaml))
	$(if $(PLUGIN_VERSION),,$(error missing Version in info.yaml))
	$(if $(PLUGIN_BINARY),,$(error PLUGIN_BINARY is required))
	$(if $(PACKAGE_STEM),,$(error PACKAGE_STEM is required))
	mkdir -p $(DIST_DIR_ABS)
	$(if $(filter wasm,$(PLUGIN_TYPE)),CGO_ENABLED=0 GOOS=wasip1 GOARCH=wasm go build -o $(DIST_DIR_ABS)/$(PLUGIN_BINARY) -buildmode=c-shared -ldflags "$(LD_FLAGS)" .,CGO_ENABLED=$(CGO_ENABLED) GOOS=linux GOARCH=amd64 go build -o $(DIST_DIR_ABS)/$(PLUGIN_BINARY) -ldflags "$(LD_FLAGS)" .)

package: build
	rm -rf $(PACKAGE_WORKDIR)
	mkdir -p $(PACKAGE_WORKDIR)/Content $(PLUGIN_DIR_ABS)
	cp $(DIST_DIR_ABS)/$(PLUGIN_BINARY) $(PACKAGE_WORKDIR)/Content/$(PLUGIN_BINARY)
	cp info.yaml $(PACKAGE_WORKDIR)/info.yaml
	$(foreach dir,$(CONTENT_DIRS),cp -R $(dir) $(PACKAGE_WORKDIR)/Content/$(dir);)
	rm -f $(PLUGIN_DIR_ABS)/$(PLUGIN_NAME).plg
	cd $(PACKAGE_WORKDIR) && zip -qr $(PLUGIN_DIR_ABS)/$(PLUGIN_NAME).plg .
	rm -rf $(PACKAGE_WORKDIR)

clean:
	rm -rf $(DIST_DIR_ABS)/$(PLUGIN_BINARY) $(PACKAGE_WORKDIR) $(PLUGIN_DIR_ABS)/$(PLUGIN_NAME).plg
