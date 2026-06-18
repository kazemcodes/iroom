<script lang="ts">
	import { api } from '$lib/api';
	import { onMount } from 'svelte';
	import type { Class, Session, User } from '$lib/types';
	import { goto } from '$app/navigation';
	import ConfirmModal from '$lib/components/ConfirmModal.svelte';

	const toPersian = (n: number) => String(n).replace(/[0-9]/g, d => '۰۱۲۳۴۵۶۷۸۹'[+d]);

	let classes = $state<Class[]>([]);
	let teachers = $state<Record<number, User>>({});
	let sessions = $state<Session[]>([]);
	let studentCounts = $state<Record<number, number>>({});
	let loading = $state(true);

	let searchQuery = $state('');
	let statusFilter = $state<'all' | 'active' | 'inactive'>('all');

	let showCreateModal = $state(false);
	let newRoom = $state({ name: '', description: '', color: '#3B82F6', max_students: 30, teacher_id: 0 });
	let createLoading = $state(false);
	let teacherList = $state<User[]>([]);

	let showDeleteConfirm = $state(false);
	let deleteTargetId = $state(0);

	onMount(() => loadRooms());

	async function loadRooms() {
		loading = true;
		const [classesRes, sessionsRes, usersRes] = await Promise.all([
			api.get<Class[]>('/classes'),
			api.get<Session[]>('/sessions'),
			api.get<{ items: User[] }>('/admin/users', { per_page: '1000' })
		]);

		if (classesRes.success && classesRes.data) {
			classes = Array.isArray(classesRes.data) ? classesRes.data : [];
		}
		if (sessionsRes.success && sessionsRes.data) {
			sessions = Array.isArray(sessionsRes.data) ? sessionsRes.data : [];
		}
		if (usersRes.success && usersRes.data) {
			const list = Array.isArray(usersRes.data.items) ? usersRes.data.items : [];
			list.forEach(u => { teachers[u.id] = u; });
		}
		loading = false;

		// Load actual student counts for each class
		for (const cls of classes) {
			const res = await api.get<{ items: User[] }>(`/classes/${cls.id}/students`);
			if (res.success && res.data) {
				const items = Array.isArray(res.data.items) ? res.data.items : (Array.isArray(res.data) ? res.data : []);
				studentCounts[cls.id] = items.length;
			}
		}
	}

	function getActiveSessions(classId: number) {
		return sessions.filter(s => s.class_id === classId && s.status === 'live');
	}

	function getSessionsCount(classId: number) {
		return sessions.filter(s => s.class_id === classId).length;
	}

	function getStudentsCount(classId: number) {
		// Return actual count from loaded data
		return studentCounts[classId] || 0;
	}

	function isRoomActive(classId: number) {
		return getActiveSessions(classId).length > 0;
	}

	function formatDate(d: string) {
		if (!d) return '';
		return new Date(d).toLocaleDateString('fa-IR', { year: 'numeric', month: 'long', day: 'numeric' });
	}

	let filteredClasses = $derived(classes.filter(cls => {
		const matchesSearch = !searchQuery ||
			cls.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
			(teachers[cls.teacher_id]?.display_name || '').toLowerCase().includes(searchQuery.toLowerCase());
		const matchesStatus = statusFilter === 'all' ||
			(statusFilter === 'active' && isRoomActive(cls.id)) ||
			(statusFilter === 'inactive' && !isRoomActive(cls.id));
		return matchesSearch && matchesStatus;
	}));

	async function loadTeachersForModal() {
		const res = await api.get<{ items: User[] }>('/admin/users', { per_page: '100' });
		if (res.success && res.data) {
			teacherList = (res.data.items || []).filter(u => u.role === 'teacher' || u.role === 'admin');
		}
	}

	async function createRoom() {
		createLoading = true;
		const res = await api.post('/admin/classes', newRoom);
		if (res.success && res.data) {
			classes = [res.data as Class, ...classes];
			showCreateModal = false;
			newRoom = { name: '', description: '', color: '#3B82F6', max_students: 30, teacher_id: 0 };
		}
		createLoading = false;
	}

	async function deleteRoom(id: number) {
		const res = await api.delete(`/admin/classes/${id}`);
		if (res.success) {
			classes = classes.filter(c => c.id !== id);
		}
	}

	function confirmDeleteRoom(id: number) {
		deleteTargetId = id;
		showDeleteConfirm = true;
	}

	const colorOptions = ['#3B82F6', '#10B981', '#F59E0B', '#EF4444', '#8B5CF6', '#EC4899', '#06B6D4', '#F97316'];
</script>

