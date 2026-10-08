import * as m from '$lib/paraglide/messages.js';
import type { SecretInfo, Protection } from './types';

type ApiPayload = {
	success: boolean;
	message?: string;
	[key: string]: unknown;
};

export class ApiError extends Error {}

const secretCollectionPath = '/secret-manager/secrets';

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

export type SecretValueInput = {
	plaintext: string;
	protection: Protection;
	passphrase?: string;
};

export type CreateSecretInput = {
	name: string;
	description?: string;
	value: SecretValueInput;
	allowed_plugins?: string[];
};

export type UpdateSecretInput = {
	name: string;
	description?: string;
	value?: SecretValueInput;
	allowed_plugins?: string[];
};

export type SaveSecretInput =
	| { kind: 'create'; input: CreateSecretInput }
	| { kind: 'update'; input: UpdateSecretInput };

export async function listSecrets(): Promise<SecretInfo[]> {
	const payload = await api<ApiPayload & { keys: SecretInfo[] }>(secretCollectionPath);
	return payload.keys || [];
}

export async function createSecret(input: CreateSecretInput): Promise<void> {
	await api(secretCollectionPath, { method: 'POST', body: JSON.stringify(input) });
}

export async function updateSecret(input: UpdateSecretInput): Promise<void> {
	const { name, ...fields } = input;
	await api(secretResourcePath(name), { method: 'PATCH', body: JSON.stringify(fields) });
}

export async function revealSecret(name: string, passphrase: string): Promise<string> {
	const payload = await api<ApiPayload & { value: string }>(`${secretResourcePath(name)}/reveal`, {
		method: 'POST',
		body: JSON.stringify({ passphrase })
	});
	return payload.value;
}

export async function deleteSecret(name: string): Promise<void> {
	await api(secretResourcePath(name), { method: 'DELETE' });
}

function secretResourcePath(name: string): string {
	// A single-dot path segment is normalized away by the browser.
	if (name === '.') return `${secretCollectionPath}/~.`;
	return `${secretCollectionPath}/${encodeURIComponent(name)}`;
}
