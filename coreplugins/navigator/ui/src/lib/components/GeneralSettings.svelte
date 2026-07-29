<script lang="ts">
	import DangerTriangleIcon from '@iconify-svelte/mynaui/danger-triangle';
	import RefreshIcon from '@iconify-svelte/mynaui/refresh';
	import * as m from '$lib/paraglide/messages.js';
	import type { Locale } from '$lib/paraglide/runtime.js';
	import type { AvailableLanguage } from '$lib/utils/locale';

	let {
		darkTheme,
		selectedLanguage,
		languages,
		kernelVersion,
		busy,
		onThemeChange,
		onLanguageChange,
		onReload
	}: {
		darkTheme: boolean;
		selectedLanguage: Locale;
		languages: AvailableLanguage[];
		kernelVersion: string;
		busy: boolean;
		onThemeChange: (enabled: boolean) => void;
		onLanguageChange: (language: string) => void;
		onReload: () => void;
	} = $props();

	let confirmReload = $state(false);

	function languageLabel(language: AvailableLanguage): string {
		return language.nativeName && language.nativeName !== language.name
			? `${language.name} (${language.nativeName})`
			: language.name;
	}
</script>

<div class="space-y-4">
	<label class="flex items-center justify-between gap-4 rounded-lg border border-base-300 p-4">
		<span>
			<span class="block font-medium">{m.theme()}</span>
			<span class="text-xs text-base-content/60">{darkTheme ? m.theme_dark() : m.theme_light()}</span>
		</span>
		<input
			class="toggle toggle-primary"
			type="checkbox"
			checked={darkTheme}
			aria-label={m.toggle_dark_theme()}
			onchange={(event) => onThemeChange(event.currentTarget.checked)}
		/>
	</label>

	<label class="grid gap-2 rounded-lg border border-base-300 p-4">
		<span class="font-medium">{m.language()}</span>
		<select
			class="select w-full"
			value={selectedLanguage}
			aria-label={m.language()}
			onchange={(event) => onLanguageChange(event.currentTarget.value)}
		>
			{#each languages as language}
				<option value={language.code}>{languageLabel(language)}</option>
			{/each}
		</select>
	</label>

	<section class="rounded-lg border border-base-300 p-4" aria-labelledby="kernel-heading">
		<div class="flex items-center justify-between gap-3">
			<div>
				<h3 class="font-medium" id="kernel-heading">{m.kernel()}</h3>
				<p class="text-xs text-base-content/60">
					{m.kernel_version({ version: kernelVersion || m.loading() })}
				</p>
			</div>
			{#if !confirmReload}
				<button
					class="btn btn-warning btn-sm"
					type="button"
					disabled={busy}
					onclick={() => (confirmReload = true)}
				>
					<RefreshIcon class="size-4" aria-hidden="true" />
					{m.reload_configuration()}
				</button>
			{/if}
		</div>

		{#if confirmReload}
			<div class="alert alert-warning mt-4 items-start text-sm" role="alert">
				<DangerTriangleIcon class="mt-0.5 size-5 shrink-0" aria-hidden="true" />
				<div class="min-w-0">
					<p class="font-semibold">{m.reload_danger_title()}</p>
					<p class="mt-1">{m.reload_danger_description()}</p>
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
							{busy ? m.reloading() : m.reload_now()}
						</button>
						<button
							class="btn btn-ghost btn-sm"
							type="button"
							disabled={busy}
							onclick={() => (confirmReload = false)}
						>
							{m.cancel()}
						</button>
					</div>
				</div>
			</div>
		{/if}
	</section>

	<a class="btn btn-error w-full" href="/pages/logout.html">{m.logout()}</a>
</div>
