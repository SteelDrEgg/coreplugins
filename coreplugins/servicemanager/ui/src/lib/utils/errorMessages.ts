export const serviceAPIErrorMessages: Readonly<Partial<Record<number, string>>> = {
	401: 'Your session has expired.',
	403: 'Service management access was denied.',
	404: 'Service management is not enabled.'
};
