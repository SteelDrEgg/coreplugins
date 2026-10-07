<script lang="ts">
	import * as m from '$lib/paraglide/messages.js';
	import { locale } from '$lib/utils/locale';
	import type { ServiceStatus } from '$lib/utils/types';
	let { status }: { status: ServiceStatus } = $props();
	const messages = { discovered: m.discovered, running: m.running, degraded: m.degraded, starting: m.starting, stopping: m.stopping, failed: m.failed, backoff: m.backoff, exited: m.exited };
	let options = $derived({ locale: $locale });
	let color = $derived(status === 'running' ? 'badge-success' : status === 'failed' ? 'badge-error' : ['degraded', 'backoff'].includes(status) ? 'badge-warning' : ['starting', 'stopping'].includes(status) ? 'badge-info' : 'badge-ghost');
</script>

<span class="badge badge-soft badge-sm whitespace-nowrap {color}">{(messages[status] ?? m.unavailable)({}, options)}</span>
