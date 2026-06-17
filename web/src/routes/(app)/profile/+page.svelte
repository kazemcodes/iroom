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

	const roleLabels: Record<string, string> = { admin: 'مدیر سیستم', teacher: 'مدرس', student: 'دانش‌آموز' };

	onMount(async () => {
		avatarData = localStorage.getItem('user_avatar') || '';
		const res = await api.get<User>('/auth/me');
		if (res.success && res.data) {
			user = res.data;
			displayName = user!.display_name;
			phone = user!.phone;
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
	{/if}
</div>
