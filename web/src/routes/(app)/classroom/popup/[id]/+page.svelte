<script lang="ts">
	import { page } from '$app/state';
	import { auth } from '$lib/stores';
	import { api } from '$lib/api';
	import { onMount, onDestroy } from 'svelte';
	import { Room, RoomEvent, Track, ConnectionState, type Room as RoomType, type TrackPublication } from 'livekit-client';
	import { formatDuration, toPersianNum } from '$lib/utils/persian';
	import Whiteboard from '$lib/components/Whiteboard.svelte';
	import SettingsPopup from '$lib/components/SettingsPopup.svelte';
	import type { Session } from '$lib/types';

	let session = $state<Session | null>(null);
	let loading = $state(true);
	let room = $state<RoomType | null>(null);
	let connectionState = $state<ConnectionState>(ConnectionState.Disconnected);
	let audioEnabled = $state(true);
	let videoEnabled = $state(false);
	let screenSharing = $state(false);
	let whiteboardOpen = $state(false);
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
	let elapsedSeconds = $state(0);
	let timerInterval: ReturnType<typeof setInterval> | null = null;
	let handRaised = $state(false);
	let handsRaised = $state<Record<string, boolean>>({});
	let unreadCount = $state(0);
	let chatOpen = $state(true);
	let participantsOpen = $state(true);
	let showSettings = $state(false);
	let activeView = $state<'video' | 'whiteboard' | 'screenshare'>('video');
	let pollsOpen = $state(false);
	let showCreatePoll = $state(false);
	let polls = $state<{id: number; question: string; options: string[]; is_active: boolean}[]>([]);
	let pollResults = $state<Record<number, {votes: number[]; total_votes: number}>>({});
	let votedPolls = $state<Record<number, number>>({});
	let loadingPolls = $state(false);

	// Create poll form state
	let pollQuestion = $state('');
	let pollOptions = $state<string[]>(['', '']);
	let creatingPoll = $state(false);

	const sessionId = $derived(page.params.id);
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
	onDestroy(() => { stopRecording(); disconnect(); if (timerInterval) clearInterval(timerInterval); });

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
				startTimer();
			});

			room.on(RoomEvent.Disconnected, () => {
				connectionState = ConnectionState.Disconnected;
				stopTimer();
			});

			room.on(RoomEvent.TrackSubscribed, (track: TrackPublication, participant) => {
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

			await room.connect(url, token);
		} catch (err) {
			console.error('Failed to connect:', err);
			connectionState = ConnectionState.Disconnected;
		}
	}

	function startTimer() {
		elapsedSeconds = 0;
		timerInterval = setInterval(() => { elapsedSeconds++; }, 1000);
	}

	function stopTimer() {
		if (timerInterval) { clearInterval(timerInterval); timerInterval = null; }
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

	function disconnect() {
		connected = false;
		stopTimer();
		chatWs?.close();
		stopRecording();
		if (localVideoEl?.srcObject) {
			(localVideoEl.srcObject as MediaStream).getTracks().forEach(t => t.stop());
			localVideoEl.srcObject = null;
		}
		if (room) { room.disconnect(); room = null; }
		connectionState = ConnectionState.Disconnected;
	}

	function leaveRoom() {
		disconnect();
		window.close();
	}

	// Poll functions
	async function loadPolls() {
		loadingPolls = true;
		const res = await api.get<typeof polls>(`/sessions/${sessionId}/polls`);
		if (res.success && res.data) {
			polls = res.data;
			// Load results for each poll
			for (const poll of polls) {
				if (!pollResults[poll.id]) {
					const resultRes = await api.get<{votes: number[]; total_votes: number}>(`/polls/${poll.id}/results`);
					if (resultRes.success && resultRes.data) {
						pollResults[poll.id] = resultRes.data;
					}
				}
			}
		}
		loadingPolls = false;
	}

	async function vote(pollId: number, optionIndex: number) {
		const res = await api.post(`/polls/${pollId}/vote`, { option_index: optionIndex });
		if (res.success) {
			votedPolls[pollId] = optionIndex;
			// Reload results
			const resultRes = await api.get<{votes: number[]; total_votes: number}>(`/polls/${pollId}/results`);
			if (resultRes.success && resultRes.data) {
				pollResults[pollId] = resultRes.data;
			}
		}
	}

	async function closePoll(pollId: number) {
		if (!confirm('آیا مطمئن هستید که می‌خواهید این نظرسنجی را ببندید؟')) return;
		const res = await api.post(`/polls/${pollId}/close`, {});
		if (res.success) {
			await loadPolls();
		}
	}

	function addPollOption() {
		pollOptions = [...pollOptions, ''];
	}

	function removePollOption(index: number) {
		if (pollOptions.length <= 2) return;
		pollOptions = pollOptions.filter((_, i) => i !== index);
	}

	async function createPoll() {
		if (!pollQuestion.trim()) return;
		const validOptions = pollOptions.filter(o => o.trim());
		if (validOptions.length < 2) return;

		creatingPoll = true;
		const res = await api.post(`/sessions/${sessionId}/polls`, {
			question: pollQuestion.trim(),
			options: validOptions
		});
		if (res.success) {
			showCreatePoll = false;
			pollQuestion = '';
			pollOptions = ['', ''];
			await loadPolls();
		}
		creatingPoll = false;
	}

	function togglePollsPanel() {
		pollsOpen = !pollsOpen;
		if (pollsOpen && polls.length === 0) {
			loadPolls();
		}
	}
</script>

<div class="h-screen flex flex-col text-white" style="background-color: #1a1a2e;">
	{#if loading}
		<div class="flex-1 flex items-center justify-center">
			<div class="animate-spin h-8 w-8 border-4 border-blue-500 border-t-transparent rounded-full"></div>
		</div>
	{:else if session}
		<!-- Top bar -->
		<div class="flex items-center justify-between px-4 py-2 border-b shrink-0" style="background-color: #16213e; border-color: #2a2a4a;">
			<div class="flex items-center gap-3">
				<h1 class="font-bold text-sm">{session.title}</h1>
				{#if connectionState === ConnectionState.Connected}
					<span class="flex items-center gap-1.5 text-xs text-green-400"><span class="w-2 h-2 bg-green-400 rounded-full animate-pulse"></span>متصل</span>
				{:else}
					<span class="text-xs text-gray-400">قطع شده</span>
				{/if}
				{#if isRecording}
					<span class="flex items-center gap-1.5 text-xs text-red-400"><span class="w-2 h-2 bg-red-500 rounded-full animate-pulse"></span>ضبط</span>
				{/if}
			</div>
			<div class="flex items-center gap-3">
				<span class="text-xs text-gray-400">{formatDuration(elapsedSeconds)}</span>
				<button onclick={() => window.close()} class="text-gray-400 hover:text-white p-1" title="بستن">
					<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" /></svg>
				</button>
			</div>
		</div>

		<!-- Main content: 3-column layout -->
		<div class="flex-1 flex overflow-hidden">
			<!-- Video / Whiteboard area (~55%) -->
			<div class="flex-1 relative min-w-0 flex flex-col">
				<div class="flex gap-2 px-3 py-2 shrink-0" style="background-color: #16213e; border-color: #2a2a4a;">
					<button onclick={() => activeView = 'video'} class="px-3 py-1.5 text-xs rounded-lg {activeView === 'video' ? 'bg-blue-600 text-white' : 'bg-gray-700 text-gray-300'}">🎥 ویدیو</button>
					<button onclick={() => activeView = 'whiteboard'} class="px-3 py-1.5 text-xs rounded-lg {activeView === 'whiteboard' ? 'bg-purple-600 text-white' : 'bg-gray-700 text-gray-300'}">📋 تخته‌سفید</button>
				</div>
				<div class="flex-1 relative min-w-0">
				{#if activeView === 'whiteboard' && room}
					<Whiteboard {room} {sessionId} />
				{:else}
					<div bind:this={remoteContainer} class="absolute inset-0 {gridCols} gap-2 p-2 auto-rows-fr"></div>
					{#if connectionState !== ConnectionState.Connected}
						<div class="absolute inset-0 flex flex-col items-center justify-center">
							<div class="w-20 h-20 bg-gray-700 rounded-full flex items-center justify-center mb-3">
								<span class="text-2xl font-bold text-gray-400">{$auth.user?.display_name?.charAt(0) || '?'}</span>
							</div>
							<p class="text-gray-400 text-sm mb-3">آماده پیوستن به کلاس</p>
							<button onclick={joinRoom} class="px-5 py-2.5 bg-blue-600 text-white rounded-xl text-sm font-medium hover:bg-blue-700 transition-colors">پیوستن</button>
						</div>
					{/if}
					<div class="absolute bottom-3 left-3 w-36 h-28 rounded-lg overflow-hidden border-2 bg-gray-800 {videoEnabled ? 'border-gray-600' : 'border-transparent hidden'}">
						<video bind:this={localVideoEl} autoplay muted playsinline class="w-full h-full object-cover"></video>
					</div>
				{/if}
				</div>
			</div>

			<!-- Chat panel (280px) -->
			<div class="w-[280px] flex flex-col shrink-0 border-l" style="background-color: #16213e; border-color: #2a2a4a;">
				<div class="px-3 py-2.5 border-b" style="border-color: #2a2a4a;">
					<h3 class="font-bold text-xs text-gray-300">گفتگو</h3>
				</div>
				<div class="flex-1 overflow-y-auto px-3 py-2 space-y-2">
					{#each chatMessages as msg}
						<div>
							<div class="flex items-center gap-2">
								<span class="text-[10px] font-bold text-blue-400">{msg.sender}</span>
								<span class="text-[9px] text-gray-500">{msg.time}</span>
							</div>
							<p class="text-xs mt-0.5 text-gray-200">{msg.content}</p>
						</div>
					{/each}
				</div>
				<div class="px-2 py-2 border-t" style="border-color: #2a2a4a;">
					<form onsubmit={(e) => { e.preventDefault(); sendChat(); }} class="flex gap-1.5">
						<input type="text" bind:value={chatInput} class="flex-1 px-2.5 py-1.5 rounded-lg text-xs focus:ring-1 focus:ring-blue-500 outline-none" style="background-color: #2a2a4a;" placeholder="پیام..." />
						<button type="submit" class="px-2.5 py-1.5 bg-blue-600 rounded-lg text-xs hover:bg-blue-700">ارسال</button>
					</form>
				</div>
			</div>

			<!-- Participants panel (220px) -->
			<div class="w-[220px] flex flex-col shrink-0 border-l" style="background-color: #16213e; border-color: #2a2a4a;">
				<div class="px-3 py-2.5 border-b" style="border-color: #2a2a4a;">
					<h3 class="font-bold text-xs text-gray-300">شرکت‌کنندگان ({participants.length})</h3>
				</div>
				<div class="flex-1 overflow-y-auto px-2 py-2 space-y-1">
					{#each participants as p}
						<div class="flex items-center gap-2 px-2 py-1.5 rounded-lg {p.isSpeaking ? 'bg-blue-900/30' : ''}">
							<div class="relative">
								<div class="w-7 h-7 rounded-full bg-gray-600 flex items-center justify-center text-[10px] font-bold">
									{p.name.charAt(0)}
								</div>
								{#if p.isSpeaking}
									<span class="absolute -bottom-0.5 -right-0.5 w-2.5 h-2.5 bg-green-500 rounded-full border-2" style="border-color: #16213e;"></span>
								{/if}
							</div>
							<div class="flex-1 min-w-0">
								<p class="text-xs font-medium truncate">
									{p.name}{'isLocal' in p && p.isLocal ? ' (شما)' : ''}
									{#if p.handRaised}<span class="text-yellow-400 ml-1">✋</span>{/if}
								</p>
							</div>
							<div class="flex items-center gap-1">
								{#if p.hasAudio}<span class="w-1.5 h-1.5 bg-green-400 rounded-full"></span>{:else}<span class="w-1.5 h-1.5 bg-red-400 rounded-full"></span>{/if}
								{#if p.hasVideo}<span class="w-1.5 h-1.5 bg-green-400 rounded-full"></span>{:else}<span class="w-1.5 h-1.5 bg-gray-500 rounded-full"></span>{/if}
							</div>
							{#if isTeacherOrAdmin && !('isLocal' in p && p.isLocal)}
								<button onclick={() => muteParticipant(p.identity)} class="text-[10px] px-1.5 py-0.5 rounded hover:bg-gray-700 text-gray-400" title="بی‌صدا کردن">🔇</button>
								<button onclick={() => removeParticipant(p.identity)} class="text-[10px] px-1.5 py-0.5 rounded hover:bg-red-600 text-gray-400 hover:text-white" title="حذف">✕</button>
							{/if}
						</div>
					{/each}
					{#if participants.length === 0}
						<p class="text-center text-gray-500 text-xs py-3">هنوز کسی متصل نیست</p>
					{/if}
				</div>
			</div>
		</div>

		<!-- Bottom control bar -->
		<div class="flex items-center justify-center gap-2.5 px-4 py-3 border-t shrink-0" style="background-color: #16213e; border-color: #2a2a4a;">
			<button onclick={toggleAudio} class="w-10 h-10 rounded-full flex items-center justify-center transition-colors {audioEnabled ? 'bg-gray-700 hover:bg-gray-600' : 'bg-red-600 hover:bg-red-700'}" title={audioEnabled ? 'بی‌صدا' : 'صدا'}>
				{#if audioEnabled}<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M19 11a7 7 0 01-7 7m0 0a7 7 0 01-7-7m7 7v4m0 0H8m4 0h4m-4-8a3 3 0 01-3-3V5a3 3 0 116 0v6a3 3 0 01-3 3z" /></svg>{:else}<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M5.586 15H4a1 1 0 01-1-1v-4a1 1 0 011-1h1.586l4.707-4.707C10.923 3.663 12 4.109 12 5v14c0 .891-1.077 1.337-1.707.707L5.586 15z" /><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M17 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2" /></svg>{/if}
			</button>
			<button onclick={toggleVideo} class="w-10 h-10 rounded-full flex items-center justify-center transition-colors {videoEnabled ? 'bg-gray-700 hover:bg-gray-600' : 'bg-red-600 hover:bg-red-700'}" title={videoEnabled ? 'ویدیو خاموش' : 'روشن کردن ویدیو'}>
				{#if videoEnabled}<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" /></svg>{:else}<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" /><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M17 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2" /></svg>{/if}
			</button>
			<button onclick={toggleScreenShare} class="w-10 h-10 rounded-full flex items-center justify-center transition-colors {screenSharing ? 'bg-blue-600 hover:bg-blue-700' : 'bg-gray-700 hover:bg-gray-600'}" title="اشتراک‌گذاری صفحه">
				<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" /></svg>
			</button>
			<button onclick={() => whiteboardOpen = !whiteboardOpen} class="w-10 h-10 rounded-full flex items-center justify-center transition-colors {whiteboardOpen ? 'bg-purple-600 hover:bg-purple-700' : 'bg-gray-700 hover:bg-gray-600'}" title="تخته‌سفید">
				<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" /></svg>
			</button>

			<!-- Hand raise -->
			<button onclick={toggleHandRaise} class="w-10 h-10 rounded-full flex items-center justify-center transition-colors {handRaised ? 'bg-yellow-500 hover:bg-yellow-600' : 'bg-gray-700 hover:bg-gray-600'}" title={handRaised ? 'پایین آوردن دست' : 'بالا بردن دست'}>
				<span class="text-lg">✋</span>
			</button>

			<button onclick={isRecording ? stopRecording : startRecording} class="w-10 h-10 rounded-full flex items-center justify-center transition-colors {isRecording ? 'bg-red-600 hover:bg-red-700 animate-pulse' : 'bg-gray-700 hover:bg-gray-600'}" title={isRecording ? 'پایان ضبط' : 'ضبط'}>
				<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><circle cx="12" cy="12" r="10" /><circle cx="12" cy="12" r="4" fill="currentColor" /></svg>
			</button>

			<button onclick={() => showSettings = true} class="w-10 h-10 rounded-full flex items-center justify-center transition-colors bg-gray-700 hover:bg-gray-600" title="تنظیمات">
				<span class="text-lg">⚙️</span>
			</button>

			<!-- Polls button -->
			<button onclick={togglePollsPanel} class="w-10 h-10 rounded-full flex items-center justify-center transition-colors {pollsOpen ? 'bg-green-600 hover:bg-green-700' : 'bg-gray-700 hover:bg-gray-600'}" title="نظرسنجی‌ها">
				<span class="text-lg">📊</span>
			</button>

			<div class="w-px h-6 mx-1" style="background-color: #2a2a4a;"></div>

			<button onclick={leaveRoom} class="px-4 py-2 bg-red-600 text-white rounded-full text-xs font-medium hover:bg-red-700 transition-colors flex items-center gap-1.5">
				<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" /></svg>
				خروج
			</button>
		</div>

		<SettingsPopup bind:show={showSettings} />

		<!-- Polls Panel Overlay -->
		{#if pollsOpen}
			<div class="absolute inset-0 z-40 flex items-center justify-center" style="background-color: rgba(0,0,0,0.5);" onclick={(e) => { if (e.target === e.currentTarget) pollsOpen = false; }}>
				<div class="w-[480px] max-h-[80vh] rounded-2xl shadow-2xl flex flex-col" style="background-color: #1e1e3a; border: 1px solid #2a2a4a;">
					<!-- Header -->
					<div class="flex items-center justify-between px-5 py-4 border-b" style="border-color: #2a2a4a;">
						<h2 class="text-lg font-bold text-white flex items-center gap-2">
							<span>📊</span> نظرسنجی‌ها
						</h2>
						<button onclick={() => pollsOpen = false} class="text-gray-400 hover:text-white p-1">
							<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" /></svg>
						</button>
					</div>

					<!-- Content -->
					<div class="flex-1 overflow-y-auto p-4 space-y-4">
						{#if loadingPolls}
							<div class="flex items-center justify-center py-8">
								<div class="animate-spin h-6 w-6 border-3 border-blue-500 border-t-transparent rounded-full"></div>
							</div>
						{:else if polls.length === 0}
							<div class="text-center py-8">
								<p class="text-gray-400 text-sm">هنوز نظرسنجی‌ای ایجاد نشده است</p>
							</div>
						{:else}
							{#each polls as poll (poll.id)}
								<div class="rounded-xl p-4" style="background-color: #252545; border: 1px solid #3a3a5a;">
									<div class="flex items-start justify-between mb-3">
										<h3 class="text-sm font-medium text-white flex-1">{poll.question}</h3>
										<div class="flex items-center gap-2">
											{#if poll.is_active}
												<span class="text-[10px] px-2 py-0.5 rounded-full bg-green-600/20 text-green-400">فعال</span>
											{:else}
												<span class="text-[10px] px-2 py-0.5 rounded-full bg-gray-600/20 text-gray-400">بسته شده</span>
											{/if}
											{#if isTeacherOrAdmin && poll.is_active}
												<button onclick={() => closePoll(poll.id)} class="text-[10px] px-2 py-0.5 rounded hover:bg-red-600/20 text-gray-400 hover:text-red-400" title="بستن نظرسنجی">بستن</button>
											{/if}
										</div>
									</div>

									<!-- Voting UI or Results -->
									{#if votedPolls[poll.id] !== undefined || !poll.is_active}
										<!-- Show results -->
										{#if pollResults[poll.id]}
											<div class="space-y-2">
												{#each poll.options as option, i}
													<div class="relative">
														<div class="flex items-center justify-between text-xs mb-1">
															<span class="text-gray-300">{option}</span>
															<span class="text-gray-400">
																{toPersianNum(pollResults[poll.id].votes[i] || 0)} رأی
																({toPersianNum(pollResults[poll.id].total_votes > 0 ? Math.round(((pollResults[poll.id].votes[i] || 0) / pollResults[poll.id].total_votes) * 100) : 0)}%)
															</span>
														</div>
														<div class="h-2 rounded-full bg-gray-700 overflow-hidden">
															<div
																class="h-full rounded-full transition-all duration-500 {votedPolls[poll.id] === i ? 'bg-blue-500' : 'bg-blue-600'}"
																style="width: {pollResults[poll.id].total_votes > 0 ? ((pollResults[poll.id].votes[i] || 0) / pollResults[poll.id].total_votes) * 100 : 0}%"
															></div>
														</div>
													</div>
												{/each}
												<p class="text-[10px] text-gray-500 mt-2">مجموع آرا: {toPersianNum(pollResults[poll.id].total_votes)}</p>
											</div>
										{:else}
											<div class="flex items-center justify-center py-4">
												<div class="animate-spin h-4 w-4 border-2 border-blue-500 border-t-transparent rounded-full"></div>
											</div>
										{/if}
									{:else}
										<!-- Show voting options -->
										<div class="space-y-2">
											{#each poll.options as option, i}
												<label class="flex items-center gap-3 p-2 rounded-lg cursor-pointer hover:bg-white/5 transition-colors">
													<input type="radio" name="poll-{poll.id}" value={i} class="w-4 h-4 text-blue-600 focus:ring-blue-500" style="accent-color: #3b82f6;" />
													<span class="text-sm text-gray-200">{option}</span>
												</label>
											{/each}
											<button
												onclick={() => {
													const selected = document.querySelector(`input[name="poll-${poll.id}"]:checked`) as HTMLInputElement;
													if (selected) vote(poll.id, parseInt(selected.value));
												}}
												class="w-full mt-2 px-4 py-2 bg-blue-600 text-white rounded-lg text-sm font-medium hover:bg-blue-700 transition-colors"
											>
												ثبت رأی
											</button>
										</div>
									{/if}
								</div>
							{/each}
						{/if}
					</div>

					<!-- Footer with Create Poll button -->
					{#if isTeacherOrAdmin}
						<div class="px-4 py-3 border-t" style="border-color: #2a2a4a;">
							<button
								onclick={() => showCreatePoll = true}
								class="w-full px-4 py-2.5 bg-green-600 text-white rounded-xl text-sm font-medium hover:bg-green-700 transition-colors flex items-center justify-center gap-2"
							>
								<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" /></svg>
								ایجاد نظرسنجی جدید
							</button>
						</div>
					{/if}
				</div>
			</div>
		{/if}

		<!-- Create Poll Modal -->
		{#if showCreatePoll}
			<div class="absolute inset-0 z-50 flex items-center justify-center" style="background-color: rgba(0,0,0,0.6);" onclick={(e) => { if (e.target === e.currentTarget) showCreatePoll = false; }}>
				<div class="w-[440px] rounded-2xl shadow-2xl" style="background-color: #1e1e3a; border: 1px solid #2a2a4a;">
					<!-- Header -->
					<div class="flex items-center justify-between px-5 py-4 border-b" style="border-color: #2a2a4a;">
						<h2 class="text-lg font-bold text-white">ایجاد نظرسنجی جدید</h2>
						<button onclick={() => showCreatePoll = false} class="text-gray-400 hover:text-white p-1">
							<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" /></svg>
						</button>
					</div>

					<!-- Form -->
					<div class="p-5 space-y-4">
						<div>
							<label class="block text-sm text-gray-300 mb-2">سوال</label>
							<input
								type="text"
								bind:value={pollQuestion}
								class="w-full px-4 py-2.5 rounded-xl text-sm text-white focus:ring-2 focus:ring-blue-500 outline-none"
								style="background-color: #252545; border: 1px solid #3a3a5a;"
								placeholder="سوال نظرسنجی را وارد کنید..."
							/>
						</div>

						<div>
							<label class="block text-sm text-gray-300 mb-2">گزینه‌ها</label>
							<div class="space-y-2">
								{#each pollOptions as option, i}
									<div class="flex items-center gap-2">
										<input
											type="text"
											bind:value={pollOptions[i]}
											class="flex-1 px-4 py-2 rounded-lg text-sm text-white focus:ring-2 focus:ring-blue-500 outline-none"
											style="background-color: #252545; border: 1px solid #3a3a5a;"
											placeholder="گزینه {toPersianNum(i + 1)}"
										/>
										{#if pollOptions.length > 2}
											<button
												onclick={() => removePollOption(i)}
												class="p-2 text-gray-400 hover:text-red-400 hover:bg-red-600/10 rounded-lg transition-colors"
												title="حذف گزینه"
											>
												<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" /></svg>
											</button>
										{/if}
									</div>
								{/each}
							</div>
							<button
								onclick={addPollOption}
								class="mt-3 w-full px-4 py-2 border-2 border-dashed border-gray-600 text-gray-400 rounded-xl text-sm hover:border-blue-500 hover:text-blue-400 transition-colors flex items-center justify-center gap-2"
							>
								<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" /></svg>
								افزودن گزینه
							</button>
						</div>
					</div>

					<!-- Footer -->
					<div class="px-5 py-4 border-t flex gap-3" style="border-color: #2a2a4a;">
						<button
							onclick={() => showCreatePoll = false}
							class="flex-1 px-4 py-2.5 bg-gray-700 text-white rounded-xl text-sm font-medium hover:bg-gray-600 transition-colors"
						>
							انصراف
						</button>
						<button
							onclick={createPoll}
							disabled={creatingPoll || !pollQuestion.trim() || pollOptions.filter(o => o.trim()).length < 2}
							class="flex-1 px-4 py-2.5 bg-green-600 text-white rounded-xl text-sm font-medium hover:bg-green-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
						>
							{#if creatingPoll}
								<div class="animate-spin h-4 w-4 border-2 border-white border-t-transparent rounded-full"></div>
							{/if}
							ایجاد نظرسنجی
						</button>
					</div>
				</div>
			</div>
		{/if}
	{:else}
		<div class="flex-1 flex items-center justify-center"><p class="text-gray-400">جلسه یافت نشد</p></div>
	{/if}
</div>
