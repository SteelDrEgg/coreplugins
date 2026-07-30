export type NavigationEntry = {
	id: string;
	service: string;
	route_id: string;
	href: string;
	label: string;
	icon?: string;
	icon_solid?: string;
};

export type EntriesPayload = {
	entries: NavigationEntry[];
};

export type NavigatorConfig = {
	icon: string;
	order: string[];
	hide: string[];
};

export type ServerFriendlyName = {
	name: string;
};

export type LanguageDefinition = {
	name: string;
	nativeName: string;
};

export type MessageKind = 'info' | 'success' | 'warning' | 'error';

export type BannerMessage = {
	text: string;
	kind: MessageKind;
};
