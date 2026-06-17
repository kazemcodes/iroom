<script lang="ts">
	import { api } from '$lib/api';
	import { onMount } from 'svelte';
	import type { User } from '$lib/types';
	import ConfirmModal from '$lib/components/ConfirmModal.svelte';
	import { toPersianNum } from '$lib/utils/persian';

	let users = $state<User[]>([]);
	let total = $state(0);
	let currentPage = $state(1);
	let search = $state('');
	let roleFilter = $state('all');
	let loading = $state(true);

	const perPage = 15;

	let showCreateModal = $state(false);
	let showEditModal = $state(false);
	let editingUser = $state<User | null>(null);
	let showDeleteConfirm = $state(false);
	let deleteUserId = $state(0);

	let newUser = $state({ email: '', password: '', display_name: '', phone: '', role: 'student' });
	let editForm = $state({ role: 'student', is_active: true });
	let createLoading = $state(false);
	let createError = $state('');

	// Bulk import state
	let showImportModal = $state(false);
	let importFile = $state<File | null>(null);
	let importPreview = $state<string[][]>([]);
	let importLoading = $state(false);
	let importResult = $state<{ success_count: number; errors: { row: number; message: string }[] } | null>(null);
	let isDragging = $state(false);
	let fileInput = $state<HTMLInputElement | null>(null);

	onMount(() => loadUsers());

	async function loadUsers() {
		loading = true;
		const params: Record<string, string> = { page: String(currentPage), per_page: String(perPage) };
		if (search) params.search = search;
		if (roleFilter !== 'all') params.role = roleFilter;
		const res = await api.get<{ items: User[]; total: number }>('/admin/users', params);
		if (res.success && res.data) {
			users = res.data.items || [];
			total = res.data.total;
		}
		loading = false;
	}

	function searchUsers() {
		currentPage = 1;
		loadUsers();
	}

	function openEdit(user: User) {
		editingUser = user;
		editForm = { role: user.role, is_active: user.is_active };
		showEditModal = true;
	}

	async function createUser() {
		createLoading = true;
		createError = '';
		const res = await api.post('/admin/users', newUser);
		if (!res.success) {
			createError = res.error || 'خطا';
			createLoading = false;
			return;
		}
		showCreateModal = false;
		newUser = { email: '', password: '', display_name: '', phone: '', role: 'student' };
		createLoading = false;
		await loadUsers();
	}

	async function saveEdit() {
		if (!editingUser) return;
		const res = await api.put(`/admin/users/${editingUser.id}`, editForm);
		if (res.success) {
			showEditModal = false;
			editingUser = null;
			await loadUsers();
		}
	}

	async function toggleActive(user: User) {
		const res = await api.put(`/admin/users/${user.id}`, { is_active: !user.is_active });
		if (res.success) await loadUsers();
	}

	function confirmDelete(id: number) {
		deleteUserId = id;
		showDeleteConfirm = true;
	}

	async function deleteUser() {
		const res = await api.delete(`/admin/users/${deleteUserId}`);
		if (res.success) await loadUsers();
	}

	// --- Bulk Import ---

	function openImportModal() {
		importFile = null;
		importPreview = [];
		importResult = null;
		importLoading = false;
		showImportModal = true;
	}

	function handleDragOver(e: DragEvent) {
		e.preventDefault();
		isDragging = true;
	}

	function handleDragLeave(e: DragEvent) {
		e.preventDefault();
		isDragging = false;
	}

	function handleDrop(e: DragEvent) {
		e.preventDefault();
		isDragging = false;
		const files = e.dataTransfer?.files;
		if (files && files.length > 0) {
			validateAndSetFile(files[0]);
		}
	}

	function handleFileSelect(e: Event) {
		const target = e.target as HTMLInputElement;
		const files = target.files;
		if (files && files.length > 0) {
			validateAndSetFile(files[0]);
		}
	}

	function validateAndSetFile(file: File) {
		if (!file.name.endsWith('.csv')) {
			return;
		}
		importFile = file;
		parseCSVPreview(file);
	}

	function parseCSVPreview(file: File) {
		const reader = new FileReader();
		reader.onload = (e) => {
			const text = e.target?.result as string;
			const lines = text.split(/\r?\n/).filter((line) => line.trim() !== '');
			// Show header + up to 5 data rows
			importPreview = lines.slice(0, 6).map((line) => {
				// Simple CSV parse (handles basic commas, not quoted fields with commas)
				return line.split(',').map((cell) => cell.trim());
			});
		};
		reader.readAsText(file);
	}

	function downloadTemplate() {
		const csvContent = 'display_name,email,password,role,phone\nعلی رضایی,ali@example.com,Password123,student,09121234567';
		const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
		const url = URL.createObjectURL(blob);
		const a = document.createElement('a');
		a.href = url;
		a.download = 'users_template.csv';
		a.click();
		URL.revokeObjectURL(url);
	}

	async function runImport() {
		if (!importFile) return;
		importLoading = true;
		importResult = null;
		const formData = new FormData();
		formData.append('file', importFile);
		const res = await api.postFormData<{ success_count: number; errors: { row: number; message: string }[] }>(
			'/admin/users/import',
			formData
		);
		if (res.success && res.data) {
			importResult = res.data;
		} else {
			importResult = { success_count: 0, errors: [{ row: 0, message: res.error || 'خطا در وارد کردن فایل' }] };
		}
		importLoading = false;
	}

	function closeImportModal() {
		showImportModal = false;
		importFile = null;
		importPreview = [];
		importResult = null;
		loadUsers();
	}

	function toPersian(n: number): string {
		return n.toLocaleString('fa-IR');
	}

	const roleLabels: Record<string, string> = { admin: 'مدیر', teacher: 'مدرس', student: 'دانش‌آموز' };
	const roleColors: Record<string, string> = { admin: 'bg-red-100 text-red-700', teacher: 'bg-purple-100 text-purple-700', student: 'bg-blue-100 text-blue-700' };
