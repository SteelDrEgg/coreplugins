
// this file is generated — do not edit it


/// <reference types="@sveltejs/kit" />

/**
 * This module provides access to environment variables that are injected _statically_ into your bundle at build time and are limited to _private_ access.
 * 
 * |         | Runtime                                                                    | Build time                                                               |
 * | ------- | -------------------------------------------------------------------------- | ------------------------------------------------------------------------ |
 * | Private | [`$env/dynamic/private`](https://svelte.dev/docs/kit/$env-dynamic-private) | [`$env/static/private`](https://svelte.dev/docs/kit/$env-static-private) |
 * | Public  | [`$env/dynamic/public`](https://svelte.dev/docs/kit/$env-dynamic-public)   | [`$env/static/public`](https://svelte.dev/docs/kit/$env-static-public)   |
 * 
 * Static environment variables are [loaded by Vite](https://vitejs.dev/guide/env-and-mode.html#env-files) from `.env` files and `process.env` at build time and then statically injected into your bundle at build time, enabling optimisations like dead code elimination.
 * 
 * **_Private_ access:**
 * 
 * - This module cannot be imported into client-side code
 * - This module only includes variables that _do not_ begin with [`config.kit.env.publicPrefix`](https://svelte.dev/docs/kit/configuration#env) _and do_ start with [`config.kit.env.privatePrefix`](https://svelte.dev/docs/kit/configuration#env) (if configured)
 * 
 * For example, given the following build time environment:
 * 
 * ```env
 * ENVIRONMENT=production
 * PUBLIC_BASE_URL=http://site.com
 * ```
 * 
 * With the default `publicPrefix` and `privatePrefix`:
 * 
 * ```ts
 * import { ENVIRONMENT, PUBLIC_BASE_URL } from '$env/static/private';
 * 
 * console.log(ENVIRONMENT); // => "production"
 * console.log(PUBLIC_BASE_URL); // => throws error during build
 * ```
 * 
 * The above values will be the same _even if_ different values for `ENVIRONMENT` or `PUBLIC_BASE_URL` are set at runtime, as they are statically replaced in your code with their build time values.
 */
declare module '$env/static/private' {
	export const NODE_ENV: string;
	export const FIG_TERM: string;
	export const SVELTEKIT_FORK: string;
	export const _: string;
	export const npm_node_execpath: string;
	export const RUBYMINE_VM_OPTIONS: string;
	export const JETBRAINS_CLIENT_VM_OPTIONS: string;
	export const npm_config_user_agent: string;
	export const FNM_DIR: string;
	export const GOROOT: string;
	export const STUDIO_VM_OPTIONS: string;
	export const npm_config_prefix: string;
	export const SHLVL: string;
	export const RIDER_VM_OPTIONS: string;
	export const HOME: string;
	export const AQUA_VM_OPTIONS: string;
	export const DATASPELL_VM_OPTIONS: string;
	export const DATAGRIP_VM_OPTIONS: string;
	export const WEBSTORM_VM_OPTIONS: string;
	export const LANG: string;
	export const GOPATH: string;
	export const CLION_VM_OPTIONS: string;
	export const IDEA_VM_OPTIONS: string;
	export const EDITOR: string;
	export const npm_lifecycle_event: string;
	export const npm_config_node_gyp: string;
	export const PWD: string;
	export const USER: string;
	export const __CFBundleIdentifier: string;
	export const __CF_USER_TEXT_ENCODING: string;
	export const npm_config_init_module: string;
	export const GATEWAY_VM_OPTIONS: string;
	export const npm_package_json: string;
	export const RUSTROVER_VM_OPTIONS: string;
	export const OSLogRateLimit: string;
	export const PROCESS_LAUNCHED_BY_CW: string;
	export const FNM_NODE_DIST_MIRROR: string;
	export const DEVECOSTUDIO_VM_OPTIONS: string;
	export const PLUGIN_DIR: string;
	export const TERMINAL_EMULATOR: string;
	export const LC_CTYPE: string;
	export const TERM_SESSION_ID: string;
	export const PATH: string;
	export const MFLAGS: string;
	export const MAKELEVEL: string;
	export const FNM_RESOLVE_ENGINES: string;
	export const SSH_AUTH_SOCK: string;
	export const npm_lifecycle_script: string;
	export const PHPSTORM_VM_OPTIONS: string;
	export const COMMAND_MODE: string;
	export const FNM_MULTISHELL_PATH: string;
	export const MAKEOVERRIDES: string;
	export const PYCHARM_VM_OPTIONS: string;
	export const npm_execpath: string;
	export const npm_config_userconfig: string;
	export const XPC_FLAGS: string;
	export const npm_command: string;
	export const FNM_VERSION_FILE_STRATEGY: string;
	export const FNM_COREPACK_ENABLED: string;
	export const ROOT_DIR: string;
	export const LOGNAME: string;
	export const npm_config_local_prefix: string;
	export const npm_package_name: string;
	export const npm_config_noproxy: string;
	export const SHELL: string;
	export const COLOR: string;
	export const npm_config_global_prefix: string;
	export const npm_package_version: string;
	export const npm_config_allow_scripts: string;
	export const TERM: string;
	export const TMPDIR: string;
	export const GOLAND_VM_OPTIONS: string;
	export const npm_config_cache: string;
	export const MAKEFLAGS: string;
	export const FNM_ARCH: string;
	export const DIST_DIR: string;
	export const GO111MODULE: string;
	export const XPC_SERVICE_NAME: string;
	export const FNM_LOGLEVEL: string;
	export const INIT_CWD: string;
	export const JETBRAINSCLIENT_VM_OPTIONS: string;
	export const PNPM_HOME: string;
	export const npm_config_globalconfig: string;
	export const NODE: string;
	export const npm_config_npm_version: string;
	export const PROCESS_LAUNCHED_BY_Q: string;
	export const MANPATH: string;
	export const WEBIDE_VM_OPTIONS: string;
}

