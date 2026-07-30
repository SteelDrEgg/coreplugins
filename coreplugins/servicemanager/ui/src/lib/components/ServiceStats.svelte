<script lang="ts">
	import * as m from '$lib/paraglide/messages.js';
	import { locale } from '$lib/utils/locale';
	import type { DiscoveredService } from '$lib/utils/types';

	let { services }: { services: DiscoveredService[] } = $props();
	let localeOptions = $derived({ locale: $locale });

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

<section
	class="grid min-w-0 gap-3 sm:grid-cols-2 lg:grid-cols-5"
	aria-label={m.service_status({}, localeOptions)}
>
	<div class="stat min-w-0 rounded-lg border border-base-300 bg-base-100 shadow-sm">
		<div class="stat-title">{m.discovered({}, localeOptions)}</div>
		<div class="stat-value text-2xl">{services.length}</div>
	</div>
	<div class="stat min-w-0 rounded-lg border border-base-300 bg-base-100 shadow-sm">
		<div class="stat-title">{m.running({}, localeOptions)}</div>
		<div class="stat-value text-2xl">{running}</div>
	</div>
	<div class="stat min-w-0 rounded-lg border border-base-300 bg-base-100 shadow-sm">
		<div class="stat-title">{m.degraded({}, localeOptions)}</div>
		<div class="stat-value text-2xl">{degraded}</div>
	</div>
	<div class="stat min-w-0 rounded-lg border border-base-300 bg-base-100 shadow-sm">
		<div class="stat-title">{m.inactive({}, localeOptions)}</div>
		<div class="stat-value text-2xl">{inactive}</div>
	</div>
	<div class="stat min-w-0 rounded-lg border border-base-300 bg-base-100 shadow-sm">
		<div class="stat-title">{m.failed({}, localeOptions)}</div>
		<div class="stat-value text-2xl">{failed}</div>
	</div>
</section>
