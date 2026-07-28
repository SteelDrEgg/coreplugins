<script lang="ts">
	import ChevronDownIcon from '@iconify-svelte/mynaui/chevron-down';
	import ChevronUpIcon from '@iconify-svelte/mynaui/chevron-up';
	import GripVerticalIcon from '@iconify-svelte/mynaui/grip-vertical';
	import PlusIcon from '@iconify-svelte/mynaui/plus';
	import TrashIcon from '@iconify-svelte/mynaui/trash';
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
	let newName = $state('');
	let draggedName = $state('');
	let dropTarget = $state('');

	$effect(() => {
		icon = config.icon;
		order = [...config.order];
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
		<span class="font-medium">Brand icon URL</span>
		<span class="text-xs text-base-content/60">
			Absolute and cross-origin URLs are supported. Leave blank to use /Arupa.svg.
		</span>
		<div class="flex items-center gap-3">
			<div class="grid size-11 shrink-0 place-items-center rounded-lg border border-base-300 bg-white p-2">
				<BrandIcon src={icon} class="size-full" />
			</div>
			<input
				class="input min-w-0 flex-1 font-mono text-xs"
				type="text"
				inputmode="url"
				placeholder="/Arupa.svg or https://…"
				bind:value={icon}
				disabled={busy}
			/>
		</div>
	</label>

	<section aria-labelledby="service-order-heading">
		<div class="mb-2">
			<h3 class="font-medium" id="service-order-heading">Service order</h3>
			<p class="text-xs text-base-content/60">
				Add service names, then drag the blocks or use the arrow buttons to reorder them.
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
				placeholder="service name"
				aria-label="Service name"
				bind:value={newName}
				disabled={busy}
			/>
			<button class="btn btn-primary" type="submit" disabled={busy || !newName.trim()}>
				<PlusIcon class="size-4" aria-hidden="true" />
				Add
			</button>
		</form>

		<div class="grid gap-2" role="list" aria-live="polite">
			{#if order.length === 0}
				<div class="rounded-lg border border-dashed border-base-300 p-6 text-center text-sm text-base-content/50">
					No explicit order. Services will be sorted by display name.
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
							title="Move up"
							aria-label={`Move ${name} up`}
							disabled={busy || index === 0}
							onclick={() => move(name, -1)}
						>
							<ChevronUpIcon class="size-4" aria-hidden="true" />
						</button>
						<button
							class="btn btn-square btn-ghost btn-xs"
							type="button"
							title="Move down"
							aria-label={`Move ${name} down`}
							disabled={busy || index === order.length - 1}
							onclick={() => move(name, 1)}
						>
							<ChevronDownIcon class="size-4" aria-hidden="true" />
						</button>
						<button
							class="btn btn-square btn-ghost btn-xs text-error"
							type="button"
							title="Remove"
							aria-label={`Remove ${name}`}
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

	<button
		class="btn btn-primary w-full"
		type="button"
		disabled={busy}
		onclick={() => onsave({ icon: icon.trim() || '/Arupa.svg', order })}
	>
		{busy ? 'Saving…' : 'Save Navigator settings'}
	</button>
</div>
