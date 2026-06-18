<script lang="ts">
	import { api } from '$lib/api';
	import { onMount } from 'svelte';
	import type { ActivityLog } from '$lib/types';
	import { toPersianNum } from '$lib/utils/persian';
	import { toPersianDateTime } from '$lib/utils/persian';

	let logs = $state<ActivityLog[]>([]);
	let total = $state(0);
	let page = $state(1);
	let loading = $state(true);

	const perPage = 30;

	onMount(() => loadLogs());

	async function loadLogs() {
		loading = true;
		const res = await api.get<{ items: ActivityLog[]; total: number }>('/admin/logs', { page: String(page), per_page: String(perPage) });
		if (res.success && res.data) {
			logs = res.data.items || [];
			total = res.data.total;
		}
		loading = false;
	}

	async function loadAllLogs(): Promise<ActivityLog[]> {
		const allLogs: ActivityLog[] = [];
		let currentPage = 1;
		const totalPages = Math.ceil(total / perPage);
		while (currentPage <= totalPages) {
			const res = await api.get<{ items: ActivityLog[]; total: number }>('/admin/logs', { page: String(currentPage), per_page: String(perPage) });
			if (res.success && res.data) {
				allLogs.push(...(res.data.items || []));
			}
			currentPage++;
		}
		return allLogs;
	}

	const actionLabels: Record<string, string> = {
		login: 'ورود',
		register: 'ثبت‌نام',
		create_class: 'ایجاد کلاس',
		delete_class: 'حذف کلاس',
		create_session: 'ایجاد جلسه',
		start_session: 'شروع جلسه',
		end_session: 'پایان جلسه',
		upload_file: 'آپلود فایل',
		upload_recording: 'آپلود ضبط',
		update_settings: 'بروزرسانی تنظیمات',
	};

	const actionColors: Record<string, string> = {
		login: 'sky-badge-success',
		register: 'sky-badge-info',
		create_class: 'badge-purple',
		delete_class: 'sky-badge-danger',
		create_session: 'sky-badge-warning',
		start_session: 'sky-badge-success',
		end_session: 'sky-badge-default',
		upload_file: 'sky-badge-info',
		upload_recording: 'badge-purple',
		update_settings: 'sky-badge-warning',
	};

	async function exportCSV() {
		const allLogs = await loadAllLogs();
		const headers = ['زمان', 'کاربر', 'عملیات', 'جزئیات'];
		const rows = allLogs.map(l => [toPersianDateTime(l.created_at), l.user_id || '-', actionLabels[l.action] || l.action, l.details || '-']);
		const csv = [headers.join(','), ...rows.map(r => r.map(v => `"${String(v).replace(/"/g, '""')}"`).join(','))].join('\n');
		const blob = new Blob(['\uFEFF' + csv], { type: 'text/csv;charset=utf-8;' });
		const url = URL.createObjectURL(blob);
		const a = document.createElement('a');
		a.href = url; a.download = 'activity-logs.csv'; a.click();
		URL.revokeObjectURL(url);
	}
</script>

<div class="space-y-5">
	<div class="flex items-center justify-between">
		<div>
			<h1 class="sky-page-title">لاگ فعالیت‌ها</h1>
			<p class="sky-page-subtitle">{toPersianNum(total)} رویداد</p>
		</div>
		{#if !loading && logs.length > 0}
			<button onclick={exportCSV} class="sky-btn sky-btn-primary flex items-center gap-2" style="background: var(--color-lush-meadow);">
				<svg width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"/></svg>
				دانلود CSV
			</button>
		{/if}
	</div>

	{#if loading}
		<div class="flex items-center justify-center py-16"><svg class="sky-spinner lg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="color: var(--color-crystal-clear);"><path d="M12 2v4M12 18v4M4.93 4.93l2.83 2.83M16.24 16.24l2.83 2.83M2 12h4M18 12h4M4.93 19.07l2.83-2.83M16.24 7.76l2.83-2.83"/></svg></div>
	{:else if logs.length === 0}
		<div class="sky-card"><div class="sky-empty"><p class="sky-empty-desc">لاگی ثبت نشده</p></div></div>
	{:else}
		<div class="sky-card overflow-hidden">
			<table class="sky-table">
				<thead><tr><th>عملیات</th><th>نوع</th><th>شناسه</th><th>جزئیات</th><th>تاریخ</th></tr></thead>
				<tbody>
					{#each logs as log}
						<tr>
							<td><span class="sky-badge {actionColors[log.action] || 'sky-badge-default'}">{actionLabels[log.action] || log.action}</span></td>
							<td style="color: var(--color-mystic-sea);">{log.entity_type}</td>
							<td style="color: var(--color-mystic-sea);">#{toPersianNum(log.entity_id)}</td>
							<td style="color: var(--color-mystic-sea); max-width: 200px;" class="truncate">{log.details || '-'}</td>
							<td style="color: var(--color-mystic-sea);">{toPersianDateTime(log.created_at)}</td>
						</tr>
					{/each}
				</tbody>
			</table>
			{#if total > perPage}
				<div class="px-5 py-3 flex items-center justify-between text-sm" style="border-top: 1px solid var(--color-zen-garden); color: var(--color-mystic-sea);">
					<span>{toPersianNum(total)} رویداد</span>
					<div class="sky-pagination">
						<button class="sky-page-btn" disabled={page <= 1} onclick={() => { page--; loadLogs(); }}>قبلی</button>
						<span class="sky-page-btn" style="cursor:default;">{toPersianNum(page)}/{toPersianNum(Math.ceil(total / perPage))}</span>
						<button class="sky-page-btn" disabled={page >= Math.ceil(total / perPage)} onclick={() => { page++; loadLogs(); }}>بعدی</button>
					</div>
				</div>
			{/if}
		</div>
	{/if}
</div>
