<script lang="ts">
	import { onMount } from 'svelte';
	import X from '@iconify-svelte/mynaui/x';
	import * as m from '$lib/paraglide/messages.js';
	import { locale } from '$lib/utils/locale';
	import { getServiceDirectory, updateServiceDirectory } from '$lib/utils/api';
	import ServiceSettings from './ServiceSettings.svelte';

	let { mode, disabled, onclose, onupdated, onbusy }: {
		mode: 'directory' | 'defaults';
		disabled: boolean;
		onclose: () => void;
		onupdated: () => Promise<boolean>;
		onbusy: (busy: boolean) => void;
	} = $props();
	let dialog: HTMLDialogElement;
	let options = $derived({ locale: $locale });
	let current = $state('');
	let draft = $state('');
	let loading = $state(true);
	let ready = $state(false);
	let saving = $state(false);
	let error = $state('');
	let notice: 'directory_saved' | 'directory_refresh_failed' | 'no_changes' | null = $state(null);
	let alive = true;

	function failure(value: unknown) {
		error = value instanceof Error ? value.message : m.unknown_error({}, options);
	}

	async function loadDirectory() {
		loading = true;
		error = '';
		try {
			const data = await getServiceDirectory();
			if (!alive) return;
			current = data.service_dir;
			draft = current;
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
		if (!path) { error = m.directory_required({}, options); return; }
		if (path === current) { notice = 'no_changes'; return; }
		saving = true;
		onbusy(true);
		try {
			const data = await updateServiceDirectory(path);
			if (!alive) return;
			current = data.service_dir;
			draft = current;
			notice = (await onupdated()) ? 'directory_saved' : 'directory_refresh_failed';
		} catch (value) {
			if (!alive) return;
			failure(value);
			// The Kernel may have persisted the path before its scan failed.
			await onupdated();
			try {
				const data = await getServiceDirectory();
				if (alive) current = data.service_dir;
			} catch { /* Keep the draft available for correction. */ }
		} finally {
			onbusy(false);
			if (alive) saving = false;
		}
	}

	onMount(() => {
		alive = true;
		dialog.showModal();
		if (mode === 'directory') void loadDirectory();
		return () => { alive = false; };
	});
</script>

<dialog bind:this={dialog} class="modal" onclose={onclose} aria-labelledby="manager-settings-title">
	<div class="modal-box max-h-[90dvh] w-[calc(100%-2rem)] space-y-5 p-4 sm:p-6 {mode === 'defaults' ? 'max-w-3xl' : 'max-w-xl'}">
		<header class="flex items-start justify-between gap-4">
			<h2 id="manager-settings-title" class="text-xl font-semibold">
				{mode === 'directory' ? m.directory_settings({}, options) : m.service_defaults({}, options)}
			</h2>
			<form method="dialog"><button class="btn btn-ghost btn-square btn-sm" aria-label={m.close({}, options)}><X class="size-5" aria-hidden="true" /></button></form>
		</header>
		{#if mode === 'defaults'}
			<ServiceSettings name="default" isDefault {disabled} {onupdated} {onbusy} />
		{:else}
			<p class="text-sm text-base-content/60">{m.directory_hint({}, options)}</p>
			{#if error}<div class="alert alert-error alert-soft" role="alert">{error}</div>{/if}
			{#if notice}<div class="alert alert-soft {notice === 'directory_refresh_failed' ? 'alert-warning' : 'alert-success'}" role="status">{m[notice]({}, options)}</div>{/if}
			{#if loading}
				<p class="flex items-center gap-2 py-4 text-sm" role="status"><span class="loading loading-spinner loading-sm" aria-hidden="true"></span>{m.loading_settings({}, options)}</p>
			{:else if !ready}
				<button type="button" class="btn" onclick={loadDirectory}>{m.retry({}, options)}</button>
			{:else}
				<dl class="text-sm"><dt class="text-base-content/50">{m.current_directory({}, options)}</dt><dd class="mt-1 break-all font-mono text-xs">{current}</dd></dl>
				<form onsubmit={saveDirectory} class="space-y-4">
					<label class="block space-y-2 text-sm"><span class="block">{m.service_directory({}, options)}</span><input class="input w-full font-mono text-sm" type="text" required bind:value={draft} disabled={saving} /></label>
					<button class="btn" type="submit" disabled={disabled || saving}>{#if saving}<span class="loading loading-spinner loading-xs" aria-hidden="true"></span>{/if}{m.save_directory({}, options)}</button>
				</form>
			{/if}
		{/if}
	</div>
	<form method="dialog" class="modal-backdrop"><button>{m.close({}, options)}</button></form>
</dialog>
