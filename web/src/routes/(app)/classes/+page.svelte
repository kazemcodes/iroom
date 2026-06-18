<script lang="ts">
	import { isAdmin, isTeacher } from '$lib/stores';
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

	let formName = $state('');
	let formDesc = $state('');
	let formColor = $state('#23b9d7');
	let formMaxStudents = $state(30);
	let formLoading = $state(false);
	let formError = $state('');

	let showJoinByCode = $state(false);
	let joinCode = $state('');
	let joinLoading = $state(false);
	let joinError = $state('');

	const colors = ['#23b9d7', '#3B82F6', '#10B981', '#F59E0B', '#EF4444', '#8B5CF6', '#EC4899'];
	const totalPages = $derived(Math.ceil(totalClasses / perPage));

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
		formLoading = true; formError = '';
		const res = await api.post<Class>('/classes', {
			name: formName, description: formDesc, color: formColor, max_students: formMaxStudents
		});
		if (!res.success) { formError = res.error || 'خطا در ایجاد کلاس'; formLoading = false; return; }
		classes = [res.data!, ...classes]; totalClasses++;
		showCreate = false; formName = ''; formDesc = ''; formColor = '#23b9d7'; formMaxStudents = 30;
		formLoading = false;
	}

	async function joinByCode() {
		if (!joinCode.trim()) { joinError = 'کد دعوت را وارد کنید'; return; }
		joinLoading = true; joinError = '';
		const res = await api.post<{ class_id: number }>(`/classes/join/${joinCode.trim()}`);
		if (!res.success) { joinError = res.error || 'خطا در پیوستن'; joinLoading = false; return; }
		if (res.data?.class_id) window.location.href = `/classes/${res.data.class_id}`;
		showJoinByCode = false; joinCode = ''; joinLoading = false;
	}

	function handleSearch() { currentPage = 1; loadClasses(); }
</script>

