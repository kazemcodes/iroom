<script lang="ts">
	import { auth, isAdmin } from '$lib/stores';
	import { api } from '$lib/api';
	import { onMount } from 'svelte';
	import type { Class, Session } from '$lib/types';

	let stats = $state({ users: 0, classes: 0, sessions: 0, messages: 0 });
	let classes = $state<Class[]>([]);
	let sessions = $state<Session[]>([]);
	let loading = $state(true);

	onMount(async () => {
		await loadData();
	});

	async function loadData() {
		loading = true;

		if ($isAdmin) {
			const statsRes = await api.get<{ users: number; classes: number; sessions: number; messages: number }>('/admin/dashboard/stats');
			if (statsRes.success) stats = statsRes.data!;

			const classesRes = await api.get<Class[]>('/classes');
			if (classesRes.success && classesRes.data) classes = Array.isArray(classesRes.data) ? classesRes.data : [];

			const sessionsRes = await api.get<Session[]>('/sessions');
			if (sessionsRes.success && sessionsRes.data) sessions = Array.isArray(sessionsRes.data) ? sessionsRes.data : [];
		} else {
			const classesRes = await api.get<Class[]>('/classes');
			if (classesRes.success && classesRes.data) classes = Array.isArray(classesRes.data) ? classesRes.data : [];
		}

		loading = false;
	}

	function formatDate(dateStr: string) {
		if (!dateStr) return '';
		const d = new Date(dateStr);
		return d.toLocaleDateString('fa-IR', { year: 'numeric', month: 'long', day: 'numeric' });
	}

	function formatTime(dateStr: string) {
		if (!dateStr) return '';
		const d = new Date(dateStr);
		return d.toLocaleTimeString('fa-IR', { hour: '2-digit', minute: '2-digit' });
	}

	const statusLabels: Record<string, string> = {
		scheduled: 'برنامه‌ریزی شده',
		live: 'در حال برگزاری',
		ended: 'پایان یافته'
	};

	const statusColors: Record<string, string> = {
		scheduled: 'bg-blue-100 text-blue-700',
		live: 'bg-green-100 text-green-700',
		ended: 'bg-gray-100 text-gray-500'
	};
</script>

