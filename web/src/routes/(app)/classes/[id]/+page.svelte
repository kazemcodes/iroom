<script lang="ts">
	import { page } from '$app/state';
	import { auth, isAdmin, isTeacher } from '$lib/stores';
	import { api } from '$lib/api';
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import type { Class, Session, User, Announcement, RecurringSession } from '$lib/types';
	import ConfirmModal from '$lib/components/ConfirmModal.svelte';
	import JalaliDatePicker from '$lib/components/JalaliDatePicker.svelte';
	import { classroomWindow } from '$lib/classroom/ClassroomWindow';
	import { toPersianNum, toPersianDate, toPersianDateTime } from '$lib/utils/persian';

	let classData = $state<Class | null>(null);
	let sessions = $state<Session[]>([]);
	let students = $state<User[]>([]);
	let announcements = $state<Announcement[]>([]);
	let loading = $state(true);
	let activeTab = $state<'sessions' | 'students' | 'announcements'>('sessions');

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

	// Announcements
	let showCreateAnnouncement = $state(false);
	let showEditAnnouncement = $state(false);
	let editingAnnouncement = $state<Announcement | null>(null);
	let announcementTitle = $state('');
	let announcementContent = $state('');
	let announcementPinned = $state(false);
	let announcementLoading = $state(false);
	let announcementError = $state('');
	let showDeleteAnnouncementConfirm = $state(false);
	let deletingAnnouncementId = $state<number | null>(null);

	// Recurring sessions
	let recurringSessions = $state<RecurringSession[]>([]);
	let showCreateRecurring = $state(false);
	let recurringTitle = $state('');
	let recurringDayOfWeek = $state(0);
	let recurringTime = $state('09:00');
	let recurringDuration = $state(60);
	let recurringWeekCount = $state(4);
	let recurringLoading = $state(false);
	let recurringError = $state('');
	let showDeleteRecurringConfirm = $state(false);
	let deletingRecurringId = $state<number | null>(null);

	// Invite code
	let inviteCode = $state('');
	let joinLink = $state('');
	let showRegenerateConfirm = $state(false);
	let copySuccess = $state(false);

	const persianDays = ['شنبه', 'یکشنبه', 'دوشنبه', 'سه‌شنبه', 'چهارشنبه', 'پنج‌شنبه', 'جمعه'];

	const classId = $derived(page.params.id as string);

	onMount(() => loadData());

	async function loadData() {
		loading = true;
		const [classRes, sessionsRes, studentsRes, announcementsRes, recurringRes] = await Promise.all([
			api.get<Class>(`/classes/${classId}`),
			api.get<Session[]>(`/sessions?class_id=${classId}`),
			api.get<User[]>(`/classes/${classId}/students`),
			api.get<{ items: Announcement[] }>(`/classes/${classId}/announcements`),
			api.get<RecurringSession[]>(`/sessions/recurring?class_id=${classId}`)
		]);

		if (classRes.success) {
			classData = classRes.data!;
			inviteCode = classRes.data!.invite_code || '';
			joinLink = `${window.location.origin}/join/${inviteCode}`;
		}
		if (sessionsRes.success && sessionsRes.data) sessions = Array.isArray(sessionsRes.data) ? sessionsRes.data : [];
		if (studentsRes.success && studentsRes.data) students = Array.isArray(studentsRes.data) ? studentsRes.data : [];
		if (announcementsRes.success && announcementsRes.data) {
			announcements = Array.isArray(announcementsRes.data.items) ? announcementsRes.data.items : [];
		}
		if (recurringRes.success && recurringRes.data) {
			recurringSessions = Array.isArray(recurringRes.data) ? recurringRes.data : [];
		}
		loading = false;
	}

	async function createSession() {
		if (!sessionTitle?.trim()) {
			sessionError = 'عنوان الزامی است';
			return;
		}
		if (!sessionDate) {
			sessionError = 'تاریخ را انتخاب کنید';
			return;
		}
		if (!sessionTime) {
			sessionError = 'ساعت را انتخاب کنید';
			return;
		}
		sessionLoading = true;
		sessionError = '';
		const dt = `${sessionDate}T${sessionTime}:00`;
		const parsed = new Date(dt);
		if (isNaN(parsed.getTime())) {
			sessionError = `تاریخ نامعتبر: ${sessionDate} ${sessionTime}`;
			sessionLoading = false;
			return;
		}
		const res = await api.post('/sessions', {
			class_id: parseInt(classId),
			title: sessionTitle,
			scheduled_at: parsed.toISOString(),
			duration: sessionDuration
		});

		if (!res.success) {
			sessionError = res.error || 'خطا در ایجاد جلسه';
			sessionLoading = false;
			return;
		}

		sessions = [res.data! as Session, ...sessions];
		showCreateSession = false;
		sessionTitle = '';
		sessionDate = '';
		sessionTime = '';
		sessionDuration = 60;
		sessionLoading = false;
	}

	async function createRecurringSession() {
		recurringLoading = true;
		recurringError = '';
		const res = await api.post<RecurringSession>('/sessions/recurring', {
			class_id: parseInt(classId),
			title: recurringTitle,
			day_of_week: recurringDayOfWeek,
			time: recurringTime,
			duration: recurringDuration,
			week_count: recurringWeekCount
		});

		if (!res.success) {
			recurringError = res.error || 'خطا در ایجاد جلسه تکرارشونده';
			recurringLoading = false;
			return;
		}

		recurringSessions = [res.data!, ...recurringSessions];
		showCreateRecurring = false;
		recurringTitle = '';
		recurringDayOfWeek = 0;
		recurringTime = '09:00';
		recurringDuration = 60;
		recurringWeekCount = 4;
		recurringLoading = false;
	}

	async function deleteRecurringSession() {
		if (!deletingRecurringId) return;
		const res = await api.delete(`/sessions/recurring/${deletingRecurringId}`);
		if (res.success) {
			recurringSessions = recurringSessions.filter(r => r.id !== deletingRecurringId);
		}
		showDeleteRecurringConfirm = false;
		deletingRecurringId = null;
	}

	// Generate preview dates for recurring session
	function generatePreviewDates() {
		const dates: Date[] = [];
		const today = new Date();
		const currentDay = today.getDay(); // 0=Sunday, 1=Monday, ...
		
		// Convert Persian day (0=Saturday) to JS day (0=Sunday)
		// Persian: 0=شنبه(Sat), 1=یکشنبه(Sun), ..., 6=جمعه(Fri)
		// JS: 0=Sun, 1=Mon, ..., 6=Sat
		const targetDay = recurringDayOfWeek === 0 ? 6 : recurringDayOfWeek - 1;
		
		let daysUntilTarget = targetDay - currentDay;
		if (daysUntilTarget < 0) daysUntilTarget += 7;
		
		const firstDate = new Date(today);
		firstDate.setDate(today.getDate() + daysUntilTarget);
		
		for (let i = 0; i < recurringWeekCount; i++) {
			const date = new Date(firstDate);
			date.setDate(firstDate.getDate() + i * 7);
			dates.push(date);
		}
		
		return dates;
	}

	async function startSession(id: number) {
		const res = await api.post(`/sessions/${id}/start`);
		if (res.success) {
			const data = res.data as any;
			sessions = sessions.map(s => s.id === id ? { ...s, status: 'live', livekit_room: data?.livekit_room || '' } : s);
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

	// Announcement functions
	async function createAnnouncement() {
		announcementLoading = true;
		announcementError = '';
		const res = await api.post<Announcement>(`/classes/${classId}/announcements`, {
			title: announcementTitle,
			content: announcementContent,
			is_pinned: announcementPinned
		});

		if (!res.success) {
			announcementError = res.error || 'خطا در ایجاد اعلان';
			announcementLoading = false;
			return;
		}

		announcements = [res.data!, ...announcements];
		showCreateAnnouncement = false;
		announcementTitle = '';
		announcementContent = '';
		announcementPinned = false;
		announcementLoading = false;
	}

	async function updateAnnouncement() {
		if (!editingAnnouncement) return;
		announcementLoading = true;
		announcementError = '';
		const res = await api.put<Announcement>(`/announcements/${editingAnnouncement.id}`, {
			title: announcementTitle,
			content: announcementContent
		});

		if (!res.success) {
			announcementError = res.error || 'خطا در ویرایش اعلان';
			announcementLoading = false;
			return;
		}

		announcements = announcements.map(a => a.id === editingAnnouncement?.id ? { ...a, title: announcementTitle, content: announcementContent } : a);
		showEditAnnouncement = false;
		editingAnnouncement = null;
		announcementTitle = '';
		announcementContent = '';
		announcementLoading = false;
	}

	async function deleteAnnouncement() {
		if (!deletingAnnouncementId) return;
		const res = await api.delete(`/announcements/${deletingAnnouncementId}`);
		if (res.success) {
			announcements = announcements.filter(a => a.id !== deletingAnnouncementId);
		}
		showDeleteAnnouncementConfirm = false;
		deletingAnnouncementId = null;
	}

	async function togglePin(announcement: Announcement) {
		const res = await api.post(`/announcements/${announcement.id}/pin`);
		if (res.success) {
			announcements = announcements.map(a => 
				a.id === announcement.id ? { ...a, is_pinned: !a.is_pinned } : a
			);
		}
	}

	// Invite code functions
	async function copyToClipboard(text: string) {
		try {
			await navigator.clipboard.writeText(text);
			copySuccess = true;
			setTimeout(() => copySuccess = false, 2000);
		} catch (err) {
			console.error('Failed to copy:', err);
		}
	}

	async function regenerateCode() {
		const res = await api.post<{ invite_code: string }>(`/classes/${classId}/regenerate-code`);
		if (res.success && res.data) {
			inviteCode = res.data.invite_code;
			joinLink = `${window.location.origin}/join/${inviteCode}`;
		}
		showRegenerateConfirm = false;
	}

	function openEditAnnouncement(announcement: Announcement) {
		editingAnnouncement = announcement;
		announcementTitle = announcement.title;
		announcementContent = announcement.content;
		showEditAnnouncement = true;
	}

	const sortedAnnouncements = $derived([...announcements].sort((a, b) => {
		if (a.is_pinned && !b.is_pinned) return -1;
		if (!a.is_pinned && b.is_pinned) return 1;
		return new Date(b.created_at).getTime() - new Date(a.created_at).getTime();
	}));

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
		<div class="bg-white rounded-xl p-6">
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

		<!-- Invite Code Section (Teacher/Admin only) -->
		{#if ($isAdmin || $isTeacher) && inviteCode}
			<div class="bg-white rounded-xl p-5">
				<h3 class="font-semibold text-gray-900 mb-4 flex items-center gap-2">
					<svg class="w-5 h-5 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z" />
					</svg>
					کد دعوت
				</h3>
				<div class="space-y-4">
					<!-- Invite Code -->
					<div>
						<label class="block text-sm text-gray-500 mb-1">کد دعوت</label>
						<div class="flex items-center gap-2">
							<code class="flex-1 px-4 py-2.5 bg-gray-50 border rounded-lg font-mono text-lg tracking-wider text-center select-all">{inviteCode}</code>
							<button
								onclick={() => copyToClipboard(inviteCode)}
								class="px-4 py-2.5 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors flex items-center gap-2 shrink-0"
								title="کپی کد"
							>
								{#if copySuccess}
									<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
									</svg>
								{:else}
									<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
									</svg>
								{/if}
								کپی
							</button>
						</div>
					</div>

					<!-- Join Link -->
					<div>
						<label class="block text-sm text-gray-500 mb-1">لینک پیوستن</label>
						<div class="flex items-center gap-2">
							<input
								type="text"
								readonly
								value={joinLink}
								class="flex-1 px-4 py-2.5 bg-gray-50 border rounded-lg text-sm text-gray-600 select-all"
							/>
							<button
								onclick={() => copyToClipboard(joinLink)}
								class="px-4 py-2.5 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors flex items-center gap-2 shrink-0"
								title="کپی لینک"
							>
								<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
								</svg>
								کپی لینک
							</button>
						</div>
					</div>

					<!-- Regenerate Button -->
					<div class="pt-2 border-t">
						<button
							onclick={() => showRegenerateConfirm = true}
							class="px-4 py-2 text-sm text-orange-600 hover:bg-orange-50 rounded-lg transition-colors flex items-center gap-2"
						>
							<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
							</svg>
							تولید مجدد کد
						</button>
						<p class="text-xs text-gray-400 mt-1">توجه: تولید مجدد کد، کد قبلی را غیرفعال می‌کند</p>
					</div>
				</div>
			</div>
		{/if}

		<!-- Tabs -->
		<div class="flex gap-1 bg-gray-100 p-1 rounded-lg w-fit">
			<button
				class="px-4 py-2 rounded-md text-sm font-medium transition-all {activeTab === 'sessions' ? 'bg-white text-blue-600 shadow-sm' : 'text-gray-500 hover:text-gray-700'}"
				onclick={() => activeTab = 'sessions'}
			>جلسات ({toPersianNum(sessions.length)})</button>
			<button
				class="px-4 py-2 rounded-md text-sm font-medium transition-all {activeTab === 'students' ? 'bg-white text-blue-600 shadow-sm' : 'text-gray-500 hover:text-gray-700'}"
				onclick={() => activeTab = 'students'}
			>دانش‌آموزان ({toPersianNum(students.length)})</button>
			<button
				class="px-4 py-2 rounded-md text-sm font-medium transition-all {activeTab === 'announcements' ? 'bg-white text-blue-600 shadow-sm' : 'text-gray-500 hover:text-gray-700'}"
				onclick={() => activeTab = 'announcements'}
			>اعلان‌ها ({toPersianNum(announcements.length)})</button>
		</div>

		{#if activeTab === 'sessions'}
			<div class="flex justify-end gap-2">
				{#if $isAdmin || $isTeacher}
					<button onclick={() => showCreateRecurring = true} class="px-4 py-2.5 bg-purple-600 text-white rounded-lg text-sm font-medium hover:bg-purple-700 transition-colors flex items-center gap-2">
						<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" /></svg>
						جلسه تکرارشونده
					</button>
					<button onclick={() => showCreateSession = true} class="px-4 py-2.5 bg-blue-600 text-white rounded-lg text-sm font-medium hover:bg-blue-700 transition-colors flex items-center gap-2">
						<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" /></svg>
						جلسه جدید
					</button>
				{/if}
			</div>

			<!-- Recurring Sessions Section -->
			{#if recurringSessions.length > 0}
				<div class="bg-white rounded-xl p-4">
					<h3 class="font-semibold text-gray-900 mb-4 flex items-center gap-2">
						<svg class="w-5 h-5 text-purple-600" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" /></svg>
						جلسات تکرارشونده
					</h3>
					<div class="space-y-3">
						{#each recurringSessions as template (template.id)}
							<div class="flex items-center justify-between p-3 bg-purple-50 rounded-lg border border-purple-100">
								<div class="flex items-center gap-3">
									<div class="w-10 h-10 rounded-lg bg-purple-100 flex items-center justify-center">
										<svg class="w-5 h-5 text-purple-600" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" /></svg>
									</div>
									<div>
										<p class="font-medium text-gray-900">{template.title}</p>
										<p class="text-sm text-gray-500">
											{persianDays[template.day_of_week]} ساعت {template.time} • {toPersianNum(template.duration)} دقیقه • {toPersianNum(template.week_count)} هفته
										</p>
									</div>
								</div>
								<div class="flex items-center gap-3">
									<span class="text-xs px-2.5 py-1 rounded-full font-medium bg-purple-100 text-purple-700">
										{toPersianNum(template.sessions_generated)} جلسه ایجاد شده
									</span>
									{#if $isAdmin || $isTeacher}
										<button
											onclick={() => { deletingRecurringId = template.id; showDeleteRecurringConfirm = true; }}
											class="p-2 rounded-lg hover:bg-red-50 transition-colors"
											title="حذف"
										>
											<svg class="w-4 h-4 text-red-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
											</svg>
										</button>
									{/if}
								</div>
							</div>
						{/each}
					</div>
				</div>
			{/if}

			{#if sessions.length === 0}
				<div class="text-center py-12 bg-white rounded-xl">
					<p class="text-gray-500">هنوز جلسه‌ای ایجاد نشده</p>
				</div>
			{:else}
				<div class="space-y-3">
					{#each sessions as s}
						<div class="bg-white rounded-xl p-4 flex items-center justify-between">
							<div class="flex items-center gap-4">
								<div class="w-10 h-10 rounded-lg flex items-center justify-center {s.status === 'live' ? 'bg-green-100' : 'bg-gray-100'}">
									<svg class="w-5 h-5 {s.status === 'live' ? 'text-green-600' : 'text-gray-400'}" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" />
									</svg>
								</div>
								<div>
									<p class="font-medium text-gray-900">{s.title}</p>
									<p class="text-sm text-gray-500">{toPersianDateTime(s.scheduled_at)} • {toPersianNum(s.duration)} دقیقه</p>
								</div>
							</div>
							<div class="flex items-center gap-3">
								<span class="text-xs px-2.5 py-1 rounded-full font-medium {statusColors[s.status]}">{statusLabels[s.status]}</span>
								{#if ($isAdmin || $isTeacher) && s.status === 'scheduled'}
									<button onclick={() => startSession(s.id)} class="px-3 py-1.5 bg-green-600 text-white text-xs rounded-lg hover:bg-green-700 transition-colors">شروع</button>
								{/if}
							{#if ($isAdmin || $isTeacher) && s.status === 'live'}
									<button onclick={() => classroomWindow.open(String(s.id), s.title)} class="px-3 py-1.5 bg-blue-600 text-white text-xs rounded-lg hover:bg-blue-700 transition-colors">ورود</button>
									<button onclick={() => endSession(s.id)} class="px-3 py-1.5 bg-red-600 text-white text-xs rounded-lg hover:bg-red-700 transition-colors">پایان</button>
								{/if}
								{#if s.status === 'live'}
									<button onclick={() => classroomWindow.open(String(s.id), s.title)} class="px-3 py-1.5 bg-green-600 text-white text-xs rounded-lg hover:bg-green-700 transition-colors">پیوستن</button>
								{/if}
							</div>
						</div>
					{/each}
				</div>
			{/if}
		{:else if activeTab === 'students'}
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
				<div class="text-center py-12 bg-white rounded-xl">
					<p class="text-gray-500">هنوز دانش‌آموزی ثبت‌نام نکرده</p>
				</div>
			{:else}
				<div class="bg-white rounded-xl divide-y">
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
		{:else}
			<!-- Announcements tab -->
			<div class="flex justify-end">
				{#if $isAdmin || $isTeacher}
					<button onclick={() => showCreateAnnouncement = true} class="px-4 py-2.5 bg-blue-600 text-white rounded-lg text-sm font-medium hover:bg-blue-700 transition-colors flex items-center gap-2">
						<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" /></svg>
						اعلان جدید
					</button>
				{/if}
			</div>

			{#if announcements.length === 0}
				<div class="text-center py-12 bg-white rounded-xl">
					<p class="text-gray-500">هنوز اعلانی ایجاد نشده</p>
				</div>
			{:else}
				<div class="space-y-3">
					{#each sortedAnnouncements as announcement}
						<div class="bg-white rounded-xl p-4 {announcement.is_pinned ? 'border-blue-200 bg-blue-50/30' : ''}">
							<div class="flex items-start justify-between gap-4">
								<div class="flex-1 min-w-0">
									<div class="flex items-center gap-2 mb-2">
										{#if announcement.is_pinned}
											<span class="text-blue-600" title="پین شده">📌</span>
										{/if}
										<h3 class="font-semibold text-gray-900 truncate">{announcement.title}</h3>
										{#if !announcement.is_read}
											<span class="px-2 py-0.5 bg-red-100 text-red-600 text-xs rounded-full font-medium">جدید</span>
										{/if}
									</div>
									<p class="text-gray-600 text-sm whitespace-pre-wrap mb-3">{announcement.content}</p>
									<div class="flex items-center gap-4 text-xs text-gray-500">
										<span>{announcement.author_name || 'مدیر'}</span>
										<span>{toPersianDate(announcement.created_at)}</span>
									</div>
								</div>
								{#if $isAdmin || $isTeacher}
									<div class="flex items-center gap-1 shrink-0">
										<button
											onclick={() => togglePin(announcement)}
											class="p-2 rounded-lg hover:bg-gray-100 transition-colors"
											title={announcement.is_pinned ? 'برداشتن پین' : 'پین کردน'}
										>
											<svg class="w-4 h-4 {announcement.is_pinned ? 'text-blue-600' : 'text-gray-400'}" fill="currentColor" viewBox="0 0 24 24">
												<path d="M16 12V4h1V2H7v2h1v8l-2 2v2h5.2v6h1.6v-6H18v-2l-2-2z"/>
											</svg>
										</button>
										<button
											onclick={() => openEditAnnouncement(announcement)}
											class="p-2 rounded-lg hover:bg-gray-100 transition-colors"
											title="ویرایش"
										>
											<svg class="w-4 h-4 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
											</svg>
										</button>
										<button
											onclick={() => { deletingAnnouncementId = announcement.id; showDeleteAnnouncementConfirm = true; }}
											class="p-2 rounded-lg hover:bg-red-50 transition-colors"
											title="حذف"
										>
											<svg class="w-4 h-4 text-red-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
											</svg>
										</button>
									</div>
								{/if}
							</div>
						</div>
					{/each}
				</div>
			{/if}
		{/if}
	</div>

	<!-- Create Session Modal -->
	{#if showCreateSession}
		<div class="modal-overlay" role="button" tabindex="-1" onclick={() => showCreateSession = false}>
			<div class="modal-content" onclick={(e) => e.stopPropagation()}>
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
							<JalaliDatePicker bind:value={sessionDate} label="تاریخ" />
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

	<!-- Create Recurring Session Modal -->
	{#if showCreateRecurring}
		<div class="modal-overlay" role="button" tabindex="-1" onclick={() => showCreateRecurring = false}>
			<div class="modal-content" onclick={(e) => e.stopPropagation()}>
				<div class="px-6 py-4 border-b"><h2 class="font-bold text-lg">جلسه تکرارشونده</h2></div>
				<div class="px-6 py-4 space-y-4">
					{#if recurringError}
						<div class="p-3 bg-red-50 text-red-600 rounded-lg text-sm">{recurringError}</div>
					{/if}
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">عنوان</label>
						<input type="text" bind:value={recurringTitle} class="w-full px-4 py-2.5 border rounded-lg focus:ring-2 focus:ring-purple-500 focus:border-transparent outline-none text-sm" placeholder="عنوان جلسه" required />
					</div>
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">روز هفته</label>
						<select bind:value={recurringDayOfWeek} class="w-full px-4 py-2.5 border rounded-lg focus:ring-2 focus:ring-purple-500 focus:border-transparent outline-none text-sm">
							{#each persianDays as day, i}
								<option value={i}>{day}</option>
							{/each}
						</select>
					</div>
					<div class="grid grid-cols-2 gap-3">
						<div>
							<label class="block text-sm font-medium text-gray-700 mb-1">ساعت</label>
							<input type="time" bind:value={recurringTime} class="w-full px-4 py-2.5 border rounded-lg focus:ring-2 focus:ring-purple-500 focus:border-transparent outline-none text-sm" required />
						</div>
						<div>
							<label class="block text-sm font-medium text-gray-700 mb-1">مدت (دقیقه)</label>
							<input type="number" bind:value={recurringDuration} class="w-full px-4 py-2.5 border rounded-lg focus:ring-2 focus:ring-purple-500 focus:border-transparent outline-none text-sm" min="15" max="480" />
						</div>
					</div>
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">تعداد هفته‌ها</label>
						<input type="number" bind:value={recurringWeekCount} class="w-full px-4 py-2.5 border rounded-lg focus:ring-2 focus:ring-purple-500 focus:border-transparent outline-none text-sm" min="1" max="52" />
					</div>
					<!-- Preview -->
					<div class="bg-purple-50 rounded-lg p-3">
						<p class="text-sm font-medium text-purple-700 mb-2">پیش‌نمایش تاریخ‌ها:</p>
						<div class="space-y-1 max-h-32 overflow-y-auto">
							{#each generatePreviewDates() as date, i}
								<p class="text-sm text-purple-600">{toPersianDate(date)} ({persianDays[recurringDayOfWeek]})</p>
							{/each}
						</div>
					</div>
				</div>
				<div class="px-6 py-4 border-t flex justify-end gap-3">
					<button onclick={() => showCreateRecurring = false} class="px-4 py-2 text-sm text-gray-600 hover:bg-gray-100 rounded-lg">انصراف</button>
					<button onclick={createRecurringSession} disabled={recurringLoading || !recurringTitle} class="px-4 py-2 bg-purple-600 text-white text-sm rounded-lg font-medium hover:bg-purple-700 disabled:opacity-50">
						{recurringLoading ? 'در حال ایجاد...' : 'ایجاد جلسه تکرارشونده'}
					</button>
				</div>
			</div>
		</div>
	{/if}

	<!-- Enroll Modal -->
	{#if showEnroll}
		<div class="modal-overlay" role="button" tabindex="-1" onclick={() => showEnroll = false}>
			<div class="modal-content" onclick={(e) => e.stopPropagation()}>
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
		<div class="modal-overlay" role="button" tabindex="-1" onclick={() => showEdit = false}>
			<div class="modal-content" onclick={(e) => e.stopPropagation()}>
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

	<!-- Create Announcement Modal -->
	{#if showCreateAnnouncement}
		<div class="modal-overlay" role="button" tabindex="-1" onclick={() => showCreateAnnouncement = false}>
			<div class="modal-content" onclick={(e) => e.stopPropagation()}>
				<div class="px-6 py-4 border-b"><h2 class="font-bold text-lg">اعلان جدید</h2></div>
				<div class="px-6 py-4 space-y-4">
					{#if announcementError}
						<div class="p-3 bg-red-50 text-red-600 rounded-lg text-sm">{announcementError}</div>
					{/if}
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">عنوان</label>
						<input type="text" bind:value={announcementTitle} class="w-full px-4 py-2.5 border rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none text-sm" placeholder="عنوان اعلان" required />
					</div>
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">محتوا</label>
						<textarea bind:value={announcementContent} class="w-full px-4 py-2.5 border rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none text-sm resize-none" rows="4" placeholder="متن اعلان" required></textarea>
					</div>
					<label class="flex items-center gap-2 cursor-pointer">
						<input type="checkbox" bind:checked={announcementPinned} class="w-4 h-4 text-blue-600 rounded focus:ring-blue-500" />
						<span class="text-sm text-gray-700">پین کردن اعلان</span>
					</label>
				</div>
				<div class="px-6 py-4 border-t flex justify-end gap-3">
					<button onclick={() => showCreateAnnouncement = false} class="px-4 py-2 text-sm text-gray-600 hover:bg-gray-100 rounded-lg">انصراف</button>
					<button onclick={createAnnouncement} disabled={announcementLoading || !announcementTitle || !announcementContent} class="px-4 py-2 bg-blue-600 text-white text-sm rounded-lg font-medium hover:bg-blue-700 disabled:opacity-50">
						{announcementLoading ? 'در حال ایجاد...' : 'ایجاد اعلان'}
					</button>
				</div>
			</div>
		</div>
	{/if}

	<!-- Edit Announcement Modal -->
	{#if showEditAnnouncement}
		<div class="modal-overlay" role="button" tabindex="-1" onclick={() => showEditAnnouncement = false}>
			<div class="modal-content" onclick={(e) => e.stopPropagation()}>
				<div class="px-6 py-4 border-b"><h2 class="font-bold text-lg">ویرایش اعلان</h2></div>
				<div class="px-6 py-4 space-y-4">
					{#if announcementError}
						<div class="p-3 bg-red-50 text-red-600 rounded-lg text-sm">{announcementError}</div>
					{/if}
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">عنوان</label>
						<input type="text" bind:value={announcementTitle} class="w-full px-4 py-2.5 border rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none text-sm" placeholder="عنوان اعلان" required />
					</div>
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">محتوا</label>
						<textarea bind:value={announcementContent} class="w-full px-4 py-2.5 border rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none text-sm resize-none" rows="4" placeholder="متن اعلان" required></textarea>
					</div>
				</div>
				<div class="px-6 py-4 border-t flex justify-end gap-3">
					<button onclick={() => showEditAnnouncement = false} class="px-4 py-2 text-sm text-gray-600 hover:bg-gray-100 rounded-lg">انصراف</button>
					<button onclick={updateAnnouncement} disabled={announcementLoading || !announcementTitle || !announcementContent} class="px-4 py-2 bg-blue-600 text-white text-sm rounded-lg font-medium hover:bg-blue-700 disabled:opacity-50">
						{announcementLoading ? 'در حال ذخیره...' : 'ذخیره تغییرات'}
					</button>
				</div>
			</div>
		</div>
	{/if}

<ConfirmModal bind:show={showDeleteConfirm} title="حذف کلاس" message="آیا از حذف این کلاس اطمینان دارید؟" onConfirm={deleteClass} onCancel={() => {}} />
<ConfirmModal bind:show={showDeleteAnnouncementConfirm} title="حذف اعلان" message="آیا از حذف این اعلان اطمینان دارید؟" onConfirm={deleteAnnouncement} onCancel={() => { deletingAnnouncementId = null; }} />
<ConfirmModal bind:show={showDeleteRecurringConfirm} title="حذف جلسه تکرارشونده" message="آیا از حذف این الگوی جلسه تکرارشونده اطمینان دارید؟" onConfirm={deleteRecurringSession} onCancel={() => { deletingRecurringId = null; }} />
<ConfirmModal bind:show={showRegenerateConfirm} title="تولید مجدد کد دعوت" message="آیا از تولید مجدد کد دعوت اطمینان دارید؟ کد قبلی دیگر قابل استفاده نخواهد بود." onConfirm={regenerateCode} onCancel={() => {}} />
