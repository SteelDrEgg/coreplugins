<script lang="ts">
	import type { SecretMeta } from './types';
	import ActionMenu from './ActionMenu.svelte';

	let {
		keys,
		filterQuery,
		onEdit,
		onReveal,
		onDelete
	}: {
		keys: SecretMeta[];
		filterQuery: string;
		onEdit: (name: string) => void;
		onReveal: (name: string) => void;
		onDelete: (name: string) => void;
	} = $props();

	let openMenuName: string | null = $state(null);
	let openMenuAnchor: HTMLElement | null = $state(null);

	let filteredKeys = $derived.by(() => {
		const query = filterQuery.trim().toLowerCase();
		return keys.filter((item) => !query || item.name.toLowerCase().includes(query));
	});

	function toggleMenu(name: string, button: HTMLElement) {
		if (openMenuName === name) {
			closeMenu();
			return;
		}
		openMenuName = name;
		openMenuAnchor = button;
	}

	function closeMenu() {
		openMenuName = null;
		openMenuAnchor = null;
	}

	function withClose(action: (name: string) => void) {
		return () => {
			if (!openMenuName) return;
			const name = openMenuName;
			closeMenu();
			action(name);
		};
	}
</script>

{#if filteredKeys.length === 0}
	<div class="grid min-h-48 place-items-center p-6 text-center text-base-content/60">
		No secrets found
	</div>
{:else}
	<table class="table table-zebra">
		<thead>
			<tr>
				<th>Name</th>
				<th>Allowed plugins</th>
				<th>Updated</th>
			</tr>
		</thead>
		<tbody>
			{#each filteredKeys as item (item.name)}
				<tr>
					<td class="max-w-64">
						<div class="min-w-0 flex-1">
							<div class="truncate font-mono text-sm" title={item.name}>{item.name}</div>
							<div class="truncate text-xs text-base-content/60">
								{item.description || 'No description'}
							</div>
						</div>
					</td>
					<td class="max-w-52">
						<div class="truncate text-sm" title={(item.allowed_plugins || []).join(', ')}>
							{(item.allowed_plugins || []).join(', ') || 'None'}
						</div>
					</td>
					<td class="whitespace-nowrap text-xs text-base-content/60">
						<div class="flex items-center justify-between gap-2">
							<span>{item.updated_at || '-'}</span>
							<button
								type="button"
								class="action-menu-trigger btn btn-ghost btn-xs px-2"
								aria-haspopup="menu"
								aria-expanded={openMenuName === item.name}
								aria-label={`Actions for ${item.name}`}
								onclick={(event) => toggleMenu(item.name, event.currentTarget as HTMLElement)}
							>
								⋮
							</button>
						</div>
					</td>
				</tr>
			{/each}
		</tbody>
	</table>
{/if}

<ActionMenu
	anchor={openMenuAnchor}
	onEdit={withClose(onEdit)}
	onReveal={withClose(onReveal)}
	onDelete={withClose(onDelete)}
	onClose={closeMenu}
/>
