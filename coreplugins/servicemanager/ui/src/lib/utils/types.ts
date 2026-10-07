export type ServiceStatus = 'discovered' | 'starting' | 'running' | 'degraded' | 'stopping' | 'failed' | 'backoff' | 'exited';
export type ServiceAction = 'start' | 'stop' | 'restart';

export interface AccessPolicy {
	require_auth?: boolean;
	groups?: string[];
}

export interface ServiceRoute {
	id: string;
	transport: string;
	http?: {
		method?: string;
		pattern: string;
		access?: AccessPolicy;
		rewrite?: { prefix?: boolean; location?: boolean };
	};
	socket_io?: {
		namespace: string;
		events?: string[];
		access?: AccessPolicy;
		event_access?: Record<string, AccessPolicy>;
	};
}

export interface ServiceTransport {
	id: string;
	type: 'static' | 'http' | 'proxy' | 'socket.io';
	source?: string;
	proxy?: { network: string; address?: string; scheme?: string };
}

export interface ServiceConfig {
	restart?: string;
	run_as_user?: string;
	checksum?: string;
	allow?: string[] | null;
}

export interface ConfigResponse extends ServiceConfig {
	name: string;
	configured: boolean;
}

export type ConfigPatch = {
	[K in keyof ServiceConfig]?: ServiceConfig[K] | null;
};

export interface DiscoveredService {
	name: string;
	version: string;
	type: string;
	contract_version: number;
	command: string;
	package_path: string;
	config: ServiceConfig;
	metadata?: Record<string, unknown>;
	status: ServiceStatus;
	retry_count: number;
	next_retry_at?: string;
	last_exit?: { exit_code: number; error?: string; exited_at: string };
	last_error?: string;
}

export interface RunningService {
	instance_id: string;
	name: string;
	version: string;
	type: string;
	path: string;
	routes?: ServiceRoute[];
	transports?: ServiceTransport[];
}

export interface ServiceRow {
	name: string;
	disk?: DiscoveredService;
	instances: RunningService[];
	status: ServiceStatus;
	versionMismatch: boolean;
}

export interface ServiceDirectory { service_dir: string }
export interface ServiceTempDirectory { temp_dir: string; requires_restart?: boolean }
export interface ParamsResponse { name: string; params: Record<string, string> | null }
export interface ParamsPatch { set?: Record<string, string>; remove?: string[] }
