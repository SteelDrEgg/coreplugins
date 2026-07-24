<script lang="ts">
	import { base } from '$app/paths';
	import { onMount } from 'svelte';
	import ThemeToggle from '$lib/ThemeToggle.svelte';

	type ApiResponse = {
		success?: boolean;
	};

	let submitting = false;
	let checkingAuth = true;
	let message = '';
	let hasError = false;

	onMount(() => {
		const controller = new AbortController();

		async function checkAuthentication() {
			try {
				const response = await fetch('/api/check-auth', { signal: controller.signal });
				const data = (await response.json()) as ApiResponse;

				if (!data.success) {
					window.location.replace(`${base}/login.html`);
					return;
				}
			} catch (error) {
				if (!(error instanceof DOMException && error.name === 'AbortError')) {
					window.location.replace(`${base}/login.html`);
					return;
				}
			}

			checkingAuth = false;
		}

		void checkAuthentication();
		return () => controller.abort();
	});

	async function submitLogout() {
		submitting = true;
		message = '';
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
				message = 'Logged out successfully! Redirecting...';
				window.setTimeout(() => window.location.assign(`${base}/login.html`), 1000);
				return;
			}

			message = 'Unable to sign out. Please try again.';
			hasError = true;
		} catch {
			message = 'Network error. Please try again.';
			hasError = true;
		} finally {
			submitting = false;
		}
	}
</script>

<svelte:head>
	<title>Logout</title>
	<meta name="description" content="Sign out of Arupa" />
</svelte:head>

<main class="grid min-h-screen place-items-center bg-base-200 px-4 py-8 text-base-content">
	<section class="card relative w-full max-w-md rounded-lg border border-base-300 bg-base-100 shadow-xl">
		<div class="absolute right-3 top-3">
			<ThemeToggle />
		</div>

		<div class="card-body gap-6 p-6 pt-16 text-center sm:p-8 sm:pt-16">
			<header>
				<h1 class="text-3xl font-semibold tracking-normal">Arupa</h1>
				<p class="mt-2 text-base text-base-content/70">Are you sure you want to sign out?</p>
			</header>

			<div class="grid gap-3">
				<button
					type="button"
					class="btn btn-error w-full"
					disabled={submitting || checkingAuth}
					onclick={submitLogout}
				>
					{#if submitting || checkingAuth}
						<span class="loading loading-spinner loading-sm" aria-hidden="true"></span>
						{submitting ? 'Signing out...' : 'Checking session...'}
					{:else}
						Sign Out
					{/if}
				</button>

				<a href={`${base}/login.html`} class="btn btn-ghost w-full">Back to Login</a>
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
