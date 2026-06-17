<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib/api';
	import { toPersianNum } from '$lib/utils/persian';

	let classes = $state<any[]>([]);
	let loading = $state(true);
	let showCreate = $state(false);
	let editingClass = $state<any>(null);
	let searchQuery = $state('');

	let formName = $state('');
	let formDescription = $state('');
	let formColor = $state('#3B82F6');
	let formMaxStudents = $state(30);

	onMount(loadClasses);

	async function loadClasses() {
		loading = true;
		const res = await api.get('/classes');
		if (res.success && res.data) classes = Array.isArray(res.data) ? res.data : [];
		loading = false;
	}

	async function createClass() {
		const res = await api.post('/classes', {
			name: formName, description: formDescription, color: formColor, max_students: formMaxStudents
		});
		if (res.success) { showCreate = false; resetForm(); await loadClasses(); }
	}

	async function updateClass() {
		if (!editingClass) return;
		const res = await api.put(`/classes/${editingClass.id}`, {
			name: formName, description: formDescription, color: formColor, max_students: formMaxStudents
		});
		if (res.success) { editingClass = null; resetForm(); await loadClasses(); }
	}

	async function deleteClass(id: number) {
		if (!confirm('آیا از حذف این کلاس اطمینان دارید؟')) return;
		await api.delete(`/classes/${id}`);
		await loadClasses();
	}

	function startEdit(c: any) {
		editingClass = c;
		formName = c.name;
		formDescription = c.description || '';
		formColor = c.color || '#3B82F6';
		formMaxStudents = c.max_students || 30;
	}

	function resetForm() {
		formName = ''; formDescription = ''; formColor = '#3B82F6'; formMaxStudents = 30;
	}

	const filteredClasses = $derived(classes.filter(c => !searchQuery || c.name.includes(searchQuery)));
</script>

<div class="space-y-6">
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-bold text-gray-900">مدیریت کلاس‌ها</h1>
			<p class="text-gray-500 mt-1">ایجاد و مدیریت کلاس‌های آموزشی</p>
		</div>
		<button onclick={() => { showCreate = true; editingClass = null; resetForm(); }} class="px-4 py-2.5 bg-blue-600 text-white text-sm rounded-xl hover:bg-blue-700 transition-colors font-medium">
			+ کلاس جدید
		</button>
	</div>

	<!-- Search -->
	<div class="relative">
		<svg class="absolute right-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" /></svg>
		<input bind:value={searchQuery} class="w-full pr-10 pl-4 py-2.5 border border-gray-200 rounded-xl text-sm focus:ring-2 focus:ring-blue-500 outline-none bg-white" placeholder="جستجو در نام کلاس..." />
	</div>

	<!-- Create/Edit Form -->
	{#if showCreate || editingClass}
		<div class="bg-white rounded-xl p-5">
			<h3 class="font-bold text-gray-900 mb-4">{editingClass ? 'ویرایش کلاس' : 'کلاس جدید'}</h3>
			<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">نام کلاس</label>
					<input bind:value={formName} class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-blue-500 outline-none" />
				</div>
				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">حداکثر دانش‌آموز</label>
					<input type="number" bind:value={formMaxStudents} class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-blue-500 outline-none" />
				</div>
				<div class="md:col-span-2">
					<label class="block text-sm font-medium text-gray-700 mb-1">توضیحات</label>
					<textarea bind:value={formDescription} class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-blue-500 outline-none" rows="2"></textarea>
				</div>
				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">رنگ</label>
					<input type="color" bind:value={formColor} class="w-10 h-10 rounded-lg border cursor-pointer" />
				</div>
			</div>
			<div class="flex gap-2 mt-4">
				<button onclick={editingClass ? updateClass : createClass} class="px-4 py-2 bg-blue-600 text-white text-sm rounded-lg hover:bg-blue-700">
					{editingClass ? 'ذخیره' : 'ایجاد'}
				</button>
				<button onclick={() => { showCreate = false; editingClass = null; }} class="px-4 py-2 bg-gray-100 text-gray-700 text-sm rounded-lg hover:bg-gray-200">انصراف</button>
			</div>
		</div>
	{/if}

	<!-- Classes List -->
	{#if loading}
		<div class="flex items-center justify-center py-20">
			<div class="animate-spin h-8 w-8 border-4 border-blue-600 border-t-transparent rounded-full"></div>
		</div>
	{:else if filteredClasses.length === 0}
		<div class="text-center py-20 bg-white rounded-xl">
			<p class="text-gray-500">کلاسی یافت نشد</p>
		</div>
	{:else}
		<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
			{#each filteredClasses as c (c.id)}
				<div class="bg-white rounded-xl p-4 hover:shadow-md transition-all">
					<div class="flex items-start justify-between">
						<div class="flex items-center gap-3">
							<div class="w-10 h-10 rounded-lg flex items-center justify-center text-white font-bold text-sm" style="background: {c.color || '#3B82F6'}">
								{c.name.charAt(0)}
							</div>
							<div>
								<h3 class="font-bold text-gray-900 text-sm">{c.name}</h3>
								<p class="text-xs text-gray-500">{toPersianNum(c.max_students || 0)} دانش‌آموز</p>
							</div>
						</div>
						<div class="flex gap-1">
							<button onclick={() => startEdit(c)} class="p-1.5 text-gray-400 hover:text-blue-600 rounded-lg hover:bg-blue-50" title="ویرایش">
								<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" /></svg>
							</button>
							<button onclick={() => deleteClass(c.id)} class="p-1.5 text-gray-400 hover:text-red-600 rounded-lg hover:bg-red-50" title="حذف">
								<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" /></svg>
							</button>
						</div>
					</div>
					{#if c.description}
						<p class="text-xs text-gray-500 mt-2 line-clamp-2">{c.description}</p>
					{/if}
					{#if c.invite_code}
						<p class="text-[10px] text-gray-400 mt-2 font-mono">کد دعوت: {c.invite_code}</p>
					{/if}
				</div>
			{/each}
		</div>
	{/if}
</div>
