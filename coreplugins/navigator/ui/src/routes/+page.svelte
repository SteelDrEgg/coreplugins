<script lang="ts">
	import { onMount } from 'svelte';
	import NavigatorSidebar from '$lib/components/NavigatorSidebar.svelte';
	import ServiceViewport from '$lib/components/ServiceViewport.svelte';
	import SettingsDialog from '$lib/components/SettingsDialog.svelte';
	import * as m from '$lib/paraglide/messages.js';
	import type { Locale } from '$lib/paraglide/runtime.js';
	import { loadEntries, loadNavigatorConfig, loadServerFriendlyName } from '$lib/utils/api';
	import {
		availableLanguages,
		changeLocale,
		initializeLocale,
		locale
	} from '$lib/utils/locale';
	import { getTheme, setTheme, subscribeTheme } from '$lib/utils/preferences';
	import type { EntriesPayload, NavigationEntry, NavigatorConfig } from '$lib/utils/types';

	let entries: NavigationEntry[] = $state([]);
	let activeID = $state('');
	let openedIDs: string[] = $state([]);
	let loadedIDs: string[] = $state([]);
	let selectedLanguage: Locale = $derived($locale);
	let localeOptions = $derived({ locale: $locale });
	let darkTheme = $state(false);
	let loadingEntries = $state(true);
	let message = $state('');
	let mobileOpen = $state(false);
	let navigatorConfig: NavigatorConfig = $state({ icon: '/Arupa.svg', order: [], hide: [] });
	let serverFriendlyName = $state('Arupa');
	let settingsDialog: { show: () => void };

	let openedEntries = $derived(
		openedIDs
			.map((id) => entries.find((entry) => entry.id === id))
			.filter((entry): entry is NavigationEntry => Boolean(entry))
	);
	let activeLoaded = $derived(activeID !== '' && loadedIDs.includes(activeID));

	function selectEntry(id: string) {
		if (!entries.some((entry) => entry.id === id)) return;
		activeID = id;
		if (!openedIDs.includes(id)) openedIDs = [...openedIDs, id];
		mobileOpen = false;
	}

	function markLoaded(id: string) {
		if (!loadedIDs.includes(id)) loadedIDs = [...loadedIDs, id];
	}

	function changeLanguage(candidate: string) {
		const language = availableLanguages.find(({ code }) => code === candidate);
		if (!language || language.code === selectedLanguage) return;
		changeLocale(language.code);
	}

	function changeTheme(enabled: boolean) {
		darkTheme = setTheme(enabled ? 'dark' : 'light') === 'dark';
	}

	function applyEntries(payload: EntriesPayload) {
		entries = payload.entries;
		openedIDs = openedIDs.filter((id) => entries.some((entry) => entry.id === id));
		loadedIDs = loadedIDs.filter((id) => entries.some((entry) => entry.id === id));
		if (!entries.some((entry) => entry.id === activeID)) activeID = entries[0]?.id || '';
		if (activeID && !openedIDs.includes(activeID)) openedIDs = [...openedIDs, activeID];
		if (entries.length === 0) message = m.no_accessible_services({}, localeOptions);
	}

	async function refresh() {
		loadingEntries = true;
		message = '';
		try {
			applyEntries(await loadEntries());
		} catch (error) {
			message =
				error instanceof Error ? error.message : m.failed_load_navigation({}, localeOptions);
		} finally {
			loadingEntries = false;
		}
	}

	async function initialize() {
		await Promise.all([
			refresh(),
			loadNavigatorConfig()
				.then((config) => (navigatorConfig = config))
				.catch((error) => {
					message =
						error instanceof Error
							? error.message
							: m.failed_load_navigator_settings({}, localeOptions);
				}),
			loadServerFriendlyName()
				.then((name) => (serverFriendlyName = name))
				.catch(() => (serverFriendlyName = 'Arupa'))
		]);
	}

	function configSaved(config: NavigatorConfig) {
		navigatorConfig = config;
		void refresh();
	}

	function serverFriendlyNameSaved(name: string) {
		serverFriendlyName = name || 'Arupa';
	}

	onMount(() => {
		const disconnectLocale = initializeLocale();
		darkTheme = setTheme(getTheme()) === 'dark';
		void initialize();
		const unsubscribe = subscribeTheme((theme) => {
			darkTheme = setTheme(theme) === 'dark';
		});
		const resize = () => {
			if (window.innerWidth > 768) mobileOpen = false;
		};
		window.addEventListener('resize', resize);
		return () => {
			disconnectLocale();
			unsubscribe();
			window.removeEventListener('resize', resize);
		};
	});
</script>

<svelte:head>
	<title>{m.page_title({}, localeOptions)}</title>
	<meta name="description" content={m.page_description({}, localeOptions)} />
</svelte:head>

<div
	class="grid h-screen grid-cols-[4.75rem_minmax(0,1fr)] grid-rows-1 overflow-hidden max-md:grid-cols-1 max-md:grid-rows-[3.75rem_minmax(0,1fr)]"
>
	<NavigatorSidebar
		{entries}
		{activeID}
		loading={loadingEntries}
		{mobileOpen}
		brandIcon={navigatorConfig.icon}
		serverName={serverFriendlyName}
		onselect={selectEntry}
		onopen={() => (mobileOpen = true)}
		onclose={() => (mobileOpen = false)}
		onsettings={() => settingsDialog.show()}
	/>

	<ServiceViewport
		loading={loadingEntries}
		{activeLoaded}
		{message}
		{openedEntries}
		{activeID}
		onrefresh={refresh}
		onload={markLoaded}
	/>
</div>

<SettingsDialog
	bind:this={settingsDialog}
	{darkTheme}
	{selectedLanguage}
	languages={availableLanguages}
	config={navigatorConfig}
	{serverFriendlyName}
	onThemeChange={changeTheme}
	onLanguageChange={changeLanguage}
	onConfigSaved={configSaved}
	onServerFriendlyNameSaved={serverFriendlyNameSaved}
/>
