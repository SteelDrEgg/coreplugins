<script lang="ts">
	import { onMount } from 'svelte';
	import X from '@iconify-svelte/mynaui/x';
	import * as m from '$lib/paraglide/messages.js';
	import { locale } from '$lib/utils/locale';
	import { displayName } from '$lib/utils/services';
	import type { ServiceRow, ServiceAction, AccessPolicy } from '$lib/utils/types';
	import ServiceStatus from './ServiceStatus.svelte';
	import ServiceActions from './ServiceActions.svelte';
	import ServiceSettings from './ServiceSettings.svelte';
	let { row, disabled, pending, onaction, onclose, onupdated }: { row: ServiceRow; disabled: boolean; pending: ServiceAction | null; onaction: (row: ServiceRow, action: ServiceAction) => void; onclose: () => void; onupdated: () => Promise<boolean> } = $props();
	let dialog: HTMLDialogElement;
	let tab: 'overview' | 'routes' | 'transports' | 'settings' = $state('overview');
	let settingsOpened = $state(false);
	const tabs = ['overview', 'routes', 'transports', 'settings'] as const;
	function selectTab(next: typeof tab) { tab = next; if (next === 'settings') settingsOpened = true; }
	function moveTab(event: KeyboardEvent, index: number) {
		if (!['ArrowLeft', 'ArrowRight', 'Home', 'End'].includes(event.key)) return;
		event.preventDefault();
		const next = event.key === 'Home' ? 0 : event.key === 'End' ? tabs.length - 1 : (index + (event.key === 'ArrowRight' ? 1 : -1) + tabs.length) % tabs.length;
		selectTab(tabs[next]);
		(event.currentTarget as HTMLElement).parentElement?.querySelector<HTMLButtonElement>(`#detail-tab-${tabs[next]}`)?.focus();
	}
	let options = $derived({ locale: $locale });
	let routes = $derived(row.instances.flatMap((instance) => (instance.routes ?? []).map((route) => ({ instance: instance.instance_id, route }))));
	let transports = $derived(row.instances.flatMap((instance) => (instance.transports ?? []).map((transport) => ({ instance: instance.instance_id, transport }))));
	function date(value: string) { const parsed = new Date(value); return Number.isNaN(parsed.getTime()) ? value : parsed.toLocaleString($locale); }
	onMount(() => dialog.showModal());
</script>

