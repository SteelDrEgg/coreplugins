<script lang="ts">
	import { onMount } from 'svelte';
	import { loadEntries } from '$lib/api';
	import { getLanguage, getTheme, setLanguage, setTheme, subscribePreferences } from '$lib/preferences';
	import type { LanguageDefinition, NavigationEntry } from '$lib/types';

	const fallbackIcon = '/assets/icon/puzzle.svg';

	let entries: NavigationEntry[] = $state([]);
	let activeID = $state('');
	let openedIDs: string[] = $state([]);
	let loadedIDs: string[] = $state([]);
	let languages: string[] = $state(['en']);
	let languageDefinitions: Record<string, LanguageDefinition> = $state({});
	let selectedLanguage = $state('en');
	let darkTheme = $state(false);
	let loadingEntries = $state(true);
	let message = $state('');
	let mobileOpen = $state(false);
	let settingsDialog: HTMLDialogElement;

	let openedEntries = $derived(
		openedIDs
			.map((id) => entries.find((entry) => entry.id === id))
			.filter((entry): entry is NavigationEntry => Boolean(entry))
	);
	let activeLoaded = $derived(activeID !== '' && loadedIDs.includes(activeID));

	function iconURL(entry: NavigationEntry, active: boolean): string {
		return (active ? entry.icon_solid || entry.icon : entry.icon) || fallbackIcon;
	}

	function cssURL(url: string): string {
		return `url("${url.replaceAll('\\', '\\\\').replaceAll('"', '\\"').replaceAll('\n', '')}")`;
	}

	function selectEntry(id: string) {
		if (!entries.some((entry) => entry.id === id)) return;
		activeID = id;
		if (!openedIDs.includes(id)) openedIDs = [...openedIDs, id];
		mobileOpen = false;
	}

	function markLoaded(id: string) {
		if (!loadedIDs.includes(id)) loadedIDs = [...loadedIDs, id];
	}

	function languageDefinition(code: string): LanguageDefinition {
		return languageDefinitions[code] || { name: code };
	}

	function languageLabel(code: string): string {
		const definition = languageDefinition(code);
		const name = definition.name || code;
		if (definition.nativeName && definition.nativeName !== name) {
			return `${name} (${definition.nativeName})`;
		}
		return name;
	}

	function normalizeSelectedLanguage(candidate: string): string {
		const normalized = candidate.trim().toLowerCase();
		if (languages.includes(normalized)) return normalized;
		const browserLanguage = navigator.language.toLowerCase().split('-')[0];
		if (languages.includes(browserLanguage)) return browserLanguage;
		return languages[0] || 'en';
	}

	function applyLanguage(candidate: string) {
		selectedLanguage = normalizeSelectedLanguage(candidate);
		document.documentElement.lang = selectedLanguage;
	}

	function changeLanguage(candidate: string) {
		applyLanguage(setLanguage(normalizeSelectedLanguage(candidate)));
	}

	function changeTheme(enabled: boolean) {
		darkTheme = setTheme(enabled ? 'dark' : 'light') === 'dark';
	}

	function openSettings() {
		if (typeof settingsDialog?.showModal === 'function') settingsDialog.showModal();
	}

	async function fetchLanguageDefinitions() {
		try {
			const response = await fetch('/assets/js/lang.json', { credentials: 'include' });
			if (response.ok) languageDefinitions = (await response.json()) as Record<string, LanguageDefinition>;
		} catch {
			languageDefinitions = {};
		}
	}

	async function refresh() {
		loadingEntries = true;
		message = '';
		try {
			const payload = await loadEntries();
			entries = payload.entries;
			languages = payload.languages.map((language) => language.toLowerCase());
			openedIDs = openedIDs.filter((id) => entries.some((entry) => entry.id === id));
			loadedIDs = loadedIDs.filter((id) => entries.some((entry) => entry.id === id));
			if (!entries.some((entry) => entry.id === activeID)) {
				activeID = entries[0]?.id || '';
			}
			if (activeID && !openedIDs.includes(activeID)) openedIDs = [...openedIDs, activeID];
			applyLanguage(getLanguage());
			if (entries.length === 0) message = 'No accessible services expose an entry route.';
		} catch (error) {
			message = error instanceof Error ? error.message : 'Failed to load navigation entries.';
		} finally {
			loadingEntries = false;
		}
	}

	onMount(() => {
		darkTheme = setTheme(getTheme()) === 'dark';
		void fetchLanguageDefinitions();
		void refresh();
		const unsubscribe = subscribePreferences(
			(theme) => {
				darkTheme = setTheme(theme) === 'dark';
			},
			(language) => applyLanguage(language)
		);
		const resize = () => {
			if (window.innerWidth > 768) mobileOpen = false;
		};
		window.addEventListener('resize', resize);
		return () => {
			unsubscribe();
			window.removeEventListener('resize', resize);
		};
	});
</script>

<svelte:head>
	<title>Arupa</title>
	<meta name="description" content="Navigate Arupa services" />
</svelte:head>

<div
	class="grid h-screen grid-cols-[4.75rem_minmax(0,1fr)] grid-rows-1 overflow-hidden max-md:grid-cols-1 max-md:grid-rows-[3.75rem_minmax(0,1fr)]"
