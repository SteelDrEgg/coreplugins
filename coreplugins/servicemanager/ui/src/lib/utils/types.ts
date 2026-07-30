export type ServiceStatus =
	| 'discovered'
	| 'starting'
	| 'running'
	| 'degraded'
	| 'stopping'
	| 'failed'
	| string;

export type ServiceMetadata = Record<string, unknown>;

export type DiscoveredService = {
	name: string;
	version?: string;
	type?: string;
	contract_version?: number;
	command?: string;
	package_path?: string;
	config?: Record<string, unknown>;
	metadata?: ServiceMetadata;
	status?: ServiceStatus;
};

export type ServiceTransport = {
	id?: string;
	type?: string;
	source?: string;
	[key: string]: unknown;
};

export type ServiceRoute = {
	id?: string;
	transport?: string;
	http?: {
		method?: string;
		pattern?: string;
		[key: string]: unknown;
	};
	[key: string]: unknown;
};

export type RunningService = {
	instance_id?: string;
	name?: string;
	version?: string;
	type?: string;
	path?: string;
	transports?: ServiceTransport[];
	routes?: ServiceRoute[];
};

export type ServiceDirectory = {
	service_dir: string;
};

export type ServiceTempDirectory = {
	temp_dir: string;
	requires_restart: boolean;
};

export type ServiceAction = 'start' | 'stop' | 'restart';

export type MessageKind = 'info' | 'success' | 'warning' | 'error';

export type BannerMessage = {
	text: string;
	kind: MessageKind;
};
