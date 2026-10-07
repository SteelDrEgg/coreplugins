<script lang="ts">
	import Play from '@iconify-svelte/mynaui/play';
	import Stop from '@iconify-svelte/mynaui/stop';
	import * as m from '$lib/paraglide/messages.js';
	import { locale } from '$lib/utils/locale';
	import { canAct } from '$lib/utils/services';
	import type { ServiceRow, ServiceAction } from '$lib/utils/types';
	let { row, disabled = false, pending = null, onaction }: { row: ServiceRow; disabled?: boolean; pending?: ServiceAction | null; onaction: (row: ServiceRow, action: ServiceAction) => void } = $props();
	let options = $derived({ locale: $locale });
</script>

<div class="flex flex-wrap items-center gap-1">
	{#if canAct(row, 'start')}
		<button type="button" class="btn btn-sm" disabled={disabled} onclick={() => onaction(row, 'start')} aria-label={`${m.start({}, options)} ${row.name}`}>
			{#if pending === 'start'}<span class="loading loading-spinner loading-xs" aria-hidden="true"></span>{:else}<Play class="size-4" aria-hidden="true" />{/if}
			{m.start({}, options)}
		</button>
	{/if}
	{#if canAct(row, 'restart')}
		<button type="button" class="btn btn-ghost btn-sm" disabled={disabled} onclick={() => onaction(row, 'restart')} aria-label={`${m.restart({}, options)} ${row.name}`}>
			{#if pending === 'restart'}<span class="loading loading-spinner loading-xs" aria-hidden="true"></span>{/if}
			{m.restart({}, options)}
		</button>
	{/if}
	{#if canAct(row, 'stop')}
		<button type="button" class="btn btn-ghost btn-sm" disabled={disabled} onclick={() => onaction(row, 'stop')} aria-label={`${m.stop({}, options)} ${row.name}`}>
			{#if pending === 'stop'}<span class="loading loading-spinner loading-xs" aria-hidden="true"></span>{:else}<Stop class="size-4" aria-hidden="true" />{/if}
			{m.stop({}, options)}
		</button>
	{/if}
</div>
