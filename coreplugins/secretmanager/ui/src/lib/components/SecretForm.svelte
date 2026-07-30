<script lang="ts">
	import * as m from '$lib/paraglide/messages.js';
	import { locale } from '$lib/utils/locale';
	import type { SecretMeta } from '../utils/types';
	import type { WriteSecretInput } from '../utils/api';

	let {
		editing,
		busy,
		onSave,
		onCancel
	}: {
		editing: SecretMeta | null;
		busy: boolean;
		onSave: (input: WriteSecretInput, isUpdate: boolean) => void;
		onCancel: () => void;
	} = $props();

	let name = $state('');
	let description = $state('');
	let value = $state('');
	let passphrase = $state('');
	let allowedPlugins = $state('');
	let localeOptions = $derived({ locale: $locale });

	// Reset the draft whenever the target of editing changes (including
	// switching back to "add" mode, where editing becomes null).
	$effect(() => {
		if (editing) {
			name = editing.name;
			description = editing.description || '';
			value = '*';
			passphrase = '';
			allowedPlugins = (editing.allowed_plugins || []).join('\n');
		} else {
			name = '';
			description = '';
			value = '';
			passphrase = '';
			allowedPlugins = '';
		}
	});

	let valuePlaceholder = $derived(
		editing
			? m.value_edit_placeholder({}, localeOptions)
			: m.secret_value({}, localeOptions)
	);
	let passphrasePlaceholder = $derived(
		editing
			? editing.encryption === 'scrypt'
				? '••••••••••••'
				: m.passphrase_new_value_placeholder({}, localeOptions)
			: m.passphrase_encrypt_placeholder({}, localeOptions)
	);

	function handleSubmit(event: SubmitEvent) {
		event.preventDefault();
		if (!value && !window.confirm(m.confirm_empty_value({}, localeOptions))) return;
		const plugins = allowedPlugins
			.split(/[\n,]/)
			.map((item) => item.trim())
			.filter(Boolean);
		onSave(
			{
				name: name.trim(),
				description: description.trim(),
				value,
				passphrase,
				allowed_plugins: plugins
			},
			editing !== null
		);
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
		<span class="text-sm font-semibold text-base-content/70">
			{m.description_optional({}, localeOptions)}
		</span>
		<input
			class="input w-full"
			maxlength="240"
			placeholder={m.description_placeholder({}, localeOptions)}
			disabled={busy}
			bind:value={description}
		/>
	</label>
	<label class="grid gap-2">
		<span class="text-sm font-semibold text-base-content/70">{m.value({}, localeOptions)}</span>
		<textarea
			class="textarea min-h-32 w-full font-mono text-sm"
			required
			autocomplete="off"
			placeholder={valuePlaceholder}
			disabled={busy}
			bind:value
		></textarea>
	</label>
	<label class="grid gap-2">
		<span class="text-sm font-semibold text-base-content/70">
			{m.passphrase_optional({}, localeOptions)}
		</span>
		<input
			class="input w-full font-mono"
			type="password"
			autocomplete="new-password"
			placeholder={passphrasePlaceholder}
			disabled={busy}
			bind:value={passphrase}
		/>
		<span class="text-xs text-base-content/60">
			{m.passphrase_help_needed({}, localeOptions)}<br />
			{m.passphrase_help_slow({}, localeOptions)}<br />
			{m.passphrase_help_unrecoverable({}, localeOptions)}
		</span>
	</label>
	<label class="grid gap-2">
		<span class="text-sm font-semibold text-base-content/70">
			{m.allowed_plugins({}, localeOptions)}
		</span>
		<textarea
			class="textarea min-h-24 w-full font-mono text-sm"
			placeholder={m.allowed_plugins_placeholder({}, localeOptions)}
			disabled={busy}
			bind:value={allowedPlugins}
		></textarea>
		<span class="text-xs text-base-content/60">
			{m.allowed_plugins_empty({}, localeOptions)}
		</span>
	</label>
	<div class="flex flex-wrap justify-end gap-2">
		<button class="btn" type="button" disabled={busy} onclick={onCancel}>
			{m.cancel({}, localeOptions)}
		</button>
		<button class="btn btn-primary" type="submit" disabled={busy}>
			{editing ? m.update_secret({}, localeOptions) : m.add_secret({}, localeOptions)}
		</button>
	</div>
</form>
