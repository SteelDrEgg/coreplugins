<script lang="ts">
	import * as m from '$lib/paraglide/messages.js';
	import { locale } from '$lib/utils/locale';
	import type { NavigationEntry } from '$lib/utils/types';

	let {
		loading,
		activeLoaded,
		message,
		openedEntries,
		activeID,
		onrefresh,
		onload
	}: {
		loading: boolean;
		activeLoaded: boolean;
		message: string;
		openedEntries: NavigationEntry[];
		activeID: string;
		onrefresh: () => void;
		onload: (id: string) => void;
	} = $props();

	let localeOptions = $derived({ locale: $locale });
</script>

<main class="relative min-h-0 min-w-0 overflow-hidden bg-base-100">
	{#if loading || (activeID && !activeLoaded)}
		<div class="absolute inset-0 z-10 grid place-items-center bg-base-100 text-base-content/60">
			<span
				class="loading loading-spinner loading-md"
				aria-label={m.loading_page({}, localeOptions)}
			></span>
		</div>
	{/if}

	{#if message}
		<div
			class="absolute inset-x-4 top-4 z-20 flex items-center justify-between gap-3 rounded-lg border border-error/20 bg-base-100 p-4 text-sm text-error shadow-sm"
			role="status"
		>
			<span>{message}</span>
			<button class="btn btn-sm" type="button" onclick={onrefresh}
				>{m.retry({}, localeOptions)}</button
			>
		</div>
	{/if}

	{#each openedEntries as entry (entry.id)}
		<iframe
			class:hidden={entry.id !== activeID}
			class="absolute inset-0 size-full border-0 bg-base-100"
			title={entry.label}
			src={entry.href}
			onload={() => onload(entry.id)}
		></iframe>
	{/each}
</main>
