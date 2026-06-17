<script lang="ts">
	import { api } from '$lib/api';
	import { goto } from '$app/navigation';

	let step = $state<'email' | 'reset'>('email');
	let email = $state('');
	let token = $state('');
	let newPassword = $state('');
	let confirmPassword = $state('');
	let loading = $state(false);
	let error = $state('');
	let success = $state('');

	async function handleRequestReset() {
		loading = true;
		error = '';
		success = '';
		const res = await api.post<{ token: string; message: string }>('/auth/forgot-password', { email });
		if (!res.success) { error = res.error || 'خطایی رخ داد'; loading = false; return; }
		token = res.data!.token;
		success = res.data!.message;
		step = 'reset';
		loading = false;
	}

	async function handleResetPassword() {
		error = '';
		success = '';
		if (newPassword.length < 6) { error = 'رمز عبور باید حداقل ۶ کاراکتر باشد'; return; }
		if (newPassword !== confirmPassword) { error = 'رمزهای عبور مطابقت ندارند'; return; }
		loading = true;
		const res = await api.post<{ message: string }>('/auth/reset-password', { token, password: newPassword });
		if (!res.success) { error = res.error || 'خطایی رخ داد'; loading = false; return; }
		success = 'رمز عبور با موفقیت تغییر کرد. در حال انتقال به صفحه ورود...';
		loading = false;
		setTimeout(() => goto('/auth'), 2000);
	}
</script>

<div class="min-h-screen flex items-center justify-center px-4 py-12" style="background: var(--sky-bg-dark);">
	<div class="w-full max-w-md">
		<div class="text-center mb-8">
			<div class="w-16 h-16 rounded-2xl flex items-center justify-center text-white font-extrabold text-2xl mx-auto mb-4"
				style="background: linear-gradient(135deg, #1a56db, #2563eb); box-shadow: 0 4px 20px rgba(26, 86, 219, 0.4);">
				آ
			</div>
			<h1 class="text-3xl font-extrabold" style="color: var(--sky-text-primary);">بازنشانی رمز عبور</h1>
			<p class="mt-2 font-medium" style="color: var(--sky-text-secondary);">آی‌روم — کلاس آنلاین هوشمند</p>
		</div>

		<div class="card p-8">
			{#if error}
				<div class="mb-5 p-3 rounded-xl text-sm font-medium flex items-center gap-2"
					style="background: rgba(233, 69, 96, 0.15); color: var(--sky-accent-red); border: 1px solid rgba(233, 69, 96, 0.2);">
					<svg class="w-4 h-4 shrink-0" fill="currentColor" viewBox="0 0 20 20">
						<path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clip-rule="evenodd" />
					</svg>
					{error}
				</div>
			{/if}

			{#if success && step === 'reset'}
				<div class="mb-5 p-3 rounded-xl text-sm font-medium flex items-center gap-2"
					style="background: rgba(0, 210, 106, 0.15); color: var(--sky-accent-green); border: 1px solid rgba(0, 210, 106, 0.2);">
					<svg class="w-4 h-4 shrink-0" fill="currentColor" viewBox="0 0 20 20">
						<path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
					</svg>
					{success}
				</div>
			{/if}

			{#if step === 'email'}
				<form onsubmit={(e) => { e.preventDefault(); handleRequestReset(); }} class="space-y-4">
					<p class="text-sm font-medium text-center mb-2" style="color: var(--sky-text-secondary);">
						ایمیل خود را وارد کنید تا لینک بازنشانی رمز عبور برایتان ارسال شود.
					</p>
					<div>
						<label class="block text-sm font-semibold mb-1.5" style="color: var(--sky-text-secondary);">ایمیل</label>
						<input type="email" bind:value={email} class="input-field" placeholder="example@email.com" dir="ltr" required />
					</div>
					<button type="submit" disabled={loading}
						class="btn-primary w-full py-3 text-center disabled:opacity-50 disabled:cursor-not-allowed">
						{#if loading}
							<span class="inline-flex items-center gap-2">
								<svg class="animate-spin h-4 w-4" viewBox="0 0 24 24">
									<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" fill="none" />
									<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
								</svg>
								در حال ارسال...
							</span>
						{:else}
							ارسال لینک بازنشانی
						{/if}
					</button>
				</form>
			{:else}
				<form onsubmit={(e) => { e.preventDefault(); handleResetPassword(); }} class="space-y-4">
					<p class="text-sm font-medium text-center mb-2" style="color: var(--sky-text-secondary);">
						رمز عبور جدید خود را وارد کنید.
					</p>
					<div>
						<label class="block text-sm font-semibold mb-1.5" style="color: var(--sky-text-secondary);">رمز عبور جدید</label>
						<input type="password" bind:value={newPassword} class="input-field" placeholder="حداقل ۶ کاراکتر" dir="ltr" required minlength="6" />
					</div>
					<div>
						<label class="block text-sm font-semibold mb-1.5" style="color: var(--sky-text-secondary);">تکرار رمز عبور</label>
						<input type="password" bind:value={confirmPassword} class="input-field" placeholder="رمز عبور را مجدداً وارد کنید" dir="ltr" required minlength="6" />
					</div>
					<button type="submit" disabled={loading}
						class="btn-primary w-full py-3 text-center disabled:opacity-50 disabled:cursor-not-allowed">
						{#if loading}
							<span class="inline-flex items-center gap-2">
								<svg class="animate-spin h-4 w-4" viewBox="0 0 24 24">
									<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" fill="none" />
									<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
								</svg>
								در حال پردازش...
							</span>
						{:else}
							تغییر رمز عبور
						{/if}
					</button>
				</form>
			{/if}

			<div class="mt-6 text-center">
				<a href="/auth" class="text-sm font-medium hover:underline" style="color: var(--sky-accent-blue);">
					بازگشت به صفحه ورود
				</a>
			</div>
		</div>

		<p class="text-center text-xs mt-6 font-medium" style="color: var(--sky-text-secondary);">نسخه ۰.۱.۰ — متن‌باز و رایگان</p>
	</div>
</div>
