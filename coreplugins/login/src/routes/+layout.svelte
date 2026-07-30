<script lang="ts">
	import { onMount } from 'svelte';
	import { initializeLocale } from '$lib/locale';
	import '../app.css';

	let { children } = $props();

	onMount(initializeLocale);
</script>

{@render children()}
