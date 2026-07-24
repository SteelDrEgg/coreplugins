<script lang="ts">
	import { onMount } from 'svelte';
	import MessageBanner from '$lib/MessageBanner.svelte';
	import SecretTable from '$lib/SecretTable.svelte';
	import SecretForm from '$lib/SecretForm.svelte';
	import RevealDialog from '$lib/RevealDialog.svelte';
	import { applyTheme, getTheme, onStoredThemeChange } from '$lib/theme';
	import { listSecrets, addSecret, updateSecret, revealSecret, deleteSecret } from '$lib/api';
	import type { WriteSecretInput } from '$lib/api';
	import type { SecretMeta, BannerMessage, MessageKind } from '$lib/types';

	let keys: SecretMeta[] = $state([]);
	let loaded = $state(false);
	let editing: SecretMeta | null = $state(null);
	let filterQuery = $state('');

	let message: BannerMessage | null = $state(null);
	let pendingMessage: BannerMessage | null = null;
	let busyCount = $state(0);
	let busy = $derived(busyCount > 0);

	let revealOpen = $state(false);
	let revealName = $state('');
	let revealValue = $state('');

	function showMessage(text: string, kind: MessageKind = 'info') {
		message = text ? { text, kind } : null;
	}

	// Mirrors the previous vanilla page's setBusy(): while any request is in
	// flight, show a transient "waiting" banner and restore whatever was
	// there before once the last concurrent request settles.
	async function withBusy<T>(fn: () => Promise<T>): Promise<T> {
		if (busyCount === 0) {
			pendingMessage = message;
			showMessage('Waiting for server…', 'warning');
		}
		busyCount += 1;
		try {
			return await fn();
		} finally {
			busyCount -= 1;
			if (busyCount === 0 && message?.kind === 'warning') {
				message = pendingMessage;
			}
		}
	}

	async function loadKeys(showLoadedMessage = false) {
		try {
			keys = await withBusy(() => listSecrets());
			if (showLoadedMessage) showMessage('List refreshed', 'success');
		} catch (error) {
			showMessage((error as Error).message, 'error');
		} finally {
			loaded = true;
		}
	}

	function startEdit(name: string) {
		const item = keys.find((candidate) => candidate.name === name);
		if (item) editing = item;
	}

	function cancelEdit() {
		editing = null;
	}

	async function handleSave(input: WriteSecretInput, isUpdate: boolean) {
		try {
			await withBusy(() => (isUpdate ? updateSecret(input) : addSecret(input)));
			editing = null;
			await loadKeys();
			showMessage(`${isUpdate ? 'Updated' : 'Added'} ${input.name}`, 'success');
		} catch (error) {
			showMessage((error as Error).message, 'error');
		}
	}

	async function handleReveal(name: string) {
		const item = keys.find((candidate) => candidate.name === name);
		let passphrase = '';
		if (item?.encryption === 'scrypt') {
			const entered = window.prompt(`Passphrase for ${name}`);
			if (entered === null) return;
			passphrase = entered;
		}
		try {
			const value = await withBusy(() => revealSecret(name, passphrase));
			revealName = name;
			revealValue = value;
			revealOpen = true;
			showMessage(`Revealed ${name}`, 'success');
		} catch (error) {
			showMessage((error as Error).message, 'error');
		}
	}

	function closeReveal() {
		revealOpen = false;
		revealValue = '';
	}

	async function handleDelete(name: string) {
		if (!window.confirm(`Delete ${name}? This cannot be undone.`)) return;
		try {
			await withBusy(() => deleteSecret(name));
			if (editing?.name === name) editing = null;
			await loadKeys();
			showMessage(`Deleted ${name}`, 'success');
		} catch (error) {
			showMessage((error as Error).message, 'error');
		}
	}

	onMount(() => {
		applyTheme(getTheme());
		loadKeys();
		return onStoredThemeChange(applyTheme);
	});
</script>

<svelte:head>
	<title>Secrets</title>
</svelte:head>

<main class="mx-auto grid w-full max-w-[1320px] gap-4 px-4 py-6 sm:py-8">
	<header class="flex min-w-0 flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
		<div class="flex min-w-0 items-center gap-3">
			<div
				class="grid size-10 shrink-0 place-items-center rounded-lg bg-primary text-primary-content"
				aria-hidden="true"
			>
				<span class="icon key size-5"></span>
			</div>
			<div class="min-w-0">
				<h1 class="text-xl font-semibold">Secrets</h1>
				<p class="truncate text-sm text-base-content/60">Centralized secrets manager</p>
			</div>
		</div>
		<div class="flex flex-wrap gap-2">
			<button class="btn btn-primary" type="button" disabled={busy} onclick={() => loadKeys(true)}>
				Refresh
			</button>
		</div>
	</header>

	<MessageBanner {message} />

	<section class="grid min-w-0 gap-4 lg:grid-cols-[minmax(0,1fr)_360px]">
		<div class="card min-w-0 rounded-lg border border-base-300 bg-base-100 shadow-sm">
			<div
				class="flex flex-col gap-3 border-b border-base-300 p-4 sm:flex-row sm:items-center sm:justify-between"
			>
				<h2 class="font-semibold">Secrets</h2>
				<input
					class="input w-full sm:w-64"
					type="search"
					placeholder="Filter by name"
					disabled={busy}
					bind:value={filterQuery}
				/>
			</div>
			<div class="min-w-0 overflow-x-auto">
				{#if !loaded}
					<div class="grid min-h-48 place-items-center p-6 text-base-content/60">Loading...</div>
				{:else}
					<SecretTable
						{keys}
						{filterQuery}
						onEdit={startEdit}
						onReveal={handleReveal}
						onDelete={handleDelete}
					/>
				{/if}
			</div>
		</div>

		<div class="grid gap-4">
			<div class="stat rounded-lg border border-base-300 bg-base-100 shadow-sm">
				<div class="stat-title">Managed secrets</div>
				<div class="stat-value text-2xl">{keys.length}</div>
			</div>
			<aside class="card h-fit min-w-0 rounded-lg border border-base-300 bg-base-100 shadow-sm">
				<div class="border-b border-base-300 p-4">
					<h2 class="font-semibold">{editing ? 'Edit secret' : 'Add secret'}</h2>
				</div>
				<SecretForm {editing} {busy} onSave={handleSave} onCancel={cancelEdit} />
			</aside>
		</div>
	</section>
</main>

<RevealDialog
	open={revealOpen}
	title={`Secret: ${revealName}`}
	value={revealValue}
	warning="HTTP is not safe, this secret may be compromised"
	onClose={closeReveal}
	onMessage={showMessage}
/>
