<script lang="ts">
	import Search from '@iconify-svelte/mynaui/search';
	import ServiceTable from './ServiceTable.svelte';
	import type { DiscoveredService, ServiceAction } from '$utils/types';

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
		<h2 class="font-semibold">Service catalog</h2>
		<div class="flex min-w-0 flex-col gap-2 sm:flex-row">
			<label class="input w-full sm:w-64">
				<Search class="size-4 opacity-50" aria-hidden="true" />
				<input bind:value={query} type="search" placeholder="Search name or package" />
			</label>
			<select class="select w-full sm:w-40" bind:value={statusFilter} aria-label="Status filter">
				<option value="all">All</option>
				<option value="running">Running</option>
				<option value="degraded">Degraded</option>
				<option value="inactive">Inactive</option>
				<option value="busy">Busy</option>
				<option value="failed">Failed</option>
			</select>
		</div>
	</div>
	<div class="min-w-0">
		{#if !loaded}
			<div class="grid min-h-56 place-items-center p-6 text-base-content/60">
				<span class="loading loading-spinner loading-md" aria-label="Loading services"></span>
			</div>
		{:else}
			<ServiceTable services={filteredServices} {busy} {onAction} />
		{/if}
	</div>
</div>
