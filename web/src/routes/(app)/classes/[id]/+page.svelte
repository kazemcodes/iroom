<script lang="ts">
	import { page } from '$app/state';
	import { auth, isAdmin, isTeacher } from '$lib/stores';
	import { api } from '$lib/api';
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import type { Class, Session, User } from '$lib/types';
	import ConfirmModal from '$lib/components/ConfirmModal.svelte';

	let classData = $state<Class | null>(null);
	let sessions = $state<Session[]>([]);
	let students = $state<User[]>([]);
	let loading = $state(true);
	let activeTab = $state<'sessions' | 'students'>('sessions');

	// Enroll
	let showEnroll = $state(false);
	let enrollSearch = $state('');
	let searchResults = $state<User[]>([]);
	let enrollLoading = $state(false);

	// Create session
	let showCreateSession = $state(false);
	let sessionTitle = $state('');
	let sessionDate = $state('');
	let sessionTime = $state('');
	let sessionDuration = $state(60);
	let sessionLoading = $state(false);
	let sessionError = $state('');

	// Edit class
	let showEdit = $state(false);
	let editName = $state('');
	let editDesc = $state('');
	let editColor = $state('#3B82F6');
	let editMax = $state(30);
	let editLoading = $state(false);

	// Delete confirm
	let showDeleteConfirm = $state(false);

	const classId = $derived(page.params.id);

	onMount(() => loadData());

	async function loadData() {
		loading = true;
		const [classRes, sessionsRes, studentsRes] = await Promise.all([
			api.get<Class>(`/classes/${classId}`),
			api.get<Session[]>(`/sessions?class_id=${classId}`),
			api.get<User[]>(`/classes/${classId}/students`)
		]);

		if (classRes.success) classData = classRes.data!;
		if (sessionsRes.success && sessionsRes.data) sessions = Array.isArray(sessionsRes.data) ? sessionsRes.data : [];
		if (studentsRes.success && studentsRes.data) students = Array.isArray(studentsRes.data) ? studentsRes.data : [];
		loading = false;
	}

	async function createSession() {
		sessionLoading = true;
		sessionError = '';
		const dt = `${sessionDate}T${sessionTime}:00Z`;
		const res = await api.post('/sessions', {
			class_id: parseInt(classId),
			title: sessionTitle,
			scheduled_at: new Date(dt).toISOString(),
			duration: sessionDuration
		});

		if (!res.success) {
			sessionError = res.error || 'خطا در ایجاد جلسه';
			sessionLoading = false;
			return;
		}

		sessions = [res.data!, ...sessions];
		showCreateSession = false;
		sessionTitle = '';
		sessionDate = '';
		sessionTime = '';
		sessionDuration = 60;
		sessionLoading = false;
	}

	async function startSession(id: number) {
		const res = await api.post(`/sessions/${id}/start`);
		if (res.success) {
			sessions = sessions.map(s => s.id === id ? { ...s, status: 'live', livekit_room: res.data!.livekit_room } : s);
		}
	}

	async function endSession(id: number) {
		const res = await api.post(`/sessions/${id}/end`);
		if (res.success) {
			sessions = sessions.map(s => s.id === id ? { ...s, status: 'ended' } : s);
		}
	}

	async function searchUsers() {
		if (!enrollSearch) { searchResults = []; return; }
		enrollLoading = true;
		const res = await api.get<{ items: User[] }>('/admin/users', { search: enrollSearch, page: '1', per_page: '10' });
		if (res.success && res.data) {
			searchResults = res.data.items || [];
		}
		enrollLoading = false;
	}

	async function enrollStudent(studentId: number) {
		const res = await api.post(`/classes/${classId}/enroll`, { student_id: studentId });
		if (res.success) {
			const found = searchResults.find(u => u.id === studentId);
			if (found) students = [...students, found];
			searchResults = searchResults.filter(u => u.id !== studentId);
		}
	}

	async function deleteClass() {
		const res = await api.delete(`/classes/${classId}`);
		if (res.success) goto('/classes');
	}

	function formatDate(d: string) {
		if (!d) return '';
		return new Date(d).toLocaleDateString('fa-IR', { year: 'numeric', month: 'short', day: 'numeric' });
	}

	function formatTime(d: string) {
		if (!d) return '';
		return new Date(d).toLocaleTimeString('fa-IR', { hour: '2-digit', minute: '2-digit' });
	}

	const statusLabels: Record<string, string> = { scheduled: 'برنامه‌ریزی شده', live: 'در حال برگزاری', ended: 'پایان یافته' };
	const statusColors: Record<string, string> = { scheduled: 'bg-blue-100 text-blue-700', live: 'bg-green-100 text-green-700', ended: 'bg-gray-100 text-gray-500' };
