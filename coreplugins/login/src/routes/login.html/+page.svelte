<script lang="ts">
	import ThemeToggle from '$lib/ThemeToggle.svelte';
	import { locale, messagesFor } from '$lib/locale';

	type LoginResponse = {
		success?: boolean;
		message?: string;
	};

	let username = '';
	let password = '';
	let submitting = false;
	let messageKey: 'success' | 'failed' | 'network' | null = null;
	let serverMessage = '';
	let messageType: 'error' | 'success' | null = null;
	$: messages = messagesFor($locale);
	$: message =
		serverMessage ||
			(messageKey === 'success'
				? messages.login_success
				: messageKey === 'failed'
					? messages.login_failed
					: messageKey === 'network'
						? messages.network_error
						: '');

	async function submitLogin(event: SubmitEvent) {
		event.preventDefault();
		submitting = true;
		messageKey = null;
		serverMessage = '';
		messageType = null;

		try {
			const response = await fetch('/api/login', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json'
				},
				body: JSON.stringify({ username, password })
			});
			const data = (await response.json()) as LoginResponse;

			if (data.success) {
				messageKey = 'success';
				messageType = 'success';
				window.setTimeout(() => window.location.assign('/'), 1000);
			} else {
				serverMessage = data.message || '';
				messageKey = serverMessage ? null : 'failed';
				messageType = 'error';
			}
		} catch {
			messageKey = 'network';
			messageType = 'error';
		} finally {
			submitting = false;
		}
	}
</script>

<svelte:head>
	<title>{messages.login_page_title}</title>
	<meta name="description" content={messages.login_page_description} />
</svelte:head>

<main class="grid min-h-screen place-items-center bg-base-200 px-4 py-8 text-base-content">
	<section class="card relative w-full max-w-md rounded-lg border border-base-300 bg-base-100 shadow-xl">
		<div class="absolute right-3 top-3">
			<ThemeToggle />
		</div>

		<div class="card-body gap-6 p-6 pt-16 sm:p-8 sm:pt-16">
			<header class="text-center">
				<h1 class="text-3xl font-semibold tracking-normal">Arupa</h1>
				<p class="mt-2 text-sm text-base-content/60">{messages.sign_in_prompt}</p>
			</header>

			<form class="grid gap-4" onsubmit={submitLogin}>
				<label class="grid gap-2" for="username">
					<span class="text-sm font-medium text-base-content/80">
						{messages.username}
					</span>
					<input
						id="username"
						name="username"
						type="text"
						class="input w-full"
						autocomplete="username"
						bind:value={username}
						required
					/>
				</label>

				<label class="grid gap-2" for="password">
					<span class="text-sm font-medium text-base-content/80">
						{messages.password}
					</span>
					<input
						id="password"
						name="password"
						type="password"
						class="input w-full"
						autocomplete="current-password"
						bind:value={password}
						required
					/>
				</label>

				<button type="submit" class="btn btn-primary mt-2 w-full" disabled={submitting}>
					{#if submitting}
						<span class="loading loading-spinner loading-sm" aria-hidden="true"></span>
						{messages.signing_in}
					{:else}
						{messages.sign_in}
					{/if}
				</button>
			</form>

			{#if messageType}
				<div
					class:alert-error={messageType === 'error'}
					class:alert-success={messageType === 'success'}
					class="alert text-sm"
					role="status"
				>
					{message}
				</div>
			{/if}
		</div>
	</section>
</main>