/**
 * This module provides access to environment variables that are injected _statically_ into your bundle at build time and are _publicly_ accessible.
 * 
 * |         | Runtime                                                                    | Build time                                                               |
 * | ------- | -------------------------------------------------------------------------- | ------------------------------------------------------------------------ |
 * | Private | [`$env/dynamic/private`](https://svelte.dev/docs/kit/$env-dynamic-private) | [`$env/static/private`](https://svelte.dev/docs/kit/$env-static-private) |
 * | Public  | [`$env/dynamic/public`](https://svelte.dev/docs/kit/$env-dynamic-public)   | [`$env/static/public`](https://svelte.dev/docs/kit/$env-static-public)   |
 * 
 * Static environment variables are [loaded by Vite](https://vitejs.dev/guide/env-and-mode.html#env-files) from `.env` files and `process.env` at build time and then statically injected into your bundle at build time, enabling optimisations like dead code elimination.
 * 
 * **_Public_ access:**
 * 
 * - This module _can_ be imported into client-side code
 * - **Only** variables that begin with [`config.kit.env.publicPrefix`](https://svelte.dev/docs/kit/configuration#env) (which defaults to `PUBLIC_`) are included
 * 
 * For example, given the following build time environment:
 * 
 * ```env
 * ENVIRONMENT=production
 * PUBLIC_BASE_URL=http://site.com
 * ```
 * 
 * With the default `publicPrefix` and `privatePrefix`:
 * 
 * ```ts
 * import { ENVIRONMENT, PUBLIC_BASE_URL } from '$env/static/public';
 * 
 * console.log(ENVIRONMENT); // => throws error during build
 * console.log(PUBLIC_BASE_URL); // => "http://site.com"
 * ```
 * 
 * The above values will be the same _even if_ different values for `ENVIRONMENT` or `PUBLIC_BASE_URL` are set at runtime, as they are statically replaced in your code with their build time values.
 */
declare module '$env/static/public' {
	
}

