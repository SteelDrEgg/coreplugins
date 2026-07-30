import { writable } from 'svelte/store';

type ServerFriendlyNameResponse = {
	success?: boolean;
	data?: {
		name?: string;
	};
};

const DEFAULT_SERVER_NAME = 'Arupa';
const SERVER_FRIENDLY_NAME_ENDPOINT = '/usr/server-friendly-name';

export const serverName = writable(DEFAULT_SERVER_NAME);

export function initializeServerName(): () => void {
	const controller = new AbortController();

	void fetch(SERVER_FRIENDLY_NAME_ENDPOINT, { signal: controller.signal })
		.then(async (response) => {
			if (!response.ok) return;

			const result = (await response.json()) as ServerFriendlyNameResponse;
			const friendlyName = result.data?.name?.trim();
			if (result.success && friendlyName) serverName.set(friendlyName);
		})
		.catch(() => {
			// Keep the default name when the endpoint is unavailable.
		});

	return () => controller.abort();
}
