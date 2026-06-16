<script lang="ts">
	import '../app.css';
	import { auth, isAdmin } from '$lib/stores';
	import { page } from '$app/state';

	let { children } = $props();
	let currentPath = $derived(page.url.pathname);
	let mobileOpen = $state(false);
	let sidebarCollapsed = $state(false);

	const isOnAdmin = $derived(currentPath.startsWith('/admin'));

	const navItems = $derived.by(() => {
		const items = [
			{ href: '/dashboard', label: 'داشبورد', icon: 'M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-4 0h4' },
			{ href: '/classes', label: 'کلاس‌ها', icon: 'M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253' },
			{ href: '/sessions', label: 'جلسات', icon: 'M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z' },
			{ href: '/files', label: 'فایل‌ها', icon: 'M19.5 14.25v-2.625a3.375 3.375 0 00-3.375-3.375h-1.5A1.125 1.125 0 0113.5 7.125v-1.5a3.375 3.375 0 00-3.375-3.375H8.25m6.75 12H9.75m0-3h6m-6 6h6M3.375 6.75h17.25a.375.375 0 01.375.375v11.25a.375.375 0 01-.375.375H3.375a.375.375 0 01-.375-.375V7.125a.375.375 0 01.375-.375z' },
			{ href: '/support', label: 'پشتیبانی', icon: 'M20.25 8.511c.884.284 1.5 1.128 1.5 2.097v4.286c0 1.136-.847 2.1-1.98 2.193-.34.027-.68.052-1.02.072v3.091l-3-3c-1.354 0-2.694-.055-4.02-.163a2.115 2.115 0 01-.825-.242m9.345-8.334a2.126 2.126 0 00-.476-.095 48.64 48.64 0 00-8.048 0c-1.131.094-1.976.94-1.976 2.097v4.286c0 .837.46 1.58 1.155 1.951m9.345-8.334V6.637c0-1.621-1.152-3.026-2.76-3.235A48.455 48.455 0 0011.25 3c-2.115 0-4.198.137-6.24.402-1.608.209-2.76 1.614-2.76 3.235v6.226c0 1.621 1.152 3.026 2.76 3.235.577.075 1.157.14 1.74.194V21l4.155-4.155' }
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
			{ href: '/admin/rooms', label: 'اتاق‌ها', icon: 'M2.25 21h19.5m-18-18v18m10.5-18v18m6-13.5V21M6.75 6.75h.75m-.75 3h.75m-.75 3h.75m3-6h.75m-.75 3h.75m-.75 3h.75M6.75 21v-3.375c0-.621.504-1.125 1.125-1.125h2.25c.621 0 1.125.504 1.125 1.125V21M3 3h12m-.75 4.5H21m-3.75 3H21m-3.75 3H21' },
			{ href: '/support', label: 'تیکت‌ها', icon: 'M20.25 8.511c.884.284 1.5 1.128 1.5 2.097v4.286c0 1.136-.847 2.1-1.98 2.193-.34.027-.68.052-1.02.072v3.091l-3-3c-1.354 0-2.694-.055-4.02-.163a2.115 2.115 0 01-.825-.242m9.345-8.334a2.126 2.126 0 00-.476-.095 48.64 48.64 0 00-8.048 0c-1.131.094-1.976.94-1.976 2.097v4.286c0 .837.46 1.58 1.155 1.951m9.345-8.334V6.637c0-1.621-1.152-3.026-2.76-3.235A48.455 48.455 0 0011.25 3c-2.115 0-4.198.137-6.24.402-1.608.209-2.76 1.614-2.76 3.235v6.226c0 1.621 1.152 3.026 2.76 3.235.577.075 1.157.14 1.74.194V21l4.155-4.155' },
			{ href: '/admin/recordings', label: 'ضبط‌ها', icon: 'M15.75 10.5l4.72-4.72a.75.75 0 011.28.53v11.38a.75.75 0 01-1.28.53l-4.72-4.72M4.5 18.75h9a2.25 2.25 0 002.25-2.25v-9a2.25 2.25 0 00-2.25-2.25h-9A2.25 2.25 0 002.25 7.5v9a2.25 2.25 0 002.25 2.25z' },
			{ href: '/admin/logs', label: 'لاگ‌ها', icon: 'M19.5 14.25v-2.625a3.375 3.375 0 00-3.375-3.375h-1.5A1.125 1.125 0 0113.5 7.125v-1.5a3.375 3.375 0 00-3.375-3.375H8.25m0 12.75h7.5m-7.5 3H12M10.5 2.25H5.625c-.621 0-1.125.504-1.125 1.125v17.25c0 .621.504 1.125 1.125 1.125h12.75c.621 0 1.125-.504 1.125-1.125V11.25a9 9 0 00-9-9z' },
			{ href: '/admin/settings', label: 'تنظیمات', icon: 'M9.594 3.94c.09-.542.56-.94 1.11-.94h2.593c.55 0 1.02.398 1.11.94l.213 1.281c.063.374.313.686.645.87.074.04.147.083.22.127.325.196.72.257 1.075.124l1.217-.456a1.125 1.125 0 011.37.49l1.296 2.247a1.125 1.125 0 01-.26 1.431l-1.003.827c-.293.241-.438.613-.43.992a7.723 7.723 0 010 .255c-.008.378.137.75.43.991l1.004.827c.424.35.534.955.26 1.43l-1.298 2.247a1.125 1.125 0 01-1.369.491l-1.217-.456c-.355-.133-.75-.072-1.076.124a6.47 6.47 0 01-.22.128c-.331.183-.581.495-.644.869l-.213 1.281c-.09.543-.56.94-1.11.94h-2.594c-.55 0-1.019-.398-1.11-.94l-.213-1.281c-.062-.374-.312-.686-.644-.87a6.52 6.52 0 01-.22-.127c-.325-.196-.72-.257-1.076-.124l-1.217.456a1.125 1.125 0 01-1.369-.49l-1.297-2.247a1.125 1.125 0 01.26-1.431l1.004-.827c.292-.24.437-.613.43-.991a6.932 6.932 0 010-.255c.007-.38-.138-.751-.43-.992l-1.004-.827a1.125 1.125 0 01-.26-1.43l1.297-2.247a1.125 1.125 0 011.37-.491l1.216.456c.356.133.751.072 1.076-.124.072-.044.146-.086.22-.128.332-.183.582-.495.644-.869l.214-1.28z M15 12a3 3 0 11-6 0 3 3 0 016 0z' }
		];
	});

	const roleLabels: Record<string, string> = { admin: 'مدیر سیستم', teacher: 'مدرس', student: 'دانش‌آموز' };
	const roleBadgeColors: Record<string, string> = {
		admin: 'bg-amber-100 text-amber-700',
		teacher: 'bg-violet-100 text-violet-700',
		student: 'bg-teal-100 text-teal-700'
	};
