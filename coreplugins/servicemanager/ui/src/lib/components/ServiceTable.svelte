<script lang="ts">
	import Play from '@iconify-svelte/mynaui/play';
	import Refresh from '@iconify-svelte/mynaui/refresh';
	import Stop from '@iconify-svelte/mynaui/stop';
	import StatusBadge from './StatusBadge.svelte';
	import * as m from '$lib/paraglide/messages.js';
	import { locale } from '$lib/utils/locale';
	import type { DiscoveredService, ServiceAction } from '$lib/utils/types';

	let {
		services,
		busy,
		onAction
	}: {
		services: DiscoveredService[];
		busy: boolean;
		onAction: (action: ServiceAction, name: string) => void;
	} = $props();
	let localeOptions = $derived({ locale: $locale });

	const fallbackIcon = '/services/icon/puzzle.svg';

	function normalizeStatus(status: unknown): string {
		return String(status || 'discovered').trim().toLowerCase();
	}

	function canRun(action: ServiceAction, statusValue: unknown): boolean {
		const status = normalizeStatus(statusValue);
		if (status === 'starting' || status === 'stopping') return false;
		if (action === 'start') return status === 'discovered' || status === 'failed';
		return status === 'running' || status === 'degraded';
	}

	function metadataString(service: DiscoveredService, ...keys: string[]): string {
		for (const key of keys) {
			const value = service.metadata?.[key];
			if (typeof value === 'string' && value.trim()) return value.trim();
		}
		return '';
	}

	function displayName(service: DiscoveredService): string {
		return metadataString(service, 'DisplayName', 'display_name') || service.name;
	}

	function iconURL(service: DiscoveredService): string {
		return metadataString(service, 'Icon', 'icon') || fallbackIcon;
	}

	function iconStyle(service: DiscoveredService): string {
		const url = iconURL(service).replaceAll('\\', '\\\\').replaceAll('"', '\\"').replaceAll('\n', '');
		return `url("${url}")`;
	}

	function packagePath(service: DiscoveredService): string {
		return String(service.package_path || '-').replace(/^(?:\.?[\\/])?services(?:[\\/]+|$)/, '') || '-';
	}
</script>

{#if services.length === 0}
	<div class="grid min-h-56 place-items-center p-6 text-center text-base-content/60">
		{m.no_services_match({}, localeOptions)}
	</div>
{:else}
	<div class="overflow-x-auto">
		<table class="table table-zebra min-w-[760px]">
			<thead>
				<tr>
					<th>{m.service({}, localeOptions)}</th>
					<th>{m.status({}, localeOptions)}</th>
					<th>{m.type({}, localeOptions)}</th>
					<th>{m.package({}, localeOptions)}</th>
					<th>{m.actions({}, localeOptions)}</th>
				</tr>
			</thead>
			<tbody>
				{#each services as service (service.name)}
					<tr>
						<td>
							<div class="flex min-w-52 items-center gap-3">
								<div
									class="grid size-8 shrink-0 place-items-center rounded-lg bg-accent/20 text-accent-content"
									aria-hidden="true"
								>
									<span
										class="service-icon"
										style:--service-icon-url={iconStyle(service)}
									></span>
								</div>
								<div class="min-w-0">
									<div class="truncate font-semibold" title={service.name}>
										{displayName(service)}
									</div>
									<div class="truncate text-xs text-base-content/60">
										{service.name} v{service.version || '-'}
									</div>
								</div>
							</div>
						</td>
						<td><StatusBadge status={service.status} /></td>
						<td><span class="badge badge-neutral badge-sm">{service.type || '-'}</span></td>
						<td
							class="max-w-64 truncate font-mono text-xs text-base-content/70"
							title={packagePath(service)}
						>
							{packagePath(service)}
						</td>
						<td>
							<div class="flex min-w-52 flex-wrap items-center gap-2">
								<button
									class="btn btn-sm"
									type="button"
									disabled={busy || !canRun('start', service.status)}
									onclick={() => onAction('start', service.name)}
								>
									<Play class="size-4" aria-hidden="true" />
									{m.start({}, localeOptions)}
								</button>
								<button
									class="btn btn-sm"
									type="button"
									disabled={busy || !canRun('restart', service.status)}
									onclick={() => onAction('restart', service.name)}
								>
									<Refresh class="size-4" aria-hidden="true" />
									{m.restart({}, localeOptions)}
								</button>
								<button
									class="btn btn-sm"
									type="button"
									disabled={busy || !canRun('stop', service.status)}
									onclick={() => onAction('stop', service.name)}
								>
									<Stop class="size-4" aria-hidden="true" />
									{m.stop({}, localeOptions)}
								</button>
							</div>
						</td>
					</tr>
				{/each}
			</tbody>
		</table>
	</div>
{/if}
