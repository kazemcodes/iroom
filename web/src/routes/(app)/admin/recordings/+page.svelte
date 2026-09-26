<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib/api';
	import { toPersianNum } from '$lib/utils/persian';

	let recordings = $state<any[]>([]);
	let loading = $state(true);
	let searchQuery = $state('');

	onMount(loadRecordings);

	async function loadRecordings() {
		loading = true;
		const res = await api.get('/admin/recordings');
		if (res.success && res.data) recordings = Array.isArray(res.data) ? res.data : [];
		loading = false;
	}

	async function deleteRecording(id: number) {
		if (!confirm('آیا از حذف این ضبط اطمینان دارید؟')) return;
		await api.delete(`/admin/recordings/${id}`);
		await loadRecordings();
	}

	function formatDuration(seconds: number): string {
		const m = Math.floor(seconds / 60);
		const s = seconds % 60;
		return `${toPersianNum(m)}:${toPersianNum(s).padStart(2, '0')}`;
	}

	function formatSize(bytes: number): string {
		if (bytes < 1024) return `${bytes} B`;
		if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
		return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
	}

	const filteredRecordings = $derived(recordings.filter(r => !searchQuery || r.session_title?.includes(searchQuery)));
</script>

<div class="space-y-5">
	<div>
		<h1 class="sky-page-title">مدیریت ضبط‌ها</h1>
		<p class="sky-page-subtitle">مشاهده و مدیریت ضبط‌های جلسات</p>
	</div>

	<div class="sky-search max-w-sm">
		<div class="sky-search-icon"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg></div>
		<input bind:value={searchQuery} class="sky-input" placeholder="جستجو در عنوان جلسه..." style="padding-right: 2.5rem;" />
	</div>

	{#if loading}
		<div class="flex items-center justify-center py-16"><svg class="sky-spinner lg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="color: var(--color-crystal-clear);"><path d="M12 2v4M12 18v4M4.93 4.93l2.83 2.83M16.24 16.24l2.83 2.83M2 12h4M18 12h4M4.93 19.07l2.83-2.83M16.24 7.76l2.83-2.83"/></svg></div>
	{:else if filteredRecordings.length === 0}
		<div class="sky-card"><div class="sky-empty"><div class="sky-empty-icon"><svg width="48" height="48" fill="none" stroke="currentColor" stroke-width="1" viewBox="0 0 24 24" style="color: var(--color-muted-mountain);"><circle cx="12" cy="12" r="10"/><circle cx="12" cy="12" r="3"/></svg></div><p class="sky-empty-title">ضبطی یافت نشد</p></div></div>
	{:else}
		<div class="sky-card overflow-hidden">
			<table class="sky-table">
				<thead><tr><th>جلسه</th><th>مدت</th><th>حجم</th><th>تاریخ</th><th>عملیات</th></tr></thead>
				<tbody>
					{#each filteredRecordings as rec (rec.id)}
						<tr>
							<td class="font-semibold">{rec.session_title || '—'}</td>
							<td style="color: var(--color-mystic-sea);">{rec.duration ? formatDuration(rec.duration) : '—'}</td>
							<td style="color: var(--color-mystic-sea);">{rec.file_size ? formatSize(rec.file_size) : '—'}</td>
							<td style="color: var(--color-moonlit-mist);">{rec.created_at ? new Date(rec.created_at).toLocaleDateString('fa-IR') : '—'}</td>
							<td>
								<div class="flex gap-1">
									<a href="/recordings/{rec.id}" class="sky-btn-icon" style="width:32px;height:32px;" title="مشاهده">
										<svg width="16" height="16" fill="none" stroke="currentColor" stroke-width="1.75" viewBox="0 0 24 24"><path d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"/><path d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"/></svg>
									</a>
									<button onclick={() => deleteRecording(rec.id)} class="sky-btn-icon" style="width:32px;height:32px;" title="حذف">
										<svg width="16" height="16" fill="none" stroke="var(--color-fiery-passion)" stroke-width="1.75" viewBox="0 0 24 24"><polyline points="3 6 5 6 21 6"/><path d="M19 6l-1 14a2 2 0 01-2 2H8a2 2 0 01-2-2L5 6"/><path d="M10 11v6M14 11v6"/></svg>
									</button>
								</div>
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	{/if}
</div>
