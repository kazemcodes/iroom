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
			if ($auth.isLoggedIn) {
				goto('/dashboard');
			}
		});
		return unsub;
	});

	async function handleSubmit() {
		loading = true;
		error = '';

		const endpoint = isRegister ? '/auth/register' : '/auth/login';
		const body: any = { email, password };
		if (isRegister) {
			body.display_name = displayName;
			body.phone = phone;
		}

		const res = await api.post<{ user: User; tokens: Tokens }>(endpoint, body);

		if (!res.success) {
			error = res.error || 'خطایی رخ داد';
			loading = false;
			return;
		}

		auth.login(res.data!.user, res.data!.tokens);
		goto('/dashboard');
	}
</script>

<div class="min-h-screen flex items-center justify-center bg-gradient-to-br from-blue-50 via-white to-indigo-50 px-4">
	<div class="w-full max-w-md">
		<!-- Logo -->
		<div class="text-center mb-8">
			<div class="w-16 h-16 bg-blue-600 rounded-2xl flex items-center justify-center text-white font-bold text-2xl mx-auto mb-4 shadow-lg shadow-blue-200">آ</div>
			<h1 class="text-3xl font-bold text-gray-900">آی‌روم</h1>
			<p class="text-gray-500 mt-2">کلاس آنلاین هوشمند</p>
		</div>

		<!-- Card -->
		<div class="bg-white rounded-2xl shadow-xl p-8">
			<!-- Tabs -->
			<div class="flex mb-6 bg-gray-100 rounded-lg p-1">
				<button
					class="flex-1 py-2.5 rounded-md text-sm font-medium transition-all"
					class:bg-white={!isRegister}
					class:text-blue-600={!isRegister}
					class:shadow-sm={!isRegister}
					class:text-gray-500={isRegister}
					onclick={() => { isRegister = false; error = ''; }}
				>
					ورود
				</button>
				<button
					class="flex-1 py-2.5 rounded-md text-sm font-medium transition-all"
					class:bg-white={isRegister}
					class:text-blue-600={isRegister}
					class:shadow-sm={isRegister}
					class:text-gray-500={!isRegister}
					onclick={() => { isRegister = true; error = ''; }}
				>
					ثبت‌نام
				</button>
			</div>

			{#if error}
				<div class="mb-4 p-3 bg-red-50 text-red-600 rounded-lg text-sm flex items-center gap-2">
					<svg class="w-4 h-4 shrink-0" fill="currentColor" viewBox="0 0 20 20">
						<path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clip-rule="evenodd" />
					</svg>
					{error}
				</div>
			{/if}

			<form onsubmit={(e) => { e.preventDefault(); handleSubmit(); }}>
				{#if isRegister}
					<div class="mb-4">
						<label class="block text-sm font-medium text-gray-700 mb-1.5">نام نمایشی</label>
						<input
							type="text"
							bind:value={displayName}
							class="w-full px-4 py-3 border border-gray-200 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none transition-all bg-gray-50 focus:bg-white"
							placeholder="نام خود را وارد کنید"
							required
						/>
					</div>
					<div class="mb-4">
						<label class="block text-sm font-medium text-gray-700 mb-1.5">شماره تلفن</label>
						<input
							type="tel"
							bind:value={phone}
							class="w-full px-4 py-3 border border-gray-200 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none transition-all bg-gray-50 focus:bg-white"
							placeholder="09120000000"
							dir="ltr"
						/>
					</div>
				{/if}

				<div class="mb-4">
					<label class="block text-sm font-medium text-gray-700 mb-1.5">ایمیل</label>
					<input
						type="email"
						bind:value={email}
						class="w-full px-4 py-3 border border-gray-200 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none transition-all bg-gray-50 focus:bg-white"
						placeholder="example@email.com"
						dir="ltr"
						required
					/>
				</div>

				<div class="mb-6">
					<label class="block text-sm font-medium text-gray-700 mb-1.5">رمز عبور</label>
					<input
						type="password"
						bind:value={password}
						class="w-full px-4 py-3 border border-gray-200 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none transition-all bg-gray-50 focus:bg-white"
						placeholder="حداقل ۶ کاراکتر"
						dir="ltr"
						required
						minlength="6"
					/>
				</div>

				<button
					type="submit"
					disabled={loading}
					class="w-full py-3 bg-blue-600 text-white rounded-lg font-medium hover:bg-blue-700 focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 transition-all disabled:opacity-50 disabled:cursor-not-allowed"
				>
					{#if loading}
						<span class="inline-flex items-center gap-2">
							<svg class="animate-spin h-4 w-4" viewBox="0 0 24 24">
								<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" fill="none" />
								<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
							</svg>
							لطفاً صبر کنید...
						</span>
					{:else}
						{isRegister ? 'ثبت‌نام' : 'ورود'}
					{/if}
				</button>
			</form>
		</div>

		<p class="text-center text-xs text-gray-400 mt-6">نسخه ۰.۱.۰ — متن‌باز و رایگان</p>
	</div>
</div>
