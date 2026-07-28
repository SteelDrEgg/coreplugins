<script lang="ts">
	let {
		src = '/Arupa.svg',
		class: className = ''
	}: {
		src?: string;
		class?: string;
	} = $props();

	let currentSource = $state('/Arupa.svg');

	$effect(() => {
		currentSource = src.trim() || '/Arupa.svg';
	});

	function useFallback() {
		if (currentSource !== '/Arupa.svg') currentSource = '/Arupa.svg';
	}
</script>

<img
	class={`block object-contain ${className}`}
	src={currentSource}
	alt=""
	aria-hidden="true"
	referrerpolicy="no-referrer"
	onerror={useFallback}
/>
