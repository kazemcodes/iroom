<script lang="ts">
	let {
		isInSidebar = false,
		isLoading = false,
	}: {
		isInSidebar: boolean;
		isLoading: boolean;
	} = $props();
</script>

<div class="flex flex-col h-full {isInSidebar ? '' : ''}">
	{#if isInSidebar}
		<div class="flex items-center gap-2 px-3 py-2.5 shrink-0" style="border-bottom: 1px solid #3a3a5a;">
			<svg class="w-4 h-4 text-[#94a3b8]" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M7 4v16M17 4v16M3 8h4m10 0h4M3 12h18M3 16h4m10 0h4M4 20h16a1 1 0 001-1V5a1 1 0 00-1-1H4a1 1 0 00-1 1v14a1 1 0 001 1z" /></svg>
			<span class="text-xs font-medium text-[#94a3b8]">اسلاید</span>
		</div>
	{/if}

	<div class="relative flex-1 min-h-0" style="background-color: #000;">
		{#if isLoading}
			<div class="absolute inset-0 flex items-center justify-center">
				<svg class="w-10 h-10 animate-spin" viewBox="0 0 24 24" fill="none">
					<circle cx="12" cy="12" r="10" stroke="#3a3a5a" stroke-width="3"></circle>
					<path d="M12 2a10 10 0 019.95 9" stroke="#23b9d7" stroke-width="3" stroke-linecap="round"></path>
				</svg>
			</div>
		{/if}
		<slot />
	</div>
</div>
