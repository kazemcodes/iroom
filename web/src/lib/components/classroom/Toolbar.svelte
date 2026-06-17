<script lang="ts">
	let {
		audioEnabled = $bindable(true),
		videoEnabled = $bindable(false),
		screenSharing = $bindable(false),
		whiteboardOpen = $bindable(false),
		handRaised = $bindable(false),
		isRecording = $bindable(false),
		onToggleAudio,
		onToggleVideo,
		onToggleScreenShare,
		onToggleWhiteboard,
		onToggleHandRaise,
		onToggleRecording,
		onLeave,
	}: {
		audioEnabled: boolean;
		videoEnabled: boolean;
		screenSharing: boolean;
		whiteboardOpen: boolean;
		handRaised: boolean;
		isRecording: boolean;
		onToggleAudio: () => void;
		onToggleVideo: () => void;
		onToggleScreenShare: () => void;
		onToggleWhiteboard: () => void;
		onToggleHandRaise: () => void;
		onToggleRecording: () => void;
		onLeave: () => void;
	} = $props();
</script>

<div class="flex items-center justify-center gap-2.5 px-4 py-3 shrink-0" style="background-color: #16213e; border-top: 1px solid #2a2a4a;">
	<!-- Microphone -->
	<button onclick={onToggleAudio} class="w-10 h-10 rounded-full flex items-center justify-center transition-colors {audioEnabled ? 'bg-gray-700 hover:bg-gray-600' : 'bg-red-600 hover:bg-red-700'}" title={audioEnabled ? 'بی‌صدا' : 'صدا'}>
		{#if audioEnabled}
			<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M19 11a7 7 0 01-7 7m0 0a7 7 0 01-7-7m7 7v4m0 0H8m4 0h4m-4-8a3 3 0 01-3-3V5a3 3 0 116 0v6a3 3 0 01-3 3z" /></svg>
		{:else}
			<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M5.586 15H4a1 1 0 01-1-1v-4a1 1 0 011-1h1.586l4.707-4.707C10.923 3.663 12 4.109 12 5v14c0 .891-1.077 1.337-1.707.707L5.586 15z" /><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M17 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2" /></svg>
		{/if}
	</button>

	<!-- Camera -->
	<button onclick={onToggleVideo} class="w-10 h-10 rounded-full flex items-center justify-center transition-colors {videoEnabled ? 'bg-gray-700 hover:bg-gray-600' : 'bg-red-600 hover:bg-red-700'}" title={videoEnabled ? 'ویدیو خاموش' : 'روشن کردن ویدیو'}>
		{#if videoEnabled}
			<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" /></svg>
		{:else}
			<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" /><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M17 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2" /></svg>
		{/if}
	</button>

	<!-- Screen Share -->
	<button onclick={onToggleScreenShare} class="w-10 h-10 rounded-full flex items-center justify-center transition-colors {screenSharing ? 'bg-blue-600 hover:bg-blue-700' : 'bg-gray-700 hover:bg-gray-600'}" title="اشتراک‌گذاری صفحه">
		<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" /></svg>
	</button>

	<!-- Whiteboard -->
	<button onclick={onToggleWhiteboard} class="w-10 h-10 rounded-full flex items-center justify-center transition-colors {whiteboardOpen ? 'bg-purple-600 hover:bg-purple-700' : 'bg-gray-700 hover:bg-gray-600'}" title="تخته‌سفید">
		<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" /></svg>
	</button>

	<!-- Raise Hand -->
	<button onclick={onToggleHandRaise} class="w-10 h-10 rounded-full flex items-center justify-center transition-colors {handRaised ? 'bg-yellow-500 hover:bg-yellow-600' : 'bg-gray-700 hover:bg-gray-600'}" title={handRaised ? 'پایین آوردن دست' : 'بالا بردن دست'}>
		<span class="text-lg">✋</span>
	</button>

	<!-- Recording -->
	<button onclick={onToggleRecording} class="w-10 h-10 rounded-full flex items-center justify-center transition-colors {isRecording ? 'bg-red-600 hover:bg-red-700 animate-pulse' : 'bg-gray-700 hover:bg-gray-600'}" title={isRecording ? 'پایان ضبط' : 'ضبط'}>
		<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><circle cx="12" cy="12" r="10" /><circle cx="12" cy="12" r="4" fill="currentColor" /></svg>
	</button>

	<!-- Settings -->
	<button class="w-10 h-10 rounded-full flex items-center justify-center transition-colors bg-gray-700 hover:bg-gray-600" title="تنظیمات">
		<span class="text-lg">⚙️</span>
	</button>

	<div class="w-px h-6 mx-1" style="background-color: #2a2a4a;"></div>

	<!-- Leave -->
	<button onclick={onLeave} class="px-4 py-2 bg-red-600 text-white rounded-full text-xs font-medium hover:bg-red-700 transition-colors flex items-center gap-1.5">
		<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" /></svg>
		خروج
	</button>
</div>
