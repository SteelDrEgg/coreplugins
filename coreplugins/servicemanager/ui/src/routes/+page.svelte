<script lang="ts">
	import { onMount } from 'svelte';
	import Search from '@iconify-svelte/mynaui/search';
	import * as m from '$lib/paraglide/messages.js';
	import { locale } from '$lib/utils/locale';
	import { listDiscoveredServices, listRunningServices, getServiceDirectory, runServiceAction } from '$lib/utils/api';
	import { canAct, displayName, mergeServices } from '$lib/utils/services';
	import type { DiscoveredService, RunningService, ServiceRow, ServiceAction } from '$lib/utils/types';
	import ServiceStatus from '$lib/components/ServiceStatus.svelte';
	import ServiceActions from '$lib/components/ServiceActions.svelte';
	import ServiceDetails from '$lib/components/ServiceDetails.svelte';
	import ManagerSettings from '$lib/components/ManagerSettings.svelte';

	let discovered: DiscoveredService[] = $state([]);
	let running: RunningService[] = $state([]);
	let directory = $state('');
	let loaded = $state(false);
	let refreshing = $state(false);
	let catalogAvailable = $state(false);
	let runtimeAvailable = $state(false);
	let refreshError = $state('');
	let error = $state('');
	let notice: 'service_started' | 'service_stopped' | 'service_restarted' | 'action_refresh_failed' | 'self_action_done' | null = $state(null);
	let busyName = $state('');
	let pending: ServiceAction | null = $state(null);
	let query = $state('');
	let filter: 'all' | 'disk' | 'running' | 'attention' = $state('all');
	let selectedName = $state('');
	let managerSettings: 'directory' | 'defaults' | null = $state(null);
	let globalBusy = $state(false);
	let lastUpdated = $state<Date | null>(null);
	let alive = true;
	let options = $derived({ locale: $locale });
	let rows = $derived(mergeServices(discovered, running));
	let selected = $derived(rows.find((row) => row.name === selectedName));
	let counts = $derived({ all: rows.length, disk: discovered.length, running: rows.filter((row) => row.instances.length).length, attention: rows.filter((row) => row.versionMismatch || ['degraded', 'failed', 'backoff'].includes(row.status) || !row.disk).length });
	let filtered = $derived(rows.filter((row) => {
		if (filter === 'disk' && !row.disk) return false;
		if (filter === 'running' && !row.instances.length) return false;
		if (filter === 'attention' && !row.versionMismatch && !['degraded', 'failed', 'backoff'].includes(row.status) && row.disk) return false;
		const term = query.trim().toLowerCase();
		return !term || [row.name, displayName(row), row.disk?.package_path ?? '', ...row.instances.map((item) => item.path)].some((value) => value.toLowerCase().includes(term));
	}));
	let managerDisabled = $derived(!!busyName || refreshing || globalBusy);
	let actionsDisabled = $derived(managerDisabled || !catalogAvailable || !runtimeAvailable);

	function errorText(value: unknown): string { return value instanceof Error ? value.message : m.unknown_error({}, options); }

	async function refresh(): Promise<boolean> {
		if (refreshing) return false;
		refreshing = true;
		const results = await Promise.allSettled([listDiscoveredServices(), listRunningServices(), getServiceDirectory()]);
		if (!alive) return false;
		catalogAvailable = results[0].status === 'fulfilled';
		runtimeAvailable = results[1].status === 'fulfilled';
		if (results[0].status === 'fulfilled') discovered = results[0].value;
		if (results[1].status === 'fulfilled') running = results[1].value;
		if (results[2].status === 'fulfilled') directory = results[2].value.service_dir;
		const failed = results.slice(0, 2).find((result) => result.status === 'rejected');
		refreshError = failed?.status === 'rejected' ? errorText(failed.reason) : '';
		if (!failed) lastUpdated = new Date();
		loaded = true;
		refreshing = false;
		return !failed;
	}

	async function act(row: ServiceRow, action: ServiceAction) {
		if (actionsDisabled || !canAct(row, action)) return;
		if (action !== 'start') {
			let text: string = action === 'stop' ? m.confirm_stop({ name: row.name }, options) : m.confirm_restart({ name: row.name }, options);
			if (row.name === 'service-manager') text += `\n\n${m.confirm_self({}, options)}`;
			if (!window.confirm(text)) return;
		}
		busyName = row.name;
		pending = action;
		error = '';
		notice = null;
		try {
			await runServiceAction(action, row.name);
			if (row.name === 'service-manager') notice = 'self_action_done';
			else {
				notice = action === 'start' ? 'service_started' : action === 'stop' ? 'service_stopped' : 'service_restarted';
				if (!(await refresh())) notice = 'action_refresh_failed';
			}
		} catch (failure) {
			error = errorText(failure);
			// A failed launch may already be waiting for an automatic retry.
			await refresh();
		} finally { busyName = ''; pending = null; }
	}

	onMount(() => {
		alive = true;
		void refresh();
		const timer = window.setInterval(() => { if (!document.hidden && !busyName && !globalBusy) void refresh(); }, 5000);
		return () => { alive = false; window.clearInterval(timer); };
	});
</script>

<svelte:head>
	<title>{m.page_title({}, options)}</title>
	<meta name="description" content={m.page_description({}, options)} />
</svelte:head>

