<script lang="ts">
	import type { BannerMessage } from './types';

	let { message }: { message: BannerMessage | null } = $props();

	const classByKind: Record<string, string> = {
		info: 'alert alert-info',
		success: 'alert alert-success',
		warning: 'alert alert-warning',
		error: 'alert alert-error'
	};
</script>

{#if message}
	<div class={classByKind[message.kind]} role="status">{message.text}</div>
{/if}
