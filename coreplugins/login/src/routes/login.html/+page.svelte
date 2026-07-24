<script lang="ts">
	import ThemeToggle from '$lib/ThemeToggle.svelte';

	type LoginResponse = {
		success?: boolean;
		message?: string;
	};

	let username = '';
	let password = '';
	let submitting = false;
	let message = '';
	let messageType: 'error' | 'success' | null = null;

	async function submitLogin(event: SubmitEvent) {
		event.preventDefault();
		submitting = true;
		message = '';
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
				message = 'Login successful! Redirecting...';
				messageType = 'success';
				window.setTimeout(() => window.location.assign('/'), 1000);
			} else {
				message = data.message || 'Login failed';
				messageType = 'error';
			}
		} catch {
			message = 'Network error. Please try again.';
			messageType = 'error';
		} finally {
			submitting = false;
		}
	}
</script>

<svelte:head>
	<title>Login</title>
	<meta name="description" content="Sign in to Arupa" />
</svelte:head>

<main class="grid min-h-screen place-items-center bg-base-200 px-4 py-8 text-base-content">
	<section class="card relative w-full max-w-md rounded-lg border border-base-300 bg-base-100 shadow-xl">
		<div class="absolute right-3 top-3">
			<ThemeToggle />
		</div>

		<div class="card-body gap-6 p-6 pt-16 sm:p-8 sm:pt-16">
			<header class="text-center">
				<h1 class="text-3xl font-semibold tracking-normal">Arupa</h1>
				<p class="mt-2 text-sm text-base-content/60">Please sign in to continue</p>
			</header>

			<form class="grid gap-4" onsubmit={submitLogin}>
				<label class="grid gap-2" for="username">
					<span class="text-sm font-medium text-base-content/80">Username</span>
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
					<span class="text-sm font-medium text-base-content/80">Password</span>
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
						Signing in...
					{:else}
						Sign In
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