<div class="space-y-5">
	<!-- Page header -->
	<div class="flex items-center justify-between">
		<div>
			<h1 class="sky-page-title">کلاس‌ها</h1>
			<p class="sky-page-subtitle">{toPersianNum(totalClasses)} کلاس</p>
		</div>
		<div class="flex items-center gap-2">
			<button onclick={() => showJoinByCode = true} class="sky-btn sky-btn-outline flex items-center gap-2">
				<svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M15 3h4a2 2 0 012 2v14a2 2 0 01-2 2h-4"/><polyline points="10 17 15 12 10 7"/><line x1="15" y1="12" x2="3" y2="12"/></svg>
				پیوستن با کد
			</button>
			{#if $isAdmin || $isTeacher}
				<button onclick={() => showCreate = true} class="sky-btn sky-btn-primary flex items-center gap-2">
					<svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
					ایجاد کلاس
				</button>
			{/if}
		</div>
	</div>

	<!-- Search -->
	<div class="sky-search max-w-sm">
		<div class="sky-search-icon">
			<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg>
		</div>
		<input type="text" bind:value={search} onkeydown={(e) => e.key === 'Enter' && handleSearch()}
			class="sky-input" placeholder="جستجوی کلاس..." style="padding-right: 2.5rem;" />
	</div>

	<!-- Classes grid -->
	{#if loading}
		<div class="flex items-center justify-center py-16">
			<svg class="sky-spinner lg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="color: var(--color-crystal-clear);"><path d="M12 2v4M12 18v4M4.93 4.93l2.83 2.83M16.24 16.24l2.83 2.83M2 12h4M18 12h4M4.93 19.07l2.83-2.83M16.24 7.76l2.83-2.83"/></svg>
		</div>
	{:else if classes.length === 0}
		<div class="sky-card">
			<div class="sky-empty">
				<div class="sky-empty-icon">
					<svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1" style="color: var(--color-muted-mountain);"><path d="M2 3h6a4 4 0 014 4v14a3 3 0 00-3-3H2z"/><path d="M22 3h-6a4 4 0 00-4 4v14a3 3 0 013-3h7z"/></svg>
				</div>
				<p class="sky-empty-title">کلاسی یافت نشد</p>
				<p class="sky-empty-desc">اولین کلاس خود را بسازید</p>
				{#if $isAdmin || $isTeacher}
					<button onclick={() => showCreate = true} class="sky-btn sky-btn-primary">ایجاد کلاس</button>
				{/if}
			</div>
		</div>
	{:else}
		<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
			{#each classes as cls}
				<a href="/classes/{cls.id}" class="sky-card block p-5 group" style="text-decoration: none;">
					<div class="flex items-center gap-3 mb-3">
						<div class="w-10 h-10 rounded-xl flex items-center justify-center text-white font-bold text-sm shrink-0"
							style="background: {cls.color || 'var(--color-crystal-clear)'};">
							{cls.name.charAt(0)}
						</div>
						<div class="flex-1 min-w-0">
							<h3 class="font-bold text-sm truncate group-hover:underline" style="color: var(--color-midnight-sky);">{cls.name}</h3>
						</div>
					</div>
					{#if cls.description}
						<p class="text-xs line-clamp-2 mb-3" style="color: var(--color-mystic-sea);">{cls.description}</p>
					{/if}
					<div class="flex items-center justify-between text-xs pt-3" style="color: var(--color-moonlit-mist); border-top: 1px solid var(--color-zen-garden);">
						<span class="flex items-center gap-1">
							<svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.75"><path d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z"/></svg>
							حداکثر {toPersianNum(cls.max_students)} نفر
						</span>
						<svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="9 18 15 12 9 6"/></svg>
					</div>
				</a>
			{/each}
		</div>
	{/if}

	<!-- Pagination -->
	{#if totalPages > 1}
		<div class="flex items-center justify-between text-sm" style="color: var(--color-mystic-sea);">
			<span>{toPersianNum(totalClasses)} کلاس</span>
			<div class="sky-pagination">
				<button class="sky-page-btn" disabled={currentPage <= 1} onclick={() => { currentPage--; loadClasses(); }}>
					<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="9 18 15 12 9 6"/></svg>
				</button>
				{#each Array.from({length: totalPages}, (_, i) => i + 1) as p}
					<button class="sky-page-btn {currentPage === p ? 'active' : ''}" onclick={() => { currentPage = p; loadClasses(); }}>{toPersianNum(p)}</button>
				{/each}
				<button class="sky-page-btn" disabled={currentPage >= totalPages} onclick={() => { currentPage++; loadClasses(); }}>
					<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="15 18 9 12 15 6"/></svg>
				</button>
			</div>
			<span class="text-xs">صفحه {toPersianNum(currentPage)} از {toPersianNum(totalPages)}</span>
		</div>
	{/if}
</div>

<!-- Create Class Modal -->
{#if showCreate}
	<div class="modal-overlay" onclick={() => showCreate = false} role="button" tabindex="-1">
		<div class="modal-content" onclick={(e) => e.stopPropagation()}>
			<div class="sky-modal-header">
				<h2>ایجاد کلاس جدید</h2>
				<button onclick={() => showCreate = false} class="sky-btn-icon">
					<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
				</button>
			</div>
			<div class="sky-modal-body space-y-4">
				{#if formError}
					<div class="p-3 rounded-lg text-sm" style="background: rgba(224,82,82,0.1); color: var(--color-fiery-passion);">{formError}</div>
				{/if}
				<div>
					<label class="sky-label">نام کلاس</label>
					<input type="text" bind:value={formName} class="sky-input" placeholder="مثال: ریاضی پایه دهم" required />
				</div>
				<div>
					<label class="sky-label">توضیحات (اختیاری)</label>
					<textarea bind:value={formDesc} class="sky-input resize-none" rows="2" placeholder="توضیحات کلاس..."></textarea>
				</div>
				<div>
					<label class="sky-label">رنگ</label>
					<div class="flex gap-2 flex-wrap">
						{#each colors as color}
							<button type="button"
								class="w-8 h-8 rounded-full transition-all"
								style="background: {color}; outline: {formColor === color ? '3px solid var(--color-zen-garden)' : 'none'}; outline-offset: 2px; transform: scale({formColor === color ? 1.15 : 1});"
								onclick={() => formColor = color}
							></button>
						{/each}
					</div>
				</div>
				<div>
					<label class="sky-label">حداکثر دانش‌آموز</label>
					<input type="number" bind:value={formMaxStudents} class="sky-input" min="1" max="200" />
				</div>
			</div>
			<div class="sky-modal-footer">
				<button onclick={() => showCreate = false} class="sky-btn sky-btn-secondary">انصراف</button>
				<button onclick={createClass} disabled={formLoading || !formName} class="sky-btn sky-btn-primary">
					{formLoading ? 'در حال ایجاد...' : 'ایجاد کلاس'}
				</button>
			</div>
		</div>
	</div>
{/if}

<!-- Join by Code Modal -->
{#if showJoinByCode}
	<div class="modal-overlay" onclick={() => showJoinByCode = false} role="button" tabindex="-1">
		<div class="modal-content" onclick={(e) => e.stopPropagation()}>
			<div class="sky-modal-header">
				<h2>پیوستن به کلاس با کد</h2>
				<button onclick={() => { showJoinByCode = false; joinCode = ''; joinError = ''; }} class="sky-btn-icon">
					<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
				</button>
			</div>
			<div class="sky-modal-body space-y-4">
				{#if joinError}
					<div class="p-3 rounded-lg text-sm" style="background: rgba(224,82,82,0.1); color: var(--color-fiery-passion);">{joinError}</div>
				{/if}
				<div>
					<label class="sky-label">کد دعوت</label>
					<input type="text" bind:value={joinCode} onkeydown={(e) => e.key === 'Enter' && joinByCode()}
						class="sky-input text-center font-mono tracking-widest uppercase text-lg"
						placeholder="کد دعوت را وارد کنید" autofocus />
					<p class="text-xs mt-2" style="color: var(--color-moonlit-mist);">کد را از معلم یا مدیر کلاس دریافت کنید</p>
				</div>
			</div>
			<div class="sky-modal-footer">
				<button onclick={() => { showJoinByCode = false; joinCode = ''; joinError = ''; }} class="sky-btn sky-btn-secondary">انصراف</button>
				<button onclick={joinByCode} disabled={joinLoading || !joinCode.trim()} class="sky-btn sky-btn-primary">
					{joinLoading ? 'در حال پیوستن...' : 'پیوستن'}
				</button>
			</div>
		</div>
	</div>
{/if}
