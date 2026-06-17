<script lang="ts">
	import { api } from '$lib/api';
	import { onMount } from 'svelte';
	import type { Recording } from '$lib/types';
	import ConfirmModal from '$lib/components/ConfirmModal.svelte';

	let recordings = $state<Recording[]>([]);
	let total = $state(0);
	let page = $state(1);
	let loading = $state(true);
	let search = $state('');

	let showDeleteConfirm = $state(false);
	let deleteTargetId = $state(0);

	const perPage = 20;

	onMount(() => loadRecordings());

	async function loadRecordings() {
		loading = true;
		const params: Record<string, string> = { page: String(page), per_page: String(perPage) };
		if (search) params.search = search;
		const res = await api.get<{ items: Recording[]; total: number }>('/admin/recordings', params);
		if (res.success && res.data) {
			recordings = res.data.items || [];
			total = res.data.total;
		}
		loading = false;
	}

	async function deleteRecording(id: number) {
		const res = await api.delete(`/admin/recordings/${id}`);
		if (res.success) {
			recordings = recordings.filter(r => r.id !== id);
			total--;
		}
	}

	function confirmDeleteRecording(id: number) {
		deleteTargetId = id;
		showDeleteConfirm = true;
	}

	function formatSize(bytes: number) {
		if (bytes < 1024) return bytes + ' B';
		if (bytes < 1048576) return (bytes / 1024).toFixed(1) + ' KB';
		return (bytes / 1048576).toFixed(1) + ' MB';
	}

	function formatDate(d: string) {
		return new Date(d).toLocaleDateString('fa-IR', { year: 'numeric', month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' });
	}
</script>

<div class="space-y-6">
	<div>
		<h1 class="text-2xl font-bold text-gray-900">ضبط‌ها</h1>
		<p class="text-gray-500 mt-1">{total} ضبط</p>
	</div>

	<div class="flex gap-3">
		<input type="text" bind:value={search} onkeydown={(e) => e.key === 'Enter' && (page = 1, loadRecordings())} class="flex-1 px-4 py-2.5 border rounded-lg text-sm focus:ring-2 focus:ring-blue-500 outline-none bg-white" placeholder="جستجو..." />
	</div>

	{#if loading}
		<div class="flex items-center justify-center py-12"><div class="animate-spin h-8 w-8 border-4 border-blue-600 border-t-transparent rounded-full"></div></div>
	{:else if recordings.length === 0}
		<div class="text-center py-20 bg-white rounded-xl border">
			<p class="text-gray-500">ضبطی یافت نشد</p>
		</div>
	{:else}
		<div class="bg-white rounded-xl border overflow-hidden">
			<table class="w-full text-sm">
				<thead class="bg-gray-50 border-b">
					<tr>
						<th class="px-5 py-3 text-right font-medium text-gray-600">فایل</th>
						<th class="px-5 py-3 text-right font-medium text-gray-600">جلسه</th>
						<th class="px-5 py-3 text-right font-medium text-gray-600">حجم</th>
						<th class="px-5 py-3 text-right font-medium text-gray-600">وضعیت</th>
						<th class="px-5 py-3 text-right font-medium text-gray-600">تاریخ</th>
						<th class="px-5 py-3 text-right font-medium text-gray-600">عملیات</th>
					</tr>
				</thead>
				<tbody class="divide-y">
					{#each recordings as rec}
						<tr class="hover:bg-gray-50">
							<td class="px-5 py-3 font-medium">{rec.filename}</td>
							<td class="px-5 py-3 text-gray-500">جلسه #{rec.session_id}</td>
							<td class="px-5 py-3">{formatSize(rec.filesize)}</td>
							<td class="px-5 py-3">
								<span class="text-xs px-2 py-1 rounded-full font-medium {rec.status === 'ready' ? 'bg-green-100 text-green-700' : 'bg-yellow-100 text-yellow-700'}">
									{rec.status === 'ready' ? 'آماده' : 'پردازش'}
								</span>
							</td>
							<td class="px-5 py-3 text-gray-500">{formatDate(rec.created_at)}</td>
							<td class="px-5 py-3">
								<div class="flex items-center gap-1">
									<a href="/recordings/{rec.session_id}" class="px-2 py-1 text-xs text-blue-600 hover:bg-blue-50 rounded">مشاهده</a>
									<button onclick={() => confirmDeleteRecording(rec.id)} class="px-2 py-1 text-xs text-red-600 hover:bg-red-50 rounded">حذف</button>
								</div>
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
			{#if total > perPage}
				<div class="px-5 py-3 border-t flex items-center justify-between text-sm text-gray-500">
					<span>{total} ضبط</span>
					<div class="flex gap-1">
						<button disabled={page <= 1} onclick={() => { page--; loadRecordings(); }} class="px-3 py-1 border rounded hover:bg-gray-50 disabled:opacity-50">قبلی</button>
						<span class="px-3 py-1">صفحه {page} از {Math.ceil(total / perPage)}</span>
						<button disabled={page >= Math.ceil(total / perPage)} onclick={() => { page++; loadRecordings(); }} class="px-3 py-1 border rounded hover:bg-gray-50 disabled:opacity-50">بعدی</button>
					</div>
				</div>
			{/if}
		</div>

<ConfirmModal bind:show={showDeleteConfirm} title="حذف ضبط" message="آیا از حذف این ضبط اطمینان دارید؟" onConfirm={() => deleteRecording(deleteTargetId)} onCancel={() => {}} />
	{/if}
</div>
