<script lang="ts">
	import { api } from '$lib/api';
	import { onMount, onDestroy } from 'svelte';
	import type { User, Class, Session, DashboardStats, ActivityLog } from '$lib/types';
	import ConfirmModal from '$lib/components/ConfirmModal.svelte';

	let refreshInterval: ReturnType<typeof setInterval> | null = null;

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
		await loadHealth();
		// Refresh dashboard every 30 seconds
		refreshInterval = setInterval(async () => {
			await loadDashboard();
			await loadHealth();
		}, 30000);
	});

	onDestroy(() => {
		if (refreshInterval) clearInterval(refreshInterval);
	});

	async function loadDashboard() {
		loading = true;
		// Single API call for stats
		const statsRes = await api.get<DashboardStats>('/admin/stats');
		if (statsRes.success && statsRes.data) {
			stats = statsRes.data;
		}

		// Load activity logs (separate call needed for list)
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
		users = [res.data! as User, ...users];
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
			classes = [res.data! as Class, ...classes];
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

	// System Health
	let healthLoading = $state(true);
	let healthData = $state({
		serverUptime: '',
		dbSize: '',
		webrtcStatus: 'pion_builtin',
		activeRooms: 0,
		totalUsers: 0,
		totalSessions: 0,
		totalClasses: 0
	});

	const roleLabels: Record<string, string> = { admin: 'مدیر', teacher: 'مدرس', student: 'دانش‌آموز' };
	const roleColors: Record<string, string> = { admin: 'bg-red-100 text-red-700', teacher: 'bg-purple-100 text-purple-700', student: 'bg-blue-100 text-blue-700' };

	function toPersian(n: number): string {
		return n.toLocaleString('fa-IR');
	}

	let loading = $state(true);
	let stats = $state<DashboardStats>({ users: 0, classes: 0, sessions: 0, messages: 0 });
	let todaySessions = $state(0);
	let recordings = $state(0);
	let liveRooms = $state<any[]>([]);
	let activityLogs = $state<ActivityLog[]>([]);

	async function loadHealth() {
		healthLoading = true;
		const res = await api.get<any>('/health');
		if (res.success && res.data) {
			healthData = {
				serverUptime: res.data.uptime || res.data.server_uptime || '',
				dbSize: res.data.db_size || '',
				webrtcStatus: res.data.webrtc_status || 'pion_builtin',
				activeRooms: res.data.active_rooms ?? liveRooms.length,
				totalUsers: res.data.total_users ?? stats.users,
				totalSessions: res.data.total_sessions ?? stats.sessions,
				totalClasses: res.data.total_classes ?? stats.classes
			};
		} else {
			healthData = {
				...healthData,
				activeRooms: liveRooms.length,
				totalUsers: stats.users,
				totalSessions: stats.sessions,
				totalClasses: stats.classes
			};
		}
		healthLoading = false;
	}
</script>

<div class="space-y-6">
	<div>
		<h1 style="font-size:1.5rem;font-weight:700;color:var(--color-midnight-sky);">پنل مدیریت</h1>
		<p style="font-size:0.875rem;color:var(--color-mystic-sea);margin-top:4px;">داشبورد مدیریت سیستم</p>
	</div>

	{#if loading}
		<div class="flex items-center justify-center py-20">
			<div class="animate-spin h-8 w-8 border-4 border-[#23b9d7] border-t-transparent rounded-full"></div>
		</div>
	{:else}
		<!-- Stats Cards — Skyroom Style -->
		<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
			<div class="sky-stat-card">
				<div class="sky-stat-icon" style="background:linear-gradient(135deg,#23b9d7,#004ff2);">
					<svg width="24" height="24" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M15 19.128a9.38 9.38 0 002.625.372 9.337 9.337 0 004.121-.952 4.125 4.125 0 00-7.533-2.493M15 19.128v-.003c0-1.113-.285-2.16-.786-3.07M15 19.128v.106A12.318 12.318 0 018.624 21c-2.331 0-4.512-.645-6.374-1.766l-.001-.109a6.375 6.375 0 0111.964-3.07M12 6.375a3.375 3.375 0 11-6.75 0 3.375 3.375 0 016.75 0zm8.25 2.25a2.625 2.625 0 11-5.25 0 2.625 2.625 0 015.25 0z" /></svg>
				</div>
				<div>
					<p class="sky-stat-label">کاربران</p>
					<p class="sky-stat-value">{toPersian(stats.users)}</p>
				</div>
			</div>

			<div class="sky-stat-card">
				<div class="sky-stat-icon" style="background:linear-gradient(135deg,#40bf7f,#059669);">
					<svg width="24" height="24" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M2.25 21h19.5m-18-18v18m10.5-18v18m6-13.5V21M6.75 6.75h.75m-.75 3h.75m-.75 3h.75m3-6h.75m-.75 3h.75m-.75 3h.75M6.75 21v-3.375c0-.621.504-1.125 1.125-1.125h2.25c.621 0 1.125.504 1.125 1.125V21" /></svg>
				</div>
				<div>
					<p class="sky-stat-label">اتاق‌های فعال</p>
					<p class="sky-stat-value">{toPersian(liveRooms.length)}</p>
				</div>
			</div>

			<div class="sky-stat-card">
				<div class="sky-stat-icon" style="background:linear-gradient(135deg,#f59e0b,#d97706);">
					<svg width="24" height="24" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M6.75 3v2.25M17.25 3v2.25M3 18.75V7.5a2.25 2.25 0 012.25-2.25h13.5A2.25 2.25 0 0121 7.5v11.25m-18 0A2.25 2.25 0 005.25 21h13.5A2.25 2.25 0 0021 18.75m-18 0v-7.5A2.25 2.25 0 015.25 9h13.5A2.25 2.25 0 0121 11.25v7.5" /></svg>
				</div>
				<div>
					<p class="sky-stat-label">جلسات امروز</p>
					<p class="sky-stat-value">{toPersian(todaySessions)}</p>
				</div>
			</div>

			<div class="sky-stat-card">
				<div class="sky-stat-icon" style="background:linear-gradient(135deg,#8b5cf6,#7c3aed);">
					<svg width="24" height="24" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M15.75 10.5l4.72-4.72a.75.75 0 011.28.53v11.38a.75.75 0 01-1.28.53l-4.72-4.72M4.5 18.75h9a2.25 2.25 0 002.25-2.25v-9a2.25 2.25 0 00-2.25-2.25h-9A2.25 2.25 0 002.25 7.5v9a2.25 2.25 0 002.25 2.25z" /></svg>
				</div>
				<div>
					<p class="sky-stat-label">ضبط‌ها</p>
					<p class="sky-stat-value">{toPersian(recordings)}</p>
				</div>
			</div>
		</div>

		<!-- System Health -->
		<div class="sky-card">
			<div class="sky-card-header">
				<h2>وضعیت سیستم</h2>
			</div>
			<div class="sky-card-body">
				{#if healthLoading}
					<div class="flex items-center justify-center py-4">
						<svg class="sky-spinner md" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="color: var(--color-crystal-clear);"><path d="M12 2v4M12 18v4M4.93 4.93l2.83 2.83M16.24 16.24l2.83 2.83M2 12h4M18 12h4M4.93 19.07l2.83-2.83M16.24 7.76l2.83-2.83"/></svg>
					</div>
				{:else}
					<div class="grid grid-cols-2 sm:grid-cols-4 gap-4">
						{#each [
							{ label: 'آپتایم سرور', value: healthData.serverUptime || '—', color: 'var(--color-midnight-sky)' },
							{ label: 'WebRTC', value: 'فعال (Pion)', color: 'var(--color-lush-meadow)' },
							{ label: 'حجم پایگاه داده', value: healthData.dbSize || '—', color: 'var(--color-midnight-sky)' },
							{ label: 'اتاق‌های فعال', value: toPersian(healthData.activeRooms), color: 'var(--color-midnight-sky)' },
						] as item}
							<div class="text-center p-3 rounded-lg" style="background: var(--color-secret-glow);">
								<p class="text-xs mb-1" style="color: var(--color-moonlit-mist);">{item.label}</p>
								<p class="text-sm font-bold" style="color: {item.color};">{item.value}</p>
							</div>
						{/each}
					</div>
				{/if}
			</div>
		</div>

		<!-- Live Rooms + Activity Feed -->
		<div class="grid grid-cols-1 lg:grid-cols-3 gap-5">
			<!-- Live Rooms -->
			<div class="lg:col-span-2">
				<div class="sky-card">
					<div class="sky-card-header">
						<h2>اتاق‌های فعال</h2>
						<button onclick={() => goto('/admin/rooms')} class="sky-btn sky-btn-ghost" style="padding: 0.25rem 0.75rem; font-size: 12px;">مشاهده همه</button>
					</div>
					<div class="sky-card-body">
						{#if liveRooms.length === 0}
							<div class="sky-empty" style="padding: 2rem;">
								<p class="sky-empty-desc">اتاق فعالی وجود ندارد</p>
							</div>
						{:else}
							<div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
								{#each liveRooms as room}
									<div class="flex items-center justify-between p-3 rounded-xl" style="background: var(--color-secret-glow);">
										<div class="flex items-center gap-3">
											<div class="w-9 h-9 rounded-lg flex items-center justify-center text-white font-bold text-sm shrink-0" style="background: {room.color || 'var(--color-crystal-clear)'}">
												{room.name.charAt(0)}
											</div>
											<div>
												<p class="font-bold text-sm" style="color: var(--color-midnight-sky);">{room.name}</p>
												<p class="text-xs" style="color: var(--color-moonlit-mist);">{room.teacherName}</p>
											</div>
										</div>
										<span class="sky-live-dot">{toPersian(room.activeSessions)} زنده</span>
									</div>
								{/each}
							</div>
						{/if}
					</div>
				</div>
			</div>

			<!-- Activity Feed -->
			<div class="sky-card">
				<div class="sky-card-header"><h2>آخرین فعالیت‌ها</h2></div>
				<div class="overflow-y-auto" style="max-height: 320px;">
					{#if activityLogs.length === 0}
						<div class="sky-empty" style="padding: 2rem;">
							<p class="sky-empty-desc">فعالیتی ثبت نشده</p>
						</div>
					{:else}
						{#each activityLogs as log}
							<div class="px-4 py-3 flex items-start gap-3" style="border-bottom: 1px solid var(--color-zen-garden);">
								<div class="mt-1.5 w-2 h-2 rounded-full shrink-0 {actionColor(log.action)}"></div>
								<div class="flex-1 min-w-0">
									<p class="text-sm truncate" style="color: var(--color-midnight-sky);">{actionLabel(log.action)}</p>
									<p class="text-xs mt-0.5" style="color: var(--color-moonlit-mist);">{formatTime(log.created_at)} · {formatDate(log.created_at)}</p>
								</div>
							</div>
						{/each}
					{/if}
				</div>
			</div>
		</div>
	{/if}

	<!-- Tabs -->
	<div class="sky-filter-bar w-fit">
		<button class="sky-filter-btn {activeTab === 'users' ? 'active' : ''}" onclick={() => activeTab = 'users'}>کاربران</button>
		<button class="sky-filter-btn {activeTab === 'classes' ? 'active' : ''}" onclick={() => activeTab = 'classes'}>کلاس‌ها</button>
		<button class="sky-filter-btn {activeTab === 'sessions' ? 'active' : ''}" onclick={() => activeTab = 'sessions'}>جلسات</button>
	</div>

	{#if activeTab === 'users'}
		<div class="flex items-center gap-3">
			<div class="sky-search flex-1">
				<div class="sky-search-icon"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg></div>
				<input type="text" bind:value={userSearch} onkeydown={(e) => e.key === 'Enter' && (userPage = 1, loadUsers())} class="sky-input" placeholder="جستجوی کاربر..." style="padding-right: 2.5rem;" />
			</div>
			<button onclick={() => { showCreateUser = true; }} class="sky-btn sky-btn-primary flex items-center gap-2 shrink-0">
				<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
				کاربر جدید
			</button>
		</div>

		{#if userLoading}
			<div class="flex items-center justify-center py-12"><svg class="sky-spinner lg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="color: var(--color-crystal-clear);"><path d="M12 2v4M12 18v4M4.93 4.93l2.83 2.83M16.24 16.24l2.83 2.83M2 12h4M18 12h4M4.93 19.07l2.83-2.83M16.24 7.76l2.83-2.83"/></svg></div>
		{:else}
			<div class="sky-card overflow-hidden">
				<table class="sky-table">
					<thead>
						<tr>
							<th>نام</th><th>ایمیل</th><th>نقش</th><th>وضعیت</th><th>عملیات</th>
						</tr>
					</thead>
					<tbody>
						{#each users as user}
							<tr>
								<td class="font-semibold">{user.display_name}</td>
								<td dir="ltr" style="color: var(--color-mystic-sea);">{user.email}</td>
								<td><span class="sky-badge {user.role === 'admin' ? 'sky-badge-danger' : user.role === 'teacher' ? 'sky-badge-info' : 'sky-badge-default'}">{roleLabels[user.role]}</span></td>
								<td><span class="sky-badge {user.is_active ? 'sky-badge-success' : 'sky-badge-danger'}">{user.is_active ? 'فعال' : 'غیرفعال'}</span></td>
								<td>
									<button onclick={() => toggleUserActive(user)} class="sky-btn sky-btn-ghost" style="padding: 0.2rem 0.6rem; font-size: 12px; color: {user.is_active ? 'var(--color-dawn-warm)' : 'var(--color-lush-meadow)'};">
										{user.is_active ? 'غیرفعال‌سازی' : 'فعال‌سازی'}
									</button>
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
				{#if userTotal > perPage}
					<div class="px-5 py-3 flex items-center justify-between text-sm" style="border-top: 1px solid var(--color-zen-garden); color: var(--color-mystic-sea);">
						<span>{toPersian(userTotal)} کاربر</span>
						<div class="flex gap-1">
							<button disabled={userPage <= 1} onclick={() => { userPage--; loadUsers(); }} class="sky-page-btn">قبلی</button>
							<span class="sky-page-btn" style="cursor:default;">{toPersian(userPage)}/{toPersian(Math.ceil(userTotal / perPage))}</span>
							<button disabled={userPage >= Math.ceil(userTotal / perPage)} onclick={() => { userPage++; loadUsers(); }} class="sky-page-btn">بعدی</button>
						</div>
					</div>
				{/if}
			</div>
		{/if}

	{:else if activeTab === 'classes'}
		<div class="flex items-center gap-3">
			<div class="sky-search flex-1">
				<div class="sky-search-icon"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg></div>
				<input type="text" bind:value={classSearch} onkeydown={(e) => e.key === 'Enter' && (classPage = 1, loadClasses())} class="sky-input" placeholder="جستجوی کلاس..." style="padding-right: 2.5rem;" />
			</div>
			<button onclick={() => { showCreateClass = true; loadTeachers(); }} class="sky-btn sky-btn-primary flex items-center gap-2 shrink-0">
				<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
				کلاس جدید
			</button>
		</div>

		{#if classLoading}
			<div class="flex items-center justify-center py-12"><svg class="sky-spinner lg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="color: var(--color-crystal-clear);"><path d="M12 2v4M12 18v4M4.93 4.93l2.83 2.83M16.24 16.24l2.83 2.83M2 12h4M18 12h4M4.93 19.07l2.83-2.83M16.24 7.76l2.83-2.83"/></svg></div>
		{:else}
			<div class="sky-card overflow-hidden">
				<table class="sky-table">
					<thead><tr><th>نام</th><th>توضیحات</th><th>حداکثر</th><th>تاریخ ایجاد</th></tr></thead>
					<tbody>
						{#each classes as cls}
							<tr>
								<td>
									<div class="flex items-center gap-2">
										<div class="w-3 h-3 rounded-full shrink-0" style="background-color: {cls.color}"></div>
										<span class="font-semibold">{cls.name}</span>
									</div>
								</td>
								<td style="color: var(--color-mystic-sea); max-width: 200px;" class="truncate">{cls.description || '—'}</td>
								<td>{toPersian(cls.max_students)}</td>
								<td style="color: var(--color-mystic-sea);">{formatDate(cls.created_at)}</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		{/if}

	{:else if activeTab === 'sessions'}
		<div class="sky-search">
			<div class="sky-search-icon"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg></div>
			<input type="text" bind:value={sessionSearch} onkeydown={(e) => e.key === 'Enter' && (sessionPage = 1, loadSessions())} class="sky-input" placeholder="جستجوی جلسه..." style="padding-right: 2.5rem;" />
		</div>

		{#if sessionLoading}
			<div class="flex items-center justify-center py-12"><svg class="sky-spinner lg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="color: var(--color-crystal-clear);"><path d="M12 2v4M12 18v4M4.93 4.93l2.83 2.83M16.24 16.24l2.83 2.83M2 12h4M18 12h4M4.93 19.07l2.83-2.83M16.24 7.76l2.83-2.83"/></svg></div>
		{:else}
			<div class="sky-card overflow-hidden">
				<table class="sky-table">
					<thead><tr><th>عنوان</th><th>تاریخ</th><th>مدت</th><th>وضعیت</th></tr></thead>
					<tbody>
						{#each sessions as s}
							<tr>
								<td class="font-semibold">{s.title}</td>
								<td style="color: var(--color-mystic-sea);">{formatDate(s.scheduled_at)}</td>
								<td>{toPersian(s.duration)} دقیقه</td>
								<td>
									<span class="sky-badge {s.status === 'live' ? 'sky-badge-success' : s.status === 'scheduled' ? 'sky-badge-info' : 'sky-badge-default'}">
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
	<div class="modal-overlay" onclick={() => showCreateUser = false} role="button" tabindex="-1">
		<div class="modal-content" onclick={(e) => e.stopPropagation()}>
			<div class="sky-modal-header">
				<h2>ایجاد کاربر جدید</h2>
				<button onclick={() => showCreateUser = false} class="sky-btn-icon"><svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg></button>
			</div>
			<div class="sky-modal-body space-y-4">
				{#if createUserError}
					<div class="p-3 rounded-lg text-sm" style="background: rgba(224,82,82,0.1); color: var(--color-fiery-passion);">{createUserError}</div>
				{/if}
				<div><label class="sky-label">نام نمایشی</label><input type="text" bind:value={newUser.display_name} class="sky-input" required /></div>
				<div><label class="sky-label">ایمیل</label><input type="email" bind:value={newUser.email} class="sky-input" dir="ltr" required /></div>
				<div><label class="sky-label">رمز عبور</label><input type="password" bind:value={newUser.password} class="sky-input" dir="ltr" required /></div>
				<div class="grid grid-cols-2 gap-3">
					<div>
						<label class="sky-label">نقش</label>
						<select bind:value={newUser.role} class="sky-input">
							<option value="student">دانش‌آموز</option>
							<option value="teacher">مدرس</option>
							<option value="admin">مدیر</option>
						</select>
					</div>
					<div><label class="sky-label">تلفن</label><input type="tel" bind:value={newUser.phone} class="sky-input" dir="ltr" /></div>
				</div>
			</div>
			<div class="sky-modal-footer">
				<button onclick={() => showCreateUser = false} class="sky-btn sky-btn-secondary">انصراف</button>
				<button onclick={createUser} disabled={createUserLoading || !newUser.email || !newUser.password || !newUser.display_name} class="sky-btn sky-btn-primary">
					{createUserLoading ? 'در حال ایجاد...' : 'ایجاد کاربر'}
				</button>
			</div>
		</div>
	</div>
{/if}

<!-- Create Class Modal -->
{#if showCreateClass}
	<div class="modal-overlay" onclick={() => showCreateClass = false} role="button" tabindex="-1">
		<div class="modal-content" onclick={(e) => e.stopPropagation()}>
			<div class="sky-modal-header">
				<h2>ایجاد کلاس جدید</h2>
				<button onclick={() => showCreateClass = false} class="sky-btn-icon"><svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg></button>
			</div>
			<div class="sky-modal-body space-y-4">
				<div><label class="sky-label">نام کلاس</label><input type="text" bind:value={newClass.name} class="sky-input" required /></div>
				<div><label class="sky-label">توضیحات</label><textarea bind:value={newClass.description} class="sky-input resize-none" rows="2"></textarea></div>
				<div><label class="sky-label">حداکثر دانش‌آموز</label><input type="number" bind:value={newClass.max_students} class="sky-input" min="1" /></div>
			</div>
			<div class="sky-modal-footer">
				<button onclick={() => showCreateClass = false} class="sky-btn sky-btn-secondary">انصراف</button>
				<button onclick={createClass} disabled={!newClass.name} class="sky-btn sky-btn-primary">ایجاد کلاس</button>
			</div>
		</div>
	</div>
{/if}

<ConfirmModal bind:show={showDeleteUserConfirm} title="غیرفعال‌سازی کاربر" message="آیا از غیرفعال‌سازی این کاربر اطمینان دارید؟" onConfirm={() => deleteUser(deleteUserId)} onCancel={() => {}} />
