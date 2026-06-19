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

<div class="space-y-5">
	<div>
		<h1 class="sky-page-title">مدیریت فایل‌ها</h1>
		<p class="sky-page-subtitle">مشاهده و مدیریت فایل‌های آپلود شده</p>
	</div>

	<div class="sky-search max-w-sm">
		<div class="sky-search-icon"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg></div>
		<input bind:value={searchQuery} class="sky-input" placeholder="جستجو..." style="padding-right: 2.5rem;" />
	</div>

	{#if loading}
		<div class="flex items-center justify-center py-16"><svg class="sky-spinner lg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="color: var(--color-crystal-clear);"><path d="M12 2v4M12 18v4M4.93 4.93l2.83 2.83M16.24 16.24l2.83 2.83M2 12h4M18 12h4M4.93 19.07l2.83-2.83M16.24 7.76l2.83-2.83"/></svg></div>
	{:else}
		<div class="sky-card"><div class="sky-empty">
			<div class="sky-empty-icon"><svg width="48" height="48" fill="none" stroke="currentColor" stroke-width="1" viewBox="0 0 24 24" style="color: var(--color-muted-mountain);"><path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z"/><polyline points="14 2 14 8 20 8"/></svg></div>
			<p class="sky-empty-title">فایل‌ها از طریق جلسات آپلود می‌شوند</p>
			<p class="sky-empty-desc">برای مدیریت فایل‌ها به صفحه جلسات مراجعه کنید</p>
		</div></div>
	{/if}
</div>
