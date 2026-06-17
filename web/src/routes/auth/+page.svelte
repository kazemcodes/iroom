<script lang="ts">
	import { auth } from '$lib/stores';
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import { api } from '$lib/api';
	import type { User, Tokens } from '$lib/types';

	let isRegister = $state(false);
	let email = $state('');
	let password = $state('');
	let displayName = $state('');
	let phone = $state('');
	let loading = $state(false);
	let error = $state('');

	onMount(() => {
		auth.init();
		const unsub = auth.subscribe(($auth) => {
			if ($auth.isLoggedIn) goto('/dashboard');
		});
		return unsub;
	});

	async function handleSubmit() {
		loading = true;
		error = '';
		const endpoint = isRegister ? '/auth/register' : '/auth/login';
		const body: any = { email, password };
		if (isRegister) { body.display_name = displayName; body.phone = phone; }

		const res = await api.post<{ user: User; tokens: Tokens }>(endpoint, body);
		if (!res.success) { error = res.error || 'خطایی رخ داد'; loading = false; return; }
		auth.login(res.data!.user, res.data!.tokens);
		goto('/dashboard');
	}
</script>

<div class="login-page">
	<div class="login-box">
		<!-- Logo -->
		<div class="login-logo">
			<a href="/">
				<div class="logo-icon">آ</div>
				<span class="logo-text">آی‌روم</span>
			</a>
		</div>

		<!-- Error message -->
		{#if error}
			<div class="login-error">
				{error}
			</div>
		{/if}

		<!-- Login form -->
		<div class="login-form-container">
			<form onsubmit={(e) => { e.preventDefault(); handleSubmit(); }} class="login-form">
				<div class="login-input">
					<input type="email" bind:value={email} placeholder="ایمیل" dir="ltr" autocomplete="username" required />
				</div>
				<div class="login-input">
					<input type="password" bind:value={password} placeholder="گذرواژه" dir="ltr" autocomplete="off" required minlength="6" />
				</div>

				{#if isRegister}
					<div class="login-input">
						<input type="text" bind:value={displayName} placeholder="نام نمایشی" dir="auto" />
					</div>
					<div class="login-input">
						<input type="tel" bind:value={phone} placeholder="شماره تلفن" dir="ltr" />
					</div>
				{/if}

				<div class="login-buttons">
					<button type="submit" class="btn-login blue" disabled={loading}>
						{loading ? 'در حال پردازش...' : (isRegister ? 'ثبت‌نام' : 'ورود')}
					</button>
				</div>
			</form>
		</div>

		<!-- Bottom links -->
		<div class="login-footer">
			<a href="/auth/forgot-password">راهنما</a>
			<span class="dot">·</span>
			<a href="/auth/forgot-password">حریم خصوصی</a>
		</div>
	</div>

	<!-- Trademark -->
	<div class="trademark">
		<a href="/">© آی‌روم</a>
	</div>
</div>

<style>
	.login-page {
		min-height: 100vh;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		background: linear-gradient(135deg, #0b1120 0%, #1a1a2e 50%, #0d1b2a 100%);
		padding: 2rem;
		position: relative;
	}

	.login-box {
		width: 100%;
		max-width: 400px;
		background: white;
		border-radius: 16px;
		padding: 2.5rem 2rem;
		box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
	}

	.login-logo {
		text-align: center;
		margin-bottom: 2rem;
	}
	.login-logo a {
		text-decoration: none;
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 0.75rem;
	}
	.logo-icon {
		width: 56px;
		height: 56px;
		border-radius: 14px;
		background: linear-gradient(135deg, #23b9d7, #004ff2);
		color: white;
		font-size: 1.75rem;
		font-weight: 800;
		display: flex;
		align-items: center;
		justify-content: center;
		box-shadow: 0 4px 16px rgba(35, 185, 215, 0.3);
	}
	.logo-text {
		font-size: 1.5rem;
		font-weight: 800;
		color: #1c293a;
	}

	.login-error {
		background: rgba(224, 82, 82, 0.08);
		color: #e05252;
		border: 1px solid rgba(224, 82, 82, 0.2);
		border-radius: 8px;
		padding: 0.75rem 1rem;
		font-size: 0.875rem;
		margin-bottom: 1.25rem;
		text-align: center;
	}

	.login-form-container {
		margin-bottom: 1.5rem;
	}

	.login-form {
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
	}

	.login-input {
		display: flex;
		align-items: center;
	}
	.login-input input {
		width: 100%;
		padding: 0.75rem 1rem;
		border: 2px solid #e0e4eb;
		border-radius: 8px;
		font-size: 1rem;
		color: #1c293a;
		background: white;
		outline: none;
		transition: border-color 0.15s ease;
		font-family: 'Vazirmatn', system-ui, sans-serif;
	}
	.login-input input:focus {
		border-color: #23b9d7;
	}
	.login-input input::placeholder {
		color: #9fa2b4;
	}

	.login-buttons {
		display: flex;
		gap: 0.75rem;
		margin-top: 0.5rem;
	}

	.btn-login {
		flex: 1;
		padding: 0.75rem 1rem;
		border: none;
		border-radius: 8px;
		font-size: 1rem;
		font-weight: 600;
		cursor: pointer;
		transition: all 0.15s ease;
		font-family: 'Vazirmatn', system-ui, sans-serif;
	}
	.btn-login.blue {
		background: #23b9d7;
		color: white;
	}
	.btn-login.blue:hover {
		background: #1a9ad4;
	}
	.btn-login.blue:disabled {
		opacity: 0.6;
		cursor: not-allowed;
	}

	.login-footer {
		text-align: center;
		font-size: 0.8rem;
		color: #9fa2b4;
		display: flex;
		justify-content: center;
		gap: 0.5rem;
	}
	.login-footer a {
		color: #6790a0;
		text-decoration: none;
		transition: color 0.15s ease;
	}
	.login-footer a:hover {
		color: #23b9d7;
	}
	.dot {
		color: #dadada;
	}

	.trademark {
		position: absolute;
		bottom: 1.5rem;
		text-align: center;
		width: 100%;
	}
	.trademark a {
		color: rgba(255, 255, 255, 0.4);
		text-decoration: none;
		font-size: 0.75rem;
		transition: color 0.15s ease;
	}
	.trademark a:hover {
		color: rgba(255, 255, 255, 0.7);
	}
</style>
