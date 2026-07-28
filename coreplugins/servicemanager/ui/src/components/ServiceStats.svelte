<script lang="ts">
	import type { DiscoveredService } from '$utils/types';

	let { services }: { services: DiscoveredService[] } = $props();

	function count(status: string): number {
		return services.filter(
			(service) => String(service.status || 'discovered').trim().toLowerCase() === status
		).length;
	}

	let running = $derived(count('running'));
	let degraded = $derived(count('degraded'));
	let inactive = $derived(count('discovered'));
	let failed = $derived(count('failed'));
</script>

<section class="grid min-w-0 gap-3 sm:grid-cols-2 lg:grid-cols-5" aria-label="Service status">
	<div class="stat min-w-0 rounded-lg border border-base-300 bg-base-100 shadow-sm">
		<div class="stat-title">Discovered</div>
		<div class="stat-value text-2xl">{services.length}</div>
	</div>
	<div class="stat min-w-0 rounded-lg border border-base-300 bg-base-100 shadow-sm">
		<div class="stat-title">Running</div>
		<div class="stat-value text-2xl">{running}</div>
	</div>
	<div class="stat min-w-0 rounded-lg border border-base-300 bg-base-100 shadow-sm">
		<div class="stat-title">Degraded</div>
		<div class="stat-value text-2xl">{degraded}</div>
	</div>
	<div class="stat min-w-0 rounded-lg border border-base-300 bg-base-100 shadow-sm">
		<div class="stat-title">Inactive</div>
		<div class="stat-value text-2xl">{inactive}</div>
	</div>
	<div class="stat min-w-0 rounded-lg border border-base-300 bg-base-100 shadow-sm">
		<div class="stat-title">Failed</div>
		<div class="stat-value text-2xl">{failed}</div>
	</div>
</section>
