<script lang="ts">
	import { auth, isAdmin, isTeacher } from '$lib/stores';
	import { api } from '$lib/api';
	import { onMount } from 'svelte';
	import type { Session } from '$lib/types';
	import ConfirmModal from '$lib/components/ConfirmModal.svelte';
	import { classroomWindow } from '$lib/classroom/ClassroomWindow';
	import { toPersianNum, toPersianDateTime } from '$lib/utils/persian';

	let sessions = $state<Session[]>([]);
	let loading = $state(true);
	let search = $state('');
	let filter = $state<'all' | 'scheduled' | 'live' | 'ended'>('all');

	let currentPage = $state(1);
	let totalSessions = $state(0);
	let perPage = $state(10);
	let perPageInitialized = $state(false);

	$effect(() => {
		if (!perPageInitialized) { perPageInitialized = true; return; }
		perPage; currentPage = 1; loadSessions();
	});

	let showDeleteConfirm = $state(false);
	let deleteTargetId = $state(0);
	let showCreate = $state(false);
	let newTitle = $state('');
	let newDuration = $state(60);
	let newScheduled = $state('');
	let newClassId = $state<number | ''>('');
	let classes = $state<any[]>([]);
	let createLoading = $state(false);
	let createError = $state('');

	const totalPages = $derived(Math.ceil(totalSessions / perPage));
	const filtered = $derived(filter === 'all' ? sessions : sessions.filter(s => s.status === filter));

	onMount(() => loadSessions());

	async function loadSessions() {
		loading = true;
		const params: Record<string, string> = { page: String(currentPage), per_page: String(perPage) };
		if (search) params.search = search;
		const res = await api.get<{ items: Session[]; total: number }>('/sessions', params);
		if (res.success && res.data) {
			sessions = res.data.items || (Array.isArray(res.data) ? res.data : []);
			totalSessions = res.data.total || sessions.length;
		}
		loading = false;
	}

	async function loadClasses() {
		const res = await api.get<any>('/classes');
		if (res.success && res.data) classes = Array.isArray(res.data) ? res.data : (res.data.items || []);
	}

	async function createSession() {
		if (!newTitle.trim() || !newScheduled) return;
		createLoading = true; createError = '';
		const body: any = { title: newTitle, duration: newDuration, scheduled_at: new Date(newScheduled).toISOString() };
		if (newClassId) body.class_id = newClassId;
		const res = await api.post<Session>('/sessions', body);
		if (!res.success) { createError = res.error || 'خطا در ایجاد جلسه'; createLoading = false; return; }
		sessions = [res.data!, ...sessions]; totalSessions++;
		showCreate = false; newTitle = ''; newDuration = 60; newScheduled = ''; newClassId = '';
		createLoading = false;
	}

	async function startSession(id: number) {
		const res = await api.post<any>(`/sessions/${id}/start`);
		if (res.success) sessions = sessions.map(s => s.id === id ? { ...s, status: 'live' } as any : s);
	}

	async function endSession(id: number) {
		const res = await api.post(`/sessions/${id}/end`);
		if (res.success) sessions = sessions.map(s => s.id === id ? { ...s, status: 'ended' } : s);
	}

	async function deleteSession(id: number) {
		const res = await api.delete(`/sessions/${id}`);
		if (res.success) { sessions = sessions.filter(s => s.id !== id); totalSessions--; }
	}

	function confirmDeleteSession(id: number) { deleteTargetId = id; showDeleteConfirm = true; }

	const statusLabel: Record<string, string> = { scheduled: 'برنامه‌ریزی شده', live: 'در حال برگزاری', ended: 'پایان یافته' };
	const statusBadge: Record<string, string> = { scheduled: 'sky-badge sky-badge-info', live: 'sky-badge sky-badge-success', ended: 'sky-badge sky-badge-default' };
</script>

