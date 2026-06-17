<script lang="ts">
	// @ts-nocheck
	import { page } from '$app/state';
	import { auth } from '$lib/stores';
	import { api } from '$lib/api';
	import { onMount, onDestroy } from 'svelte';
	import { JanusClient } from '$lib/classroom/janus-client';
	import { formatDuration, toPersianNum } from '$lib/utils/persian';
	import type { Participant, ChatMessage } from '$lib/classroom/types';
	import Toolbar from '$lib/components/classroom/Toolbar.svelte';
	import UsersPanel from '$lib/components/classroom/UsersPanel.svelte';
	import ChatPanel from '$lib/components/classroom/ChatPanel.svelte';
	import AppMenu from '$lib/components/classroom/AppMenu.svelte';
	import Whiteboard from '$lib/components/Whiteboard.svelte';

	let session = $state<any>(null);
	let loading = $state(true);

	let janus = $state<JanusClient | null>(null);
	let connected = $state(false);
	let audioEnabled = $state(true);
	let videoEnabled = $state(false);
	let screenSharing = $state(false);
	let whiteboardOpen = $state(false);
	let showUsersPanel = $state(true);
	let showChatPanel = $state(true);
	let showAppMenu = $state(false);
	let handRaised = $state(false);
	let isRecording = $state(false);
	let elapsedSeconds = $state(0);
	let timerInterval: ReturnType<typeof setInterval> | null = null;

	let participants = $state<Participant[]>([]);
	let chatMessages = $state<ChatMessage[]>([]);
	let localVideoEl: HTMLVideoElement;
	let remoteContainer: HTMLDivElement;
	let chatWs: WebSocket | null = null;
	let mediaRecorder = $state<MediaRecorder | null>(null);
	let recordingChunks = $state<Blob[]>([]);
	let recordingStartTime = $state(0);
	let showSettings = $state(false);

	const sessionId = $derived(page.params.id!);
	const currentUserRole = $derived($auth.user?.role || 'student');
	const isTeacherOrAdmin = $derived(currentUserRole === 'teacher' || currentUserRole === 'admin');
	const isOwner = $derived(currentUserRole === 'admin');

	const gridCols = $derived.by(() => {
		const count = participants.length;
		if (count <= 1) return 'grid-cols-1';
		if (count <= 2) return 'grid-cols-2';
		if (count <= 4) return 'grid-cols-2';
		if (count <= 6) return 'grid-cols-3';
		return 'grid-cols-4';
	});

	onMount(async () => {
		const res = await api.get(`/sessions/${sessionId}`);
		if (res.success) session = res.data!;
		loading = false;
		connectChatWs();
	});
	onDestroy(() => { stopRecording(); disconnect(); if (timerInterval) clearInterval(timerInterval); });

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
						id: String(Date.now()),
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
		const joinRes = await api.get(`/sessions/${sessionId}/classroom`);
		if (!joinRes.success || !joinRes.data) { alert(joinRes.error || 'خطا در دریافت اطلاعات اتاق'); return; }

		const { ws_url, room_id, user_id, role } = joinRes.data;
		try {
			janus = new JanusClient({
				wsUrl: ws_url,
				roomId: room_id,
				userId: user_id,
				role: role,
				displayName: $auth.user?.display_name || 'کاربر',
			});

			await janus.connect();
			connected = true;
			startTimer();

			janus.onParticipantJoined = (p) => {
				participants = [...participants, {
					id: String(p.id),
					name: p.display || 'ناشناس',
					role: 'user',
					isSpeaking: false,
					hasVideo: p.video,
					hasAudio: p.audio,
					hasScreen: false,
					hasWhiteboard: false,
					handRaised: false,
				}];
			};

			janus.onStream = (stream, participant) => {
				if (remoteContainer) {
					const video = document.createElement('video');
					video.autoplay = true;
					video.playsInline = true;
					video.className = 'w-full h-full object-cover rounded-lg';
					video.srcObject = stream;
					remoteContainer.appendChild(video);
				}
			};

			await janus.joinRoom();
		} catch (err) {
			console.error('Failed to connect:', err);
			connected = false;
		}
	}

	function startTimer() {
		elapsedSeconds = 0;
		timerInterval = setInterval(() => { elapsedSeconds++; }, 1000);
	}

	function stopTimer() {
		if (timerInterval) { clearInterval(timerInterval); timerInterval = null; }
	}

	function toggleAudio() { audioEnabled = !audioEnabled; janus?.toggleAudio(); }
	function toggleVideo() { videoEnabled = !videoEnabled; janus?.toggleVideo(); }
	function toggleScreenShare() { screenSharing = !screenSharing; }
	function toggleWhiteboard() { whiteboardOpen = !whiteboardOpen; }
	function toggleHandRaise() { handRaised = !handRaised; }

	function startRecording() {
		const streams: MediaStream[] = [];
		if (localVideoEl?.srcObject) streams.push(localVideoEl.srcObject as MediaStream);
		if (streams.length === 0) { alert('ابتدا صدا یا ویدیو را فعال کنید'); return; }

		const combined = new MediaStream(streams.flatMap(s => s.getTracks()));
		try { mediaRecorder = new MediaRecorder(combined, { mimeType: 'video/webm;codecs=vp9' }); }
		catch { mediaRecorder = new MediaRecorder(combined); }

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
				method: 'POST', headers: { 'Authorization': `Bearer ${token}` }, body: formData
			});
		};
		mediaRecorder.start(1000);
		isRecording = true;
		recordingStartTime = Date.now();
	}

	function stopRecording() {
		if (mediaRecorder && mediaRecorder.state !== 'inactive') mediaRecorder.stop();
		isRecording = false;
	}

	function sendChatMessage(content: string) {
		if (chatWs?.readyState === WebSocket.OPEN) {
			chatWs.send(JSON.stringify({ type: 'message', content }));
		} else {
			api.post(`/sessions/${sessionId}/messages`, { content });
		}
		chatMessages = [...chatMessages, {
			id: String(Date.now()),
			sender: 'شما', content, isOwn: true,
			time: new Date().toLocaleTimeString('fa-IR', { hour: '2-digit', minute: '2-digit' })
		}];
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
		janus?.leave();
		janus = null;
	}

	function leaveRoom() {
		disconnect();
		window.close();
	}
