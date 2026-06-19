<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib/api';
	import { toPersianNum } from '$lib/utils/persian';

	let sessions = $state<any[]>([]);
	let loading = $state(true);
	let searchQuery = $state('');
	let filterStatus = $state('all');

	onMount(loadSessions);

	async function loadSessions() {
		loading = true;
		const res = await api.get('/admin/sessions');
		if (res.success && res.data) sessions = Array.isArray(res.data) ? res.data : [];
		loading = false;
	}

	async function deleteSession(id: number) {
		if (!confirm('آیا از حذف این جلسه اطمینان دارید؟')) return;
		await api.delete(`/admin/sessions/${id}`);
		await loadSessions();
	}

	const statusLabels: Record<string, string> = { scheduled: 'برنامه‌ریزی شده', live: 'در حال برگزاری', ended: 'پایان یافته' };
	const statusBadge: Record<string, string> = { scheduled: 'sky-badge sky-badge-info', live: 'sky-badge sky-badge-success', ended: 'sky-badge sky-badge-default' };

	const filteredSessions = $derived(sessions.filter(s => {
		if (filterStatus !== 'all' && s.status !== filterStatus) return false;
		if (searchQuery && !s.title?.includes(searchQuery)) return false;
		return true;
	}));
</script>

<div class="space-y-5">
	<div>
		<h1 class="sky-page-title">مدیریت جلسات</h1>
		<p class="sky-page-subtitle">مشاهده و مدیریت تمام جلسات</p>
	</div>

	<div class="flex items-center gap-3">
		<div class="sky-search flex-1">
			<div class="sky-search-icon"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg></div>
			<input bind:value={searchQuery} class="sky-input" style="padding-right: 2.5rem;" placeholder="جستجو در عنوان جلسه..." />
		</div>
		<select bind:value={filterStatus} class="sky-input" style="width:auto;min-width:140px;">
			<option value="all">همه</option>
			<option value="scheduled">برنامه‌ریزی شده</option>
			<option value="live">در حال برگزاری</option>
			<option value="ended">پایان یافته</option>
		</select>
	</div>

	{#if loading}
		<div class="flex items-center justify-center py-16"><svg class="sky-spinner lg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="color: var(--color-crystal-clear);"><path d="M12 2v4M12 18v4M4.93 4.93l2.83 2.83M16.24 16.24l2.83 2.83M2 12h4M18 12h4M4.93 19.07l2.83-2.83M16.24 7.76l2.83-2.83"/></svg></div>
	{:else if filteredSessions.length === 0}
		<div class="sky-card"><div class="sky-empty"><p class="sky-empty-desc">جلسه‌ای یافت نشد</p></div></div>
	{:else}
		<div class="sky-card overflow-hidden">
			<table class="sky-table">
				<thead><tr><th>عنوان</th><th>وضعیت</th><th>تاریخ</th><th>مدت</th><th>عملیات</th></tr></thead>
				<tbody>
					{#each filteredSessions as s (s.id)}
						<tr>
							<td class="font-semibold">{s.title}</td>
							<td><span class="{statusBadge[s.status] || 'sky-badge sky-badge-default'}">{statusLabels[s.status] || s.status}</span></td>
							<td style="color: var(--color-moonlit-mist);">{s.scheduled_at ? new Date(s.scheduled_at).toLocaleDateString('fa-IR') : '—'}</td>
							<td style="color: var(--color-mystic-sea);">{s.duration ? `${toPersianNum(s.duration)} دقیقه` : '—'}</td>
							<td>
								<button onclick={() => deleteSession(s.id)} class="sky-btn-icon" style="width:32px;height:32px;" title="حذف">
									<svg width="16" height="16" fill="none" stroke="var(--color-fiery-passion)" stroke-width="1.75" viewBox="0 0 24 24"><polyline points="3 6 5 6 21 6"/><path d="M19 6l-1 14a2 2 0 01-2 2H8a2 2 0 01-2-2L5 6"/><path d="M10 11v6M14 11v6"/></svg>
								</button>
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	{/if}
</div>
