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

<div class="space-y-6">
	<div>
		<h1 style="font-size:1.5rem;font-weight:700;color:var(--color-midnight-sky);">مدیریت ضبط‌ها</h1>
		<p style="font-size:0.875rem;color:var(--color-mystic-sea);margin-top:4px;">مشاهده و مدیریت ضبط‌های جلسات</p>
	</div>

	<div class="relative">
		<svg class="absolute right-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" /></svg>
		<input bind:value={searchQuery} class="w-full pr-10 pl-4 py-2.5 border border-gray-200 rounded-xl text-sm focus:ring-2 focus:ring-blue-500 outline-none bg-white" placeholder="جستجو در عنوان جلسه..." />
	</div>

	{#if loading}
		<div class="flex items-center justify-center py-20">
			<div class="animate-spin h-8 w-8 border-4 border-blue-600 border-t-transparent rounded-full"></div>
		</div>
	{:else if filteredRecordings.length === 0}
		<div class="text-center py-20 bg-white rounded-xl">
			<svg class="w-16 h-16 text-gray-300 mx-auto mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1" d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" /></svg>
			<p class="text-gray-500">ضبطی یافت نشد</p>
		</div>
	{:else}
		<div class="bg-white rounded-xl overflow-hidden">
			<table class="w-full text-sm">
				<thead class="bg-gray-50 border-b">
					<tr>
						<th class="px-4 py-3 text-right font-medium text-gray-600">جلسه</th>
						<th class="px-4 py-3 text-right font-medium text-gray-600">مدت</th>
						<th class="px-4 py-3 text-right font-medium text-gray-600">حجم</th>
						<th class="px-4 py-3 text-right font-medium text-gray-600">تاریخ</th>
						<th class="px-4 py-3 text-right font-medium text-gray-600">عملیات</th>
					</tr>
				</thead>
				<tbody class="divide-y">
					{#each filteredRecordings as rec (rec.id)}
						<tr class="hover:bg-gray-50">
							<td class="px-4 py-3 font-medium text-gray-900">{rec.session_title || '—'}</td>
							<td class="px-4 py-3 text-gray-600">{rec.duration ? formatDuration(rec.duration) : '—'}</td>
							<td class="px-4 py-3 text-gray-600">{rec.file_size ? formatSize(rec.file_size) : '—'}</td>
							<td class="px-4 py-3 text-gray-500 text-xs">{rec.created_at ? new Date(rec.created_at).toLocaleDateString('fa-IR') : '—'}</td>
							<td class="px-4 py-3">
								<div class="flex gap-1">
									<a href="/recordings/{rec.id}" class="p-1.5 text-gray-400 hover:text-blue-600 rounded-lg hover:bg-blue-50" title="مشاهده">
										<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" /><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" /></svg>
									</a>
									<button onclick={() => deleteRecording(rec.id)} class="p-1.5 text-gray-400 hover:text-red-600 rounded-lg hover:bg-red-50" title="حذف">
										<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" /></svg>
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