</script>

<div class="h-screen flex flex-col text-white" style="background-color: #1a1a2e;">
	{#if loading}
		<div class="flex-1 flex items-center justify-center">
			<div class="animate-spin h-8 w-8 border-4 border-blue-500 border-t-transparent rounded-full"></div>
		</div>
	{:else if session}
		<!-- Header -->
		<div class="flex items-center justify-between px-4 py-2 border-b shrink-0" style="background-color: #16213e; border-color: #2a2a4a;">
			<div class="flex items-center gap-3">
				<h1 class="font-bold text-sm">{session.title}</h1>
				{#if connected}
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

		<!-- 3-column layout -->
		<div class="flex-1 flex overflow-hidden">
			<!-- Users Panel -->
			{#if showUsersPanel}
				<UsersPanel
					{participants}
					{currentUserRole}
					onClose={() => showUsersPanel = false}
				/>
			{/if}

			<!-- Main Content -->
			<div class="flex-1 relative min-w-0 flex flex-col">
				<!-- View Tabs -->
				<div class="flex gap-2 px-3 py-2 shrink-0" style="background-color: #16213e; border-color: #2a2a4a;">
					<button onclick={() => whiteboardOpen = false} class="px-3 py-1.5 text-xs rounded-lg {!whiteboardOpen ? 'bg-blue-600 text-white' : 'bg-gray-700 text-gray-300'}">🎥 ویدیو</button>
					<button onclick={() => whiteboardOpen = true} class="px-3 py-1.5 text-xs rounded-lg {whiteboardOpen ? 'bg-purple-600 text-white' : 'bg-gray-700 text-gray-300'}">📋 تخته‌سفید</button>
				</div>

				<div class="flex-1 relative min-w-0">
					{#if whiteboardOpen}
						<Whiteboard {sessionId} />
					{:else}
						<div bind:this={remoteContainer} class="absolute inset-0 {gridCols} gap-2 p-2 auto-rows-fr"></div>
						{#if !connected}
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

			<!-- Chat Panel -->
			{#if showChatPanel}
				<ChatPanel
					messages={chatMessages}
					isAdmin={isTeacherOrAdmin}
					onSend={sendChatMessage}
					onClose={() => showChatPanel = false}
				/>
			{/if}
		</div>

		<!-- Toolbar -->
		<Toolbar
			bind:audioEnabled
			bind:videoEnabled
			bind:screenSharing
			bind:whiteboardOpen
			bind:handRaised
			bind:isRecording
			onToggleAudio={toggleAudio}
			onToggleVideo={toggleVideo}
			onToggleScreenShare={toggleScreenShare}
			onToggleWhiteboard={toggleWhiteboard}
			onToggleHandRaise={toggleHandRaise}
			onToggleRecording={() => { isRecording ? stopRecording() : startRecording(); }}
			onLeave={leaveRoom}
		/>
	{:else}
		<div class="flex-1 flex items-center justify-center"><p class="text-gray-400">جلسه یافت نشد</p></div>
	{/if}
</div>
