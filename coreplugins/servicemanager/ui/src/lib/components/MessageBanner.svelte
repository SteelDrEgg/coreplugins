<script lang="ts">
	import type { BannerMessage } from '$lib/utils/types';

	let { message }: { message: BannerMessage | null } = $props();

	const classes: Record<BannerMessage['kind'], string> = {
		info: 'alert-info',
		success: 'alert-success',
		warning: 'alert-warning',
		error: 'alert-error'
	};
</script>

{#if message}
	<div class={`alert ${classes[message.kind]} text-sm`} role="status" aria-live="polite">
		<span>{message.text}</span>
	</div>
{/if}
