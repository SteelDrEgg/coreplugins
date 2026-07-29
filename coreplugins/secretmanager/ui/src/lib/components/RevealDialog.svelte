<script lang="ts">
	import * as m from '$lib/paraglide/messages.js';
	import { locale } from '$lib/utils/locale';
	import type { MessageKind } from '../utils/types';

	let {
		open,
		title,
		value,
		warning,
		onClose,
		onMessage
	}: {
		open: boolean;
		title: string;
		value: string;
		warning: string;
		onClose: () => void;
		onMessage: (text: string, kind: MessageKind) => void;
	} = $props();

	let dialogEl: HTMLDialogElement | undefined = $state();
	let localeOptions = $derived({ locale: $locale });

	$effect(() => {
		if (!dialogEl) return;
		if (open && !dialogEl.open) {
			dialogEl.showModal();
		} else if (!open && dialogEl.open) {
			dialogEl.close();
		}
	});

	async function copyValue() {
		try {
			await navigator.clipboard.writeText(value);
			onMessage(m.copied_to_clipboard({}, localeOptions), 'success');
		} catch {
			dialogEl?.querySelector('textarea')?.select();
			onMessage(m.clipboard_access_failed({}, localeOptions), 'info');
		}
	}
</script>

<dialog class="modal" bind:this={dialogEl} onclose={onClose}>
	<div class="modal-box max-w-2xl">
		<h3 class="text-lg font-semibold">{title}</h3>
		<p class="mt-2 text-sm text-warning">{warning}</p>
		<textarea class="textarea mt-4 min-h-40 w-full select-all font-mono text-sm" readonly value={value}
		></textarea>
		<div class="modal-action">
			<button class="btn" type="button" onclick={copyValue}>{m.copy({}, localeOptions)}</button>
			<button class="btn btn-primary" type="button" onclick={() => dialogEl?.close()}>
				{m.close({}, localeOptions)}
			</button>
		</div>
	</div>
	<form method="dialog" class="modal-backdrop">
		<button aria-label={m.close_dialog({}, localeOptions)}>{m.close({}, localeOptions)}</button>
	</form>
</dialog>
