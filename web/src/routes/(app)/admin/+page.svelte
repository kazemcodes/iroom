<script lang="ts">
	import { api } from '$lib/api';
	import { onMount, onDestroy } from 'svelte';
	import { goto } from '$app/navigation';
	import type { User, Session, Room, ActivityLog } from '$lib/types';
	import ConfirmModal from '$lib/components/ConfirmModal.svelte';

	let refreshInterval: ReturnType<typeof setInterval> | null = null;
	let activeTab = $state<'users' | 'rooms' | 'sessions'>('users');

	// Stats
	let loading = $state(true);
	let stats = $state({ users: 0, rooms: 0, sessions: 0 });
	let todaySessions = $state(0);
	let recordings = $state(0);
	let liveRooms = $state<any[]>([]);
	let activityLogs = $state<ActivityLog[]>([]);

	// Users
	let users = $state<User[]>([]);
	let userTotal = $state(0);
	let userPage = $state(1);
	let userSearch = $state('');
	let userLoading = $state(true);

	// Rooms
	let rooms = $state<Room[]>([]);
	let roomTotal = $state(0);
	let roomPage = $state(1);
	let roomSearch = $state('');
	let roomLoading = $state(true);
	let userCounts = $state<Record<number, number>>({});

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

	// Delete confirm
	let showDeleteUserConfirm = $state(false);
	let deleteUserId = $state(0);

	const perPage = 15;

	onMount(async () => {
		await Promise.all([loadDashboard(), loadUsers(), loadRooms(), loadSessions()]);
		await loadHealth();
		refreshInterval = setInterval(async () => {
			await loadDashboard();
			await loadHealth();
		}, 30000);
	});

	onDestroy(() => { if (refreshInterval) clearInterval(refreshInterval); });

	async function loadDashboard() {
		loading = true;
		const statsRes = await api.get<any>('/admin/dashboard/stats');
		if (statsRes.success && statsRes.data) {
			stats = { users: statsRes.data.users || 0, rooms: statsRes.data.rooms || 0, sessions: statsRes.data.sessions || 0 };
		}

		const logsRes = await api.get<{ items: ActivityLog[] }>('/admin/activity-logs', { per_page: '10' });
		if (logsRes.success && logsRes.data) {
			activityLogs = (logsRes.data as any).items || (Array.isArray(logsRes.data) ? logsRes.data : []);
		}

		// Load live rooms
		const roomsRes = await api.get<any>('/admin/rooms', { per_page: '100' });
		if (roomsRes.success && roomsRes.data) {
			const allRooms = roomsRes.data.items || [];
			// Check which rooms have active sessions
			const sessRes = await api.get<any>('/admin/sessions');
			if (sessRes.success && sessRes.data) {
				const allSessions = Array.isArray(sessRes.data) ? sessRes.data : (sessRes.data.items || []);
				const liveSessionIds = new Set(allSessions.filter((s: any) => s.status === 'live').map((s: any) => s.class_id));
				liveRooms = allRooms.filter((r: any) => liveSessionIds.has(r.id));
			}
		}

		// Load today's sessions count
		const sessRes2 = await api.get<any>('/admin/sessions');
		if (sessRes2.success && sessRes2.data) {
			const allSessions = Array.isArray(sessRes2.data) ? sessRes2.data : (sessRes2.data.items || []);
			const today = new Date().toDateString();
			todaySessions = allSessions.filter((s: any) => new Date(s.created_at).toDateString() === today).length;
		}

		// Load recordings count
		const recRes = await api.get<any>('/admin/recordings');
		if (recRes.success && recRes.data) {
			recordings = Array.isArray(recRes.data) ? recRes.data.length : (recRes.data.items?.length || 0);
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

	async function loadRooms() {
		roomLoading = true;
		const params: Record<string, string> = { page: String(roomPage), per_page: String(perPage) };
		if (roomSearch) params.search = roomSearch;
		const res = await api.get<{ items: Room[]; total: number }>('/admin/rooms', params);
		if (res.success && res.data) {
			rooms = res.data.items || [];
			roomTotal = res.data.total;
		}
		roomLoading = false;

		for (const room of rooms) {
			const usersRes = await api.get<User[]>(`/rooms/${room.id}/users`);
			if (usersRes.success && Array.isArray(usersRes.data)) {
				userCounts[room.id] = usersRes.data.length;
			}
		}
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

	// Health
	let healthLoading = $state(true);
	let healthData = $state({ serverUptime: '', dbSize: '', webrtcStatus: 'pion_builtin', activeRooms: 0 });

	async function loadHealth() {
		healthLoading = true;
		try {
			const res = await fetch(window.location.origin + '/api/v1/health');
			const data = await res.json();
			if (data && data.status === 'ok') {
				healthData = {
					serverUptime: data.uptime || '',
					dbSize: data.db_size || '',
					webrtcStatus: data.webrtc_status || 'pion_builtin',
					activeRooms: data.active_rooms ?? 0,
				};
			}
		} catch (e) {}
		healthLoading = false;
	}

	// Create user
	async function createUser() {
		createUserLoading = true; createUserError = '';
		const res = await api.post('/admin/users', newUser);
		if (!res.success) { createUserError = res.error || 'خطا'; createUserLoading = false; return; }
		showCreateUser = false;
		newUser = { email: '', password: '', display_name: '', phone: '', role: 'student' };
		createUserLoading = false;
		await loadUsers();
	}

	// Delete user
	function confirmDeleteUser(id: number) { deleteUserId = id; showDeleteUserConfirm = true; }
	async function deleteUser() {
		const res = await api.delete(`/admin/users/${deleteUserId}`);
		if (res.success) await loadUsers();
	}

	// Activity labels
	function actionLabel(action: string): string {
		const map: Record<string, string> = {
			'create_user': 'ایجاد کاربر', 'update_user': 'بروزرسانی کاربر', 'delete_user': 'حذف کاربر',
			'create_room': 'ایجاد اتاق', 'update_room': 'بروزرسانی اتاق', 'delete_room': 'حذف اتاق',
			'create_session': 'ایجاد جلسه', 'session_action': 'عملیات جلسه', 'delete_session': 'حذف جلسه',
			'upload_file': 'آپلود فایل', 'delete_file': 'حذف فایل', 'update_settings': 'بروزرسانی تنظیمات',
			'add_room_user': 'افزودن کاربر به اتاق', 'remove_room_user': 'حذف کاربر از اتاق',
		};
		return map[action] || action;
	}
	function actionColor(action: string): string {
		const map: Record<string, string> = {
			'create_user': 'bg-blue-500', 'create_room': 'bg-green-500', 'create_session': 'bg-emerald-500',
			'delete_user': 'bg-red-500', 'delete_room': 'bg-red-500', 'upload_file': 'bg-amber-500',
		};
		return map[action] || 'bg-gray-400';
	}
	function cleanPath(path: string): string {
		return path.replace(/^\/api\/v1/, '').replace(/\/\d+/g, '/:id');
	}

	const roleLabels: Record<string, string> = { admin: 'مدیر', teacher: 'مدرس', student: 'دانش‌آموز' };
	function toPersian(n: number) { return n.toLocaleString('fa-IR'); }
	function formatDate(d: string) { return d ? new Date(d).toLocaleDateString('fa-IR', { year: 'numeric', month: 'short', day: 'numeric' }) : '—'; }
	const statusLabels: Record<string, string> = { scheduled: 'برنامه‌ریزی شده', live: 'در حال برگزاری', ended: 'پایان یافته' };
	const statusBadge: Record<string, string> = { scheduled: 'sky-badge sky-badge-info', live: 'sky-badge sky-badge-success', ended: 'sky-badge sky-badge-default' };
</script>

<div class="space-y-6">
	<div>
		<h1 style="font-size:1.5rem;font-weight:700;color:var(--color-midnight-sky);">پنل مدیریت</h1>
		<p style="font-size:0.875rem;color:var(--color-mystic-sea);margin-top:4px;">داشبورد مدیریت سیستم</p>
	</div>

	{#if loading}
		<div class="flex items-center justify-center py-20"><div class="animate-spin h-8 w-8 border-4 border-[#23b9d7] border-t-transparent rounded-full"></div></div>
	{:else}
		<!-- Stats Cards -->
		<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
			<div class="sky-stat-card">
				<div class="sky-stat-icon" style="background:linear-gradient(135deg,#23b9d7,#004ff2);">
					<svg width="24" height="24" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M15 19.128a9.38 9.38 0 002.625.372 9.337 9.337 0 004.121-.952 4.125 4.125 0 00-7.533-2.493M15 19.128v-.003c0-1.113-.285-2.16-.786-3.07M15 19.128v.106A12.318 12.318 0 018.624 21c-2.331 0-4.512-.645-6.374-1.766l-.001-.109a6.375 6.375 0 0111.964-3.07M12 6.375a3.375 3.375 0 11-6.75 0 3.375 3.375 0 016.75 0zm8.25 2.25a2.625 2.625 0 11-5.25 0 2.625 2.625 0 015.25 0z" /></svg>
				</div>
				<div><p class="sky-stat-label">کاربران</p><p class="sky-stat-value">{toPersian(stats.users)}</p></div>
			</div>
			<div class="sky-stat-card">
				<div class="sky-stat-icon" style="background:linear-gradient(135deg,#40bf7f,#059669);">
					<svg width="24" height="24" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path d="M2 12s3-7 10-7 10 7 10 7-3 7-10 7-10-7-10-7z"/><circle cx="12" cy="12" r="3"/></svg>
				</div>
				<div><p class="sky-stat-label">اتاق‌ها</p><p class="sky-stat-value">{toPersian(stats.rooms)}</p></div>
			</div>
			<div class="sky-stat-card">
				<div class="sky-stat-icon" style="background:linear-gradient(135deg,#f59e0b,#d97706);">
					<svg width="24" height="24" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M6.75 3v2.25M17.25 3v2.25M3 18.75V7.5a2.25 2.25 0 012.25-2.25h13.5A2.25 2.25 0 0121 7.5v11.25m-18 0A2.25 2.25 0 005.25 21h13.5A2.25 2.25 0 0021 18.75m-18 0v-7.5A2.25 2.25 0 015.25 9h13.5A2.25 2.25 0 0121 11.25v7.5" /></svg>
				</div>
				<div><p class="sky-stat-label">جلسات امروز</p><p class="sky-stat-value">{toPersian(todaySessions)}</p></div>
			</div>
			<div class="sky-stat-card">
				<div class="sky-stat-icon" style="background:linear-gradient(135deg,#8b5cf6,#7c3aed);">
					<svg width="24" height="24" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M15.75 10.5l4.72-4.72a.75.75 0 011.28.53v11.38a.75.75 0 01-1.28.53l-4.72-4.72M4.5 18.75h9a2.25 2.25 0 002.25-2.25v-9a2.25 2.25 0 00-2.25-2.25h-9A2.25 2.25 0 002.25 7.5v9a2.25 2.25 0 002.25 2.25z" /></svg>
				</div>
				<div><p class="sky-stat-label">ضبط‌ها</p><p class="sky-stat-value">{toPersian(recordings)}</p></div>
			</div>
		</div>

		<!-- System Health -->
		<div class="sky-card">
			<div class="sky-card-header"><h2>وضعیت سیستم</h2></div>
			<div class="sky-card-body">
				{#if healthLoading}
					<div class="flex items-center justify-center py-4"><svg class="sky-spinner md" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="color: var(--color-crystal-clear);"><path d="M12 2v4M12 18v4M4.93 4.93l2.83 2.83M16.24 16.24l2.83 2.83M2 12h4M18 12h4M4.93 19.07l2.83-2.83M16.24 7.76l2.83-2.83"/></svg></div>
				{:else}
					<div class="grid grid-cols-2 sm:grid-cols-4 gap-4">
						{#each [
							{ label: 'آپتایم سرور', value: healthData.serverUptime || '—' },
							{ label: 'WebRTC', value: 'فعال (Pion)' },
							{ label: 'حجم پایگاه داده', value: healthData.dbSize || '—' },
							{ label: 'اتاق‌های فعال', value: toPersian(healthData.activeRooms) },
						] as item}
							<div class="text-center p-3 rounded-lg" style="background: var(--color-secret-glow);">
								<p class="text-xs mb-1" style="color: var(--color-moonlit-mist);">{item.label}</p>
								<p class="text-sm font-bold" style="color: var(--color-midnight-sky);">{item.value}</p>
							</div>
						{/each}
					</div>
				{/if}
			</div>
		</div>

		<!-- Live Rooms + Activity Feed -->
		<div class="grid grid-cols-1 lg:grid-cols-3 gap-5">
			<div class="lg:col-span-2">
				<div class="sky-card">
					<div class="sky-card-header">
						<h2>اتاق‌های فعال</h2>
						<button onclick={() => goto('/admin/rooms')} class="sky-btn sky-btn-ghost" style="padding: 0.25rem 0.75rem; font-size: 12px;">مشاهده همه</button>
					</div>
					<div class="sky-card-body">
						{#if liveRooms.length === 0}
							<p class="text-sm py-4 text-center" style="color: var(--color-moonlit-mist);">اتاق فعالی وجود ندارد</p>
						{:else}
							<div class="space-y-2">
								{#each liveRooms as room}
									<a href="/room/{room.slug}" target="_blank" class="flex items-center justify-between p-3 rounded-lg transition-colors hover:bg-gray-50" style="text-decoration:none;">
										<div class="flex items-center gap-3">
											<div style="width:36px;height:36px;border-radius:8px;display:flex;align-items:center;justify-content:center;color:white;font-weight:700;font-size:14px;background:{room.color};">{room.name.charAt(0)}</div>
											<div><p class="font-semibold text-sm" style="color:var(--color-midnight-sky);">{room.name}</p></div>
										</div>
										<span class="sky-badge sky-badge-success"><span class="dot"></span> زنده</span>
									</a>
								{/each}
							</div>
						{/if}
					</div>
				</div>
			</div>

			<!-- Activity Feed -->
			<div>
				<div class="sky-card">
					<div class="sky-card-header"><h2>آخرین فعالیت‌ها</h2></div>
					<div class="sky-card-body">
						{#if activityLogs.length === 0}
							<p class="text-sm py-4 text-center" style="color: var(--color-moonlit-mist);">فعالیتی ثبت نشده</p>
						{:else}
							<div class="space-y-3">
								{#each activityLogs as log}
									<div class="flex items-start gap-3">
										<div class="w-2 h-2 rounded-full mt-2 shrink-0 {actionColor(log.action)}"></div>
										<div class="flex-1 min-w-0">
											<p class="text-sm truncate" style="color: var(--color-midnight-sky);">{actionLabel(log.action)}</p>
											<p class="text-xs" style="color: var(--color-moonlit-mist);">{cleanPath(log.details) || '—'}</p>
										</div>
									</div>
								{/each}
							</div>
						{/if}
					</div>
				</div>
			</div>
		</div>

		<!-- Tabs: Users, Rooms, Sessions -->
		<div class="sky-card">
			<div class="sky-tabs" style="padding: 0 1.25rem;">
				<button class="sky-tab {activeTab === 'users' ? 'active' : ''}" onclick={() => { activeTab = 'users'; }}>کاربران</button>
				<button class="sky-tab {activeTab === 'rooms' ? 'active' : ''}" onclick={() => { activeTab = 'rooms'; }}>اتاق‌ها</button>
				<button class="sky-tab {activeTab === 'sessions' ? 'active' : ''}" onclick={() => { activeTab = 'sessions'; }}>جلسات</button>
			</div>
			<div class="p-4">
				{#if activeTab === 'users'}
					<div class="flex items-center justify-between mb-3">
						<div class="sky-search flex-1 max-w-sm">
							<div class="sky-search-icon"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg></div>
							<input type="text" bind:value={userSearch} onkeydown={(e) => e.key === 'Enter' && loadUsers()} class="sky-input" style="padding-right:2.5rem;" placeholder="جستجو..." />
						</div>
						<button onclick={() => showCreateUser = true} class="sky-btn sky-btn-primary" style="font-size:12px;">کاربر جدید</button>
					</div>
					{#if userLoading}
						<div class="flex justify-center py-8"><div class="animate-spin h-6 w-6 border-2 border-[var(--color-crystal-clear)] border-t-transparent rounded-full"></div></div>
					{:else}
						<table class="sky-table"><thead><tr><th>نام</th><th>ایمیل</th><th>نقش</th><th>عملیات</th></tr></thead>
							<tbody>{#each users as u}<tr><td class="font-semibold">{u.display_name}</td><td dir="ltr" style="color:var(--color-mystic-sea);">{u.email}</td><td><span class="sky-badge sky-badge-{u.role === 'admin' ? 'danger' : u.role === 'teacher' ? 'info' : 'default'}">{roleLabels[u.role] || u.role}</span></td><td><button onclick={() => confirmDeleteUser(u.id)} class="sky-btn sky-btn-ghost" style="font-size:11px;color:var(--color-fiery-passion);">حذف</button></td></tr>{/each}</tbody>
						</table>
					{/if}

				{:else if activeTab === 'rooms'}
					<div class="flex items-center justify-between mb-3">
						<div class="sky-search flex-1 max-w-sm">
							<div class="sky-search-icon"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg></div>
							<input type="text" bind:value={roomSearch} onkeydown={(e) => e.key === 'Enter' && loadRooms()} class="sky-input" style="padding-right:2.5rem;" placeholder="جستجو..." />
						</div>
						<a href="/admin/rooms" class="sky-btn sky-btn-primary" style="font-size:12px;">مدیریت اتاق‌ها</a>
					</div>
					{#if roomLoading}
						<div class="flex justify-center py-8"><div class="animate-spin h-6 w-6 border-2 border-[var(--color-crystal-clear)] border-t-transparent rounded-full"></div></div>
					{:else}
						<table class="sky-table"><thead><tr><th>اتاق</th><th>توضیحات</th><th>کاربران</th><th>مهمان</th></tr></thead>
							<tbody>{#each rooms as r}<tr>
								<td><div class="flex items-center gap-2"><div style="width:28px;height:28px;border-radius:6px;display:flex;align-items:center;justify-content:center;color:white;font-weight:700;font-size:12px;background:{r.color};">{r.name.charAt(0)}</div><span class="font-semibold">{r.name}</span></div></td>
								<td style="color:var(--color-mystic-sea);max-width:150px;" class="truncate">{r.description || '—'}</td>
								<td style="color:var(--color-mystic-sea);">{toPersian(userCounts[r.id] || 0)}</td>
								<td><span class="sky-badge {r.guest_login_enabled ? 'sky-badge-success' : 'sky-badge-default'}" style="font-size:11px;">{r.guest_login_enabled ? 'فعال' : 'غیرفعال'}</span></td>
							</tr>{/each}</tbody>
						</table>
					{/if}

				{:else if activeTab === 'sessions'}
					<div class="flex items-center justify-between mb-3">
						<div class="sky-search flex-1 max-w-sm">
							<div class="sky-search-icon"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg></div>
							<input type="text" bind:value={sessionSearch} onkeydown={(e) => e.key === 'Enter' && loadSessions()} class="sky-input" style="padding-right:2.5rem;" placeholder="جستجو..." />
						</div>
					</div>
					{#if sessionLoading}
						<div class="flex justify-center py-8"><div class="animate-spin h-6 w-6 border-2 border-[var(--color-crystal-clear)] border-t-transparent rounded-full"></div></div>
					{:else}
						<table class="sky-table"><thead><tr><th>عنوان</th><th>وضعیت</th><th>تاریخ</th></tr></thead>
							<tbody>{#each sessions as s}<tr>
								<td class="font-semibold">{s.title}</td>
								<td><span class="{statusBadge[s.status] || 'sky-badge sky-badge-default'}">{statusLabels[s.status] || s.status}</span></td>
								<td style="color:var(--color-mystic-sea);font-size:12px;">{formatDate(s.scheduled_at)}</td>
							</tr>{/each}</tbody>
						</table>
					{/if}
				{/if}
			</div>
		</div>
	{/if}
</div>

<!-- Create User Modal -->
{#if showCreateUser}
	<div class="modal-overlay" onclick={() => showCreateUser = false} role="button" tabindex="-1">
		<div class="modal-content" onclick={(e) => e.stopPropagation()}>
			<div class="sky-modal-header"><h2>ایجاد کاربر جدید</h2><button onclick={() => showCreateUser = false} class="sky-btn-icon"><svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg></button></div>
			<div class="sky-modal-body space-y-3">
				{#if createUserError}<div class="p-3 rounded-lg text-sm" style="background:rgba(224,82,82,0.1);color:var(--color-fiery-passion);">{createUserError}</div>{/if}
				<div><label class="sky-label">نام نمایشی</label><input type="text" bind:value={newUser.display_name} class="sky-input" required /></div>
				<div><label class="sky-label">ایمیل</label><input type="email" bind:value={newUser.email} class="sky-input" dir="ltr" required /></div>
				<div><label class="sky-label">رمز عبور</label><input type="password" bind:value={newUser.password} class="sky-input" dir="ltr" required /></div>
				<div><label class="sky-label">نقش</label><select bind:value={newUser.role} class="sky-input"><option value="student">دانش‌آموز</option><option value="teacher">مدرس</option><option value="admin">مدیر</option></select></div>
			</div>
			<div class="sky-modal-footer"><button onclick={() => showCreateUser = false} class="sky-btn sky-btn-secondary">انصراف</button><button onclick={createUser} disabled={createUserLoading || !newUser.email || !newUser.password} class="sky-btn sky-btn-primary">{createUserLoading ? 'در حال ایجاد...' : 'ایجاد'}</button></div>
		</div>
	</div>
{/if}

<ConfirmModal bind:show={showDeleteUserConfirm} title="حذف کاربر" message="آیا از حذف این کاربر اطمینان دارید؟" onConfirm={() => { showDeleteUserConfirm = false; deleteUser(); }} onCancel={() => {}} />
