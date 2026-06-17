<script lang="ts">
	import { auth } from '$lib/stores';
	import { api } from '$lib/api';
	import { onMount } from 'svelte';
	import type { User } from '$lib/types';
	import { toPersianNum, toPersianDate } from '$lib/utils/persian';

	let user = $state<User | null>(null);
	let loading = $state(true);
	let saving = $state(false);
	let changingPassword = $state(false);
	let displayName = $state('');
	let phone = $state('');
	let currentPassword = $state('');
	let newPassword = $state('');
	let confirmNewPassword = $state('');
	let success = $state('');
	let error = $state('');
	let passwordSuccess = $state('');
	let passwordError = $state('');
	let avatarData = $state('');
	let avatarInput = $state<HTMLInputElement | null>(null);

	// 2FA State
	let twoFactorEnabled = $state(false);
	let twoFactorSetupLoading = $state(false);
	let twoFactorVerifying = $state(false);
	let twoFactorDisabling = $state(false);
	let twoFactorRegenerating = $state(false);
	let twoFactorSecret = $state('');
	let twoFactorQrUrl = $state('');
	let totpCode = $state('');
	let twoFactorError = $state('');
	let twoFactorSuccess = $state('');
	let backupCodes = $state<string[]>([]);
	let showBackupCodes = $state(false);
	let disablePassword = $state('');
	let showDisableConfirm = $state(false);

	const roleLabels: Record<string, string> = { admin: 'مدیر سیستم', teacher: 'مدرس', student: 'دانش‌آموز' };

	onMount(async () => {
		avatarData = localStorage.getItem('user_avatar') || '';
		const res = await api.get<User>('/auth/me');
		if (res.success && res.data) {
			user = res.data;
			displayName = user!.display_name;
			phone = user!.phone;
			twoFactorEnabled = (user as any).two_factor_enabled || false;
		}
		loading = false;
	});

	async function handleUpdateProfile() {
		saving = true;
		success = '';
		error = '';
		const res = await api.put<User>('/auth/me', { display_name: displayName, phone });
		if (!res.success) { error = res.error || 'خطایی رخ داد'; saving = false; return; }
		user = res.data!;
		const stored = localStorage.getItem('user');
		if (stored) {
			const parsed = JSON.parse(stored);
			parsed.display_name = res.data!.display_name;
			parsed.phone = res.data!.phone;
			localStorage.setItem('user', JSON.stringify(parsed));
			auth.init();
		}
		success = 'اطلاعات با موفقیت بروزرسانی شد';
		saving = false;
	}

	async function handleChangePassword() {
		passwordSuccess = '';
		passwordError = '';
		if (newPassword.length < 6) { passwordError = 'رمز عبور جدید باید حداقل ۶ کاراکتر باشد'; return; }
		if (newPassword !== confirmNewPassword) { passwordError = 'رمزهای عبور مطابقت ندارند'; return; }
		changingPassword = true;
		const res = await api.post<{ message: string }>('/auth/change-password', { current_password: currentPassword, new_password: newPassword });
		if (!res.success) { passwordError = res.error || 'خطایی رخ داد'; changingPassword = false; return; }
		passwordSuccess = 'رمز عبور با موفقیت تغییر کرد';
		currentPassword = '';
		newPassword = '';
		confirmNewPassword = '';
		changingPassword = false;
	}

	function formatDate(d: string) {
		if (!d) return '';
		return toPersianDate(d);
	}

	function handleAvatarUpload(e: Event) {
		const input = e.target as HTMLInputElement;
		if (!input.files?.length) return;
		const file = input.files[0];
		const reader = new FileReader();
		reader.onload = () => {
			avatarData = reader.result as string;
			localStorage.setItem('user_avatar', avatarData);
		};
		reader.readAsDataURL(file);
		input.value = '';
	}

	// 2FA Functions
	async function handleSetup2FA() {
		twoFactorSetupLoading = true;
		twoFactorError = '';
		twoFactorSecret = '';
		twoFactorQrUrl = '';
		totpCode = '';

		const res = await api.post<{ secret: string; qr_url: string; backup_codes: string[] }>('/auth/2fa/setup');
		if (!res.success) {
			twoFactorError = res.error || 'خطا در شروع تنظیم احراز هویت دو مرحله‌ای';
			twoFactorSetupLoading = false;
			return;
		}

		twoFactorSecret = res.data!.secret;
		twoFactorQrUrl = res.data!.qr_url;
		backupCodes = res.data!.backup_codes || [];
		twoFactorSetupLoading = false;
	}

	async function handleVerify2FA() {
		if (totpCode.length !== 6) {
			twoFactorError = 'کد باید ۶ رقم باشد';
			return;
		}

		twoFactorVerifying = true;
		twoFactorError = '';

		const res = await api.post<{ message: string; backup_codes: string[] }>('/auth/2fa/verify', { code: totpCode });
		if (!res.success) {
			twoFactorError = res.error || 'کد وارد شده صحیح نیست';
			twoFactorVerifying = false;
			return;
		}

		twoFactorSuccess = 'احراز هویت دو مرحله‌ای با موفقیت فعال شد';
		twoFactorEnabled = true;
		twoFactorSecret = '';
		twoFactorQrUrl = '';
		totpCode = '';
		if (res.data!.backup_codes) {
			backupCodes = res.data!.backup_codes;
		}
		showBackupCodes = true;
		twoFactorVerifying = false;
	}

	async function handleDisable2FA() {
		if (!disablePassword) {
			twoFactorError = 'رمز عبور الزامی است';
			return;
		}

		twoFactorDisabling = true;
		twoFactorError = '';

		const res = await api.post<{ message: string }>('/auth/2fa/disable', { password: disablePassword });
		if (!res.success) {
			twoFactorError = res.error || 'خطا در غیرفعال کردن احراز هویت دو مرحله‌ای';
			twoFactorDisabling = false;
			return;
		}

		twoFactorSuccess = 'احراز هویت دو مرحله‌ای غیرفعال شد';
		twoFactorEnabled = false;
		disablePassword = '';
		showDisableConfirm = false;
		backupCodes = [];
		twoFactorDisabling = false;
	}

	async function handleRegenerateBackupCodes() {
		twoFactorRegenerating = true;
		twoFactorError = '';

		const res = await api.post<{ backup_codes: string[] }>('/auth/2fa/backup');
		if (!res.success) {
			twoFactorError = res.error || 'خطا در تولید مجدد کدهای پشتیبان';
			twoFactorRegenerating = false;
			return;
		}

		backupCodes = res.data!.backup_codes;
		showBackupCodes = true;
		twoFactorRegenerating = false;
	}

	function getQrCodeImageUrl(otpauthUrl: string): string {
		const encoded = encodeURIComponent(otpauthUrl);
		return `https://api.qrserver.com/v1/create-qr-code/?size=200x200&data=${encoded}`;
	}

	function copyToClipboard(text: string) {
		navigator.clipboard.writeText(text);
	}
