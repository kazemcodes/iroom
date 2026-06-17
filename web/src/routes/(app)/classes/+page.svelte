<script lang="ts">
	import { auth, isAdmin, isTeacher } from '$lib/stores';
	import { api } from '$lib/api';
	import { onMount } from 'svelte';
	import type { Class } from '$lib/types';
	import { toPersianNum } from '$lib/utils/persian';

	let classes = $state<Class[]>([]);
	let loading = $state(true);
	let showCreate = $state(false);
	let search = $state('');

	let currentPage = $state(1);
	let totalClasses = $state(0);
	let perPage = $state(12);

	const perPageOptions = [6, 12, 24, 48];

	let formName = $state('');
	let formDesc = $state('');
	let formColor = $state('#3B82F6');
	let formMaxStudents = $state(30);
	let formLoading = $state(false);
	let formError = $state('');

	let showJoinByCode = $state(false);
	let joinCode = $state('');
	let joinLoading = $state(false);
	let joinError = $state('');

	const colors = ['#3B82F6', '#10B981', '#F59E0B', '#EF4444', '#8B5CF6', '#EC4899', '#06B6D4'];

	const totalPages = $derived(Math.ceil(totalClasses / perPage));

	function handlePerPageChange() {
		currentPage = 1;
		loadClasses();
	}

	onMount(() => loadClasses());

	async function loadClasses() {
		loading = true;
		const params: Record<string, string> = { page: String(currentPage), per_page: String(perPage) };
		if (search) params.search = search;
		const res = await api.get<{ items: Class[]; total: number }>('/classes', params);
		if (res.success && res.data) {
			classes = res.data.items || (Array.isArray(res.data) ? res.data : []);
			totalClasses = res.data.total || classes.length;
		}
		loading = false;
	}

	async function createClass() {
		formLoading = true;
		formError = '';
		const res = await api.post<Class>('/classes', {
			name: formName,
			description: formDesc,
			color: formColor,
			max_students: formMaxStudents
		});
		if (!res.success) {
			formError = res.error || 'خطا در ایجاد کلاس';
			formLoading = false;
			return;
		}
		classes = [res.data!, ...classes];
		totalClasses++;
		showCreate = false;
		formName = '';
		formDesc = '';
		formColor = '#3B82F6';
		formMaxStudents = 30;
		formLoading = false;
	}

	async function joinByCode() {
		if (!joinCode.trim()) {
			joinError = 'لطفاً کد دعوت را وارد کنید';
			return;
		}
		joinLoading = true;
		joinError = '';
		const res = await api.post<{ class_id: number }>(`/classes/join/${joinCode.trim()}`);
		if (!res.success) {
			joinError = res.error || 'خطا در پیوستن به کلاس';
			joinLoading = false;
			return;
		}
		if (res.data?.class_id) {
			window.location.href = `/classes/${res.data.class_id}`;
		}
		showJoinByCode = false;
		joinCode = '';
		joinLoading = false;
	}

	function handleSearch() {
		currentPage = 1;
		loadClasses();
	}
</script>

