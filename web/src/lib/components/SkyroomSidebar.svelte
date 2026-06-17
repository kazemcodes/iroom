<script lang="ts">
	let {
		collapsed = $bindable(false),
		currentPath = '',
		isAdmin = false,
	}: {
		collapsed: boolean;
		currentPath: string;
		isAdmin: boolean;
	} = $props();

	const navItems = [
		{ href: '/dashboard', label: 'داشبورد', icon: '📊' },
		{ href: '/classes', label: 'کلاس‌ها', icon: '📚' },
		{ href: '/sessions', label: 'جلسات', icon: '🎥' },
		{ href: '/files', label: 'فایل‌ها', icon: '📁' },
		{ href: '/support', label: 'پشتیبانی', icon: '🎧' },
		{ href: '/profile', label: 'حساب کاربری', icon: '👤' },
	];

	const adminNavItems = [
		{ href: '/admin', label: 'داشبورد مدیریت', icon: '📈' },
		{ href: '/admin/users', label: 'کاربران', icon: '👥' },
		{ href: '/admin/rooms', label: 'اتاق‌ها', icon: '🏢' },
		{ href: '/admin/recordings', label: 'ضبط‌ها', icon: '⏺' },
		{ href: '/admin/logs', label: 'لاگ‌ها', icon: '📋' },
		{ href: '/admin/settings', label: 'تنظیمات', icon: '⚙️' },
	];
</script>

<aside class="fixed inset-y-0 right-0 z-30 flex flex-col transition-all duration-200"
	style="width: {collapsed ? '60px' : '260px'}; background-color: #1c293a;">

	<!-- Logo -->
	<div class="flex items-center gap-3 px-4 py-4" style="border-bottom: 1px solid rgba(255,255,255,0.1);">
		<div class="w-8 h-8 rounded-lg flex items-center justify-center text-white font-bold text-sm shrink-0" style="background: #23b9d7;">
			آ
		</div>
		{#if !collapsed}
			<span class="text-sm font-bold text-[#e2e8f0]">آی‌روم</span>
		{/if}
	</div>

	<!-- Navigation -->
	<nav class="flex-1 px-2 py-3 space-y-0.5 overflow-y-auto">
		{#each navItems as item}
			<a href={item.href} class="flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm transition-all {currentPath === item.href || (item.href !== '/admin' && currentPath.startsWith(item.href)) ? 'bg-white/10 text-[#e2e8f0] font-medium' : 'text-[#94a3b8] hover:bg-white/5 hover:text-[#e2e8f0]'}">
				<span class="text-base">{item.icon}</span>
				{#if !collapsed}
					<span>{item.label}</span>
				{/if}
			</a>
		{/each}

		{#if isAdmin}
			<div class="pt-2 mt-2" style="border-top: 1px solid rgba(255,255,255,0.1);">
				{#if !collapsed}
					<p class="px-3 mb-1 text-[10px] font-bold text-[#6790a0] uppercase tracking-wider">مدیریت</p>
				{/if}
				{#each adminNavItems as item}
					<a href={item.href} class="flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm transition-all {currentPath === item.href || (item.href === '/admin' && currentPath === '/admin') ? 'bg-white/10 text-[#e2e8f0] font-medium' : 'text-[#94a3b8] hover:bg-white/5 hover:text-[#e2e8f0]'}">
						<span class="text-base">{item.icon}</span>
						{#if !collapsed}
							<span>{item.label}</span>
						{/if}
					</a>
				{/each}
			</div>
		{/if}
	</nav>

	<!-- Footer -->
	<div class="px-3 py-3" style="border-top: 1px solid rgba(255,255,255,0.1);">
		{#if !collapsed}
			<p class="text-[10px] text-[#6790a0] text-center">© آی‌روم</p>
		{/if}
	</div>
</aside>

<!-- Mobile overlay -->
{#if !collapsed}
	<div class="fixed inset-0 bg-black/50 z-20 lg:hidden" onclick={() => collapsed = true} role="button" tabindex="-1"></div>
{/if}
