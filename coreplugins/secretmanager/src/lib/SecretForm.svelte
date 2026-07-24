<script lang="ts">
	import type { SecretMeta } from './types';
	import type { WriteSecretInput } from './api';

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
		editing ? 'Leave * unchanged, or enter a new value to replace it' : 'Secret value'
	);
	let passphrasePlaceholder = $derived(
		editing
			? editing.encryption === 'scrypt'
				? '••••••••••••'
				: 'Enter a passphrase to protect the new value'
			: 'Used to encrypt secret'
	);

	function handleSubmit(event: SubmitEvent) {
		event.preventDefault();
		if (!value && !window.confirm('Save an empty value?')) return;
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
		<span class="text-sm font-semibold text-base-content/70">Name</span>
		<input
			class="input w-full font-mono"
			required
			maxlength="128"
			placeholder="Unique identifier"
			disabled={editing !== null || busy}
			bind:value={name}
		/>
	</label>
	<label class="grid gap-2">
		<span class="text-sm font-semibold text-base-content/70">Description (Optional)</span>
		<input
			class="input w-full"
			maxlength="240"
			placeholder="Set an alias or comment"
			disabled={busy}
			bind:value={description}
		/>
	</label>
	<label class="grid gap-2">
		<span class="text-sm font-semibold text-base-content/70">Value</span>
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
		<span class="text-sm font-semibold text-base-content/70">Passphrase (optional)</span>
		<input
			class="input w-full font-mono"
			type="password"
			autocomplete="new-password"
			placeholder={passphrasePlaceholder}
			disabled={busy}
			bind:value={passphrase}
		/>
		<span class="text-xs text-base-content/60">
			Don't set if you're not sure that you need it.<br />
			It takes extra time to decrypt<br />
			You cannot retrieve secret if you forget passphrase
		</span>
	</label>
	<label class="grid gap-2">
		<span class="text-sm font-semibold text-base-content/70">Allowed plugins</span>
		<textarea
			class="textarea min-h-24 w-full font-mono text-sm"
			placeholder="one-plugin-per-line"
			disabled={busy}
			bind:value={allowedPlugins}
		></textarea>
		<span class="text-xs text-base-content/60">Empty means no plugin can access this secret.</span>
	</label>
	<div class="flex flex-wrap justify-end gap-2">
		<button class="btn" type="button" disabled={busy} onclick={onCancel}>Cancel</button>
		<button class="btn btn-primary" type="submit" disabled={busy}>
			{editing ? 'Update secret' : 'Add secret'}
		</button>
	</div>
</form>
