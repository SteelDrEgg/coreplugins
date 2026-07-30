export type Theme = 'light' | 'dark';

const THEME_KEY = 'arupa.theme';

function read(key: string): string {
	try {
		return window.localStorage.getItem(key) || '';
	} catch {
		return '';
	}
}

function write(key: string, value: string): void {
	try {
		window.localStorage.setItem(key, value);
	} catch {
		// The current document can still apply the preference.
	}
}

export function getTheme(): Theme {
	return read(THEME_KEY) === 'dark' ? 'dark' : 'light';
}

export function setTheme(theme: Theme): Theme {
	const normalized = theme === 'dark' ? 'dark' : 'light';
	document.body.classList.remove('light', 'dark');
	document.body.classList.add(normalized);
	document.documentElement.style.colorScheme = normalized;
	write(THEME_KEY, normalized);
	return normalized;
}

export function subscribeTheme(onTheme: (theme: Theme) => void): () => void {
	const listener = (event: StorageEvent) => {
		if (event.key === THEME_KEY) onTheme(event.newValue === 'dark' ? 'dark' : 'light');
	};
	window.addEventListener('storage', listener);
	return () => window.removeEventListener('storage', listener);
}
