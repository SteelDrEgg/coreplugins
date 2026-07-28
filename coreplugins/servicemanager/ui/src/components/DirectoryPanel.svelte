<script lang="ts">
	import Folder from '@iconify-svelte/mynaui/folder';

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
		Directories
	</div>
	<form class="card-body gap-4 p-4" onsubmit={submit}>
		<label class="grid gap-2">
			<span class="text-sm font-semibold text-base-content/70">Service directory</span>
			<input
				class="input w-full font-mono text-xs"
				bind:value={serviceDirInput}
				autocomplete="off"
				disabled={busy}
			/>
		</label>
		<label class="grid gap-2">
			<span class="text-sm font-semibold text-base-content/70">Temporary directory</span>
			<input
				class="input w-full font-mono text-xs"
				bind:value={tempDirInput}
				autocomplete="off"
				disabled={busy}
			/>
			{#if requiresRestart}
				<span class="text-xs text-warning">Kernel restart required</span>
			{/if}
		</label>
		<button class="btn btn-primary" type="submit" disabled={busy}>Save directories</button>
	</form>
</div>
