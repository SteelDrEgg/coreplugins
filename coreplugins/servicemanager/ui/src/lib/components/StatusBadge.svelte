<script lang="ts">
	import * as m from '$lib/paraglide/messages.js';
	import { locale } from '$lib/utils/locale';
	import type { ServiceStatus } from '$lib/utils/types';

	let { status }: { status?: ServiceStatus } = $props();
	let localeOptions = $derived({ locale: $locale });

	let normalized = $derived(String(status || 'discovered').trim().toLowerCase());
	let label = $derived(
		({
			discovered: m.inactive({}, localeOptions),
			starting: m.starting({}, localeOptions),
			running: m.running({}, localeOptions),
			degraded: m.degraded({}, localeOptions),
			stopping: m.stopping({}, localeOptions),
			failed: m.failed({}, localeOptions)
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

<span class={`badge badge-sm ${tone}`}>{label}</span>
