<script lang="ts">
	import { auth, isAdmin } from '$lib/stores';
	import { api } from '$lib/api';
	import { onMount } from 'svelte';
	import type { Class, Session } from '$lib/types';

	let stats = $state({ users: 0, classes: 0, sessions: 0, messages: 0 });
	let classes = $state<Class[]>([]);
	let sessions = $state<Session[]>([]);
	let loading = $state(true);

	onMount(() => loadData());

	async function loadData() {
		loading = true;
		if ($isAdmin) {
			const statsRes = await api.get<{ users: number; classes: number; sessions: number; messages: number }>('/admin/dashboard/stats');
			if (statsRes.success) stats = statsRes.data!;
		}
		const classesRes = await api.get<Class[]>('/classes');
		if (classesRes.success && classesRes.data) classes = Array.isArray(classesRes.data) ? classesRes.data : [];
		const sessionsRes = await api.get<Session[]>('/sessions');
		if (sessionsRes.success && sessionsRes.data) sessions = Array.isArray(sessionsRes.data) ? sessionsRes.data : [];
		loading = false;
	}

	function formatDate(d: string) {
		if (!d) return '';
		return new Date(d).toLocaleDateString('fa-IR', { month: 'long', day: 'numeric' });
	}

	function formatTime(d: string) {
		if (!d) return '';
		return new Date(d).toLocaleTimeString('fa-IR', { hour: '2-digit', minute: '2-digit' });
	}

	const statusLabels: Record<string, string> = { scheduled: 'برنامه‌ریزی شده', live: 'در حال برگزاری', ended: 'پایان یافته' };
	const statusColors: Record<string, string> = { scheduled: 'bg-blue-50 text-blue-600', live: 'bg-green-50 text-green-600', ended: 'bg-gray-50 text-gray-500' };
	const statusDots: Record<string, string> = { scheduled: 'bg-blue-400', live: 'bg-green-500', ended: 'bg-gray-400' };
</script>

