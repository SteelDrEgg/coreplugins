<script lang="ts">
	import { onMount } from 'svelte';
	import '$lib/assets/app.css';
	import { initializeLocale } from '$lib/utils/locale';
	import { connectTheme } from '$lib/utils/theme';
	let { children } = $props();
	onMount(() => {
		const disconnectLocale = initializeLocale();
		const disconnectTheme = connectTheme();
		return () => { disconnectLocale(); disconnectTheme(); };
	});
</script>

{@render children()}
