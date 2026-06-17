<script lang="ts">
	let {
		audioOn = $bindable(true),
		micOn = $bindable(true),
		webcamOn = $bindable(false),
		screenShareOn = $bindable(false),
		whiteboardOn = $bindable(false),
		filesOn = $bindable(false),
		handRaised = $bindable(false),
		isRecording = $bindable(false),
		onToggleAudio,
		onToggleMic,
		onToggleWebcam,
		onToggleScreenShare,
		onToggleWhiteboard,
		onToggleFiles,
		onToggleHand,
		onLeave,
	}: {
		audioOn: boolean;
		micOn: boolean;
		webcamOn: boolean;
		screenShareOn: boolean;
		whiteboardOn: boolean;
		filesOn: boolean;
		handRaised: boolean;
		isRecording: boolean;
		onToggleAudio: () => void;
		onToggleMic: () => void;
		onToggleWebcam: () => void;
		onToggleScreenShare: () => void;
		onToggleWhiteboard: () => void;
		onToggleFiles: () => void;
		onToggleHand: () => void;
		onLeave: () => void;
	} = $props();
</script>

<div class="absolute bottom-4 left-1/2 -translate-x-1/2 flex items-center gap-2 px-3 py-2 rounded-full z-30" style="background-color: rgba(37, 37, 64, 0.95); backdrop-filter: blur(8px);">
	<!-- Audio Output -->
	<button onclick={onToggleAudio} class="w-10 h-10 rounded-full flex items-center justify-center transition-all duration-150 {audioOn ? 'bg-[#3a3a5a] border border-[#23b9d7]' : 'bg-[#252540] border border-[#3a3a5a]'}" title="خروجی صدا">
		{#if audioOn}
			<svg class="w-5 h-5 text-[#23b9d7]" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M15.536 8.464a5 5 0 010 7.072m2.828-9.9a9 9 0 010 12.728M5.586 15H4a1 1 0 01-1-1v-4a1 1 0 011-1h1.586l4.707-4.707C10.923 3.663 12 4.109 12 5v14c0 .891-1.077 1.337-1.707.707L5.586 15z" /></svg>
		{:else}
			<svg class="w-5 h-5 text-[#94a3b8]" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M5.586 15H4a1 1 0 01-1-1v-4a1 1 0 011-1h1.586l4.707-4.707C10.923 3.663 12 4.109 12 5v14c0 .891-1.077 1.337-1.707.707L5.586 15z" /><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M17 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2" /></svg>
		{/if}
	</button>

	<!-- Microphone -->
	<button onclick={onToggleMic} class="w-10 h-10 rounded-full flex items-center justify-center transition-all duration-150 {micOn ? 'bg-[#3a3a5a] border border-[#23b9d7]' : 'bg-[#252540] border border-[#3a3a5a]'}" title="میکروفون">
		{#if micOn}
			<svg class="w-5 h-5 text-[#23b9d7]" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M19 11a7 7 0 01-7 7m0 0a7 7 0 01-7-7m7 7v4m0 0H8m4 0h4m-4-8a3 3 0 01-3-3V5a3 3 0 116 0v6a3 3 0 01-3 3z" /></svg>
		{:else}
			<svg class="w-5 h-5 text-[#94a3b8]" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M5.586 15H4a1 1 0 01-1-1v-4a1 1 0 011-1h1.586l4.707-4.707C10.923 3.663 12 4.109 12 5v14c0 .891-1.077 1.337-1.707.707L5.586 15z" /><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M17 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2" /></svg>
		{/if}
	</button>

	<!-- Webcam -->
	<button onclick={onToggleWebcam} class="w-10 h-10 rounded-full flex items-center justify-center transition-all duration-150 {webcamOn ? 'bg-[#3a3a5a] border border-[#23b9d7]' : 'bg-[#252540] border border-[#3a3a5a]'}" title="وبکم">
		{#if webcamOn}
			<svg class="w-5 h-5 text-[#23b9d7]" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" /></svg>
		{:else}
			<svg class="w-5 h-5 text-[#94a3b8]" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" /><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M17 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2" /></svg>
		{/if}
	</button>

	<!-- Screen Share -->
	<button onclick={onToggleScreenShare} class="w-10 h-10 rounded-full flex items-center justify-center transition-all duration-150 {screenShareOn ? 'bg-[#3a3a5a] border border-[#23b9d7]' : 'bg-[#252540] border border-[#3a3a5a]'}" title="اشتراک‌گذاری صفحه">
		<svg class="w-5 h-5 {screenShareOn ? 'text-[#23b9d7]' : 'text-[#94a3b8]'}" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" /></svg>
	</button>

	<!-- Whiteboard -->
	<button onclick={onToggleWhiteboard} class="w-10 h-10 rounded-full flex items-center justify-center transition-all duration-150 {whiteboardOn ? 'bg-[#3a3a5a] border border-[#23b9d7]' : 'bg-[#252540] border border-[#3a3a5a]'}" title="تخته">
		<svg class="w-5 h-5 {whiteboardOn ? 'text-[#23b9d7]' : 'text-[#94a3b8]'}" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" /></svg>
	</button>

	<!-- Files -->
	<button onclick={onToggleFiles} class="w-10 h-10 rounded-full flex items-center justify-center transition-all duration-150 {filesOn ? 'bg-[#3a3a5a] border border-[#23b9d7]' : 'bg-[#252540] border border-[#3a3a5a]'}" title="فایل‌ها">
		<svg class="w-5 h-5 {filesOn ? 'text-[#23b9d7]' : 'text-[#94a3b8]'}" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M7 4v16M17 4v16M3 8h4m10 0h4M3 12h18M3 16h4m10 0h4M4 20h16a1 1 0 001-1V5a1 1 0 00-1-1H4a1 1 0 00-1 1v14a1 1 0 001 1z" /></svg>
	</button>

	<!-- Hand Raise -->
	<button onclick={onToggleHand} class="w-10 h-10 rounded-full flex items-center justify-center transition-all duration-150 {handRaised ? 'bg-[#3a3a5a] border border-[#23b9d7]' : 'bg-[#252540] border border-[#3a3a5a]'}" title="بالا بردن دست">
		<span class="text-lg {handRaised ? '' : 'grayscale'}">{handRaised ? '✋' : '🖐️'}</span>
	</button>

	<div class="w-px h-6 bg-[#3a3a5a] mx-1"></div>

	<!-- Leave -->
	<button onclick={onLeave} class="px-4 py-2 bg-[#e05252] text-white rounded-full text-xs font-medium hover:bg-[#c44040] transition-colors flex items-center gap-1.5">
		<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" /></svg>
		خروج
	</button>
</div>