<div class="space-y-5">
	<!-- Page header -->
	<div class="flex items-center justify-between">
		<div>
			<h1 class="sky-page-title">جلسات</h1>
			<p class="sky-page-subtitle">{toPersianNum(totalSessions)} جلسه</p>
		</div>
		{#if $isAdmin || $isTeacher}
			<button onclick={() => { showCreate = true; loadClasses(); }} class="sky-btn sky-btn-primary flex items-center gap-2">
				<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
				جلسه جدید
			</button>
		{/if}
	</div>

	<!-- Filters + Search -->
	<div class="flex flex-wrap items-center gap-3">
		<div class="sky-filter-bar">
			{#each [['all','همه'], ['scheduled','برنامه‌ریزی شده'], ['live','زنده'], ['ended','پایان یافته']] as [val, lbl]}
				<button class="sky-filter-btn {filter === val ? 'active' : ''}" onclick={() => filter = val as any}>{lbl}</button>
			{/each}
		</div>
		<div class="sky-search flex-1 min-w-[180px]">
			<div class="sky-search-icon">
				<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg>
			</div>
			<input type="text" bind:value={search} onkeydown={(e) => e.key === 'Enter' && loadSessions()}
				class="sky-input" placeholder="جستجو در جلسات..." style="padding-right: 2.5rem;" />
		</div>
	</div>

	<!-- Sessions Table -->
	<div class="sky-card">
		{#if loading}
			<div class="flex items-center justify-center py-16">
				<svg class="sky-spinner lg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="color: var(--color-crystal-clear);"><path d="M12 2v4M12 18v4M4.93 4.93l2.83 2.83M16.24 16.24l2.83 2.83M2 12h4M18 12h4M4.93 19.07l2.83-2.83M16.24 7.76l2.83-2.83"/></svg>
			</div>
		{:else if filtered.length === 0}
			<div class="sky-empty">
				<div class="sky-empty-icon">
					<svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1" style="color: var(--color-muted-mountain);"><path d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14"/><rect x="2" y="6" width="13" height="12" rx="2"/></svg>
				</div>
				<p class="sky-empty-title">جلسه‌ای یافت نشد</p>
				<p class="sky-empty-desc">جلسه جدیدی بسازید یا فیلتر را تغییر دهید</p>
			</div>
		{:else}
			<div class="overflow-x-auto">
				<table class="sky-table">
					<thead>
						<tr>
							<th>عنوان جلسه</th>
							<th>تاریخ</th>
							<th>مدت</th>
							<th>وضعیت</th>
							<th>عملیات</th>
						</tr>
					</thead>
					<tbody>
						{#each filtered as s}
							<tr>
								<td>
									<div class="flex items-center gap-3">
										<div class="w-9 h-9 rounded-lg flex items-center justify-center shrink-0"
											style="background: {s.status === 'live' ? 'rgba(64,191,127,0.12)' : 'var(--color-secret-glow)'};">
											{#if s.status === 'live'}
												<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.75" style="color: var(--color-lush-meadow);"><path d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14"/><rect x="2" y="6" width="13" height="12" rx="2"/></svg>
											{:else}
												<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.75" style="color: var(--color-moonlit-mist);"><path d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14"/><rect x="2" y="6" width="13" height="12" rx="2"/></svg>
											{/if}
										</div>
										<div>
											<p class="font-semibold text-sm" style="color: var(--color-midnight-sky);">{s.title}</p>
										</div>
									</div>
								</td>
								<td>
									<span class="text-sm" style="color: var(--color-mystic-sea);">{toPersianDateTime(s.scheduled_at)}</span>
								</td>
								<td>
									<span class="text-sm" style="color: var(--color-mystic-sea);">{toPersianNum(s.duration)} دقیقه</span>
								</td>
								<td>
									<span class="{statusBadge[s.status] || 'sky-badge sky-badge-default'}">
										{#if s.status === 'live'}<span class="dot"></span>{/if}
										{statusLabel[s.status]}
									</span>
								</td>
								<td>
									<div class="flex items-center gap-1">
										{#if s.status === 'live'}
											<button onclick={() => classroomWindow.open(String(s.id), s.title)} class="sky-btn sky-btn-primary" style="padding: 0.3rem 0.75rem; font-size: 12px;">پیوستن</button>
										{/if}
										{#if ($isAdmin || $isTeacher) && s.status === 'scheduled'}
											<button onclick={() => startSession(s.id)} class="sky-btn sky-btn-primary" style="padding: 0.3rem 0.75rem; font-size: 12px; background: var(--color-lush-meadow);">شروع</button>
										{/if}
										{#if ($isAdmin || $isTeacher) && s.status === 'live'}
											<button onclick={() => endSession(s.id)} class="sky-btn sky-btn-danger" style="padding: 0.3rem 0.75rem; font-size: 12px;">پایان</button>
										{/if}
										<a href="/sessions/{s.id}/logs" class="sky-btn sky-btn-ghost" style="padding: 0.3rem 0.75rem; font-size: 12px;">جزئیات</a>
										{#if $isAdmin}
											<button onclick={() => confirmDeleteSession(s.id)} class="sky-btn-icon" style="width:32px;height:32px;">
												<svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.75" style="color: var(--color-fiery-passion);"><polyline points="3 6 5 6 21 6"/><path d="M19 6l-1 14a2 2 0 01-2 2H8a2 2 0 01-2-2L5 6"/><path d="M10 11v6M14 11v6"/><path d="M9 6V4h6v2"/></svg>
											</button>
										{/if}
									</div>
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		{/if}
	</div>

	<!-- Pagination -->
	{#if totalPages > 1}
		<div class="flex items-center justify-between text-sm" style="color: var(--color-mystic-sea);">
			<div class="flex items-center gap-2">
				<span>{toPersianNum(totalSessions)} جلسه</span>
				<select bind:value={perPage} class="sky-input" style="width:auto;padding:0.25rem 1.5rem 0.25rem 0.5rem;font-size:12px;">
					{#each [10,25,50] as n}<option value={n}>{toPersianNum(n)}</option>{/each}
				</select>
			</div>
			<div class="sky-pagination">
				<button class="sky-page-btn" disabled={currentPage <= 1} onclick={() => { currentPage--; loadSessions(); }}>
					<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="9 18 15 12 9 6"/></svg>
				</button>
				{#each Array.from({length: Math.min(totalPages, 7)}, (_, i) => i + 1) as p}
					<button class="sky-page-btn {currentPage === p ? 'active' : ''}" onclick={() => { currentPage = p; loadSessions(); }}>{toPersianNum(p)}</button>
				{/each}
				<button class="sky-page-btn" disabled={currentPage >= totalPages} onclick={() => { currentPage++; loadSessions(); }}>
					<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="15 18 9 12 15 6"/></svg>
				</button>
			</div>
			<span class="text-xs">صفحه {toPersianNum(currentPage)} از {toPersianNum(totalPages)}</span>
		</div>
	{/if}
</div>

<!-- Create Session Modal -->
{#if showCreate}
	<div class="modal-overlay" onclick={() => showCreate = false} role="button" tabindex="-1">
		<div class="modal-content" onclick={(e) => e.stopPropagation()}>
			<div class="sky-modal-header">
				<h2>جلسه جدید</h2>
				<button onclick={() => showCreate = false} class="sky-btn-icon">
					<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
				</button>
			</div>
			<div class="sky-modal-body space-y-4">
				{#if createError}
					<div class="p-3 rounded-lg text-sm" style="background: rgba(224,82,82,0.1); color: var(--color-fiery-passion);">{createError}</div>
				{/if}
				<div>
					<label class="sky-label">عنوان جلسه</label>
					<input type="text" bind:value={newTitle} class="sky-input" placeholder="مثال: جلسه ریاضی — هفته اول" required />
				</div>
				<div class="grid grid-cols-2 gap-3">
					<div>
						<label class="sky-label">تاریخ و ساعت</label>
						<input type="datetime-local" bind:value={newScheduled} class="sky-input" required />
					</div>
					<div>
						<label class="sky-label">مدت (دقیقه)</label>
						<input type="number" bind:value={newDuration} class="sky-input" min="5" max="480" />
					</div>
				</div>
				{#if classes.length > 0}
					<div>
						<label class="sky-label">کلاس مرتبط (اختیاری)</label>
						<select bind:value={newClassId} class="sky-input">
							<option value="">— انتخاب کلاس —</option>
							{#each classes as cls}<option value={cls.id}>{cls.name}</option>{/each}
						</select>
					</div>
				{/if}
			</div>
			<div class="sky-modal-footer">
				<button onclick={() => showCreate = false} class="sky-btn sky-btn-secondary">انصراف</button>
				<button onclick={createSession} disabled={createLoading || !newTitle || !newScheduled} class="sky-btn sky-btn-primary">
					{createLoading ? 'در حال ایجاد...' : 'ایجاد جلسه'}
				</button>
			</div>
		</div>
	</div>
{/if}

<ConfirmModal bind:show={showDeleteConfirm} title="حذف جلسه" message="آیا از حذف این جلسه اطمینان دارید؟" onConfirm={() => deleteSession(deleteTargetId)} onCancel={() => {}} />
