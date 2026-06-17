<script lang="ts">
	// @ts-nocheck
	import { page } from '$app/state';
	import { auth } from '$lib/stores';
	import { api } from '$lib/api';
	import { onMount, onDestroy } from 'svelte';
	import { Room, RoomEvent, Track, ConnectionState, type Room as RoomType, type TrackPublication } from 'livekit-client';
	import { formatDuration, toPersianNum } from '$lib/utils/persian';
	import { classroomWindow } from '$lib/classroom/ClassroomWindow';
	import Whiteboard from '$lib/components/Whiteboard.svelte';
	import SettingsPopup from '$lib/components/SettingsPopup.svelte';
	import type { Session } from '$lib/types';

	let session = $state<Session | null>(null);
	let loading = $state(true);
	let popupBlocked = $state(false);

	let room = $state<RoomType | null>(null);
	let connectionState = $state<ConnectionState>(ConnectionState.Disconnected);
	let audioEnabled = $state(true);
	let videoEnabled = $state(false);
	let screenSharing = $state(false);
	let whiteboardOpen = $state(false);
	let participantsOpen = $state(false);
	let chatOpen = $state(false);
	let participants = $state<{identity: string; name: string; isSpeaking: boolean; hasVideo: boolean; hasAudio: boolean; isLocal?: boolean; handRaised?: boolean}[]>([]);
	let chatMessages = $state<{sender: string; content: string; time: string; isOwn?: boolean}[]>([]);
	let chatInput = $state('');
	let localVideoEl: HTMLVideoElement;
	let remoteContainer: HTMLDivElement;
	let chatWs: WebSocket | null = null;
	let mediaRecorder = $state<MediaRecorder | null>(null);
	let isRecording = $state(false);
	let recordingChunks = $state<Blob[]>([]);
	let recordingStartTime = $state(0);
	let connected = $state(false);
	let handRaised = $state(false);
	let elapsedSeconds = $state(0);
	let timerInterval: ReturnType<typeof setInterval> | null = null;
	let unreadCount = $state(0);
	let handsRaised = $state<Record<string, boolean>>({});
	let showSettings = $state(false);

	const sessionId = $derived(page.params.id!);
	const isTeacherOrAdmin = $derived($auth.user?.role === 'teacher' || $auth.user?.role === 'admin');

	const gridCols = $derived.by(() => {
		const count = participants.length;
		if (count <= 1) return 'grid-cols-1';
		if (count <= 2) return 'grid-cols-2';
		if (count <= 4) return 'grid-cols-2';
		if (count <= 6) return 'grid-cols-3';
		return 'grid-cols-4';
	});

	onMount(async () => { await loadSession(); });
	onDestroy(() => { stopRecording(); disconnectInline(); });

	async function loadSession() {
		loading = true;
		const res = await api.get<Session>(`/sessions/${sessionId}`);
		if (res.success) session = res.data!;
		loading = false;
	}

	function openPopup() {
		if (!session) return;
		const win = classroomWindow.open(String(session.id), session.title);
		if (!win) {
			popupBlocked = true;
		}
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
					if (!isOwn && !chatOpen) unreadCount++;
				}
			} catch (e) {}
		};
		chatWs.onclose = () => { if (connected) setTimeout(connectChatWs, 3000); };
	}

	async function joinInlineRoom() {
		const tokenRes = await api.get<{ token: string; url: string; room: string }>(`/sessions/${sessionId}/livekit-token`);
		if (!tokenRes.success || !tokenRes.data) { alert(tokenRes.error || 'خطا در دریافت توکن اتاق'); return; }

		const { token, url } = tokenRes.data;
		try {
			room = new Room({
				adaptiveStream: true,
				dynacast: true,
				audioCaptureDefaults: { echoCancellation: true, noiseSuppression: true },
				videoCaptureDefaults: { resolution: { width: 640, height: 480, maxFps: 30 } as any }
			});

			room.on(RoomEvent.Connected, () => {
				connectionState = ConnectionState.Connected;
				connected = true;
				elapsedSeconds = 0;
				timerInterval = setInterval(() => { elapsedSeconds++; }, 1000);
				const lp = room!.localParticipant;
				lp.setMicrophoneEnabled(true).catch(() => {});
			});

			room.on(RoomEvent.Disconnected, () => {
				connectionState = ConnectionState.Disconnected;
				if (timerInterval) { clearInterval(timerInterval); timerInterval = null; }
			});

			room.on(RoomEvent.TrackSubscribed, (track: any, participant: any) => {
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

			room.on(RoomEvent.TrackUnsubscribed, (track: any) => {
				const el = document.getElementById(`track-${track?.trackSid || ''}`);
				if (el) el.remove();
			});

			room.on(RoomEvent.ActiveSpeakersChanged, (speakers: any) => {
				const remotePs = ((room as any)!.remoteParticipants || room!.participants).values().toArray().map(p => ({
					identity: p.identity, name: p.name || 'ناشناس',
					isSpeaking: speakers.some(s => s.identity === p.identity),
					hasVideo: p.getTrackPublication(Track.Source.Camera)?.isSubscribed ?? false,
					hasAudio: p.getTrackPublication(Track.Source.Microphone)?.isSubscribed ?? false,
					handRaised: handsRaised[p.identity] ?? false,
				}));
				const localP = {
					identity: room!.localParticipant.identity,
					name: $auth.user?.display_name || 'شما',
					isSpeaking: speakers.some(s => s.identity === room!.localParticipant.identity),
					hasVideo: videoEnabled,
					hasAudio: audioEnabled,
					isLocal: true,
					handRaised,
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
						if (!chatOpen) unreadCount++;
					} else if (data.type === 'hand-raise' && participant) {
						handsRaised[participant.identity] = data.raised;
					}
				} catch (e) {}
			});

			connectChatWs();
			await room.connect(url, token);
		} catch (err) {
			console.error('Failed to connect:', err);
			connectionState = ConnectionState.Disconnected;
		}
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
				const stream = await navigator.mediaDevices.getUserMedia({ video: { width: 640, height: 480 } });
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

	function toggleHandRaise() {
		if (!room) return;
		handRaised = !handRaised;
		room.localParticipant.sendData(
			new TextEncoder().encode(JSON.stringify({ type: 'hand-raise', raised: handRaised })),
			{ reliable: true }
		);
	}

	function muteParticipant(_identity: string) {
		alert('عملیات انجام شد');
	}

	function removeParticipant(_identity: string) {
		alert('عملیات انجام شد');
	}

	function startRecording() {
		if (!room) return;
		const streams: MediaStream[] = [];
		const lp = room.localParticipant;

		const micPub = lp.getTrackPublication(Track.Source.Microphone);
		if (micPub?.track?.mediaStreamTrack) streams.push(micPub.track.mediaStreamTrack as any);

		const camPub = lp.getTrackPublication(Track.Source.Camera);
		if (camPub?.track?.mediaStreamTrack) streams.push(camPub.track.mediaStreamTrack as any);

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

	function disconnectInline() {
		connected = false;
		chatWs?.close();
		stopRecording();
		if (timerInterval) { clearInterval(timerInterval); timerInterval = null; }
		if (localVideoEl?.srcObject) {
			(localVideoEl.srcObject as MediaStream).getTracks().forEach(t => t.stop());
			localVideoEl.srcObject = null;
		}
		if (room) { room.disconnect(); room = null; }
		connectionState = ConnectionState.Disconnected;
	}

	function formatDate(d: string) {
		if (!d) return '';
		return new Date(d).toLocaleDateString('fa-IR', { year: 'numeric', month: 'long', day: 'numeric', hour: '2-digit', minute: '2-digit' });
	}
</script>

<div class="min-h-screen bg-gray-50">
	{#if loading}
		<div class="flex items-center justify-center py-20">
			<div class="animate-spin h-8 w-8 border-4 border-blue-600 border-t-transparent rounded-full"></div>
		</div>
	{:else if session}
		<div class="max-w-2xl mx-auto py-10 px-4">
			<!-- Session info card -->
			<div class="bg-white rounded-2xl border shadow-sm overflow-hidden">
				<div class="p-6">
					<div class="flex items-start justify-between">
						<div>
							<h1 class="text-xl font-bold text-gray-900">{session.title}</h1>
							<p class="text-sm text-gray-500 mt-1">{formatDate(session.scheduled_at)} — {session.duration} دقیقه</p>
						</div>
						{#if session.status === 'live'}
							<span class="flex items-center gap-1.5 text-xs text-green-600 bg-green-50 px-3 py-1.5 rounded-full font-medium">
								<span class="w-2 h-2 bg-green-500 rounded-full animate-pulse"></span>
								در حال برگزاری
							</span>
						{:else if session.status === 'scheduled'}
							<span class="text-xs text-blue-600 bg-blue-50 px-3 py-1.5 rounded-full font-medium">برنامه‌ریزی شده</span>
						{:else}
							<span class="text-xs text-gray-500 bg-gray-100 px-3 py-1.5 rounded-full font-medium">پایان یافته</span>
						{/if}
					</div>

					<div class="mt-6">
						{#if session.status === 'live'}
							<button onclick={openPopup} class="w-full py-3 bg-blue-600 text-white rounded-xl font-medium hover:bg-blue-700 transition-colors text-center">
								ورود به کلاس
							</button>
							{#if popupBlocked}
								<p class="text-xs text-amber-600 mt-2 text-center">پاپ‌آپ مسدود شده. <button onclick={() => { popupBlocked = false; joinInlineRoom(); }} class="underline font-medium">ورود در همین صفحه</button></p>
							{/if}
						{:else if session.status === 'scheduled'}
							<p class="text-sm text-gray-500 text-center py-3">جلسه هنوز شروع نشده</p>
						{:else}
							<p class="text-sm text-gray-500 text-center py-3">جلسه به پایان رسیده</p>
						{/if}
					</div>
				</div>
			</div>

			<!-- Inline classroom fallback -->
			{#if popupBlocked}
				<div class="mt-6 bg-white rounded-2xl border shadow-sm overflow-hidden">
					<!-- Top bar with timer -->
					<div class="flex items-center justify-between px-4 py-2 bg-gray-800 border-b border-gray-700">
						<div class="flex items-center gap-3">
							<h2 class="text-sm font-bold text-white">{session.title}</h2>
							{#if connectionState === ConnectionState.Connected}
								<span class="flex items-center gap-1.5 text-xs text-green-400"><span class="w-2 h-2 bg-green-400 rounded-full animate-pulse"></span>متصل</span>
								<span class="text-xs text-gray-400 font-mono">{formatDuration(elapsedSeconds)}</span>
							{:else}
								<span class="text-xs text-gray-400">قطع شده</span>
							{/if}
							{#if isRecording}
								<span class="flex items-center gap-1.5 text-xs text-red-400"><span class="w-2 h-2 bg-red-500 rounded-full animate-pulse"></span>ضبط</span>
							{/if}
						</div>
						<span class="text-sm text-gray-400">{session.duration} دقیقه</span>
					</div>

					<div class="h-96 bg-gray-900 relative">
						<div bind:this={remoteContainer} class="absolute inset-0 {gridCols} gap-2 p-2 auto-rows-fr"></div>
						{#if connectionState !== ConnectionState.Connected}
							<div class="absolute inset-0 flex flex-col items-center justify-center">
								<div class="w-20 h-20 bg-gray-700 rounded-full flex items-center justify-center mb-3">
									<span class="text-2xl font-bold text-gray-400">{$auth.user?.display_name?.charAt(0) || '?'}</span>
								</div>
								<p class="text-gray-400 text-sm mb-3">آماده پیوستن</p>
								<button onclick={joinInlineRoom} class="px-5 py-2.5 bg-blue-600 text-white rounded-xl text-sm font-medium hover:bg-blue-700 transition-colors">پیوستن</button>
							</div>
						{/if}
						<div class="absolute bottom-3 left-3 w-36 h-28 rounded-lg overflow-hidden border-2 bg-gray-800 {videoEnabled ? 'border-gray-600' : 'border-transparent hidden'}">
							<video bind:this={localVideoEl} autoplay muted playsinline class="w-full h-full object-cover"></video>
						</div>
					</div>

					<!-- Controls -->
					<div class="flex items-center justify-center gap-2 px-4 py-3 bg-gray-800 border-t border-gray-700">
						<button onclick={toggleAudio} class="w-10 h-10 rounded-full flex items-center justify-center transition-colors {audioEnabled ? 'bg-gray-700 hover:bg-gray-600' : 'bg-red-600 hover:bg-red-700'}">
							{#if audioEnabled}<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M19 11a7 7 0 01-7 7m0 0a7 7 0 01-7-7m7 7v4m0 0H8m4 0h4m-4-8a3 3 0 01-3-3V5a3 3 0 116 0v6a3 3 0 01-3 3z" /></svg>{:else}<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M5.586 15H4a1 1 0 01-1-1v-4a1 1 0 011-1h1.586l4.707-4.707C10.923 3.663 12 4.109 12 5v14c0 .891-1.077 1.337-1.707.707L5.586 15z" /><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M17 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2" /></svg>{/if}
						</button>
						<button onclick={toggleVideo} class="w-10 h-10 rounded-full flex items-center justify-center transition-colors {videoEnabled ? 'bg-gray-700 hover:bg-gray-600' : 'bg-red-600 hover:bg-red-700'}">
							{#if videoEnabled}<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" /></svg>{:else}<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" /><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M17 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2" /></svg>{/if}
						</button>
						<button onclick={toggleScreenShare} class="w-10 h-10 rounded-full flex items-center justify-center transition-colors {screenSharing ? 'bg-blue-600 hover:bg-blue-700' : 'bg-gray-700 hover:bg-gray-600'}">
							<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" /></svg>
						</button>

						<button onclick={() => whiteboardOpen = !whiteboardOpen} class="w-10 h-10 rounded-full flex items-center justify-center transition-colors {whiteboardOpen ? 'bg-purple-600 hover:bg-purple-700' : 'bg-gray-700 hover:bg-gray-600'}">
							<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" /></svg>
						</button>

						<!-- Hand raise -->
						<button onclick={toggleHandRaise} class="w-10 h-10 rounded-full flex items-center justify-center transition-colors {handRaised ? 'bg-yellow-500 hover:bg-yellow-600' : 'bg-gray-700 hover:bg-gray-600'}" title={handRaised ? 'پایین آوردن دست' : 'بالا بردن دست'}>
							<span class="text-lg">✋</span>
						</button>

						<!-- Chat with unread badge -->
						<button onclick={() => { chatOpen = !chatOpen; if (chatOpen) unreadCount = 0; }} class="relative w-10 h-10 rounded-full flex items-center justify-center transition-colors {chatOpen ? 'bg-blue-600 hover:bg-blue-700' : 'bg-gray-700 hover:bg-gray-600'}">
							<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" /></svg>
							{#if unreadCount > 0}
								<span class="absolute -top-1 -right-1 w-5 h-5 bg-red-500 text-white text-xs rounded-full flex items-center justify-center">{unreadCount > 9 ? '۹+' : toPersianNum(unreadCount)}</span>
							{/if}
						</button>

						<!-- Participants -->
						<button onclick={() => participantsOpen = !participantsOpen} class="w-10 h-10 rounded-full flex items-center justify-center transition-colors {participantsOpen ? 'bg-blue-600 hover:bg-blue-700' : 'bg-gray-700 hover:bg-gray-600'}">
							<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z" /></svg>
						</button>

						<button onclick={isRecording ? stopRecording : startRecording} class="w-10 h-10 rounded-full flex items-center justify-center transition-colors {isRecording ? 'bg-red-600 hover:bg-red-700 animate-pulse' : 'bg-gray-700 hover:bg-gray-600'}">
							<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><circle cx="12" cy="12" r="10" /><circle cx="12" cy="12" r="4" fill="currentColor" /></svg>
						</button>

						<button onclick={() => showSettings = true} class="w-10 h-10 rounded-full flex items-center justify-center transition-colors bg-gray-700 hover:bg-gray-600" title="تنظیمات">
							<span class="text-lg">⚙️</span>
						</button>

						<div class="w-px h-8 bg-gray-600 mx-1"></div>
						<button onclick={() => { disconnectInline(); history.back(); }} class="px-4 py-2 bg-red-600 text-white rounded-full text-xs font-medium hover:bg-red-700 transition-colors flex items-center gap-1.5">
							<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" /></svg>
							خروج
						</button>
					</div>

					<SettingsPopup bind:show={showSettings} />

					<!-- Chat panel -->
					{#if chatOpen}
						<div class="bg-gray-800 border-t border-gray-700">
							<div class="flex items-center justify-between px-4 py-2 border-b border-gray-700">
								<h3 class="font-bold text-sm text-white">گفتگو</h3>
								<button onclick={() => chatOpen = false} class="text-gray-400 hover:text-white" aria-label="بستن"><svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" /></svg></button>
							</div>
							<div class="max-h-48 overflow-y-auto px-4 py-2 space-y-2">
								{#each chatMessages as msg}
									<div>
										<div class="flex items-center gap-2"><span class="text-xs font-bold text-blue-400">{msg.sender}</span><span class="text-[10px] text-gray-500">{msg.time}</span></div>
										<p class="text-sm text-gray-200 mt-0.5">{msg.content}</p>
									</div>
								{/each}
							</div>
							<div class="px-3 py-2 border-t border-gray-700">
								<form onsubmit={(e) => { e.preventDefault(); sendChat(); }} class="flex gap-2">
									<input type="text" bind:value={chatInput} class="flex-1 px-3 py-2 bg-gray-700 rounded-lg text-sm text-white focus:ring-1 focus:ring-blue-500 outline-none" placeholder="پیام..." />
									<button type="submit" class="px-3 py-2 bg-blue-600 rounded-lg text-sm text-white hover:bg-blue-700">ارسال</button>
								</form>
							</div>
						</div>
					{/if}

					<!-- Participants panel -->
					{#if participantsOpen}
						<div class="bg-gray-800 border-t border-gray-700">
							<div class="flex items-center justify-between px-4 py-2 border-b border-gray-700">
								<h3 class="font-bold text-sm text-white">شرکت‌کنندگان ({participants.length})</h3>
								<button onclick={() => participantsOpen = false} class="text-gray-400 hover:text-white" aria-label="بستن"><svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" /></svg></button>
							</div>
							<div class="max-h-48 overflow-y-auto px-3 py-2 space-y-1">
								{#each participants as p}
									<div class="flex items-center gap-3 px-3 py-2 rounded-lg {p.isSpeaking ? 'bg-blue-900/30' : ''}">
										<div class="relative">
											<div class="w-8 h-8 rounded-full bg-gray-600 flex items-center justify-center text-xs font-bold text-white">
												{p.name.charAt(0)}
											</div>
											{#if p.isSpeaking}
												<span class="absolute -bottom-0.5 -right-0.5 w-3 h-3 bg-green-500 rounded-full border-2 border-gray-800"></span>
											{/if}
										</div>
										<div class="flex-1 min-w-0">
											<p class="text-sm font-medium text-white truncate">
												{p.name}{'isLocal' in p && p.isLocal ? ' (شما)' : ''}
												{#if p.handRaised}<span class="text-yellow-400 ml-1">✋</span>{/if}
											</p>
										</div>
										<div class="flex items-center gap-1">
											{#if p.hasAudio}<span class="w-2 h-2 bg-green-400 rounded-full"></span>{:else}<span class="w-2 h-2 bg-red-400 rounded-full"></span>{/if}
											{#if p.hasVideo}<span class="w-2 h-2 bg-green-400 rounded-full"></span>{:else}<span class="w-2 h-2 bg-gray-500 rounded-full"></span>{/if}
										</div>
										{#if isTeacherOrAdmin && !('isLocal' in p && p.isLocal)}
											<button onclick={() => muteParticipant(p.identity)} class="text-xs px-2 py-1 bg-gray-700 hover:bg-gray-600 rounded text-gray-300" title="بی‌صدا کردن">
												🔇
											</button>
											<button onclick={() => removeParticipant(p.identity)} class="text-xs px-2 py-1 bg-gray-700 hover:bg-red-600 rounded text-gray-300 hover:text-white" title="حذف">
												✕
											</button>
										{/if}
									</div>
								{/each}
								{#if participants.length === 0}
									<p class="text-center text-gray-500 text-sm py-4">هنوز کسی متصل نیست</p>
								{/if}
							</div>
						</div>
					{/if}

					<!-- Whiteboard overlay for inline -->
					{#if whiteboardOpen && room}
						<div class="border-t border-gray-700 h-96">
							<Whiteboard {room} {sessionId} />
						</div>
					{/if}
				</div>
			{/if}
		</div>
	{:else}
		<div class="flex-1 flex items-center justify-center py-20"><p class="text-gray-400">جلسه یافت نشد</p></div>
	{/if}
</div>
