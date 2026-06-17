<script lang="ts">
	import '../app.css';
	import { auth, isAdmin } from '$lib/stores';
	import { page } from '$app/state';
	import { onMount } from 'svelte';
	import { api } from '$lib/api';
	import { toPersianNum } from '$lib/utils/persian';

	let { children } = $props();
	let currentPath = $derived(page.url.pathname);
	let mobileOpen = $state(false);
	let sidebarCollapsed = $state(false);
	let counts = $state({ classes: 0, sessions: 0 });
	let showNotifications = $state(false);
	let unreadCount = $state(3); // Placeholder - would come from API

	onMount(async () => {
		try {
			const [c, s] = await Promise.all([
				api.get<any>('/classes'),
				api.get<any>('/sessions')
			]);
			if (c.success && c.data) counts.classes = Array.isArray(c.data) ? c.data.length : (c.data?.total || 0);
			if (s.success && s.data) counts.sessions = Array.isArray(s.data) ? s.data.length : (s.data?.total || 0);
		} catch {}
	});

	const isOnAdmin = $derived(currentPath.startsWith('/admin'));

	const navItems = $derived.by(() => {
		const items = [
			{ href: '/dashboard', label: 'داشبورد', icon: 'M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-4 0h4' },
			{ href: '/classes', label: 'کلاس‌ها', icon: 'M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253' },
			{ href: '/sessions', label: 'جلسات', icon: 'M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z' },
			{ href: '/files', label: 'فایل‌ها', icon: 'M19.5 14.25v-2.625a3.375 3.375 0 00-3.375-3.375h-1.5A1.125 1.125 0 0113.5 7.125v-1.5a3.375 3.375 0 00-3.375-3.375H8.25m6.75 12H9.75m0-3h6m-6 6h6M3.375 6.75h17.25a.375.375 0 01.375.375v11.25a.375.375 0 01-.375-.375H3.375a.375.375 0 01-.375-.375V7.125a.375.375 0 01.375-.375z' },
			{ href: '/support', label: 'پشتیبانی', icon: 'M20.25 8.511c.884.284 1.5 1.128 1.5 2.097v4.286c0 1.136-.847 2.1-1.98 2.193-.34.027-.68.052-1.02.072v3.091l-3-3c-1.354 0-2.694-.055-4.02-.163a2.115 2.115 0 01-.825-.242m9.345-8.334a2.126 2.126 0 00-.476-.095 48.64 48.64 0 00-8.048 0c-1.131.094-1.976.94-1.976 2.097v4.286c0 .837.46 1.58 1.155 1.951m9.345-8.334V6.637c0-1.621-1.152-3.026-2.76-3.235A48.455 48.455 0 0011.25 3c-2.115 0-4.198.137-6.24.402-1.608.209-2.76 1.614-2.76 3.235v6.226c0 1.621 1.152 3.026 2.76 3.235.577.075 1.157.14 1.74.194V21l4.155-4.155' },
			{ href: '/profile', label: 'حساب کاربری', icon: 'M15.75 6a3.75 3.75 0 11-7.5 0 3.75 3.75 0 017.5 0zM4.501 20.118a7.5 7.5 0 0114.998 0A17.933 17.933 0 0112 21.75c-2.676 0-5.216-.584-7.499-1.632z' }
		];

		if ($isAdmin) {
			items.push(
				{ href: '/admin', label: 'داشبورد مدیریت', icon: 'M3.75 6A2.25 2.25 0 016 3.75h2.25A2.25 2.25 0 0110.5 6v2.25a2.25 2.25 0 01-2.25 2.25H6a2.25 2.25 0 01-2.25-2.25V6zM3.75 15.75A2.25 2.25 0 016 13.5h2.25a2.25 2.25 0 012.25 2.25V18a2.25 2.25 0 01-2.25 2.25H6A2.25 2.25 0 013.75 18v-2.25zM13.5 6a2.25 2.25 0 012.25-2.25H18A2.25 2.25 0 0120.25 6v2.25A2.25 2.25 0 0118 10.5h-2.25a2.25 2.25 0 01-2.25-2.25V6zM13.5 15.75a2.25 2.25 0 012.25-2.25H18a2.25 2.25 0 012.25 2.25V18A2.25 2.25 0 0118 20.25h-2.25A2.25 2.25 0 0113.5 18v-2.25z' },
				{ href: '/admin/rooms', label: 'اتاق‌ها', icon: 'M2.25 21h19.5m-18-18v18m10.5-18v18m6-13.5V21M6.75 6.75h.75m-.75 3h.75m-.75 3h.75m3-6h.75m-.75 3h.75m-.75 3h.75M6.75 21v-3.375c0-.621.504-1.125 1.125-1.125h2.25c.621 0 1.125.504 1.125 1.125V21M3 3h12m-.75 4.5H21m-3.75 3H21m-3.75 3H21' },
				{ href: '/admin/recordings', label: 'ضبط‌ها', icon: 'M15.75 10.5l4.72-4.72a.75.75 0 011.28.53v11.38a.75.75 0 01-1.28.53l-4.72-4.72M4.5 18.75h9a2.25 2.25 0 002.25-2.25v-9a2.25 2.25 0 00-2.25-2.25h-9A2.25 2.25 0 002.25 7.5v9a2.25 2.25 0 002.25 2.25z' },
				{ href: '/admin/logs', label: 'لاگ‌ها', icon: 'M19.5 14.25v-2.625a3.375 3.375 0 00-3.375-3.375h-1.5A1.125 1.125 0 0113.5 7.125v-1.5a3.375 3.375 0 00-3.375-3.375H8.25m0 12.75h7.5m-7.5 3H12M10.5 2.25H5.625c-.621 0-1.125.504-1.125 1.125v17.25c0 .621.504 1.125 1.125 1.125h12.75c.621 0 1.125-.504 1.125-1.125V11.25a9 9 0 00-9-9z' },
				{ href: '/admin/settings', label: 'تنظیمات', icon: 'M9.594 3.94c.09-.542.56-.94 1.11-.94h2.593c.55 0 1.02.398 1.11.94l.213 1.281c.063.374.313.686.645.87.074.04.147.083.22.127.325.196.72.257 1.075.124l1.217-.456a1.125 1.125 0 011.37.49l1.296 2.247a1.125 1.125 0 01-.26 1.431l-1.003.827c-.293.241-.438.613-.43.992a7.723 7.723 0 010 .255c-.008.378.137.75.43.991l1.004.827c.424.35.534.955.26 1.43l-1.298 2.247a1.125 1.125 0 01-1.369.491l-1.217-.456c-.355-.133-.75-.072-1.076.124a6.47 6.47 0 01-.22.128c-.331.183-.581.495-.644.869l-.213 1.281c-.09.543-.56.94-1.11.94h-2.594c-.55 0-1.019-.398-1.11-.94l-.213-1.281c-.062-.374-.312-.686-.644-.87a6.52 6.52 0 01-.22-.127c-.325-.196-.72-.257-1.076-.124l-1.217.456a1.125 1.125 0 01-1.369-.49l-1.297-2.247a1.125 1.125 0 01.26-1.431l1.004-.827c.292-.24.437-.613.43-.991a6.932 6.932 0 010-.255c.007-.38-.138-.751-.43-.992l-1.004-.827a1.125 1.125 0 01-.26-1.43l1.297-2.247a1.125 1.125 0 011.37-.491l1.216.456c.356.133.751.072 1.076-.124.072-.044.146-.086.22-.128.332-.183.582-.495.644-.869l.214-1.28z M15 12a3 3 0 11-6 0 3 3 0 016 0z' }
			);
		}

		return items;
	});

	const adminNavItems = $derived.by(() => {
		if (!$isAdmin) return [];
		return [
			{ href: '/admin', label: 'داشبورد', icon: 'M3.75 6A2.25 2.25 0 016 3.75h2.25A2.25 2.25 0 0110.5 6v2.25a2.25 2.25 0 01-2.25 2.25H6a2.25 2.25 0 01-2.25-2.25V6zM3.75 15.75A2.25 2.25 0 016 13.5h2.25a2.25 2.25 0 012.25 2.25V18a2.25 2.25 0 01-2.25 2.25H6A2.25 2.25 0 013.75 18v-2.25zM13.5 6a2.25 2.25 0 012.25-2.25H18A2.25 2.25 0 0120.25 6v2.25A2.25 2.25 0 0118 10.5h-2.25a2.25 2.25 0 01-2.25-2.25V6zM13.5 15.75a2.25 2.25 0 012.25-2.25H18a2.25 2.25 0 012.25 2.25V18A2.25 2.25 0 0118 20.25h-2.25A2.25 2.25 0 0113.5 18v-2.25z' },
			{ href: '/admin/users', label: 'کاربران', icon: 'M15 19.128a9.38 9.38 0 002.625.372 9.337 9.337 0 004.121-.952 4.125 4.125 0 00-7.533-2.493M15 19.128v-.003c0-1.113-.285-2.16-.786-3.07M15 19.128v.106A12.318 12.318 0 018.624 21c-2.331 0-4.512-.645-6.374-1.766l-.001-.109a6.375 6.375 0 0111.964-3.07M12 6.375a3.375 3.375 0 11-6.75 0 3.375 3.375 0 016.75 0zm8.25 2.25a2.625 2.625 0 11-5.25 0 2.625 2.625 0 015.25 0z' },
			{ href: '/admin/sessions', label: 'جلسات', icon: 'M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z' },
			{ href: '/admin/rooms', label: 'اتاق‌ها', icon: 'M2.25 21h19.5m-18-18v18m10.5-18v18m6-13.5V21M6.75 6.75h.75m-.75 3h.75m-.75 3h.75m3-6h.75m-.75 3h.75m-.75 3h.75M6.75 21v-3.375c0-.621.504-1.125 1.125-1.125h2.25c.621 0 1.125.504 1.125 1.125V21M3 3h12m-.75 4.5H21m-3.75 3H21m-3.75 3H21' },
			{ href: '/support', label: 'تیکت‌ها', icon: 'M20.25 8.511c.884.284 1.5 1.128 1.5 2.097v4.286c0 1.136-.847 2.1-1.98 2.193-.34.027-.68.052-1.02.072v3.091l-3-3c-1.354 0-2.694-.055-4.02-.163a2.115 2.115 0 01-.825-.242m9.345-8.334a2.126 2.126 0 00-.476-.095 48.64 48.64 0 00-8.048 0c-1.131.094-1.976.94-1.976 2.097v4.286c0 .837.46 1.58 1.155 1.951m9.345-8.334V6.637c0-1.621-1.152-3.026-2.76-3.235A48.455 48.455 0 0011.25 3c-2.115 0-4.198.137-6.24.402-1.608.209-2.76 1.614-2.76 3.235v6.226c0 1.621 1.152 3.026 2.76 3.235.577.075 1.157.14 1.74.194V21l4.155-4.155' },
			{ href: '/admin/recordings', label: 'ضبط‌ها', icon: 'M15.75 10.5l4.72-4.72a.75.75 0 011.28.53v11.38a.75.75 0 01-1.28.53l-4.72-4.72M4.5 18.75h9a2.25 2.25 0 002.25-2.25v-9a2.25 2.25 0 00-2.25-2.25h-9A2.25 2.25 0 002.25 7.5v9a2.25 2.25 0 002.25 2.25z' },
			{ href: '/admin/logs', label: 'لاگ‌ها', icon: 'M19.5 14.25v-2.625a3.375 3.375 0 00-3.375-3.375h-1.5A1.125 1.125 0 0113.5 7.125v-1.5a3.375 3.375 0 00-3.375-3.375H8.25m0 12.75h7.5m-7.5 3H12M10.5 2.25H5.625c-.621 0-1.125.504-1.125 1.125v17.25c0 .621.504 1.125 1.125 1.125h12.75c.621 0 1.125-.504 1.125-1.125V11.25a9 9 0 00-9-9z' },
			{ href: '/admin/settings', label: 'تنظیمات', icon: 'M9.594 3.94c.09-.542.56-.94 1.11-.94h2.593c.55 0 1.02.398 1.11.94l.213 1.281c.063.374.313.686.645.87.074.04.147.083.22.127.325.196.72.257 1.075.124l1.217-.456a1.125 1.125 0 011.37.49l1.296 2.247a1.125 1.125 0 01-.26 1.431l-1.003.827c-.293.241-.438.613-.43.992a7.723 7.723 0 010 .255c-.008.378.137.75.43.991l1.004.827c.424.35.534.955.26 1.43l-1.298 2.247a1.125 1.125 0 01-1.369.491l-1.217-.456c-.355-.133-.75-.072-1.076.124a6.47 6.47 0 01-.22.128c-.331.183-.581.495-.644.869l-.213 1.281c-.09.543-.56.94-1.11.94h-2.594c-.55 0-1.019-.398-1.11-.94l-.213-1.281c-.062-.374-.312-.686-.644-.87a6.52 6.52 0 01-.22-.127c-.325-.196-.72-.257-1.076-.124l-1.217.456a1.125 1.125 0 01-1.369-.49l-1.297-2.247a1.125 1.125 0 01.26-1.431l1.004-.827c.292-.24.437-.613.43-.991a6.932 6.932 0 010-.255c.007-.38-.138-.751-.43-.992l-1.004-.827a1.125 1.125 0 01-.26-1.43l1.297-2.247a1.125 1.125 0 011.37-.491l1.216.456c.356.133.751.072 1.076-.124.072-.044.146-.086.22-.128.332-.183.582-.495.644-.869l.214-1.28z M15 12a3 3 0 11-6 0 3 3 0 016 0z' }
		];
	});

	const roleLabels: Record<string, string> = { admin: 'مدیر سیستم', teacher: 'مدرس', student: 'دانش‌آموز' };
	const roleBadgeColors: Record<string, string> = {
		admin: 'bg-amber-500/20 text-amber-400',
		teacher: 'bg-violet-500/20 text-violet-400',
		student: 'bg-teal-500/20 text-teal-400'
	};

	// Close notifications when clicking outside
	function handleClickOutside(event: MouseEvent) {
		const target = event.target as HTMLElement;
		if (!target.closest('.notification-dropdown') && !target.closest('.notification-bell')) {
			showNotifications = false;
		}
	}
