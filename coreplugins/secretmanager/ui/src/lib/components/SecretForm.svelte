<script lang="ts">
	import * as m from '$lib/paraglide/messages.js';
	import { locale } from '$lib/utils/locale';
	import type { SecretInfo, Protection } from '../utils/types';
	import type { SaveSecretInput, UpdateSecretInput, SecretValueInput } from '../utils/api';

	let {
		editing,
		busy,
		onSave,
		onCancel
	}: {
		editing: SecretInfo | null;
		busy: boolean;
		onSave: (command: SaveSecretInput) => void;
		onCancel: () => void;
	} = $props();

	let name = $state('');
	let description = $state('');
	let value = $state('');
	let protection: Protection = $state('identity');
	let passphrase = $state('');
	let allowedPlugins = $state('');
	let replaceValue = $state(false);
	let localeOptions = $derived({ locale: $locale });
	let changingValue = $derived(editing === null || replaceValue);
	let plugins = $derived(
		[...new Set(allowedPlugins.split(/[\n,]/).map((item) => item.trim()).filter(Boolean))].sort()
	);
	let descriptionChanged = $derived(editing !== null && description.trim() !== editing.description);
	let pluginsChanged = $derived(
		editing !== null && JSON.stringify(plugins) !== JSON.stringify([...editing.allowed_plugins].sort())
	);
	let canSave = $derived(editing === null || replaceValue || descriptionChanged || pluginsChanged);

	$effect(() => {
		name = editing?.name || '';
		description = editing?.description || '';
		allowedPlugins = (editing?.allowed_plugins || []).join('\n');
		protection = editing?.protection || 'identity';
		value = '';
		passphrase = '';
		replaceValue = false;
	});

	function handleSubmit(event: SubmitEvent) {
		event.preventDefault();
		if (busy || !canSave) return;
		if (changingValue && !value && !window.confirm(m.confirm_empty_value({}, localeOptions))) return;
		const replacement: SecretValueInput = {
			plaintext: value,
			protection,
			...(protection === 'passphrase' ? { passphrase } : {})
		};
		if (editing) {
			const input: UpdateSecretInput = { name: editing.name };
			if (descriptionChanged) input.description = description.trim();
			if (pluginsChanged) input.allowed_plugins = plugins;
			if (replaceValue) input.value = replacement;
			onSave({ kind: 'update', input });
		} else {
			onSave({
				kind: 'create',
				input: {
					name: name.trim(),
					description: description.trim(),
					allowed_plugins: plugins,
					value: replacement
				}
			});
		}
	}
</script>

<form class="grid gap-4 p-4" onsubmit={handleSubmit}>
	<label class="grid gap-2">
		<span class="text-sm font-semibold text-base-content/70">{m.name({}, localeOptions)}</span>
		<input
			class="input w-full font-mono"
			required
			maxlength="128"
			placeholder={m.unique_identifier({}, localeOptions)}
			disabled={editing !== null || busy}
			bind:value={name}
		/>
	</label>
	<label class="grid gap-2">
		<span class="text-sm font-semibold text-base-content/70">{m.description_optional({}, localeOptions)}</span>
		<input
			class="input w-full"
			maxlength="240"
			placeholder={m.description_placeholder({}, localeOptions)}
			disabled={busy}
			bind:value={description}
		/>
	</label>
	{#if editing}
		<div class="grid gap-2">
			<label class="flex items-center gap-2">
				<input
					class="checkbox checkbox-sm"
					type="checkbox"
					disabled={busy}
					bind:checked={replaceValue}
				/>
				<span class="text-sm font-semibold">{m.replace_value({}, localeOptions)}</span>
			</label>
			<p class="text-xs text-base-content/60">{m.keep_value_help({}, localeOptions)}</p>
		</div>
	{/if}
	{#if changingValue}
		<label class="grid gap-2">
			<span class="text-sm font-semibold text-base-content/70">{m.value({}, localeOptions)}</span>
			<textarea
				class="textarea min-h-32 w-full font-mono text-sm"
				autocomplete="off"
				placeholder={m.secret_value({}, localeOptions)}
				disabled={busy}
				bind:value
			></textarea>
		</label>
		<label class="grid gap-2">
			<span class="text-sm font-semibold text-base-content/70">{m.protection({}, localeOptions)}</span>
			<select class="select w-full" disabled={busy} bind:value={protection}>
				<option value="identity">{m.identity_protection({}, localeOptions)}</option>
				<option value="passphrase">{m.passphrase_protection({}, localeOptions)}</option>
			</select>
		</label>
		{#if protection === 'passphrase'}
			<label class="grid gap-2">
				<span class="text-sm font-semibold text-base-content/70">{m.passphrase({}, localeOptions)}</span>
				<input
					class="input w-full font-mono"
					type="password"
					required
					autocomplete="new-password"
					placeholder={m.passphrase_placeholder({}, localeOptions)}
					disabled={busy}
					bind:value={passphrase}
				/>
				<span class="text-xs text-base-content/60">
					{m.passphrase_help_slow({}, localeOptions)}<br />
					{m.passphrase_help_unrecoverable({}, localeOptions)}
				</span>
			</label>
		{/if}
	{/if}
	<label class="grid gap-2">
		<span class="text-sm font-semibold text-base-content/70">{m.allowed_plugins({}, localeOptions)}</span>
		<textarea
			class="textarea min-h-24 w-full font-mono text-sm"
			placeholder={m.allowed_plugins_placeholder({}, localeOptions)}
			disabled={busy}
			bind:value={allowedPlugins}
		></textarea>
		<span class="text-xs text-base-content/60">{m.allowed_plugins_empty({}, localeOptions)}</span>
	</label>
	<div class="flex flex-wrap justify-end gap-2">
		<button class="btn" type="button" disabled={busy} onclick={onCancel}>{m.cancel({}, localeOptions)}</button>
		<button class="btn btn-primary" type="submit" disabled={busy || !canSave}>
			{editing ? m.update_secret({}, localeOptions) : m.add_secret({}, localeOptions)}
		</button>
	</div>
</form>
