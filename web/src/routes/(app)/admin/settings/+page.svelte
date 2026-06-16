<script lang="ts">
	import { api } from '$lib/api';
	import { onMount } from 'svelte';

	let settings = $state({
		max_users_per_room: 100,
		recording_enabled: true,
		maintenance_mode: false,
		allow_student_video: false,
		max_file_size_mb: 50,
		session_auto_end_minutes: 120,
	});
	let loading = $state(true);
	let saving = $state(false);
	let saved = $state(false);

	onMount(async () => {
		const res = await api.get<any>('/admin/settings');
		if (res.success && res.data) settings = { ...settings, ...res.data };
		loading = false;
	});

	async function saveSettings() {
		saving = true;
		saved = false;
		const res = await api.put('/admin/settings', settings);
		if (res.success) {
			saved = true;
			setTimeout(() => saved = false, 3000);
		}
		saving = false;
	}
</script>

<div class="max-w-2xl mx-auto space-y-6">
	<div>
		<h1 class="text-2xl font-bold text-gray-900">تنظیمات سیستم</h1>
		<p class="text-gray-500 mt-1">تنظیمات کلی پلتفرم کلاس آنلاین</p>
	</div>

	{#if loading}
		<div class="flex items-center justify-center py-12">
			<div class="animate-spin h-8 w-8 border-4 border-blue-600 border-t-transparent rounded-full"></div>
		</div>
	{:else}
		<div class="bg-white rounded-xl border divide-y">
			<!-- Max users per room -->
			<div class="px-6 py-4 flex items-center justify-between">
				<div>
					<p class="font-medium text-gray-900">حداکثر کاربر در اتاق</p>
					<p class="text-sm text-gray-500 mt-0.5">تعداد maksimum شرکت‌کنندگان در هر جلسه</p>
				</div>
				<input type="number" bind:value={settings.max_users_per_room} min="2" max="500" class="w-20 px-3 py-2 border rounded-lg text-sm text-center focus:ring-2 focus:ring-blue-500 outline-none" />
			</div>

			<!-- Recording enabled -->
			<div class="px-6 py-4 flex items-center justify-between">
				<div>
					<p class="font-medium text-gray-900">ضبط جلسات</p>
					<p class="text-sm text-gray-500 mt-0.5">امکان ضبط جلسات توسط مدرس</p>
				</div>
				<button
					onclick={() => settings.recording_enabled = !settings.recording_enabled}
					class="relative w-11 h-6 rounded-full transition-colors {settings.recording_enabled ? 'bg-blue-600' : 'bg-gray-300'}"
				>
					<span class="absolute top-0.5 right-0.5 w-5 h-5 bg-white rounded-full transition-transform {settings.recording_enabled ? 'translate-x-[-20px]' : ''}"></span>
				</button>
			</div>

			<!-- Allow student video -->
			<div class="px-6 py-4 flex items-center justify-between">
				<div>
					<p class="font-medium text-gray-900">ارسال ویدیو توسط دانش‌آموز</p>
					<p class="text-sm text-gray-500 mt-0.5">اجازه ارسال ویدیو به دانش‌آموزان</p>
				</div>
				<button
					onclick={() => settings.allow_student_video = !settings.allow_student_video}
					class="relative w-11 h-6 rounded-full transition-colors {settings.allow_student_video ? 'bg-blue-600' : 'bg-gray-300'}"
				>
					<span class="absolute top-0.5 right-0.5 w-5 h-5 bg-white rounded-full transition-transform {settings.allow_student_video ? 'translate-x-[-20px]' : ''}"></span>
				</button>
			</div>

			<!-- Max file size -->
			<div class="px-6 py-4 flex items-center justify-between">
				<div>
					<p class="font-medium text-gray-900">حداکثر حجم فایل (MB)</p>
					<p class="text-sm text-gray-500 mt-0.5">حداکثر اندازه آپلود فایل</p>
				</div>
				<input type="number" bind:value={settings.max_file_size_mb} min="1" max="500" class="w-20 px-3 py-2 border rounded-lg text-sm text-center focus:ring-2 focus:ring-blue-500 outline-none" />
			</div>

			<!-- Auto end session -->
			<div class="px-6 py-4 flex items-center justify-between">
				<div>
					<p class="font-medium text-gray-900">پایان خودکار جلسه (دقیقه)</p>
					<p class="text-sm text-gray-500 mt-0.5">زمان پایان خودکار پس از شروع</p>
				</div>
				<input type="number" bind:value={settings.session_auto_end_minutes} min="30" max="480" class="w-20 px-3 py-2 border rounded-lg text-sm text-center focus:ring-2 focus:ring-blue-500 outline-none" />
			</div>

			<!-- Maintenance mode -->
			<div class="px-6 py-4 flex items-center justify-between">
				<div>
					<p class="font-medium text-gray-900">حالت تعمیر و نگهداری</p>
					<p class="text-sm text-gray-500 mt-0.5">غیرفعال‌سازی موقت سیستم</p>
				</div>
				<button
					onclick={() => settings.maintenance_mode = !settings.maintenance_mode}
					class="relative w-11 h-6 rounded-full transition-colors {settings.maintenance_mode ? 'bg-red-600' : 'bg-gray-300'}"
				>
					<span class="absolute top-0.5 right-0.5 w-5 h-5 bg-white rounded-full transition-transform {settings.maintenance_mode ? 'translate-x-[-20px]' : ''}"></span>
				</button>
			</div>
		</div>

		<div class="flex items-center justify-between">
			{#if saved}
				<span class="text-sm text-green-600">ذخیره شد</span>
			{:else}
				<span></span>
			{/if}
			<button onclick={saveSettings} disabled={saving} class="px-6 py-2.5 bg-blue-600 text-white rounded-lg font-medium text-sm hover:bg-blue-700 transition-colors disabled:opacity-50">
				{saving ? 'در حال ذخیره...' : 'ذخیره تنظیمات'}
			</button>
		</div>
	{/if}
</div>
