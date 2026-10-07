<script lang="ts">
	import { onMount } from 'svelte';
	import * as m from '$lib/paraglide/messages.js';
	import { locale } from '$lib/utils/locale';
	import { getServiceTempDirectory, updateServiceTempDirectory } from '$lib/utils/api';
	import type { ServiceTempDirectory } from '$lib/utils/types';

	let { disabled, onbusy }: {
		disabled: boolean;
		onbusy: (busy: boolean) => void;
	} = $props();
	let options = $derived({ locale: $locale });
	let current = $state('');
	let draft = $state('');
	let requiresRestart = $state(false);
	let loading = $state(true);
	let ready = $state(false);
	let saving = $state(false);
	let error = $state('');
	let notice: 'temporary_directory_saved' | 'no_changes' | null = $state(null);
	let alive = true;

	function applyDirectory(data: ServiceTempDirectory, resetDraft: boolean) {
		current = data.temp_dir;
		requiresRestart = data.requires_restart === true;
		if (resetDraft) draft = current;
	}

	function failure(value: unknown) {
		error = value instanceof Error ? value.message : m.unknown_error({}, options);
	}

	async function loadDirectory() {
		loading = true;
		error = '';
		try {
			const data = await getServiceTempDirectory();
			if (!alive) return;
			applyDirectory(data, true);
			ready = true;
		} catch (value) {
			if (alive) failure(value);
		} finally {
			if (alive) loading = false;
		}
	}

	async function saveDirectory(event: SubmitEvent) {
		event.preventDefault();
		if (disabled || saving || !ready) return;
		error = '';
		notice = null;
		const path = draft.trim();
		if (!path) { error = m.temporary_directory_required({}, options); return; }
		if (path === current) { notice = 'no_changes'; return; }
		saving = true;
		onbusy(true);
		try {
			const data = await updateServiceTempDirectory(path);
			if (!alive) return;
			applyDirectory(data, true);
			notice = 'temporary_directory_saved';
		} catch (value) {
			if (!alive) return;
			failure(value);
			// Reconcile persisted state after an interrupted response, retaining the draft.
			try {
				const data = await getServiceTempDirectory();
				if (alive) applyDirectory(data, false);
			} catch { /* Keep the last known directory and the draft for correction. */ }
		} finally {
			onbusy(false);
			if (alive) saving = false;
		}
	}

	onMount(() => {
		alive = true;
		void loadDirectory();
		return () => { alive = false; };
	});
</script>

<section class="space-y-4 border-t border-base-300 pt-5" aria-labelledby="temporary-directory-title">
	<h3 id="temporary-directory-title" class="font-semibold">{m.temporary_directory({}, options)}</h3>
	<p class="text-sm text-base-content/60">{m.temporary_directory_hint({}, options)}</p>
	{#if error}<div class="alert alert-error alert-soft" role="alert">{error}</div>{/if}
	{#if notice}<div class="alert alert-success alert-soft" role="status">{m[notice]({}, options)}</div>{/if}
	{#if loading}
		<p class="flex items-center gap-2 py-4 text-sm" role="status"><span class="loading loading-spinner loading-sm" aria-hidden="true"></span>{m.loading_settings({}, options)}</p>
	{:else if !ready}
		<button type="button" class="btn" onclick={loadDirectory} {disabled}>{m.retry({}, options)}</button>
	{:else}
		<dl class="text-sm"><dt class="text-base-content/50">{m.current_temporary_directory({}, options)}</dt><dd class="mt-1 break-all font-mono text-xs">{current}</dd></dl>
		{#if requiresRestart}<div class="alert alert-warning alert-soft" role="status">{m.temporary_directory_restart_required({}, options)}</div>{/if}
		<form onsubmit={saveDirectory} class="space-y-4">
			<label class="block space-y-2 text-sm"><span class="block">{m.temporary_directory({}, options)}</span><input class="input w-full font-mono text-sm" type="text" required bind:value={draft} disabled={disabled || saving} /></label>
			<button class="btn" type="submit" disabled={disabled || saving}>{#if saving}<span class="loading loading-spinner loading-xs" aria-hidden="true"></span>{/if}{m.save_temporary_directory({}, options)}</button>
		</form>
	{/if}
</section>
