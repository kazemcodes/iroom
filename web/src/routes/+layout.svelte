<script lang="ts">
	import { auth, isAdmin } from '$lib/stores';
	import { page } from '$app/state';

	let { children } = $props();
	let currentPath = $derived(page.url.pathname);

	let mobileOpen = $state(false);

	const navItems = $derived.by(() => {
		const items = [
			{ href: '/dashboard', label: 'داشبورد', icon: 'M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-4 0h4' },
			{ href: '/classes', label: 'کلاس‌ها', icon: 'M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253' },
			{ href: '/sessions', label: 'جلسات', icon: 'M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z' }
		];

		if ($isAdmin) {
			items.push({ href: '/admin', label: 'مدیریت', icon: 'M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.066 2.573c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.573 1.066c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.066-2.573c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z M15 12a3 3 0 11-6 0 3 3 0 016 0z' });
		}

		return items;
	});
</script>

<div class="flex min-h-screen">
	<!-- Sidebar -->
	<aside class="fixed inset-y-0 right-0 z-30 w-64 bg-white border-l shadow-lg transform transition-transform duration-200
		{mobileOpen ? 'translate-x-0' : 'translate-x-full'}
		lg:translate-x-0 lg:static lg:shadow-none">
		<div class="flex flex-col h-full">
			<!-- Logo -->
			<div class="flex items-center gap-3 px-6 py-5 border-b">
				<div class="w-10 h-10 bg-blue-600 rounded-xl flex items-center justify-center text-white font-bold text-lg">آ</div>
				<div>
					<h1 class="font-bold text-lg text-gray-900">آی‌روم</h1>
					<p class="text-xs text-gray-500">کلاس آنلاین</p>
				</div>
			</div>

			<!-- Navigation -->
			<nav class="flex-1 px-3 py-4 space-y-1">
				{#each navItems as item}
					<a
						href={item.href}
						class="flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm font-medium transition-colors
							{currentPath.startsWith(item.href) ? 'bg-blue-50 text-blue-700' : 'text-gray-600 hover:bg-gray-50 hover:text-gray-900'}"
						onclick={() => mobileOpen = false}
					>
						<svg class="w-5 h-5 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d={item.icon} />
						</svg>
						{item.label}
					</a>
				{/each}
			</nav>

			<!-- User info -->
			<div class="px-4 py-4 border-t">
				{#if $auth.user}
					<div class="flex items-center gap-3">
						<div class="w-9 h-9 rounded-full bg-blue-100 flex items-center justify-center text-blue-700 font-bold text-sm">
							{$auth.user.display_name.charAt(0)}
						</div>
						<div class="flex-1 min-w-0">
							<p class="text-sm font-medium text-gray-900 truncate">{$auth.user.display_name}</p>
							<p class="text-xs text-gray-500">
								{$auth.user.role === 'admin' ? 'مدیر' : $auth.user.role === 'teacher' ? 'مدرس' : 'دانش‌آموز'}
							</p>
						</div>
					</div>
				{/if}
				<button
					onclick={() => auth.logout()}
					class="mt-3 w-full flex items-center justify-center gap-2 px-3 py-2 text-sm text-red-600 hover:bg-red-50 rounded-lg transition-colors"
				>
					<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
					</svg>
					خروج
				</button>
			</div>
		</div>
	</aside>

	<!-- Mobile overlay -->
	{#if mobileOpen}
		<div class="fixed inset-0 bg-black/30 z-20 lg:hidden" onclick={() => mobileOpen = false}></div>
	{/if}

	<!-- Main content -->
	<div class="flex-1 min-w-0">
		<!-- Mobile header -->
		<header class="sticky top-0 z-10 bg-white border-b px-4 py-3 flex items-center justify-between lg:hidden">
			<button onclick={() => mobileOpen = true} class="p-2 hover:bg-gray-100 rounded-lg">
				<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M4 6h16M4 12h16M4 18h16" />
				</svg>
			</button>
			<h1 class="font-bold text-blue-600">آی‌روم</h1>
			<div class="w-9"></div>
		</header>

		<main class="p-4 lg:p-8">
			{@render children()}
		</main>
	</div>
</div>
