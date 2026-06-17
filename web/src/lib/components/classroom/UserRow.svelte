<script lang="ts">
	import type { Participant, UserRole } from '$lib/classroom/types';
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

	const canControl = $derived(isAdmin && !participant.isLocal);
</script>

<div class="flex items-center gap-2 px-2 py-1.5 rounded-lg {participant.isSpeaking ? 'bg-blue-900/30' : ''}">
	<div class="relative">
		<div class="w-7 h-7 rounded-full bg-gray-600 flex items-center justify-center text-[10px] font-bold text-white">
			{participant.name.charAt(0)}
		</div>
		{#if participant.isSpeaking}
			<span class="absolute -bottom-0.5 -right-0.5 w-2.5 h-2.5 bg-green-500 rounded-full border-2" style="border-color: #16213e;"></span>
		{/if}
	</div>
	<div class="flex-1 min-w-0">
		<p class="text-xs font-medium truncate">
			{participant.name}{participant.isLocal ? ' (شما)' : ''}
			{#if participant.handRaised}<span class="text-yellow-400 ml-1">✋</span>{/if}
			{#if participant.role !== 'user' && participant.role !== 'student'}
				<span class="text-[9px] text-gray-500 ml-1">({ROLE_LABELS[participant.role as UserRole]})</span>
			{/if}
		</p>
	</div>
	<div class="flex items-center gap-1">
		{#if participant.hasAudio}<span class="w-1.5 h-1.5 bg-green-400 rounded-full"></span>{:else}<span class="w-1.5 h-1.5 bg-red-400 rounded-full"></span>{/if}
		{#if participant.hasVideo}<span class="w-1.5 h-1.5 bg-green-400 rounded-full"></span>{:else}<span class="w-1.5 h-1.5 bg-gray-500 rounded-full"></span>{/if}
	</div>
	{#if canControl && !participant.isLocal}
		<button onclick={() => onMuteAudio?.(participant.id)} class="text-[10px] px-1.5 py-0.5 rounded hover:bg-gray-700 text-gray-400" title="بی‌صدا کردن">🔇</button>
		<button onclick={() => onMuteVideo?.(participant.id)} class="text-[10px] px-1.5 py-0.5 rounded hover:bg-gray-700 text-gray-400" title="خاموش کردن ویدیو">📷</button>
		{#if isOwner}
			<button onclick={() => onKick?.(participant.id)} class="text-[10px] px-1.5 py-0.5 rounded hover:bg-red-600 text-gray-400 hover:text-white" title="حذف">✕</button>
		{/if}
	{/if}
</div>
