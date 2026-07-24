export type SecretMeta = {
	name: string;
	description?: string;
	allowed_plugins: string[];
	updated_at: string;
	encryption?: string;
};

export type MessageKind = 'info' | 'success' | 'warning' | 'error';

export type BannerMessage = {
	text: string;
	kind: MessageKind;
};
