<!--
  App Layout — Main authenticated layout with sidebar navigation.
  
  Responsibilities:
    - Redirects unauthenticated users to /auth
    - Renders sidebar with navigation links
    - Handles real-time notifications via WebSocket
    - Shows unread notification count badge
    - Collapsible sidebar (240px → 60px)
    - Mobile responsive sidebar
    - Admin section visible only to admin/teacher roles
  
  Navigation items:
    - Dashboard, Classes, Sessions, Files, Support, Profile
    - Admin: Users, Sessions, Rooms, Tickets, Recordings, Logs, Settings
  
  Route: (app)/* — all authenticated pages
-->
<script lang="ts">
	import { browser } from '$app/environment';
	import { auth, isAdmin } from '$lib/stores';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { onMount, onDestroy } from 'svelte';
	import { api } from '$lib/api';
	import { toPersianNum } from '$lib/utils/persian';
	import { notifications, unreadCount } from '$lib/stores/notifications';
	import type { Notification } from '$lib/stores/notifications';

	let { children } = $props();
	let currentPath = $derived(page.url.pathname);
	let mobileOpen = $state(false);
	let sidebarCollapsed = $state(false);
	let counts = $state({ classes: 0, sessions: 0 });
	let showNotifications = $state(false);
	let showUserMenu = $state(false);

	let ws = $state<WebSocket | null>(null);
	let reconnectTimeout: ReturnType<typeof setTimeout>;
	let mounted = $state(false);

	function connectWebSocket() {
		if (!browser) return;
		const token = localStorage.getItem('access_token');
		if (!token) return;
		const wsBase = api.getWsUrl();
		ws = new WebSocket(`${wsBase}/api/v1/ws?token=${token}`);
		ws.onmessage = (event) => {
			try {
				const msg = JSON.parse(event.data);
				if (msg.type === 'notification' && msg.data) notifications.add(msg.data as Notification);
			} catch {}
		};
		ws.onclose = () => { reconnectTimeout = setTimeout(connectWebSocket, 5000); };
		ws.onerror = () => { ws?.close(); };
	}

	onMount(async () => {
		auth.init();
		// Validate token on mount
		const token = localStorage.getItem('access_token');
		if (token) {
			const res = await api.get('/auth/me');
			if (!res.success) {
				auth.logout();
				goto('/auth');
				return;
			}
		}
		mounted = true;
	});

	$effect(() => {
		if (!mounted) return;
		const token = localStorage.getItem('access_token');
		if (!token) return;
		(async () => {
			try {
				const [c, s] = await Promise.all([api.get<any>('/classes'), api.get<any>('/sessions')]);
				if (c.success && c.data) counts.classes = Array.isArray(c.data) ? c.data.length : (c.data?.total || 0);
				if (s.success && s.data) counts.sessions = Array.isArray(s.data) ? s.data.length : (s.data?.total || 0);
			} catch {}
			await notifications.load();
			connectWebSocket();
		})();
		return () => {
			if (reconnectTimeout) clearTimeout(reconnectTimeout);
			if (ws) ws.close();
			ws = null;
		};
	});

	const icons = {
		dashboard: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="3" width="7" height="7" rx="1"/><rect x="14" y="3" width="7" height="7" rx="1"/><rect x="14" y="14" width="7" height="7" rx="1"/><rect x="3" y="14" width="7" height="7" rx="1"/></svg>`,
		classes: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round"><path d="M2 3h6a4 4 0 0 1 4 4v14a3 3 0 0 0-3-3H2z"/><path d="M22 3h-6a4 4 0 0 0-4 4v14a3 3 0 0 1 3-3h7z"/></svg>`,
		sessions: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round"><path d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14"/><rect x="2" y="6" width="13" height="12" rx="2"/></svg>`,
		files: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round"><path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z"/><polyline points="14 2 14 8 20 8"/></svg>`,
		support: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15a2 2 0 01-2 2H7l-4 4V5a2 2 0 012-2h14a2 2 0 012 2z"/></svg>`,
		profile: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round"><path d="M20 21v-2a4 4 0 00-4-4H8a4 4 0 00-4 4v2"/><circle cx="12" cy="7" r="4"/></svg>`,
		adminHome: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round"><polyline points="22 12 18 12 15 21 9 3 6 12 2 12"/></svg>`,
		users: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round"><path d="M17 21v-2a4 4 0 00-4-4H5a4 4 0 00-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M23 21v-2a4 4 0 00-3-3.87M16 3.13a4 4 0 010 7.75"/></svg>`,
		rooms: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round"><path d="M2 12s3-7 10-7 10 7 10 7-3 7-10 7-10-7-10-7z"/><circle cx="12" cy="12" r="3"/></svg>`,
		recordings: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><circle cx="12" cy="12" r="3"/></svg>`,
		logs: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round"><path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z"/><polyline points="14 2 14 8 20 8"/><line x1="16" y1="13" x2="8" y2="13"/><line x1="16" y1="17" x2="8" y2="17"/></svg>`,
		settings: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 00.33 1.82l.06.06a2 2 0 11-2.83 2.83l-.06-.06a1.65 1.65 0 00-1.82-.33 1.65 1.65 0 00-1 1.51V21a2 2 0 01-4 0v-.09A1.65 1.65 0 009 19.4a1.65 1.65 0 00-1.82.33l-.06.06a2 2 0 11-2.83-2.83l.06-.06a1.65 1.65 0 00.33-1.82 1.65 1.65 0 00-1.51-1H3a2 2 0 010-4h.09A1.65 1.65 0 004.6 9a1.65 1.65 0 00-.33-1.82l-.06-.06a2 2 0 112.83-2.83l.06.06a1.65 1.65 0 001.82.33H9a1.65 1.65 0 001-1.51V3a2 2 0 014 0v.09a1.65 1.65 0 001 1.51 1.65 1.65 0 001.82-.33l.06-.06a2 2 0 112.83 2.83l-.06.06a1.65 1.65 0 00-.33 1.82V9a1.65 1.65 0 001.51 1H21a2 2 0 010 4h-.09a1.65 1.65 0 00-1.51 1z"/></svg>`
	};

	const navItems = $derived.by(() => [
		{ href: '/dashboard', label: 'داشبورد', icon: icons.dashboard },
		{ href: '/classes', label: 'کلاس‌ها', icon: icons.classes },
		{ href: '/sessions', label: 'جلسات', icon: icons.sessions },
		{ href: '/files', label: 'فایل‌ها', icon: icons.files },
		{ href: '/support', label: 'پشتیبانی', icon: icons.support },
		{ href: '/profile', label: 'حساب کاربری', icon: icons.profile },
	]);

	const adminNavItems = $derived.by(() => {
		if (!$isAdmin) return [];
		return [
			{ href: '/admin', label: 'داشبورد مدیریت', icon: icons.adminHome },
			{ href: '/admin/users', label: 'کاربران', icon: icons.users },
			{ href: '/admin/rooms', label: 'اتاق‌ها', icon: icons.rooms },
			{ href: '/admin/recordings', label: 'ضبط‌ها', icon: icons.recordings },
			{ href: '/admin/logs', label: 'لاگ‌ها', icon: icons.logs },
			{ href: '/admin/settings', label: 'تنظیمات', icon: icons.settings },
		];
	});

	function confirmLogout() {
		if (confirm('آیا از خروج از حساب کاربری اطمینان دارید؟')) auth.logout();
	}

	function handleClickOutside(event: MouseEvent) {
		const target = event.target as HTMLElement;
		if (!target.closest('.notification-dropdown') && !target.closest('.notification-bell')) showNotifications = false;
		if (!target.closest('.user-menu')) showUserMenu = false;
	}
</script>

<svelte:document onclick={handleClickOutside} />

<div class="flex min-h-screen" style="background: var(--color-eternal-snow);">
	<!-- Skyroom Sidebar -->
	<aside class="sky-sidebar" class:collapsed={sidebarCollapsed} class:mobile-open={mobileOpen}>
		<!-- Logo -->
		<div class="sky-sidebar-logo">
			<div class="logo-icon">
				<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="white" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14"/><rect x="2" y="6" width="13" height="12" rx="2"/></svg>
			</div>
			{#if !sidebarCollapsed}
				<div class="logo-text">آی‌روم</div>
			{/if}
		</div>

		<!-- Navigation -->
		<nav class="flex-1 px-2 py-3 space-y-0.5 overflow-y-auto">
			{#each navItems as item}
				<a href={item.href}
					class="sky-nav-item {currentPath === item.href || (item.href !== '/admin' && currentPath.startsWith(item.href)) ? 'active' : ''}"
					onclick={() => mobileOpen = false}
					title={sidebarCollapsed ? item.label : undefined}
				>
					{@html item.icon}
					{#if !sidebarCollapsed}
						<span>{item.label}</span>
					{/if}
				</a>
			{/each}

			{#if $isAdmin && adminNavItems.length > 0}
				<div class="pt-2 mt-2" style="border-top: 1px solid rgba(255,255,255,0.08);">
					{#if !sidebarCollapsed}
						<p class="sky-sidebar-section-title">مدیریت</p>
					{/if}
					{#each adminNavItems as item}
						<a href={item.href}
							class="sky-nav-item {currentPath === item.href || (item.href === '/admin' && currentPath === '/admin') ? 'active' : ''}"
							onclick={() => mobileOpen = false}
							title={sidebarCollapsed ? item.label : undefined}
						>
							{@html item.icon}
							{#if !sidebarCollapsed}
								<span>{item.label}</span>
							{/if}
						</a>
					{/each}
				</div>
			{/if}
		</nav>

		<!-- User Info -->
		<div class="px-3 py-3" style="border-top: 1px solid rgba(255,255,255,0.08);">
			{#if $auth.user}
				<div class="flex items-center gap-3 mb-2">
					<div class="w-8 h-8 rounded-full flex items-center justify-center text-white text-xs font-bold shrink-0" style="background: var(--color-crystal-clear);">
						{$auth.user.display_name.charAt(0)}
					</div>
					{#if !sidebarCollapsed}
						<div class="flex-1 min-w-0">
							<p class="text-xs font-bold truncate" style="color: var(--color-pure);">{$auth.user.display_name}</p>
						</div>
					{/if}
				</div>
			{/if}
			<button onclick={confirmLogout}
				class="w-full flex items-center gap-2 px-3 py-2 text-xs rounded-lg transition-all duration-150 font-medium hover:bg-white/5"
				style="color: var(--color-fiery-passion);"
				title={sidebarCollapsed ? 'خروج' : undefined}
			>
				<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round" style="flex-shrink:0;"><path d="M9 21H5a2 2 0 01-2-2V5a2 2 0 012-2h4"/><polyline points="16 17 21 12 16 7"/><line x1="21" y1="12" x2="9" y2="12"/></svg>
				{#if !sidebarCollapsed}
					خروج
				{/if}
			</button>
		</div>
	</aside>

	<!-- Mobile overlay -->
	{#if mobileOpen}
		<div class="fixed inset-0 z-20 lg:hidden" style="background: rgba(0,0,0,0.5); backdrop-filter: blur(4px);" onclick={() => mobileOpen = false} role="button" tabindex="-1"></div>
	{/if}

	<!-- Main content -->
	<div class="flex-1 min-w-0 app-content" style="margin-right: {sidebarCollapsed ? '60px' : '260px'};">
		<!-- Skyroom Header -->
		<header class="sky-header">
			<div class="flex items-center gap-3">
				<button onclick={() => mobileOpen = true} class="sky-btn-icon lg:hidden">
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path stroke-linecap="round" stroke-linejoin="round" d="M3.75 6.75h16.5M3.75 12h16.5m-16.5 5.25h16.5" /></svg>
				</button>
				<button onclick={() => sidebarCollapsed = !sidebarCollapsed} class="sky-btn-icon hidden lg:flex">
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path stroke-linecap="round" stroke-linejoin="round" d="M3.75 6.75h16.5M3.75 12h16.5m-16.5 5.25h16.5" /></svg>
				</button>
			</div>

			<div class="flex items-center gap-3">
				<!-- Notification Bell -->
				<div class="relative">
					<button class="notification-bell sky-btn-icon relative" onclick={() => showNotifications = !showNotifications}>
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path stroke-linecap="round" stroke-linejoin="round" d="M14.857 17.082a23.848 23.848 0 005.454-1.31A8.967 8.967 0 0118 9.75v-.7V9A6 6 0 006 9v.75a8.967 8.967 0 01-2.312 6.022c1.733.64 3.56 1.085 5.455 1.31m5.714 0a24.255 24.255 0 01-5.714 0m5.714 0a3 3 0 11-5.714 0" /></svg>
						{#if $unreadCount > 0}
							<span class="absolute -top-1 -right-1 w-4 h-4 flex items-center justify-center text-[9px] font-bold text-white rounded-full" style="background: var(--color-fiery-passion);">
								{toPersianNum($unreadCount)}
							</span>
						{/if}
					</button>

					{#if showNotifications}
						<div class="notification-dropdown absolute top-full left-0 mt-2 w-80 rounded-xl shadow-xl overflow-hidden" style="background: var(--color-pure); border: 1px solid var(--color-zen-garden);">
							<div class="px-4 py-3 flex items-center justify-between" style="border-bottom: 1px solid var(--color-zen-garden);">
								<h3 class="font-bold text-sm" style="color: var(--color-midnight-sky);">اعلان‌ها</h3>
								{#if $unreadCount > 0}
									<button class="text-xs font-medium hover:underline" style="color: var(--color-crystal-clear);" onclick={() => notifications.markAllRead()}>
										علامت‌گذاری همه به عنوان خوانده شده
									</button>
								{/if}
							</div>
							<div class="max-h-80 overflow-y-auto">
								{#if $notifications.length === 0}
									<div class="px-4 py-8 text-center">
										<p class="text-sm" style="color: var(--color-moonlit-mist);">اعلانی وجود ندارد</p>
									</div>
								{:else}
									{#each $notifications as notification (notification.id)}
										<div class="px-4 py-3 transition-colors cursor-pointer {notification.is_read ? 'opacity-60' : ''}"
											style="border-bottom: 1px solid var(--color-zen-garden);"
											onclick={() => notifications.markRead(notification.id)}>
											<div class="flex items-start gap-3">
												<div class="w-8 h-8 rounded-full flex items-center justify-center shrink-0" style="background: var(--color-polar-ice);">
													<svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="var(--color-crystal-clear)" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round"><path d="M18 8A6 6 0 006 8c0 7-3 9-3 9h18s-3-2-3-9"/><path d="M13.73 21a2 2 0 01-3.46 0"/></svg>
												</div>
												<div class="flex-1 min-w-0">
													<p class="text-sm font-medium" style="color: var(--color-midnight-sky);">{notification.title}</p>
													{#if notification.message}
														<p class="text-xs mt-0.5 truncate" style="color: var(--color-moonlit-mist);">{notification.message}</p>
													{/if}
												</div>
												{#if !notification.is_read}
													<div class="w-2 h-2 rounded-full shrink-0 mt-1.5" style="background: var(--color-crystal-clear);"></div>
												{/if}
											</div>
										</div>
									{/each}
								{/if}
							</div>
						</div>
					{/if}
				</div>

				<!-- User Menu -->
				<div class="relative user-menu">
					<button onclick={() => showUserMenu = !showUserMenu} class="flex items-center gap-2 px-3 py-1.5 rounded-lg hover:bg-gray-100 transition-colors">
						{#if $auth.user}
							<div class="w-8 h-8 rounded-full flex items-center justify-center text-white text-xs font-bold" style="background: var(--color-crystal-clear);">
								{$auth.user.display_name.charAt(0)}
							</div>
							<span class="text-sm font-medium hidden sm:block" style="color: var(--color-midnight-sky);">{$auth.user.display_name}</span>
						{/if}
						<svg class="w-4 h-4" style="color: var(--color-moonlit-mist);" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M19 9l-7 7-7-7" /></svg>
					</button>

					{#if showUserMenu}
						<div class="absolute top-full left-0 mt-2 w-48 rounded-xl shadow-xl overflow-hidden" style="background: var(--color-pure); border: 1px solid var(--color-zen-garden);">
							<a href="/profile" class="block px-4 py-2.5 text-[13px] hover:bg-gray-50 transition-colors" style="color: var(--color-midnight-sky);">حساب کاربری</a>
							<a href="/dashboard" class="block px-4 py-2.5 text-[13px] hover:bg-gray-50 transition-colors" style="color: var(--color-midnight-sky);">داشبورد</a>
							<div class="h-px" style="background: var(--color-zen-garden);"></div>
							<a href="/auth" class="block px-4 py-2.5 text-[13px] hover:bg-red-50 transition-colors" style="color: var(--color-fiery-passion);">خروج</a>
						</div>
					{/if}
				</div>
			</div>
		</header>

		<!-- Page Content -->
		<main class="p-4 lg:p-6" style="max-width: 1226px; margin: 0 auto;">
			{@render children()}
		</main>
	</div>
</div>
