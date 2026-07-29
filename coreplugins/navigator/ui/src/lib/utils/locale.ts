import languageDefinitions from '../../../locale/languages.json';
import {
	getLocale,
	getTextDirection,
	locales,
	setLocale,
	type Locale
} from '$lib/paraglide/runtime.js';
import type { LanguageDefinition } from './types';

export type AvailableLanguage = LanguageDefinition & {
	code: Locale;
};

export const availableLanguages: AvailableLanguage[] = locales.map((code) => ({
	code,
	...(languageDefinitions[code as keyof typeof languageDefinitions] ?? {
		name: code,
		nativeName: code
	})
}));

export function currentLocale(): Locale {
	return getLocale();
}

export function changeLocale(locale: Locale): void | Promise<void> {
	return setLocale(locale);
}

export function applyDocumentLocale(): void {
	const locale = getLocale();
	document.documentElement.lang = locale;
	document.documentElement.dir = getTextDirection(locale);
}
