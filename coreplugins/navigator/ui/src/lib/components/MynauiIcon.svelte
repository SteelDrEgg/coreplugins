<script lang="ts">
	import KeyIcon from '@iconify-svelte/mynaui/key';
	import KeySolidIcon from '@iconify-svelte/mynaui/key-solid';
	import LetterMSquareIcon from '@iconify-svelte/mynaui/letter-m-square';
	import LetterMSquareSolidIcon from '@iconify-svelte/mynaui/letter-m-square-solid';
	import PuzzleIcon from '@iconify-svelte/mynaui/puzzle';
	import PuzzleSolidIcon from '@iconify-svelte/mynaui/puzzle-solid';
	import TerminalIcon from '@iconify-svelte/mynaui/terminal';
	import TerminalSolidIcon from '@iconify-svelte/mynaui/terminal-solid';
	import type { SvelteHTMLElements } from 'svelte/elements';

	type Props = Omit<SvelteHTMLElements['svg'], 'viewBox' | 'xmlns'> & {
		name?: string;
		width?: string;
		height?: string;
	};

	const icons = {
		key: KeyIcon,
		'key-solid': KeySolidIcon,
		'letter-m-square': LetterMSquareIcon,
		'letter-m-square-solid': LetterMSquareSolidIcon,
		puzzle: PuzzleIcon,
		'puzzle-solid': PuzzleSolidIcon,
		terminal: TerminalIcon,
		'terminal-solid': TerminalSolidIcon
	};

	type IconName = keyof typeof icons;

	let { name = '', ...props }: Props = $props();

	function iconName(value: string): string {
		const path = value.split(/[?#]/, 1)[0];
		const filename = path.split('/').pop() || path;
		return filename.replace(/\.svg$/i, '').toLowerCase();
	}

	let Icon = $derived(icons[iconName(name) as IconName] || PuzzleIcon);
</script>

<Icon {...props} />
