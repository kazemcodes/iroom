<script lang="ts">
	import { api } from '$lib/api';
	import { onMount } from 'svelte';
	import type { Class, Session, User } from '$lib/types';

	let classes = $state<Class[]>([]);
	let teachers = $state<Record<number, User>>({});
	let sessions = $state<Session[]>([]);
	let loading = $state(true);

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
		return 0;
	}

	function formatDate(d: string) {
		if (!d) return '';
		return new Date(d).toLocaleDateString('fa-IR', { year: 'numeric', month: 'long', day: 'numeric' });
	}
</script>

<div class="space-y-6">
	<div>
		<h1 class="text-2xl font-bold text-gray-900">مدیریت اتاق‌ها</h1>
		<p class="text-gray-500 mt-1">{classes.length} کلاس</p>
	</div>

	{#if loading}
		<div class="flex items-center justify-center py-20">
			<div class="animate-spin h-8 w-8 border-4 border-blue-600 border-t-transparent rounded-full"></div>
		</div>
	{:else if classes.length === 0}
		<div class="text-center py-20 bg-white rounded-xl border">
			<p class="text-gray-500">کلاسی یافت نشد</p>
		</div>
	{:else}
		<div class="grid gap-4">
			{#each classes as cls}
				{@const activeSessions = getActiveSessions(cls.id)}
				{@const sessionCount = getSessionsCount(cls.id)}
				<div class="bg-white border border-gray-200 rounded-xl p-5 hover:shadow-sm transition-shadow">
					<div class="flex items-center justify-between">
						<div class="flex items-center gap-4">
							<div class="w-12 h-12 rounded-xl flex items-center justify-center text-white font-bold text-lg" style="background: {cls.color || 'linear-gradient(135deg, #1a56db, #7c3aed)'}">
								{cls.name.charAt(0)}
							</div>
							<div>
								<h3 class="font-bold text-gray-900">{cls.name}</h3>
								<p class="text-sm text-gray-500 mt-0.5">
									مدرس: {teachers[cls.teacher_id]?.display_name || '—'}
									• {sessionCount} جلسه
									{#if activeSessions.length > 0}
										• <span class="text-green-600 font-medium">{activeSessions.length} زنده</span>
									{/if}
								</p>
								<p class="text-xs text-gray-400 mt-0.5">حداکثر {cls.max_students} دانش‌آموز • {formatDate(cls.created_at)}</p>
							</div>
						</div>
						<div class="flex items-center gap-2">
							<a href="/classes/{cls.id}" class="px-3 py-1.5 text-xs text-gray-600 border border-gray-200 rounded-lg hover:bg-gray-50 transition-colors font-medium">
								جزئیات
							</a>
						</div>
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>
