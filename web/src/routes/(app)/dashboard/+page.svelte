<script lang="ts">
	import { auth, isAdmin } from '$lib/stores';
	import { api } from '$lib/api';
	import { onMount } from 'svelte';
	import type { Class, Session } from '$lib/types';
	import { toPersianNum, toPersianDate, toPersianDateTime } from '$lib/utils/persian';

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

	const statusLabels: Record<string, string> = { scheduled: 'برنامه‌ریزی شده', live: 'در حال برگزاری', ended: 'پایان یافته' };
	const statusClasses: Record<string, string> = { scheduled: 'badge-info', live: 'badge-success', ended: 'badge' };
	const statusDots: Record<string, string> = { scheduled: 'bg-blue-400', live: 'bg-green-500', ended: 'bg-gray-400' };
</script>

<div class="space-y-6">
	<div>
		<h1 class="text-2xl font-extrabold" style="color: var(--sr-text);">سلام {$auth.user?.display_name} 👋</h1>
		<p class="mt-1 font-medium" style="color: var(--sr-text-secondary);">
			{toPersianDate(new Date())} — {new Date().toLocaleDateString('fa-IR', { weekday: 'long' })}
		</p>
	</div>

	{#if loading}
		<div class="flex items-center justify-center py-20">
			<div class="animate-spin h-8 w-8 border-4 border-t-transparent rounded-full" style="border-color: var(--sr-border); border-top-color: var(--sr-primary);"></div>
		</div>
	{:else}
		{#if $isAdmin}
			<div class="grid grid-cols-2 lg:grid-cols-4 gap-4">
				{#each [
					{ label: 'کاربران', value: stats.users, color: '#3b82f6', icon: 'M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z' },
					{ label: 'کلاس‌ها', value: stats.classes, color: '#8b5cf6', icon: 'M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253' },
					{ label: 'جلسات', value: stats.sessions, color: '#10b981', icon: 'M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z' },
					{ label: 'پیام‌ها', value: stats.messages, color: '#f59e0b', icon: 'M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z' }
				] as stat}
					<div class="rounded-xl p-5" style="background: var(--sr-pure); border: 1px solid var(--sr-border);">
						<div class="flex items-center justify-between mb-3">
							<div class="w-10 h-10 rounded-xl flex items-center justify-center text-white" style="background: {stat.color};">
								<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="1.5">
									<path stroke-linecap="round" stroke-linejoin="round" d={stat.icon} />
								</svg>
							</div>
						</div>
						<p class="text-2xl font-extrabold" style="color: var(--sr-text);">{toPersianNum(stat.value)}</p>
						<p class="text-sm font-medium mt-0.5" style="color: var(--sr-text-secondary);">{stat.label}</p>
					</div>
				{/each}
			</div>
		{/if}

		<div class="rounded-xl p-0" style="background: var(--sr-pure); border: 1px solid var(--sr-border);">
			<div class="px-6 py-4 flex items-center justify-between" style="border-bottom: 1px solid var(--sr-border);">
				<h2 class="font-bold text-sm" style="color: var(--sr-text);">کلاس‌های من</h2>
				<a href="/classes" class="text-xs font-medium hover:underline" style="color: var(--sr-primary);">مشاهده همه →</a>
			</div>
			<div class="p-6">
				{#if classes.length === 0}
					<div class="text-center py-10">
						<div class="w-16 h-16 rounded-2xl flex items-center justify-center mx-auto mb-4" style="background: var(--sr-bg-alt);">
							<svg class="w-8 h-8" style="color: var(--sr-text-muted);" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="1">
								<path stroke-linecap="round" stroke-linejoin="round" d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253" />
							</svg>
						</div>
						<p class="font-medium" style="color: var(--sr-text-muted);">هنوز کلاسی ایجاد نشده</p>
						<a href="/classes" class="btn-primary inline-block mt-4">ایجاد کلاس</a>
					</div>
				{:else}
					<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
						{#each classes as cls}
							<a href="/classes/{cls.id}" class="block p-5 rounded-xl transition-all hover:shadow-sm group" style="background: var(--sr-bg-alt); border: 1px solid var(--sr-border);">
								<div class="flex items-center gap-3 mb-3">
									<div class="w-3 h-3 rounded-full shrink-0" style="background-color: {cls.color};"></div>
									<h3 class="font-bold text-sm group-hover:underline" style="color: var(--sr-text);">{cls.name}</h3>
								</div>
								{#if cls.description}
									<p class="text-xs line-clamp-2 mb-3" style="color: var(--sr-text-muted);">{cls.description}</p>
								{/if}
								<div class="flex items-center justify-between text-[10px] pt-3" style="color: var(--sr-text-muted); border-top: 1px solid var(--sr-border);">
									<span class="flex items-center gap-1">
										<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="1.5"><path stroke-linecap="round" stroke-linejoin="round" d="M15 19.128a9.38 9.38 0 002.625.372 9.337 9.337 0 004.121-.952 4.125 4.125 0 00-7.533-2.493M15 19.128v-.003c0-1.113-.285-2.16-.786-3.07M15 19.128v.106A12.318 12.318 0 018.624 21c-2.331 0-4.512-.645-6.374-1.766l-.001-.109a6.375 6.375 0 0111.964-3.07M12 6.375a3.375 3.375 0 11-6.75 0 3.375 3.375 0 016.75 0zm8.25 2.25a2.625 2.625 0 11-5.25 0 2.625 2.625 0 015.25 0z" /></svg>
										حداکثر {toPersianNum(cls.max_students)} نفر
									</span>
								</div>
							</a>
						{/each}
					</div>
				{/if}
			</div>
		</div>

		{#if sessions.length > 0}
			<div class="rounded-xl overflow-hidden" style="background: var(--sr-pure); border: 1px solid var(--sr-border);">
				<div class="px-6 py-4 flex items-center justify-between" style="border-bottom: 1px solid var(--sr-border);">
					<h2 class="font-bold text-sm" style="color: var(--sr-text);">جلسات اخیر</h2>
					<a href="/sessions" class="text-xs font-medium hover:underline" style="color: var(--sr-primary);">مشاهده همه →</a>
				</div>
				<div>
					{#each sessions.slice(0, 5) as session}
						<a href="/sessions/{session.id}" class="px-6 py-3 flex items-center justify-between transition-colors" style="border-bottom: 1px solid var(--sr-border);"
							onmouseenter={(e) => e.currentTarget.style.background = 'var(--sr-bg-alt)'}
							onmouseleave={(e) => e.currentTarget.style.background = 'transparent'}>
							<div class="flex items-center gap-3">
								<div class="relative">
									<div class="w-9 h-9 rounded-lg flex items-center justify-center" style="background: {session.status === 'live' ? 'rgba(64, 191, 127, 0.15)' : 'var(--sr-bg-alt)'};">
										<svg class="w-4 h-4" style="color: {session.status === 'live' ? 'var(--sr-success)' : 'var(--sr-text-muted)'};" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="1.5">
											<path stroke-linecap="round" stroke-linejoin="round" d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" />
										</svg>
									</div>
									{#if session.status === 'live'}
										<span class="absolute -top-0.5 -right-0.5 w-2 h-2 rounded-full animate-pulse" style="background: var(--sr-success);"></span>
									{/if}
								</div>
								<div>
									<p class="text-sm font-medium" style="color: var(--sr-text);">{session.title}</p>
									<p class="text-[10px]" style="color: var(--sr-text-muted);">{toPersianDateTime(session.scheduled_at)}</p>
								</div>
							</div>
							<span class="badge {statusClasses[session.status]}">
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
