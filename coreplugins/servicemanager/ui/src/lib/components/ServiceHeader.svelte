<script lang="ts">
	import PuzzleSolid from '@iconify-svelte/mynaui/puzzle-solid';
	import Refresh from '@iconify-svelte/mynaui/refresh';
	import * as m from '$lib/paraglide/messages.js';
	import { locale } from '$lib/utils/locale';

	let {
		serviceDir,
		busy,
		onRefresh
	}: {
		serviceDir: string;
		busy: boolean;
		onRefresh: () => void;
	} = $props();
	let localeOptions = $derived({ locale: $locale });
</script>

<header class="flex min-w-0 flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
	<div class="flex min-w-0 items-center gap-3">
		<div
			class="grid size-10 shrink-0 place-items-center rounded-lg bg-primary text-primary-content"
			aria-hidden="true"
		>
			<PuzzleSolid class="size-5" />
		</div>
		<div class="min-w-0">
			<h1 class="text-xl font-semibold">{m.services({}, localeOptions)}</h1>
			<p class="truncate text-sm text-base-content/60">
				<!--{serviceDir || m.no_service_directory({}, localeOptions)}-->
				{m.page_description({}, localeOptions)}
			</p>
		</div>
	</div>
	<button
		class="btn btn-primary w-full sm:w-auto"
		type="button"
		disabled={busy}
		onclick={onRefresh}
	>
		<Refresh class={`size-4 ${busy ? 'animate-spin' : ''}`} aria-hidden="true" />
		{m.refresh({}, localeOptions)}
	</button>
</header>
