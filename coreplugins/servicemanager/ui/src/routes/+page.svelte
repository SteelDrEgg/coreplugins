<script lang="ts">
	import { onMount } from 'svelte';
	import DirectoryPanel from '$components/DirectoryPanel.svelte';
	import MessageBanner from '$components/MessageBanner.svelte';
	import RuntimePanel from '$components/RuntimePanel.svelte';
	import ServiceCatalog from '$components/ServiceCatalog.svelte';
	import ServiceHeader from '$components/ServiceHeader.svelte';
	import ServiceStats from '$components/ServiceStats.svelte';
	import { ServiceManager } from '$stores/serviceManager.svelte';
	import { connectTheme } from '$utils/theme';

	const manager = new ServiceManager();

	onMount(() => {
		const disconnectTheme = connectTheme();
		void manager.refresh();
		return disconnectTheme;
	});
</script>

<svelte:head>
	<title>Services</title>
	<meta name="description" content="Discover, configure, and manage Arupa services" />
</svelte:head>

<main
	class="mx-auto grid w-full max-w-[1320px] gap-4 px-4 py-6 sm:py-8"
	aria-busy={manager.busy}
>
	<ServiceHeader
		serviceDir={manager.serviceDir}
		busy={manager.busy}
		onRefresh={() => void manager.refresh(true)}
	/>

	<MessageBanner message={manager.message} />
	<ServiceStats services={manager.discovered} />

	<section class="grid min-w-0 gap-4 lg:grid-cols-[minmax(0,1fr)_320px]">
		<ServiceCatalog
			services={manager.discovered}
			loaded={manager.loaded}
			busy={manager.busy}
			onAction={(action, name) => void manager.runAction(action, name)}
		/>

		<aside class="grid min-w-0 gap-4 self-start">
			<DirectoryPanel
				serviceDir={manager.serviceDir}
				tempDir={manager.tempDir}
				requiresRestart={manager.tempDirRequiresRestart}
				busy={manager.busy}
				onSave={(serviceDir, tempDir) => void manager.saveDirectories(serviceDir, tempDir)}
			/>
			<RuntimePanel services={manager.running} />
		</aside>
	</section>
</main>
