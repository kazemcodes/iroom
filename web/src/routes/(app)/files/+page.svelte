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

	// Per-file upload progress tracking
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

			// Initialize upload state
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
		// Clear upload states after a delay
		setTimeout(() => {
			uploads = [];
		}, 3000);
	}

	async function handleUpload(e: Event) {
		const input = e.target as HTMLInputElement;
		await handleFiles(input.files);
		input.value = '';
	}

	// Drag and drop handlers
	function handleDragOver(e: DragEvent) {
		e.preventDefault();
		e.stopPropagation();
		isDragging = true;
	}

	function handleDragLeave(e: DragEvent) {
		e.preventDefault();
		e.stopPropagation();
		isDragging = false;
	}

	function handleDrop(e: DragEvent) {
		e.preventDefault();
		e.stopPropagation();
		isDragging = false;
		if (e.dataTransfer?.files) {
			handleFiles(e.dataTransfer.files);
		}
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

	function confirmDelete(file: FileItem) {
		deleteTarget = file;
		showDeleteModal = true;
	}

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

	function cancelDelete() {
		showDeleteModal = false;
		deleteTarget = null;
	}

	function goToPage(page: number) {
		if (page < 1 || page > totalPages || page === currentPage) return;
		currentPage = page;
		loadFiles();
	}

	// Generate page numbers to display
	const pageNumbers = $derived(() => {
		const pages: number[] = [];
		const maxVisible = 5;
		let start = Math.max(1, currentPage - Math.floor(maxVisible / 2));
		const end = Math.min(totalPages, start + maxVisible - 1);
		if (end - start + 1 < maxVisible) {
			start = Math.max(1, end - maxVisible + 1);
		}
		for (let i = start; i <= end; i++) {
			pages.push(i);
		}
		return pages;
	});
</script>

<div class="space-y-6">
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-bold" style="color: var(--sr-text);">فایل‌ها</h1>
			<p style="color: var(--sr-text-secondary);">{toPersianNum(totalFiles)} فایل</p>
		</div>
		<div class="flex items-center gap-3">
			<input type="file" multiple bind:this={fileInput} onchange={handleUpload} class="hidden" />
		</div>
	</div>

	<!-- Drag and Drop Upload Zone -->
	<div
		bind:this={dropZone}
		ondragover={handleDragOver}
		ondragleave={handleDragLeave}
		ondrop={handleDrop}
		onclick={() => fileInput?.click()}
		class="relative border-2 border-dashed rounded-xl p-8 text-center cursor-pointer transition-all duration-200"
		style={isDragging
			? 'border-color: var(--sr-primary); background: rgba(35, 185, 215, 0.1);'
			: 'border-color: var(--sr-border); background: var(--sr-pure);'}
	>
		{#if isDragging}
			<div class="pointer-events-none">
				<svg class="w-12 h-12 text-blue-500 mx-auto mb-3 animate-bounce" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" /></svg>
				<p class="text-blue-600 font-medium text-lg">فایل‌ها را اینجا رها کنید</p>
			</div>
		{:else}
			<div class="pointer-events-none">
				<svg class="w-12 h-12 text-gray-400 mx-auto mb-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12" /></svg>
				<p class="text-gray-600 font-medium">فایل‌ها را بکشید و اینجا رها کنید</p>
				<p class="text-gray-400 text-sm mt-1">یا کلیک کنید تا فایل انتخاب شود</p>
			</div>
		{/if}
	</div>

	<!-- Upload Progress -->
	{#if uploads.length > 0}
		<div class="card p-4 space-y-3">
			{#each uploads as upload, i}
				<div class="flex items-center gap-3">
					<div class="flex-1 min-w-0">
						<div class="flex items-center justify-between mb-1">
							<span class="text-sm truncate" style="color: var(--sr-text);">{upload.filename}</span>
							<span class="text-sm font-medium" style="color: {upload.status === 'success'
								? 'var(--sr-success)'
								: upload.status === 'error'
									? 'var(--sr-danger)'
									: 'var(--sr-primary)'};">
								{#if upload.status === 'success'}
									✓ آپلود شد
								{:else if upload.status === 'error'}
									✗ {upload.errorMsg || 'خطا'}
								{:else}
									{toPersianNum(upload.progress)}%
								{/if}
							</span>
						</div>
						<div class="w-full bg-gray-200 rounded-full h-2">
							<div
								class="h-2 rounded-full transition-all duration-300 {upload.status === 'success'
									? 'bg-green-500'
									: upload.status === 'error'
										? 'bg-red-500'
										: 'bg-blue-600'}"
								style="width: {upload.progress}%"
							></div>
						</div>
					</div>
				</div>
			{/each}
		</div>
	{/if}

	<!-- Session Selector -->
	<div class="card p-5">
		<div class="flex items-center gap-3">
			<label class="text-sm font-medium" style="color: var(--sr-text-secondary);">جلسه:</label>
			<select
				bind:value={selectedSessionId}
				onchange={() => { currentPage = 1; activeFilter = 'all'; loadFiles(); }}
				class="input-field flex-1"
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

	<!-- File Type Filters -->
	<div class="flex items-center gap-2 flex-wrap">
		{#each filters as filter}
			<button
				onclick={() => { activeFilter = filter.key; }}
				class="px-4 py-2 text-sm rounded-xl font-medium transition-colors"
				style={activeFilter === filter.key
					? 'background: var(--sr-primary); color: white;'
					: 'background: var(--sr-pure); color: var(--sr-text-secondary); border: 1px solid var(--sr-border);'}
			>
				{filter.label}
				{#if filter.key !== 'all'}
					<span class="mr-1 text-xs opacity-75">
						({toPersianNum(files.filter((f) => getFileCategory(f.filename) === filter.key).length)})
					</span>
				{:else}
					<span class="mr-1 text-xs opacity-75">({toPersianNum(files.length)})</span>
				{/if}
			</button>
		{/each}
	</div>

	{#if loading}
		<div class="flex items-center justify-center py-20">
			<div class="animate-spin h-8 w-8 border-4 border-blue-600 border-t-transparent rounded-full"></div>
		</div>
	{:else if filteredFiles.length === 0}
		<div class="text-center py-20 card">
			<svg class="w-12 h-12 mx-auto mb-3" style="color: var(--sr-text-secondary);" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M19.5 14.25v-2.625a3.375 3.375 0 00-3.375-3.375h-1.5A1.125 1.125 0 0113.5 7.125v-1.5a3.375 3.375 0 00-3.375-3.375H8.25m6.75 12H9.75m0-3h6m-6 6h6M3.375 6.75h17.25a.375.375 0 01.375.375v11.25a.375.375 0 01-.375.375H3.375a.375.375 0 01-.375-.375V7.125a.375.375 0 01.375-.375z" /></svg>
			<p style="color: var(--sr-text-secondary);">فایلی وجود ندارد</p>
		</div>
	{:else}
		<div class="overflow-hidden" style="background: var(--sr-pure); border: 1px solid var(--sr-border); border-radius: 0.75rem;">
			<table class="w-full">
				<thead>
					<tr style="border-bottom: 1px solid var(--sr-border);">
						<th class="text-right px-5 py-3 text-xs font-semibold" style="color: var(--sr-text-secondary);">فایل</th>
						<th class="text-right px-5 py-3 text-xs font-semibold" style="color: var(--sr-text-secondary);">اندازه</th>
						<th class="text-right px-5 py-3 text-xs font-semibold" style="color: var(--sr-text-secondary);">تاریخ</th>
						<th class="text-right px-5 py-3 text-xs font-semibold" style="color: var(--sr-text-secondary);">عملیات</th>
					</tr>
				</thead>
				<tbody>
					{#each filteredFiles as file}
						<tr class="table-row" style="border-bottom: 1px solid var(--sr-border);">
							<td class="px-5 py-3.5">
								<div class="flex items-center gap-3">
									<span class="text-xl">{getFileIcon(file.filename)}</span>
									<span class="text-sm font-medium truncate max-w-[300px]" style="color: var(--sr-text);">{file.filename}</span>
								</div>
							</td>
							<td class="px-5 py-3.5 text-sm" style="color: var(--sr-text-secondary);">{formatSize(file.filesize)}</td>
							<td class="px-5 py-3.5 text-sm" style="color: var(--sr-text-secondary);">{toPersianDateTime(file.created_at)}</td>
							<td class="px-5 py-3.5">
								<a href="{api.getBaseUrl()}/files/{file.id}/download" 
									class="p-2 rounded-lg transition-colors inline-flex"
									style="color: var(--sr-text-secondary);"
									title="دانلود" download>
									<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" /></svg>
								</a>
								<button
									onclick={() => confirmDelete(file)}
									class="p-2 rounded-lg transition-colors"
									style="color: var(--sr-text-secondary);"
									title="حذف فایل"
								>
									<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" /></svg>
								</button>
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	{/if}

	<!-- Pagination -->
	{#if totalPages > 1}
		<div class="flex items-center justify-between text-sm text-gray-500">
			<span>{toPersianNum(totalFiles)} فایل</span>
			<div class="flex items-center gap-1">
				<!-- Previous button -->
				<button
					disabled={currentPage <= 1}
					onclick={() => goToPage(currentPage - 1)}
					class="px-3 py-1.5 border rounded-lg hover:bg-gray-50 disabled:opacity-40 disabled:cursor-not-allowed transition-colors"
				>
					قبلی
				</button>

				<!-- First page + ellipsis -->
				{#if pageNumbers()[0] > 1}
					<button
						onclick={() => goToPage(1)}
						class="px-3 py-1.5 border rounded-lg hover:bg-gray-50 transition-colors"
					>
						{toPersianNum(1)}
					</button>
					{#if pageNumbers()[0] > 2}
						<span class="px-2 text-gray-400">...</span>
					{/if}
				{/if}

				<!-- Page numbers -->
				{#each pageNumbers() as p}
					<button
						onclick={() => goToPage(p)}
						class="px-3 py-1.5 border rounded-lg transition-colors {p === currentPage
							? 'bg-blue-600 text-white border-blue-600'
							: 'hover:bg-gray-50'}"
					>
						{toPersianNum(p)}
					</button>
				{/each}

				<!-- Last page + ellipsis -->
				{#if pageNumbers()[pageNumbers().length - 1] < totalPages}
					{#if pageNumbers()[pageNumbers().length - 1] < totalPages - 1}
						<span class="px-2 text-gray-400">...</span>
					{/if}
					<button
						onclick={() => goToPage(totalPages)}
						class="px-3 py-1.5 border rounded-lg hover:bg-gray-50 transition-colors"
					>
						{toPersianNum(totalPages)}
					</button>
				{/if}

				<!-- Next button -->
				<button
					disabled={currentPage >= totalPages}
					onclick={() => goToPage(currentPage + 1)}
					class="px-3 py-1.5 border rounded-lg hover:bg-gray-50 disabled:opacity-40 disabled:cursor-not-allowed transition-colors"
				>
					بعدی
				</button>
			</div>
			<span>صفحه {toPersianNum(currentPage)} از {toPersianNum(totalPages)}</span>
		</div>
	{/if}
</div>

<!-- Delete Confirmation Modal -->
<ConfirmModal
	show={showDeleteModal}
	title="حذف فایل"
	message="آیا از حذف فایل {deleteTarget?.filename} اطمینان دارید؟ این عمل قابل بازگشت نیست."
	onConfirm={deleteFile}
	onCancel={cancelDelete}
/>
