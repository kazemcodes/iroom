<script lang="ts">
	import '../app.css';
	import { auth, isAdmin } from '$lib/stores';
	import { page } from '$app/state';

	let { children } = $props();
	let currentPath = $derived(page.url.pathname);
	let mobileOpen = $state(false);

	const navItems = $derived.by(() => {
		const items = [
			{ href: '/dashboard', label: 'داشبورد', icon: 'M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-4 0h4' },
			{ href: '/classes', label: 'کلاس‌ها', icon: 'M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253' },
			{ href: '/sessions', label: 'جلسات', icon: 'M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z' },
			{ href: '/files', label: 'فایل‌ها', icon: 'M19.5 14.25v-2.625a3.375 3.375 0 00-3.375-3.375h-1.5A1.125 1.125 0 0113.5 7.125v-1.5a3.375 3.375 0 00-3.375-3.375H8.25m6.75 12H9.75m0-3h6m-6 6h6M3.375 6.75h17.25a.375.375 0 01.375.375v11.25a.375.375 0 01-.375.375H3.375a.375.375 0 01-.375-.375V7.125a.375.375 0 01.375-.375z' },
			{ href: '/support', label: 'پشتیبانی', icon: 'M20.25 8.511c.884.284 1.5 1.128 1.5 2.097v4.286c0 1.136-.847 2.1-1.98 2.193-.34.027-.68.052-1.02.072v3.091l-3-3c-1.354 0-2.694-.055-4.02-.163a2.115 2.115 0 01-.825-.242m9.345-8.334a2.126 2.126 0 00-.476-.095 48.64 48.64 0 00-8.048 0c-1.131.094-1.976.94-1.976 2.097v4.286c0 .837.46 1.58 1.155 1.951m9.345-8.334V6.637c0-1.621-1.152-3.026-2.76-3.235A48.455 48.455 0 0011.25 3c-2.115 0-4.198.137-6.24.402-1.608.209-2.76 1.614-2.76 3.235v6.226c0 1.621 1.152 3.026 2.76 3.235.577.075 1.157.14 1.74.194V21l4.155-4.155' }
		];

		if ($isAdmin) {
			items.push({ href: '/admin', label: 'مدیریت', icon: 'M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.066 2.573c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.573 1.066c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.066-2.573c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z M15 12a3 3 0 11-6 0 3 3 0 016 0z' });
			items.push({ href: '/admin/rooms', label: 'اتاق‌ها', icon: 'M2.25 21h19.5m-18-18v18m10.5-18v18m6-13.5V21M6.75 6.75h.75m-.75 3h.75m-.75 3h.75m3-6h.75m-.75 3h.75m-.75 3h.75M6.75 21v-3.375c0-.621.504-1.125 1.125-1.125h2.25c.621 0 1.125.504 1.125 1.125V21M3 3h12m-.75 4.5H21m-3.75 3H21m-3.75 3H21' });
		}

		return items;
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
	<aside class="fixed inset-y-0 right-0 z-30 w-72 bg-white/80 backdrop-blur-xl border-l border-gray-100 shadow-xl shadow-gray-200/50 transform transition-all duration-300
		{mobileOpen ? 'translate-x-0' : 'translate-x-full'}
		lg:translate-x-0 lg:static lg:shadow-none">
		<div class="flex flex-col h-full">
			<!-- Logo -->
			<div class="px-6 py-6 border-b border-gray-100">
				<div class="flex items-center gap-3">
					<div class="w-11 h-11 rounded-xl flex items-center justify-center text-white font-bold text-lg shadow-lg shadow-blue-500/25"
						style="background: linear-gradient(135deg, #1a56db, #0891b2);">
						آ
					</div>
					<div>
						<h1 class="font-extrabold text-lg gradient-text">آی‌روم</h1>
						<p class="text-[11px] text-gray-400 font-medium">کلاس آنلاین هوشمند</p>
					</div>
				</div>
			</div>

			<!-- Navigation -->
			<nav class="flex-1 px-4 py-5 space-y-1">
				{#each navItems as item}
					<a
						href={item.href}
						class="sidebar-link {currentPath.startsWith(item.href) ? 'active' : ''}"
						onclick={() => mobileOpen = false}
					>
						<svg class="w-5 h-5 shrink-0" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" d={item.icon} />
						</svg>
						{item.label}
					</a>
				{/each}
			</nav>

			<!-- User info -->
			<div class="px-4 py-4 border-t border-gray-100">
				{#if $auth.user}
					<div class="flex items-center gap-3 mb-3">
						<div class="w-10 h-10 rounded-xl flex items-center justify-center text-white font-bold text-sm shadow-md"
							style="background: linear-gradient(135deg, #1a56db, #7c3aed);">
							{$auth.user.display_name.charAt(0)}
						</div>
						<div class="flex-1 min-w-0">
							<p class="text-sm font-bold text-gray-800 truncate">{$auth.user.display_name}</p>
							<span class="text-[11px] px-2 py-0.5 rounded-full font-semibold {roleBadgeColors[$auth.user.role]}">
								{roleLabels[$auth.user.role]}
							</span>
						</div>
					</div>
				{/if}
				<button
					onclick={() => auth.logout()}
					class="w-full flex items-center justify-center gap-2 px-3 py-2.5 text-sm text-gray-500 hover:text-red-600 hover:bg-red-50 rounded-xl transition-all font-medium"
				>
					<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="1.5">
						<path stroke-linecap="round" stroke-linejoin="round" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
					</svg>
					خروج از حساب
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
