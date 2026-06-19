<!--
  Auth Page — Admin login only.
  Route: /auth
-->
<script lang="ts">
	import { auth } from '$lib/stores';
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import { api } from '$lib/api';
	import type { User, Tokens } from '$lib/types';

	let email = $state('');
	let password = $state('');
	let loading = $state(false);
	let error = $state('');

	onMount(() => {
		auth.init();
		const unsub = auth.subscribe(($auth) => {
			if ($auth.isLoggedIn) goto('/admin');
		});
		return unsub;
	});

	async function handleLogin() {
		if (!email || !password) { error = 'ایمیل و رمز عبور الزامی است'; return; }
		loading = true; error = '';
		const res = await api.post<{ user: User; tokens: Tokens }>('/auth/login', { email, password });
		if (!res.success) { error = res.error || 'خطایی رخ داد'; loading = false; return; }
		// Only allow admin users
		if (res.data!.user.role !== 'admin') {
			error = 'فقط مدیران اجازه ورود دارند';
			loading = false;
			return;
		}
		auth.login(res.data!.user, res.data!.tokens);
		goto('/admin');
	}
</script>

<div class="min-h-screen flex flex-col items-center justify-center px-4 py-8"
	style="background: linear-gradient(135deg, var(--color-midnight-sky) 0%, var(--color-ocean-wave) 50%, #0d1b2a 100%);">

	<div class="w-full" style="max-width: 420px;">
		<!-- Logo -->
		<div class="text-center mb-8">
			<div class="w-16 h-16 rounded-2xl mx-auto mb-4 flex items-center justify-center shadow-lg" style="background: var(--color-crystal-clear);">
				<svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="white" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
					<path d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14"/>
					<rect x="2" y="6" width="13" height="12" rx="2"/>
				</svg>
			</div>
			<h1 class="text-2xl font-extrabold text-white">آی‌روم</h1>
			<p class="text-sm mt-1" style="color: rgba(255,255,255,0.5);">پنل مدیریت</p>
		</div>

		<!-- Card -->
		<div class="rounded-2xl p-7 shadow-2xl" style="background: var(--color-pure);">
			<h2 class="text-lg font-bold mb-4" style="color: var(--color-midnight-sky);">ورود مدیر</h2>

			{#if error}
				<div class="mb-4 px-4 py-3 rounded-xl text-sm text-center"
					style="background: rgba(224,82,82,0.08); color: var(--color-fiery-passion); border: 1px solid rgba(224,82,82,0.2);">
					{error}
				</div>
			{/if}

			<form onsubmit={(e) => { e.preventDefault(); handleLogin(); }} class="space-y-3">
				<div>
					<label class="sky-label">ایمیل</label>
					<input type="email" bind:value={email} placeholder="admin@example.com" dir="ltr" autocomplete="username" required class="sky-input" />
				</div>
				<div>
					<label class="sky-label">گذرواژه</label>
					<input type="password" bind:value={password} placeholder="رمز عبور" dir="ltr" autocomplete="off" required class="sky-input" />
				</div>
				<div class="pt-1">
					<button type="submit" disabled={loading} class="sky-btn sky-btn-primary w-full" style="height: 44px; font-size: 15px;">
						{#if loading}
							<svg class="sky-spinner sm" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
								<path d="M12 2v4M12 18v4M4.93 4.93l2.83 2.83M16.24 16.24l2.83 2.83M2 12h4M18 12h4M4.93 19.07l2.83-2.83M16.24 7.76l2.83-2.83"/>
							</svg>
							در حال پردازش...
						{:else}
							ورود
						{/if}
					</button>
				</div>
			</form>
		</div>

		<p class="text-center mt-6 text-xs" style="color: rgba(255,255,255,0.3);">© آی‌روم — کلاس آنلاین متن‌باز</p>
	</div>
</div>
