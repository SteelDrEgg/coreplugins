<script lang="ts">
	import ChevronDownIcon from '@iconify-svelte/mynaui/chevron-down';
	import ChevronUpIcon from '@iconify-svelte/mynaui/chevron-up';
	import EyeOffIcon from '@iconify-svelte/mynaui/eye-off';
	import GripVerticalIcon from '@iconify-svelte/mynaui/grip-vertical';
	import PlusIcon from '@iconify-svelte/mynaui/plus';
	import TrashIcon from '@iconify-svelte/mynaui/trash';
	import * as m from '$lib/paraglide/messages.js';
	import { locale } from '$lib/utils/locale';
	import type { NavigatorConfig } from '$lib/utils/types';
	import BrandIcon from './BrandIcon.svelte';

	let {
		config,
		busy,
		onsave
	}: {
		config: NavigatorConfig;
		busy: boolean;
		onsave: (config: NavigatorConfig) => void;
	} = $props();

	let icon = $state('/Arupa.svg');
	let order: string[] = $state([]);
	let hide: string[] = $state([]);
	let newName = $state('');
	let newHiddenService = $state('');
	let draggedName = $state('');
	let dropTarget = $state('');
	let localeOptions = $derived({ locale: $locale });

	$effect(() => {
		icon = config.icon;
		order = [...config.order];
		hide = [...config.hide];
	});

	function addName() {
		const name = newName.trim();
		if (!name || order.includes(name)) return;
		order = [...order, name];
		newName = '';
	}

	function move(name: string, offset: number) {
		const index = order.indexOf(name);
		const target = index + offset;
		if (index < 0 || target < 0 || target >= order.length) return;
		const next = [...order];
		[next[index], next[target]] = [next[target], next[index]];
		order = next;
	}

	function remove(name: string) {
		order = order.filter((candidate) => candidate !== name);
	}

	function addHiddenService() {
		const name = newHiddenService.trim();
		if (!name || hide.includes(name)) return;
		hide = [...hide, name];
		newHiddenService = '';
	}

	function showService(name: string) {
		hide = hide.filter((candidate) => candidate !== name);
	}

	function startDrag(event: DragEvent, name: string) {
		draggedName = name;
		event.dataTransfer?.setData('text/plain', name);
		if (event.dataTransfer) event.dataTransfer.effectAllowed = 'move';
	}

	function dragOver(event: DragEvent, name: string) {
		if (!draggedName || draggedName === name) return;
		event.preventDefault();
		dropTarget = name;
		if (event.dataTransfer) event.dataTransfer.dropEffect = 'move';
	}

	function drop(event: DragEvent, targetName: string) {
		event.preventDefault();
		const sourceName = draggedName || event.dataTransfer?.getData('text/plain') || '';
		const sourceIndex = order.indexOf(sourceName);
		const targetIndex = order.indexOf(targetName);
		if (sourceIndex >= 0 && targetIndex >= 0 && sourceIndex !== targetIndex) {
			const next = order.filter((name) => name !== sourceName);
			next.splice(targetIndex, 0, sourceName);
			order = next;
		}
		draggedName = '';
		dropTarget = '';
	}
</script>