</script>

{#if loading}
	<div class="flex items-center justify-center py-20">
		<div class="animate-spin h-8 w-8 border-4 border-blue-600 border-t-transparent rounded-full"></div>
	</div>
{:else if classData}
	<div class="space-y-6">
		<!-- Header -->
		<div class="bg-white rounded-xl border p-6">
			<div class="flex items-start justify-between">
				<div class="flex items-center gap-4">
					<div class="w-5 h-5 rounded-full shrink-0" style="background-color: {classData.color}"></div>
					<div>
						<h1 class="text-2xl font-bold text-gray-900">{classData.name}</h1>
						{#if classData.description}
							<p class="text-gray-500 mt-1">{classData.description}</p>
						{/if}
					</div>
				</div>
				<div class="flex items-center gap-2">
					{#if $isAdmin || $isTeacher}
						<button
							onclick={() => {
								showEdit = true;
								editName = classData!.name;
								editDesc = classData!.description;
								editColor = classData!.color;
								editMax = classData!.max_students;
							}}
							class="px-3 py-1.5 text-sm text-gray-600 hover:bg-gray-100 rounded-lg transition-colors"
						>ویرایش</button>
						<button onclick={() => showDeleteConfirm = true} class="px-3 py-1.5 text-sm text-red-600 hover:bg-red-50 rounded-lg transition-colors">حذف</button>
					{/if}
				</div>
			</div>
		</div>

		<!-- Tabs -->
		<div class="flex gap-1 bg-gray-100 p-1 rounded-lg w-fit">
			<button
				class="px-4 py-2 rounded-md text-sm font-medium transition-all {activeTab === 'sessions' ? 'bg-white text-blue-600 shadow-sm' : 'text-gray-500 hover:text-gray-700'}"
				onclick={() => activeTab = 'sessions'}
			>جلسات ({sessions.length})</button>
			<button
				class="px-4 py-2 rounded-md text-sm font-medium transition-all {activeTab === 'students' ? 'bg-white text-blue-600 shadow-sm' : 'text-gray-500 hover:text-gray-700'}"
				onclick={() => activeTab = 'students'}
			>دانش‌آموزان ({students.length})</button>
		</div>

		{#if activeTab === 'sessions'}
			<div class="flex justify-end">
				{#if $isAdmin || $isTeacher}
					<button onclick={() => showCreateSession = true} class="px-4 py-2.5 bg-blue-600 text-white rounded-lg text-sm font-medium hover:bg-blue-700 transition-colors flex items-center gap-2">
						<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" /></svg>
						جلسه جدید
					</button>
				{/if}
			</div>

			{#if sessions.length === 0}
				<div class="text-center py-12 bg-white rounded-xl border">
					<p class="text-gray-500">هنوز جلسه‌ای ایجاد نشده</p>
				</div>
			{:else}
				<div class="space-y-3">
					{#each sessions as s}
						<div class="bg-white rounded-xl border p-4 flex items-center justify-between">
							<div class="flex items-center gap-4">
								<div class="w-10 h-10 rounded-lg flex items-center justify-center {s.status === 'live' ? 'bg-green-100' : 'bg-gray-100'}">
									<svg class="w-5 h-5 {s.status === 'live' ? 'text-green-600' : 'text-gray-400'}" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" />
									</svg>
								</div>
								<div>
									<p class="font-medium text-gray-900">{s.title}</p>
									<p class="text-sm text-gray-500">{formatDate(s.scheduled_at)} — {formatTime(s.scheduled_at)} • {s.duration} دقیقه</p>
								</div>
							</div>
							<div class="flex items-center gap-3">
								<span class="text-xs px-2.5 py-1 rounded-full font-medium {statusColors[s.status]}">{statusLabels[s.status]}</span>
								{#if ($isAdmin || $isTeacher) && s.status === 'scheduled'}
									<button onclick={() => startSession(s.id)} class="px-3 py-1.5 bg-green-600 text-white text-xs rounded-lg hover:bg-green-700 transition-colors">شروع</button>
								{/if}
								{#if ($isAdmin || $isTeacher) && s.status === 'live'}
									<a href="/classroom/{s.id}" class="px-3 py-1.5 bg-blue-600 text-white text-xs rounded-lg hover:bg-blue-700 transition-colors">ورود</a>
									<button onclick={() => endSession(s.id)} class="px-3 py-1.5 bg-red-600 text-white text-xs rounded-lg hover:bg-red-700 transition-colors">پایان</button>
								{/if}
								{#if s.status === 'live'}
									<a href="/classroom/{s.id}" class="px-3 py-1.5 bg-green-600 text-white text-xs rounded-lg hover:bg-green-700 transition-colors">پیوستن</a>
								{/if}
							</div>
						</div>
					{/each}
				</div>
			{/if}
		{:else}
			<!-- Students tab -->
			<div class="flex justify-end">
				{#if $isAdmin || $isTeacher}
					<button onclick={() => showEnroll = true} class="px-4 py-2.5 bg-blue-600 text-white rounded-lg text-sm font-medium hover:bg-blue-700 transition-colors flex items-center gap-2">
						<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M18 9v3m0 0v3m0-3h3m-3 0h-3m-2-5a4 4 0 11-8 0 4 4 0 018 0zM3 20a6 6 0 0112 0v1H3v-1z" /></svg>
						افزودن دانش‌آموز
					</button>
				{/if}
			</div>

			{#if students.length === 0}
				<div class="text-center py-12 bg-white rounded-xl border">
					<p class="text-gray-500">هنوز دانش‌آموزی ثبت‌نام نکرده</p>
				</div>
			{:else}
				<div class="bg-white rounded-xl border divide-y">
					{#each students as student}
						<div class="px-5 py-3 flex items-center justify-between">
							<div class="flex items-center gap-3">
								<div class="w-9 h-9 rounded-full bg-blue-100 flex items-center justify-center text-blue-700 font-bold text-sm">{student.display_name.charAt(0)}</div>
								<div>
									<p class="font-medium text-gray-900 text-sm">{student.display_name}</p>
									<p class="text-xs text-gray-500">{student.email}</p>
								</div>
							</div>
						</div>
					{/each}
				</div>
			{/if}
		{/if}
	</div>

	<!-- Create Session Modal -->
	{#if showCreateSession}
		<div class="fixed inset-0 bg-black/40 z-50 flex items-center justify-center p-4" onclick={() => showCreateSession = false}>
			<div class="bg-white rounded-2xl w-full max-w-md shadow-xl" onclick={(e) => e.stopPropagation()}>
				<div class="px-6 py-4 border-b"><h2 class="font-bold text-lg">جلسه جدید</h2></div>
				<div class="px-6 py-4 space-y-4">
					{#if sessionError}
						<div class="p-3 bg-red-50 text-red-600 rounded-lg text-sm">{sessionError}</div>
					{/if}
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">عنوان</label>
						<input type="text" bind:value={sessionTitle} class="w-full px-4 py-2.5 border rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none text-sm" placeholder="عنوان جلسه" required />
					</div>
					<div class="grid grid-cols-2 gap-3">
						<div>
							<label class="block text-sm font-medium text-gray-700 mb-1">تاریخ</label>
							<input type="date" bind:value={sessionDate} class="w-full px-4 py-2.5 border rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none text-sm" required />
						</div>
						<div>
							<label class="block text-sm font-medium text-gray-700 mb-1">ساعت</label>
							<input type="time" bind:value={sessionTime} class="w-full px-4 py-2.5 border rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none text-sm" required />
						</div>
					</div>
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">مدت (دقیقه)</label>
						<input type="number" bind:value={sessionDuration} class="w-full px-4 py-2.5 border rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none text-sm" min="15" max="480" />
					</div>
				</div>
				<div class="px-6 py-4 border-t flex justify-end gap-3">
					<button onclick={() => showCreateSession = false} class="px-4 py-2 text-sm text-gray-600 hover:bg-gray-100 rounded-lg">انصراف</button>
					<button onclick={createSession} disabled={sessionLoading || !sessionTitle || !sessionDate || !sessionTime} class="px-4 py-2 bg-blue-600 text-white text-sm rounded-lg font-medium hover:bg-blue-700 disabled:opacity-50">
						{sessionLoading ? 'در حال ایجاد...' : 'ایجاد جلسه'}
					</button>
				</div>
			</div>
		</div>
	{/if}

	<!-- Enroll Modal -->
	{#if showEnroll}
		<div class="fixed inset-0 bg-black/40 z-50 flex items-center justify-center p-4" onclick={() => showEnroll = false}>
			<div class="bg-white rounded-2xl w-full max-w-md shadow-xl" onclick={(e) => e.stopPropagation()}>
				<div class="px-6 py-4 border-b"><h2 class="font-bold text-lg">افزودن دانش‌آموز</h2></div>
				<div class="px-6 py-4">
					<div class="flex gap-2 mb-4">
						<input type="text" bind:value={enrollSearch} onkeydown={(e) => e.key === 'Enter' && searchUsers()} class="flex-1 px-4 py-2.5 border rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none" placeholder="جستجو با ایمیل..." />
						<button onclick={searchUsers} class="px-4 py-2.5 bg-gray-100 rounded-lg text-sm hover:bg-gray-200">جستجو</button>
					</div>
					{#if enrollLoading}
						<p class="text-sm text-gray-500 text-center py-4">در حال جستجو...</p>
					{:else if searchResults.length > 0}
						<div class="space-y-2 max-h-60 overflow-y-auto">
							{#each searchResults as user}
								<div class="flex items-center justify-between p-3 bg-gray-50 rounded-lg">
									<div>
										<p class="text-sm font-medium">{user.display_name}</p>
										<p class="text-xs text-gray-500">{user.email}</p>
									</div>
									<button onclick={() => enrollStudent(user.id)} class="px-3 py-1 bg-blue-600 text-white text-xs rounded-lg hover:bg-blue-700">افزودن</button>
								</div>
							{/each}
						</div>
					{:else if enrollSearch}
						<p class="text-sm text-gray-500 text-center py-4">نتیجه‌ای یافت نشد</p>
					{/if}
				</div>
				<div class="px-6 py-4 border-t flex justify-end">
					<button onclick={() => showEnroll = false} class="px-4 py-2 text-sm text-gray-600 hover:bg-gray-100 rounded-lg">بستن</button>
				</div>
			</div>
		</div>
	{/if}

	<!-- Edit Class Modal -->
	{#if showEdit}
		<div class="fixed inset-0 bg-black/40 z-50 flex items-center justify-center p-4" onclick={() => showEdit = false}>
			<div class="bg-white rounded-2xl w-full max-w-md shadow-xl" onclick={(e) => e.stopPropagation()}>
				<div class="px-6 py-4 border-b"><h2 class="font-bold text-lg">ویرایش کلاس</h2></div>
				<div class="px-6 py-4 space-y-4">
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">نام</label>
						<input type="text" bind:value={editName} class="w-full px-4 py-2.5 border rounded-lg text-sm focus:ring-2 focus:ring-blue-500 outline-none" />
					</div>
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">توضیحات</label>
						<textarea bind:value={editDesc} class="w-full px-4 py-2.5 border rounded-lg text-sm focus:ring-2 focus:ring-blue-500 outline-none resize-none" rows="2"></textarea>
					</div>
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">حداکثر</label>
						<input type="number" bind:value={editMax} class="w-full px-4 py-2.5 border rounded-lg text-sm focus:ring-2 focus:ring-blue-500 outline-none" min="1" />
					</div>
				</div>
				<div class="px-6 py-4 border-t flex justify-end gap-3">
					<button onclick={() => showEdit = false} class="px-4 py-2 text-sm text-gray-600 hover:bg-gray-100 rounded-lg">انصراف</button>
					<button
						onclick={async () => {
							editLoading = true;
							await api.put(`/classes/${classId}`, { name: editName, description: editDesc, color: editColor, max_students: editMax });
							await loadData();
							showEdit = false;
							editLoading = false;
						}}
						disabled={editLoading}
						class="px-4 py-2 bg-blue-600 text-white text-sm rounded-lg font-medium hover:bg-blue-700 disabled:opacity-50"
					>ذخیره</button>
				</div>
			</div>
		</div>
	{/if}
{:else}
	<div class="text-center py-20">
		<p class="text-gray-500">کلاس یافت نشد</p>
		<a href="/classes" class="text-blue-600 text-sm mt-2 inline-block">بازگشت به کلاس‌ها</a>
	</div>
{/if}

<ConfirmModal bind:show={showDeleteConfirm} title="حذف کلاس" message="آیا از حذف این کلاس اطمینان دارید؟" onConfirm={deleteClass} onCancel={() => {}} />
