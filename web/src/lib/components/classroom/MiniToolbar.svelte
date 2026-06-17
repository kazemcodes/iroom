<script lang="ts">
	let {
		usersPanelOpen = $bindable(true),
		chatPanelOpen = $bindable(true),
		onToggleUsers,
		onToggleChat,
	}: {
		usersPanelOpen: boolean;
		chatPanelOpen: boolean;
		onToggleUsers: () => void;
		onToggleChat: () => void;
	} = $props();
</script>

<div class="flex items-center gap-2">
	<!-- Users Panel Toggle -->
	<button onclick={onToggleUsers} class="w-10 h-10 rounded-full flex items-center justify-center transition-all duration-150 {usersPanelOpen ? 'bg-[#23b9d7] text-white' : 'bg-[#252540] border border-[#3a3a5a] text-[#94a3b8] hover:bg-[#2d2d4a]'}" title="کاربران">
		<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z" /></svg>
	</button>

	<!-- Chat Panel Toggle -->
	<button onclick={onToggleChat} class="w-10 h-10 rounded-full flex items-center justify-center transition-all duration-150 {chatPanelOpen ? 'bg-[#23b9d7] text-white' : 'bg-[#252540] border border-[#3a3a5a] text-[#94a3b8] hover:bg-[#2d2d4a]'}" title="گفتگو">
		<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" /></svg>
	</button>
</div>
