<!--
  Auth Page — Login/Register for admin panel (teachers/admins only).
  Route: /auth
-->
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
		loading = true; error = '';
		const endpoint = isRegister ? '/auth/register' : '/auth/login';
		const body: any = { email, password };
		if (isRegister) { body.display_name = displayName; body.phone = phone; }
		const res = await api.post<{ user: User; tokens: Tokens }>(endpoint, body);
		if (!res.success) { error = res.error || 'خطایی رخ داد'; loading = false; return; }
		auth.login(res.data!.user, res.data!.tokens);
		goto('/dashboard');
	}
</script>

<div class="min-h-screen flex flex-col items-center justify-center px-4 py-8"
	style="background: linear-gradient(135deg, var(--color-midnight-sky) 0%, var(--color-ocean-wave) 50%, #0d1b2a 100%);">

	<!-- Login Card -->
	<div class="w-full" style="max-width: 420px;">
		<!-- Logo area -->
		<div class="text-center mb-8">
			<div class="w-16 h-16 rounded-2xl mx-auto mb-4 flex items-center justify-center shadow-lg"
				style="background: var(--color-crystal-clear);">
				<svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="white" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
					<path d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14"/>
					<rect x="2" y="6" width="13" height="12" rx="2"/>
				</svg>
			</div>
			<h1 class="text-2xl font-extrabold text-white">آی‌روم</h1>
			<p class="text-sm mt-1" style="color: rgba(255,255,255,0.5);">پلتفرم کلاس آنلاین</p>
		</div>

		<!-- Card -->
		<div class="rounded-2xl p-7 shadow-2xl" style="background: var(--color-pure);">
			<!-- Tab toggle -->
			<div class="flex gap-1 p-1 rounded-xl mb-6" style="background: var(--color-secret-glow);">
				<button
					class="flex-1 py-2 rounded-lg text-sm font-semibold transition-all"
					style="{!isRegister ? 'background: var(--color-pure); color: var(--color-crystal-clear); box-shadow: 0 1px 3px rgba(0,0,0,0.08);' : 'color: var(--color-moonlit-mist);'}"
					onclick={() => { isRegister = false; error = ''; }}
				>ورود</button>
				<button
					class="flex-1 py-2 rounded-lg text-sm font-semibold transition-all"
					style="{isRegister ? 'background: var(--color-pure); color: var(--color-crystal-clear); box-shadow: 0 1px 3px rgba(0,0,0,0.08);' : 'color: var(--color-moonlit-mist);'}"
					onclick={() => { isRegister = true; error = ''; }}
				>ثبت‌نام</button>
			</div>

			<!-- Error -->
			{#if error}
				<div class="mb-4 px-4 py-3 rounded-xl text-sm text-center"
					style="background: rgba(224,82,82,0.08); color: var(--color-fiery-passion); border: 1px solid rgba(224,82,82,0.2);">
					{error}
				</div>
			{/if}

			<!-- Form -->
			<form onsubmit={(e) => { e.preventDefault(); handleSubmit(); }} class="space-y-3">
				{#if isRegister}
					<div>
						<label class="sky-label">نام نمایشی</label>
						<input type="text" bind:value={displayName} placeholder="نام و نام خانوادگی" dir="auto" required
							class="sky-input" />
					</div>
				{/if}
				<div>
					<label class="sky-label">ایمیل</label>
					<input type="email" bind:value={email} placeholder="example@mail.com" dir="ltr" autocomplete="username" required
						class="sky-input" />
				</div>
				<div>
					<label class="sky-label">گذرواژه</label>
					<input type="password" bind:value={password} placeholder="حداقل ۶ کاراکتر" dir="ltr" autocomplete="off" required minlength="6"
						class="sky-input" />
				</div>
				{#if isRegister}
					<div>
						<label class="sky-label">شماره تلفن (اختیاری)</label>
						<input type="tel" bind:value={phone} placeholder="09120000000" dir="ltr"
							class="sky-input" />
					</div>
				{/if}

				<div class="pt-1">
					<button type="submit" disabled={loading} class="sky-btn sky-btn-primary w-full" style="height: 44px; font-size: 15px;">
						{#if loading}
							<svg class="sky-spinner sm" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
								<path d="M12 2v4M12 18v4M4.93 4.93l2.83 2.83M16.24 16.24l2.83 2.83M2 12h4M18 12h4M4.93 19.07l2.83-2.83M16.24 7.76l2.83-2.83"/>
							</svg>
							در حال پردازش...
						{:else}
							{isRegister ? 'ثبت‌نام' : 'ورود به حساب'}
						{/if}
					</button>
				</div>
			</form>

			{#if !isRegister}
				<div class="mt-4 text-center">
					<a href="/auth/forgot-password" class="text-xs hover:underline" style="color: var(--color-mystic-sea);">رمز عبور را فراموش کرده‌اید؟</a>
				</div>
			{/if}
		</div>

		<p class="text-center mt-6 text-xs" style="color: rgba(255,255,255,0.3);">© آی‌روم — کلاس آنلاین متن‌باز</p>
	</div>
</div>
