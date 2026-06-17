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

<div class="min-h-screen flex flex-col items-center justify-center px-4" style="background: linear-gradient(135deg, #0b1120 0%, #1a1a2e 50%, #0d1b2a 100%);">
	<!-- Login Box -->
	<div class="w-full max-w-[400px] rounded-2xl p-6" style="background: white; box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);">
		<!-- Logo -->
		<div class="text-center mb-6">
			<div class="w-14 h-14 rounded-xl mx-auto mb-3 flex items-center justify-center text-white font-bold text-xl" style="background: linear-gradient(135deg, #23b9d7, #004ff2);">
				آ
			</div>
			<h1 class="text-xl font-bold" style="color: #1c293a;">آی‌روم</h1>
		</div>

		<!-- Error message -->
		{#if error}
			<div class="mb-4 px-4 py-3 rounded-lg text-sm text-center" style="background: rgba(224, 82, 82, 0.08); color: #e05252; border: 1px solid rgba(224, 82, 82, 0.2);">
				{error}
			</div>
		{/if}

		<!-- Form -->
		<form onsubmit={(e) => { e.preventDefault(); handleSubmit(); }} class="space-y-3">
			<div>
				<input type="email" bind:value={email} placeholder="ایمیل" dir="ltr" autocomplete="username" required
					class="w-full px-4 py-2.5 rounded-lg text-sm outline-none transition-colors"
					style="border: 1px solid #e0e4eb; color: #1c293a; font-family: 'Vazirmatn', system-ui, sans-serif;"
					onfocus={(e) => e.currentTarget.style.borderColor = '#23b9d7'}
					onblur={(e) => e.currentTarget.style.borderColor = '#e0e4eb'}
				/>
			</div>
			<div>
				<input type="password" bind:value={password} placeholder="گذرواژه" dir="ltr" autocomplete="off" required minlength="6"
					class="w-full px-4 py-2.5 rounded-lg text-sm outline-none transition-colors"
					style="border: 1px solid #e0e4eb; color: #1c293a; font-family: 'Vazirmatn', system-ui, sans-serif;"
					onfocus={(e) => e.currentTarget.style.borderColor = '#23b9d7'}
					onblur={(e) => e.currentTarget.style.borderColor = '#e0e4eb'}
				/>
			</div>

			{#if isRegister}
				<div>
					<input type="text" bind:value={displayName} placeholder="نام نمایشی" dir="auto"
						class="w-full px-4 py-2.5 rounded-lg text-sm outline-none transition-colors"
						style="border: 1px solid #e0e4eb; color: #1c293a;"
						onfocus={(e) => e.currentTarget.style.borderColor = '#23b9d7'}
						onblur={(e) => e.currentTarget.style.borderColor = '#e0e4eb'}
					/>
				</div>
				<div>
					<input type="tel" bind:value={phone} placeholder="شماره تلفن" dir="ltr"
						class="w-full px-4 py-2.5 rounded-lg text-sm outline-none transition-colors"
						style="border: 1px solid #e0e4eb; color: #1c293a;"
						onfocus={(e) => e.currentTarget.style.borderColor = '#23b9d7'}
						onblur={(e) => e.currentTarget.style.borderColor = '#e0e4eb'}
					/>
				</div>
			{/if}

			<button type="submit" disabled={loading}
				class="w-full py-2.5 rounded-lg text-sm font-semibold text-white transition-colors"
				style="background: #23b9d7; font-family: 'Vazirmatn', system-ui, sans-serif; {loading ? 'opacity: 0.6; cursor: not-allowed;' : ''}"
			>
				{loading ? 'در حال پردازش...' : (isRegister ? 'ثبت‌نام' : 'ورود')}
			</button>
		</form>

		<!-- Toggle register/login -->
		<div class="mt-4 text-center">
			<button onclick={() => { isRegister = !isRegister; error = ''; }} class="text-sm hover:underline" style="color: #23b9d7;">
				{isRegister ? 'حساب کاربری دارید؟ ورود' : 'حساب کاربری ندارید؟ ثبت‌نام'}
			</button>
		</div>

		<!-- Forgot password -->
		<div class="mt-3 text-center">
			<a href="/auth/forgot-password" class="text-xs hover:underline" style="color: #6790a0;">رمز عبور را فراموش کرده‌اید؟</a>
		</div>
	</div>

	<!-- Footer -->
	<div class="mt-6 text-center">
		<a href="/" class="text-xs" style="color: rgba(255, 255, 255, 0.4);">© آی‌روم</a>
	</div>
</div>