>
	<header class="hidden items-center justify-between border-b border-base-300 bg-base-100 px-3 max-md:flex">
		<div class="flex min-w-0 items-center gap-2 font-semibold">
			<span
				class="mask-icon size-7 shrink-0"
				style:--icon-url={cssURL('/assets/icon/Arupa.svg')}
				aria-hidden="true"
			></span>
			<span class="truncate">Arupa</span>
		</div>
		<button
			class="btn btn-square btn-ghost"
			type="button"
			aria-label="Open navigation"
			onclick={() => (mobileOpen = true)}
		>
			<span
				class="mask-icon"
				style:--icon-url={cssURL('/assets/icon/sidebar.svg')}
				aria-hidden="true"
			></span>
		</button>
	</header>

	<aside
		class:!translate-x-0={mobileOpen}
		class="flex min-h-0 flex-col border-r border-base-300 bg-base-100 max-md:fixed max-md:inset-y-0 max-md:left-0 max-md:z-40 max-md:w-20 max-md:-translate-x-full max-md:transition-transform"
	>
		<div class="grid min-h-16 shrink-0 place-items-center border-b border-base-300">
			<span
				class="mask-icon size-9"
				style:--icon-url={cssURL('/assets/icon/Arupa.svg')}
				aria-hidden="true"
			></span>
		</div>

		<nav
			class="flex min-h-0 flex-1 flex-col items-center gap-2 overflow-y-auto px-2 py-3"
			aria-label="Service navigation"
		>
			{#if loadingEntries}
				<div class="grid min-h-24 place-items-center text-base-content/50">
					<span class="loading loading-spinner loading-sm" aria-label="Loading services"></span>
				</div>
			{:else}
				{#each entries as entry (entry.id)}
					{@const active = entry.id === activeID}
					<button
						class:text-primary={active}
						class="group flex min-h-[4.5rem] w-[3.75rem] flex-col items-center justify-center gap-1 border-0 bg-transparent p-0 font-medium text-base-content transition-colors hover:text-secondary focus-visible:text-secondary focus-visible:outline-none"
						type="button"
						title={entry.label}
						aria-current={active ? 'page' : undefined}
						onclick={() => selectEntry(entry.id)}
					>
						<span
							class:bg-primary={active}
							class:text-primary-content={active}
							class="grid size-9 place-items-center rounded-md transition-colors group-hover:bg-secondary group-hover:text-secondary-content group-focus-visible:bg-secondary group-focus-visible:text-secondary-content"
							aria-hidden="true"
						>
							<span
								class="mask-icon"
								style:--icon-url={cssURL(iconURL(entry, active))}
							></span>
						</span>
						<span class="w-full overflow-hidden text-ellipsis whitespace-nowrap text-center text-xs">
							{entry.label}
						</span>
					</button>
				{/each}
			{/if}
		</nav>

		<div class="grid shrink-0 place-items-center border-t border-base-300 p-2">
			<button
				class="btn btn-square btn-ghost"
				type="button"
				title="Settings"
				aria-label="Settings"
				onclick={openSettings}
			>
				<span
					class="mask-icon"
					style:--icon-url={cssURL('/assets/icon/cog-four.svg')}
					aria-hidden="true"
				></span>
			</button>
		</div>
	</aside>

	{#if mobileOpen}
		<button
			class="fixed inset-0 z-30 bg-neutral/40 md:hidden"
			type="button"
			aria-label="Close navigation"
			onclick={() => (mobileOpen = false)}
		></button>
	{/if}

	<main class="relative min-h-0 min-w-0 overflow-hidden bg-base-100">
		{#if loadingEntries || (activeID && !activeLoaded)}
			<div class="absolute inset-0 z-10 grid place-items-center bg-base-100 text-base-content/60">
				<span class="loading loading-spinner loading-md" aria-label="Loading page"></span>
			</div>
		{/if}

		{#if message}
			<div
				class="absolute inset-x-4 top-4 z-20 flex items-center justify-between gap-3 rounded-lg border border-error/20 bg-base-100 p-4 text-sm text-error shadow-sm"
				role="status"
			>
				<span>{message}</span>
				<button class="btn btn-sm" type="button" onclick={refresh}>Retry</button>
			</div>
		{/if}

		{#each openedEntries as entry (entry.id)}
			<iframe
				class:hidden={entry.id !== activeID}
				class="absolute inset-0 size-full border-0 bg-base-100"
				title={entry.label}
				src={entry.href}
				onload={() => markLoaded(entry.id)}
			></iframe>
		{/each}
	</main>
</div>

<dialog class="modal" bind:this={settingsDialog} aria-labelledby="settings-title">
	<div class="modal-box max-w-md">
		<div class="mb-5 flex items-center gap-3">
			<span
				class="mask-icon size-6"
				style:--icon-url={cssURL('/assets/icon/cog-four.svg')}
				aria-hidden="true"
			></span>
			<h2 class="text-lg font-semibold" id="settings-title">Settings</h2>
		</div>

		<div class="space-y-4">
			<label class="flex items-center justify-between gap-4 rounded-lg border border-base-300 p-4">
				<span>
					<span class="block font-medium">Theme</span>
					<span class="text-xs text-base-content/60">{darkTheme ? 'Dark' : 'Light'}</span>
				</span>
				<input
					class="toggle toggle-primary"
					type="checkbox"
					checked={darkTheme}
					aria-label="Toggle dark theme"
					onchange={(event) => changeTheme(event.currentTarget.checked)}
				/>
			</label>

			<label class="grid gap-2 rounded-lg border border-base-300 p-4">
				<span class="font-medium">Language</span>
				<select
					class="select w-full"
					value={selectedLanguage}
					aria-label="Language"
					onchange={(event) => changeLanguage(event.currentTarget.value)}
				>
					{#each languages as language}
						<option value={language}>{languageLabel(language)}</option>
					{/each}
				</select>
			</label>
		</div>

		<div class="modal-action">
			<a class="btn btn-error btn-outline" href="/pages/logout.html">Logout</a>
			<form method="dialog">
				<button class="btn" type="submit">Close</button>
			</form>
		</div>
	</div>
	<form class="modal-backdrop" method="dialog">
		<button aria-label="Close settings">close</button>
	</form>
</dialog>
