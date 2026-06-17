<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib/api';
	import { toPersianNum } from '$lib/utils/persian';

	let files = $state<any[]>([]);
	let loading = $state(true);
	let searchQuery = $state('');

	onMount(loadFiles);

	async function loadFiles() {
		loading = true;
		const res = await api.get('/admin/sessions');
		if (res.success && res.data) {
			sessions = Array.isArray(res.data) ? res.data : [];
		}
		loading = false;
	}

	let sessions = $state<any[]>([]);

	async function deleteFile(id: number) {
		if (!confirm('آیا از حذف این فایل اطمینان دارید؟')) return;
		await api.delete(`/files/${id}`);
	}

	function formatSize(bytes: number): string {
		if (bytes < 1024) return `${bytes} B`;
		if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
		return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
	}
</script>

<div class="space-y-6">
	<div>
		<h1 class="text-2xl font-bold text-gray-900">مدیریت فایل‌ها</h1>
		<p class="text-gray-500 mt-1">مشاهده و مدیریت فایل‌های آپلود شده</p>
	</div>

	<div class="relative">
		<svg class="absolute right-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" /></svg>
		<input bind:value={searchQuery} class="w-full pr-10 pl-4 py-2.5 border border-gray-200 rounded-xl text-sm focus:ring-2 focus:ring-blue-500 outline-none bg-white" placeholder="جستجو..." />
	</div>

	{#if loading}
		<div class="flex items-center justify-center py-20">
			<div class="animate-spin h-8 w-8 border-4 border-blue-600 border-t-transparent rounded-full"></div>
		</div>
	{:else}
		<div class="text-center py-20 bg-white rounded-xl border">
			<svg class="w-16 h-16 text-gray-300 mx-auto mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1" d="M19.5 14.25v-2.625a3.375 3.375 0 00-3.375-3.375h-1.5A1.125 1.125 0 0113.5 7.125v-1.5a3.375 3.375 0 00-3.375-3.375H8.25m0 12.75h7.5m-7.5 3H12M10.5 2.25H5.625c-.621 0-1.125.504-1.125 1.125v17.25c0 .621.504 1.125 1.125 1.125h12.75c.621 0 1.125-.504 1.125-1.125V11.25a9 9 0 00-9-9z" /></svg>
			<p class="text-gray-500">فایل‌ها از طریق جلسات آپلود می‌شوند</p>
			<p class="text-sm text-gray-400 mt-1">برای مدیریت فایل‌ها به صفحه جلسات مراجعه کنید</p>
		</div>
	{/if}
</div>
