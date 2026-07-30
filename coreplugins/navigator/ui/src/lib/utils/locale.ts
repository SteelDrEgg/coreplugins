import { writable } from 'svelte/store';
import languageDefinitions from '../../../locale/languages.json';
import {
	getLocale,
	getTextDirection,
	locales,
	setLocale,
	toLocale,
	type Locale
} from '$lib/paraglide/runtime.js';
import type { LanguageDefinition } from './types';

export const LANGUAGE_STORAGE_KEY = 'arupa.language';

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

let activeLocale = getLocale();
export const locale = writable<Locale>(activeLocale);

function applyLocale(next: Locale, persist: boolean): void {
	activeLocale = next;
	if (persist) void setLocale(next, { reload: false });
	locale.set(next);
	document.documentElement.lang = next;
	document.documentElement.dir = getTextDirection(next);
}

export function changeLocale(next: Locale): void {
	if (next === activeLocale) return;
	applyLocale(next, true);
}

export function initializeLocale(): () => void {
	applyLocale(getLocale(), false);

	const handleStorage = (event: StorageEvent) => {
		if (event.key !== LANGUAGE_STORAGE_KEY && event.key !== null) return;
		const next = toLocale(event.newValue) ?? getLocale();
		if (next === activeLocale) return;
		applyLocale(next, true);
	};

	window.addEventListener('storage', handleStorage);
	return () => window.removeEventListener('storage', handleStorage);
}
