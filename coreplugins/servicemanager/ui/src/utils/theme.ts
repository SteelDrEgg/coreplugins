export type Theme = 'light' | 'dark';

function normalizeTheme(value: unknown): Theme {
	return value === 'dark' ? 'dark' : 'light';
}

function storedTheme(): Theme {
	try {
		return normalizeTheme(window.localStorage.getItem('arupa.theme'));
	} catch {
		return 'light';
	}
}

export function applyTheme(theme: Theme): void {
	document.body.classList.remove('light', 'dark');
	document.body.classList.add(theme);
	document.documentElement.style.colorScheme = theme;
}

export function connectTheme(): () => void {
	const sdk = window.webSDK;
	if (sdk) {
		applyTheme(normalizeTheme(sdk.getTheme()));
		return sdk.onThemeChange((theme) => applyTheme(normalizeTheme(theme)));
	}

	applyTheme(storedTheme());
	const handleStorage = (event: StorageEvent) => {
		if (event.key === 'arupa.theme') applyTheme(normalizeTheme(event.newValue));
	};
	window.addEventListener('storage', handleStorage);
	return () => window.removeEventListener('storage', handleStorage);
}
