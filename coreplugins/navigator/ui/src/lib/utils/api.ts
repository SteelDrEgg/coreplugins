import type { EntriesPayload, NavigatorConfig } from './types';

type APIResponse<T> = {
	success?: boolean;
	message?: string;
	error?: string;
	data?: T;
};

export class APIError extends Error {
	readonly status: number;
	readonly path: string;

	constructor(message: string, status = 0, path = '') {
		super(message);
		this.name = 'APIError';
		this.status = status;
		this.path = path;
	}
}

async function request<T>(path: string, options: RequestInit = {}): Promise<APIResponse<T>> {
	const headers = new Headers(options.headers);
	headers.set('Accept', 'application/json');
	if (options.body != null && !headers.has('Content-Type')) {
		headers.set('Content-Type', 'application/json');
	}

	const response = await fetch(path, {
		...options,
		credentials: 'include',
		headers
	});
	const rawBody = await response.text();
	let payload: APIResponse<T>;
	try {
		payload = rawBody ? (JSON.parse(rawBody) as APIResponse<T>) : {};
	} catch {
		throw new APIError('API returned an invalid JSON response', response.status, path);
	}
	if (!response.ok || !payload.success) {
		throw new APIError(
			payload.message || payload.error || `Request failed with HTTP ${response.status}`,
			response.status,
			path
		);
	}
	return payload;
}

function requiredData<T>(payload: APIResponse<T>, message: string): T {
	if (payload.data == null) throw new APIError(payload.message || message);
	return payload.data;
}

export async function loadEntries(): Promise<EntriesPayload> {
	const payload = await request<EntriesPayload>('/navigator/api/entries');
	const data = requiredData(payload, 'Navigation response has no data');
	return {
		entries: Array.isArray(data.entries) ? data.entries : [],
		languages:
			Array.isArray(data.languages) && data.languages.length > 0 ? data.languages : ['en']
	};
}

export async function loadNavigatorConfig(): Promise<NavigatorConfig> {
	const payload = await request<NavigatorConfig>('/navigator/api/config');
	const data = requiredData(payload, 'Navigator settings response has no data');
	return {
		icon: typeof data.icon === 'string' && data.icon.trim() ? data.icon : '/Arupa.svg',
		order: Array.isArray(data.order) ? data.order.filter((name) => typeof name === 'string') : []
	};
}

export async function saveNavigatorConfig(config: NavigatorConfig): Promise<NavigatorConfig> {
	const payload = await request<NavigatorConfig>('/navigator/api/config', {
		method: 'PUT',
		body: JSON.stringify(config)
	});
	const data = requiredData(payload, 'Saved Navigator settings response has no data');
	return {
		icon: typeof data.icon === 'string' && data.icon.trim() ? data.icon : '/Arupa.svg',
		order: Array.isArray(data.order) ? data.order.filter((name) => typeof name === 'string') : []
	};
}

export async function loadKernelVersion(): Promise<string> {
	const payload = await request<{ version?: string }>('/api/kernel/version');
	const data = requiredData(payload, 'Kernel version response has no data');
	return typeof data.version === 'string' && data.version ? data.version : 'unknown';
}

export async function reloadKernel(): Promise<string> {
	const payload = await request<null>('/api/kernel/reload', { method: 'POST' });
	return payload.message || 'Configuration reloaded';
}
