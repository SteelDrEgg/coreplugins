<script lang="ts">
	import { onMount } from 'svelte';
	import { initializeLocale } from '$lib/locale';
	import { initializeServerName } from '$lib/serverName';
	import { initializeTheme } from '$lib/theme';
	import '../app.css';

	let { children } = $props();

	onMount(() => {
		const disposeLocale = initializeLocale();
		const disposeServerName = initializeServerName();
		const disposeTheme = initializeTheme();

		return () => {
			disposeLocale();
			disposeServerName();
			disposeTheme();
		};
	});
</script>

{@render children()}
