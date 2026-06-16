<script lang="ts">
	import { api } from '$lib/api';
	import { onMount } from 'svelte';
	import type { User, Class, Session, DashboardStats } from '$lib/types';

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

	const perPage = 15;

	onMount(() => {
		loadUsers();
		loadClasses();
		loadSessions();
	});

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
		if (!confirm('آیا از غیرفعال‌سازی این کاربر اطمینان دارید؟')) return;
		const res = await api.delete(`/admin/users/${id}`);
		if (res.success) {
			users = users.map(u => u.id === id ? { ...u, is_active: false } : u);
		}
	}

	async function createClass() {
		const res = await api.post('/classes', newClass);
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

	const roleLabels: Record<string, string> = { admin: 'مدیر', teacher: 'مدرس', student: 'دانش‌آموز' };
	const roleColors: Record<string, string> = { admin: 'bg-red-100 text-red-700', teacher: 'bg-purple-100 text-purple-700', student: 'bg-blue-100 text-blue-700' };
</script>

<div class="space-y-6">
	<div>
		<h1 class="text-2xl font-bold text-gray-900">پنل مدیریت</h1>
		<p class="text-gray-500 mt-1">مدیریت کاربران، کلاس‌ها و جلسات</p>
	</div>

	<!-- Tabs -->
	<div class="flex items-center gap-3 flex-wrap">
		<div class="flex gap-1 bg-gray-100 p-1 rounded-lg w-fit">
			<button class="px-4 py-2 rounded-md text-sm font-medium transition-all {activeTab === 'users' ? 'bg-white text-blue-600 shadow-sm' : 'text-gray-500'}" onclick={() => activeTab = 'users'}>کاربران</button>
			<button class="px-4 py-2 rounded-md text-sm font-medium transition-all {activeTab === 'classes' ? 'bg-white text-blue-600 shadow-sm' : 'text-gray-500'}" onclick={() => activeTab = 'classes'}>کلاس‌ها</button>
			<button class="px-4 py-2 rounded-md text-sm font-medium transition-all {activeTab === 'sessions' ? 'bg-white text-blue-600 shadow-sm' : 'text-gray-500'}" onclick={() => activeTab = 'sessions'}>جلسات</button>
		</div>
		<div class="flex gap-1">
			<button onclick={() => goto('/admin/recordings')} class="px-3 py-2 text-sm text-gray-600 hover:bg-gray-100 rounded-lg">ضبط‌ها</button>
			<button onclick={() => goto('/admin/logs')} class="px-3 py-2 text-sm text-gray-600 hover:bg-gray-100 rounded-lg">لاگ‌ها</button>
			<button onclick={() => goto('/admin/settings')} class="px-3 py-2 text-sm text-gray-600 hover:bg-gray-100 rounded-lg flex items-center gap-1">
				<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.066 2.573c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-1.066 2.573c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" /><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" /></svg>
				تنظیمات
			</button>
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
						<span>{userTotal} کاربر</span>
						<div class="flex gap-1">
							<button disabled={userPage <= 1} onclick={() => { userPage--; loadUsers(); }} class="px-3 py-1 border rounded hover:bg-gray-50 disabled:opacity-50">قبلی</button>
							<span class="px-3 py-1">صفحه {userPage} از {Math.ceil(userTotal / perPage)}</span>
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
								<td class="px-5 py-3">{cls.max_students}</td>
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
								<td class="px-5 py-3">{s.duration} دقیقه</td>
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
