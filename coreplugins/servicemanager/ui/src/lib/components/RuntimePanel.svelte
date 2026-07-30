<script lang="ts">
	import RuntimeList from './RuntimeList.svelte';
	import * as m from '$lib/paraglide/messages.js';
	import { locale } from '$lib/utils/locale';
	import type { RunningService } from '$lib/utils/types';

	let { services }: { services: RunningService[] } = $props();
	let localeOptions = $derived({ locale: $locale });
</script>

<div class="card min-w-0 rounded-lg border border-base-300 bg-base-100 shadow-sm">
	<div class="border-b border-base-300 p-4 font-semibold">{m.runtime({}, localeOptions)}</div>
	<div class="grid gap-3 p-4">
		<RuntimeList {services} />
	</div>
</div>
