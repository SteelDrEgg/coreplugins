PLUGIN_TYPE := static
PACKAGE_STEM := service_manager
CONTENT_DIRS := ui/build

WORKSPACE_ROOT := ../..
UI_DIR := ui
ROOT_PACKAGE_JSON := $(WORKSPACE_ROOT)/package.json
ROOT_PACKAGE_LOCK := $(WORKSPACE_ROOT)/package-lock.json
ROOT_NODE_MODULES_STAMP := $(WORKSPACE_ROOT)/node_modules/.package-lock.json
FRONTEND_SOURCES := $(UI_DIR)/package.json $(ROOT_PACKAGE_JSON) $(ROOT_PACKAGE_LOCK) $(UI_DIR)/svelte.config.js $(UI_DIR)/tsconfig.json $(UI_DIR)/vite.config.ts $(shell find $(UI_DIR)/src $(UI_DIR)/static -type f)

.PHONY: frontend-build frontend-clean

build: frontend-build

frontend-build: $(ROOT_NODE_MODULES_STAMP) $(FRONTEND_SOURCES)
	npm --prefix $(WORKSPACE_ROOT) run build --workspace @arupa/servicemanager

$(ROOT_NODE_MODULES_STAMP): $(ROOT_PACKAGE_JSON) $(ROOT_PACKAGE_LOCK) $(UI_DIR)/package.json
	npm --prefix $(WORKSPACE_ROOT) ci

clean: frontend-clean

frontend-clean:
	rm -rf $(UI_DIR)/build $(UI_DIR)/.svelte-kit

include ../../mk/plugin.mk