{#snippet policy(access: AccessPolicy | undefined)}
	<div class="space-y-1 text-xs">
		{#if access?.require_auth}<p>{m.authenticated({}, options)}</p>{/if}
		{#if access?.groups?.length}<p>{m.groups({}, options)}: {access.groups.join(', ')}</p>{/if}
		{#if !access?.require_auth && !access?.groups?.length}<p class="text-base-content/50">{m.public_access({}, options)}</p>{/if}
	</div>
{/snippet}

<dialog bind:this={dialog} class="modal" onclose={onclose} aria-labelledby="service-detail-title">
	<div class="modal-box max-h-[90dvh] w-[calc(100%-2rem)] max-w-5xl space-y-5 p-4 sm:p-6">
		<header class="flex items-start justify-between gap-4">
			<div class="min-w-0"><h2 id="service-detail-title" class="truncate text-xl font-semibold">{displayName(row)}</h2><div class="mt-2 flex flex-wrap items-center gap-2"><span class="text-sm text-base-content/50">{row.name}</span><ServiceStatus status={row.status} /></div></div>
			<form method="dialog"><button class="btn btn-ghost btn-square btn-sm" aria-label={m.close({}, options)}><X class="size-5" aria-hidden="true" /></button></form>
		</header>
		<ServiceActions {row} {disabled} {pending} {onaction} />
		{#if row.versionMismatch}<div class="alert alert-warning alert-soft" role="status">{m.version_hint({}, options)}</div>{/if}
		{#if !row.disk}<div class="alert alert-warning alert-soft" role="status">{m.missing_package({}, options)}</div>{/if}
		<div class="tabs tabs-border overflow-x-auto" role="tablist" aria-label={m.details({}, options)}>
			{#each tabs as name, index}
				<button type="button" id={`detail-tab-${name}`} role="tab" aria-selected={tab === name} aria-controls={`detail-panel-${name}`} tabindex={tab === name ? 0 : -1} class="tab whitespace-nowrap {tab === name ? 'tab-active' : ''}" onclick={() => selectTab(name)} onkeydown={(event) => moveTab(event, index)}>
					{name === 'overview' ? m.overview({}, options) : name === 'routes' ? m.routes({}, options) : name === 'transports' ? m.transports({}, options) : m.settings({}, options)}
					{#if name === 'routes'}<span class="ml-2 text-xs opacity-50">{routes.length}</span>{:else if name === 'transports'}<span class="ml-2 text-xs opacity-50">{transports.length}</span>{/if}
				</button>
			{/each}
		</div>
		<div id="detail-panel-overview" role="tabpanel" tabindex="0" aria-labelledby="detail-tab-overview" hidden={tab !== 'overview'} class="space-y-5">
			<dl class="grid grid-cols-1 gap-4 text-sm sm:grid-cols-2">
				<div><dt class="text-base-content/50">{m.disk_version({}, options)}</dt><dd class="mt-1 font-mono">{row.disk?.version ?? m.unavailable({}, options)}</dd></div>
				<div><dt class="text-base-content/50">{m.running_version({}, options)}</dt><dd class="mt-1 font-mono">{row.instances.map((item) => item.version).join(', ') || m.not_set({}, options)}</dd></div>
				<div><dt class="text-base-content/50">{m.type({}, options)}</dt><dd class="mt-1">{row.disk?.type ?? row.instances[0]?.type}</dd></div>
				{#if row.disk}
					<div><dt class="text-base-content/50">{m.contract_version({}, options)}</dt><dd class="mt-1">{row.disk.contract_version}</dd></div>
					<div class="sm:col-span-2"><dt class="text-base-content/50">{m.package_path({}, options)}</dt><dd class="mt-1 break-all font-mono text-xs">{row.disk.package_path}</dd></div>
					{#if row.disk.command}<div class="sm:col-span-2"><dt class="text-base-content/50">{m.command({}, options)}</dt><dd class="mt-1 break-all font-mono text-xs">{row.disk.command}</dd></div>{/if}
					<div><dt class="text-base-content/50">{m.retry_count({}, options)}</dt><dd class="mt-1">{row.disk.retry_count}</dd></div>
					{#if row.disk.next_retry_at}<div><dt class="text-base-content/50">{m.next_retry({}, options)}</dt><dd class="mt-1">{date(row.disk.next_retry_at)}</dd></div>{/if}
					{#if row.disk.last_exit}<div class="sm:col-span-2"><dt class="text-base-content/50">{m.last_exit({}, options)}</dt><dd class="mt-1">{date(row.disk.last_exit.exited_at)} · {m.exit_code({}, options)} {row.disk.last_exit.exit_code}{#if row.disk.last_exit.error}<p class="mt-1 break-all text-error">{row.disk.last_exit.error}</p>{/if}</dd></div>{/if}
				{/if}
				{#each row.instances as instance (instance.instance_id)}<div class="sm:col-span-2"><dt class="text-base-content/50">{m.instance({}, options)}: {instance.instance_id}</dt><dd class="mt-1 break-all font-mono text-xs">{m.runtime_path({}, options)}: {instance.path}</dd></div>{/each}
			</dl>
			{#if row.disk?.last_error}<div class="alert alert-error alert-soft" role="alert"><div><h3 class="font-semibold">{m.last_error({}, options)}</h3><p class="mt-1 break-all text-sm">{row.disk.last_error}</p></div></div>{/if}
			{#if row.disk?.metadata && Object.keys(row.disk.metadata).length}<div><h3 class="mb-2 text-sm font-semibold">{m.metadata({}, options)}</h3><pre class="overflow-x-auto rounded-box bg-base-200 p-4 text-xs">{JSON.stringify(row.disk.metadata, null, 2)}</pre></div>{/if}
		</div>
		<div id="detail-panel-routes" role="tabpanel" tabindex="0" aria-labelledby="detail-tab-routes" hidden={tab !== 'routes'}>
			{#if !row.instances.length}<p class="py-6 text-sm text-base-content/60">{m.no_runtime({}, options)}</p>{:else if !routes.length}<p class="py-6 text-sm text-base-content/60">{m.no_routes({}, options)}</p>{:else}
				<div class="overflow-x-auto"><table class="table table-sm"><thead><tr><th>{m.route_id({}, options)}</th><th>{m.binding({}, options)}</th><th>{m.transport_id({}, options)}</th><th>{m.access({}, options)}</th></tr></thead><tbody>
					{#each routes as { instance, route } (`${instance}:${route.id}`)}<tr>
						<td class="font-mono text-xs">{route.id}<div class="mt-1 text-base-content/50">{instance}</div></td>
						<td><div class="break-all font-mono text-xs">{route.http ? `${route.http.method || m.all_methods({}, options)} ${route.http.pattern}` : route.socket_io?.namespace}</div>
							{#if route.socket_io}<p class="mt-1 text-xs">{m.events({}, options)}: {(route.socket_io.events ?? []).join(', ')}</p>{/if}
							{#if route.http?.rewrite?.prefix || route.http?.rewrite?.location}<p class="mt-1 text-xs text-base-content/50">{m.rewrite({}, options)}: {route.http.rewrite.prefix ? m.rewrite_prefix({}, options) : ''}{route.http.rewrite.location ? ` · ${m.rewrite_location({}, options)}` : ''}</p>{/if}
						</td><td class="font-mono text-xs">{route.transport}</td><td>{@render policy(route.http?.access ?? route.socket_io?.access)}
							{#if route.socket_io?.event_access && Object.keys(route.socket_io.event_access).length}<div class="mt-2 text-xs"><p class="font-semibold">{m.event_access({}, options)}</p>{#each Object.entries(route.socket_io.event_access) as [event, access]}<div class="mt-2"><p class="font-mono">{event}</p>{@render policy(access)}</div>{/each}</div>{/if}
						</td>
					</tr>{/each}
				</tbody></table></div>
			{/if}
		</div>
		<div id="detail-panel-transports" role="tabpanel" tabindex="0" aria-labelledby="detail-tab-transports" hidden={tab !== 'transports'}>
			{#if !row.instances.length}<p class="py-6 text-sm text-base-content/60">{m.no_runtime({}, options)}</p>{:else if !transports.length}<p class="py-6 text-sm text-base-content/60">{m.no_transports({}, options)}</p>{:else}
				<div class="overflow-x-auto"><table class="table table-sm"><thead><tr><th>{m.transport_id({}, options)}</th><th>{m.type({}, options)}</th><th>{m.details({}, options)}</th></tr></thead><tbody>
					{#each transports as { instance, transport } (`${instance}:${transport.id}`)}<tr><td class="font-mono text-xs">{transport.id}<div class="mt-1 text-base-content/50">{instance}</div></td><td><span class="badge badge-ghost badge-sm">{transport.type}</span></td><td class="break-all text-xs">
						{#if transport.source}<p>{m.source({}, options)}: <span class="font-mono">{transport.source}</span></p>{/if}
						{#if transport.proxy}<p>{m.proxy_target({}, options)}: <span class="font-mono">{transport.proxy.network} · {transport.proxy.scheme || 'http'} · {transport.proxy.address || m.not_set({}, options)}</span></p>{/if}
					</td></tr>{/each}
				</tbody></table></div>
			{/if}
		</div>
		<div id="detail-panel-settings" role="tabpanel" tabindex="0" aria-labelledby="detail-tab-settings" hidden={tab !== 'settings'}>
			{#if settingsOpened}<ServiceSettings name={row.name} effective={row.disk?.config} {disabled} {onupdated} />{/if}
		</div>
	</div>
	<form method="dialog" class="modal-backdrop"><button>{m.close({}, options)}</button></form>
</dialog>
