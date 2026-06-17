<script lang="ts">
	// @ts-nocheck
	import { page } from '$app/state';
	import { auth } from '$lib/stores';
	import { api } from '$lib/api';
	import { onMount, onDestroy } from 'svelte';
	import { JanusClient } from '$lib/classroom/janus-client';
	import { formatDuration, toPersianNum } from '$lib/utils/persian';
	import { classroomWindow } from '$lib/classroom/ClassroomWindow';
	import type { Participant, ChatMessage } from '$lib/classroom/types';
	import Toolbar from '$lib/components/classroom/Toolbar.svelte';
	import UsersPanel from '$lib/components/classroom/UsersPanel.svelte';
	import ChatPanel from '$lib/components/classroom/ChatPanel.svelte';
	import AppMenu from '$lib/components/classroom/AppMenu.svelte';
	import Whiteboard from '$lib/components/Whiteboard.svelte';

	let session = $state<any>(null);
	let loading = $state(true);
	let popupBlocked = $state(false);

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

	onMount(async () => { await loadSession(); });
	onDestroy(() => { stopRecording(); disconnectInline(); });

	async function loadSession() {
		loading = true;
		const res = await api.get(`/sessions/${sessionId}`);
		if (res.success) session = res.data!;
		loading = false;
	}

	function openPopup() {
		if (!session) return;
		const win = classroomWindow.open(String(session.id), session.title);
		if (!win) popupBlocked = true;
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

	async function joinInlineRoom() {
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
			elapsedSeconds = 0;
			timerInterval = setInterval(() => { elapsedSeconds++; }, 1000);

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
			connectChatWs();
		} catch (err) {
			console.error('Failed to connect:', err);
			connected = false;
		}
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

	function disconnectInline() {
		connected = false;
		chatWs?.close();
		stopRecording();
		if (timerInterval) { clearInterval(timerInterval); timerInterval = null; }
		if (localVideoEl?.srcObject) {
			(localVideoEl.srcObject as MediaStream).getTracks().forEach(t => t.stop());
			localVideoEl.srcObject = null;
		}
		janus?.leave();
		janus = null;
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
					<!-- Header -->
					<div class="flex items-center justify-between px-4 py-2" style="background-color: #16213e; border-bottom: 1px solid #2a2a4a;">
						<div class="flex items-center gap-3">
							<h2 class="text-sm font-bold text-white">{session.title}</h2>
							{#if connected}
								<span class="flex items-center gap-1.5 text-xs text-green-400"><span class="w-2 h-2 bg-green-400 rounded-full animate-pulse"></span>متصل</span>
								<span class="text-xs text-gray-400 font-mono">{formatDuration(elapsedSeconds)}</span>
							{:else}
								<span class="text-xs text-gray-400">قطع شده</span>
							{/if}
							{#if isRecording}
								<span class="flex items-center gap-1.5 text-xs text-red-400"><span class="w-2 h-2 bg-red-500 rounded-full animate-pulse"></span>ضبط</span>
							{/if}
						</div>
						<div class="flex items-center gap-2">
							<!-- App Menu Toggle -->
							<button onclick={() => showAppMenu = !showAppMenu} class="relative p-1.5 rounded-lg hover:bg-gray-700 text-gray-400 hover:text-white" title="منو">
								<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M3.75 6.75h16.5M3.75 12h16.5m-16.5 5.25h16.5" /></svg>
							</button>
							{#if showAppMenu}
								<AppMenu
									bind:showUsersPanel
									bind:showChatPanel
									{isOwner}
									onSettings={() => showSettings = true}
									onExit={() => { disconnectInline(); history.back(); }}
									onCloseRoom={() => {}}
									onClose={() => showAppMenu = false}
								/>
							{/if}
						</div>
					</div>

					<!-- 3-column layout -->
					<div class="flex" style="height: 500px;">
						<!-- Users Panel -->
						{#if showUsersPanel}
							<UsersPanel
								{participants}
								{currentUserRole}
								onClose={() => showUsersPanel = false}
							/>
						{/if}

						<!-- Main Content -->
						<div class="flex-1 relative min-w-0 bg-gray-900">
							{#if !connected}
								<div class="absolute inset-0 flex flex-col items-center justify-center">
									<div class="w-20 h-20 bg-gray-700 rounded-full flex items-center justify-center mb-3">
										<span class="text-2xl font-bold text-gray-400">{$auth.user?.display_name?.charAt(0) || '?'}</span>
									</div>
									<p class="text-gray-400 text-sm mb-3">آماده پیوستن</p>
									<button onclick={joinInlineRoom} class="px-5 py-2.5 bg-blue-600 text-white rounded-xl text-sm font-medium hover:bg-blue-700 transition-colors">پیوستن</button>
								</div>
							{/if}
							<div bind:this={remoteContainer} class="absolute inset-0 {gridCols} gap-2 p-2 auto-rows-fr"></div>
							<div class="absolute bottom-3 left-3 w-36 h-28 rounded-lg overflow-hidden border-2 bg-gray-800 {videoEnabled ? 'border-gray-600' : 'border-transparent hidden'}">
								<video bind:this={localVideoEl} autoplay muted playsinline class="w-full h-full object-cover"></video>
							</div>
							{#if whiteboardOpen}
								<div class="absolute inset-0 bg-white">
									<Whiteboard {sessionId} />
								</div>
							{/if}
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
						onLeave={() => { disconnectInline(); history.back(); }}
					/>
				</div>
			{/if}
		</div>
	{:else}
		<div class="flex-1 flex items-center justify-center py-20"><p class="text-gray-400">جلسه یافت نشد</p></div>
	{/if}
</div>
