<script lang="ts">
	import { onMount } from 'svelte';
	import MessageBanner from '$lib/components/MessageBanner.svelte';
	import SecretTable from '$lib/components/SecretTable.svelte';
	import SecretForm from '$lib/components/SecretForm.svelte';
	import RevealDialog from '$lib/components/RevealDialog.svelte';
	import * as m from '$lib/paraglide/messages.js';
	import { initializeLocale, locale } from '$lib/utils/locale';
	import { applyTheme, getTheme, onStoredThemeChange } from '$lib/utils/theme';
	import { listSecrets, addSecret, updateSecret, revealSecret, deleteSecret } from '$lib/utils/api';
	import type { WriteSecretInput } from '$lib/utils/api';
	import type { SecretMeta, BannerMessage, MessageKind } from '$lib/utils/types';

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
	let localeOptions = $derived({ locale: $locale });

	function showMessage(text: string, kind: MessageKind = 'info') {
		message = text ? { text, kind } : null;
	}

	// Mirrors the previous vanilla page's setBusy(): while any request is in
	// flight, show a transient "waiting" banner and restore whatever was
	// there before once the last concurrent request settles.
	async function withBusy<T>(fn: () => Promise<T>): Promise<T> {
		if (busyCount === 0) {
			pendingMessage = message;
			showMessage(m.waiting_for_server({}, localeOptions), 'warning');
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
			if (showLoadedMessage) showMessage(m.list_refreshed({}, localeOptions), 'success');
		} catch (error) {
			showMessage(
				error instanceof Error ? error.message : m.unknown_error({}, localeOptions),
				'error'
			);
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
			showMessage(
				isUpdate
					? m.secret_updated({ name: input.name }, localeOptions)
					: m.secret_added({ name: input.name }, localeOptions),
				'success'
			);
		} catch (error) {
			showMessage(
				error instanceof Error ? error.message : m.unknown_error({}, localeOptions),
				'error'
			);
		}
	}

	async function handleReveal(name: string) {
		const item = keys.find((candidate) => candidate.name === name);
		let passphrase = '';
		if (item?.encryption === 'scrypt') {
			const entered = window.prompt(m.passphrase_for({ name }, localeOptions));
			if (entered === null) return;
			passphrase = entered;
		}
		try {
			const value = await withBusy(() => revealSecret(name, passphrase));
			revealName = name;
			revealValue = value;
			revealOpen = true;
			showMessage(m.secret_revealed({ name }, localeOptions), 'success');
		} catch (error) {
			showMessage(
				error instanceof Error ? error.message : m.unknown_error({}, localeOptions),
				'error'
			);
		}
	}

	function closeReveal() {
		revealOpen = false;
		revealValue = '';
	}

	async function handleDelete(name: string) {
		if (!window.confirm(m.confirm_delete({ name }, localeOptions))) return;
		try {
			await withBusy(() => deleteSecret(name));
			if (editing?.name === name) editing = null;
			await loadKeys();
			showMessage(m.secret_deleted({ name }, localeOptions), 'success');
		} catch (error) {
			showMessage(
				error instanceof Error ? error.message : m.unknown_error({}, localeOptions),
				'error'
			);
		}
	}

	onMount(() => {
		const unsubscribeLocale = initializeLocale();
		applyTheme(getTheme());
		void loadKeys();
		const unsubscribeTheme = onStoredThemeChange(applyTheme);
		return () => {
			unsubscribeLocale();
			unsubscribeTheme();
		};
	});
</script>

<svelte:head>
	<title>{m.page_title({}, localeOptions)}</title>
	<meta name="description" content={m.page_description({}, localeOptions)} />
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
				<h1 class="text-xl font-semibold">{m.secrets({}, localeOptions)}</h1>
				<p class="truncate text-sm text-base-content/60">
					{m.page_description({}, localeOptions)}
				</p>
			</div>
		</div>
		<div class="flex flex-wrap gap-2">
			<button class="btn btn-primary" type="button" disabled={busy} onclick={() => loadKeys(true)}>
				{m.refresh({}, localeOptions)}
			</button>
		</div>
	</header>

	<MessageBanner {message} />

	<section class="grid min-w-0 gap-4 lg:grid-cols-[minmax(0,1fr)_360px]">
		<div class="card min-w-0 rounded-lg border border-base-300 bg-base-100 shadow-sm">
			<div
				class="flex flex-col gap-3 border-b border-base-300 p-4 sm:flex-row sm:items-center sm:justify-between"
			>
				<h2 class="font-semibold">{m.secrets({}, localeOptions)}</h2>
				<input
					class="input w-full sm:w-64"
					type="search"
					placeholder={m.filter_by_name({}, localeOptions)}
					disabled={busy}
					bind:value={filterQuery}
				/>
			</div>
			<div class="min-w-0 overflow-x-auto">
				{#if !loaded}
					<div class="grid min-h-48 place-items-center p-6 text-base-content/60">
						{m.loading({}, localeOptions)}
					</div>
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
				<div class="stat-title">{m.managed_secrets({}, localeOptions)}</div>
				<div class="stat-value text-2xl">{keys.length}</div>
			</div>
			<aside class="card h-fit min-w-0 rounded-lg border border-base-300 bg-base-100 shadow-sm">
				<div class="border-b border-base-300 p-4">
					<h2 class="font-semibold">
						{editing
							? m.edit_secret({}, localeOptions)
							: m.add_secret({}, localeOptions)}
					</h2>
				</div>
				<SecretForm {editing} {busy} onSave={handleSave} onCancel={cancelEdit} />
			</aside>
		</div>
	</section>
</main>

<RevealDialog
	open={revealOpen}
	title={m.revealed_secret_title({ name: revealName }, localeOptions)}
	value={revealValue}
	warning={m.http_warning({}, localeOptions)}
	onClose={closeReveal}
	onMessage={showMessage}
/>
