<script lang="ts">
	import { auth, isAdmin } from '$lib/stores';
	import { api } from '$lib/api';
	import { onMount } from 'svelte';
	import type { Class, Session } from '$lib/types';
	import { toPersianNum, toPersianDate, toPersianDateTime } from '$lib/utils/persian';

	let stats = $state({ users: 0, classes: 0, sessions: 0, messages: 0 });
	let classes = $state<Class[]>([]);
	let sessions = $state<Session[]>([]);
	let currentUser = $state<any>(null);

	// Subscribe to auth store
	auth.subscribe(s => { currentUser = s.user; });
	let loading = $state(true);

	onMount(() => loadData());

	async function loadData() {
		loading = true;
		if ($isAdmin) {
			const statsRes = await api.get<{ users: number; classes: number; sessions: number; messages: number }>('/admin/dashboard/stats');
			if (statsRes.success) stats = statsRes.data!;
		}
		const classesRes = await api.get<Class[]>('/classes');
		if (classesRes.success && classesRes.data) classes = Array.isArray(classesRes.data) ? classesRes.data : ((classesRes.data as any).items || []);
		const sessionsRes = await api.get<Session[]>('/sessions');
		if (sessionsRes.success && sessionsRes.data) sessions = Array.isArray(sessionsRes.data) ? sessionsRes.data : ((sessionsRes.data as any).items || []);
		loading = false;
	}

	const statusLabels: Record<string, string> = { scheduled: 'برنامه‌ریزی شده', live: 'در حال برگزاری', ended: 'پایان یافته' };
	const statusBadge: Record<string, string> = { scheduled: 'sky-badge sky-badge-info', live: 'sky-badge sky-badge-success', ended: 'sky-badge sky-badge-default' };

	const statCards = $derived([
		{ label: 'کاربران', value: stats.users, grad: 'linear-gradient(135deg,#23b9d7,#004ff2)', icon: `<svg width="24" height="24" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M15 19.128a9.38 9.38 0 002.625.372 9.337 9.337 0 004.121-.952 4.125 4.125 0 00-7.533-2.493M15 19.128v-.003c0-1.113-.285-2.16-.786-3.07M15 19.128v.106A12.318 12.318 0 018.624 21c-2.331 0-4.512-.645-6.374-1.766l-.001-.109a6.375 6.375 0 0111.964-3.07M12 6.375a3.375 3.375 0 11-6.75 0 3.375 3.375 0 016.75 0zm8.25 2.25a2.625 2.625 0 11-5.25 0 2.625 2.625 0 015.25 0z"/></svg>` },
		{ label: 'کلاس‌ها', value: stats.classes, grad: 'linear-gradient(135deg,#8b5cf6,#7c3aed)', icon: `<svg width="24" height="24" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253"/></svg>` },
		{ label: 'جلسات', value: stats.sessions, grad: 'linear-gradient(135deg,#40bf7f,#059669)', icon: `<svg width="24" height="24" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z"/></svg>` },
		{ label: 'پیام‌ها', value: stats.messages, grad: 'linear-gradient(135deg,#f59e0b,#d97706)', icon: `<svg width="24" height="24" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z"/></svg>` }
	]);
</script>