</script>

<div class="flex min-h-screen">
	<!-- Sidebar -->
	<aside class="fixed inset-y-0 right-0 z-30 bg-white/80 backdrop-blur-xl border-l border-gray-100 shadow-xl shadow-gray-200/50 transform transition-all duration-300
		{mobileOpen ? 'translate-x-0' : 'translate-x-full'}
		lg:translate-x-0 lg:static lg:shadow-none
		{sidebarCollapsed ? 'w-[68px]' : 'w-72'}">
		<div class="flex flex-col h-full">
			<!-- Logo -->
			<div class="px-6 py-6 border-b border-gray-100">
				<div class="flex items-center gap-3">
					<div class="w-11 h-11 rounded-xl flex items-center justify-center text-white font-bold text-lg shadow-lg shadow-blue-500/25 shrink-0"
						style="background: linear-gradient(135deg, #1a56db, #0891b2);">
						آ
					</div>
					{#if !sidebarCollapsed}
						<div class="min-w-0">
							<h1 class="font-extrabold text-lg gradient-text">آی‌روم</h1>
							<p class="text-[11px] text-gray-400 font-medium">کلاس آنلاین هوشمند</p>
						</div>
					{/if}
				</div>
			</div>

			<!-- Collapse toggle (desktop) -->
			<div class="hidden lg:flex justify-start px-3 pt-2">
				<button onclick={() => sidebarCollapsed = !sidebarCollapsed} class="p-1.5 rounded-lg text-gray-400 hover:text-gray-600 hover:bg-gray-100 transition-colors" title={sidebarCollapsed ? 'باز کردن منو' : 'بستن منو'}>
					<svg class="w-4 h-4 transition-transform {sidebarCollapsed ? 'rotate-180' : ''}" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 19l-7-7 7-7m8 14l-7-7 7-7" /></svg>
				</button>
			</div>

			<!-- Navigation -->
			<nav class="flex-1 px-3 py-4 space-y-1 overflow-y-auto">
				{#each navItems as item}
					<a
						href={item.href}
						class="sidebar-link {currentPath === item.href || (item.href !== '/admin' && currentPath.startsWith(item.href)) ? 'active' : ''}"
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

				{#if $isAdmin && adminNavItems.length > 0}
					<div class="pt-3 mt-3 border-t border-gray-100">
						{#if !sidebarCollapsed}
							<p class="px-3 mb-2 text-[10px] font-bold text-gray-400 uppercase tracking-wider">مدیریت</p>
						{/if}
						{#each adminNavItems as item}
							<a
								href={item.href}
								class="sidebar-link {currentPath === item.href || (item.href === '/admin' && currentPath === '/admin') ? 'active' : ''}"
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
			<div class="px-4 py-4 border-t border-gray-100">
				{#if $auth.user}
					<div class="flex items-center gap-3 mb-3">
						<div class="w-10 h-10 rounded-xl flex items-center justify-center text-white font-bold text-sm shadow-md shrink-0"
							style="background: linear-gradient(135deg, #1a56db, #7c3aed);">
							{$auth.user.display_name.charAt(0)}
						</div>
						{#if !sidebarCollapsed}
							<div class="flex-1 min-w-0">
								<p class="text-sm font-bold text-gray-800 truncate">{$auth.user.display_name}</p>
								<span class="text-[11px] px-2 py-0.5 rounded-full font-semibold {roleBadgeColors[$auth.user.role]}">
									{roleLabels[$auth.user.role]}
								</span>
							</div>
						{/if}
					</div>
				{/if}
				<button
					onclick={() => auth.logout()}
					class="w-full flex items-center justify-center gap-2 px-3 py-2.5 text-sm text-gray-500 hover:text-red-600 hover:bg-red-50 rounded-xl transition-all font-medium"
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
		<div class="fixed inset-0 bg-black/20 backdrop-blur-sm z-20 lg:hidden" onclick={() => mobileOpen = false} role="button" tabindex="-1"></div>
	{/if}

	<!-- Main content -->
	<div class="flex-1 min-w-0">
		<!-- Mobile header -->
		<header class="sticky top-0 z-10 bg-white/80 backdrop-blur-xl border-b border-gray-100 px-4 py-3 flex items-center justify-between lg:hidden">
			<button onclick={() => mobileOpen = true} class="p-2.5 hover:bg-gray-100 rounded-xl transition-colors" aria-label="منو">
				<svg class="w-5 h-5 text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="1.5">
					<path stroke-linecap="round" stroke-linejoin="round" d="M3.75 6.75h16.5M3.75 12h16.5m-16.5 5.25h16.5" />
				</svg>
			</button>
			<h1 class="font-extrabold gradient-text">آی‌روم</h1>
			<div class="w-10"></div>
		</header>

		<main class="p-4 lg:p-8 max-w-7xl mx-auto">
			{@render children()}
		</main>
	</div>
</div>