<div class="space-y-5">
	<label class="grid gap-2">
		<span class="font-medium">{m.brand_icon_url({}, localeOptions)}</span>
		<span class="text-xs text-base-content/60">
			{m.brand_icon_help({}, localeOptions)}
		</span>
		<div class="flex items-center gap-3">
			<div class="grid size-11 shrink-0 place-items-center rounded-lg border border-base-300 bg-white p-2">
				<BrandIcon src={icon} class="size-full" />
			</div>
			<input
				class="input min-w-0 flex-1 font-mono text-xs"
				type="text"
				inputmode="url"
				placeholder={m.brand_icon_placeholder({}, localeOptions)}
				bind:value={icon}
				disabled={busy}
			/>
		</div>
	</label>

	<section aria-labelledby="service-order-heading">
		<div class="mb-2">
			<h3 class="font-medium" id="service-order-heading"
				>{m.service_order({}, localeOptions)}</h3
			>
			<p class="text-xs text-base-content/60">
				{m.service_order_help({}, localeOptions)}
			</p>
		</div>

		<form
			class="mb-3 flex gap-2"
			onsubmit={(event) => {
				event.preventDefault();
				addName();
			}}
		>
			<input
				class="input min-w-0 flex-1"
				placeholder={m.service_name({}, localeOptions)}
				aria-label={m.service_name({}, localeOptions)}
				bind:value={newName}
				disabled={busy}
			/>
			<button class="btn btn-primary" type="submit" disabled={busy || !newName.trim()}>
				<PlusIcon class="size-4" aria-hidden="true" />
				{m.add({}, localeOptions)}
			</button>
		</form>

		<div class="grid gap-2" role="list" aria-live="polite">
			{#if order.length === 0}
				<div class="rounded-lg border border-dashed border-base-300 p-6 text-center text-sm text-base-content/50">
					{m.no_explicit_order({}, localeOptions)}
				</div>
			{:else}
				{#each order as name, index (name)}
					<div
						role="listitem"
						class:border-primary={dropTarget === name}
						class:opacity-50={draggedName === name}
						class="flex items-center gap-2 rounded-lg border border-base-300 bg-base-100 p-2 transition-colors"
						draggable="true"
						ondragstart={(event) => startDrag(event, name)}
						ondragover={(event) => dragOver(event, name)}
						ondrop={(event) => drop(event, name)}
						ondragend={() => {
							draggedName = '';
							dropTarget = '';
						}}
					>
						<GripVerticalIcon
							class="size-5 shrink-0 cursor-grab text-base-content/40"
							aria-hidden="true"
						/>
						<span class="min-w-0 flex-1 truncate font-mono text-sm">{name}</span>
						<button
							class="btn btn-square btn-ghost btn-xs"
							type="button"
							title={m.move_up({}, localeOptions)}
							aria-label={m.move_service_up({ name }, localeOptions)}
							disabled={busy || index === 0}
							onclick={() => move(name, -1)}
						>
							<ChevronUpIcon class="size-4" aria-hidden="true" />
						</button>
						<button
							class="btn btn-square btn-ghost btn-xs"
							type="button"
							title={m.move_down({}, localeOptions)}
							aria-label={m.move_service_down({ name }, localeOptions)}
							disabled={busy || index === order.length - 1}
							onclick={() => move(name, 1)}
						>
							<ChevronDownIcon class="size-4" aria-hidden="true" />
						</button>
						<button
							class="btn btn-square btn-ghost btn-xs text-error"
							type="button"
							title={m.remove({}, localeOptions)}
							aria-label={m.remove_service({ name }, localeOptions)}
							disabled={busy}
							onclick={() => remove(name)}
						>
							<TrashIcon class="size-4" aria-hidden="true" />
						</button>
					</div>
				{/each}
			{/if}
		</div>
	</section>

	<section aria-labelledby="hidden-services-heading">
		<div class="mb-2">
			<h3 class="font-medium" id="hidden-services-heading"
				>{m.hidden_services({}, localeOptions)}</h3
			>
			<p class="text-xs text-base-content/60">{m.hidden_services_help({}, localeOptions)}</p>
		</div>

		<form
			class="mb-3 flex gap-2"
			onsubmit={(event) => {
				event.preventDefault();
				addHiddenService();
			}}
		>
			<input
				class="input min-w-0 flex-1"
				placeholder={m.hidden_service_name({}, localeOptions)}
				aria-label={m.hidden_service_name({}, localeOptions)}
				bind:value={newHiddenService}
				disabled={busy}
			/>
			<button
				class="btn btn-primary"
				type="submit"
				disabled={busy || !newHiddenService.trim()}
			>
				<EyeOffIcon class="size-4" aria-hidden="true" />
				{m.hide({}, localeOptions)}
			</button>
		</form>

		<div class="grid gap-2" role="list" aria-live="polite">
			{#if hide.length === 0}
				<div
					class="rounded-lg border border-dashed border-base-300 p-6 text-center text-sm text-base-content/50"
				>
					{m.no_hidden_services({}, localeOptions)}
				</div>
			{:else}
				{#each hide as name (name)}
					<div
						class="flex items-center gap-2 rounded-lg border border-base-300 bg-base-100 p-2"
						role="listitem"
					>
						<EyeOffIcon class="size-5 shrink-0 text-base-content/40" aria-hidden="true" />
						<span class="min-w-0 flex-1 truncate font-mono text-sm">{name}</span>
						<button
							class="btn btn-square btn-ghost btn-xs"
							type="button"
							title={m.show({}, localeOptions)}
							aria-label={m.show_service({ name }, localeOptions)}
							disabled={busy}
							onclick={() => showService(name)}
						>
							<TrashIcon class="size-4" aria-hidden="true" />
						</button>
					</div>
				{/each}
			{/if}
		</div>
	</section>

	<button
		class="btn btn-primary w-full"
		type="button"
		disabled={busy}
		onclick={() => onsave({ icon: icon.trim() || '/Arupa.svg', order, hide })}
	>
		{busy ? m.saving({}, localeOptions) : m.save_navigator_settings({}, localeOptions)}
	</button>
</div>
