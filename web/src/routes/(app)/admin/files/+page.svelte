<script lang="ts">
	import { api } from '$lib/api';
	import { onMount } from 'svelte';
	import type { FileItem, Session } from '$lib/types';
	import { toPersianNum, toPersianDateTime } from '$lib/utils/persian';
	import ConfirmModal from '$lib/components/ConfirmModal.svelte';

	let sessions = $state<Session[]>([]);
	let selectedSessionId = $state<number | null>(null);
	let files = $state<FileItem[]>([]);
	let loading = $state(false);
	let uploading = $state(false);
	let fileInput = $state<HTMLInputElement | null>(null);
	let dropZone = $state<HTMLDivElement | null>(null);
	let isDragging = $state(false);

	let currentPage = $state(1);
	let totalFiles = $state(0);
	let activeFilter = $state<string>('all');
	const perPage = 20;

	let deleteTarget = $state<FileItem | null>(null);
	let showDeleteModal = $state(false);

	interface UploadState {
		filename: string;
		progress: number;
		status: 'uploading' | 'success' | 'error';
		errorMsg?: string;
	}
	let uploads = $state<UploadState[]>([]);

	const totalPages = $derived(Math.ceil(totalFiles / perPage));

	const filters = [
		{ key: 'all', label: 'همه' },
		{ key: 'images', label: 'تصاویر' },
		{ key: 'documents', label: 'اسناد' },
		{ key: 'videos', label: 'ویدیو' },
		{ key: 'other', label: 'سایر' }
	];

	function getFileCategory(filename: string): string {
		const ext = filename.split('.').pop()?.toLowerCase() || '';
		if (['jpg', 'jpeg', 'png', 'gif', 'webp', 'svg', 'bmp', 'ico'].includes(ext)) return 'images';
		if (['mp4', 'webm', 'mov', 'avi', 'mkv', 'flv'].includes(ext)) return 'videos';
		if (['pdf', 'doc', 'docx', 'xls', 'xlsx', 'ppt', 'pptx', 'txt', 'csv', 'rtf'].includes(ext)) return 'documents';
		return 'other';
	}

	const filteredFiles = $derived(
		activeFilter === 'all' ? files : files.filter((f) => getFileCategory(f.filename) === activeFilter)
	);

	onMount(async () => {
		const res = await api.get<Session[]>('/sessions');
		if (res.success && res.data) {
			sessions = Array.isArray(res.data) ? res.data : [];
			if (sessions.length > 0) {
				selectedSessionId = sessions[0].id;
				currentPage = 1;
				await loadFiles();
			}
		}
	});

	async function loadFiles() {
		if (!selectedSessionId) return;
		loading = true;
		const params: Record<string, string> = { page: String(currentPage), per_page: String(perPage) };
		const res = await api.get<{ items: FileItem[]; total: number }>(`/sessions/${selectedSessionId}/files`, params);
		if (res.success && res.data) {
			files = res.data.items || (Array.isArray(res.data) ? res.data : []);
			totalFiles = res.data.total || files.length;
		} else {
			files = [];
			totalFiles = 0;
		}
		loading = false;
	}

	function uploadSingleFile(file: File, sessionId: number, index: number): Promise<void> {
		return new Promise((resolve) => {
			const xhr = new XMLHttpRequest();
			const formData = new FormData();
			formData.append('file', file);

			uploads[index] = { filename: file.name, progress: 0, status: 'uploading' };
			uploads = [...uploads];

			xhr.upload.onprogress = (e) => {
				if (e.lengthComputable) {
					const pct = Math.round((e.loaded / e.total) * 100);
					uploads[index] = { ...uploads[index], progress: pct };
					uploads = [...uploads];
				}
			};

			xhr.onload = () => {
				if (xhr.status >= 200 && xhr.status < 300) {
					try {
						const data = JSON.parse(xhr.responseText);
						if (data.success && data.data) {
							files = [data.data, ...files];
							uploads[index] = { ...uploads[index], progress: 100, status: 'success' };
						} else {
							uploads[index] = { ...uploads[index], progress: 100, status: 'error', errorMsg: data.error || 'خطا در آپلود' };
						}
					} catch {
						uploads[index] = { ...uploads[index], progress: 100, status: 'error', errorMsg: 'خطا در پاسخ سرور' };
					}
				} else {
					uploads[index] = { ...uploads[index], progress: 100, status: 'error', errorMsg: `خطای ${xhr.status}` };
				}
				uploads = [...uploads];
				resolve();
			};

			xhr.onerror = () => {
				uploads[index] = { ...uploads[index], progress: 100, status: 'error', errorMsg: 'خطا در اتصال' };
				uploads = [...uploads];
				resolve();
			};

			xhr.open('POST', `/api/v1/sessions/${sessionId}/files`);
			const token = localStorage.getItem('access_token');
			if (token) xhr.setRequestHeader('Authorization', `Bearer ${token}`);
			xhr.send(formData);
		});
	}

	async function handleFiles(fileList: FileList | null) {
		if (!fileList?.length || !selectedSessionId) return;

		uploading = true;
		uploads = [];

		for (let i = 0; i < fileList.length; i++) {
			const file = fileList[i];
			uploads = [...uploads, { filename: file.name, progress: 0, status: 'uploading' as const }];
		}

		for (let i = 0; i < fileList.length; i++) {
			await uploadSingleFile(fileList[i], selectedSessionId, i);
		}

		uploading = false;
		setTimeout(() => { uploads = []; }, 3000);
	}

	async function handleUpload(e: Event) {
		const input = e.target as HTMLInputElement;
		await handleFiles(input.files);
		input.value = '';
	}

	function handleDragOver(e: DragEvent) { e.preventDefault(); e.stopPropagation(); isDragging = true; }
	function handleDragLeave(e: DragEvent) { e.preventDefault(); e.stopPropagation(); isDragging = false; }
	function handleDrop(e: DragEvent) {
		e.preventDefault(); e.stopPropagation(); isDragging = false;
		if (e.dataTransfer?.files) handleFiles(e.dataTransfer.files);
	}

	function formatSize(bytes: number) {
		if (bytes < 1024) return toPersianNum(bytes) + ' بایت';
		if (bytes < 1048576) return toPersianNum((bytes / 1024).toFixed(1)) + ' کیلوبایت';
		return toPersianNum((bytes / 1048576).toFixed(1)) + ' مگابایت';
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

	function confirmDelete(file: FileItem) { deleteTarget = file; showDeleteModal = true; }

	async function deleteFile() {
		if (!deleteTarget) return;
		const res = await api.delete(`/files/${deleteTarget.id}`);
		if (res.success) {
			files = files.filter((f) => f.id !== deleteTarget!.id);
			totalFiles = Math.max(0, totalFiles - 1);
		}
		showDeleteModal = false;
		deleteTarget = null;
	}

	function cancelDelete() { showDeleteModal = false; deleteTarget = null; }

	function goToPage(page: number) {
		if (page < 1 || page > totalPages || page === currentPage) return;
		currentPage = page;
		loadFiles();
	}

	const pageNumbers = $derived(() => {
		const pages: number[] = [];
		const maxVisible = 5;
		let start = Math.max(1, currentPage - Math.floor(maxVisible / 2));
		const end = Math.min(totalPages, start + maxVisible - 1);
		if (end - start + 1 < maxVisible) start = Math.max(1, end - maxVisible + 1);
		for (let i = start; i <= end; i++) pages.push(i);
		return pages;
	});
</script>

<div class="space-y-5">
	<div class="flex items-center justify-between">
		<div>
			<h1 class="sky-page-title">فایل‌ها</h1>
			<p class="sky-page-subtitle">{toPersianNum(totalFiles)} فایل</p>
		</div>
		<input type="file" multiple bind:this={fileInput} onchange={handleUpload} class="hidden" />
	</div>

	<div bind:this={dropZone} ondragover={handleDragOver} ondragleave={handleDragLeave} ondrop={handleDrop}
		onclick={() => fileInput?.click()} role="button" tabindex="-1"
		class="relative rounded-xl p-8 text-center cursor-pointer transition-all duration-200"
		style="border: 2px dashed {isDragging ? 'var(--color-crystal-clear)' : 'var(--color-zen-garden)'}; background: {isDragging ? 'rgba(35,185,215,0.08)' : 'var(--color-pure)'};">
		<div class="pointer-events-none">
			<svg class="mx-auto mb-3 {isDragging ? 'animate-bounce' : ''}" width="44" height="44" fill="none" stroke="{isDragging ? 'var(--color-crystal-clear)' : 'var(--color-moonlit-mist)'}" stroke-width="1.5" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12"/></svg>
			{#if isDragging}
				<p class="font-medium text-lg" style="color: var(--color-crystal-clear);">فایل‌ها را اینجا رها کنید</p>
			{:else}
				<p class="font-medium" style="color: var(--color-midnight-sky);">فایل‌ها را بکشید و اینجا رها کنید</p>
				<p class="text-sm mt-1" style="color: var(--color-moonlit-mist);">یا کلیک کنید تا فایل انتخاب شود</p>
			{/if}
		</div>
	</div>

	{#if uploads.length > 0}
		<div class="sky-card p-4 space-y-3">
			{#each uploads as upload}
				<div class="flex-1 min-w-0">
					<div class="flex items-center justify-between mb-1">
						<span class="text-sm truncate" style="color: var(--color-midnight-sky);">{upload.filename}</span>
						<span class="text-sm font-medium" style="color: {upload.status === 'success' ? 'var(--color-lush-meadow)' : upload.status === 'error' ? 'var(--color-fiery-passion)' : 'var(--color-crystal-clear)'};">
							{#if upload.status === 'success'}آپلود شد{:else if upload.status === 'error'}{upload.errorMsg || 'خطا'}{:else}{toPersianNum(upload.progress)}٪{/if}
						</span>
					</div>
					<div class="w-full rounded-full h-2" style="background: var(--color-zen-garden);">
						<div class="h-2 rounded-full transition-all duration-300" style="width: {upload.progress}%; background: {upload.status === 'success' ? 'var(--color-lush-meadow)' : upload.status === 'error' ? 'var(--color-fiery-passion)' : 'var(--color-crystal-clear)'};"></div>
					</div>
				</div>
			{/each}
		</div>
	{/if}

	<div class="sky-card p-4">
		<div class="flex items-center gap-3">
			<label class="text-sm font-medium shrink-0" style="color: var(--color-mystic-sea);">جلسه:</label>
			<select bind:value={selectedSessionId} onchange={() => { currentPage = 1; activeFilter = 'all'; loadFiles(); }} class="sky-input flex-1">
				{#if sessions.length === 0}
					<option value={null}>جلسه‌ای موجود نیست</option>
				{:else}
					{#each sessions as s}<option value={s.id}>{s.title} ({s.status})</option>{/each}
				{/if}
			</select>
		</div>
	</div>

	<div class="sky-filter-bar w-fit flex-wrap">
		{#each filters as filter}
			<button onclick={() => { activeFilter = filter.key; }} class="sky-filter-btn {activeFilter === filter.key ? 'active' : ''}">
				{filter.label}
				<span class="text-xs opacity-70">({toPersianNum(filter.key === 'all' ? files.length : files.filter((f) => getFileCategory(f.filename) === filter.key).length)})</span>
			</button>
		{/each}
	</div>

	{#if loading}
		<div class="flex items-center justify-center py-16"><svg class="sky-spinner lg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="color: var(--color-crystal-clear);"><path d="M12 2v4M12 18v4M4.93 4.93l2.83 2.83M16.24 16.24l2.83 2.83M2 12h4M18 12h4M4.93 19.07l2.83-2.83M16.24 7.76l2.83-2.83"/></svg></div>
	{:else if filteredFiles.length === 0}
		<div class="sky-card">
			<div class="sky-empty">
				<div class="sky-empty-icon"><svg width="48" height="48" fill="none" stroke="currentColor" stroke-width="1" viewBox="0 0 24 24" style="color: var(--color-muted-mountain);"><path stroke-linecap="round" stroke-linejoin="round" d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z"/><polyline points="14 2 14 8 20 8"/></svg></div>
				<p class="sky-empty-title">فایلی وجود ندارد</p>
				<p class="sky-empty-desc">فایل‌ها را بالا بکشید تا آپلود شوند</p>
			</div>
		</div>
	{:else}
		<div class="sky-card overflow-hidden">
			<table class="sky-table">
				<thead><tr><th>فایل</th><th>اندازه</th><th>تاریخ</th><th>عملیات</th></tr></thead>
				<tbody>
					{#each filteredFiles as file}
						<tr>
							<td>
								<div class="flex items-center gap-3">
									<span class="text-xl">{getFileIcon(file.filename)}</span>
									<span class="text-sm font-medium truncate max-w-[300px]" style="color: var(--color-midnight-sky);">{file.filename}</span>
								</div>
							</td>
							<td style="color: var(--color-mystic-sea);">{formatSize(file.filesize)}</td>
							<td style="color: var(--color-mystic-sea);">{toPersianDateTime(file.created_at)}</td>
							<td>
								<div class="flex items-center gap-1">
									<a href="{api.getBaseUrl()}/files/{file.id}/download" class="sky-btn-icon" style="width:32px;height:32px;" title="دانلود" download>
										<svg width="16" height="16" fill="none" stroke="currentColor" stroke-width="1.75" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4"/></svg>
									</a>
									<button onclick={() => confirmDelete(file)} class="sky-btn-icon" style="width:32px;height:32px;" title="حذف فایل">
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

	{#if totalPages > 1}
		<div class="flex items-center justify-between text-sm" style="color: var(--color-mystic-sea);">
			<span>{toPersianNum(totalFiles)} فایل</span>
			<div class="sky-pagination">
				<button class="sky-page-btn" disabled={currentPage <= 1} onclick={() => goToPage(currentPage - 1)}>
					<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="9 18 15 12 9 6"/></svg>
				</button>
				{#each pageNumbers() as p}
					<button class="sky-page-btn {p === currentPage ? 'active' : ''}" onclick={() => goToPage(p)}>{toPersianNum(p)}</button>
				{/each}
				<button class="sky-page-btn" disabled={currentPage >= totalPages} onclick={() => goToPage(currentPage + 1)}>
					<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="15 18 9 12 15 6"/></svg>
				</button>
			</div>
			<span class="text-xs">صفحه {toPersianNum(currentPage)} از {toPersianNum(totalPages)}</span>
		</div>
	{/if}
</div>

<ConfirmModal show={showDeleteModal} title="حذف فایل" message="آیا از حذف فایل {deleteTarget?.filename} اطمینان دارید؟" onConfirm={deleteFile} onCancel={cancelDelete} />
