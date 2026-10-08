export type Protection = 'identity' | 'passphrase';

export type SecretInfo = {
	name: string;
	description?: string;
	allowed_plugins: string[];
	updated_at: string;
	protection: Protection;
};

export type MessageKind = 'info' | 'success' | 'warning' | 'error';

export type BannerMessage = {
	text: string;
	kind: MessageKind;
};
