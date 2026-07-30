<script lang="ts">
	import { locale, messagesFor } from '$lib/locale';

	$: messages = messagesFor($locale);
</script>

<svelte:head>
	<title>{messages.forbidden_page_title}</title>
	<meta name="description" content={messages.forbidden_page_description} />
</svelte:head>

<main class="grid min-h-screen place-items-center bg-base-200 px-4 py-8 text-base-content">
	<section
		class="card relative min-h-64 w-full max-w-md rounded-lg border border-base-300 bg-base-100 shadow-xl"
	>
		<div class="card-body justify-center gap-3 px-6 py-12 pb-16 sm:px-8">
			<h1 class="text-3xl font-semibold tracking-normal">
				{messages.forbidden_page_title}
			</h1>
			<p class="max-w-sm text-base leading-relaxed text-base-content/70">
				{messages.forbidden_page_description}
			</p>

			<p class="absolute bottom-5 right-6 font-mono text-xs text-base-content/45">
				error code: 403
			</p>
		</div>
	</section>
</main>
