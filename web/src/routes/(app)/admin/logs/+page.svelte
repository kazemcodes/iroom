<script lang="ts">
	import { api } from '$lib/api';
	import { onMount } from 'svelte';
	import type { ActivityLog } from '$lib/types';

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

	function formatDate(d: string) {
		return new Date(d).toLocaleDateString('fa-IR', { year: 'numeric', month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' });
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
</script>

<div class="space-y-6">
	<div>
		<h1 class="text-2xl font-bold text-gray-900">لاگ فعالیت‌ها</h1>
		<p class="text-gray-500 mt-1">{total} رویداد</p>
	</div>

	{#if loading}
		<div class="flex items-center justify-center py-12"><div class="animate-spin h-8 w-8 border-4 border-blue-600 border-t-transparent rounded-full"></div></div>
	{:else if logs.length === 0}
		<div class="text-center py-20 bg-white rounded-xl border">
			<p class="text-gray-500">لاگی ثبت نشده</p>
		</div>
	{:else}
		<div class="bg-white rounded-xl border overflow-hidden">
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
							<td class="px-5 py-3 text-gray-500">#{log.entity_id}</td>
							<td class="px-5 py-3 text-gray-500 max-w-[200px] truncate">{log.details || '-'}</td>
							<td class="px-5 py-3 text-gray-500">{formatDate(log.created_at)}</td>
						</tr>
					{/each}
				</tbody>
			</table>
			{#if total > perPage}
				<div class="px-5 py-3 border-t flex items-center justify-between text-sm text-gray-500">
					<span>{total} رویداد</span>
					<div class="flex gap-1">
						<button disabled={page <= 1} onclick={() => { page--; loadLogs(); }} class="px-3 py-1 border rounded hover:bg-gray-50 disabled:opacity-50">قبلی</button>
						<span class="px-3 py-1">صفحه {page} از {Math.ceil(total / perPage)}</span>
						<button disabled={page >= Math.ceil(total / perPage)} onclick={() => { page++; loadLogs(); }} class="px-3 py-1 border rounded hover:bg-gray-50 disabled:opacity-50">بعدی</button>
					</div>
				</div>
			{/if}
		</div>
	{/if}
</div>