</script>

<div class="space-y-6">
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-bold text-gray-900">مدیریت کاربران</h1>
			<p class="text-gray-500 mt-1">{toPersian(total)} کاربر</p>
		</div>
		<div class="flex items-center gap-3">
			<button onclick={openImportModal} class="px-4 py-2.5 bg-white border border-gray-300 text-gray-700 rounded-lg text-sm font-medium hover:bg-gray-50 flex items-center gap-2">
				<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12" /></svg>
				وارد کردن گروهی
			</button>
			<button onclick={() => showCreateModal = true} class="px-4 py-2.5 bg-blue-600 text-white rounded-lg text-sm font-medium hover:bg-blue-700 flex items-center gap-2">
				<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" /></svg>
				کاربر جدید
			</button>
		</div>
	</div>

	<!-- Filters -->
	<div class="flex items-center gap-3 flex-wrap">
		<input type="text" bind:value={search} onkeydown={(e) => e.key === 'Enter' && searchUsers()} class="flex-1 min-w-[200px] px-4 py-2.5 border rounded-lg text-sm focus:ring-2 focus:ring-blue-500 outline-none bg-white" placeholder="جستجوی نام یا ایمیل..." />
		<select bind:value={roleFilter} onchange={() => { currentPage = 1; loadUsers(); }} class="px-4 py-2.5 border rounded-lg text-sm focus:ring-2 focus:ring-blue-500 outline-none bg-white">
			<option value="all">همه نقش‌ها</option>
			<option value="admin">مدیر</option>
			<option value="teacher">مدرس</option>
			<option value="student">دانش‌آموز</option>
		</select>
	</div>

	{#if loading}
		<div class="flex items-center justify-center py-12"><div class="animate-spin h-8 w-8 border-4 border-blue-600 border-t-transparent rounded-full"></div></div>
	{:else if users.length === 0}
		<div class="text-center py-20 bg-white rounded-xl border">
			<p class="text-gray-500">کاربری یافت نشد</p>
		</div>
	{:else}
		<div class="bg-white rounded-xl border overflow-hidden">
			<table class="w-full text-sm">
				<thead class="bg-gray-50 border-b">
					<tr>
						<th class="px-5 py-3 text-right font-medium text-gray-600">نام</th>
						<th class="px-5 py-3 text-right font-medium text-gray-600">ایمیل</th>
						<th class="px-5 py-3 text-right font-medium text-gray-600">نقش</th>
						<th class="px-5 py-3 text-right font-medium text-gray-600">وضعیت</th>
						<th class="px-5 py-3 text-right font-medium text-gray-600">عملیات</th>
					</tr>
				</thead>
				<tbody class="divide-y">
					{#each users as user}
						<tr class="hover:bg-gray-50">
							<td class="px-5 py-3 font-medium">{user.display_name}</td>
							<td class="px-5 py-3 text-gray-500" dir="ltr">{user.email}</td>
							<td class="px-5 py-3"><span class="text-xs px-2 py-1 rounded-full font-medium {roleColors[user.role]}">{roleLabels[user.role]}</span></td>
							<td class="px-5 py-3">
								<span class="text-xs px-2 py-1 rounded-full font-medium {user.is_active ? 'bg-green-100 text-green-700' : 'bg-red-100 text-red-700'}">
									{user.is_active ? 'فعال' : 'غیرفعال'}
								</span>
							</td>
							<td class="px-5 py-3">
								<div class="flex items-center gap-1">
									<button onclick={() => openEdit(user)} class="px-2 py-1 text-xs text-blue-600 hover:bg-blue-50 rounded">ویرایش</button>
									<button onclick={() => toggleActive(user)} class="px-2 py-1 text-xs rounded {user.is_active ? 'text-orange-600 hover:bg-orange-50' : 'text-green-600 hover:bg-green-50'}">
										{user.is_active ? 'غیرفعال' : 'فعال'}
									</button>
									<button onclick={() => confirmDelete(user.id)} class="px-2 py-1 text-xs text-red-600 hover:bg-red-50 rounded">حذف</button>
								</div>
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
			{#if total > perPage}
				<div class="px-5 py-3 border-t flex items-center justify-between text-sm text-gray-500">
					<span>{toPersian(total)} کاربر</span>
					<div class="flex gap-1">
						<button disabled={currentPage <= 1} onclick={() => { currentPage--; loadUsers(); }} class="px-3 py-1 border rounded hover:bg-gray-50 disabled:opacity-50">قبلی</button>
						<span class="px-3 py-1">صفحه {toPersian(currentPage)} از {toPersian(Math.ceil(total / perPage))}</span>
						<button disabled={currentPage >= Math.ceil(total / perPage)} onclick={() => { currentPage++; loadUsers(); }} class="px-3 py-1 border rounded hover:bg-gray-50 disabled:opacity-50">بعدی</button>
					</div>
				</div>
			{/if}
		</div>
	{/if}
</div>

<!-- Create User Modal -->
{#if showCreateModal}
	<div class="fixed inset-0 bg-black/40 z-50 flex items-center justify-center p-4" onclick={() => showCreateModal = false}>
		<div class="bg-white rounded-2xl w-full max-w-md shadow-xl" onclick={(e) => e.stopPropagation()}>
			<div class="px-6 py-4 border-b"><h2 class="font-bold text-lg">ایجاد کاربر جدید</h2></div>
			<div class="px-6 py-4 space-y-4">
				{#if createError}
					<div class="p-3 bg-red-50 text-red-600 rounded-lg text-sm">{createError}</div>
				{/if}
				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">نام نمایشی</label>
					<input type="text" bind:value={newUser.display_name} class="input-field" required />
				</div>
				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">ایمیل</label>
					<input type="email" bind:value={newUser.email} class="input-field" dir="ltr" required />
				</div>
				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">رمز عبور</label>
					<input type="password" bind:value={newUser.password} class="input-field" dir="ltr" required />
				</div>
				<div class="grid grid-cols-2 gap-3">
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">نقش</label>
						<select bind:value={newUser.role} class="input-field bg-white">
							<option value="student">دانش‌آموز</option>
							<option value="teacher">مدرس</option>
							<option value="admin">مدیر</option>
						</select>
					</div>
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">تلفن</label>
						<input type="tel" bind:value={newUser.phone} class="input-field" dir="ltr" />
					</div>
				</div>
			</div>
			<div class="px-6 py-4 border-t flex justify-end gap-3">
				<button onclick={() => showCreateModal = false} class="btn-ghost">انصراف</button>
				<button onclick={createUser} disabled={createLoading || !newUser.email || !newUser.password || !newUser.display_name} class="btn-primary disabled:opacity-50">
					{createLoading ? 'در حال ایجاد...' : 'ایجاد کاربر'}
				</button>
			</div>
		</div>
	</div>
{/if}

<!-- Edit User Modal -->
{#if showEditModal && editingUser}
	<div class="fixed inset-0 bg-black/40 z-50 flex items-center justify-center p-4" onclick={() => showEditModal = false}>
		<div class="bg-white rounded-2xl w-full max-w-md shadow-xl" onclick={(e) => e.stopPropagation()}>
			<div class="px-6 py-4 border-b"><h2 class="font-bold text-lg">ویرایش کاربر</h2></div>
			<div class="px-6 py-4 space-y-4">
				<p class="text-sm text-gray-500">{editingUser.display_name} — {editingUser.email}</p>
				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">نقش</label>
					<select bind:value={editForm.role} class="input-field bg-white">
						<option value="student">دانش‌آموز</option>
						<option value="teacher">مدرس</option>
						<option value="admin">مدیر</option>
					</select>
				</div>
				<div class="flex items-center justify-between">
					<span class="text-sm font-medium text-gray-700">وضعیت فعال</span>
					<button onclick={() => editForm.is_active = !editForm.is_active} class="relative inline-flex h-6 w-11 items-center rounded-full transition-colors {editForm.is_active ? 'bg-blue-600' : 'bg-gray-300'}">
						<span class="inline-block h-4 w-4 transform rounded-full bg-white transition-transform {editForm.is_active ? 'translate-x-6' : 'translate-x-1'}"></span>
					</button>
				</div>
			</div>
			<div class="px-6 py-4 border-t flex justify-end gap-3">
				<button onclick={() => showEditModal = false} class="btn-ghost">انصراف</button>
				<button onclick={saveEdit} class="btn-primary">ذخیره</button>
			</div>
		</div>
	</div>
{/if}

<ConfirmModal bind:show={showDeleteConfirm} title="حذف کاربر" message="آیا از حذف این کاربر اطمینان دارید؟" onConfirm={deleteUser} onCancel={() => {}} />

<!-- Bulk Import Modal -->
{#if showImportModal}
	<div class="fixed inset-0 bg-black/40 z-50 flex items-center justify-center p-4" onclick={closeImportModal}>
		<div class="bg-white rounded-2xl w-full max-w-2xl shadow-xl max-h-[90vh] overflow-y-auto" onclick={(e) => e.stopPropagation()}>
			<div class="px-6 py-4 border-b flex items-center justify-between">
				<h2 class="font-bold text-lg">وارد کردن گروهی کاربران</h2>
				<button onclick={closeImportModal} class="text-gray-400 hover:text-gray-600">
					<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" /></svg>
				</button>
			</div>

			<div class="px-6 py-4 space-y-4">
				<!-- Download Template -->
				<button onclick={downloadTemplate} class="text-sm text-blue-600 hover:text-blue-700 flex items-center gap-1">
					<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" /></svg>
					دانلود قالب CSV
				</button>

				<!-- File Upload Area -->
				{#if !importFile}
					<div
						onDragOver={handleDragOver}
						onDragLeave={handleDragLeave}
						onDrop={handleDrop}
						onclick={() => fileInput?.click()}
						class="border-2 border-dashed rounded-xl p-8 text-center cursor-pointer transition-colors {isDragging ? 'border-blue-500 bg-blue-50' : 'border-gray-300 hover:border-gray-400 bg-gray-50'}"
					>
						<svg class="w-10 h-10 mx-auto text-gray-400 mb-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" /></svg>
						<p class="text-sm text-gray-600 mb-1">فایل CSV را اینجا بکشید و رها کنید</p>
						<p class="text-xs text-gray-400">یا کلیک کنید برای انتخاب فایل</p>
						<input bind:this={fileInput} type="file" accept=".csv" onchange={handleFileSelect} class="hidden" />
					</div>
				{/if}

				<!-- Selected File & Preview -->
				{#if importFile && !importResult}
					<div class="space-y-3">
						<div class="flex items-center justify-between p-3 bg-gray-50 rounded-lg">
							<div class="flex items-center gap-2">
								<svg class="w-5 h-5 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" /></svg>
								<span class="text-sm font-medium">{importFile.name}</span>
								<span class="text-xs text-gray-400">({toPersianNum(Math.round(importFile.size / 1024))} کیلوبایت)</span>
							</div>
							<button onclick={() => { importFile = null; importPreview = []; }} class="text-gray-400 hover:text-red-500">
								<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" /></svg>
							</button>
						</div>

						{#if importPreview.length > 0}
							<div class="border rounded-lg overflow-hidden">
								<div class="bg-gray-50 px-4 py-2 border-b">
									<span class="text-xs font-medium text-gray-500">پیش‌نمایش (اولین {toPersianNum(Math.min(importPreview.length - 1, 5))} ردیف داده)</span>
								</div>
								<div class="overflow-x-auto">
									<table class="w-full text-xs">
										<thead>
											<tr class="bg-gray-100">
												{#each importPreview[0] as header}
													<th class="px-3 py-2 text-right font-medium text-gray-600">{header}</th>
												{/each}
											</tr>
										</thead>
										<tbody class="divide-y">
											{#each importPreview.slice(1) as row}
												<tr>
													{#each row as cell}
														<td class="px-3 py-2 text-gray-700">{cell}</td>
													{/each}
												</tr>
											{/each}
										</tbody>
									</table>
								</div>
							</div>
						{/if}
					</div>
				{/if}

				<!-- Import Results -->
				{#if importResult}
					<div class="space-y-3">
						{#if importResult.success_count > 0}
							<div class="p-4 bg-green-50 border border-green-200 rounded-lg flex items-center gap-3">
								<svg class="w-6 h-6 text-green-600 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
								<div>
									<p class="text-sm font-medium text-green-800">وارد کردن با موفقیت انجام شد</p>
									<p class="text-sm text-green-600">{toPersianNum(importResult.success_count)} کاربر جدید ایجاد شد</p>
								</div>
							</div>
						{/if}

						{#if importResult.errors.length > 0}
							<div class="border border-red-200 rounded-lg overflow-hidden">
								<div class="bg-red-50 px-4 py-2 border-b border-red-200">
									<span class="text-sm font-medium text-red-700">{toPersianNum(importResult.errors.length)} خطا</span>
								</div>
								<div class="max-h-48 overflow-y-auto divide-y divide-red-100">
									{#each importResult.errors as error}
										<div class="px-4 py-2 flex items-start gap-2 text-sm">
											<span class="text-red-500 font-mono shrink-0">ردیف {toPersianNum(error.row)}:</span>
											<span class="text-red-700">{error.message}</span>
										</div>
									{/each}
								</div>
							</div>
						{/if}
					</div>
				{/if}
			</div>

			<div class="px-6 py-4 border-t flex justify-end gap-3">
				<button onclick={closeImportModal} class="btn-ghost">
					{importResult ? 'انجام شد' : 'انصراف'}
				</button>
				{#if importFile && !importResult}
					<button onclick={runImport} disabled={importLoading} class="btn-primary disabled:opacity-50 flex items-center gap-2">
						{#if importLoading}
							<div class="animate-spin h-4 w-4 border-2 border-white border-t-transparent rounded-full"></div>
							در حال وارد کردن...
						{:else}
							<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12" /></svg>
							وارد کردن
						{/if}
					</button>
				{/if}
			</div>
		</div>
	</div>
{/if}
