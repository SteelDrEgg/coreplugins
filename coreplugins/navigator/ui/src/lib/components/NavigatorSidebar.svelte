<script lang="ts">
	import CogFourIcon from '@iconify-svelte/mynaui/cog-four';
	import SidebarIcon from '@iconify-svelte/mynaui/sidebar';
	import * as m from '$lib/paraglide/messages.js';
	import { locale } from '$lib/utils/locale';
	import type { NavigationEntry } from '$lib/utils/types';
	import BrandIcon from './BrandIcon.svelte';
	import MynauiIcon from './MynauiIcon.svelte';

	let {
		entries,
		activeID,
		loading,
		mobileOpen,
		brandIcon,
		serverName,
		onselect,
		onopen,
		onclose,
		onsettings
	}: {
		entries: NavigationEntry[];
		activeID: string;
		loading: boolean;
		mobileOpen: boolean;
		brandIcon: string;
		serverName: string;
		onselect: (id: string) => void;
		onopen: () => void;
		onclose: () => void;
		onsettings: () => void;
	} = $props();

	let localeOptions = $derived({ locale: $locale });

	function iconName(entry: NavigationEntry, active: boolean): string {
		return (active ? entry.icon_solid || entry.icon : entry.icon) || 'puzzle';
	}
</script>

<header class="hidden items-center justify-between border-b border-base-300 bg-base-100 px-3 max-md:flex">
	<div class="flex min-w-0 items-center gap-2 font-semibold">
		<BrandIcon src={brandIcon} class="size-7 shrink-0" />
		<span class="truncate">{serverName || 'Arupa'}</span>
	</div>
	<button
		class="btn btn-square btn-ghost"
		type="button"
		aria-label={m.open_navigation({}, localeOptions)}
		onclick={onopen}
	>
		<SidebarIcon height="1.25em" aria-hidden="true" />
	</button>
</header>

<aside
	class:!translate-x-0={mobileOpen}
	class="flex min-h-0 flex-col border-r border-base-300 bg-base-100 max-md:fixed max-md:inset-y-0 max-md:left-0 max-md:z-40 max-md:w-20 max-md:-translate-x-full max-md:transition-transform"
>
	<div class="grid min-h-16 shrink-0 place-items-center border-b border-base-300">
		<BrandIcon src={brandIcon} class="size-9" />
	</div>

	<nav
		class="flex min-h-0 flex-1 flex-col items-center gap-2 overflow-y-auto px-2 py-3"
		aria-label={m.service_navigation({}, localeOptions)}
	>
		{#if loading}
			<div class="grid min-h-24 place-items-center text-base-content/50">
				<span
					class="loading loading-spinner loading-sm"
					aria-label={m.loading_services({}, localeOptions)}
				></span>
			</div>
		{:else}
			{#each entries as entry (entry.id)}
				{@const active = entry.id === activeID}
				<button
					class:text-primary={active}
					class="group flex min-h-[4.5rem] w-[3.75rem] flex-col items-center justify-center gap-1 border-0 bg-transparent p-0 font-medium text-base-content transition-colors hover:text-secondary focus-visible:text-secondary focus-visible:outline-none"
					type="button"
					title={entry.label}
					aria-current={active ? 'page' : undefined}
					onclick={() => onselect(entry.id)}
				>
					<span
						class:bg-primary={active}
						class:text-primary-content={active}
						class="grid size-9 place-items-center rounded-md transition-colors group-hover:bg-secondary group-hover:text-secondary-content group-focus-visible:bg-secondary group-focus-visible:text-secondary-content"
						aria-hidden="true"
					>
						<MynauiIcon name={iconName(entry, active)} height="1.25rem" />
					</span>
					<span class="w-full overflow-hidden text-ellipsis whitespace-nowrap text-center text-xs">
						{entry.label}
					</span>
				</button>
			{/each}
		{/if}
	</nav>

	<div class="grid shrink-0 place-items-center border-t border-base-300 p-2">
		<button
			class="btn btn-square btn-ghost"
			type="button"
			title={m.settings({}, localeOptions)}
			aria-label={m.settings({}, localeOptions)}
			onclick={onsettings}
		>
			<CogFourIcon height="1.25em" aria-hidden="true" />
		</button>
	</div>
</aside>

{#if mobileOpen}
	<button
		class="fixed inset-0 z-30 bg-neutral/40 md:hidden"
		type="button"
		aria-label={m.close_navigation({}, localeOptions)}
		onclick={onclose}
	></button>
{/if}