<main class="mx-auto min-h-screen w-full max-w-7xl space-y-5 px-4 py-6 sm:px-6 sm:py-8">
	<header class="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
		<div><h1 class="text-2xl font-semibold tracking-tight">{m.page_title({}, options)}</h1><p class="mt-1 text-sm text-base-content/60">{m.page_description({}, options)}</p></div>
		<div class="flex flex-wrap gap-2">
			<button class="btn" type="button" disabled={managerDisabled} onclick={() => managerSettings = 'directory'}>{m.directory_settings({}, options)}</button>
			<button class="btn" type="button" disabled={managerDisabled} onclick={() => managerSettings = 'defaults'}>{m.service_defaults({}, options)}</button>
			<button class="btn" type="button" disabled={managerDisabled} onclick={() => refresh()}>
				{#if refreshing}<span class="loading loading-spinner loading-sm" aria-hidden="true"></span>{/if}{m.refresh({}, options)}
			</button>
		</div>
	</header>
	{#if refreshError}<div class="alert alert-warning alert-soft" role="alert"><div><p>{m.refresh_failed({}, options)}</p><p class="mt-1 text-sm break-all">{refreshError}</p></div></div>{/if}
	{#if error}<div class="alert alert-error alert-soft" role="alert">{error}</div>{/if}
	{#if notice}<div class="alert alert-soft {notice === 'action_refresh_failed' ? 'alert-warning' : 'alert-success'}" role="status">{m[notice]({}, options)}</div>{/if}
	<section class="card card-border min-w-0 bg-base-100">
		<div class="flex flex-col gap-4 border-b border-base-300 p-4 sm:flex-row sm:items-center sm:justify-between">
			<div class="flex flex-wrap gap-1" aria-label={m.status({}, options)}>
				{#each ['all', 'disk', 'running', 'attention'] as choice}
					<button type="button" class="btn btn-sm {filter === choice ? 'btn-active' : 'btn-ghost'}" aria-pressed={filter === choice} onclick={() => filter = choice as typeof filter}>
						{choice === 'all' ? m.all_services({}, options) : choice === 'disk' ? m.on_disk({}, options) : choice === 'running' ? m.running({}, options) : m.attention({}, options)}
						<span class="badge badge-sm badge-ghost">{counts[choice as keyof typeof counts]}</span>
					</button>
				{/each}
			</div>
			<label class="input w-full sm:w-72"><Search class="size-4 shrink-0 opacity-50" aria-hidden="true" /><input type="search" bind:value={query} placeholder={m.search_services({}, options)} aria-label={m.search_services({}, options)} /></label>
		</div>
		<div class="overflow-x-auto" aria-busy={refreshing}>
			<table class="table">
				<thead><tr><th>{m.service({}, options)}</th><th>{m.status({}, options)}</th><th>{m.disk_version({}, options)}</th><th>{m.running_version({}, options)}</th><th>{m.actions({}, options)}</th></tr></thead>
				<tbody>
					{#each filtered as row (row.name)}
						<tr class="hover:bg-base-200/50">
							<td><button class="text-left font-semibold underline-offset-4 hover:underline" type="button" onclick={() => selectedName = row.name}>{displayName(row)}</button><div class="mt-1 text-xs text-base-content/50">{row.name} · {row.disk?.type ?? row.instances[0]?.type}</div></td>
							<td><ServiceStatus status={row.status} />{#if row.versionMismatch}<div class="mt-1 text-xs text-warning">{m.version_changed({}, options)}</div>{/if}</td>
							<td class="font-mono text-xs">{row.disk?.version ?? m.unavailable({}, options)}</td>
							<td class="font-mono text-xs">{row.instances.length ? [...new Set(row.instances.map((item) => item.version))].join(', ') : m.not_set({}, options)}</td>
							<td><div class="flex flex-wrap items-center gap-1"><ServiceActions {row} disabled={actionsDisabled} pending={busyName === row.name ? pending : null} onaction={act} /><button class="btn btn-ghost btn-sm" type="button" onclick={() => selectedName = row.name} aria-label={`${m.details({}, options)} ${row.name}`}>{m.details({}, options)}</button></div></td>
						</tr>
					{:else}<tr><td colspan="5" class="py-16 text-center text-base-content/50">{!loaded ? m.loading_services({}, options) : rows.length ? m.no_matches({}, options) : m.no_services({}, options)}</td></tr>{/each}
				</tbody>
			</table>
		</div>
		<footer class="flex flex-col gap-2 border-t border-base-300 px-4 py-3 text-xs text-base-content/50 sm:flex-row sm:justify-between">
			<p class="break-all">{m.service_directory({}, options)}: {directory || m.unavailable({}, options)}</p>
			<p>{m.auto_refresh({}, options)}{#if lastUpdated} · {m.last_updated({ time: lastUpdated.toLocaleTimeString($locale) }, options)}{/if}</p>
		</footer>
	</section>
</main>

{#if selected}
	{#key selected.name}<ServiceDetails row={selected} disabled={actionsDisabled} pending={busyName === selected.name ? pending : null} onaction={act} onclose={() => selectedName = ''} onupdated={refresh} />{/key}
{/if}

{#if managerSettings}
	<ManagerSettings mode={managerSettings} disabled={managerDisabled} onclose={() => managerSettings = null} onupdated={refresh} onbusy={(busy) => globalBusy = busy} />
{/if}
