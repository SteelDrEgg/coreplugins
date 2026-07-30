import * as m from '$lib/paraglide/messages.js';
import type { EntriesPayload, NavigatorConfig, ServerFriendlyName } from './types';

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
		throw new APIError(m.api_invalid_json(), response.status, path);
	}
	if (!response.ok || !payload.success) {
		throw new APIError(
			payload.message || payload.error || m.request_failed({ status: String(response.status) }),
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
	const data = requiredData(payload, m.navigation_no_data());
	return {
		entries: Array.isArray(data.entries) ? data.entries : []
	};
}

export async function loadNavigatorConfig(): Promise<NavigatorConfig> {
	const payload = await request<NavigatorConfig>('/navigator/api/config');
	const data = requiredData(payload, m.navigator_settings_no_data());
	return {
		icon: typeof data.icon === 'string' && data.icon.trim() ? data.icon : '/Arupa.svg',
		order: Array.isArray(data.order) ? data.order.filter((name) => typeof name === 'string') : [],
		hide: Array.isArray(data.hide) ? data.hide.filter((name) => typeof name === 'string') : []
	};
}

export async function saveNavigatorConfig(config: NavigatorConfig): Promise<NavigatorConfig> {
	const payload = await request<NavigatorConfig>('/navigator/api/config', {
		method: 'PUT',
		body: JSON.stringify(config)
	});
	const data = requiredData(payload, m.saved_navigator_settings_no_data());
	return {
		icon: typeof data.icon === 'string' && data.icon.trim() ? data.icon : '/Arupa.svg',
		order: Array.isArray(data.order) ? data.order.filter((name) => typeof name === 'string') : [],
		hide: Array.isArray(data.hide) ? data.hide.filter((name) => typeof name === 'string') : []
	};
}

export async function loadServerFriendlyName(): Promise<string> {
	const payload = await request<ServerFriendlyName>('/usr/server-friendly-name');
	const data = requiredData(payload, m.server_friendly_name_no_data());
	return typeof data.name === 'string' && data.name.trim() ? data.name.trim() : 'Arupa';
}

export async function saveServerFriendlyName(name: string): Promise<string> {
	const payload = await request<ServerFriendlyName>('/usr/server-friendly-name', {
		method: 'PUT',
		body: JSON.stringify({ name })
	});
	const data = requiredData(payload, m.server_friendly_name_no_data());
	return typeof data.name === 'string' && data.name.trim() ? data.name.trim() : 'Arupa';
}

export async function loadKernelVersion(): Promise<string> {
	const payload = await request<{ version?: string }>('/api/kernel/version');
	const data = requiredData(payload, m.kernel_version_no_data());
	return typeof data.version === 'string' && data.version ? data.version : m.unknown();
}

export async function reloadKernel(): Promise<string> {
	const payload = await request<null>('/api/kernel/reload', { method: 'POST' });
	return payload.message || m.configuration_reloaded();
}
