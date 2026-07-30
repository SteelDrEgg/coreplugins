<script lang="ts">
	import CogFourIcon from '@iconify-svelte/mynaui/cog-four';
	import * as m from '$lib/paraglide/messages.js';
	import type { Locale } from '$lib/paraglide/runtime.js';
	import {
		loadKernelVersion,
		loadNavigatorConfig,
		loadServerFriendlyName,
		reloadKernel,
		saveNavigatorConfig,
		saveServerFriendlyName
	} from '$lib/utils/api';
	import { locale, type AvailableLanguage } from '$lib/utils/locale';
	import type { BannerMessage, NavigatorConfig } from '$lib/utils/types';
	import GeneralSettings from './GeneralSettings.svelte';
	import NavigatorSettings from './NavigatorSettings.svelte';

	let {
		darkTheme,
		selectedLanguage,
		languages,
		config,
		serverFriendlyName,
		onThemeChange,
		onLanguageChange,
		onConfigSaved,
		onServerFriendlyNameSaved
	}: {
		darkTheme: boolean;
		selectedLanguage: Locale;
		languages: AvailableLanguage[];
		config: NavigatorConfig;
		serverFriendlyName: string;
		onThemeChange: (enabled: boolean) => void;
		onLanguageChange: (language: string) => void;
		onConfigSaved: (config: NavigatorConfig) => void;
		onServerFriendlyNameSaved: (name: string) => void;
	} = $props();

	let dialog: HTMLDialogElement;
	let activeTab = $state<'general' | 'navigator'>('general');
	let kernelVersion = $state('');
	let busy = $state(false);
	let message = $state<BannerMessage | null>(null);
	let localeOptions = $derived({ locale: $locale });

	export function show() {
		activeTab = 'general';
		message = null;
		dialog.showModal();
		void refreshVersion();
		void refreshNavigatorConfig();
		void refreshServerFriendlyName();
	}

	async function refreshVersion() {
		try {
			kernelVersion = await loadKernelVersion();
		} catch (error) {
			message = {
				kind: 'error',
				text:
					error instanceof Error
						? error.message
						: m.failed_load_kernel_version({}, localeOptions)
			};
		}
	}

	async function refreshNavigatorConfig() {
		try {
			onConfigSaved(await loadNavigatorConfig());
		} catch (error) {
			message = {
				kind: 'error',
				text:
					error instanceof Error
						? error.message
						: m.failed_load_navigator_settings({}, localeOptions)
			};
		}
	}

	async function refreshServerFriendlyName() {
		try {
			onServerFriendlyNameSaved(await loadServerFriendlyName());
		} catch (error) {
			message = {
				kind: 'error',
				text:
					error instanceof Error
						? error.message
						: m.failed_load_server_friendly_name({}, localeOptions)
			};
		}
	}

	async function runReload() {
		busy = true;
		message = null;
		try {
			message = { kind: 'success', text: await reloadKernel() };
			await refreshVersion();
		} catch (error) {
			message = {
				kind: 'error',
				text:
					error instanceof Error
						? error.message
						: m.failed_reload_configuration({}, localeOptions)
			};
		} finally {
			busy = false;
		}
	}

	async function saveConfig(next: NavigatorConfig) {
		busy = true;
		message = null;
		try {
			const saved = await saveNavigatorConfig(next);
			onConfigSaved(saved);
			message = { kind: 'success', text: m.navigator_settings_saved({}, localeOptions) };
		} catch (error) {
			message = {
				kind: 'error',
				text:
					error instanceof Error
						? error.message
						: m.failed_save_navigator_settings({}, localeOptions)
			};
		} finally {
			busy = false;
		}
	}

	async function saveFriendlyName(name: string) {
		busy = true;
		message = null;
		try {
			const saved = await saveServerFriendlyName(name);
			onServerFriendlyNameSaved(saved);
			message = { kind: 'success', text: m.server_friendly_name_saved({}, localeOptions) };
		} catch (error) {
			message = {
				kind: 'error',
				text:
					error instanceof Error
						? error.message
						: m.failed_save_server_friendly_name({}, localeOptions)
			};
		} finally {
			busy = false;
		}
	}

	const messageClass = {
		info: 'alert-info',
		success: 'alert-success',
		warning: 'alert-warning',
		error: 'alert-error'
	};
</script>

<dialog class="modal" bind:this={dialog} aria-labelledby="settings-title">
	<div class="modal-box max-h-[90vh] max-w-2xl overflow-y-auto">
		<div class="mb-4 flex items-center gap-3">
			<CogFourIcon class="size-6" aria-hidden="true" />
			<h2 class="text-lg font-semibold" id="settings-title"
				>{m.settings({}, localeOptions)}</h2
			>
		</div>

		<div
			class="tabs tabs-border mb-5"
			role="tablist"
			aria-label={m.settings_sections({}, localeOptions)}
		>
			<button
				class:tab-active={activeTab === 'general'}
				class="tab"
				type="button"
				role="tab"
				aria-selected={activeTab === 'general'}
				onclick={() => (activeTab = 'general')}
			>
				{m.general({}, localeOptions)}
			</button>
			<button
				class:tab-active={activeTab === 'navigator'}
				class="tab"
				type="button"
				role="tab"
				aria-selected={activeTab === 'navigator'}
				onclick={() => (activeTab = 'navigator')}
			>
				{m.navigator({}, localeOptions)}
			</button>
		</div>

		{#if message}
			<div
				class={`alert ${messageClass[message.kind]} mb-4 text-sm`}
				role="status"
				aria-live="polite"
			>
				<span>{message.text}</span>
			</div>
		{/if}

		{#if activeTab === 'general'}
			<GeneralSettings
				{darkTheme}
				{selectedLanguage}
				{languages}
				{serverFriendlyName}
				{kernelVersion}
				{busy}
				{onThemeChange}
				{onLanguageChange}
				onSaveServerFriendlyName={saveFriendlyName}
				onReload={runReload}
			/>
		{:else}
			<NavigatorSettings {config} {busy} onsave={saveConfig} />
		{/if}

		<div class="modal-action">
			<form method="dialog">
				<button class="btn" type="submit" disabled={busy}>{m.close({}, localeOptions)}</button>
			</form>
		</div>
	</div>
	<form class="modal-backdrop" method="dialog">
		<button aria-label={m.close_settings({}, localeOptions)}>{m.close({}, localeOptions)}</button>
	</form>
</dialog>
