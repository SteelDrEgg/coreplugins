import type { EntriesPayload } from './types';

type APIResponse<T> = {
	success?: boolean;
	message?: string;
	error?: string;
	data?: T;
};

export async function loadEntries(): Promise<EntriesPayload> {
	const response = await fetch('/navigator/api/entries', {
		credentials: 'include',
		headers: { Accept: 'application/json' }
	});
	let payload: APIResponse<EntriesPayload>;
	try {
		payload = (await response.json()) as APIResponse<EntriesPayload>;
	} catch {
		throw new Error('Navigation service returned an invalid response');
	}
	if (!response.ok || !payload.success || !payload.data) {
		throw new Error(payload.message || payload.error || 'Failed to load navigation entries');
	}
	return {
		entries: Array.isArray(payload.data.entries) ? payload.data.entries : [],
		languages:
			Array.isArray(payload.data.languages) && payload.data.languages.length > 0
				? payload.data.languages
				: ['en']
	};
}
