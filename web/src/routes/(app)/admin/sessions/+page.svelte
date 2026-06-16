<script lang="ts">
	import { api } from '$lib/api';
	import { onMount } from 'svelte';
	import type { Session } from '$lib/types';

	let sessions = $state<Session[]>([]);
	let total = $state(0);
	let currentPage = $state(1);
	let search = $state('');
	let statusFilter = $state('all');
	let loading = $state(true);

	const perPage = 20;

	onMount(() => loadSessions());

	async function loadSessions() {
		loading = true;
		const params: Record<string, string> = { page: String(currentPage), per_page: String(perPage) };
		if (search) params.search = search;
		if (statusFilter !== 'all') params.status = statusFilter;
		const res = await api.get<{ items: Session[]; total: number }>('/admin/sessions', params);
		if (res.success && res.data) {
			sessions = res.data.items || [];
			total = res.data.total;
		}
		loading = false;
	}

	function searchSessions() {
		currentPage = 1;
		loadSessions();
	}

	function formatDate(d: string) {
		if (!d) return '';
		return new Date(d).toLocaleDateString('fa-IR', { year: 'numeric', month: 'short', day: 'numeric' });
	}

	function formatDuration(d: number) {
		return toPersian(d) + ' دقیقه';
	}

	function toPersian(n: number): string {
		return n.toLocaleString('fa-IR');
	}

	const statusLabels: Record<string, string> = {
		live: 'در حال برگزاری',
		scheduled: 'برنامه‌ریزی شده',
		ended: 'پایان یافته'
	};
	const statusColors: Record<string, string> = {
		live: 'bg-green-100 text-green-700',
		scheduled: 'bg-blue-100 text-blue-700',
		ended: 'bg-gray-100 text-gray-500'
	};
</script>

<div class="space-y-6">
	<div>
		<h1 class="text-2xl font-bold text-gray-900">مدیریت جلسات</h1>
		<p class="text-gray-500 mt-1">{toPersian(total)} جلسه</p>
	</div>

	<!-- Filters -->
	<div class="flex items-center gap-3 flex-wrap">
		<input type="text" bind:value={search} onkeydown={(e) => e.key === 'Enter' && searchSessions()} class="flex-1 min-w-[200px] px-4 py-2.5 border rounded-lg text-sm focus:ring-2 focus:ring-blue-500 outline-none bg-white" placeholder="جستجوی عنوان..." />
		<select bind:value={statusFilter} onchange={() => { currentPage = 1; loadSessions(); }} class="px-4 py-2.5 border rounded-lg text-sm focus:ring-2 focus:ring-blue-500 outline-none bg-white">
			<option value="all">همه وضعیت‌ها</option>
			<option value="scheduled">برنامه‌ریزی شده</option>
			<option value="live">در حال برگزاری</option>
			<option value="ended">پایان یافته</option>
		</select>
	</div>

	{#if loading}
		<div class="flex items-center justify-center py-12"><div class="animate-spin h-8 w-8 border-4 border-blue-600 border-t-transparent rounded-full"></div></div>
	{:else if sessions.length === 0}
		<div class="text-center py-20 bg-white rounded-xl border">
			<p class="text-gray-500">جلسه‌ای یافت نشد</p>
		</div>
	{:else}
		<div class="bg-white rounded-xl border overflow-hidden">
			<table class="w-full text-sm">
				<thead class="bg-gray-50 border-b">
					<tr>
						<th class="px-5 py-3 text-right font-medium text-gray-600">عنوان</th>
						<th class="px-5 py-3 text-right font-medium text-gray-600">تاریخ</th>
						<th class="px-5 py-3 text-right font-medium text-gray-600">مدت</th>
						<th class="px-5 py-3 text-right font-medium text-gray-600">وضعیت</th>
					</tr>
				</thead>
				<tbody class="divide-y">
					{#each sessions as s}
						<tr class="hover:bg-gray-50">
							<td class="px-5 py-3 font-medium">{s.title}</td>
							<td class="px-5 py-3 text-gray-500">{formatDate(s.scheduled_at)}</td>
							<td class="px-5 py-3">{formatDuration(s.duration)}</td>
							<td class="px-5 py-3">
								<span class="text-xs px-2 py-1 rounded-full font-medium {statusColors[s.status] || 'bg-gray-100 text-gray-500'}">
									{statusLabels[s.status] || s.status}
								</span>
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
			{#if total > perPage}
				<div class="px-5 py-3 border-t flex items-center justify-between text-sm text-gray-500">
					<span>{toPersian(total)} جلسه</span>
					<div class="flex gap-1">
						<button disabled={currentPage <= 1} onclick={() => { currentPage--; loadSessions(); }} class="px-3 py-1 border rounded hover:bg-gray-50 disabled:opacity-50">قبلی</button>
						<span class="px-3 py-1">صفحه {toPersian(currentPage)} از {toPersian(Math.ceil(total / perPage))}</span>
						<button disabled={currentPage >= Math.ceil(total / perPage)} onclick={() => { currentPage++; loadSessions(); }} class="px-3 py-1 border rounded hover:bg-gray-50 disabled:opacity-50">بعدی</button>
					</div>
				</div>
			{/if}
		</div>
	{/if}
</div>
