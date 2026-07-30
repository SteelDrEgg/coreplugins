ROOT_DIR := $(CURDIR)

# PLUGIN_DIR is the directory containing the final .plg packages.  DIST_DIR
# contains binaries and temporary package directories used while building.
PLUGIN_DIR ?= plugins
DIST_DIR ?= dist

PLUGIN_DIR_ABS := $(abspath $(PLUGIN_DIR))
DIST_DIR_ABS := $(abspath $(DIST_DIR))

SERVICES := hello login navigator secret-manager service-manager ssh web-assets

.PHONY: all plugins clean $(SERVICES) secretmanager servicemanager webassets

all: plugins

## plugins: build every service under coreplugins/.
plugins: $(SERVICES)

define build-service
	$(MAKE) -C coreplugins/$(1) package \
		ROOT_DIR="$(ROOT_DIR)" \
		DIST_DIR="$(DIST_DIR_ABS)" \
		PLUGIN_DIR="$(PLUGIN_DIR_ABS)"
endef

hello:
	$(call build-service,hello)

login:
	$(call build-service,login)

navigator:
	$(call build-service,navigator)

secret-manager:
	$(call build-service,secretmanager)

service-manager:
	$(call build-service,servicemanager)

ssh:
	$(call build-service,ssh)

web-assets:
	$(call build-service,webassets)

# Directory-name aliases are useful when invoking make from the repository
# root, while the canonical targets above follow the package names in info.yaml.
secretmanager: secret-manager
servicemanager: service-manager
webassets: web-assets

## clean: remove package outputs and service build artifacts.
clean:
	$(MAKE) -C coreplugins/hello clean ROOT_DIR="$(ROOT_DIR)" DIST_DIR="$(DIST_DIR_ABS)" PLUGIN_DIR="$(PLUGIN_DIR_ABS)"
	$(MAKE) -C coreplugins/login clean ROOT_DIR="$(ROOT_DIR)" DIST_DIR="$(DIST_DIR_ABS)" PLUGIN_DIR="$(PLUGIN_DIR_ABS)"
	$(MAKE) -C coreplugins/navigator clean ROOT_DIR="$(ROOT_DIR)" DIST_DIR="$(DIST_DIR_ABS)" PLUGIN_DIR="$(PLUGIN_DIR_ABS)"
	$(MAKE) -C coreplugins/secretmanager clean ROOT_DIR="$(ROOT_DIR)" DIST_DIR="$(DIST_DIR_ABS)" PLUGIN_DIR="$(PLUGIN_DIR_ABS)"
	$(MAKE) -C coreplugins/servicemanager clean ROOT_DIR="$(ROOT_DIR)" DIST_DIR="$(DIST_DIR_ABS)" PLUGIN_DIR="$(PLUGIN_DIR_ABS)"
	$(MAKE) -C coreplugins/ssh clean ROOT_DIR="$(ROOT_DIR)" DIST_DIR="$(DIST_DIR_ABS)" PLUGIN_DIR="$(PLUGIN_DIR_ABS)"
	$(MAKE) -C coreplugins/webassets clean ROOT_DIR="$(ROOT_DIR)" DIST_DIR="$(DIST_DIR_ABS)" PLUGIN_DIR="$(PLUGIN_DIR_ABS)"
	rm -rf "$(DIST_DIR_ABS)" "$(ROOT_DIR)/tmp"
