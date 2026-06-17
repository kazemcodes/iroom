<script lang="ts">
	let {
		onToggleSidebar,
		userName = '',
		userRole = '',
	}: {
		onToggleSidebar: () => void;
		userName: string;
		userRole: string;
	} = $props();

	let showUserMenu = $state(false);
</script>

<header class="sticky top-0 z-10 flex items-center justify-between px-4 h-14" style="background: white; border-bottom: 1px solid #e0e4eb;">
	<!-- Left: Hamburger + Breadcrumb -->
	<div class="flex items-center gap-3">
		<button onclick={onToggleSidebar} class="p-2 rounded-lg text-[#94a3b8] hover:bg-[#f0f2f5] transition-colors lg:hidden">
			<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M3.75 6.75h16.5M3.75 12h16.5m-16.5 5.25h16.5" /></svg>
		</button>
	</div>

	<!-- Right: User Menu -->
	<div class="flex items-center gap-3">
		<button onclick={() => showUserMenu = !showUserMenu} class="flex items-center gap-2 px-3 py-1.5 rounded-lg hover:bg-[#f0f2f5] transition-colors">
			<div class="w-8 h-8 rounded-full flex items-center justify-center text-white text-xs font-bold" style="background: #23b9d7;">
				{userName.charAt(0)}
			</div>
			<span class="text-sm font-medium text-[#1c293a] hidden sm:block">{userName}</span>
			<svg class="w-4 h-4 text-[#94a3b8]" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" /></svg>
		</button>

		{#if showUserMenu}
			<div class="absolute top-12 left-4 w-48 rounded-lg overflow-hidden shadow-xl" style="background: white; border: 1px solid #e0e4eb;">
				<a href="/profile" class="block px-4 py-2.5 text-[13px] text-[#1c293a] hover:bg-[#f0f2f5]">حساب کاربری</a>
				<a href="/dashboard" class="block px-4 py-2.5 text-[13px] text-[#1c293a] hover:bg-[#f0f2f5]">داشبورد</a>
				<div class="h-px bg-[#e0e4eb]"></div>
				<a href="/auth" class="block px-4 py-2.5 text-[13px] text-[#e05252] hover:bg-[#fde8e8]">خروج</a>
			</div>
		{/if}
	</div>
</header>
