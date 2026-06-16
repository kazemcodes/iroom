<script lang="ts">
	import { api } from '$lib/api';
	import { onMount } from 'svelte';
	import type { FileItem, Session } from '$lib/types';

	let sessions = $state<Session[]>([]);
	let selectedSessionId = $state<number | null>(null);
	let files = $state<FileItem[]>([]);
	let loading = $state(false);
	let uploading = $state(false);
	let fileInput = $state<HTMLInputElement | null>(null);

	onMount(async () => {
		const res = await api.get<Session[]>('/sessions');
		if (res.success && res.data) {
			sessions = Array.isArray(res.data) ? res.data : [];
			if (sessions.length > 0) {
				selectedSessionId = sessions[0].id;
				await loadFiles();
			}
		}
	});

	async function loadFiles() {
		if (!selectedSessionId) return;
		loading = true;
		const res = await api.get<FileItem[]>(`/sessions/${selectedSessionId}/files`);
		if (res.success && res.data) {
			files = Array.isArray(res.data) ? res.data : [];
		} else {
			files = [];
		}
		loading = false;
	}

	async function handleUpload(e: Event) {
		const input = e.target as HTMLInputElement;
		if (!input.files?.length || !selectedSessionId) return;

		uploading = true;

		for (let i = 0; i < input.files.length; i++) {
			const file = input.files[i];
			const formData = new FormData();
			formData.append('file', file);

			try {
				const token = localStorage.getItem('access_token');
				const res = await fetch(`/api/v1/sessions/${selectedSessionId}/files`, {
					method: 'POST',
					headers: token ? { Authorization: `Bearer ${token}` } : {},
					body: formData
				});

				const data = await res.json();
				if (data.success && data.data) {
					files = [data.data, ...files];
				}
			} catch (err) {
				console.error('Upload failed:', err);
			}
		}

		uploading = false;
		input.value = '';
	}

	function formatSize(bytes: number) {
		if (bytes < 1024) return bytes + ' B';
		if (bytes < 1048576) return (bytes / 1024).toFixed(1) + ' KB';
		return (bytes / 1048576).toFixed(1) + ' MB';
	}

	function formatDate(d: string) {
		if (!d) return '';
		return new Date(d).toLocaleDateString('fa-IR', { year: 'numeric', month: 'long', day: 'numeric', hour: '2-digit', minute: '2-digit' });
	}

	function getFileIcon(filename: string) {
		const ext = filename.split('.').pop()?.toLowerCase();
		if (['jpg', 'jpeg', 'png', 'gif', 'webp', 'svg'].includes(ext || '')) return '🖼️';
		if (['mp4', 'webm', 'mov'].includes(ext || '')) return '🎬';
		if (['mp3', 'wav', 'ogg', 'flac'].includes(ext || '')) return '🎵';
		if (['pdf'].includes(ext || '')) return '📄';
		if (['doc', 'docx'].includes(ext || '')) return '📝';
		if (['xls', 'xlsx'].includes(ext || '')) return '📊';
		if (['zip', 'rar', '7z'].includes(ext || '')) return '📦';
		return '📎';
	}
</script>

<div class="space-y-6">
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-bold text-gray-900">فایل‌ها</h1>
			<p class="text-gray-500 mt-1">{files.length} فایل</p>
		</div>
		<div class="flex items-center gap-3">
			<input type="file" multiple bind:this={fileInput} onchange={handleUpload} class="hidden" />
			<button onclick={() => fileInput?.click()} disabled={uploading || !selectedSessionId} class="px-4 py-2.5 bg-blue-600 text-white text-sm rounded-xl hover:bg-blue-700 disabled:opacity-50 transition-colors font-medium flex items-center gap-2">
				{#if uploading}
					<div class="animate-spin h-4 w-4 border-2 border-white border-t-transparent rounded-full"></div>
					در حال آپلود...
				{:else}
					<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12" /></svg>
					آپلود فایل
				{/if}
			</button>
		</div>
	</div>

	<!-- Session Selector -->
	<div class="bg-white border border-gray-200 rounded-xl p-5">
		<div class="flex items-center gap-3">
			<label class="text-sm font-medium text-gray-700">جلسه:</label>
			<select
				bind:value={selectedSessionId}
				onchange={loadFiles}
				class="flex-1 px-4 py-2.5 border border-gray-200 rounded-xl text-sm focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none bg-white"
			>
				{#if sessions.length === 0}
					<option value={null}>جلسه‌ای موجود نیست</option>
				{:else}
					{#each sessions as s}
						<option value={s.id}>{s.title} ({s.status})</option>
					{/each}
				{/if}
			</select>
		</div>
	</div>

	{#if loading}
		<div class="flex items-center justify-center py-20">
			<div class="animate-spin h-8 w-8 border-4 border-blue-600 border-t-transparent rounded-full"></div>
		</div>
	{:else if files.length === 0}
		<div class="text-center py-20 bg-white rounded-xl border">
			<svg class="w-12 h-12 text-gray-300 mx-auto mb-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M19.5 14.25v-2.625a3.375 3.375 0 00-3.375-3.375h-1.5A1.125 1.125 0 0113.5 7.125v-1.5a3.375 3.375 0 00-3.375-3.375H8.25m6.75 12H9.75m0-3h6m-6 6h6M3.375 6.75h17.25a.375.375 0 01.375.375v11.25a.375.375 0 01-.375.375H3.375a.375.375 0 01-.375-.375V7.125a.375.375 0 01.375-.375z" /></svg>
			<p class="text-gray-500">فایلی وجود ندارد</p>
		</div>
	{:else}
		<div class="bg-white border border-gray-200 rounded-xl overflow-hidden">
			<table class="w-full">
				<thead>
					<tr class="border-b border-gray-100">
						<th class="text-right px-5 py-3 text-xs font-semibold text-gray-500">فایل</th>
						<th class="text-right px-5 py-3 text-xs font-semibold text-gray-500">اندازه</th>
						<th class="text-right px-5 py-3 text-xs font-semibold text-gray-500">تاریخ</th>
					</tr>
				</thead>
				<tbody>
					{#each files as file}
						<tr class="border-b border-gray-50 hover:bg-gray-50 transition-colors">
							<td class="px-5 py-3.5">
								<div class="flex items-center gap-3">
									<span class="text-xl">{getFileIcon(file.filename)}</span>
									<span class="text-sm font-medium text-gray-800 truncate max-w-[300px]">{file.filename}</span>
								</div>
							</td>
							<td class="px-5 py-3.5 text-sm text-gray-500">{formatSize(file.filesize)}</td>
							<td class="px-5 py-3.5 text-sm text-gray-500">{formatDate(file.created_at)}</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	{/if}
</div>