<div class="space-y-6">
	<div class="flex items-center justify-between">
		<div>
			<h1 style="font-size:1.5rem;font-weight:700;color:var(--color-midnight-sky);">مدیریت اتاق‌ها</h1>
			<p style="font-size:0.875rem;color:var(--color-mystic-sea);margin-top:4px;">{toPersian(classes.length)} اتاق</p>
		</div>
		<button onclick={() => { showCreateModal = true; loadTeachersForModal(); }} class="sky-btn sky-btn-primary">
			<svg width="16" height="16" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" /></svg>
			اتاق جدید
		</button>
	</div>

	<!-- Search & Filter -->
	<div class="flex items-center gap-3 flex-wrap">
		<div class="sky-search flex-1 min-w-[200px]">
			<div class="sky-search-icon"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg></div>
			<input type="text" bind:value={searchQuery} class="sky-input" style="padding-right: 2.5rem;" placeholder="جستجو بر اساس نام اتاق یا مدرس..." />
		</div>
		<div class="sky-filter-bar">
			<button class="sky-filter-btn {statusFilter === 'all' ? 'active' : ''}" onclick={() => statusFilter = 'all'}>همه</button>
			<button class="sky-filter-btn {statusFilter === 'active' ? 'active' : ''}" onclick={() => statusFilter = 'active'}>فعال</button>
			<button class="sky-filter-btn {statusFilter === 'inactive' ? 'active' : ''}" onclick={() => statusFilter = 'inactive'}>غیرفعال</button>
		</div>
	</div>

	{#if loading}
		<div class="flex items-center justify-center py-16"><svg class="sky-spinner lg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="color: var(--color-crystal-clear);"><path d="M12 2v4M12 18v4M4.93 4.93l2.83 2.83M16.24 16.24l2.83 2.83M2 12h4M18 12h4M4.93 19.07l2.83-2.83M16.24 7.76l2.83-2.83"/></svg></div>
	{:else if filteredClasses.length === 0}
		<div class="sky-card"><div class="sky-empty"><div class="sky-empty-icon"><svg width="48" height="48" fill="none" stroke="currentColor" stroke-width="1" viewBox="0 0 24 24" style="color: var(--color-muted-mountain);"><path d="M2 12s3-7 10-7 10 7 10 7-3 7-10 7-10-7-10-7z"/><circle cx="12" cy="12" r="3"/></svg></div><p class="sky-empty-title">اتاقی یافت نشد</p><p class="sky-empty-desc">اولین اتاق خود را ایجاد کنید</p></div></div>
	{:else}
		<div style="display:grid;gap:16px;grid-template-columns:repeat(auto-fill,minmax(320px,1fr));">
			{#each filteredClasses as cls}
				{@const activeSessions = getActiveSessions(cls.id)}
				{@const sessionCount = getSessionsCount(cls.id)}
				{@const studentCount = getStudentsCount(cls.id)}
				{@const active = isRoomActive(cls.id)}
				<div style="background:var(--color-pure);border:1px solid var(--color-zen-garden);border-radius:12px;padding:20px;transition:box-shadow 0.2s;">
					<div style="display:flex;align-items:flex-start;justify-content:space-between;margin-bottom:16px;">
						<div style="display:flex;align-items:center;gap:12px;">
							<div style="position:relative;">
								<div style="width:48px;height:48px;border-radius:12px;display:flex;align-items:center;justify-content:center;color:white;font-weight:700;font-size:18px;background:{cls.color || 'var(--color-crystal-clear)'};">
									{cls.name.charAt(0)}
								</div>
								{#if active}
									<span style="position:absolute;top:-4px;left:-4px;width:12px;height:12px;border-radius:50%;background:var(--color-lush-meadow);border:2px solid var(--color-pure);"></span>
								{/if}
							</div>
							<div>
								<h3 style="font-weight:700;color:var(--color-midnight-sky);">{cls.name}</h3>
								<p style="font-size:0.75rem;color:var(--color-mystic-sea);margin-top:2px;">{teachers[cls.teacher_id]?.display_name || 'بدون مدرس'}</p>
							</div>
						</div>
						<span style="font-size:0.75rem;padding:4px 10px;border-radius:20px;font-weight:600;background:{active ? 'rgba(64,191,127,0.1)' : 'var(--color-secret-glow)'};color:{active ? 'var(--color-lush-meadow)' : 'var(--color-moonlit-mist)'};">
							{active ? 'فعال' : 'غیرفعال'}
						</span>
					</div>

					<div style="display:flex;align-items:center;gap:16px;font-size:0.75rem;color:var(--color-mystic-sea);margin-bottom:16px;">
						<div style="display:flex;align-items:center;gap:6px;">
							<svg width="14" height="14" fill="none" stroke="currentColor" viewBox="0 0 24 24" style="color:var(--color-moonlit-mist);"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M15 19.128a9.38 9.38 0 002.625.372 9.337 9.337 0 004.121-.952 4.125 4.125 0 00-7.533-2.493M15 19.128v-.003c0-1.113-.285-2.16-.786-3.07M15 19.128v.106A12.318 12.318 0 018.624 21c-2.331 0-4.512-.645-6.374-1.766l-.001-.109a6.375 6.375 0 0111.964-3.07M12 6.375a3.375 3.375 0 11-6.75 0 3.375 3.375 0 016.75 0zm8.25 2.25a2.625 2.625 0 11-5.25 0 2.625 2.625 0 015.25 0z" /></svg>
							<span>{toPersian(studentCount)} دانش‌آموز</span>
						</div>
						<div style="display:flex;align-items:center;gap:6px;">
							<svg width="14" height="14" fill="none" stroke="currentColor" viewBox="0 0 24 24" style="color:var(--color-moonlit-mist);"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M6.75 3v2.25M17.25 3v2.25M3 18.75V7.5a2.25 2.25 0 012.25-2.25h13.5A2.25 2.25 0 0121 7.5v11.25m-18 0A2.25 2.25 0 005.25 21h13.5A2.25 2.25 0 0021 18.75m-18 0v-7.5A2.25 2.25 0 015.25 9h13.5A2.25 2.25 0 0121 11.25v7.5" /></svg>
							<span>{toPersian(sessionCount)} جلسه</span>
						</div>
						{#if activeSessions.length > 0}
							<div class="flex items-center gap-1.5 text-green-600 font-medium">
								<span class="w-1.5 h-1.5 rounded-full bg-green-500"></span>
								<span>{toPersian(activeSessions.length)} زنده</span>
							</div>
						{/if}
					</div>

					<div class="flex items-center gap-2">
						{#if active}
							<a href="/classes/{cls.id}" class="sky-btn flex-1" style="background: rgba(64,191,127,0.12); color: var(--color-lush-meadow); font-size: 12px; padding: 0.45rem;">ورود</a>
						{/if}
						<a href="/classes/{cls.id}" class="sky-btn sky-btn-secondary flex-1" style="font-size: 12px; padding: 0.45rem;">جلسات</a>
						<button onclick={() => confirmDeleteRoom(cls.id)} class="sky-btn-icon" style="width:34px;height:34px;" title="حذف">
							<svg width="16" height="16" fill="none" stroke="var(--color-fiery-passion)" stroke-width="1.75" viewBox="0 0 24 24"><polyline points="3 6 5 6 21 6"/><path d="M19 6l-1 14a2 2 0 01-2 2H8a2 2 0 01-2-2L5 6"/><path d="M10 11v6M14 11v6"/></svg>
						</button>
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>

<!-- Create Room Modal -->
{#if showCreateModal}
	<div class="modal-overlay" onclick={() => showCreateModal = false} role="button" tabindex="-1">
		<div class="modal-content" onclick={(e) => e.stopPropagation()}>
			<div class="sky-modal-header">
				<h2>ایجاد اتاق جدید</h2>
				<button onclick={() => showCreateModal = false} class="sky-btn-icon"><svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg></button>
			</div>
			<div class="sky-modal-body space-y-4">
				<div><label class="sky-label">نام اتاق</label><input type="text" bind:value={newRoom.name} class="sky-input" required /></div>
				<div><label class="sky-label">توضیحات</label><textarea bind:value={newRoom.description} class="sky-input resize-none" rows="2"></textarea></div>
				<div class="grid grid-cols-2 gap-3">
					<div>
						<label class="sky-label">مدرس</label>
						<select bind:value={newRoom.teacher_id} class="sky-input">
							<option value={0}>انتخاب مدرس</option>
							{#each teacherList as t}<option value={t.id}>{t.display_name}</option>{/each}
						</select>
					</div>
					<div><label class="sky-label">حداکثر ظرفیت</label><input type="number" bind:value={newRoom.max_students} class="sky-input" min="1" /></div>
				</div>
				<div>
					<label class="sky-label">رنگ</label>
					<div class="flex gap-2 flex-wrap">
						{#each colorOptions as color}
							<button type="button" class="w-8 h-8 rounded-full transition-all" style="background: {color}; outline: {newRoom.color === color ? '3px solid var(--color-zen-garden)' : 'none'}; outline-offset: 2px; transform: scale({newRoom.color === color ? 1.15 : 1});" onclick={() => newRoom.color = color}></button>
						{/each}
					</div>
				</div>
			</div>
			<div class="sky-modal-footer">
				<button onclick={() => showCreateModal = false} class="sky-btn sky-btn-secondary">انصراف</button>
				<button onclick={createRoom} disabled={createLoading || !newRoom.name} class="sky-btn sky-btn-primary">
					{createLoading ? 'در حال ایجاد...' : 'ایجاد اتاق'}
				</button>
			</div>
		</div>
	</div>
{/if}

<ConfirmModal bind:show={showDeleteConfirm} title="حذف اتاق" message="آیا از حذف این اتاق اطمینان دارید؟" onConfirm={() => deleteRoom(deleteTargetId)} onCancel={() => {}} />
