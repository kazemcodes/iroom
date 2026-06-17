<script lang="ts">
	import type { Participant } from '$lib/classroom/types';
	import { ROLE_LABELS } from '$lib/classroom/types';

	let {
		participant,
		isAdmin = false,
		isOwner = false,
		onMuteAudio,
		onMuteVideo,
		onKick,
	}: {
		participant: Participant;
		isAdmin: boolean;
		isOwner: boolean;
		onMuteAudio?: (id: string) => void;
		onMuteVideo?: (id: string) => void;
		onKick?: (id: string) => void;
	} = $props();

	const canControl = $derived(isAdmin && !participant.isLocal && participant.role !== 'owner');
</script>

<div class="flex items-center gap-2 px-3 py-2 transition-colors hover:bg-white/5 {participant.isSpeaking ? 'bg-white/5' : ''}">
	<!-- Avatar -->
	<div class="relative">
		<div class="w-8 h-8 rounded-full flex items-center justify-center text-xs font-bold {participant.isLocal ? 'bg-[#23b9d7] text-white' : 'bg-[#3a3a5a] text-[#e2e8f0]'}">
			{participant.name.charAt(0)}
		</div>
		{#if participant.isSpeaking}
			<span class="absolute -bottom-0.5 -right-0.5 w-2.5 h-2.5 bg-[#40bf7f] rounded-full border-2" style="border-color: #252540;"></span>
		{/if}
	</div>

	<!-- Info -->
	<div class="flex-1 min-w-0">
		<p class="text-[13px] font-medium truncate text-[#e2e8f0]">
			{participant.name}{participant.isLocal ? ' (شما)' : ''}
			{#if participant.handRaised}<span class="text-yellow-400 ml-1">✋</span>{/if}
		</p>
	</div>

	<!-- Status Indicators -->
	<div class="flex items-center gap-1">
		{#if participant.hasAudio}
			<span class="w-1.5 h-1.5 bg-[#40bf7f] rounded-full"></span>
		{:else}
			<span class="w-1.5 h-1.5 bg-[#e05252] rounded-full"></span>
		{/if}
		{#if participant.hasVideo}
			<span class="w-1.5 h-1.5 bg-[#40bf7f] rounded-full"></span>
		{:else}
			<span class="w-1.5 h-1.5 bg-[#94a3b8] rounded-full"></span>
		{/if}
	</div>

	<!-- Action Buttons (admin only) -->
	{#if canControl}
		<div class="flex items-center gap-0.5 ml-1">
			<button onclick={() => onMuteAudio?.(participant.id)} class="w-6 h-6 rounded flex items-center justify-center text-[#94a3b8] hover:text-[#e2e8f0] hover:bg-white/10" title="بی‌صدا کردن">
				<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M5.586 15H4a1 1 0 01-1-1v-4a1 1 0 011-1h1.586l4.707-4.707C10.923 3.663 12 4.109 12 5v14c0 .891-1.077 1.337-1.707.707L5.586 15z" /><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M17 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2" /></svg>
			</button>
			<button onclick={() => onMuteVideo?.(participant.id)} class="w-6 h-6 rounded flex items-center justify-center text-[#94a3b8] hover:text-[#e2e8f0] hover:bg-white/10" title="خاموش کردن ویدیو">
				<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" /><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M17 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2" /></svg>
			</button>
			{#if isOwner}
				<button onclick={() => onKick?.(participant.id)} class="w-6 h-6 rounded flex items-center justify-center text-[#94a3b8] hover:text-[#e05252] hover:bg-white/10" title="حذف">
					<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M6 18L18 6M6 6l12 12" /></svg>
				</button>
			{/if}
		</div>
	{/if}
</div>
