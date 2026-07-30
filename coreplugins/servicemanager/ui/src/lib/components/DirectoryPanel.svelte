<script lang="ts">
	import Folder from '@iconify-svelte/mynaui/folder';
	import * as m from '$lib/paraglide/messages.js';
	import { locale } from '$lib/utils/locale';

	let {
		serviceDir,
		tempDir,
		requiresRestart,
		busy,
		onSave
	}: {
		serviceDir: string;
		tempDir: string;
		requiresRestart: boolean;
		busy: boolean;
		onSave: (serviceDir: string, tempDir: string) => void;
	} = $props();

	let serviceDirInput = $state('');
	let tempDirInput = $state('');
	let localeOptions = $derived({ locale: $locale });

	$effect(() => {
		serviceDirInput = serviceDir;
	});

	$effect(() => {
		tempDirInput = tempDir;
	});

	function submit(event: SubmitEvent) {
		event.preventDefault();
		onSave(serviceDirInput.trim(), tempDirInput.trim());
	}
</script>

<div class="card min-w-0 rounded-lg border border-base-300 bg-base-100 shadow-sm">
	<div class="flex items-center gap-2 border-b border-base-300 p-4 font-semibold">
		<Folder class="size-4" aria-hidden="true" />
		{m.directories({}, localeOptions)}
	</div>
	<form class="card-body gap-4 p-4" onsubmit={submit}>
		<label class="grid gap-2">
			<span class="text-sm font-semibold text-base-content/70">
				{m.service_directory({}, localeOptions)}
			</span>
			<input
				class="input w-full font-mono text-xs"
				bind:value={serviceDirInput}
				autocomplete="off"
				disabled={busy}
			/>
		</label>
		<label class="grid gap-2">
			<span class="text-sm font-semibold text-base-content/70">
				{m.temporary_directory({}, localeOptions)}
			</span>
			<input
				class="input w-full font-mono text-xs"
				bind:value={tempDirInput}
				autocomplete="off"
				disabled={busy}
			/>
			{#if requiresRestart}
				<span class="text-xs text-warning">{m.kernel_restart_required({}, localeOptions)}</span>
			{/if}
		</label>
		<button class="btn btn-primary" type="submit" disabled={busy}>
			{m.save_directories({}, localeOptions)}
		</button>
	</form>
</div>
