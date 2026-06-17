<script lang="ts">
	// @ts-nocheck
	import { page } from '$app/state';
	import { auth } from '$lib/stores';
	import { api } from '$lib/api';
	import { onMount, onDestroy } from 'svelte';
	import { JanusClient } from '$lib/classroom/janus-client';
	import { formatDuration } from '$lib/utils/persian';
	import type { Participant, ChatMessage } from '$lib/classroom/types';
	import ClassroomHeader from '$lib/components/classroom/ClassroomHeader.svelte';
	import Toolbar from '$lib/components/classroom/Toolbar.svelte';
	import ChatPanel from '$lib/components/classroom/ChatPanel.svelte';

	let session = $state<any>(null);
	let loading = $state(true);

	let janus = $state<JanusClient | null>(null);
	let connected = $state(false);
	let audioOn = $state(true);
	let micOn = $state(true);
	let webcamOn = $state(false);
	let screenShareOn = $state(false);
	let whiteboardOn = $state(false);
	let filesOn = $state(false);
	let handRaised = $state(false);
	let isRecording = $state(false);
	let elapsedSeconds = $state(0);
	let timerInterval: ReturnType<typeof setInterval> | null = null;

	let showChatPanel = $state(true);
	let participants = $state<Participant[]>([]);
	let chatMessages = $state<ChatMessage[]>([]);
	let localVideoEl: HTMLVideoElement;
	let remoteContainer: HTMLDivElement;
	let chatWs: WebSocket | null = null;
	let mediaRecorder = $state<MediaRecorder | null>(null);
	let recordingChunks = $state<Blob[]>([]);
	let recordingStartTime = $state(0);

	const sessionId = $derived(page.params.id!);
	const currentUserRole = $derived($auth.user?.role || 'student');
	const isTeacherOrAdmin = $derived(currentUserRole === 'teacher' || currentUserRole === 'admin');

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
				wsUrl: ws_url, roomId: room_id, userId: user_id, role,
				displayName: $auth.user?.display_name || 'کاربر',
			});
			await janus.connect();
			connected = true;
			startTimer();

			janus.onParticipantJoined = (p) => {
				participants = [...participants, {
					id: String(p.id), name: p.display || 'ناشناس', role: 'user',
					isSpeaking: false, hasVideo: p.video, hasAudio: p.audio,
					hasScreen: false, hasWhiteboard: false, handRaised: false,
				}];
			};

			janus.onStream = (stream, participant) => {
				if (remoteContainer) {
					const video = document.createElement('video');
					video.autoplay = true; video.playsInline = true;
					video.className = 'w-full h-full object-cover rounded';
					video.srcObject = stream;
					remoteContainer.appendChild(video);
				}
			};

			await janus.joinRoom();
		} catch (err) { console.error('Failed to connect:', err); connected = false; }
	}

	function startTimer() { elapsedSeconds = 0; timerInterval = setInterval(() => { elapsedSeconds++; }, 1000); }
	function stopTimer() { if (timerInterval) { clearInterval(timerInterval); timerInterval = null; } }

	function toggleAudio() { audioOn = !audioOn; }
	function toggleMic() { micOn = !micOn; janus?.toggleAudio(); }
	function toggleWebcam() { webcamOn = !webcamOn; janus?.toggleVideo(); }
	function toggleScreenShare() { screenShareOn = !screenShareOn; }
	function toggleWhiteboard() { whiteboardOn = !whiteboardOn; }
	function toggleFiles() { filesOn = !filesOn; }
	function toggleHand() { handRaised = !handRaised; }

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
		mediaRecorder.start(1000); isRecording = true; recordingStartTime = Date.now();
	}

	function stopRecording() { if (mediaRecorder && mediaRecorder.state !== 'inactive') mediaRecorder.stop(); isRecording = false; }

	function sendChatMessage(content: string) {
		if (chatWs?.readyState === WebSocket.OPEN) chatWs.send(JSON.stringify({ type: 'message', content }));
		else api.post(`/sessions/${sessionId}/messages`, { content });
		chatMessages = [...chatMessages, {
			id: String(Date.now()), sender: 'شما', content, isOwn: true,
			time: new Date().toLocaleTimeString('fa-IR', { hour: '2-digit', minute: '2-digit' })
		}];
	}

	function disconnect() {
		connected = false; stopTimer(); chatWs?.close(); stopRecording();
		if (localVideoEl?.srcObject) { (localVideoEl.srcObject as MediaStream).getTracks().forEach(t => t.stop()); localVideoEl.srcObject = null; }
		janus?.leave(); janus = null;
	}

	function leaveRoom() { disconnect(); window.close(); }
</script>

<div class="h-screen flex flex-col" style="background-color: #1a1a2e;">
	{#if loading}
		<div class="flex-1 flex items-center justify-center">
			<div class="animate-spin h-8 w-8 border-4 border-[#23b9d7] border-t-transparent rounded-full"></div>
		</div>
	{:else if session}
		<ClassroomHeader ownerName={$auth.user?.display_name || 'مالک'} roomName={session.title} {timerSeconds={elapsedSeconds}} />

		<div class="flex-1 flex overflow-hidden">
			<!-- Main Content -->
			<div class="flex-1 relative min-w-0 flex flex-col">
				<div class="flex-1 relative min-w-0">
					{#if !connected}
						<div class="absolute inset-0 flex flex-col items-center justify-center">
							<div class="w-20 h-20 rounded-full flex items-center justify-center mb-3" style="background-color: #3a3a5a;">
								<span class="text-2xl font-bold text-[#94a3b8]">{$auth.user?.display_name?.charAt(0) || '?'}</span>
							</div>
							<p class="text-[#94a3b8] text-[13px] mb-3">آماده پیوستن به کلاس</p>
							<button onclick={joinRoom} class="px-5 py-2.5 bg-[#23b9d7] text-white rounded-lg text-[13px] font-medium hover:bg-[#1a9ad4] transition-colors">پیوستن</button>
						</div>
					{/if}
					<div bind:this={remoteContainer} class="absolute inset-0 {gridCols} gap-1 p-1 auto-rows-fr"></div>
					<div class="absolute bottom-20 left-3 w-36 h-28 rounded overflow-hidden {webcamOn ? 'border border-[#3a3a5a]' : 'hidden'}" style="background-color: #1a1a2e;">
						<video bind:this={localVideoEl} autoplay muted playsinline class="w-full h-full object-cover"></video>
					</div>
				</div>

				<Toolbar bind:audioOn bind:micOn bind:webcamOn bind:screenShareOn bind:whiteboardOn bind:filesOn bind:handRaised bind:isRecording
					onToggleAudio={toggleAudio} onToggleMic={toggleMic} onToggleWebcam={toggleWebcam}
					onToggleScreenShare={toggleScreenShare} onToggleWhiteboard={toggleWhiteboard}
					onToggleFiles={toggleFiles} onToggleHand={toggleHand} onLeave={leaveRoom} />
			</div>

			<!-- Chat Panel (sidebar) -->
			{#if showChatPanel}
				<div class="flex flex-col shrink-0" style="width: 280px; background-color: #252540; border-left: 1px solid #3a3a5a;">
					<div class="flex-1 min-h-0 overflow-hidden flex flex-col">
						<ChatPanel messages={chatMessages} isAdmin={isTeacherOrAdmin} onSend={sendChatMessage} onClose={() => showChatPanel = false} />
					</div>
				</div>
			{/if}
		</div>
	{:else}
		<div class="flex-1 flex items-center justify-center"><p class="text-[#94a3b8]">جلسه یافت نشد</p></div>
	{/if}
</div>
