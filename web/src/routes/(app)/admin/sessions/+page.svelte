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
	const statusColors: Record<string, string> = { scheduled: 'bg-blue-50 text-blue-700', live: 'bg-green-50 text-green-700', ended: 'bg-gray-100 text-gray-600' };

	const filteredSessions = $derived(sessions.filter(s => {
		if (filterStatus !== 'all' && s.status !== filterStatus) return false;
		if (searchQuery && !s.title?.includes(searchQuery)) return false;
		return true;
	}));
</script>

<div class="space-y-6">
	<div>
		<h1 class="text-2xl font-bold text-gray-900">مدیریت جلسات</h1>
		<p class="text-gray-500 mt-1">مشاهده و مدیریت تمام جلسات</p>
	</div>

	<div class="flex items-center gap-3">
		<div class="flex-1 relative">
			<svg class="absolute right-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" /></svg>
			<input bind:value={searchQuery} class="w-full pr-10 pl-4 py-2.5 border border-gray-200 rounded-xl text-sm focus:ring-2 focus:ring-blue-500 outline-none bg-white" placeholder="جستجو در عنوان جلسه..." />
		</div>
		<select bind:value={filterStatus} class="px-3 py-2.5 border border-gray-200 rounded-xl text-sm focus:ring-2 focus:ring-blue-500 outline-none bg-white">
			<option value="all">همه</option>
			<option value="scheduled">برنامه‌ریزی شده</option>
			<option value="live">در حال برگزاری</option>
			<option value="ended">پایان یافته</option>
		</select>
	</div>

	{#if loading}
		<div class="flex items-center justify-center py-20">
			<div class="animate-spin h-8 w-8 border-4 border-blue-600 border-t-transparent rounded-full"></div>
		</div>
	{:else if filteredSessions.length === 0}
		<div class="text-center py-20 bg-white rounded-xl border">
			<p class="text-gray-500">جلسه‌ای یافت نشد</p>
		</div>
	{:else}
		<div class="bg-white rounded-xl border overflow-hidden">
			<table class="w-full text-sm">
				<thead class="bg-gray-50 border-b">
					<tr>
						<th class="px-4 py-3 text-right font-medium text-gray-600">عنوان</th>
						<th class="px-4 py-3 text-right font-medium text-gray-600">وضعیت</th>
						<th class="px-4 py-3 text-right font-medium text-gray-600">تاریخ</th>
						<th class="px-4 py-3 text-right font-medium text-gray-600">مدت</th>
						<th class="px-4 py-3 text-right font-medium text-gray-600">عملیات</th>
					</tr>
				</thead>
				<tbody class="divide-y">
					{#each filteredSessions as s (s.id)}
						<tr class="hover:bg-gray-50">
							<td class="px-4 py-3 font-medium text-gray-900">{s.title}</td>
							<td class="px-4 py-3">
								<span class="text-xs px-2 py-1 rounded-full {statusColors[s.status] || 'bg-gray-100 text-gray-600'}">
									{statusLabels[s.status] || s.status}
								</span>
							</td>
							<td class="px-4 py-3 text-gray-500 text-xs">{s.scheduled_at ? new Date(s.scheduled_at).toLocaleDateString('fa-IR') : '—'}</td>
							<td class="px-4 py-3 text-gray-600">{s.duration ? `${toPersianNum(s.duration)} دقیقه` : '—'}</td>
							<td class="px-4 py-3">
								<button onclick={() => deleteSession(s.id)} class="p-1.5 text-gray-400 hover:text-red-600 rounded-lg hover:bg-red-50" title="حذف">
									<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" /></svg>
								</button>
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	{/if}
</div>
