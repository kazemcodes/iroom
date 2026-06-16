<script lang="ts">
	import { page } from '$app/state';
	import { auth } from '$lib/stores';
	import { api } from '$lib/api';
	import { onMount, onDestroy } from 'svelte';
	import { Room, RoomEvent, Track, ConnectionState, type Room as RoomType, type TrackPublication } from 'livekit-client';
	import Whiteboard from '$lib/components/Whiteboard.svelte';
	import type { Session } from '$lib/types';

	let session = $state<Session | null>(null);
	let loading = $state(true);
	let room = $state<RoomType | null>(null);
	let connectionState = $state<ConnectionState>(ConnectionState.Disconnected);
	let audioEnabled = $state(true);
	let videoEnabled = $state(false);
	let screenSharing = $state(false);
	let chatOpen = $state(false);
	let whiteboardOpen = $state(false);
	let participantsOpen = $state(false);
	let participants = $state<{identity: string; name: string; isSpeaking: boolean; hasVideo: boolean; hasAudio: boolean}[]>([]);
	let chatMessages = $state<any[]>([]);
	let chatInput = $state('');
	let localVideoEl: HTMLVideoElement;
	let remoteContainer: HTMLDivElement;
	let chatWs: WebSocket | null = null;
	let mediaRecorder = $state<MediaRecorder | null>(null);
	let isRecording = $state(false);
	let recordingChunks = $state<Blob[]>([]);
	let recordingStartTime = $state(0);
	let connected = $state(false);

	const sessionId = $derived(page.params.id);

	onMount(async () => { await loadSession(); });
	onDestroy(() => { stopRecording(); disconnect(); });

	async function loadSession() {
		loading = true;
		const res = await api.get<Session>(`/sessions/${sessionId}`);
		if (res.success) session = res.data!;
		loading = false;
		connectChatWs();
	}

	function connectChatWs() {
		const token = localStorage.getItem('access_token');
		if (!token) return;
		const proto = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
		chatWs = new WebSocket(`${proto}//${window.location.host}/ws/sessions/${sessionId}?token=${token}`);
		chatWs.onmessage = (event) => {
			try {
				const data = JSON.parse(event.data);
				if (data.type === 'message') {
					const msg = data.message;
					const isOwn = msg.user_id === $auth.user?.id;
					chatMessages = [...chatMessages, {
						sender: isOwn ? 'شما' : (msg.user_display_name || 'کاربر'),
						content: msg.content,
						time: new Date(msg.created_at).toLocaleTimeString('fa-IR', { hour: '2-digit', minute: '2-digit' }),
						isOwn
					}];
				}
			} catch (e) {}
		};
		chatWs.onclose = () => { if (connected) setTimeout(connectChatWs, 3000); };
	}

	async function joinRoom() {
		const tokenRes = await api.get<{ token: string; url: string; room: string }>(`/sessions/${sessionId}/livekit-token`);
		if (!tokenRes.success || !tokenRes.data) { alert(tokenRes.error || 'خطا در دریافت توکن اتاق'); return; }

		const { token, url } = tokenRes.data;
		try {
			room = new Room({
				adaptiveStream: true,
				dynacast: true,
				audioCaptureDefaults: { echoCancellation: true, noiseSuppression: true },
				videoCaptureDefaults: { resolution: { width: 640, height: 480, maxFps: 30 } }
			});

			room.on(RoomEvent.Connected, () => {
				connectionState = ConnectionState.Connected;
				connected = true;
				const lp = room!.localParticipant;
				lp.setMicrophoneEnabled(true).catch(() => {});
			});

			room.on(RoomEvent.Disconnected, () => { connectionState = ConnectionState.Disconnected; });

			room.on(RoomEvent.TrackSubscribed, (track: TrackPublication, participant) => {
				console.log('Track subscribed:', track.source, participant.identity);
				try {
					if (track.source === Track.Source.Camera) {
						const el = track.track?.attach();
						if (el) {
							el.id = `track-${participant.identity}`;
							el.className = 'w-full h-full object-cover rounded-lg';
							remoteContainer?.appendChild(el);
						}
					} else if (track.source === Track.Source.Microphone) {
						const el = track.track?.attach();
						if (el) {
							el.className = 'hidden';
							document.body.appendChild(el);
						}
					}
				} catch (e) { console.error('Track attach error:', e); }
			});

			room.on(RoomEvent.TrackUnsubscribed, (track: TrackPublication) => {
				const el = document.getElementById(`track-${track?.sid || ''}`);
				if (el) el.remove();
			});

			room.on(RoomEvent.ActiveSpeakersChanged, (speakers) => {
				const remotePs = room!.participants.values().toArray().map(p => ({
					identity: p.identity, name: p.name || 'ناشناس',
					isSpeaking: speakers.some(s => s.identity === p.identity),
					hasVideo: p.getTrackPublication(Track.Source.Camera)?.isSubscribed ?? false,
					hasAudio: p.getTrackPublication(Track.Source.Microphone)?.isSubscribed ?? false,
				}));
				const localP = {
					identity: room!.localParticipant.identity,
					name: $auth.user?.display_name || 'شما',
					isSpeaking: speakers.some(s => s.identity === room!.localParticipant.identity),
					hasVideo: videoEnabled,
					hasAudio: audioEnabled,
					isLocal: true,
				};
				participants = [localP, ...remotePs];
			});

			room.on(RoomEvent.DataReceived, (payload, participant) => {
				try {
					const data = JSON.parse(new TextDecoder().decode(payload));
					if (data.type === 'chat') {
						chatMessages = [...chatMessages, {
							sender: participant?.name || 'ناشناس',
							content: data.content,
							time: new Date().toLocaleTimeString('fa-IR', { hour: '2-digit', minute: '2-digit' })
						}];
					}
				} catch (e) {}
			});

			await room.connect(url, token);
		} catch (err) {
			console.error('Failed to connect:', err);
			connectionState = ConnectionState.Disconnected;
		}
	}

	function attachTrack(track: TrackPublication, identity: string) {
		if (!remoteContainer) return;
		const el = track.track?.attach();
		if (el) {
			el.id = `track-${identity}-${track.source}`;
			el.className = 'w-full h-full object-cover rounded-lg';
			remoteContainer.appendChild(el);
		}
	}

	function detachTrack(track: TrackPublication) {
		const el = document.getElementById(`track-${track?.sid || ''}`);
		if (el) el.remove();
	}

	function attachLocalVideo(ms: MediaStream) {
		if (localVideoEl) localVideoEl.srcObject = ms;
	}

	async function toggleAudio() {
		if (!room) return;
		const e = !audioEnabled;
		await room.localParticipant.setMicrophoneEnabled(e);
		audioEnabled = e;
	}

	async function toggleVideo() {
		if (!room) return;
		const e = !videoEnabled;
		videoEnabled = e;
		await room.localParticipant.setCameraEnabled(e);
		if (e) {
			try {
				const stream = await navigator.mediaDevices.getUserMedia({ video: true, width: 640, height: 480 });
				if (localVideoEl) localVideoEl.srcObject = stream;
			} catch (err) {
				console.error('Camera access denied:', err);
				videoEnabled = false;
				await room.localParticipant.setCameraEnabled(false);
			}
		} else {
			if (localVideoEl) {
				const stream = localVideoEl.srcObject as MediaStream;
				stream?.getTracks().forEach(t => t.stop());
				localVideoEl.srcObject = null;
			}
		}
	}

	async function toggleScreenShare() {
		if (!room) return;
		if (screenSharing) {
			await room.localParticipant.setScreenShareEnabled(false);
			screenSharing = false;
		} else {
			try {
				await room.localParticipant.setScreenShareEnabled(true);
				screenSharing = true;
			} catch (e) { console.error('Screen share failed:', e); }
		}
	}

	function startRecording() {
		if (!room) return;
		const streams: MediaStream[] = [];
		const lp = room.localParticipant;

		const micPub = lp.getTrackPublication(Track.Source.Microphone);
		if (micPub?.track?.mediaStreamTrack) streams.push(micPub.track.mediaStreamTrack);

		const camPub = lp.getTrackPublication(Track.Source.Camera);
		if (camPub?.track?.mediaStreamTrack) streams.push(camPub.track.mediaStreamTrack);

		if (streams.length === 0) { alert('ابتدا صدا یا ویدیو را فعال کنید'); return; }

		const combined = new MediaStream(streams);
		const options: MediaRecorderOptions = { mimeType: 'video/webm;codecs=vp9' };

		try {
			mediaRecorder = new MediaRecorder(combined, options);
		} catch {
			mediaRecorder = new MediaRecorder(combined);
		}

		recordingChunks = [];
		mediaRecorder.ondataavailable = (e) => { if (e.data.size > 0) recordingChunks.push(e.data); };
		mediaRecorder.onstop = async () => {
			const blob = new Blob(recordingChunks, { type: 'video/webm' });
			const duration = Math.floor((Date.now() - recordingStartTime) / 1000);
			const formData = new FormData();
			formData.append('file', blob, `recording-${Date.now()}.webm`);
			formData.append('duration', String(duration));

			const token = localStorage.getItem('access_token');
			await fetch(`${api.getBaseUrl()}/sessions/${sessionId}/recordings`, {
				method: 'POST',
				headers: { 'Authorization': `Bearer ${token}` },
				body: formData
			});
		};

		mediaRecorder.start(1000);
		isRecording = true;
		recordingStartTime = Date.now();
	}

	function stopRecording() {
		if (mediaRecorder && mediaRecorder.state !== 'inactive') {
			mediaRecorder.stop();
		}
		isRecording = false;
	}

	function sendChat() {
		if (!chatInput.trim()) return;
		const content = chatInput.trim();
		chatInput = '';

		if (chatWs?.readyState === WebSocket.OPEN) {
			chatWs.send(JSON.stringify({ type: 'message', content }));
		} else {
			api.post(`/sessions/${sessionId}/messages`, { content });
			chatMessages = [...chatMessages, {
				sender: 'شما', content, isOwn: true,
				time: new Date().toLocaleTimeString('fa-IR', { hour: '2-digit', minute: '2-digit' })
			}];
		}
	}

	function disconnect() {
		connected = false;
		chatWs?.close();
		stopRecording();
		if (localVideoEl?.srcObject) {
			(localVideoEl.srcObject as MediaStream).getTracks().forEach(t => t.stop());
			localVideoEl.srcObject = null;
		}
		if (room) { room.disconnect(); room = null; }
		connectionState = ConnectionState.Disconnected;
	}

	function leaveRoom() { disconnect(); history.back(); }
