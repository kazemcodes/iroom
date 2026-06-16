<script lang="ts">
	import { api } from '$lib/api';
	import { onMount } from 'svelte';
	import type { User, Class, Session, DashboardStats } from '$lib/types';
	import ConfirmModal from '$lib/components/ConfirmModal.svelte';

	import { goto } from '$app/navigation';
	let activeTab = $state<'users' | 'classes' | 'sessions'>('users');

	// Users
	let users = $state<User[]>([]);
	let userTotal = $state(0);
	let userPage = $state(1);
	let userSearch = $state('');
	let userLoading = $state(true);

	// Classes
	let classes = $state<Class[]>([]);
	let classTotal = $state(0);
	let classPage = $state(1);
	let classSearch = $state('');
	let classLoading = $state(true);

	// Sessions
	let sessions = $state<Session[]>([]);
	let sessionTotal = $state(0);
	let sessionPage = $state(1);
	let sessionSearch = $state('');
	let sessionLoading = $state(true);

	// Create user
	let showCreateUser = $state(false);
	let newUser = $state({ email: '', password: '', display_name: '', phone: '', role: 'student' });
	let createUserLoading = $state(false);
	let createUserError = $state('');

	// Create class
	let showCreateClass = $state(false);
	let newClass = $state({ name: '', description: '', color: '#3B82F6', max_students: 30, teacher_id: 0 });
	let teachers = $state<User[]>([]);

	// Delete confirm
	let showDeleteUserConfirm = $state(false);
	let deleteUserId = $state(0);

	const perPage = 15;

	onMount(async () => {
		await Promise.all([loadDashboard(), loadUsers(), loadClasses(), loadSessions()]);
	});

	async function loadDashboard() {
		loading = true;
		const [statsRes, classesRes, sessionsRes, usersRes, recordingsRes] = await Promise.all([
			api.get<DashboardStats>('/admin/stats'),
			api.get<Class[]>('/classes'),
			api.get<Session[]>('/sessions'),
			api.get<{ items: User[] }>('/admin/users', { per_page: '1' }),
			api.get<{ total?: number }>('/admin/recordings', { per_page: '1' })
		]);

		if (statsRes.success && statsRes.data) {
			stats = statsRes.data;
		} else {
			let uCount = 0;
			if (usersRes.success && usersRes.data) {
				uCount = (usersRes.data as any).total || 0;
			}
			const cArr = classesRes.success && classesRes.data ? (Array.isArray(classesRes.data) ? classesRes.data : []) : [];
			const sArr = sessionsRes.success && sessionsRes.data ? (Array.isArray(sessionsRes.data) ? sessionsRes.data : []) : [];
			stats = { users: uCount || cArr.length, classes: cArr.length, sessions: sArr.length, messages: 0 };
		}

		const classesArr = classesRes.success && classesRes.data ? (Array.isArray(classesRes.data) ? classesRes.data : []) : [];
		const sessionsArr = sessionsRes.success && sessionsRes.data ? (Array.isArray(sessionsRes.data) ? sessionsRes.data : []) : [];
		const usersArr = usersRes.success && usersRes.data ? ((usersRes.data as any).items || []) : [];
		const teachersMap: Record<number, string> = {};
		usersArr.forEach((u: User) => { teachersMap[u.id] = u.display_name; });

		const today = new Date().toISOString().slice(0, 10);
		todaySessions = sessionsArr.filter((s: Session) => s.scheduled_at && s.scheduled_at.startsWith(today)).length;

		if (recordingsRes.success && recordingsRes.data) {
			recordings = (recordingsRes.data as any).total || 0;
		}

		liveRooms = classesArr.map((cls: Class) => ({
			...cls,
			activeSessions: sessionsArr.filter((s: Session) => s.class_id === cls.id && s.status === 'live').length,
			teacherName: teachersMap[cls.teacher_id] || '—'
		})).filter((r: any) => r.activeSessions > 0).slice(0, 8);

		const logsRes = await api.get<{ items: ActivityLog[] }>('/admin/activity-logs', { per_page: '10' });
		if (logsRes.success && logsRes.data) {
			activityLogs = (logsRes.data as any).items || (Array.isArray(logsRes.data) ? logsRes.data : []);
		}
		loading = false;
	}

	async function loadUsers() {
		userLoading = true;
		const params: Record<string, string> = { page: String(userPage), per_page: String(perPage) };
		if (userSearch) params.search = userSearch;
		const res = await api.get<{ items: User[]; total: number }>('/admin/users', params);
		if (res.success && res.data) {
			users = res.data.items || [];
			userTotal = res.data.total;
		}
		userLoading = false;
	}

	async function loadClasses() {
		classLoading = true;
		const params: Record<string, string> = { page: String(classPage), per_page: String(perPage) };
		if (classSearch) params.search = classSearch;
		const res = await api.get<{ items: Class[]; total: number }>('/admin/classes', params);
		if (res.success && res.data) {
			classes = res.data.items || [];
			classTotal = res.data.total;
		}
		classLoading = false;
	}

	async function loadSessions() {
		sessionLoading = true;
		const params: Record<string, string> = { page: String(sessionPage), per_page: String(perPage) };
		if (sessionSearch) params.search = sessionSearch;
		const res = await api.get<{ items: Session[]; total: number }>('/admin/sessions', params);
		if (res.success && res.data) {
			sessions = res.data.items || [];
			sessionTotal = res.data.total;
		}
		sessionLoading = false;
	}

	async function loadTeachers() {
		const res = await api.get<{ items: User[] }>('/admin/users', { per_page: '100' });
		if (res.success && res.data) {
			teachers = (res.data.items || []).filter(u => u.role === 'teacher' || u.role === 'admin');
		}
	}

	async function createUser() {
		createUserLoading = true;
		createUserError = '';
		const res = await api.post('/admin/users', newUser);
		if (!res.success) {
			createUserError = res.error || 'خطا';
			createUserLoading = false;
			return;
		}
		users = [res.data!, ...users];
		showCreateUser = false;
		newUser = { email: '', password: '', display_name: '', phone: '', role: 'student' };
		createUserLoading = false;
	}

	async function toggleUserActive(user: User) {
		const res = await api.put(`/admin/users/${user.id}`, { is_active: !user.is_active });
		if (res.success) {
			users = users.map(u => u.id === user.id ? { ...u, is_active: !u.is_active } : u);
		}
	}

	async function deleteUser(id: number) {
		const res = await api.delete(`/admin/users/${id}`);
		if (res.success) {
			users = users.map(u => u.id === id ? { ...u, is_active: false } : u);
		}
	}

	function confirmDeleteUser(id: number) {
		deleteUserId = id;
		showDeleteUserConfirm = true;
	}

	async function createClass() {
		const res = await api.post('/admin/classes', newClass);
		if (res.success) {
			classes = [res.data!, ...classes];
			showCreateClass = false;
			newClass = { name: '', description: '', color: '#3B82F6', max_students: 30, teacher_id: 0 };
		}
	}

	function formatDate(d: string) {
		if (!d) return '';
		return new Date(d).toLocaleDateString('fa-IR', { year: 'numeric', month: 'short', day: 'numeric' });
	}

	function formatTime(d: string) {
		if (!d) return '';
		return new Date(d).toLocaleTimeString('fa-IR', { hour: '2-digit', minute: '2-digit' });
	}

	function actionLabel(action: string): string {
		const map: Record<string, string> = {
			'user_login': 'ورود کاربر',
			'user_created': 'ایجاد کاربر',
			'session_started': 'شروع جلسه',
			'session_ended': 'پایان جلسه',
			'class_created': 'ایجاد کلاس',
			'file_uploaded': 'آپلود فایل',
		};
		return map[action] || action;
	}

	function actionColor(action: string): string {
		const map: Record<string, string> = {
			'user_login': 'bg-blue-500',
			'user_created': 'bg-green-500',
			'session_started': 'bg-emerald-500',
			'session_ended': 'bg-gray-400',
			'class_created': 'bg-purple-500',
			'file_uploaded': 'bg-amber-500',
		};
		return map[action] || 'bg-gray-400';
	}

	const roleLabels: Record<string, string> = { admin: 'مدیر', teacher: 'مدرس', student: 'دانش‌آموز' };
	const roleColors: Record<string, string> = { admin: 'bg-red-100 text-red-700', teacher: 'bg-purple-100 text-purple-700', student: 'bg-blue-100 text-blue-700' };
