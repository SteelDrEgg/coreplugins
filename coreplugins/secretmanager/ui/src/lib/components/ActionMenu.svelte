<script lang="ts">
	import * as m from '$lib/paraglide/messages.js';
	import { locale } from '$lib/utils/locale';

	let {
		anchor,
		onEdit,
		onReveal,
		onDelete,
		onClose
	}: {
		anchor: HTMLElement | null;
		onEdit: () => void;
		onReveal: () => void;
		onDelete: () => void;
		onClose: () => void;
	} = $props();

	let menuEl: HTMLUListElement | undefined = $state();
	let left = $state(0);
	let top = $state(0);
	let localeOptions = $derived({ locale: $locale });

	$effect(() => {
		if (!anchor || !menuEl) return;

		const reposition = () => {
			if (!anchor || !menuEl) return;
			const buttonRect = anchor.getBoundingClientRect();
			const menuWidth = menuEl.offsetWidth;
			const menuHeight = menuEl.offsetHeight;
			left = Math.max(8, Math.min(buttonRect.right - menuWidth, window.innerWidth - menuWidth - 8));
			const belowTop = buttonRect.bottom + 4;
			top =
				belowTop + menuHeight <= window.innerHeight - 8
					? belowTop
					: Math.max(8, buttonRect.top - menuHeight - 4);
		};
		reposition();

		// Clicks on any row's action-menu trigger are excluded here so that
		// switching from one row's menu to another doesn't briefly close and
		// immediately reopen (the trigger's own click handler owns that).
		const handleOutsideClick = (event: MouseEvent) => {
			const target = event.target as Element | null;
			if (target?.closest('.action-menu-trigger') || menuEl?.contains(target)) return;
			onClose();
		};
		const handleKeydown = (event: KeyboardEvent) => {
			if (event.key === 'Escape') onClose();
		};

		window.addEventListener('resize', onClose);
		window.addEventListener('scroll', onClose, true);
		document.addEventListener('click', handleOutsideClick);
		document.addEventListener('keydown', handleKeydown);

		return () => {
			window.removeEventListener('resize', onClose);
			window.removeEventListener('scroll', onClose, true);
			document.removeEventListener('click', handleOutsideClick);
			document.removeEventListener('keydown', handleKeydown);
		};
	});
</script>

{#if anchor}
	<ul
		bind:this={menuEl}
		class="menu z-50 w-32 rounded-box border border-base-300 bg-base-100 p-2 shadow-lg"
		style={`position: fixed; left: ${left}px; top: ${top}px;`}
		role="menu"
	>
		<li><button type="button" onclick={onEdit}>{m.edit({}, localeOptions)}</button></li>
		<li><button type="button" onclick={onReveal}>{m.reveal({}, localeOptions)}</button></li>
		<li>
			<button type="button" class="text-error" onclick={onDelete}>
				{m.delete({}, localeOptions)}
			</button>
		</li>
	</ul>
{/if}