</script>

<div class="h-screen flex flex-col bg-gray-900 text-white">
	{#if loading}
		<div class="flex-1 flex items-center justify-center">
			<div class="animate-spin h-8 w-8 border-4 border-blue-500 border-t-transparent rounded-full"></div>
		</div>
	{:else if session}
		<!-- Top bar -->
		<div class="flex items-center justify-between px-4 py-2 bg-gray-800 border-b border-gray-700 shrink-0">
			<div class="flex items-center gap-3">
				<a href="/classes" class="text-gray-400 hover:text-white"><svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 5l7 7-7 7" /></svg></a>
				<h1 class="font-bold">{session.title}</h1>
				{#if connectionState === ConnectionState.Connected}
					<span class="flex items-center gap-1.5 text-xs text-green-400"><span class="w-2 h-2 bg-green-400 rounded-full animate-pulse"></span>متصل</span>
				{:else}
					<span class="text-xs text-gray-400">قطع شده</span>
				{/if}
				{#if isRecording}
					<span class="flex items-center gap-1.5 text-xs text-red-400"><span class="w-2 h-2 bg-red-500 rounded-full animate-pulse"></span>در حال ضبط</span>
				{/if}
			</div>
			<span class="text-sm text-gray-400">{session.duration} دقیقه</span>
		</div>

		<!-- Main content -->
		<div class="flex-1 flex overflow-hidden">
			<!-- Video / Whiteboard area -->
			<div class="flex-1 relative">
				{#if whiteboardOpen && room}
					<Whiteboard {room} {sessionId} />
				{:else}
					<div bind:this={remoteContainer} class="absolute inset-0 grid grid-cols-2 gap-2 p-2 auto-rows-fr"></div>
					{#if connectionState !== ConnectionState.Connected}
						<div class="absolute inset-0 flex flex-col items-center justify-center">
							<div class="w-24 h-24 bg-gray-700 rounded-full flex items-center justify-center mb-4">
								<span class="text-3xl font-bold text-gray-400">{$auth.user?.display_name?.charAt(0) || '?'}</span>
							</div>
							<p class="text-gray-400 mb-4">آماده پیوستن به کلاس</p>
							<button onclick={joinRoom} class="px-6 py-3 bg-blue-600 text-white rounded-xl font-medium hover:bg-blue-700 transition-colors">پیوستن به کلاس</button>
						</div>
					{/if}
					<div class="absolute bottom-4 left-4 w-40 h-30 rounded-lg overflow-hidden border-2 border-gray-600 bg-gray-800 {videoEnabled ? '' : 'hidden'}">
						<video bind:this={localVideoEl} autoplay muted playsinline class="w-full h-full object-cover"></video>
					</div>
				{/if}
			</div>

			<!-- Chat sidebar -->
			{#if chatOpen}
				<div class="w-80 bg-gray-800 border-r border-gray-700 flex flex-col shrink-0">
					<div class="px-4 py-3 border-b border-gray-700 flex items-center justify-between">
						<h3 class="font-bold text-sm">گفتگو</h3>
						<button onclick={() => chatOpen = false} class="text-gray-400 hover:text-white" aria-label="بستن"><svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" /></svg></button>
					</div>
					<div class="flex-1 overflow-y-auto px-4 py-3 space-y-3">
						{#each chatMessages as msg}
							<div>
								<div class="flex items-center gap-2"><span class="text-xs font-bold text-blue-400">{msg.sender}</span><span class="text-[10px] text-gray-500">{msg.time}</span></div>
								<p class="text-sm mt-0.5">{msg.content}</p>
							</div>
						{/each}
					</div>
					<div class="px-3 py-3 border-t border-gray-700">
						<form onsubmit={(e) => { e.preventDefault(); sendChat(); }} class="flex gap-2">
							<input type="text" bind:value={chatInput} class="flex-1 px-3 py-2 bg-gray-700 rounded-lg text-sm focus:ring-1 focus:ring-blue-500 outline-none" placeholder="پیام..." />
							<button type="submit" class="px-3 py-2 bg-blue-600 rounded-lg text-sm hover:bg-blue-700">ارسال</button>
						</form>
					</div>
				</div>
			{/if}

			<!-- Participants sidebar -->
			{#if participantsOpen}
				<div class="w-72 bg-gray-800 border-r border-gray-700 flex flex-col shrink-0">
					<div class="px-4 py-3 border-b border-gray-700 flex items-center justify-between">
						<h3 class="font-bold text-sm">شرکت‌کنندگان ({participants.length})</h3>
						<button onclick={() => participantsOpen = false} class="text-gray-400 hover:text-white" aria-label="بستن"><svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" /></svg></button>
					</div>
					<div class="flex-1 overflow-y-auto px-3 py-3 space-y-1">
						{#each participants as p}
							<div class="flex items-center gap-3 px-3 py-2 rounded-lg {p.isSpeaking ? 'bg-blue-900/30' : ''}">
								<div class="relative">
									<div class="w-8 h-8 rounded-full bg-gray-600 flex items-center justify-center text-xs font-bold">
										{p.name.charAt(0)}
									</div>
									{#if p.isSpeaking}
										<span class="absolute -bottom-0.5 -right-0.5 w-3 h-3 bg-green-500 rounded-full border-2 border-gray-800"></span>
									{/if}
								</div>
								<div class="flex-1 min-w-0">
									<p class="text-sm font-medium truncate">{p.name}{'isLocal' in p && p.isLocal ? ' (شما)' : ''}</p>
								</div>
								<div class="flex items-center gap-1">
									{#if p.hasAudio}<span class="w-2 h-2 bg-green-400 rounded-full"></span>{:else}<span class="w-2 h-2 bg-red-400 rounded-full"></span>{/if}
									{#if p.hasVideo}<span class="w-2 h-2 bg-green-400 rounded-full"></span>{:else}<span class="w-2 h-2 bg-gray-500 rounded-full"></span>{/if}
								</div>
							</div>
						{/each}
						{#if participants.length === 0}
							<p class="text-center text-gray-500 text-sm py-4">هنوز کسی متصل نیست</p>
						{/if}
					</div>
				</div>
			{/if}
		</div>

		<!-- Controls -->
		<div class="flex items-center justify-center gap-3 px-4 py-4 bg-gray-800 border-t border-gray-700 shrink-0">
			<button onclick={toggleAudio} class="w-12 h-12 rounded-full flex items-center justify-center transition-colors {audioEnabled ? 'bg-gray-700 hover:bg-gray-600' : 'bg-red-600 hover:bg-red-700'}" title={audioEnabled ? 'بی‌صدا' : 'صدا'}>
				{#if audioEnabled}<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M19 11a7 7 0 01-7 7m0 0a7 7 0 01-7-7m7 7v4m0 0H8m4 0h4m-4-8a3 3 0 01-3-3V5a3 3 0 116 0v6a3 3 0 01-3 3z" /></svg>{:else}<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M5.586 15H4a1 1 0 01-1-1v-4a1 1 0 011-1h1.586l4.707-4.707C10.923 3.663 12 4.109 12 5v14c0 .891-1.077 1.337-1.707.707L5.586 15z" /><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M17 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2" /></svg>{/if}
			</button>
			<button onclick={toggleVideo} class="w-12 h-12 rounded-full flex items-center justify-center transition-colors {videoEnabled ? 'bg-gray-700 hover:bg-gray-600' : 'bg-red-600 hover:bg-red-700'}" title={videoEnabled ? 'ویدیو خاموش' : 'روشن کردن ویدیو'}>
				{#if videoEnabled}<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" /></svg>{:else}<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" /><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M17 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2" /></svg>{/if}
			</button>
			<button onclick={toggleScreenShare} class="w-12 h-12 rounded-full flex items-center justify-center transition-colors {screenSharing ? 'bg-blue-600 hover:bg-blue-700' : 'bg-gray-700 hover:bg-gray-600'}" title="اشتراک‌گذاری صفحه">
				<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" /></svg>
			</button>

			<!-- Whiteboard toggle -->
			<button onclick={() => whiteboardOpen = !whiteboardOpen} class="w-12 h-12 rounded-full flex items-center justify-center transition-colors {whiteboardOpen ? 'bg-purple-600 hover:bg-purple-700' : 'bg-gray-700 hover:bg-gray-600'}" title="تخته‌سفید">
				<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" /></svg>
			</button>

			<button onclick={() => chatOpen = !chatOpen} class="w-12 h-12 rounded-full flex items-center justify-center transition-colors {chatOpen ? 'bg-blue-600 hover:bg-blue-700' : 'bg-gray-700 hover:bg-gray-600'}" title="گفتگو">
				<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" /></svg>
			</button>

			<!-- Participants -->
			<button onclick={() => participantsOpen = !participantsOpen} class="w-12 h-12 rounded-full flex items-center justify-center transition-colors {participantsOpen ? 'bg-blue-600 hover:bg-blue-700' : 'bg-gray-700 hover:bg-gray-600'}" title="شرکت‌کنندگان">
				<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z" /></svg>
			</button>

			<!-- Recording toggle -->
			<button onclick={isRecording ? stopRecording : startRecording} class="w-12 h-12 rounded-full flex items-center justify-center transition-colors {isRecording ? 'bg-red-600 hover:bg-red-700 animate-pulse' : 'bg-gray-700 hover:bg-gray-600'}" title={isRecording ? 'پایان ضبط' : 'ضبط'}>
				<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><circle cx="12" cy="12" r="10" /><circle cx="12" cy="12" r="4" fill="currentColor" /></svg>
			</button>

			<div class="w-px h-8 bg-gray-600 mx-2"></div>

			<button onclick={leaveRoom} class="px-6 py-2.5 bg-red-600 text-white rounded-full text-sm font-medium hover:bg-red-700 transition-colors flex items-center gap-2">
				<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" /></svg>
				خروج
			</button>
		</div>
	{:else}
		<div class="flex-1 flex items-center justify-center"><p class="text-gray-400">جلسه یافت نشد</p></div>
	{/if}
</div>
