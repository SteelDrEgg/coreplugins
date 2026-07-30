declare global {
	namespace App {}

	interface Window {
		webSDK?: {
			getTheme(): 'light' | 'dark';
			onThemeChange(listener: (theme: 'light' | 'dark') => void): () => void;
		};
	}
}

export {};