</script>

<div class="space-y-6">
	<div>
		<h1 class="text-2xl font-bold text-gray-900">پنل مدیریت</h1>
		<p class="text-gray-500 mt-1">داشبورد مدیریت سیستم</p>
	</div>

	{#if loading}
		<div class="flex items-center justify-center py-20">
			<div class="animate-spin h-8 w-8 border-4 border-blue-600 border-t-transparent rounded-full"></div>
		</div>
	{:else}
		<!-- Stats Cards -->
		<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
			<div class="stat-card">
				<div class="flex items-center gap-4">
					<div class="w-12 h-12 rounded-xl flex items-center justify-center text-white" style="background: linear-gradient(135deg, #3b82f6, #1d4ed8);">
						<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M15 19.128a9.38 9.38 0 002.625.372 9.337 9.337 0 004.121-.952 4.125 4.125 0 00-7.533-2.493M15 19.128v-.003c0-1.113-.285-2.16-.786-3.07M15 19.128v.106A12.318 12.318 0 018.624 21c-2.331 0-4.512-.645-6.374-1.766l-.001-.109a6.375 6.375 0 0111.964-3.07M12 6.375a3.375 3.375 0 11-6.75 0 3.375 3.375 0 016.75 0zm8.25 2.25a2.625 2.625 0 11-5.25 0 2.625 2.625 0 015.25 0z" /></svg>
					</div>
					<div>
						<p class="text-sm text-gray-500">کاربران</p>
						<p class="text-2xl font-extrabold text-gray-900">{toPersian(stats.users)}</p>
					</div>
				</div>
			</div>

			<div class="stat-card">
				<div class="flex items-center gap-4">
					<div class="w-12 h-12 rounded-xl flex items-center justify-center text-white" style="background: linear-gradient(135deg, #10b981, #059669);">
						<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M2.25 21h19.5m-18-18v18m10.5-18v18m6-13.5V21M6.75 6.75h.75m-.75 3h.75m-.75 3h.75m3-6h.75m-.75 3h.75m-.75 3h.75M6.75 21v-3.375c0-.621.504-1.125 1.125-1.125h2.25c.621 0 1.125.504 1.125 1.125V21" /></svg>
					</div>
					<div>
						<p class="text-sm text-gray-500">اتاق‌های فعال</p>
						<p class="text-2xl font-extrabold text-gray-900">{toPersian(liveRooms.length)}</p>
					</div>
				</div>
			</div>

			<div class="stat-card">
				<div class="flex items-center gap-4">
					<div class="w-12 h-12 rounded-xl flex items-center justify-center text-white" style="background: linear-gradient(135deg, #f59e0b, #d97706);">
						<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M6.75 3v2.25M17.25 3v2.25M3 18.75V7.5a2.25 2.25 0 012.25-2.25h13.5A2.25 2.25 0 0121 7.5v11.25m-18 0A2.25 2.25 0 005.25 21h13.5A2.25 2.25 0 0021 18.75m-18 0v-7.5A2.25 2.25 0 015.25 9h13.5A2.25 2.25 0 0121 11.25v7.5" /></svg>
					</div>
					<div>
						<p class="text-sm text-gray-500">جلسات امروز</p>
						<p class="text-2xl font-extrabold text-gray-900">{toPersian(todaySessions)}</p>
					</div>
				</div>
			</div>

			<div class="stat-card">
				<div class="flex items-center gap-4">
					<div class="w-12 h-12 rounded-xl flex items-center justify-center text-white" style="background: linear-gradient(135deg, #8b5cf6, #7c3aed);">
						<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M15.75 10.5l4.72-4.72a.75.75 0 011.28.53v11.38a.75.75 0 01-1.28.53l-4.72-4.72M4.5 18.75h9a2.25 2.25 0 002.25-2.25v-9a2.25 2.25 0 00-2.25-2.25h-9A2.25 2.25 0 002.25 7.5v9a2.25 2.25 0 002.25 2.25z" /></svg>
					</div>
					<div>
						<p class="text-sm text-gray-500">ضبط‌ها</p>
						<p class="text-2xl font-extrabold text-gray-900">{toPersian(recordings)}</p>
					</div>
				</div>
			</div>
		</div>

		<!-- Live Rooms + Activity Feed -->
		<div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
			<!-- Live Rooms -->
			<div class="lg:col-span-2">
				<div class="flex items-center justify-between mb-4">
					<h2 class="font-bold text-gray-900">اتاق‌های فعال</h2>
					<button onclick={() => goto('/admin/rooms')} class="text-sm text-blue-600 hover:text-blue-700 font-medium">مشاهده همه</button>
				</div>
				{#if liveRooms.length === 0}
					<div class="bg-white rounded-xl border p-8 text-center">
						<div class="w-12 h-12 mx-auto mb-3 rounded-full bg-gray-100 flex items-center justify-center">
							<svg class="w-6 h-6 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" /></svg>
						</div>
						<p class="text-sm text-gray-500">اتاق فعالی وجود ندارد</p>
					</div>
				{:else}
					<div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
						{#each liveRooms as room}
							<div class="bg-white border rounded-xl p-4 hover:shadow-sm transition-shadow">
								<div class="flex items-start justify-between">
									<div class="flex items-center gap-3">
										<div class="w-10 h-10 rounded-lg flex items-center justify-center text-white font-bold text-sm" style="background: {room.color || '#3b82f6'}">
											{room.name.charAt(0)}
										</div>
										<div>
											<h3 class="font-bold text-sm text-gray-900">{room.name}</h3>
											<p class="text-xs text-gray-500">{room.teacherName}</p>
										</div>
									</div>
									<div class="flex items-center gap-1.5">
										<span class="relative flex h-2.5 w-2.5">
											<span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-green-400 opacity-75"></span>
											<span class="relative inline-flex rounded-full h-2.5 w-2.5 bg-green-500"></span>
										</span>
										<span class="text-xs font-medium text-green-600">{toPersian(room.activeSessions)} زنده</span>
									</div>
								</div>
							</div>
						{/each}
					</div>
				{/if}
			</div>

			<!-- Activity Feed -->
			<div>
				<h2 class="font-bold text-gray-900 mb-4">آخرین فعالیت‌ها</h2>
				<div class="bg-white rounded-xl border divide-y max-h-[320px] overflow-y-auto">
					{#if activityLogs.length === 0}
						<div class="p-6 text-center text-sm text-gray-400">فعالیتی ثبت نشده</div>
					{:else}
						{#each activityLogs as log}
							<div class="px-4 py-3 flex items-start gap-3">
								<div class="mt-1 w-2 h-2 rounded-full shrink-0 {actionColor(log.action)}"></div>
								<div class="flex-1 min-w-0">
									<p class="text-sm text-gray-700 truncate">{actionLabel(log.action)}</p>
									<p class="text-xs text-gray-400 mt-0.5">{formatTime(log.created_at)} · {formatDate(log.created_at)}</p>
								</div>
							</div>
						{/each}
					{/if}
				</div>
			</div>
		</div>
	{/if}

	<!-- Tabs -->
	<div class="flex items-center gap-3 flex-wrap">
		<div class="flex gap-1 bg-gray-100 p-1 rounded-lg w-fit">
			<button class="px-4 py-2 rounded-md text-sm font-medium transition-all {activeTab === 'users' ? 'bg-white text-blue-600 shadow-sm' : 'text-gray-500'}" onclick={() => activeTab = 'users'}>کاربران</button>
			<button class="px-4 py-2 rounded-md text-sm font-medium transition-all {activeTab === 'classes' ? 'bg-white text-blue-600 shadow-sm' : 'text-gray-500'}" onclick={() => activeTab = 'classes'}>کلاس‌ها</button>
			<button class="px-4 py-2 rounded-md text-sm font-medium transition-all {activeTab === 'sessions' ? 'bg-white text-blue-600 shadow-sm' : 'text-gray-500'}" onclick={() => activeTab = 'sessions'}>جلسات</button>
		</div>
	</div>

	{#if activeTab === 'users'}
		<div class="flex items-center justify-between gap-3">
			<input type="text" bind:value={userSearch} onkeydown={(e) => e.key === 'Enter' && (userPage = 1, loadUsers())} class="flex-1 px-4 py-2.5 border rounded-lg text-sm focus:ring-2 focus:ring-blue-500 outline-none bg-white" placeholder="جستجوی کاربر..." />
			<button onclick={() => { showCreateUser = true; }} class="px-4 py-2.5 bg-blue-600 text-white rounded-lg text-sm font-medium hover:bg-blue-700 flex items-center gap-2 shrink-0">
				<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" /></svg>
				کاربر جدید
			</button>
		</div>

		{#if userLoading}
			<div class="flex items-center justify-center py-12"><div class="animate-spin h-8 w-8 border-4 border-blue-600 border-t-transparent rounded-full"></div></div>
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
										<button onclick={() => toggleUserActive(user)} class="px-2 py-1 text-xs rounded {user.is_active ? 'text-orange-600 hover:bg-orange-50' : 'text-green-600 hover:bg-green-50'}">
											{user.is_active ? 'غیرفعال' : 'فعال'}
										</button>
									</div>
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
				{#if userTotal > perPage}
					<div class="px-5 py-3 border-t flex items-center justify-between text-sm text-gray-500">
						<span>{toPersian(userTotal)} کاربر</span>
						<div class="flex gap-1">
							<button disabled={userPage <= 1} onclick={() => { userPage--; loadUsers(); }} class="px-3 py-1 border rounded hover:bg-gray-50 disabled:opacity-50">قبلی</button>
							<span class="px-3 py-1">صفحه {toPersian(userPage)} از {toPersian(Math.ceil(userTotal / perPage))}</span>
							<button disabled={userPage >= Math.ceil(userTotal / perPage)} onclick={() => { userPage++; loadUsers(); }} class="px-3 py-1 border rounded hover:bg-gray-50 disabled:opacity-50">بعدی</button>
						</div>
					</div>
				{/if}
			</div>
		{/if}

	{:else if activeTab === 'classes'}
		<div class="flex items-center justify-between gap-3">
			<input type="text" bind:value={classSearch} onkeydown={(e) => e.key === 'Enter' && (classPage = 1, loadClasses())} class="flex-1 px-4 py-2.5 border rounded-lg text-sm focus:ring-2 focus:ring-blue-500 outline-none bg-white" placeholder="جستجوی کلاس..." />
			<button onclick={() => { showCreateClass = true; loadTeachers(); }} class="px-4 py-2.5 bg-blue-600 text-white rounded-lg text-sm font-medium hover:bg-blue-700 flex items-center gap-2 shrink-0">
				<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" /></svg>
				کلاس جدید
			</button>
		</div>

		{#if classLoading}
			<div class="flex items-center justify-center py-12"><div class="animate-spin h-8 w-8 border-4 border-blue-600 border-t-transparent rounded-full"></div></div>
		{:else}
			<div class="bg-white rounded-xl border overflow-hidden">
				<table class="w-full text-sm">
					<thead class="bg-gray-50 border-b">
						<tr>
							<th class="px-5 py-3 text-right font-medium text-gray-600">نام</th>
							<th class="px-5 py-3 text-right font-medium text-gray-600">توضیحات</th>
							<th class="px-5 py-3 text-right font-medium text-gray-600">حداکثر</th>
							<th class="px-5 py-3 text-right font-medium text-gray-600">تاریخ ایجاد</th>
						</tr>
					</thead>
					<tbody class="divide-y">
						{#each classes as cls}
							<tr class="hover:bg-gray-50">
								<td class="px-5 py-3 font-medium">
									<div class="flex items-center gap-2">
										<div class="w-3 h-3 rounded-full" style="background-color: {cls.color}"></div>
										{cls.name}
									</div>
								</td>
								<td class="px-5 py-3 text-gray-500 max-w-[200px] truncate">{cls.description || '-'}</td>
								<td class="px-5 py-3">{toPersian(cls.max_students)}</td>
								<td class="px-5 py-3 text-gray-500">{formatDate(cls.created_at)}</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		{/if}

	{:else if activeTab === 'sessions'}
		<div class="flex items-center justify-between gap-3">
			<input type="text" bind:value={sessionSearch} onkeydown={(e) => e.key === 'Enter' && (sessionPage = 1, loadSessions())} class="flex-1 px-4 py-2.5 border rounded-lg text-sm focus:ring-2 focus:ring-blue-500 outline-none bg-white" placeholder="جستجوی جلسه..." />
		</div>

		{#if sessionLoading}
			<div class="flex items-center justify-center py-12"><div class="animate-spin h-8 w-8 border-4 border-blue-600 border-t-transparent rounded-full"></div></div>
		{:else}
			<div class="bg-white rounded-xl border overflow-hidden">
				<table class="w-full text-sm">
					<thead class="bg-gray-50 border-b">
						<tr>
							<th class="px-5 py-3 text-right font-medium text-gray-600">عنوان</th>
							<th class="px-5 py-3 text-right font-medium text-gray-600">تاریخ</th>
							<th class="px-5 py-3 text-right font-medium text-gray-600">مدت</th>
							<th class="px-5 py-3 text-right font-medium text-gray-600">وضعیت</th>
						</tr>
					</thead>
					<tbody class="divide-y">
						{#each sessions as s}
							<tr class="hover:bg-gray-50">
								<td class="px-5 py-3 font-medium">{s.title}</td>
								<td class="px-5 py-3 text-gray-500">{formatDate(s.scheduled_at)}</td>
								<td class="px-5 py-3">{toPersian(s.duration)} دقیقه</td>
								<td class="px-5 py-3">
									<span class="text-xs px-2 py-1 rounded-full font-medium {s.status === 'live' ? 'bg-green-100 text-green-700' : s.status === 'scheduled' ? 'bg-blue-100 text-blue-700' : 'bg-gray-100 text-gray-500'}">
										{s.status === 'live' ? 'در حال برگزاری' : s.status === 'scheduled' ? 'برنامه‌ریزی شده' : 'پایان یافته'}
									</span>
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		{/if}
	{/if}
</div>

<!-- Create User Modal -->
{#if showCreateUser}
	<div class="fixed inset-0 bg-black/40 z-50 flex items-center justify-center p-4" onclick={() => showCreateUser = false}>
		<div class="bg-white rounded-2xl w-full max-w-md shadow-xl" onclick={(e) => e.stopPropagation()}>
			<div class="px-6 py-4 border-b"><h2 class="font-bold text-lg">ایجاد کاربر جدید</h2></div>
			<div class="px-6 py-4 space-y-4">
				{#if createUserError}
					<div class="p-3 bg-red-50 text-red-600 rounded-lg text-sm">{createUserError}</div>
				{/if}
				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">نام نمایشی</label>
					<input type="text" bind:value={newUser.display_name} class="w-full px-4 py-2.5 border rounded-lg text-sm focus:ring-2 focus:ring-blue-500 outline-none" required />
				</div>
				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">ایمیل</label>
					<input type="email" bind:value={newUser.email} class="w-full px-4 py-2.5 border rounded-lg text-sm focus:ring-2 focus:ring-blue-500 outline-none" dir="ltr" required />
				</div>
				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">رمز عبور</label>
					<input type="password" bind:value={newUser.password} class="w-full px-4 py-2.5 border rounded-lg text-sm focus:ring-2 focus:ring-blue-500 outline-none" dir="ltr" required />
				</div>
				<div class="grid grid-cols-2 gap-3">
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">نقش</label>
						<select bind:value={newUser.role} class="w-full px-4 py-2.5 border rounded-lg text-sm focus:ring-2 focus:ring-blue-500 outline-none bg-white">
							<option value="student">دانش‌آموز</option>
							<option value="teacher">مدرس</option>
							<option value="admin">مدیر</option>
						</select>
					</div>
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">تلفن</label>
						<input type="tel" bind:value={newUser.phone} class="w-full px-4 py-2.5 border rounded-lg text-sm focus:ring-2 focus:ring-blue-500 outline-none" dir="ltr" />
					</div>
				</div>
			</div>
			<div class="px-6 py-4 border-t flex justify-end gap-3">
				<button onclick={() => showCreateUser = false} class="px-4 py-2 text-sm text-gray-600 hover:bg-gray-100 rounded-lg">انصراف</button>
				<button onclick={createUser} disabled={createUserLoading || !newUser.email || !newUser.password || !newUser.display_name} class="px-4 py-2 bg-blue-600 text-white text-sm rounded-lg font-medium hover:bg-blue-700 disabled:opacity-50">
					{createUserLoading ? 'در حال ایجاد...' : 'ایجاد کاربر'}
				</button>
			</div>
		</div>
	</div>
{/if}

<!-- Create Class Modal -->
{#if showCreateClass}
	<div class="fixed inset-0 bg-black/40 z-50 flex items-center justify-center p-4" onclick={() => showCreateClass = false}>
		<div class="bg-white rounded-2xl w-full max-w-md shadow-xl" onclick={(e) => e.stopPropagation()}>
			<div class="px-6 py-4 border-b"><h2 class="font-bold text-lg">ایجاد کلاس جدید</h2></div>
			<div class="px-6 py-4 space-y-4">
				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">نام کلاس</label>
					<input type="text" bind:value={newClass.name} class="w-full px-4 py-2.5 border rounded-lg text-sm focus:ring-2 focus:ring-blue-500 outline-none" required />
				</div>
				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">توضیحات</label>
					<textarea bind:value={newClass.description} class="w-full px-4 py-2.5 border rounded-lg text-sm focus:ring-2 focus:ring-blue-500 outline-none resize-none" rows="2"></textarea>
				</div>
				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">حداکثر دانش‌آموز</label>
					<input type="number" bind:value={newClass.max_students} class="w-full px-4 py-2.5 border rounded-lg text-sm focus:ring-2 focus:ring-blue-500 outline-none" min="1" />
				</div>
			</div>
			<div class="px-6 py-4 border-t flex justify-end gap-3">
				<button onclick={() => showCreateClass = false} class="px-4 py-2 text-sm text-gray-600 hover:bg-gray-100 rounded-lg">انصراف</button>
				<button onclick={createClass} disabled={!newClass.name} class="px-4 py-2 bg-blue-600 text-white text-sm rounded-lg font-medium hover:bg-blue-700 disabled:opacity-50">ایجاد کلاس</button>
			</div>
		</div>
	</div>
{/if}

<ConfirmModal bind:show={showDeleteUserConfirm} title="غیرفعال‌سازی کاربر" message="آیا از غیرفعال‌سازی این کاربر اطمینان دارید؟" onConfirm={() => deleteUser(deleteUserId)} onCancel={() => {}} />
