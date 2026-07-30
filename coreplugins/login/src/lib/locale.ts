import { writable } from 'svelte/store';
import en from '../../locale/en.json';
import zh from '../../locale/zh.json';

export const LANGUAGE_STORAGE_KEY = 'arupa.language';
export type Locale = 'en' | 'zh';
export type Messages = Omit<typeof en, '$schema'>;

const translations: Record<Locale, Messages> = { en, zh };

export const locale = writable<Locale>('en');

export function messagesFor(current: Locale): Messages {
	return translations[current];
}

function normalizeLocale(value: string | null | undefined): Locale | null {
	const language = String(value || '')
		.trim()
		.toLowerCase()
		.split('-')[0];
	return language === 'en' || language === 'zh' ? language : null;
}

function preferredLocale(): Locale {
	for (const language of window.navigator.languages || [window.navigator.language]) {
		const normalized = normalizeLocale(language);
		if (normalized) return normalized;
	}
	return 'en';
}

function storedLocale(): Locale {
	try {
		return normalizeLocale(window.localStorage.getItem(LANGUAGE_STORAGE_KEY)) ?? preferredLocale();
	} catch {
		return preferredLocale();
	}
}

function applyDocumentLocale(next: Locale): void {
	document.documentElement.lang = next;
	document.documentElement.dir = 'ltr';
}

export function initializeLocale(): () => void {
	let current = storedLocale();
	locale.set(current);
	applyDocumentLocale(current);

	const handleStorage = (event: StorageEvent) => {
		if (event.key !== LANGUAGE_STORAGE_KEY && event.key !== null) return;
		const next = normalizeLocale(event.newValue) ?? preferredLocale();
		if (next === current) return;
		current = next;
		locale.set(next);
		applyDocumentLocale(next);
	};

	window.addEventListener('storage', handleStorage);
	return () => window.removeEventListener('storage', handleStorage);
}
