<script lang="ts">
	import { onMount } from 'svelte';
	import * as m from '$lib/paraglide/messages.js';
	import { locale } from '$lib/utils/locale';
	import { getServiceConfig, updateServiceConfig, getServiceParams, updateServiceParams } from '$lib/utils/api';
	import { configPatch, paramsPatch } from '$lib/utils/services';
	import type { ServiceConfig } from '$lib/utils/types';
	let { name, effective, disabled, onupdated, isDefault = false, onbusy = () => {} }: { name: string; effective?: ServiceConfig; disabled: boolean; onupdated: () => Promise<boolean>; isDefault?: boolean; onbusy?: (busy: boolean) => void } = $props();
	let options = $derived({ locale: $locale });
	let loading = $state(true);
	let ready = $state(false);
	let saving = $state(false);
	let error = $state('');
	let notice: 'settings_saved' | 'settings_refresh_failed' | 'params_saved' | 'default_params_saved' | 'default_settings_saved' | 'no_changes' | 'params_invalid' | 'groups_required' | null = $state(null);
	let before: ServiceConfig = {};
	let mode = $state('');
	let retries = $state<number | undefined>(3);
	let user = $state('');
	let checksum = $state('');
	let groups = $state('');
	let groupMode: 'inherit' | 'open' | 'restricted' = $state('inherit');
	let paramsLoading = $state(false);
	let paramsLoaded = $state(false);
	let showValues = $state(false);
	let originalParams: Record<string, string> = {};
	let parameters: { id: number; key: string; value: string }[] = $state([]);
	let nextID = 0;
	let alive = true;

	function restartValue() { return mode === 'limited' ? `on-failure:${retries ?? 0}` : mode; }
	function draft(): ServiceConfig { return { restart: restartValue(), run_as_user: user, checksum, allow: groupMode === 'inherit' ? null : groupMode === 'open' ? [] : [...new Set(groups.split(',').map((value) => value.trim()).filter(Boolean))] }; }
	function populate(config: ServiceConfig) {
		const restart = (config.restart ?? '').toLowerCase().trim();
		mode = restart.startsWith('on-failure:') ? 'limited' : ['yes', 'true', 'on', 'enable', 'enabled', '1'].includes(restart) ? 'always' : ['false', 'off', 'disable', 'disabled', '0'].includes(restart) ? 'no' : restart;
		retries = mode === 'limited' ? Number(restart.split(':')[1]) : 3;
		user = config.run_as_user ?? '';
		checksum = config.checksum ?? '';
		groups = (config.allow ?? []).join(', ');
		groupMode = config.allow == null ? 'inherit' : config.allow.length ? 'restricted' : 'open';
		before = draft();
	}
	function failure(value: unknown) { error = value instanceof Error ? value.message : m.unknown_error({}, options); }
	async function loadConfig() {
		loading = true; error = ''; notice = null;
		try { const config = await getServiceConfig(name); if (alive) { populate(config); ready = true; } }
		catch (value) { if (alive) failure(value); }
		finally { if (alive) loading = false; }
	}
	async function saveConfig(event: SubmitEvent) {
		event.preventDefault();
		if (disabled || saving || paramsLoading || !ready) return;
		error = ''; notice = null;
		if (groupMode === 'restricted' && !draft().allow?.length) { notice = 'groups_required'; return; }
		const patch = configPatch(before, draft());
		if (!Object.keys(patch).length) { notice = 'no_changes'; return; }
		saving = true; error = ''; notice = null;
		onbusy(true);
		try {
			const saved = await updateServiceConfig(name, patch);
			if (!alive) return;
			populate(saved);
			notice = (await onupdated()) ? (isDefault ? 'default_settings_saved' : 'settings_saved') : 'settings_refresh_failed';
		} catch (value) { if (alive) failure(value); }
		finally { onbusy(false); if (alive) saving = false; }
	}
	async function loadParams() {
		if (paramsLoading || saving || disabled) return;
		paramsLoading = true; error = ''; notice = null;
		try {
			const response = await getServiceParams(name);
			if (!alive) return;
			originalParams = response.params ?? {};
			parameters = Object.entries(originalParams).map(([key, value]) => ({ id: nextID++, key, value }));
			paramsLoaded = true;
		} catch (value) { if (alive) failure(value); }
		finally { if (alive) paramsLoading = false; }
	}
	async function saveParams(event: SubmitEvent) {
		event.preventDefault();
		if (disabled || saving || !paramsLoaded) return;
		error = ''; notice = null;
		const entries = parameters.map((item) => [item.key.trim(), item.value] as const);
		if (entries.some(([key]) => !key) || new Set(entries.map(([key]) => key)).size !== entries.length) { notice = 'params_invalid'; return; }
		const patch = paramsPatch(originalParams, Object.fromEntries(entries));
		if (!Object.keys(patch).length) { notice = 'no_changes'; return; }
		saving = true; error = ''; notice = null;
		onbusy(true);
		try {
			const response = await updateServiceParams(name, patch);
			if (!alive) return;
			originalParams = response.params ?? {};
			parameters = Object.entries(originalParams).map(([key, value]) => ({ id: nextID++, key, value }));
			notice = isDefault ? 'default_params_saved' : 'params_saved';
		} catch (value) { if (alive) failure(value); }
		finally { onbusy(false); if (alive) saving = false; }
	}
	onMount(() => { alive = true; void loadConfig(); return () => { alive = false; }; });
