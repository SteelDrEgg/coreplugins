<script lang="ts">
	import CogFourIcon from '@iconify-svelte/mynaui/cog-four';
	import {
		loadKernelVersion,
		loadNavigatorConfig,
		reloadKernel,
		saveNavigatorConfig
	} from '$lib/utils/api';
	import type {
		BannerMessage,
		LanguageDefinition,
		NavigatorConfig
	} from '$lib/utils/types';
	import GeneralSettings from './GeneralSettings.svelte';
	import NavigatorSettings from './NavigatorSettings.svelte';

	let {
		darkTheme,
		selectedLanguage,
		languages,
		languageDefinitions,
		config,
		onThemeChange,
		onLanguageChange,
		onConfigSaved
	}: {
		darkTheme: boolean;
		selectedLanguage: string;
		languages: string[];
		languageDefinitions: Record<string, LanguageDefinition>;
		config: NavigatorConfig;
		onThemeChange: (enabled: boolean) => void;
		onLanguageChange: (language: string) => void;
		onConfigSaved: (config: NavigatorConfig) => void;
	} = $props();

	let dialog: HTMLDialogElement;
	let activeTab = $state<'general' | 'navigator'>('general');
	let kernelVersion = $state('');
	let busy = $state(false);
	let message = $state<BannerMessage | null>(null);

	export function show() {
		activeTab = 'general';
		message = null;
		dialog.showModal();
		void refreshVersion();
		void refreshNavigatorConfig();
	}

	async function refreshVersion() {
		try {
			kernelVersion = await loadKernelVersion();
		} catch (error) {
			message = {
				kind: 'error',
				text: error instanceof Error ? error.message : 'Failed to load Kernel version'
			};
		}
	}

	async function refreshNavigatorConfig() {
		try {
			onConfigSaved(await loadNavigatorConfig());
		} catch (error) {
			message = {
				kind: 'error',
				text: error instanceof Error ? error.message : 'Failed to load Navigator settings'
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
				text: error instanceof Error ? error.message : 'Failed to reload configuration'
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
			message = { kind: 'success', text: 'Navigator settings saved' };
		} catch (error) {
			message = {
				kind: 'error',
				text: error instanceof Error ? error.message : 'Failed to save Navigator settings'
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
			<h2 class="text-lg font-semibold" id="settings-title">Settings</h2>
		</div>

		<div class="tabs tabs-border mb-5" role="tablist" aria-label="Settings sections">
			<button
				class:tab-active={activeTab === 'general'}
				class="tab"
				type="button"
				role="tab"
				aria-selected={activeTab === 'general'}
				onclick={() => (activeTab = 'general')}
			>
				General
			</button>
			<button
				class:tab-active={activeTab === 'navigator'}
				class="tab"
				type="button"
				role="tab"
				aria-selected={activeTab === 'navigator'}
				onclick={() => (activeTab = 'navigator')}
			>
				Navigator
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
				{languageDefinitions}
				{kernelVersion}
				{busy}
				{onThemeChange}
				{onLanguageChange}
				onReload={runReload}
			/>
		{:else}
			<NavigatorSettings {config} {busy} onsave={saveConfig} />
		{/if}

		<div class="modal-action">
			<form method="dialog">
				<button class="btn" type="submit" disabled={busy}>Close</button>
			</form>
		</div>
	</div>
	<form class="modal-backdrop" method="dialog">
		<button aria-label="Close settings">close</button>
	</form>
</dialog>
