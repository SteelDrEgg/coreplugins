export type Theme = 'light' | 'dark';

export const THEME_STORAGE_KEY = 'arupa.theme';

export function normalizeTheme(value: unknown): Theme {
	return value === 'dark' ? 'dark' : 'light';
}

export function getTheme(): Theme {
	try {
		return normalizeTheme(window.localStorage.getItem(THEME_STORAGE_KEY));
	} catch {
		return 'light';
	}
}

export function applyTheme(theme: Theme): Theme {
	document.body.classList.remove('light', 'dark');
	document.body.classList.add(theme);
	document.documentElement.style.colorScheme = theme;
	return theme;
}

export function setTheme(theme: Theme): Theme {
	const normalized = applyTheme(normalizeTheme(theme));
	try {
		window.localStorage.setItem(THEME_STORAGE_KEY, normalized);
	} catch {}
	return normalized;
}

export function onStoredThemeChange(listener: (theme: Theme) => void): () => void {
	const handleStorage = (event: StorageEvent) => {
		if (event.key === THEME_STORAGE_KEY) listener(normalizeTheme(event.newValue));
	};
	window.addEventListener('storage', handleStorage);
	return () => window.removeEventListener('storage', handleStorage);
}