/**
 * This module provides access to environment variables set _dynamically_ at runtime and that are limited to _private_ access.
 * 
 * |         | Runtime                                                                    | Build time                                                               |
 * | ------- | -------------------------------------------------------------------------- | ------------------------------------------------------------------------ |
 * | Private | [`$env/dynamic/private`](https://svelte.dev/docs/kit/$env-dynamic-private) | [`$env/static/private`](https://svelte.dev/docs/kit/$env-static-private) |
 * | Public  | [`$env/dynamic/public`](https://svelte.dev/docs/kit/$env-dynamic-public)   | [`$env/static/public`](https://svelte.dev/docs/kit/$env-static-public)   |
 * 
 * Dynamic environment variables are defined by the platform you're running on. For example if you're using [`adapter-node`](https://github.com/sveltejs/kit/tree/main/packages/adapter-node) (or running [`vite preview`](https://svelte.dev/docs/kit/cli)), this is equivalent to `process.env`.
 * 
 * **_Private_ access:**
 * 
 * - This module cannot be imported into client-side code
 * - This module includes variables that _do not_ begin with [`config.kit.env.publicPrefix`](https://svelte.dev/docs/kit/configuration#env) _and do_ start with [`config.kit.env.privatePrefix`](https://svelte.dev/docs/kit/configuration#env) (if configured)
 * 
 * > [!NOTE] In `dev`, `$env/dynamic` includes environment variables from `.env`. In `prod`, this behavior will depend on your adapter.
 * 
 * > [!NOTE] To get correct types, environment variables referenced in your code should be declared (for example in an `.env` file), even if they don't have a value until the app is deployed:
 * >
 * > ```env
 * > MY_FEATURE_FLAG=
 * > ```
 * >
 * > You can override `.env` values from the command line like so:
 * >
 * > ```sh
 * > MY_FEATURE_FLAG="enabled" npm run dev
 * > ```
 * 
 * For example, given the following runtime environment:
 * 
 * ```env
 * ENVIRONMENT=production
 * PUBLIC_BASE_URL=http://site.com
 * ```
 * 
 * With the default `publicPrefix` and `privatePrefix`:
 * 
 * ```ts
 * import { env } from '$env/dynamic/private';
 * 
 * console.log(env.ENVIRONMENT); // => "production"
 * console.log(env.PUBLIC_BASE_URL); // => undefined
 * ```
 */
declare module '$env/dynamic/private' {
	export const env: {
		NODE_ENV: string;
		FIG_TERM: string;
		SVELTEKIT_FORK: string;
		_: string;
		npm_node_execpath: string;
		RUBYMINE_VM_OPTIONS: string;
		JETBRAINS_CLIENT_VM_OPTIONS: string;
		npm_config_user_agent: string;
		FNM_DIR: string;
		GOROOT: string;
		STUDIO_VM_OPTIONS: string;
		npm_config_prefix: string;
		SHLVL: string;
		RIDER_VM_OPTIONS: string;
		HOME: string;
		AQUA_VM_OPTIONS: string;
		DATASPELL_VM_OPTIONS: string;
		DATAGRIP_VM_OPTIONS: string;
		WEBSTORM_VM_OPTIONS: string;
		LANG: string;
		GOPATH: string;
		CLION_VM_OPTIONS: string;
		IDEA_VM_OPTIONS: string;
		EDITOR: string;
		npm_lifecycle_event: string;
		npm_config_node_gyp: string;
		PWD: string;
		USER: string;
		__CFBundleIdentifier: string;
		__CF_USER_TEXT_ENCODING: string;
		npm_config_init_module: string;
		GATEWAY_VM_OPTIONS: string;
		npm_package_json: string;
		RUSTROVER_VM_OPTIONS: string;
		OSLogRateLimit: string;
		PROCESS_LAUNCHED_BY_CW: string;
		FNM_NODE_DIST_MIRROR: string;
		DEVECOSTUDIO_VM_OPTIONS: string;
		PLUGIN_DIR: string;
		TERMINAL_EMULATOR: string;
		LC_CTYPE: string;
		TERM_SESSION_ID: string;
		PATH: string;
		MFLAGS: string;
		MAKELEVEL: string;
		FNM_RESOLVE_ENGINES: string;
		SSH_AUTH_SOCK: string;
		npm_lifecycle_script: string;
		PHPSTORM_VM_OPTIONS: string;
		COMMAND_MODE: string;
		FNM_MULTISHELL_PATH: string;
		MAKEOVERRIDES: string;
		PYCHARM_VM_OPTIONS: string;
		npm_execpath: string;
		npm_config_userconfig: string;
		XPC_FLAGS: string;
		npm_command: string;
		FNM_VERSION_FILE_STRATEGY: string;
		FNM_COREPACK_ENABLED: string;
		ROOT_DIR: string;
		LOGNAME: string;
		npm_config_local_prefix: string;
		npm_package_name: string;
		npm_config_noproxy: string;
		SHELL: string;
		COLOR: string;
		npm_config_global_prefix: string;
		npm_package_version: string;
		npm_config_allow_scripts: string;
		TERM: string;
		TMPDIR: string;
		GOLAND_VM_OPTIONS: string;
		npm_config_cache: string;
		MAKEFLAGS: string;
		FNM_ARCH: string;
		DIST_DIR: string;
		GO111MODULE: string;
		XPC_SERVICE_NAME: string;
		FNM_LOGLEVEL: string;
		INIT_CWD: string;
		JETBRAINSCLIENT_VM_OPTIONS: string;
		PNPM_HOME: string;
		npm_config_globalconfig: string;
		NODE: string;
		npm_config_npm_version: string;
		PROCESS_LAUNCHED_BY_Q: string;
		MANPATH: string;
		WEBIDE_VM_OPTIONS: string;
		[key: `PUBLIC_${string}`]: undefined;
		[key: `${string}`]: string | undefined;
	}
}

