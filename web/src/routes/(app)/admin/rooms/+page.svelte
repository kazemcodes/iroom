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
			classes = [res.data, ...classes];
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
			<h1 class="text-2xl font-bold text-gray-900">مدیریت اتاق‌ها</h1>
			<p class="text-gray-500 mt-1">{toPersian(classes.length)} اتاق</p>
		</div>
		<button onclick={() => { showCreateModal = true; loadTeachersForModal(); }} class="px-4 py-2.5 bg-blue-600 text-white rounded-xl text-sm font-medium hover:bg-blue-700 flex items-center gap-2 shadow-sm shadow-blue-500/25">
			<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" /></svg>
			اتاق جدید
		</button>
	</div>

	<!-- Search & Filter -->
	<div class="flex items-center gap-3 flex-wrap">
		<div class="relative flex-1 min-w-[200px]">
			<svg class="absolute right-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-5.197-5.197m0 0A7.5 7.5 0 105.196 5.196a7.5 7.5 0 0010.607 10.607z" /></svg>
			<input type="text" bind:value={searchQuery} class="w-full pr-10 pl-4 py-2.5 border rounded-xl text-sm focus:ring-2 focus:ring-blue-500 outline-none bg-white" placeholder="جستجو بر اساس نام اتاق یا مدرس..." />
		</div>
		<div class="flex gap-1 bg-gray-100 p-1 rounded-xl">
			<button class="px-3 py-1.5 rounded-lg text-xs font-medium transition-all {statusFilter === 'all' ? 'bg-white text-blue-600 shadow-sm' : 'text-gray-500'}" onclick={() => statusFilter = 'all'}>همه</button>
			<button class="px-3 py-1.5 rounded-lg text-xs font-medium transition-all {statusFilter === 'active' ? 'bg-white text-green-600 shadow-sm' : 'text-gray-500'}" onclick={() => statusFilter = 'active'}>فعال</button>
			<button class="px-3 py-1.5 rounded-lg text-xs font-medium transition-all {statusFilter === 'inactive' ? 'bg-white text-gray-600 shadow-sm' : 'text-gray-500'}" onclick={() => statusFilter = 'inactive'}>غیرفعال</button>
		</div>
	</div>

	{#if loading}
		<div class="flex items-center justify-center py-20">
			<div class="animate-spin h-8 w-8 border-4 border-blue-600 border-t-transparent rounded-full"></div>
		</div>
	{:else if filteredClasses.length === 0}
		<div class="text-center py-20 bg-white rounded-2xl border">
			<div class="w-16 h-16 mx-auto mb-4 rounded-2xl bg-gray-100 flex items-center justify-center">
				<svg class="w-8 h-8 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M2.25 21h19.5m-18-18v18m10.5-18v18m6-13.5V21M6.75 6.75h.75m-.75 3h.75m-.75 3h.75m3-6h.75m-.75 3h.75m-.75 3h.75M6.75 21v-3.375c0-.621.504-1.125 1.125-1.125h2.25c.621 0 1.125.504 1.125 1.125V21" /></svg>
			</div>
			<p class="text-gray-500 font-medium">اتاقی یافت نشد</p>
			<p class="text-sm text-gray-400 mt-1">اولین اتاق خود را ایجاد کنید</p>
		</div>
	{:else}
		<div class="grid gap-4 sm:grid-cols-2">
			{#each filteredClasses as cls}
				{@const activeSessions = getActiveSessions(cls.id)}
				{@const sessionCount = getSessionsCount(cls.id)}
				{@const studentCount = getStudentsCount(cls.id)}
				{@const active = isRoomActive(cls.id)}
				<div class="bg-white border border-gray-200 rounded-2xl p-5 hover:shadow-md transition-all group">
					<div class="flex items-start justify-between mb-4">
						<div class="flex items-center gap-3">
							<div class="relative">
								<div class="w-12 h-12 rounded-xl flex items-center justify-center text-white font-bold text-lg" style="background: {cls.color || 'linear-gradient(135deg, #1a56db, #7c3aed)'}">
									{cls.name.charAt(0)}
								</div>
								{#if active}
									<span class="absolute -top-1 -left-1 flex h-3 w-3">
										<span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-green-400 opacity-75"></span>
										<span class="relative inline-flex rounded-full h-3 w-3 bg-green-500 border-2 border-white"></span>
									</span>
								{/if}
							</div>
							<div>
								<h3 class="font-bold text-gray-900">{cls.name}</h3>
								<p class="text-xs text-gray-500 mt-0.5">{teachers[cls.teacher_id]?.display_name || 'بدون مدرس'}</p>
							</div>
						</div>
						<span class="text-xs px-2.5 py-1 rounded-full font-semibold {active ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-500'}">
							{active ? 'فعال' : 'غیرفعال'}
						</span>
					</div>

					<div class="flex items-center gap-4 text-xs text-gray-500 mb-4">
						<div class="flex items-center gap-1.5">
							<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M15 19.128a9.38 9.38 0 002.625.372 9.337 9.337 0 004.121-.952 4.125 4.125 0 00-7.533-2.493M15 19.128v-.003c0-1.113-.285-2.16-.786-3.07M15 19.128v.106A12.318 12.318 0 018.624 21c-2.331 0-4.512-.645-6.374-1.766l-.001-.109a6.375 6.375 0 0111.964-3.07M12 6.375a3.375 3.375 0 11-6.75 0 3.375 3.375 0 016.75 0zm8.25 2.25a2.625 2.625 0 11-5.25 0 2.625 2.625 0 015.25 0z" /></svg>
							<span>{toPersian(studentCount)} دانش‌آموز</span>
						</div>
						<div class="flex items-center gap-1.5">
							<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M6.75 3v2.25M17.25 3v2.25M3 18.75V7.5a2.25 2.25 0 012.25-2.25h13.5A2.25 2.25 0 0121 7.5v11.25m-18 0A2.25 2.25 0 005.25 21h13.5A2.25 2.25 0 0021 18.75m-18 0v-7.5A2.25 2.25 0 015.25 9h13.5A2.25 2.25 0 0121 11.25v7.5" /></svg>
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
								class="w-8 h-8 rounded-full border-2 transition-all {newRoom.color === color ? 'border-gray-900 scale-110' : 'border-transparent hover:scale-105'}"
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
