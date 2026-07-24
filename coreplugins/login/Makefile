PLUGIN_TYPE := static
PACKAGE_STEM := login
CONTENT_DIRS := build

WORKSPACE_ROOT := ../..
ROOT_PACKAGE_JSON := $(WORKSPACE_ROOT)/package.json
ROOT_PACKAGE_LOCK := $(WORKSPACE_ROOT)/package-lock.json
ROOT_NODE_MODULES_STAMP := $(WORKSPACE_ROOT)/node_modules/.package-lock.json
FRONTEND_SOURCES := package.json $(ROOT_PACKAGE_JSON) $(ROOT_PACKAGE_LOCK) svelte.config.js tsconfig.json vite.config.ts $(shell find src -type f)

.PHONY: frontend-build frontend-clean

build: frontend-build

frontend-build: $(ROOT_NODE_MODULES_STAMP) $(FRONTEND_SOURCES)
	npm --prefix $(WORKSPACE_ROOT) run build --workspace @arupa/login

$(ROOT_NODE_MODULES_STAMP): $(ROOT_PACKAGE_JSON) $(ROOT_PACKAGE_LOCK) package.json
	npm --prefix $(WORKSPACE_ROOT) ci

clean: frontend-clean

frontend-clean:
	rm -rf build .svelte-kit

include ../../mk/plugin.mk
