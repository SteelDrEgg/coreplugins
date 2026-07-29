<script lang="ts">
	import Servers from '@iconify-svelte/mynaui/servers';
	import * as m from '$lib/paraglide/messages.js';
	import { locale } from '$lib/utils/locale';
	import type { RunningService } from '$lib/utils/types';

	let { services }: { services: RunningService[] } = $props();
	let localeOptions = $derived({ locale: $locale });

	function serviceName(service: RunningService): string {
		return service.name || service.instance_id || m.unnamed_service({}, localeOptions);
	}

	function packagePath(service: RunningService): string {
		return String(service.path || '-').replace(/^(?:\.?[\\/])?services(?:[\\/]+|$)/, '') || '-';
	}
</script>

{#if services.length === 0}
	<div class="grid min-h-32 place-items-center text-center text-base-content/60">
		{m.no_running_services({}, localeOptions)}
	</div>
{:else}
	{#each services as service, index (service.instance_id || `${service.name}-${index}`)}
		<div class="rounded-lg border border-base-300 p-3">
			<div class="flex items-center justify-between gap-2 font-semibold">
				<span class="flex min-w-0 items-center gap-2">
					<Servers class="size-4 shrink-0 text-base-content/50" aria-hidden="true" />
					<span class="truncate">{serviceName(service)}</span>
				</span>
				<span class="badge badge-success badge-sm">{service.type || '-'}</span>
			</div>
			<div class="mt-2 grid gap-1 text-xs text-base-content/60">
				<div class="truncate font-mono" title={packagePath(service)}>{packagePath(service)}</div>
				<div>
					{m.runtime_counts(
						{
							routes: service.routes?.length || 0,
							transports: service.transports?.length || 0
						},
						localeOptions
					)}
				</div>
			</div>
		</div>
	{/each}
{/if}
