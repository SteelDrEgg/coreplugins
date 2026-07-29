import * as m from '$lib/paraglide/messages.js';
import type { SecretMeta } from './types';

type ApiPayload = {
	success: boolean;
	message?: string;
	[key: string]: unknown;
};

export class ApiError extends Error {}

async function api<T extends ApiPayload>(path: string, options: RequestInit = {}): Promise<T> {
	const response = await fetch(
		path,
		Object.assign(
			{
				credentials: 'include',
				headers: { Accept: 'application/json', 'Content-Type': 'application/json' }
			},
			options
		)
	);

	let payload: ApiPayload | null = null;
	try {
		payload = (await response.json()) as ApiPayload;
	} catch {
		throw new ApiError(m.api_invalid_response());
	}
	if (!response.ok || !payload.success) {
		throw new ApiError(payload.message || m.request_failed({ status: response.status }));
	}
	return payload as T;
}

export type WriteSecretInput = {
	name: string;
	description: string;
	value: string;
	passphrase: string;
	allowed_plugins: string[];
};

export async function listSecrets(): Promise<SecretMeta[]> {
	const payload = await api<ApiPayload & { keys: SecretMeta[] }>('/keys');
	return payload.keys || [];
}

export async function addSecret(input: WriteSecretInput): Promise<void> {
	await api('/keys/add', { method: 'POST', body: JSON.stringify(input) });
}

export async function updateSecret(input: WriteSecretInput): Promise<void> {
	await api('/keys/update', { method: 'POST', body: JSON.stringify(input) });
}

export async function revealSecret(name: string, passphrase: string): Promise<string> {
	const payload = await api<ApiPayload & { value: string }>('/keys/reveal', {
		method: 'POST',
		body: JSON.stringify({ name, passphrase })
	});
	return payload.value;
}

export async function deleteSecret(name: string): Promise<void> {
	await api('/keys/delete', { method: 'POST', body: JSON.stringify({ name }) });
}
