<script lang="ts">
	import MoonIcon from '@iconify-svelte/mynaui/moon';
	import SunMediumIcon from '@iconify-svelte/mynaui/sun-medium';
	import { onMount } from 'svelte';
	import {
		applyTheme,
		getTheme,
		onStoredThemeChange,
		setTheme,
		type Theme
	} from './theme';

	let theme: Theme = 'light';

	onMount(() => {
		theme = applyTheme(getTheme());

		return onStoredThemeChange((nextTheme) => {
			theme = applyTheme(nextTheme);
		});
	});

	function toggleTheme() {
		theme = setTheme(theme === 'dark' ? 'light' : 'dark');
	}
</script>

<button
	type="button"
	class="btn btn-ghost btn-circle"
	aria-label={theme === 'dark' ? 'Switch to light theme' : 'Switch to dark theme'}
	title={theme === 'dark' ? 'Light theme' : 'Dark theme'}
	onclick={toggleTheme}
>
	{#if theme === 'dark'}
		<SunMediumIcon height="1.25em" aria-hidden="true" />
	{:else}
		<MoonIcon height="1.25em" aria-hidden="true" />
	{/if}
</button>
