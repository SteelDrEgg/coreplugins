<script lang="ts">
	import Search from '@iconify-svelte/mynaui/search';
	import ServiceTable from './ServiceTable.svelte';
	import * as m from '$lib/paraglide/messages.js';
	import { locale } from '$lib/utils/locale';
	import type { DiscoveredService, ServiceAction } from '$lib/utils/types';

	let {
		services,
		loaded,
		busy,
		onAction
	}: {
		services: DiscoveredService[];
		loaded: boolean;
		busy: boolean;
		onAction: (action: ServiceAction, name: string) => void;
	} = $props();

	let query = $state('');
	let statusFilter = $state('all');
	let localeOptions = $derived({ locale: $locale });

	function normalizeStatus(status: unknown): string {
		return String(status || 'discovered').trim().toLowerCase();
	}

	let filteredServices = $derived.by(() => {
		const normalizedQuery = query.trim().toLowerCase();
		return services.filter((service) => {
			const status = normalizeStatus(service.status);
			if (statusFilter === 'running' && status !== 'running') return false;
			if (statusFilter === 'degraded' && status !== 'degraded') return false;
			if (statusFilter === 'inactive' && status !== 'discovered') return false;
			if (statusFilter === 'busy' && status !== 'starting' && status !== 'stopping') return false;
			if (statusFilter === 'failed' && status !== 'failed') return false;
			if (!normalizedQuery) return true;

			const displayName =
				typeof service.metadata?.DisplayName === 'string' ? service.metadata.DisplayName : '';
			return [
				service.name,
				service.version,
				service.type,
				service.package_path,
				service.command,
				displayName
			]
				.join(' ')
				.toLowerCase()
				.includes(normalizedQuery);
		});
	});
</script>

<div class="card min-w-0 rounded-lg border border-base-300 bg-base-100 shadow-sm">
	<div
		class="flex min-w-0 flex-col gap-3 border-b border-base-300 p-4 sm:flex-row sm:items-center sm:justify-between"
	>
		<h2 class="font-semibold">{m.service_catalog({}, localeOptions)}</h2>
		<div class="flex min-w-0 flex-col gap-2 sm:flex-row">
			<label class="input w-full sm:w-64">
				<Search class="size-4 opacity-50" aria-hidden="true" />
				<input
					bind:value={query}
					type="search"
					placeholder={m.search_name_or_package({}, localeOptions)}
				/>
			</label>
			<select
				class="select w-full sm:w-40"
				bind:value={statusFilter}
				aria-label={m.status_filter({}, localeOptions)}
			>
				<option value="all">{m.all({}, localeOptions)}</option>
				<option value="running">{m.running({}, localeOptions)}</option>
				<option value="degraded">{m.degraded({}, localeOptions)}</option>
				<option value="inactive">{m.inactive({}, localeOptions)}</option>
				<option value="busy">{m.busy({}, localeOptions)}</option>
				<option value="failed">{m.failed({}, localeOptions)}</option>
			</select>
		</div>
	</div>
	<div class="min-w-0">
		{#if !loaded}
			<div class="grid min-h-56 place-items-center p-6 text-base-content/60">
				<span
					class="loading loading-spinner loading-md"
					aria-label={m.loading_services({}, localeOptions)}
				></span>
			</div>
		{:else}
			<ServiceTable services={filteredServices} {busy} {onAction} />
		{/if}
	</div>
</div>
