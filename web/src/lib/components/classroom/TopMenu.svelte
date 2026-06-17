<script lang="ts">
	import type { UserRole } from '$lib/classroom/types';
	import { ROLE_PERMISSIONS } from '$lib/classroom/types';

	let {
		userRole = 'student',
		audioOn = $bindable(true),
		micOn = $bindable(false),
		webcamOn = $bindable(false),
		screenShareOn = $bindable(false),
		handRaised = $bindable(false),
		showChat = $bindable(true),
		onToggleMic,
		onToggleWebcam,
		onToggleScreenShare,
		onToggleHand,
		onLeave,
	}: {
		userRole: UserRole;
		audioOn: boolean;
		micOn: boolean;
		webcamOn: boolean;
		screenShareOn: boolean;
		handRaised: boolean;
		showChat: boolean;
		onToggleMic: () => void;
		onToggleWebcam: () => void;
		onToggleScreenShare: () => void;
		onToggleHand: () => void;
		onLeave: () => void;
	} = $props();

	const perms = $derived(ROLE_PERMISSIONS[userRole] || ROLE_PERMISSIONS.student);
</script>

<header class="flex items-center justify-between px-3 h-11 shrink-0" style="background-color: #1e1e3a; border-bottom: 1px solid #2a2a4a;">
	<!-- Left: Audio output -->
	<div class="flex items-center gap-1.5">
		<button onclick={() => audioOn = !audioOn}
			class="w-8 h-8 rounded-lg flex items-center justify-center transition-colors {audioOn ? 'bg-[#23b9d7]/20 text-[#23b9d7]' : 'bg-[#e05252]/20 text-[#e05252]'}"
			title={audioOn ? 'بی‌صدا' : 'صدا'}
		>
			{#if audioOn}
				<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.536 8.464a5 5 0 010 7.072m2.828-9.9a9 9 0 010 12.728M5.586 15H4a1 1 0 01-1-1v-4a1 1 0 011-1h1.586l4.707-4.707C10.923 3.663 12 4.109 12 5v14c0 .891-1.077 1.337-1.707.707L5.586 15z" /></svg>
			{:else}
				<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5.586 15H4a1 1 0 01-1-1v-4a1 1 0 011-1h1.586l4.707-4.707C10.923 3.663 12 4.109 12 5v14c0 .891-1.077 1.337-1.707.707L5.586 15z" /><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2" /></svg>
			{/if}
		</button>
	</div>

	<!-- Center: Main controls -->
	<div class="flex items-center gap-1">
		{#if perms.canMic}
			<button onclick={onToggleMic}
				class="w-9 h-9 rounded-lg flex items-center justify-center transition-colors {micOn ? 'bg-[#23b9d7]/20 text-[#23b9d7]' : 'bg-[#3a3a5a] text-[#94a3b8]'}"
				title={micOn ? 'بستن میکروفون' : 'باز کردن میکروفون'}
			>
				{#if micOn}
					<svg class="w-4.5 h-4.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11a7 7 0 01-7 7m0 0a7 7 0 01-7-7m7 7v4m0 0H8m4 0h4m-4-8a3 3 0 01-3-3V5a3 3 0 116 0v6a3 3 0 01-3 3z" /></svg>
				{:else}
					<svg class="w-4.5 h-4.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5.586 15H4a1 1 0 01-1-1v-4a1 1 0 011-1h1.586l4.707-4.707C10.923 3.663 12 4.109 12 5v14c0 .891-1.077 1.337-1.707.707L5.586 15z" /><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2" /></svg>
				{/if}
			</button>
		{/if}

		{#if perms.canWebcam}
			<button onclick={onToggleWebcam}
				class="w-9 h-9 rounded-lg flex items-center justify-center transition-colors {webcamOn ? 'bg-[#23b9d7]/20 text-[#23b9d7]' : 'bg-[#3a3a5a] text-[#94a3b8]'}"
				title={webcamOn ? 'بستن وبکم' : 'باز کردن وبکم'}
			>
				{#if webcamOn}
					<svg class="w-4.5 h-4.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" /></svg>
				{:else}
					<svg class="w-4.5 h-4.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" /><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2" /></svg>
				{/if}
			</button>
		{/if}

		{#if perms.canScreenShare}
			<button onclick={onToggleScreenShare}
				class="w-9 h-9 rounded-lg flex items-center justify-center transition-colors {screenShareOn ? 'bg-[#23b9d7]/20 text-[#23b9d7]' : 'bg-[#3a3a5a] text-[#94a3b8]'}"
				title="اشتراک‌گذاری صفحه"
			>
				<svg class="w-4.5 h-4.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" /></svg>
			</button>
		{/if}

		<div class="w-px h-5 bg-[#3a3a5a] mx-0.5"></div>

		{#if perms.canHandRaise}
			<button onclick={onToggleHand}
				class="w-9 h-9 rounded-lg flex items-center justify-center transition-colors {handRaised ? 'bg-[#f59e0b]/20 text-[#f59e0b]' : 'bg-[#3a3a5a] text-[#94a3b8]'}"
				title={handRaised ? 'پایین آوردن دست' : 'بالا بردن دست'}
			>
				<span class="text-base">{handRaised ? '✋' : '🖐️'}</span>
			</button>
		{/if}

		<button onclick={() => showChat = !showChat}
			class="w-9 h-9 rounded-lg flex items-center justify-center transition-colors {showChat ? 'bg-[#23b9d7]/20 text-[#23b9d7]' : 'bg-[#3a3a5a] text-[#94a3b8]'}"
			title="چت"
		>
			<svg class="w-4.5 h-4.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" /></svg>
		</button>
	</div>

	<!-- Right: Leave -->
	<div class="flex items-center gap-1.5">
		<button onclick={onLeave}
			class="flex items-center gap-1.5 px-3 py-1.5 bg-[#e05252] text-white rounded-lg text-xs font-medium hover:bg-[#c44040] transition-colors"
		>
			<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" /></svg>
			خروج
		</button>
	</div>
</header>
