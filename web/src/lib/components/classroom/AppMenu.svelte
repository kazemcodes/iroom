<script lang="ts">
	let {
		showUsersPanel = $bindable(true),
		showChatPanel = $bindable(true),
		isOwner = false,
		onSettings,
		onExit,
		onCloseRoom,
		onClose,
	}: {
		showUsersPanel: boolean;
		showChatPanel: boolean;
		isOwner: boolean;
		onSettings: () => void;
		onExit: () => void;
		onCloseRoom: () => void;
		onClose: () => void;
	} = $props();
</script>

<div class="absolute top-12 left-4 z-50 w-56 rounded-lg overflow-hidden shadow-xl" style="background: #ffffff; border: 1px solid #e0e4eb; border-radius: 8px; box-shadow: 0 4px 20px rgba(0,0,0,0.3);">
	<!-- Menu Items -->
	<div class="py-1">
		<!-- User Info -->
		<button class="w-full px-4 py-2.5 flex items-center gap-3 text-[13px] text-[#1c293a] hover:bg-[#f0f2f5] transition-colors">
			<svg class="w-4 h-4 text-[#94a3b8]" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
			اطلاعات کاربری
		</button>

		<!-- Connection Status -->
		<button class="w-full px-4 py-2.5 flex items-center gap-3 text-[13px] text-[#1c293a] hover:bg-[#f0f2f5] transition-colors">
			<svg class="w-4 h-4 text-[#94a3b8]" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" /></svg>
			وضعیت اتصال
		</button>

		<!-- Settings -->
		<button onclick={() => { onSettings(); onClose(); }} class="w-full px-4 py-2.5 flex items-center gap-3 text-[13px] text-[#1c293a] hover:bg-[#f0f2f5] transition-colors">
			<svg class="w-4 h-4 text-[#94a3b8]" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" /><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" /></svg>
			تنظیمات
		</button>

		<!-- Layout -->
		<button class="w-full px-4 py-2.5 flex items-center gap-3 text-[13px] text-[#1c293a] hover:bg-[#f0f2f5] transition-colors">
			<svg class="w-4 h-4 text-[#94a3b8]" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M4 5a1 1 0 011-1h14a1 1 0 011 1v2a1 1 0 01-1 1H5a1 1 0 01-1-1V5zM4 13a1 1 0 011-1h6a1 1 0 011 1v6a1 1 0 01-1 1H5a1 1 0 01-1-1v-6zM16 13a1 1 0 011-1h2a1 1 0 011 1v6a1 1 0 01-1 1h-2a1 1 0 01-1-1v-6z" /></svg>
			چیدمان
			<svg class="w-3 h-3 text-[#94a3b8] mr-auto" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" /></svg>
		</button>
	</div>

	<!-- Separator -->
	<div class="h-px bg-[#e0e4eb]"></div>

	<!-- Exit Actions -->
	<div class="py-1">
		<button onclick={() => { onExit(); onClose(); }} class="w-full px-4 py-2.5 flex items-center gap-3 text-[13px] text-[#1c293a] hover:bg-[#f0f2f5] transition-colors">
			<svg class="w-4 h-4 text-[#94a3b8]" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" /></svg>
			خروج
		</button>

		{#if isOwner}
			<button onclick={() => { onCloseRoom(); onClose(); }} class="w-full px-4 py-2.5 flex items-center gap-3 text-[13px] text-[#e05252] hover:bg-[#fde8e8] transition-colors">
				<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M5.586 15H4a1 1 0 01-1-1v-4a1 1 0 011-1h1.586l4.707-4.707C10.923 3.663 12 4.109 12 5v14c0 .891-1.077 1.337-1.707.707L5.586 15z" /><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M17 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2" /></svg>
				بستن اتاق
			</button>
		{/if}
	</div>
</div>
