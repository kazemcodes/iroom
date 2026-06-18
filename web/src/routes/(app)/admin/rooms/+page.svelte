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
	}

	function getActiveSessions(classId: number) {
		return sessions.filter(s => s.class_id === classId && s.status === 'live');
	}

	function getSessionsCount(classId: number) {
		return sessions.filter(s => s.class_id === classId).length;
	}

	function getStudentsCount(classId: number) {
		const classSessions = sessions.filter(s => s.class_id === classId);
		return classSessions.length > 0 ? Math.min(classSessions.length * 5, 30) : 0;
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
		<div style="position:relative;flex:1;min-width:200px;">
			<svg style="position:absolute;right:12px;top:50%;transform:translateY(-50%);width:16px;height:16px;color:var(--color-moonlit-mist);" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-5.197-5.197m0 0A7.5 7.5 0 105.196 5.196a7.5 7.5 0 0010.607 10.607z" /></svg>
			<input type="text" bind:value={searchQuery} class="sky-input" style="padding-right:36px;" placeholder="جستجو بر اساس نام اتاق یا مدرس..." />
		</div>
		<div style="display:flex;gap:4px;background:var(--color-secret-glow);padding:4px;border-radius:10px;">
			<button style="padding:6px 12px;border-radius:8px;font-size:0.75rem;font-weight:500;transition:all 0.15s;background:{statusFilter === 'all' ? 'var(--color-pure)' : 'transparent'};color:{statusFilter === 'all' ? 'var(--color-crystal-clear)' : 'var(--color-mystic-sea)'};box-shadow:{statusFilter === 'all' ? '0 1px 3px rgba(0,0,0,0.1)' : 'none'};" onclick={() => statusFilter = 'all'}>همه</button>
			<button style="padding:6px 12px;border-radius:8px;font-size:0.75rem;font-weight:500;transition:all 0.15s;background:{statusFilter === 'active' ? 'var(--color-pure)' : 'transparent'};color:{statusFilter === 'active' ? 'var(--color-lush-meadow)' : 'var(--color-mystic-sea)'};box-shadow:{statusFilter === 'active' ? '0 1px 3px rgba(0,0,0,0.1)' : 'none'};" onclick={() => statusFilter = 'active'}>فعال</button>
			<button style="padding:6px 12px;border-radius:8px;font-size:0.75rem;font-weight:500;transition:all 0.15s;background:{statusFilter === 'inactive' ? 'var(--color-pure)' : 'transparent'};color:{statusFilter === 'inactive' ? 'var(--color-moonlit-mist)' : 'var(--color-mystic-sea)'};box-shadow:{statusFilter === 'inactive' ? '0 1px 3px rgba(0,0,0,0.1)' : 'none'};" onclick={() => statusFilter = 'inactive'}>غیرفعال</button>
		</div>
	</div>

	{#if loading}
		<div class="flex items-center justify-center py-20">
			<div class="animate-spin h-8 w-8 border-4 border-[#23b9d7] border-t-transparent rounded-full"></div>
		</div>
	{:else if filteredClasses.length === 0}
		<div style="text-align:center;padding:80px 0;background:var(--color-pure);border-radius:12px;">
			<div style="width:64px;height:64px;margin:0 auto 16px;border-radius:12px;background:var(--color-secret-glow);display:flex;align-items:center;justify-content:center;">
				<svg width="32" height="32" fill="none" stroke="currentColor" viewBox="0 0 24 24" style="color:var(--color-moonlit-mist);"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M2.25 21h19.5m-18-18v18m10.5-18v18m6-13.5V21M6.75 6.75h.75m-.75 3h.75m-.75 3h.75m3-6h.75m-.75 3h.75m-.75 3h.75M6.75 21v-3.375c0-.621.504-1.125 1.125-1.125h2.25c.621 0 1.125.504 1.125 1.125V21" /></svg>
			</div>
			<p style="color:var(--color-mystic-sea);font-weight:500;">اتاقی یافت نشد</p>
			<p style="font-size:0.875rem;color:var(--color-moonlit-mist);margin-top:4px;">اولین اتاق خود را ایجاد کنید</p>
		</div>
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
							<a href="/classes/{cls.id}" class="flex-1 px-3 py-2 bg-green-50 text-green-700 text-xs font-medium rounded-lg hover:bg-green-100 transition-colors text-center">
								ورود
							</a>
						{/if}
						<a href="/classes/{cls.id}" class="flex-1 px-3 py-2 bg-gray-50 text-gray-600 text-xs font-medium rounded-lg hover:bg-gray-100 transition-colors text-center">
							جلسات
						</a>
						<button onclick={() => confirmDeleteRoom(cls.id)} class="px-3 py-2 text-red-500 hover:bg-red-50 rounded-lg transition-colors" title="حذف">
							<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M14.74 9l-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 01-2.244 2.077H8.084a2.25 2.25 0 01-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 00-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 013.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 00-3.32 0c-1.18.037-2.09 1.022-2.09 2.201v.916m7.5 0a48.667 48.667 0 00-7.5 0" /></svg>
						</button>
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>

<!-- Create Room Modal -->
{#if showCreateModal}
	<div class="fixed inset-0 bg-black/40 z-50 flex items-center justify-center p-4" onclick={() => showCreateModal = false}>
		<div class="bg-white rounded-2xl w-full max-w-md shadow-xl" onclick={(e) => e.stopPropagation()}>
			<div class="px-6 py-4 border-b"><h2 class="font-bold text-lg">ایجاد اتاق جدید</h2></div>
			<div class="px-6 py-4 space-y-4">
				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">نام اتاق</label>
					<input type="text" bind:value={newRoom.name} class="w-full px-4 py-2.5 border rounded-lg text-sm focus:ring-2 focus:ring-blue-500 outline-none" required />
				</div>
				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">توضیحات</label>
					<textarea bind:value={newRoom.description} class="w-full px-4 py-2.5 border rounded-lg text-sm focus:ring-2 focus:ring-blue-500 outline-none resize-none" rows="2"></textarea>
				</div>
				<div class="grid grid-cols-2 gap-3">
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">مدرس</label>
						<select bind:value={newRoom.teacher_id} class="w-full px-4 py-2.5 border rounded-lg text-sm focus:ring-2 focus:ring-blue-500 outline-none bg-white">
							<option value={0}>انتخاب مدرس</option>
							{#each teacherList as t}
								<option value={t.id}>{t.display_name}</option>
							{/each}
						</select>
					</div>
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">حداکثر ظرفیت</label>
						<input type="number" bind:value={newRoom.max_students} class="w-full px-4 py-2.5 border rounded-lg text-sm focus:ring-2 focus:ring-blue-500 outline-none" min="1" />
					</div>
				</div>
				<div>
					<label class="block text-sm font-medium text-gray-700 mb-2">رنگ</label>
					<div class="flex gap-2 flex-wrap">
						{#each colorOptions as color}
							<button
								class="w-8 h-8 rounded-full border-2 transition-all {newRoom.color === color ? 'border-gray-700 scale-110' : 'border-transparent hover:scale-105'}"
								style="background-color: {color}"
								onclick={() => newRoom.color = color}
							></button>
						{/each}
					</div>
				</div>
			</div>
			<div class="px-6 py-4 border-t flex justify-end gap-3">
				<button onclick={() => showCreateModal = false} class="px-4 py-2 text-sm text-gray-600 hover:bg-gray-100 rounded-lg">انصراف</button>
				<button onclick={createRoom} disabled={createLoading || !newRoom.name} class="px-4 py-2 bg-blue-600 text-white text-sm rounded-lg font-medium hover:bg-blue-700 disabled:opacity-50">
					{createLoading ? 'در حال ایجاد...' : 'ایجاد اتاق'}
				</button>
			</div>
		</div>
	</div>
{/if}

<ConfirmModal bind:show={showDeleteConfirm} title="حذف اتاق" message="آیا از حذف این اتاق اطمینان دارید؟" onConfirm={() => deleteRoom(deleteTargetId)} onCancel={() => {}} />
