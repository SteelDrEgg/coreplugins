import type {
	DiscoveredService,
	RunningService,
	ServiceAction,
	ServiceDirectory,
	ServiceTempDirectory
} from './types';
import * as m from '$lib/paraglide/messages.js';
import { serviceAPIErrorMessages } from './errorMessages';

type ApiEnvelope<T> = {
	success?: boolean;
	message?: string;
	error?: string;
	data?: T;
};

export class ApiError extends Error {
	readonly status: number;
	readonly path: string;

	constructor(message: string, status = 0, path = '') {
		super(message);
		this.name = 'ApiError';
		this.status = status;
		this.path = path;
	}
}

function responseMessage(payload: ApiEnvelope<unknown> | null, rawBody: string): string {
	const envelopeMessage = payload?.message || payload?.error;
	if (envelopeMessage) return envelopeMessage;

	const content = rawBody
		.replace(/<script\b[^>]*>[\s\S]*?<\/script>/gi, ' ')
		.replace(/<style\b[^>]*>[\s\S]*?<\/style>/gi, ' ')
		.replace(/<[^>]+>/g, ' ')
		.replace(/\s+/g, ' ')
		.trim();
	return content.length > 240 ? `${content.slice(0, 237)}...` : content;
}

function httpError(
	response: Response,
	path: string,
	payload: ApiEnvelope<unknown> | null,
	rawBody: string
): ApiError {
	const knownMessage = serviceAPIErrorMessages[response.status];
	if (knownMessage) return new ApiError(knownMessage(), response.status, path);

	const detail = responseMessage(payload, rawBody);
	const status = `HTTP ${response.status}${response.statusText ? ` ${response.statusText}` : ''}`;
	return new ApiError(
		m.service_api_request_failed({
			status,
			detail: detail ? `: ${detail}` : ''
		}),
		response.status,
		path
	);
}

async function request<T>(path: string, options: RequestInit = {}): Promise<ApiEnvelope<T>> {
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
	let payload: ApiEnvelope<T> | null = null;
	try {
		if (rawBody.trim()) payload = JSON.parse(rawBody) as ApiEnvelope<T>;
	} catch {
		if (!response.ok) throw httpError(response, path, null, rawBody);
		throw new ApiError(m.service_api_invalid_json(), response.status, path);
	}

	if (!response.ok) throw httpError(response, path, payload, rawBody);
	if (!payload) {
		throw new ApiError(m.service_api_empty_response(), response.status, path);
	}
	if (!payload.success) {
		throw new ApiError(
			payload.message || payload.error || m.service_api_unsuccessful(),
			response.status,
			path
		);
	}
	return payload;
}

function requiredData<T>(payload: ApiEnvelope<T>, fallbackMessage: string): T {
	if (payload.data == null) throw new ApiError(payload.message || fallbackMessage);
	return payload.data;
}

export async function listDiscoveredServices(): Promise<DiscoveredService[]> {
	const payload = await request<{ services?: DiscoveredService[] }>(
		'/api/service/discovered?include=metadata'
	);
	const data = requiredData(payload, m.discovered_services_no_data());
	return Array.isArray(data.services) ? data.services : [];
}

export async function listRunningServices(): Promise<RunningService[]> {
	const payload = await request<{ services?: RunningService[] }>('/api/service/running');
	const data = requiredData(payload, m.running_services_no_data());
	return Array.isArray(data.services) ? data.services : [];
}

export async function getServiceDirectory(): Promise<ServiceDirectory> {
	const payload = await request<ServiceDirectory>('/api/service/dir');
	return requiredData(payload, m.service_directory_no_data());
}

export async function updateServiceDirectory(serviceDir: string): Promise<ServiceDirectory> {
	const payload = await request<ServiceDirectory>('/api/service/dir', {
		method: 'PATCH',
		body: JSON.stringify({ service_dir: serviceDir })
	});
	return requiredData(payload, m.service_directory_no_data());
}

export async function getServiceTempDirectory(): Promise<ServiceTempDirectory> {
	const payload = await request<ServiceTempDirectory>('/api/service/temp-dir');
	return requiredData(payload, m.temporary_directory_no_data());
}

export async function updateServiceTempDirectory(tempDir: string): Promise<ServiceTempDirectory> {
	const payload = await request<ServiceTempDirectory>('/api/service/temp-dir', {
		method: 'PATCH',
		body: JSON.stringify({ temp_dir: tempDir })
	});
	return requiredData(payload, m.temporary_directory_no_data());
}

export async function runServiceAction(action: ServiceAction, name: string): Promise<void> {
	await request<{ name?: string }>(
		`/api/service/${action}/${encodeURIComponent(name)}`,
		{ method: 'POST' }
	);
}
