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

<div class="absolute top-12 right-4 z-50 w-56 rounded-xl shadow-2xl overflow-hidden" style="background-color: #1e1e3a; border: 1px solid #2a2a4a;">
	<!-- User Info -->
	<div class="px-4 py-3 border-b" style="border-color: #2a2a4a;">
		<p class="text-xs text-gray-400">تنظیمات اتاق</p>
	</div>

	<!-- Menu Items -->
	<div class="py-1">
		<!-- Layout: Show Users -->
		<button onclick={() => { showUsersPanel = !showUsersPanel; }} class="w-full px-4 py-2.5 flex items-center gap-3 text-sm text-gray-300 hover:bg-white/5 transition-colors">
			<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M15 19.128a9.38 9.38 0 002.625.372 9.337 9.337 0 004.121-.952 4.125 4.125 0 00-7.533-2.493M15 19.128v-.003c0-1.113-.285-2.16-.786-3.07M15 19.128v.106A12.318 12.318 0 018.624 21c-2.331 0-4.512-.645-6.374-1.766l-.001-.109a6.375 6.375 0 0111.964-3.07M12 6.375a3.375 3.375 0 11-6.75 0 3.375 3.375 0 016.75 0zm8.25 2.25a2.625 2.625 0 11-5.25 0 2.625 2.625 0 015.25 0z" /></svg>
			<span>شرکت‌کنندگان</span>
			<span class="mr-auto text-[10px] {showUsersPanel ? 'text-green-400' : 'text-gray-500'}">{showUsersPanel ? 'فعال' : 'غیرفعال'}</span>
		</button>

		<!-- Layout: Show Chat -->
		<button onclick={() => { showChatPanel = !showChatPanel; }} class="w-full px-4 py-2.5 flex items-center gap-3 text-sm text-gray-300 hover:bg-white/5 transition-colors">
			<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" /></svg>
			<span>گفتگو</span>
			<span class="mr-auto text-[10px] {showChatPanel ? 'text-green-400' : 'text-gray-500'}">{showChatPanel ? 'فعال' : 'غیرفعال'}</span>
		</button>

		<!-- Settings -->
		<button onclick={() => { onSettings(); onClose(); }} class="w-full px-4 py-2.5 flex items-center gap-3 text-sm text-gray-300 hover:bg-white/5 transition-colors">
			<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" /><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" /></svg>
			<span>تنظیمات</span>
		</button>
	</div>

	<!-- Exit Actions -->
	<div class="border-t py-1" style="border-color: #2a2a4a;">
		<button onclick={() => { onExit(); onClose(); }} class="w-full px-4 py-2.5 flex items-center gap-3 text-sm text-gray-300 hover:bg-white/5 transition-colors">
			<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" /></svg>
			<span>خروج از اتاق</span>
		</button>

		{#if isOwner}
			<button onclick={() => { onCloseRoom(); onClose(); }} class="w-full px-4 py-2.5 flex items-center gap-3 text-sm text-red-400 hover:bg-red-600/10 transition-colors">
				<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M5.586 15H4a1 1 0 01-1-1v-4a1 1 0 011-1h1.586l4.707-4.707C10.923 3.663 12 4.109 12 5v14c0 .891-1.077 1.337-1.707.707L5.586 15z" /><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M17 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2" /></svg>
				<span>بستن اتاق</span>
			</button>
		{/if}
	</div>
</div>
