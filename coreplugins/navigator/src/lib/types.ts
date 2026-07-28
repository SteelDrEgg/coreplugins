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
	languages: string[];
};

export type LanguageDefinition = {
	name?: string;
	nativeName?: string;
};