</script>

<div class="space-y-8 max-w-2xl mx-auto">
	<!-- Header -->
	<div>
		<h1 class="text-2xl font-extrabold text-gray-900">حساب کاربری</h1>
		<p class="text-gray-400 mt-1 font-medium">مدیریت اطلاعات شخصی و تنظیمات حساب</p>
	</div>

	{#if loading}
		<div class="flex items-center justify-center py-20">
			<div class="animate-spin h-8 w-8 border-4 border-blue-500 border-t-transparent rounded-full"></div>
		</div>
	{:else if user}
		<!-- Profile Card -->
		<div class="card p-6">
			<div class="flex items-center gap-5 mb-6 pb-6 border-b border-gray-50">
				<div class="relative group">
					<input type="file" accept="image/*" bind:this={avatarInput} onchange={handleAvatarUpload} class="hidden" />
					<button onclick={() => avatarInput?.click()} class="w-20 h-20 rounded-2xl flex items-center justify-center text-white font-extrabold text-3xl shadow-lg shrink-0 overflow-hidden cursor-pointer {avatarData ? '' : ''}"
						style="background: linear-gradient(135deg, #1a56db, #7c3aed);">
						{#if avatarData}
							<img src={avatarData} alt="آواتار" class="w-full h-full object-cover" />
						{:else}
							{user.display_name.charAt(0)}
						{/if}
					</button>
					<div class="absolute inset-0 rounded-2xl bg-black/40 opacity-0 group-hover:opacity-100 flex items-center justify-center transition-opacity cursor-pointer" onclick={() => avatarInput?.click()}>
						<svg class="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M3 9a2 2 0 012-2h.93a2 2 0 001.664-.89l.812-1.22A2 2 0 0110.07 4h3.86a2 2 0 011.664.89l.812 1.22A2 2 0 0018.07 7H19a2 2 0 012 2v9a2 2 0 01-2 2H5a2 2 0 01-2-2V9z" /><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M15 13a3 3 0 11-6 0 3 3 0 016 0z" /></svg>
					</div>
				</div>
				<div>
					<h2 class="text-xl font-extrabold text-gray-900">{user.display_name}</h2>
					<p class="text-sm text-gray-400 font-medium mt-0.5">{user.email}</p>
					<span class="badge {user.role === 'admin' ? 'bg-amber-100 text-amber-700' : user.role === 'teacher' ? 'bg-violet-100 text-violet-700' : 'bg-teal-100 text-teal-700'} mt-2">
						{roleLabels[user.role]}
					</span>
				</div>
			</div>

			<!-- Info -->
			<div class="grid grid-cols-2 gap-4 text-sm mb-6">
				<div>
					<span class="text-gray-400 font-medium">ایمیل</span>
					<p class="text-gray-800 font-semibold mt-1" dir="ltr">{user.email}</p>
				</div>
				<div>
					<span class="text-gray-400 font-medium">تلفن</span>
					<p class="text-gray-800 font-semibold mt-1" dir="ltr">{toPersianNum(user.phone) || '—'}</p>
				</div>
				<div>
					<span class="text-gray-400 font-medium">تاریخ عضویت</span>
					<p class="text-gray-800 font-semibold mt-1">{formatDate(user.created_at)}</p>
				</div>
				<div>
					<span class="text-gray-400 font-medium">وضعیت</span>
					<p class="text-gray-800 font-semibold mt-1">{user.is_active ? 'فعال' : 'غیرفعال'}</p>
				</div>
			</div>
		</div>

		<!-- Edit Profile -->
		<div class="card p-6">
			<h3 class="font-bold text-gray-900 mb-4">بروزرسانی اطلاعات</h3>

			{#if success}
				<div class="mb-4 p-3 bg-green-50 text-green-600 rounded-xl text-sm font-medium border border-green-100">{success}</div>
			{/if}
			{#if error}
				<div class="mb-4 p-3 bg-red-50 text-red-600 rounded-xl text-sm font-medium border border-red-100">{error}</div>
			{/if}

			<form onsubmit={(e) => { e.preventDefault(); handleUpdateProfile(); }} class="space-y-4">
				<div>
					<label class="block text-sm font-semibold text-gray-700 mb-1.5">نام نمایشی</label>
					<input type="text" bind:value={displayName} class="input-field" placeholder="نام نمایشی" required />
				</div>
				<div>
					<label class="block text-sm font-semibold text-gray-700 mb-1.5">شماره تلفن</label>
					<input type="tel" bind:value={phone} class="input-field" placeholder="09120000000" dir="ltr" />
				</div>
				<button type="submit" disabled={saving}
					class="btn-primary disabled:opacity-50 disabled:cursor-not-allowed">
					{#if saving}
						<span class="inline-flex items-center gap-2">
							<svg class="animate-spin h-4 w-4" viewBox="0 0 24 24">
								<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" fill="none" />
								<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
							</svg>
							در حال ذخیره...
						</span>
					{:else}
						ذخیره تغییرات
					{/if}
				</button>
			</form>
		</div>

		<!-- Change Password -->
		<div class="card p-6">
			<h3 class="font-bold text-gray-900 mb-4">تغییر رمز عبور</h3>

			{#if passwordSuccess}
				<div class="mb-4 p-3 bg-green-50 text-green-600 rounded-xl text-sm font-medium border border-green-100">{passwordSuccess}</div>
			{/if}
			{#if passwordError}
				<div class="mb-4 p-3 bg-red-50 text-red-600 rounded-xl text-sm font-medium border border-red-100">{passwordError}</div>
			{/if}

			<form onsubmit={(e) => { e.preventDefault(); handleChangePassword(); }} class="space-y-4">
				<div>
					<label class="block text-sm font-semibold text-gray-700 mb-1.5">رمز عبور فعلی</label>
					<input type="password" bind:value={currentPassword} class="input-field" placeholder="رمز عبور فعلی خود را وارد کنید" dir="ltr" required />
				</div>
				<div>
					<label class="block text-sm font-semibold text-gray-700 mb-1.5">رمز عبور جدید</label>
					<input type="password" bind:value={newPassword} class="input-field" placeholder="حداقل ۶ کاراکتر" dir="ltr" required minlength="6" />
				</div>
				<div>
					<label class="block text-sm font-semibold text-gray-700 mb-1.5">تکرار رمز عبور جدید</label>
					<input type="password" bind:value={confirmNewPassword} class="input-field" placeholder="رمز عبور را مجدداً وارد کنید" dir="ltr" required minlength="6" />
				</div>
				<button type="submit" disabled={changingPassword}
					class="btn-primary disabled:opacity-50 disabled:cursor-not-allowed">
					{#if changingPassword}
						<span class="inline-flex items-center gap-2">
							<svg class="animate-spin h-4 w-4" viewBox="0 0 24 24">
								<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" fill="none" />
								<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
							</svg>
							در حال تغییر...
						</span>
					{:else}
						تغییر رمز عبور
					{/if}
				</button>
			</form>
		</div>

		<!-- Two-Factor Authentication -->
		<div class="card p-6">
			<div class="flex items-center justify-between mb-4">
				<h3 class="font-bold text-gray-900">احراز هویت دو مرحله‌ای</h3>
				{#if twoFactorEnabled}
					<span class="badge bg-green-100 text-green-700 flex items-center gap-1.5">
						<svg class="w-3.5 h-3.5" fill="currentColor" viewBox="0 0 20 20">
							<path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd"/>
						</svg>
						فعال
					</span>
				{:else}
					<span class="badge bg-gray-100 text-gray-500">غیرفعال</span>
				{/if}
			</div>

			<p class="text-sm text-gray-500 mb-5">
				احراز هویت دو مرحله‌ای یک لایه امنیتی اضافی به حساب شما اضافه می‌کند. هنگام ورود، علاوه بر رمز عبور، به کد یکبار مصرف نیاز خواهید داشت.
			</p>

			{#if twoFactorSuccess}
				<div class="mb-4 p-3 bg-green-50 text-green-600 rounded-xl text-sm font-medium border border-green-100">{twoFactorSuccess}</div>
			{/if}
			{#if twoFactorError}
				<div class="mb-4 p-3 bg-red-50 text-red-600 rounded-xl text-sm font-medium border border-red-100">{twoFactorError}</div>
			{/if}

			{#if !twoFactorEnabled}
				{#if !twoFactorQrUrl}
					<!-- Enable 2FA Button -->
					<button onclick={handleSetup2FA} disabled={twoFactorSetupLoading}
						class="btn-primary disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2">
						{#if twoFactorSetupLoading}
							<svg class="animate-spin h-4 w-4" viewBox="0 0 24 24">
								<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" fill="none"/>
								<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"/>
							</svg>
							در حال آماده‌سازی...
						{:else}
							<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z"/>
							</svg>
							فعال‌سازی احراز هویت دو مرحله‌ای
						{/if}
					</button>
				{:else}
					<!-- QR Code and Verification -->
					<div class="space-y-5">
						<div class="bg-blue-50 border border-blue-100 rounded-xl p-4">
							<h4 class="font-semibold text-blue-800 mb-3 flex items-center gap-2">
								<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v1m6 11h2m-6 0h-2v4m0-11v3m0 0h.01M12 12h4.01M16 20h4M4 12h4m12 0h.01M5 8h2a1 1 0 001-1V5a1 1 0 00-1-1H5a1 1 0 00-1 1v2a1 1 0 001 1zm12 0h2a1 1 0 001-1V5a1 1 0 00-1-1h-2a1 1 0 00-1 1v2a1 1 0 001 1zM5 20h2a1 1 0 001-1v-2a1 1 0 00-1-1H5a1 1 0 00-1 1v2a1 1 0 001 1z"/>
								</svg>
								مرحله ۱: اسکن کد QR
							</h4>
							<div class="flex flex-col sm:flex-row items-center gap-5">
								<div class="bg-white p-3 rounded-xl shadow-sm">
									<img src={getQrCodeImageUrl(twoFactorQrUrl)} alt="QR Code 2FA" class="w-40 h-40" />
								</div>
								<div class="flex-1 text-sm text-blue-700">
									<p class="mb-2">این کد QR را با یک برنامه احراز هویت (مانند Google Authenticator یا Authy) اسکن کنید.</p>
									<p class="font-medium">یا کد زیر را به صورت دستی وارد کنید:</p>
									<div class="mt-2 flex items-center gap-2">
										<code class="bg-white px-3 py-1.5 rounded-lg font-mono text-xs border border-blue-200 select-all" dir="ltr">{twoFactorSecret}</code>
										<button onclick={() => copyToClipboard(twoFactorSecret)}
											class="p-1.5 hover:bg-blue-100 rounded-lg transition-colors" title="کپی">
											<svg class="w-4 h-4 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"/>
											</svg>
										</button>
									</div>
								</div>
							</div>
						</div>

						<div class="bg-amber-50 border border-amber-100 rounded-xl p-4">
							<h4 class="font-semibold text-amber-800 mb-3 flex items-center gap-2">
								<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"/>
								</svg>
								مرحله ۲: ذخیره کدهای پشتیبان
							</h4>
							{#if backupCodes.length > 0}
								<button onclick={() => showBackupCodes = !showBackupCodes}
									class="text-amber-700 text-sm font-medium hover:underline mb-2">
									{showBackupCodes ? 'مخفی کردن کدها' : 'نمایش کدهای پشتیبان'}
								</button>
								{#if showBackupCodes}
									<div class="bg-white rounded-lg p-3 border border-amber-200">
										<p class="text-xs text-amber-600 mb-2">این کدها را در مکانی امن ذخیره کنید. هر کد فقط یک بار قابل استفاده است.</p>
										<div class="grid grid-cols-2 gap-2 font-mono text-sm" dir="ltr">
											{#each backupCodes as code}
												<code class="bg-amber-50 px-2 py-1 rounded text-amber-800">{code}</code>
											{/each}
										</div>
									</div>
								{/if}
							{/if}
						</div>

						<div class="bg-green-50 border border-green-100 rounded-xl p-4">
							<h4 class="font-semibold text-green-800 mb-3 flex items-center gap-2">
								<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"/>
								</svg>
								مرحله ۳: تأیید و فعال‌سازی
							</h4>
							<div class="flex items-center gap-3">
								<input type="text" bind:value={totpCode} maxlength="6"
									class="input-field w-32 text-center text-lg tracking-widest"
									placeholder="000000" dir="ltr"
									oninput={(e) => totpCode = (e.target as HTMLInputElement).value.replace(/\D/g, '')} />
								<button onclick={handleVerify2FA} disabled={twoFactorVerifying || totpCode.length !== 6}
									class="btn-primary disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2">
									{#if twoFactorVerifying}
										<svg class="animate-spin h-4 w-4" viewBox="0 0 24 24">
											<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" fill="none"/>
											<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"/>
										</svg>
										در حال تأیید...
									{:else}
										تأیید و فعال‌سازی
									{/if}
								</button>
							</div>
						</div>
					</div>
				{/if}
			{:else}
				<!-- 2FA Enabled - Show Status and Options -->
				<div class="space-y-4">
					<div class="bg-green-50 border border-green-100 rounded-xl p-4 flex items-start gap-3">
						<svg class="w-5 h-5 text-green-600 mt-0.5 shrink-0" fill="currentColor" viewBox="0 0 20 20">
							<path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd"/>
						</svg>
						<div>
							<p class="font-medium text-green-800">احراز هویت دو مرحله‌ای فعال است</p>
							<p class="text-sm text-green-600 mt-1">حساب شما با امنیت بیشتری محافظت می‌شود.</p>
						</div>
					</div>

					<!-- Backup Codes Section -->
					<div class="border border-gray-200 rounded-xl p-4">
						<h4 class="font-semibold text-gray-800 mb-3 flex items-center gap-2">
							<svg class="w-5 h-5 text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z"/>
							</svg>
							کدهای پشتیبان
						</h4>
						{#if backupCodes.length > 0}
							<button onclick={() => showBackupCodes = !showBackupCodes}
								class="text-blue-600 text-sm font-medium hover:underline mb-2">
								{showBackupCodes ? 'مخفی کردن کدها' : 'نمایش کدهای پشتیبان'}
							</button>
							{#if showBackupCodes}
								<div class="bg-gray-50 rounded-lg p-3 border border-gray-200">
									<p class="text-xs text-gray-500 mb-2">این کدها را در مکانی امن ذخیره کنید. هر کد فقط یک بار قابل استفاده است.</p>
									<div class="grid grid-cols-2 gap-2 font-mono text-sm" dir="ltr">
										{#each backupCodes as code}
											<code class="bg-white px-2 py-1 rounded border border-gray-200">{code}</code>
										{/each}
									</div>
								</div>
							{/if}
						{:else}
							<p class="text-sm text-gray-500 mb-3">کدهای پشتیبان موجود نیست.</p>
						{/if}
						<button onclick={handleRegenerateBackupCodes} disabled={twoFactorRegenerating}
							class="mt-3 text-sm text-blue-600 hover:text-blue-700 font-medium flex items-center gap-1.5 disabled:opacity-50">
							{#if twoFactorRegenerating}
								<svg class="animate-spin h-4 w-4" viewBox="0 0 24 24">
									<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" fill="none"/>
									<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"/>
								</svg>
							{:else}
								<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"/>
								</svg>
							{/if}
							تولید مجدد کدهای پشتیبان
						</button>
					</div>

					<!-- Disable 2FA -->
					<div class="border border-red-200 rounded-xl p-4">
						<h4 class="font-semibold text-red-800 mb-2 flex items-center gap-2">
							<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"/>
							</svg>
							غیرفعال کردن احراز هویت دو مرحله‌ای
						</h4>
						<p class="text-sm text-gray-500 mb-3">غیرفعال کردن احراز هویت دو مرحله‌ای امنیت حساب شما را کاهش می‌دهد.</p>

						{#if !showDisableConfirm}
							<button onclick={() => showDisableConfirm = true}
								class="btn-danger text-sm flex items-center gap-1.5">
								<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M18.364 18.364A9 9 0 005.636 5.636m12.728 12.728A9 9 0 015.636 5.636m12.728 12.728L5.636 5.636"/>
								</svg>
								غیرفعال کردن
							</button>
						{:else}
							<div class="space-y-3">
								<p class="text-sm text-red-600 font-medium">برای غیرفعال کردن، رمز عبور خود را وارد کنید:</p>
								<input type="password" bind:value={disablePassword}
									class="input-field" placeholder="رمز عبور فعلی" dir="ltr" />
								<div class="flex items-center gap-2">
									<button onclick={handleDisable2FA} disabled={twoFactorDisabling || !disablePassword}
										class="btn-danger text-sm disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-1.5">
										{#if twoFactorDisabling}
											<svg class="animate-spin h-4 w-4" viewBox="0 0 24 24">
												<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" fill="none"/>
												<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"/>
											</svg>
											در حال غیرفعال کردن...
										{:else}
											تأیید و غیرفعال کردن
										{/if}
									</button>
									<button onclick={() => { showDisableConfirm = false; disablePassword = ''; }}
										class="btn-ghost text-sm">
										انصراف
									</button>
								</div>
							</div>
						{/if}
					</div>
				</div>
			{/if}
		</div>
	{/if}
</div>
