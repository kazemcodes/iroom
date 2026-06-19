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

<div class="space-y-5">
	<div class="flex items-center justify-between">
		<div>
			<h1 class="sky-page-title">مدیریت کلاس‌ها</h1>
			<p class="sky-page-subtitle">ایجاد و مدیریت کلاس‌های آموزشی</p>
		</div>
		<button onclick={() => { showCreate = true; editingClass = null; resetForm(); }} class="sky-btn sky-btn-primary flex items-center gap-2">
			<svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
			کلاس جدید
		</button>
	</div>

	<!-- Search -->
	<div class="sky-search max-w-sm">
		<div class="sky-search-icon"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg></div>
		<input bind:value={searchQuery} class="sky-input" placeholder="جستجو در نام کلاس..." style="padding-right: 2.5rem;" />
	</div>

	<!-- Create/Edit Form -->
	{#if showCreate || editingClass}
		<div class="sky-card p-5">
			<h3 class="font-bold mb-4" style="color: var(--color-midnight-sky);">{editingClass ? 'ویرایش کلاس' : 'کلاس جدید'}</h3>
			<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
				<div>
					<label class="sky-label">نام کلاس</label>
					<input bind:value={formName} class="sky-input" />
				</div>
				<div>
					<label class="sky-label">حداکثر دانش‌آموز</label>
					<input type="number" bind:value={formMaxStudents} class="sky-input" />
				</div>
				<div class="md:col-span-2">
					<label class="sky-label">توضیحات</label>
					<textarea bind:value={formDescription} class="sky-input resize-none" rows="2"></textarea>
				</div>
				<div>
					<label class="sky-label">رنگ</label>
					<input type="color" bind:value={formColor} class="w-10 h-10 rounded-lg cursor-pointer" style="border: 1px solid var(--color-zen-garden);" />
				</div>
			</div>
			<div class="flex gap-2 mt-4">
				<button onclick={editingClass ? updateClass : createClass} class="sky-btn sky-btn-primary">{editingClass ? 'ذخیره' : 'ایجاد'}</button>
				<button onclick={() => { showCreate = false; editingClass = null; }} class="sky-btn sky-btn-secondary">انصراف</button>
			</div>
		</div>
	{/if}

	<!-- Classes List -->
	{#if loading}
		<div class="flex items-center justify-center py-16"><svg class="sky-spinner lg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="color: var(--color-crystal-clear);"><path d="M12 2v4M12 18v4M4.93 4.93l2.83 2.83M16.24 16.24l2.83 2.83M2 12h4M18 12h4M4.93 19.07l2.83-2.83M16.24 7.76l2.83-2.83"/></svg></div>
	{:else if filteredClasses.length === 0}
		<div class="sky-card"><div class="sky-empty"><p class="sky-empty-desc">کلاسی یافت نشد</p></div></div>
	{:else}
		<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
			{#each filteredClasses as c (c.id)}
				<div class="sky-card p-4">
					<div class="flex items-start justify-between">
						<div class="flex items-center gap-3">
							<div class="w-10 h-10 rounded-lg flex items-center justify-center text-white font-bold text-sm" style="background: {c.color || 'var(--color-crystal-clear)'}">
								{c.name.charAt(0)}
							</div>
							<div>
								<h3 class="font-bold text-sm" style="color: var(--color-midnight-sky);">{c.name}</h3>
								<p class="text-xs" style="color: var(--color-mystic-sea);">{toPersianNum(c.max_students || 0)} دانش‌آموز</p>
							</div>
						</div>
						<div class="flex gap-1">
							<button onclick={() => startEdit(c)} class="sky-btn-icon" style="width:32px;height:32px;" title="ویرایش">
								<svg width="16" height="16" fill="none" stroke="currentColor" stroke-width="1.75" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"/></svg>
							</button>
							<button onclick={() => deleteClass(c.id)} class="sky-btn-icon" style="width:32px;height:32px;" title="حذف">
								<svg width="16" height="16" fill="none" stroke="var(--color-fiery-passion)" stroke-width="1.75" viewBox="0 0 24 24"><polyline points="3 6 5 6 21 6"/><path d="M19 6l-1 14a2 2 0 01-2 2H8a2 2 0 01-2-2L5 6"/><path d="M10 11v6M14 11v6"/></svg>
							</button>
						</div>
					</div>
					{#if c.description}
						<p class="text-xs mt-2 line-clamp-2" style="color: var(--color-mystic-sea);">{c.description}</p>
					{/if}
					{#if c.invite_code}
						<p class="text-[10px] mt-2 font-mono" style="color: var(--color-moonlit-mist);">کد دعوت: {c.invite_code}</p>
					{/if}
				</div>
			{/each}
		</div>
	{/if}
</div>