<div class="space-y-5">
	<!-- Greeting -->
	<div>
		<h1 class="sky-page-title">سلام {currentUser?.display_name} 👋</h1>
		<p class="sky-page-subtitle">
			{toPersianDate(new Date())} — {new Date().toLocaleDateString('fa-IR', { weekday: 'long' })}
		</p>
	</div>

	{#if loading}
		<div class="flex items-center justify-center py-20">
			<svg class="sky-spinner lg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="color: var(--color-crystal-clear);"><path d="M12 2v4M12 18v4M4.93 4.93l2.83 2.83M16.24 16.24l2.83 2.83M2 12h4M18 12h4M4.93 19.07l2.83-2.83M16.24 7.76l2.83-2.83"/></svg>
		</div>
	{:else}
		{#if $isAdmin}
			<div class="grid grid-cols-2 lg:grid-cols-4 gap-4">
				{#each statCards as stat}
					<div class="sky-stat-card">
						<div class="sky-stat-icon" style="background: {stat.grad};">{@html stat.icon}</div>
						<div>
							<p class="sky-stat-label">{stat.label}</p>
							<p class="sky-stat-value">{toPersianNum(stat.value)}</p>
						</div>
					</div>
				{/each}
			</div>
		{/if}

		<!-- My Classes -->
		<div class="sky-card">
			<div class="sky-card-header">
				<h2>کلاس‌های من</h2>
				<a href="/classes" class="text-xs font-medium hover:underline" style="color: var(--color-crystal-clear);">مشاهده همه ←</a>
			</div>
			<div class="sky-card-body">
				{#if classes.length === 0}
					<div class="sky-empty">
						<div class="sky-empty-icon">
							<svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1" style="color: var(--color-muted-mountain);"><path d="M2 3h6a4 4 0 014 4v14a3 3 0 00-3-3H2z"/><path d="M22 3h-6a4 4 0 00-4 4v14a3 3 0 013-3h7z"/></svg>
						</div>
						<p class="sky-empty-title">هنوز کلاسی ایجاد نشده</p>
						<a href="/classes" class="sky-btn sky-btn-primary mt-2">ایجاد کلاس</a>
					</div>
				{:else}
					<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
						{#each classes.slice(0, 6) as cls}
							<a href="/{(currentUser?.email || 'admin').split('@')[0]}/{cls.slug || cls.name.toLowerCase().replace(/\s+/g, '-')}" class="block p-4 rounded-xl transition-all hover:shadow-sm group" style="background: var(--color-secret-glow); text-decoration: none;">
								<div class="flex items-center gap-3 mb-2">
									<div class="w-9 h-9 rounded-lg flex items-center justify-center text-white font-bold text-sm shrink-0" style="background: {cls.color || 'var(--color-crystal-clear)'};">{cls.name.charAt(0)}</div>
									<h3 class="font-bold text-sm group-hover:underline" style="color: var(--color-midnight-sky);">{cls.name}</h3>
								</div>
								{#if cls.description}
									<p class="text-xs line-clamp-2 mb-2" style="color: var(--color-mystic-sea);">{cls.description}</p>
								{/if}
								<div class="flex items-center gap-1 text-[11px] pt-2" style="color: var(--color-moonlit-mist); border-top: 1px solid var(--color-zen-garden);">
									<svg width="13" height="13" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z"/></svg>
									حداکثر {toPersianNum(cls.max_students)} نفر
								</div>
							</a>
						{/each}
					</div>
				{/if}
			</div>
		</div>

		<!-- Recent Sessions -->
		{#if sessions.length > 0}
			<div class="sky-card">
				<div class="sky-card-header">
					<h2>جلسات اخیر</h2>
					<a href="/sessions" class="text-xs font-medium hover:underline" style="color: var(--color-crystal-clear);">مشاهده همه ←</a>
				</div>
				<div>
					{#each sessions.slice(0, 5) as session}
						<a href="/sessions/{session.id}/logs" class="sky-session-row" style="text-decoration: none;">
							<div class="flex items-center gap-3">
								<div class="relative">
									<div class="w-9 h-9 rounded-lg flex items-center justify-center" style="background: {session.status === 'live' ? 'rgba(64,191,127,0.12)' : 'var(--color-secret-glow)'};">
										<svg width="16" height="16" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24" style="color: {session.status === 'live' ? 'var(--color-lush-meadow)' : 'var(--color-moonlit-mist)'};"><path stroke-linecap="round" stroke-linejoin="round" d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z"/></svg>
									</div>
									{#if session.status === 'live'}
										<span class="absolute -top-0.5 -right-0.5 w-2 h-2 rounded-full animate-pulse" style="background: var(--color-lush-meadow);"></span>
									{/if}
								</div>
								<div>
									<p class="text-sm font-semibold" style="color: var(--color-midnight-sky);">{session.title}</p>
									<p class="text-[11px]" style="color: var(--color-moonlit-mist);">{toPersianDateTime(session.scheduled_at)}</p>
								</div>
							</div>
							<span class="{statusBadge[session.status]}">
								{#if session.status === 'live'}<span class="dot"></span>{/if}
								{statusLabels[session.status]}
							</span>
						</a>
					{/each}
				</div>
			</div>
		{/if}
	{/if}
</div>