</script>

<svelte:document onclick={handleClickOutside} />

<div class="flex min-h-screen bg-[#0d0d1a]">
	<!-- Sidebar -->
	<aside class="fixed inset-y-0 right-0 z-30 transform transition-all duration-300 ease-in-out
		{mobileOpen ? 'translate-x-0' : 'translate-x-full'}
		lg:translate-x-0 lg:static
		{sidebarCollapsed ? 'w-[60px]' : 'w-64'}"
		style="background: var(--sky-bg-dark); border-left: 1px solid var(--sky-border);">
		<div class="flex flex-col h-full">
			<!-- Logo -->
			<div class="px-4 py-5" style="border-bottom: 1px solid var(--sky-border);">
				<div class="flex items-center gap-3">
					<div class="w-10 h-10 rounded-xl flex items-center justify-center text-white font-bold text-lg shrink-0"
						style="background: linear-gradient(135deg, #1a56db, #4361ee); box-shadow: 0 4px 12px rgba(67, 97, 238, 0.3);">
						آ
					</div>
					{#if !sidebarCollapsed}
						<div class="min-w-0">
							<h1 class="font-extrabold text-lg" style="color: var(--sky-text-primary);">آی‌روم</h1>
							<p class="text-[10px] font-medium" style="color: var(--sky-text-secondary);">کلاس آنلاین هوشمند</p>
						</div>
					{/if}
				</div>
			</div>

			<!-- Collapse toggle (desktop) -->
			<div class="hidden lg:flex justify-center px-3 pt-3">
				<button onclick={() => sidebarCollapsed = !sidebarCollapsed} 
					class="p-2 rounded-lg transition-all duration-200 hover:bg-[#0f3460]"
					style="color: var(--sky-text-secondary);"
					title={sidebarCollapsed ? 'باز کردن منو' : 'بستن منو'}>
					<svg class="w-4 h-4 transition-transform duration-300 {sidebarCollapsed ? 'rotate-180' : ''}" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 19l-7-7 7-7m8 14l-7-7 7-7" />
					</svg>
				</button>
			</div>

			<!-- Navigation -->
			<nav class="flex-1 px-3 py-4 space-y-1 overflow-y-auto">
				{#each navItems as item}
					<a
						href={item.href}
						class="sky-sidebar-link {currentPath === item.href || (item.href !== '/admin' && currentPath.startsWith(item.href)) ? 'active' : ''}"
						onclick={() => mobileOpen = false}
						title={sidebarCollapsed ? item.label : undefined}
					>
						<svg class="w-5 h-5 shrink-0" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" d={item.icon} />
						</svg>
						{#if !sidebarCollapsed}
							<span>{item.label}</span>
						{/if}
						{#if !sidebarCollapsed && item.href === '/classes' && counts.classes > 0}
							<span class="text-[10px] px-1.5 py-0.5 rounded-full ms-auto" style="background: var(--sky-bg-input); color: var(--sky-text-secondary);">{counts.classes}</span>
						{:else if !sidebarCollapsed && item.href === '/sessions' && counts.sessions > 0}
							<span class="text-[10px] px-1.5 py-0.5 rounded-full ms-auto" style="background: var(--sky-bg-input); color: var(--sky-text-secondary);">{counts.sessions}</span>
						{/if}
					</a>
				{/each}

				{#if $isAdmin && adminNavItems.length > 0}
					<div class="pt-3 mt-3" style="border-top: 1px solid var(--sky-border);">
						{#if !sidebarCollapsed}
							<p class="px-3 mb-2 text-[10px] font-bold uppercase tracking-wider" style="color: var(--sky-text-secondary);">مدیریت</p>
						{/if}
						{#each adminNavItems as item}
							<a
								href={item.href}
								class="sky-sidebar-link {currentPath === item.href || (item.href === '/admin' && currentPath === '/admin') ? 'active' : ''}"
								onclick={() => mobileOpen = false}
								title={sidebarCollapsed ? item.label : undefined}
							>
								<svg class="w-5 h-5 shrink-0" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" d={item.icon} />
								</svg>
								{#if !sidebarCollapsed}
									<span>{item.label}</span>
								{/if}
							</a>
						{/each}
					</div>
				{/if}
			</nav>

			<!-- User info -->
			<div class="px-3 py-4" style="border-top: 1px solid var(--sky-border);">
				{#if $auth.user}
					<div class="flex items-center gap-3 mb-3">
						<div class="w-10 h-10 rounded-xl flex items-center justify-center text-white font-bold text-sm shrink-0"
							style="background: linear-gradient(135deg, #1a56db, #7c3aed); box-shadow: 0 4px 12px rgba(114, 9, 183, 0.3);">
							{$auth.user.display_name.charAt(0)}
						</div>
						{#if !sidebarCollapsed}
							<div class="flex-1 min-w-0">
								<p class="text-sm font-bold truncate" style="color: var(--sky-text-primary);">{$auth.user.display_name}</p>
								<span class="text-[10px] px-2 py-0.5 rounded-full font-semibold {roleBadgeColors[$auth.user.role]}">
									{roleLabels[$auth.user.role]}
								</span>
							</div>
						{/if}
					</div>
				{/if}
			<button
				onclick={() => auth.logout()}
				class="w-full flex items-center justify-center gap-2 px-3 py-2.5 text-sm rounded-xl transition-all duration-200 font-medium hover:bg-red-500/10 logout-btn"
				style="color: var(--sky-text-secondary);"
			>
					<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="1.5">
						<path stroke-linecap="round" stroke-linejoin="round" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
					</svg>
					{#if !sidebarCollapsed}
						خروج از حساب
					{/if}
				</button>
			</div>
		</div>
	</aside>

	<!-- Mobile overlay -->
	{#if mobileOpen}
		<div class="fixed inset-0 bg-black/50 backdrop-blur-sm z-20 lg:hidden" onclick={() => mobileOpen = false} role="button" tabindex="-1"></div>
	{/if}

	<!-- Main content -->
	<div class="flex-1 min-w-0">
		<!-- Desktop header with notification bell -->
		<header class="sticky top-0 z-10 px-6 py-3 flex items-center justify-between hidden lg:flex"
			style="background: var(--sky-bg-panel); border-bottom: 1px solid var(--sky-border);">
			<div class="flex items-center gap-4">
				<button onclick={() => mobileOpen = true} class="p-2 rounded-lg transition-colors hover:bg-[#0f3460]" style="color: var(--sky-text-secondary);" aria-label="منو">
					<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="1.5">
						<path stroke-linecap="round" stroke-linejoin="round" d="M3.75 6.75h16.5M3.75 12h16.5m-16.5 5.25h16.5" />
					</svg>
				</button>
			</div>
			
			<div class="flex items-center gap-4">
				<!-- Notification Bell -->
				<div class="relative">
					<button class="notification-bell p-2 rounded-lg transition-colors hover:bg-[#0f3460] relative"
						style="color: var(--sky-text-secondary);"
						onclick={() => showNotifications = !showNotifications}>
						<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="1.5">
							<path stroke-linecap="round" stroke-linejoin="round" d="M14.857 17.082a23.848 23.848 0 005.454-1.31A8.967 8.967 0 0118 9.75v-.7V9A6 6 0 006 9v.75a8.967 8.967 0 01-2.312 6.022c1.733.64 3.56 1.085 5.455 1.31m5.714 0a24.255 24.255 0 01-5.714 0m5.714 0a3 3 0 11-5.714 0" />
						</svg>
						{#if unreadCount > 0}
							<span class="absolute -top-1 -right-1 w-5 h-5 flex items-center justify-center text-[10px] font-bold text-white rounded-full"
								style="background: var(--sky-accent-red);">
								{toPersianNum(unreadCount)}
							</span>
						{/if}
					</button>
					
					<!-- Notification Dropdown -->
					{#if showNotifications}
						<div class="notification-dropdown absolute left-0 mt-2 w-80 rounded-xl shadow-2xl overflow-hidden"
							style="background: var(--sky-bg-panel); border: 1px solid var(--sky-border);">
							<div class="px-4 py-3" style="border-bottom: 1px solid var(--sky-border);">
								<h3 class="font-bold" style="color: var(--sky-text-primary);">اعلان‌ها</h3>
							</div>
							<div class="max-h-80 overflow-y-auto">
								<!-- Placeholder notifications -->
								<div class="px-4 py-3 hover:bg-[#0f3460] transition-colors cursor-pointer" style="border-bottom: 1px solid var(--sky-border);">
									<div class="flex items-start gap-3">
										<div class="w-8 h-8 rounded-full flex items-center justify-center shrink-0" style="background: rgba(67, 97, 238, 0.2);">
											<svg class="w-4 h-4" style="color: var(--sky-accent-blue);" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
											</svg>
										</div>
										<div class="flex-1 min-w-0">
											<p class="text-sm font-medium" style="color: var(--sky-text-primary);">جلسه جدید ایجاد شد</p>
											<p class="text-xs mt-1" style="color: var(--sky-text-secondary);">۵ دقیقه پیش</p>
										</div>
									</div>
								</div>
								<div class="px-4 py-3 hover:bg-[#0f3460] transition-colors cursor-pointer" style="border-bottom: 1px solid var(--sky-border);">
									<div class="flex items-start gap-3">
										<div class="w-8 h-8 rounded-full flex items-center justify-center shrink-0" style="background: rgba(0, 210, 106, 0.2);">
											<svg class="w-4 h-4" style="color: var(--sky-accent-green);" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
											</svg>
										</div>
										<div class="flex-1 min-w-0">
											<p class="text-sm font-medium" style="color: var(--sky-text-primary);">تیکت شما پاسخ داده شد</p>
											<p class="text-xs mt-1" style="color: var(--sky-text-secondary);">۱ ساعت پیش</p>
										</div>
									</div>
								</div>
								<div class="px-4 py-3 hover:bg-[#0f3460] transition-colors cursor-pointer">
									<div class="flex items-start gap-3">
										<div class="w-8 h-8 rounded-full flex items-center justify-center shrink-0" style="background: rgba(233, 69, 96, 0.2);">
											<svg class="w-4 h-4" style="color: var(--sky-accent-red);" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
											</svg>
										</div>
										<div class="flex-1 min-w-0">
											<p class="text-sm font-medium" style="color: var(--sky-text-primary);">هشدار: فضای ذخیره‌سازی پر است</p>
											<p class="text-xs mt-1" style="color: var(--sky-text-secondary);">۲ ساعت پیش</p>
										</div>
									</div>
								</div>
							</div>
							<div class="px-4 py-3 text-center" style="border-top: 1px solid var(--sky-border);">
								<a href="/notifications" class="text-sm font-medium hover:underline" style="color: var(--sky-accent-blue);">
									مشاهده همه اعلان‌ها
								</a>
							</div>
						</div>
					{/if}
				</div>
			</div>
		</header>

		<!-- Mobile header -->
		<header class="sticky top-0 z-10 px-4 py-3 flex items-center justify-between lg:hidden"
			style="background: var(--sky-bg-panel); border-bottom: 1px solid var(--sky-border);">
			<button onclick={() => mobileOpen = true} class="p-2 rounded-lg transition-colors hover:bg-[#0f3460]" style="color: var(--sky-text-secondary);" aria-label="منو">
				<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="1.5">
					<path stroke-linecap="round" stroke-linejoin="round" d="M3.75 6.75h16.5M3.75 12h16.5m-16.5 5.25h16.5" />
				</svg>
			</button>
			<h1 class="font-extrabold" style="color: var(--sky-text-primary);">آی‌روم</h1>
			<!-- Mobile notification bell -->
			<button class="p-2 rounded-lg transition-colors hover:bg-[#0f3460] relative"
				style="color: var(--sky-text-secondary);"
				onclick={() => showNotifications = !showNotifications}>
				<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="1.5">
					<path stroke-linecap="round" stroke-linejoin="round" d="M14.857 17.082a23.848 23.848 0 005.454-1.31A8.967 8.967 0 0118 9.75v-.7V9A6 6 0 006 9v.75a8.967 8.967 0 01-2.312 6.022c1.733.64 3.56 1.085 5.455 1.31m5.714 0a24.255 24.255 0 01-5.714 0m5.714 0a3 3 0 11-5.714 0" />
				</svg>
				{#if unreadCount > 0}
					<span class="absolute -top-1 -right-1 w-4 h-4 flex items-center justify-center text-[9px] font-bold text-white rounded-full"
						style="background: var(--sky-accent-red);">
						{toPersianNum(unreadCount)}
					</span>
				{/if}
			</button>
		</header>

		<main class="p-4 lg:p-8 max-w-7xl mx-auto" style="color: var(--sky-text-primary);">
			{@render children()}
		</main>
	</div>
</div>

<style>
	/* Skyroom Dark Sidebar Link Styles */
	:global(.sky-sidebar-link) {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		padding: 0.625rem 0.875rem;
		border-radius: 0.75rem;
		font-size: 0.875rem;
		font-weight: 500;
		color: var(--sky-text-secondary);
		transition: all 0.15s ease;
		position: relative;
	}

	:global(.sky-sidebar-link:hover) {
		background: #0f3460;
		color: var(--sky-text-primary);
	}

	:global(.sky-sidebar-link.active) {
		background: linear-gradient(135deg, #1a56db, #2563eb);
		color: white;
		font-weight: 600;
		box-shadow: inset 3px 0 0 0 #4361ee, 0 4px 12px rgba(26, 86, 219, 0.3);
	}

	:global(.sky-sidebar-link.active::before) {
		content: '';
		position: absolute;
		right: 0;
		top: 50%;
		transform: translateY(-50%);
		width: 3px;
		height: 60%;
		background: #4361ee;
		border-radius: 2px 0 0 2px;
	}

		/* Tooltip for collapsed sidebar */
		:global(.sky-sidebar-link[title]:hover::after) {
			content: attr(title);
			position: absolute;
			right: 100%;
			top: 50%;
			transform: translateY(-50%);
			background: var(--sky-bg-panel);
			color: var(--sky-text-primary);
			padding: 0.5rem 0.75rem;
			border-radius: 0.5rem;
			font-size: 0.75rem;
			white-space: nowrap;
			margin-right: 0.5rem;
			box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
			border: 1px solid var(--sky-border);
			z-index: 50;
			pointer-events: none;
		}

		/* Logout button hover effect */
		:global(.logout-btn:hover) {
			color: #e94560 !important;
		}
</style>