<div class="space-y-8">
	<!-- Header -->
	<div>
		<h1 class="text-2xl font-extrabold text-gray-900">سلام {$auth.user?.display_name} 👋</h1>
		<p class="text-gray-400 mt-1 font-medium">
			{new Date().toLocaleDateString('fa-IR', { weekday: 'long', year: 'numeric', month: 'long', day: 'numeric' })}
		</p>
	</div>

	{#if loading}
		<div class="flex items-center justify-center py-20">
			<div class="animate-spin h-8 w-8 border-4 border-blue-500 border-t-transparent rounded-full"></div>
		</div>
	{:else}
		<!-- Stats (admin only) -->
		{#if $isAdmin}
			<div class="grid grid-cols-2 lg:grid-cols-4 gap-4">
				{#each [
					{ label: 'کاربران', value: stats.users, color: 'from-blue-500 to-blue-600', icon: 'M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z' },
					{ label: 'کلاس‌ها', value: stats.classes, color: 'from-violet-500 to-purple-600', icon: 'M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253' },
					{ label: 'جلسات', value: stats.sessions, color: 'from-emerald-500 to-teal-600', icon: 'M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z' },
					{ label: 'پیام‌ها', value: stats.messages, color: 'from-amber-500 to-orange-500', icon: 'M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z' }
				] as stat}
					<div class="stat-card">
						<div class="flex items-center justify-between mb-3">
							<div class="w-10 h-10 rounded-xl flex items-center justify-center bg-gradient-to-br {stat.color} text-white shadow-lg shadow-{stat.color.split('-')[1]}-500/20">
								<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="1.5">
									<path stroke-linecap="round" stroke-linejoin="round" d={stat.icon} />
								</svg>
							</div>
						</div>
						<p class="text-2xl font-extrabold text-gray-900">{stat.value.toLocaleString('fa-IR')}</p>
						<p class="text-sm text-gray-400 font-medium mt-0.5">{stat.label}</p>
					</div>
				{/each}
			</div>
		{/if}

		<!-- Classes -->
		<div class="card">
			<div class="px-6 py-4 border-b border-gray-50 flex items-center justify-between">
				<h2 class="font-bold text-gray-900">کلاس‌های من</h2>
				<a href="/classes" class="text-sm text-blue-600 hover:text-blue-700 font-medium">مشاهده همه →</a>
			</div>
			<div class="p-6">
				{#if classes.length === 0}
					<div class="text-center py-10">
						<div class="w-16 h-16 rounded-2xl bg-gray-100 flex items-center justify-center mx-auto mb-4">
							<svg class="w-8 h-8 text-gray-300" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="1">
								<path stroke-linecap="round" stroke-linejoin="round" d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253" />
							</svg>
						</div>
						<p class="text-gray-500 font-medium">هنوز کلاسی ایجاد نشده</p>
						<a href="/classes" class="btn-primary inline-block mt-4">ایجاد کلاس</a>
					</div>
				{:else}
					<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
						{#each classes as cls}
							<a href="/classes/{cls.id}" class="block p-5 card group">
								<div class="flex items-center gap-3 mb-3">
									<div class="w-3 h-3 rounded-full ring-2 ring-white shadow-sm" style="background-color: {cls.color}"></div>
									<h3 class="font-bold text-gray-900 group-hover:text-blue-600 transition-colors">{cls.name}</h3>
								</div>
								{#if cls.description}
									<p class="text-sm text-gray-400 line-clamp-2 mb-3">{cls.description}</p>
								{/if}
								<div class="flex items-center justify-between text-xs text-gray-400 pt-3 border-t border-gray-50">
									<span class="flex items-center gap-1">
										<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="1.5"><path stroke-linecap="round" stroke-linejoin="round" d="M15 19.128a9.38 9.38 0 002.625.372 9.337 9.337 0 004.121-.952 4.125 4.125 0 00-7.533-2.493M15 19.128v-.003c0-1.113-.285-2.16-.786-3.07M15 19.128v.106A12.318 12.318 0 018.624 21c-2.331 0-4.512-.645-6.374-1.766l-.001-.109a6.375 6.375 0 0111.964-3.07M12 6.375a3.375 3.375 0 11-6.75 0 3.375 3.375 0 016.75 0zm8.25 2.25a2.625 2.625 0 11-5.25 0 2.625 2.625 0 015.25 0z" /></svg>
										حداکثر {cls.max_students} نفر
									</span>
								</div>
							</a>
						{/each}
					</div>
				{/if}
			</div>
		</div>

		<!-- Recent sessions -->
		{#if sessions.length > 0}
			<div class="card">
				<div class="px-6 py-4 border-b border-gray-50 flex items-center justify-between">
					<h2 class="font-bold text-gray-900">جلسات اخیر</h2>
					<a href="/sessions" class="text-sm text-blue-600 hover:text-blue-700 font-medium">مشاهده همه →</a>
				</div>
				<div class="divide-y divide-gray-50">
					{#each sessions.slice(0, 5) as session}
						<a href="/sessions/{session.id}" class="px-6 py-4 flex items-center justify-between hover:bg-gray-50/50 transition-colors">
							<div class="flex items-center gap-4">
								<div class="relative">
									<div class="w-10 h-10 rounded-xl flex items-center justify-center {session.status === 'live' ? 'bg-green-50' : 'bg-gray-50'}">
										<svg class="w-5 h-5 {session.status === 'live' ? 'text-green-500' : 'text-gray-400'}" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="1.5">
											<path stroke-linecap="round" stroke-linejoin="round" d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" />
										</svg>
									</div>
									{#if session.status === 'live'}
										<span class="absolute -top-0.5 -right-0.5 w-2.5 h-2.5 bg-green-500 rounded-full border-2 border-white animate-pulse"></span>
									{/if}
								</div>
								<div>
									<p class="font-semibold text-gray-900">{session.title}</p>
									<p class="text-sm text-gray-400">{formatDate(session.scheduled_at)} — {formatTime(session.scheduled_at)}</p>
								</div>
							</div>
							<span class="badge {statusColors[session.status]}">
								<span class="w-1.5 h-1.5 rounded-full {statusDots[session.status]} me-1.5"></span>
								{statusLabels[session.status]}
							</span>
						</a>
					{/each}
				</div>
			</div>
		{/if}
	{/if}
</div>
