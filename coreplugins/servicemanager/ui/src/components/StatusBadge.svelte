<script lang="ts">
	import type { ServiceStatus } from '$utils/types';

	let { status }: { status?: ServiceStatus } = $props();

	let normalized = $derived(String(status || 'discovered').trim().toLowerCase());
	let label = $derived(
		({
			discovered: 'Inactive',
			starting: 'Starting',
			running: 'Running',
			degraded: 'Degraded',
			stopping: 'Stopping',
			failed: 'Failed'
		}[normalized] || normalized.charAt(0).toUpperCase() + normalized.slice(1))
	);
	let tone = $derived(
		({
			discovered: 'badge-warning',
			starting: 'badge-info',
			running: 'badge-success',
			degraded: 'badge-warning',
			stopping: 'badge-info',
			failed: 'badge-error'
		}[normalized] || 'badge-neutral')
	);
</script>

<span class={`badge badge-sm badge-soft ${tone}`}>{label}</span>
