<script lang="ts">
	import { base } from '$app/paths';
	import ThemeToggle from '$lib/ThemeToggle.svelte';
	import { locale, messagesFor } from '$lib/locale';
	import { serverName } from '$lib/serverName';

	type ApiResponse = {
		success?: boolean;
	};

	let submitting = false;
	let loggedOut = false;
	let messageKey: 'success' | 'failed' | 'network' | null = null;
	let hasError = false;
	$: messages = messagesFor($locale);
	$: message =
		messageKey === 'success'
			? messages.logout_success
			: messageKey === 'failed'
				? messages.logout_failed
				: messageKey === 'network'
					? messages.network_error
					: '';

	async function submitLogout() {
		submitting = true;
		messageKey = null;
		hasError = false;

		try {
			const response = await fetch('/api/logout', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json'
				}
			});
			const data = (await response.json()) as ApiResponse;

			if (data.success) {
				messageKey = 'success';
				loggedOut = true;
				return;
			}

			messageKey = 'failed';
			hasError = true;
		} catch {
			messageKey = 'network';
			hasError = true;
		} finally {
			submitting = false;
		}
	}
</script>

<svelte:head>
	<title>{messages.logout_page_title}</title>
	<meta name="description" content={messages.logout_page_description} />
</svelte:head>

<main class="grid min-h-screen place-items-center bg-base-200 px-4 py-8 text-base-content">
	<section class="card relative w-full max-w-md rounded-lg border border-base-300 bg-base-100 shadow-xl">
		<div class="absolute right-3 top-3">
			<ThemeToggle />
		</div>

		<div class="card-body gap-6 p-6 pt-16 text-center sm:p-8 sm:pt-16">
			<header>
				<h1 class="text-3xl font-semibold tracking-normal">{$serverName}</h1>
				<p class="mt-2 text-base text-base-content/70">
					{messages.sign_out_prompt}
				</p>
			</header>

			<div class="grid gap-3">
				{#if loggedOut}
					<a href={`${base}/login.html`} class="btn btn-primary w-full">
						{messages.back_to_login}
					</a>
				{:else}
					<button
						type="button"
						class="btn btn-error w-full"
						disabled={submitting}
						onclick={submitLogout}
					>
						{#if submitting}
							<span class="loading loading-spinner loading-sm" aria-hidden="true"></span>
							{messages.signing_out}
						{:else}
							{messages.sign_out}
						{/if}
					</button>
				{/if}
			</div>

			{#if message}
				<div
					class:alert-error={hasError}
					class:alert-success={!hasError}
					class="alert text-left text-sm"
					role="status"
				>
					{message}
				</div>
			{/if}
		</div>
	</section>
</main>
