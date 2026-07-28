<script lang="ts">
	import DangerTriangleIcon from '@iconify-svelte/mynaui/danger-triangle';
	import RefreshIcon from '@iconify-svelte/mynaui/refresh';
	import type { LanguageDefinition } from '$lib/utils/types';

	let {
		darkTheme,
		selectedLanguage,
		languages,
		languageDefinitions,
		kernelVersion,
		busy,
		onThemeChange,
		onLanguageChange,
		onReload
	}: {
		darkTheme: boolean;
		selectedLanguage: string;
		languages: string[];
		languageDefinitions: Record<string, LanguageDefinition>;
		kernelVersion: string;
		busy: boolean;
		onThemeChange: (enabled: boolean) => void;
		onLanguageChange: (language: string) => void;
		onReload: () => void;
	} = $props();

	let confirmReload = $state(false);

	function languageLabel(code: string): string {
		const definition = languageDefinitions[code] || { name: code };
		const name = definition.name || code;
		return definition.nativeName && definition.nativeName !== name
			? `${name} (${definition.nativeName})`
			: name;
	}
</script>

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
			onchange={(event) => onThemeChange(event.currentTarget.checked)}
		/>
	</label>

	<label class="grid gap-2 rounded-lg border border-base-300 p-4">
		<span class="font-medium">Language</span>
		<select
			class="select w-full"
			value={selectedLanguage}
			aria-label="Language"
			onchange={(event) => onLanguageChange(event.currentTarget.value)}
		>
			{#each languages as language}
				<option value={language}>{languageLabel(language)}</option>
			{/each}
		</select>
	</label>

	<section class="rounded-lg border border-base-300 p-4" aria-labelledby="kernel-heading">
		<div class="flex items-center justify-between gap-3">
			<div>
				<h3 class="font-medium" id="kernel-heading">Kernel</h3>
				<p class="text-xs text-base-content/60">Version {kernelVersion || 'loading…'}</p>
			</div>
			{#if !confirmReload}
				<button
					class="btn btn-warning btn-outline btn-sm"
					type="button"
					disabled={busy}
					onclick={() => (confirmReload = true)}
				>
					<RefreshIcon class="size-4" aria-hidden="true" />
					Reload configuration
				</button>
			{/if}
		</div>

		{#if confirmReload}
			<div class="alert alert-warning mt-4 items-start text-sm" role="alert">
				<DangerTriangleIcon class="mt-0.5 size-5 shrink-0" aria-hidden="true" />
				<div class="min-w-0">
					<p class="font-semibold">Reloading configuration is a dangerous operation.</p>
					<p class="mt-1">
						It may restart, stop, or reconfigure services and can cause unexpected behavior.
					</p>
					<div class="mt-3 flex flex-wrap gap-2">
						<button
							class="btn btn-warning btn-sm"
							type="button"
							disabled={busy}
							onclick={() => {
								confirmReload = false;
								onReload();
							}}
						>
							{busy ? 'Reloading…' : 'Reload now'}
						</button>
						<button
							class="btn btn-ghost btn-sm"
							type="button"
							disabled={busy}
							onclick={() => (confirmReload = false)}
						>
							Cancel
						</button>
					</div>
				</div>
			</div>
		{/if}
	</section>

	<a class="btn btn-error btn-outline w-full" href="/pages/logout.html">Logout</a>
</div>
