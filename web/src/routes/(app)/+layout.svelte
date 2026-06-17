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

	onMount(() => {
		auth.init();
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

	const navItems = $derived.by(() => [
		{ href: '/dashboard', label: 'داشبورد', icon: '📊' },
		{ href: '/classes', label: 'کلاس‌ها', icon: '📚' },
		{ href: '/sessions', label: 'جلسات', icon: '🎥' },
		{ href: '/files', label: 'فایل‌ها', icon: '📁' },
		{ href: '/support', label: 'پشتیبانی', icon: '🎧' },
		{ href: '/profile', label: 'حساب کاربری', icon: '👤' },
	]);

	const adminNavItems = $derived.by(() => {
		if (!$isAdmin) return [];
		return [
			{ href: '/admin', label: 'داشبورد مدیریت', icon: '📈' },
			{ href: '/admin/users', label: 'کاربران', icon: '👥' },
			{ href: '/admin/rooms', label: 'اتاق‌ها', icon: '🏢' },
			{ href: '/admin/recordings', label: 'ضبط‌ها', icon: '⏺' },
			{ href: '/admin/logs', label: 'لاگ‌ها', icon: '📋' },
			{ href: '/admin/settings', label: 'تنظیمات', icon: '⚙️' },
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
	<aside class="sky-sidebar" class:collapsed={sidebarCollapsed}>
		<!-- Logo -->
		<div class="sky-sidebar-logo">
			<div class="logo-icon">آ</div>
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
					<span>{item.icon}</span>
					{#if !sidebarCollapsed}
						<span>{item.label}</span>
					{/if}
				</a>
			{/each}

			{#if $isAdmin && adminNavItems.length > 0}
				<div class="pt-2 mt-2" style="border-top: 1px solid rgba(255,255,255,0.08);">
					{#if !sidebarCollapsed}
						<p class="px-4 mb-1 text-[10px] font-bold uppercase tracking-wider" style="color: var(--color-mystic-sea);">مدیریت</p>
					{/if}
					{#each adminNavItems as item}
						<a href={item.href}
							class="sky-nav-item {currentPath === item.href || (item.href === '/admin' && currentPath === '/admin') ? 'active' : ''}"
							onclick={() => mobileOpen = false}
							title={sidebarCollapsed ? item.label : undefined}
						>
							<span>{item.icon}</span>
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
			>
				<span>🚪</span>
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
	<div class="flex-1 min-w-0" style="margin-right: {sidebarCollapsed ? '60px' : '260px'};">
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
												<div class="w-8 h-8 rounded-full flex items-center justify-center shrink-0" style="background: var(--color-secret-glow);">
													<span class="text-sm">🔔</span>
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
		<main class="p-4 lg:p-6" style="max-width: 1200px; margin: 0 auto;">
			{@render children()}
		</main>
	</div>
</div>