/**
 * This module provides access to environment variables set _dynamically_ at runtime and that are _publicly_ accessible.
 * 
 * |         | Runtime                                                                    | Build time                                                               |
 * | ------- | -------------------------------------------------------------------------- | ------------------------------------------------------------------------ |
 * | Private | [`$env/dynamic/private`](https://svelte.dev/docs/kit/$env-dynamic-private) | [`$env/static/private`](https://svelte.dev/docs/kit/$env-static-private) |
 * | Public  | [`$env/dynamic/public`](https://svelte.dev/docs/kit/$env-dynamic-public)   | [`$env/static/public`](https://svelte.dev/docs/kit/$env-static-public)   |
 * 
 * Dynamic environment variables are defined by the platform you're running on. For example if you're using [`adapter-node`](https://github.com/sveltejs/kit/tree/main/packages/adapter-node) (or running [`vite preview`](https://svelte.dev/docs/kit/cli)), this is equivalent to `process.env`.
 * 
 * **_Public_ access:**
 * 
 * - This module _can_ be imported into client-side code
 * - **Only** variables that begin with [`config.kit.env.publicPrefix`](https://svelte.dev/docs/kit/configuration#env) (which defaults to `PUBLIC_`) are included
 * 
 * > [!NOTE] In `dev`, `$env/dynamic` includes environment variables from `.env`. In `prod`, this behavior will depend on your adapter.
 * 
 * > [!NOTE] To get correct types, environment variables referenced in your code should be declared (for example in an `.env` file), even if they don't have a value until the app is deployed:
 * >
 * > ```env
 * > MY_FEATURE_FLAG=
 * > ```
 * >
 * > You can override `.env` values from the command line like so:
 * >
 * > ```sh
 * > MY_FEATURE_FLAG="enabled" npm run dev
 * > ```
 * 
 * For example, given the following runtime environment:
 * 
 * ```env
 * ENVIRONMENT=production
 * PUBLIC_BASE_URL=http://example.com
 * ```
 * 
 * With the default `publicPrefix` and `privatePrefix`:
 * 
 * ```ts
 * import { env } from '$env/dynamic/public';
 * console.log(env.ENVIRONMENT); // => undefined, not public
 * console.log(env.PUBLIC_BASE_URL); // => "http://example.com"
 * ```
 * 
 * ```
 * 
 * ```
 */
declare module '$env/dynamic/public' {
	export const env: {
		[key: `PUBLIC_${string}`]: string | undefined;
	}
}