<div class="space-y-6">
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-bold" style="color: var(--sr-text);">کلاس‌ها</h1>
			<p style="color: var(--sr-text-secondary);">{toPersianNum(totalClasses)} کلاس</p>
		</div>
		<div class="flex items-center gap-3">
			<button onclick={() => showJoinByCode = true} class="btn-primary flex items-center gap-2" style="background: linear-gradient(135deg, #10b981, #059669);">
				<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M18 9v3m0 0v3m0-3h3m-3 0h-3m-2-5a4 4 0 11-8 0 4 4 0 018 0zM3 20a6 6 0 0112 0v1H3v-1z" />
				</svg>
				پیوستن با کد
			</button>
			{#if $isAdmin || $isTeacher}
				<button onclick={() => showCreate = true} class="btn-primary flex items-center gap-2">
					<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
					</svg>
					ایجاد کلاس
				</button>
			{/if}
		</div>
	</div>

	<div class="flex gap-3">
		<input
			type="text"
			bind:value={search}
			onkeydown={(e) => e.key === 'Enter' && handleSearch()}
			class="input-field flex-1"
			placeholder="جستجوی کلاس..."
		/>
		<button onclick={handleSearch} class="btn-ghost">جستجو</button>
	</div>

	{#if loading}
		<div class="flex items-center justify-center py-20">
			<div class="animate-spin h-8 w-8 border-4 border-blue-500 border-t-transparent rounded-full"></div>
		</div>
	{:else if classes.length === 0}
		<div class="text-center py-20 card">
			<svg class="w-16 h-16 mx-auto mb-4" style="color: var(--sr-text-secondary);" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1" d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253" />
			</svg>
			<p style="color: var(--sr-text-secondary);">کلاسی یافت نشد</p>
		</div>
	{:else}
		<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
			{#each classes as cls}
				<a href="/classes/{cls.id}" class="block card p-5 group">
					<div class="flex items-start justify-between mb-3">
						<div class="flex items-center gap-3">
							<div class="w-4 h-4 rounded-full shrink-0" style="background-color: {cls.color}"></div>
							<h3 class="font-bold group-hover:text-blue-400 transition-colors" style="color: var(--sr-text);">{cls.name}</h3>
						</div>
					</div>
					{#if cls.description}
						<p class="text-sm line-clamp-2 mb-4" style="color: var(--sr-text-secondary);">{cls.description}</p>
					{/if}
					<div class="flex items-center justify-between text-xs pt-3" style="color: var(--sr-text-secondary); border-top: 1px solid var(--sr-border);">
						<span class="flex items-center gap-1">
							<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z" />
							</svg>
							حداکثر {toPersianNum(cls.max_students)} نفر
						</span>
					</div>
				</a>
			{/each}
		</div>
	{/if}

	{#if totalPages > 1}
		<div class="flex items-center justify-between text-sm" style="color: var(--sr-text-secondary);">
			<div class="flex items-center gap-3">
				<span>{toPersianNum(totalClasses)} کلاس</span>
				<div class="flex items-center gap-1">
					<span>نمایش</span>
					<select bind:value={perPage} onchange={handlePerPageChange} class="input-field" style="width: auto; padding: 0.25rem 0.5rem;">
						{#each perPageOptions as opt}
							<option value={opt}>{toPersianNum(opt)}</option>
						{/each}
					</select>
					<span>در هر صفحه</span>
				</div>
			</div>
			<div class="flex gap-1">
				<button disabled={currentPage <= 1} onclick={() => { currentPage--; loadClasses(); }} class="btn-ghost" style="padding: 0.25rem 0.75rem;">قبلی</button>
				<span style="padding: 0.25rem 0.75rem;">صفحه {toPersianNum(currentPage)} از {toPersianNum(totalPages)}</span>
				<button disabled={currentPage >= totalPages} onclick={() => { currentPage++; loadClasses(); }} class="btn-ghost" style="padding: 0.25rem 0.75rem;">بعدی</button>
			</div>
		</div>
	{/if}
</div>

{#if showCreate}
	<div class="modal-overlay" onclick={() => showCreate = false}>
		<div class="modal-content" onclick={(e) => e.stopPropagation()}>
			<div class="px-6 py-4" style="border-bottom: 1px solid var(--sr-border);">
				<h2 class="font-bold text-lg" style="color: var(--sr-text);">ایجاد کلاس جدید</h2>
			</div>
			<div class="px-6 py-4 space-y-4">
				{#if formError}
					<div class="p-3 rounded-lg text-sm" style="background: rgba(224, 82, 82, 0.15); color: var(--sr-danger);">{formError}</div>
				{/if}
				<div>
					<label class="block text-sm font-medium mb-1" style="color: var(--sr-text-secondary);">نام کلاس</label>
					<input type="text" bind:value={formName} class="input-field" placeholder="مثال: ریاضی پایه دهم" required />
				</div>
				<div>
					<label class="block text-sm font-medium mb-1" style="color: var(--sr-text-secondary);">توضیحات</label>
					<textarea bind:value={formDesc} class="input-field resize-none" rows="2" placeholder="توضیحات کلاس..."></textarea>
				</div>
				<div>
					<label class="block text-sm font-medium mb-2" style="color: var(--sr-text-secondary);">رنگ</label>
					<div class="flex gap-2">
						{#each colors as color}
							<button
								class="w-8 h-8 rounded-full transition-transform {formColor === color ? 'ring-2 ring-offset-2 ring-blue-500 scale-110' : 'hover:scale-105'}"
								style="background-color: {color}; --tw-ring-offset-color: var(--sr-pure);"
								onclick={() => formColor = color}
							></button>
						{/each}
					</div>
				</div>
				<div>
					<label class="block text-sm font-medium mb-1" style="color: var(--sr-text-secondary);">حداکثر دانش‌آموز</label>
					<input type="number" bind:value={formMaxStudents} class="input-field" min="1" max="200" />
				</div>
			</div>
			<div class="px-6 py-4 flex justify-end gap-3" style="border-top: 1px solid var(--sr-border);">
				<button onclick={() => showCreate = false} class="btn-ghost">انصراف</button>
				<button onclick={createClass} disabled={formLoading || !formName} class="btn-primary disabled:opacity-50">
					{formLoading ? 'در حال ایجاد...' : 'ایجاد کلاس'}
				</button>
			</div>
		</div>
	</div>
{/if}

{#if showJoinByCode}
	<div class="modal-overlay" onclick={() => showJoinByCode = false}>
		<div class="modal-content" onclick={(e) => e.stopPropagation()}>
			<div class="px-6 py-4" style="border-bottom: 1px solid var(--sr-border);">
				<h2 class="font-bold text-lg" style="color: var(--sr-text);">پیوستن به کلاس</h2>
			</div>
			<div class="px-6 py-4 space-y-4">
				{#if joinError}
					<div class="p-3 rounded-lg text-sm" style="background: rgba(224, 82, 82, 0.15); color: var(--sr-danger);">{joinError}</div>
				{/if}
				<div>
					<label class="block text-sm font-medium mb-1" style="color: var(--sr-text-secondary);">کد دعوت</label>
					<input
						type="text"
						bind:value={joinCode}
						onkeydown={(e) => e.key === 'Enter' && joinByCode()}
						class="input-field text-lg text-center font-mono tracking-wider uppercase"
						placeholder="کد دعوت را وارد کنید"
						autofocus
					/>
					<p class="text-xs mt-2" style="color: var(--sr-text-secondary);">کد دعوت را از معلم دریافت کنید</p>
				</div>
			</div>
			<div class="px-6 py-4 flex justify-end gap-3" style="border-top: 1px solid var(--sr-border);">
				<button onclick={() => { showJoinByCode = false; joinCode = ''; joinError = ''; }} class="btn-ghost">انصراف</button>
				<button onclick={joinByCode} disabled={joinLoading || !joinCode.trim()} class="btn-primary disabled:opacity-50" style="background: linear-gradient(135deg, #10b981, #059669);">
					{joinLoading ? 'در حال پیوستن...' : 'پیوستن'}
				</button>
			</div>
		</div>
	</div>
{/if}
