import { writable } from 'svelte/store';
import {
	getLocale,
	getTextDirection,
	setLocale,
	toLocale,
	baseLocale,
	type Locale
} from '$lib/paraglide/runtime.js';

export const LANGUAGE_STORAGE_KEY = 'arupa.language';
export const locale = writable<Locale>(getLocale());

function applyDocumentLocale(next: Locale): void {
	document.documentElement.lang = next;
	document.documentElement.dir = getTextDirection(next);
}

export function initializeLocale(): () => void {
	let current = getLocale();
	locale.set(current);
	applyDocumentLocale(current);

	const handleStorage = (event: StorageEvent) => {
		if (event.key !== LANGUAGE_STORAGE_KEY && event.key !== null) return;
		let stored: string | null = null;
		try { stored = localStorage.getItem(LANGUAGE_STORAGE_KEY); } catch { /* Use browser language. */ }
		const preferred = navigator.languages.map((language) => toLocale(language) ?? toLocale(language.split('-')[0])).find((value) => value !== undefined);
		const next = toLocale(stored) ?? preferred ?? baseLocale;
		if (next === current) return;
		current = next;
		void setLocale(next, { reload: false });
		locale.set(next);
		applyDocumentLocale(next);
	};

	window.addEventListener('storage', handleStorage);
	return () => window.removeEventListener('storage', handleStorage);
}
