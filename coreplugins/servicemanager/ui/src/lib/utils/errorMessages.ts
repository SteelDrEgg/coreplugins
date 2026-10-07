import * as m from '$lib/paraglide/messages.js';

export const serviceAPIErrorMessages: Readonly<Partial<Record<number, typeof m.session_expired>>> = {
	401: m.session_expired,
	403: m.service_access_denied,
	404: m.service_not_enabled
};