<div class="space-y-6">
	<!-- Header -->
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-bold text-gray-900">داشبورد</h1>
			<p class="text-gray-500 mt-1">خوش آمدید، {$auth.user?.display_name}</p>
		</div>
		<div class="text-sm text-gray-400">
			{new Date().toLocaleDateString('fa-IR', { weekday: 'long', year: 'numeric', month: 'long', day: 'numeric' })}
		</div>
	</div>

	{#if loading}
		<div class="flex items-center justify-center py-20">
			<div class="animate-spin h-8 w-8 border-4 border-blue-600 border-t-transparent rounded-full"></div>
		</div>
	{:else}
		<!-- Stats cards (admin only) -->
		{#if $isAdmin}
			<div class="grid grid-cols-2 lg:grid-cols-4 gap-4">
				<div class="bg-white rounded-xl shadow-sm p-5 border">
					<div class="flex items-center justify-between mb-3">
						<div class="w-10 h-10 bg-blue-100 rounded-lg flex items-center justify-center">
							<svg class="w-5 h-5 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z" />
							</svg>
						</div>
					</div>
					<p class="text-2xl font-bold text-gray-900">{stats.users.toLocaleString('fa-IR')}</p>
					<p class="text-sm text-gray-500 mt-1">کاربر</p>
				</div>

				<div class="bg-white rounded-xl shadow-sm p-5 border">
					<div class="flex items-center justify-between mb-3">
						<div class="w-10 h-10 bg-purple-100 rounded-lg flex items-center justify-center">
							<svg class="w-5 h-5 text-purple-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253" />
							</svg>
						</div>
					</div>
					<p class="text-2xl font-bold text-gray-900">{stats.classes.toLocaleString('fa-IR')}</p>
					<p class="text-sm text-gray-500 mt-1">کلاس</p>
				</div>

				<div class="bg-white rounded-xl shadow-sm p-5 border">
					<div class="flex items-center justify-between mb-3">
						<div class="w-10 h-10 bg-green-100 rounded-lg flex items-center justify-center">
							<svg class="w-5 h-5 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" />
							</svg>
						</div>
					</div>
					<p class="text-2xl font-bold text-gray-900">{stats.sessions.toLocaleString('fa-IR')}</p>
					<p class="text-sm text-gray-500 mt-1">جلسه</p>
				</div>

				<div class="bg-white rounded-xl shadow-sm p-5 border">
					<div class="flex items-center justify-between mb-3">
						<div class="w-10 h-10 bg-amber-100 rounded-lg flex items-center justify-center">
							<svg class="w-5 h-5 text-amber-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
							</svg>
						</div>
					</div>
					<p class="text-2xl font-bold text-gray-900">{stats.messages.toLocaleString('fa-IR')}</p>
					<p class="text-sm text-gray-500 mt-1">پیام</p>
				</div>
			</div>
		{/if}

		<!-- Classes -->
		<div class="bg-white rounded-xl shadow-sm border">
			<div class="px-6 py-4 border-b flex items-center justify-between">
				<h2 class="font-bold text-gray-900">کلاس‌های من</h2>
				<a href="/classes" class="text-sm text-blue-600 hover:text-blue-700">مشاهده همه</a>
			</div>
			<div class="p-6">
				{#if classes.length === 0}
					<div class="text-center py-8">
						<svg class="w-12 h-12 text-gray-300 mx-auto mb-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1" d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253" />
						</svg>
						<p class="text-gray-500">هنوز کلاسی ایجاد نشده</p>
						<a href="/classes" class="inline-block mt-3 px-4 py-2 bg-blue-600 text-white text-sm rounded-lg hover:bg-blue-700 transition-colors">ایجاد کلاس</a>
					</div>
				{:else}
					<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
						{#each classes as cls}
							<a href="/classes/{cls.id}" class="block p-4 rounded-xl border hover:shadow-md transition-all group">
								<div class="flex items-center gap-3 mb-3">
									<div class="w-3 h-3 rounded-full" style="background-color: {cls.color}"></div>
									<h3 class="font-bold text-gray-900 group-hover:text-blue-600 transition-colors">{cls.name}</h3>
								</div>
								{#if cls.description}
									<p class="text-sm text-gray-500 line-clamp-2">{cls.description}</p>
								{/if}
								<div class="mt-3 flex items-center justify-between text-xs text-gray-400">
									<span>حداکثر {cls.max_students} نفر</span>
									<span>{formatDate(cls.created_at)}</span>
								</div>
							</a>
						{/each}
					</div>
				{/if}
			</div>
		</div>

		<!-- Recent sessions -->
		{#if sessions.length > 0}
			<div class="bg-white rounded-xl shadow-sm border">
				<div class="px-6 py-4 border-b flex items-center justify-between">
					<h2 class="font-bold text-gray-900">جلسات اخیر</h2>
					<a href="/sessions" class="text-sm text-blue-600 hover:text-blue-700">مشاهده همه</a>
				</div>
				<div class="divide-y">
					{#each sessions.slice(0, 5) as session}
						<a href="/sessions/{session.id}" class="px-6 py-4 flex items-center justify-between hover:bg-gray-50 transition-colors">
							<div class="flex items-center gap-4">
								<div class="w-10 h-10 rounded-lg flex items-center justify-center {session.status === 'live' ? 'bg-green-100' : 'bg-gray-100'}">
									<svg class="w-5 h-5 {session.status === 'live' ? 'text-green-600' : 'text-gray-400'}" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" />
									</svg>
								</div>
								<div>
									<p class="font-medium text-gray-900">{session.title}</p>
									<p class="text-sm text-gray-500">{formatDate(session.scheduled_at)} — {formatTime(session.scheduled_at)}</p>
								</div>
							</div>
							<span class="text-xs px-2.5 py-1 rounded-full font-medium {statusColors[session.status]}">
								{statusLabels[session.status]}
							</span>
						</a>
					{/each}
				</div>
			</div>
		{/if}
	{/if}
</div>
