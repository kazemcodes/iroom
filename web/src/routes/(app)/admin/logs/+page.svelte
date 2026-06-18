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
		login: 'bg-green-100 text-green-700',
		register: 'bg-blue-100 text-blue-700',
		create_class: 'bg-purple-100 text-purple-700',
		delete_class: 'bg-red-100 text-red-700',
		create_session: 'bg-amber-100 text-amber-700',
		start_session: 'bg-green-100 text-green-700',
		end_session: 'bg-gray-100 text-gray-600',
		upload_file: 'bg-cyan-100 text-cyan-700',
		upload_recording: 'bg-pink-100 text-pink-700',
		update_settings: 'bg-orange-100 text-orange-700',
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

<div class="space-y-6">
	<div class="flex items-center justify-between">
		<div>
			<h1 style="font-size:1.5rem;font-weight:700;color:var(--color-midnight-sky);">لاگ فعالیت‌ها</h1>
			<p class="text-gray-500 mt-1">{toPersianNum(total)} رویداد</p>
		</div>
		{#if !loading && logs.length > 0}
			<button onclick={exportCSV} class="px-4 py-2.5 bg-green-600 text-white rounded-lg text-sm font-medium hover:bg-green-700 flex items-center gap-2">
				<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" /></svg>
				دانلود CSV
			</button>
		{/if}
	</div>

	{#if loading}
		<div class="flex items-center justify-center py-12"><div class="animate-spin h-8 w-8 border-4 border-blue-600 border-t-transparent rounded-full"></div></div>
	{:else if logs.length === 0}
		<div class="text-center py-20 bg-white rounded-xl">
			<p class="text-gray-500">لاگی ثبت نشده</p>
		</div>
	{:else}
		<div class="bg-white rounded-xl overflow-hidden">
			<table class="w-full text-sm">
				<thead class="bg-gray-50 border-b">
					<tr>
						<th class="px-5 py-3 text-right font-medium text-gray-600">عملیات</th>
						<th class="px-5 py-3 text-right font-medium text-gray-600">نوع</th>
						<th class="px-5 py-3 text-right font-medium text-gray-600">شناسه</th>
						<th class="px-5 py-3 text-right font-medium text-gray-600">جزئیات</th>
						<th class="px-5 py-3 text-right font-medium text-gray-600">تاریخ</th>
					</tr>
				</thead>
				<tbody class="divide-y">
					{#each logs as log}
						<tr class="hover:bg-gray-50">
							<td class="px-5 py-3">
								<span class="text-xs px-2 py-1 rounded-full font-medium {actionColors[log.action] || 'bg-gray-100 text-gray-600'}">
									{actionLabels[log.action] || log.action}
								</span>
							</td>
							<td class="px-5 py-3 text-gray-500">{log.entity_type}</td>
							<td class="px-5 py-3 text-gray-500">#{toPersianNum(log.entity_id)}</td>
							<td class="px-5 py-3 text-gray-500 max-w-[200px] truncate">{log.details || '-'}</td>
							<td class="px-5 py-3 text-gray-500">{toPersianDateTime(log.created_at)}</td>
						</tr>
					{/each}
				</tbody>
			</table>
			{#if total > perPage}
				<div class="px-5 py-3 border-t flex items-center justify-between text-sm text-gray-500">
					<span>{toPersianNum(total)} رویداد</span>
					<div class="flex gap-1">
						<button disabled={page <= 1} onclick={() => { page--; loadLogs(); }} class="px-3 py-1 border rounded hover:bg-gray-50 disabled:opacity-50">قبلی</button>
						<span class="px-3 py-1">صفحه {toPersianNum(page)} از {toPersianNum(Math.ceil(total / perPage))}</span>
						<button disabled={page >= Math.ceil(total / perPage)} onclick={() => { page++; loadLogs(); }} class="px-3 py-1 border rounded hover:bg-gray-50 disabled:opacity-50">بعدی</button>
					</div>
				</div>
			{/if}
		</div>
	{/if}
</div>