</script>

<div class="space-y-6">
	{#if error}<div class="alert alert-error alert-soft" role="alert">{error}</div>{/if}
	{#if notice}<div class="alert alert-soft {['params_invalid', 'groups_required'].includes(notice) ? 'alert-error' : notice === 'settings_refresh_failed' ? 'alert-warning' : 'alert-success'}" role="status">{m[notice]({}, options)}</div>{/if}
	{#if effective}
		<div class="rounded-box bg-base-200 p-4 text-sm"><h3 class="mb-3 font-semibold">{m.effective_config({}, options)}</h3><dl class="grid gap-3 sm:grid-cols-2">
			<div><dt class="text-base-content/50">{m.restart_policy({}, options)}</dt><dd class="mt-1 font-mono">{effective.restart || 'no'}</dd></div>
			<div><dt class="text-base-content/50">{m.allow_groups({}, options)}</dt><dd class="mt-1">{effective.allow?.join(', ') || m.no_group_restriction({}, options)}</dd></div>
			<div><dt class="text-base-content/50">{m.run_as_user({}, options)}</dt><dd class="mt-1">{effective.run_as_user || m.not_set({}, options)}</dd></div>
			<div><dt class="text-base-content/50">{m.checksum({}, options)}</dt><dd class="mt-1 break-all font-mono text-xs">{effective.checksum || m.not_set({}, options)}</dd></div>
		</dl></div>
	{/if}
	{#if loading}<p class="flex items-center gap-2 py-4 text-sm" role="status"><span class="loading loading-spinner loading-sm" aria-hidden="true"></span>{m.loading_settings({}, options)}</p>
	{:else if !ready}<button type="button" class="btn" onclick={loadConfig}>{m.retry({}, options)}</button>
	{:else}
		<form onsubmit={saveConfig} class="space-y-4">
			<h3 class="font-semibold">{isDefault ? m.default_config({}, options) : m.overrides({}, options)}</h3><p class="text-xs text-base-content/60">{isDefault ? m.defaults_hint({}, options) : m.inherited_hint({}, options)}</p>
			<fieldset disabled={saving} class="grid gap-4 sm:grid-cols-2">
				<label class="space-y-2 text-sm"><span class="block">{m.restart_policy({}, options)}</span><select class="select w-full" bind:value={mode}><option value="">{isDefault ? m.builtin_restart({}, options) : m.inherit({}, options)}</option><option value="no">{m.restart_no({}, options)}</option><option value="always">{m.restart_always({}, options)}</option><option value="on-failure">{m.restart_failure({}, options)}</option><option value="limited">{m.restart_limited({}, options)}</option></select></label>
				{#if mode === 'limited'}<label class="space-y-2 text-sm"><span class="block">{m.retry_limit({}, options)}</span><input class="input w-full" type="number" min="0" max="2147483647" step="1" required aria-label={m.retry_limit({}, options)} aria-describedby="service-retry-hint" bind:value={retries} /><span id="service-retry-hint" class="block text-xs text-base-content/50">{m.retry_hint({}, options)}</span></label>{/if}
				<label class="space-y-2 text-sm"><span class="block">{m.run_as_user({}, options)}</span><input class="input w-full" type="text" bind:value={user} placeholder={isDefault ? m.current_process_user({}, options) : m.inherit({}, options)} /></label>
				<label class="space-y-2 text-sm sm:col-span-2"><span class="block">{m.checksum({}, options)}</span><input class="input w-full font-mono text-xs" type="text" aria-label={m.checksum({}, options)} bind:value={checksum} pattern={'[sS][hH][aA]256:[0-9a-fA-F]{64}'} placeholder={isDefault ? m.checksum_disabled({}, options) : m.inherit({}, options)} />{#if isDefault}<span class="block text-xs text-base-content/50">{m.default_checksum_hint({}, options)}</span>{/if}</label>
				<label class="space-y-2 text-sm"><span class="block">{m.group_policy({}, options)}</span><select class="select w-full" bind:value={groupMode}><option value="inherit">{isDefault ? m.builtin_groups({}, options) : m.inherit({}, options)}</option><option value="open">{m.no_group_restriction({}, options)}</option><option value="restricted">{m.restrict_groups({}, options)}</option></select></label>
				{#if groupMode === 'restricted'}<label class="space-y-2 text-sm"><span class="block">{m.allow_groups({}, options)}</span><input class="input w-full" type="text" required aria-label={m.allow_groups({}, options)} bind:value={groups} /><span class="block text-xs text-base-content/50">{m.allow_hint({}, options)}</span></label>{/if}
			</fieldset>
			<p class="text-xs text-base-content/60">{m.config_hint({}, options)}</p>
			<button class="btn" type="submit" disabled={disabled || saving || paramsLoading}>{#if saving}<span class="loading loading-spinner loading-xs" aria-hidden="true"></span>{/if}{m.save_settings({}, options)}</button>
		</form>
		<section class="space-y-4 border-t border-base-300 pt-5">
			<h3 class="font-semibold">{isDefault ? m.default_params({}, options) : m.params({}, options)}</h3>{#if isDefault}<p class="text-xs text-base-content/60">{m.default_params_hint({}, options)}</p>{/if}<p class="text-xs text-base-content/60">{m.params_hint({}, options)}</p>
			{#if !paramsLoaded}<button type="button" class="btn btn-sm" disabled={disabled || paramsLoading || saving} onclick={loadParams}>{#if paramsLoading}<span class="loading loading-spinner loading-xs" aria-hidden="true"></span>{/if}{m.load_params({}, options)}</button>
			{:else}
				<form onsubmit={saveParams} class="space-y-3">
					<div class="flex flex-wrap gap-2"><button type="button" class="btn btn-ghost btn-sm" onclick={() => showValues = !showValues}>{showValues ? m.hide_values({}, options) : m.show_values({}, options)}</button><button type="button" class="btn btn-sm" disabled={saving} onclick={() => parameters = [...parameters, { id: nextID++, key: '', value: '' }]}>{m.add_parameter({}, options)}</button></div>
					{#each parameters as parameter (parameter.id)}
						<div class="flex flex-col gap-2 rounded-box border border-base-300 p-3 sm:flex-row sm:items-end">
							<label class="min-w-0 flex-1 space-y-1 text-xs"><span class="block">{m.parameter_key({}, options)}</span><input class="input input-sm w-full font-mono" type="text" required bind:value={parameter.key} disabled={saving} /></label>
							<label class="min-w-0 flex-1 space-y-1 text-xs"><span class="block">{m.parameter_value({}, options)}</span><input class="input input-sm w-full font-mono" type={showValues ? 'text' : 'password'} bind:value={parameter.value} disabled={saving} autocomplete="off" /></label>
							<button type="button" class="btn btn-ghost btn-sm" disabled={saving} onclick={() => parameters = parameters.filter((item) => item.id !== parameter.id)} aria-label={`${m.remove_parameter({}, options)} ${parameter.key}`}>{m.remove_parameter({}, options)}</button>
						</div>
					{:else}<p class="text-sm text-base-content/50">{m.no_params({}, options)}</p>{/each}
					<button type="submit" class="btn" disabled={disabled || saving}>{m.save_params({}, options)}</button>
				</form>
			{/if}
		</section>
	{/if}
</div>
