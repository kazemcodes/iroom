<script lang="ts">
	// @ts-nocheck
	import { page } from '$app/state';
	import { auth } from '$lib/stores';
	import { api } from '$lib/api';
	import { onMount, onDestroy } from 'svelte';
	import { JanusClient } from '$lib/classroom/janus-client';
	import { formatDuration } from '$lib/utils/persian';
	import { classroomWindow } from '$lib/classroom/ClassroomWindow';
	import type { Participant, ChatMessage } from '$lib/classroom/types';

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

	let showUsers = $state(true);
	let showChat = $state(true);
	let showAppMenu = $state(false);
	let showChatMenu = $state(false);
	let showUsersMenu = $state(false);

	let participants = $state<Participant[]>([]);
	let chatMessages = $state<ChatMessage[]>([]);
	let localVideoEl: HTMLVideoElement;
	let remoteContainer: HTMLDivElement;
	let chatWs: WebSocket | null = null;
	let chatInput = $state('');
	let emojiPickerOpen = $state(false);
	let replyTo = $state<ChatMessage | null>(null);
	let editingMessage = $state<ChatMessage | null>(null);
	let editContent = $state('');
	let contextMenu = $state<{ show: boolean; x: number; y: number; message: ChatMessage | null }>({ show: false, x: 0, y: 0, message: null });

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

	const formattedTime = $derived.by(() => {
		const m = Math.floor(elapsedSeconds / 60);
		const s = elapsedSeconds % 60;
		return `${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`;
	});

	onMount(async () => { loading = true; const res = await api.get(`/sessions/${sessionId}`); if (res.success) session = res.data!; loading = false; });
	onDestroy(() => { stopRecording(); disconnectInline(); });

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
					chatMessages = [...chatMessages, { id: String(Date.now()), sender: isOwn ? 'شما' : (msg.user_display_name || 'کاربر'), content: msg.content, time: new Date(msg.created_at).toLocaleTimeString('fa-IR', { hour: '2-digit', minute: '2-digit' }), isOwn }];
				}
			} catch (e) {}
		};
		chatWs.onclose = () => { if (connected) setTimeout(connectChatWs, 3000); };
	}

	async function joinRoom() {
		const joinRes = await api.get(`/sessions/${sessionId}/classroom`);
		if (!joinRes.success || !joinRes.data) { alert(joinRes.error || 'خطا'); return; }
		const { ws_url, room_id, user_id, role } = joinRes.data;
		try {
			janus = new JanusClient({ wsUrl: ws_url, roomId: room_id, userId: user_id, role, displayName: $auth.user?.display_name || 'کاربر' });
			await janus.connect();
			connected = true;
			elapsedSeconds = 0;
			timerInterval = setInterval(() => { elapsedSeconds++; }, 1000);
			janus.onParticipantJoined = (p) => { participants = [...participants, { id: String(p.id), name: p.display || 'ناشناس', role: 'user', isSpeaking: false, hasVideo: p.video, hasAudio: p.audio, hasScreen: false, hasWhiteboard: false, handRaised: false }]; };
			janus.onStream = (stream) => { if (remoteContainer) { const v = document.createElement('video'); v.autoplay = true; v.playsInline = true; v.className = 'w-full h-full object-cover rounded'; v.srcObject = stream; remoteContainer.appendChild(v); } };
			await janus.joinRoom();
			connectChatWs();
		} catch (err) { connected = false; }
	}

	function toggleMic() { micOn = !micOn; janus?.toggleAudio(); }
	function toggleWebcam() { webcamOn = !webcamOn; janus?.toggleVideo(); }
	function startRecording() { const s: MediaStream[] = []; if (localVideoEl?.srcObject) s.push(localVideoEl.srcObject as MediaStream); if (!s.length) { alert('ابتدا صدا یا ویدیو را فعال کنید'); return; } const c = new MediaStream(s.flatMap(x => x.getTracks())); try { mediaRecorder = new MediaRecorder(c, { mimeType: 'video/webm;codecs=vp9' }); } catch { mediaRecorder = new MediaRecorder(c); } recordingChunks = []; mediaRecorder.ondataavailable = (e) => { if (e.data.size > 0) recordingChunks.push(e.data); }; mediaRecorder.onstop = async () => { const b = new Blob(recordingChunks, { type: 'video/webm' }); const d = Math.floor((Date.now() - recordingStartTime) / 1000); const fd = new FormData(); fd.append('file', b, `recording-${Date.now()}.webm`); fd.append('duration', String(d)); const t = localStorage.getItem('access_token'); await fetch(`${api.getBaseUrl()}/sessions/${sessionId}/recordings`, { method: 'POST', headers: { 'Authorization': `Bearer ${t}` }, body: fd }); }; mediaRecorder.start(1000); isRecording = true; recordingStartTime = Date.now(); }
	function stopRecording() { if (mediaRecorder && mediaRecorder.state !== 'inactive') mediaRecorder.stop(); isRecording = false; }
	function sendChatMessage() { if (!chatInput.trim()) return; const c = chatInput.trim(); chatInput = ''; if (chatWs?.readyState === WebSocket.OPEN) chatWs.send(JSON.stringify({ type: 'message', content: c })); else api.post(`/sessions/${sessionId}/messages`, { content: c }); chatMessages = [...chatMessages, { id: String(Date.now()), sender: 'شما', content: c, isOwn: true, time: new Date().toLocaleTimeString('fa-IR', { hour: '2-digit', minute: '2-digit' }) }]; }
	function disconnectInline() { connected = false; chatWs?.close(); stopRecording(); if (timerInterval) { clearInterval(timerInterval); timerInterval = null; } if (localVideoEl?.srcObject) { (localVideoEl.srcObject as MediaStream).getTracks().forEach(t => t.stop()); localVideoEl.srcObject = null; } janus?.leave(); janus = null; }
	function leaveRoom() { classroomWindow.close(sessionId); disconnectInline(); window.close(); }
	function handleContextMenu(e: MouseEvent, msg: ChatMessage) { e.preventDefault(); contextMenu = { show: true, x: e.clientX, y: e.clientY, message: msg }; }
	function closeContextMenu() { contextMenu.show = false; showAppMenu = false; showChatMenu = false; showUsersMenu = false; }

	let mediaRecorder = $state<MediaRecorder | null>(null);
	let recordingChunks = $state<Blob[]>([]);
	let recordingStartTime = $state(0);
</script>

<svelte:head>
	<style>
		:global(body) { margin: 0; padding: 0; overflow: hidden; background: #1e2028; font-family: 'Vazirmatn', system-ui, sans-serif; }
		:global(.sr-room) { display: flex; flex-direction: column; height: 100vh; color: #e0e0e6; }
		:global(.sr-header) { display: flex; align-items: center; justify-content: space-between; padding: 0 16px; height: 48px; flex-shrink: 0; background: #272b35; }
		:global(.sr-header .room-name) { display: flex; align-items: center; gap: 8px; }
		:global(.sr-header .room-name img) { width: 28px; height: 28px; border-radius: 50%; }
		:global(.sr-header .room-name a) { color: #e0e0e6; text-decoration: none; font-size: 14px; font-weight: 500; }
		:global(.sr-header .room-name span) { color: #888; }
		:global(.sr-header .room-timer) { display: flex; align-items: center; gap: 6px; font-size: 13px; color: #888; font-family: monospace; }
		:global(.sr-workspace) { display: flex; flex-direction: column; flex: 1; min-height: 0; }
		:global(.sr-room-nav) { display: flex; align-items: center; padding: 8px 12px; gap: 8px; flex-shrink: 0; }
		:global(.sr-room-nav .spacer) { flex: 1; }
		:global(.sr-layout) { display: flex; flex: 1; min-height: 0; }
		:global(.sr-sidebar) { display: flex; flex-direction: column; flex-shrink: 0; width: 280px; background: #272b35; overflow-y: auto; }
		:global(.sr-mainbar) { display: flex; flex-direction: column; flex: 1; min-width: 0; background: #1e2028; position: relative; }
		:global(.sr-icon-btn) { display: inline-flex; align-items: center; justify-content: center; width: 40px; height: 40px; border-radius: 50%; border: 1px solid transparent; background: transparent; cursor: pointer; padding: 0; transition: all 0.15s; }
		:global(.sr-icon-btn:hover) { background: #2b303b; }
		:global(.sr-icon-btn.btn-on) { border-color: #23b9d7; }
		:global(.sr-icon-btn.btn-on svg) { fill: #23b9d7; }
		:global(.sr-icon-btn svg) { width: 24px; height: 24px; fill: #888; }
		:global(.sr-icon-btn.sm) { width: 28px; height: 28px; }
		:global(.sr-icon-btn.sm svg) { width: 18px; height: 18px; }
		:global(.sr-block) { display: flex; flex-direction: column; min-width: 200px; }
		:global(.sr-block-header) { display: flex; align-items: center; padding: 6px 10px; gap: 6px; border-bottom: 1px solid #3a3e47; flex-shrink: 0; }
		:global(.sr-block-header .block-title) { flex: 1; font-size: 12px; color: #888; line-height: 2; }
		:global(.sr-block-header .block-icon) { padding: 2px; }
		:global(.sr-block-header .block-icon svg) { width: 18px; height: 18px; fill: #888; }
		:global(.sr-block-header .block-menu) { padding: 2px; }
		:global(.sr-block-content) { flex: 1; min-height: 0; overflow-y: auto; }
		:global(.sr-user-row) { display: flex; align-items: center; padding: 4px 10px; gap: 8px; }
		:global(.sr-user-row:hover) { background: #2b303b; border-radius: 6px; }
		:global(.sr-user-row .user-icon) { width: 22px; min-width: 22px; }
		:global(.sr-user-row .user-icon svg) { width: 22px; height: 22px; fill: #888; }
		:global(.sr-user-row .user-nickname) { flex: 1; font-size: 13px; color: #e0e0e6; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
		:global(.sr-user-row .user-action) { display: flex; gap: 2px; }
		:global(.sr-user-row .user-action button) { background: none; border: none; padding: 2px; cursor: pointer; border-radius: 4px; }
		:global(.sr-user-row .user-action button:hover) { background: #3a3e47; }
		:global(.sr-user-row .user-action button svg) { width: 18px; height: 18px; fill: #888; }
		:global(.sr-user-row .user-action button.btn-on svg) { fill: #23b9d7; }
		:global(.sr-chat-msgs) { flex: 1; overflow-y: auto; padding: 12px; display: flex; flex-direction: column; gap: 8px; }
		:global(.sr-chat-msg) { max-width: 85%; padding: 8px 12px; border-radius: 12px; font-size: 13px; line-height: 1.4; }
		:global(.sr-chat-msg.other) { background: #2b303b; color: #e0e0e6; align-self: flex-start; }
		:global(.sr-chat-msg.own) { background: #23b9d7; color: #fff; align-self: flex-end; }
		:global(.sr-chat-msg .msg-sender) { font-size: 11px; color: #23b9d7; font-weight: 600; margin-bottom: 2px; }
		:global(.sr-chat-msg .msg-time) { font-size: 10px; color: #888; margin-top: 4px; }
		:global(.sr-chat-msg.own .msg-time) { color: rgba(255,255,255,0.7); }
		:global(.sr-chat-input) { display: flex; align-items: center; padding: 8px 10px; border-top: 1px solid #3a3e47; gap: 4px; }
		:global(.sr-chat-input .input) { flex: 1; background: transparent; border: none; color: #e0e0e6; font-size: 13px; outline: none; padding: 4px 8px; font-family: 'Vazirmatn', system-ui, sans-serif; }
		:global(.sr-chat-input .input::placeholder) { color: #888; }
		:global(.sr-toolbar) { position: absolute; bottom: 16px; left: 50%; transform: translateX(-50%); display: flex; align-items: center; gap: 8px; padding: 8px 16px; background: rgba(39, 43, 53, 0.95); border-radius: 24px; backdrop-filter: blur(8px); z-index: 10; }
		:global(.sr-toolbar .divider) { width: 1px; height: 24px; background: #3a3e47; margin: 0 4px; }
		:global(.sr-menu-items) { background: #272b35; border: 1px solid #4b4e58; border-radius: 8px; min-width: 180px; padding: 5px; list-style: none; position: absolute; z-index: 100; box-shadow: 0 4px 20px rgba(0,0,0,0.3); }
		:global(.sr-menu-items li) { display: flex; align-items: center; gap: 8px; padding: 8px 12px; border-radius: 6px; font-size: 13px; color: #e0e0e6; cursor: pointer; }
		:global(.sr-menu-items li:hover) { background: #4b4e58; color: #fff; }
		:global(.sr-menu-items li svg) { width: 18px; height: 18px; fill: #888; }
		:global(.sr-menu-items li:hover svg) { fill: #fff; }
		:global(.sr-menu-items .separator) { border-top: 1px solid #4b4e58; height: 0; margin: 2px 0; padding: 0; }
		:global(.sr-menu-items li.checked .checked-icon) { display: inline; margin-left: auto; }
		:global(.sr-menu-items li .checked-icon svg) { fill: #23b9d7; }
		:global(.sr-emoji-list) { display: grid; grid-template-columns: repeat(9, 1fr); gap: 2px; padding: 8px; background: #272b35; border: 1px solid #4b4e58; border-radius: 8px; position: absolute; bottom: 100%; left: 0; margin-bottom: 4px; }
		:global(.sr-emoji-list button) { width: 28px; height: 28px; display: flex; align-items: center; justify-content: center; font-size: 18px; border: none; background: transparent; border-radius: 4px; cursor: pointer; }
		:global(.sr-emoji-list button:hover) { background: #4b4e58; }
		:global(.sr-trademark) { text-align: center; padding: 4px; }
		:global(.sr-trademark a) { color: #888; font-size: 11px; text-decoration: none; }
		:global(.sr-msg-context) { position: fixed; z-index: 200; background: #272b35; border: 1px solid #4b4e58; border-radius: 8px; padding: 4px; min-width: 160px; }
		:global(.sr-msg-context li) { display: flex; align-items: center; gap: 8px; padding: 6px 12px; border-radius: 6px; font-size: 13px; color: #e0e0e6; cursor: pointer; }
		:global(.sr-msg-context li:hover) { background: #4b4e58; }
		:global(.sr-msg-context li svg) { width: 16px; height: 16px; fill: #888; }
	</style>
</svelte:head>

<svelte:window onclick={closeContextMenu} />

{#if loading}
	<div style="display:flex;align-items:center;justify-content:center;height:100vh;background:#1e2028;">
		<div style="width:32px;height:32px;border:3px solid #3a3e47;border-top-color:#23b9d7;border-radius:50%;animation:spin .8s linear infinite;"></div>
	</div>
	<style>@keyframes spin{to{transform:rotate(360deg)}}</style>
{:else if session}
<div class="sr-room">
	<!-- Header -->
	<header class="sr-header">
		<div class="room-name">
			<img src="./اسکای_روم - 1_files/avatar.png" alt="owner" />
			<a href="javascript:void(0)">{$auth.user?.display_name || 'مالک'}</a>
			<span>:</span>
			<span>{session.title}</span>
		</div>
		<div class="room-timer" title="زمان سپری شده">
			<span>{formattedTime}</span>
			<svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><circle cx="12" cy="12" r="10"/><path d="M12 6v6l4 2"/></svg>
		</div>
	</header>

	<!-- Workspace -->
	<div class="sr-workspace">
		<!-- Room Nav -->
		<div class="sr-room-nav">
			<!-- App Menu -->
			<div style="position:relative;">
				<button class="sr-icon-btn" onclick={() => { showAppMenu = !showAppMenu; closeContextMenu(); }} title="منو">
					<svg viewBox="0 0 24 24" fill="currentColor"><path d="M3 18h18v-2H3v2zm0-5h18v-2H3v2zm0-7v2h18V6H3z"/></svg>
				</button>
				{#if showAppMenu}
					<ul class="sr-menu-items" style="top:100%;left:0;margin-top:4px;">
						<li><svg viewBox="0 0 24 24" fill="currentColor"><circle cx="12" cy="12" r="10" fill="none" stroke="currentColor" stroke-width="1.5"/><path d="M12 16v-4m0-4h.01" fill="none" stroke="currentColor" stroke-width="1.5"/></svg> اطلاعات کاربری</li>
						<li><svg viewBox="0 0 24 24" fill="currentColor"><path d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" fill="none" stroke="currentColor" stroke-width="1.5"/></svg> وضعیت اتصال</li>
						<li><svg viewBox="0 0 24 24" fill="currentColor"><path d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065zM15 12a3 3 0 11-6 0 3 3 0 016 0z" fill="none" stroke="currentColor" stroke-width="1.5"/></svg> تنظیمات</li>
						<li><svg viewBox="0 0 24 24" fill="currentColor"><path d="M4 5a1 1 0 011-1h14a1 1 0 011 1v2a1 1 0 01-1 1H5a1 1 0 01-1-1V5zM4 13a1 1 0 011-1h6a1 1 0 011 1v6a1 1 0 01-1 1H5a1 1 0 01-1-1v-6zM16 13a1 1 0 011-1h2a1 1 0 011 1v6a1 1 0 01-1 1h-2a1 1 0 01-1-1v-6z" fill="none" stroke="currentColor" stroke-width="1.5"/></svg> چیدمان <svg viewBox="0 0 24 24" fill="currentColor" style="margin-left:auto;width:14px;height:14px;"><path d="M9 5l7 7-7 7" fill="none" stroke="currentColor" stroke-width="2"/></svg></li>
						<div class="separator"></div>
						<li onclick={leaveRoom}><svg viewBox="0 0 24 24" fill="currentColor"><path d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" fill="none" stroke="currentColor" stroke-width="1.5"/></svg> خروج</li>
						{#if isOwner}
							<li style="color:#e05252;"><svg viewBox="0 0 24 24" fill="currentColor"><path d="M5.586 15H4a1 1 0 01-1-1v-4a1 1 0 011-1h1.586l4.707-4.707C10.923 3.663 12 4.109 12 5v14c0 .891-1.077 1.337-1.707.707L5.586 15z" fill="none" stroke="currentColor" stroke-width="1.5"/><path d="M17 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2" fill="none" stroke="currentColor" stroke-width="1.5"/></svg> بستن اتاق</li>
						{/if}
					</ul>
				{/if}
			</div>

			<!-- Mini Toolbar -->
			<button class="sr-icon-btn {showUsers ? 'btn-on' : ''}" title="کاربران" onclick={() => { showUsers = !showUsers; }}>
				<svg viewBox="0 0 24 24" fill="currentColor"><path d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z" fill="none" stroke="currentColor" stroke-width="1.5"/></svg>
			</button>
			<button class="sr-icon-btn {showChat ? 'btn-on' : ''}" title="پیام‌ها" onclick={() => { showChat = !showChat; }}>
				<svg viewBox="0 0 24 24" fill="currentColor"><path d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" fill="none" stroke="currentColor" stroke-width="1.5"/></svg>
			</button>

			<div class="spacer"></div>

			<!-- Toolbar -->
			<button class="sr-icon-btn {audioOn ? 'btn-on' : ''}" title="خروجی صدا" onclick={() => audioOn = !audioOn}>
				{#if audioOn}<svg viewBox="0 0 24 24" fill="currentColor"><path d="M15.536 8.464a5 5 0 010 7.072m2.828-9.9a9 9 0 010 12.728M5.586 15H4a1 1 0 01-1-1v-4a1 1 0 011-1h1.586l4.707-4.707C10.923 3.663 12 4.109 12 5v14c0 .891-1.077 1.337-1.707.707L5.586 15z" fill="none" stroke="currentColor" stroke-width="1.5"/></svg>{:else}<svg viewBox="0 0 24 24" fill="currentColor"><path d="M5.586 15H4a1 1 0 01-1-1v-4a1 1 0 011-1h1.586l4.707-4.707C10.923 3.663 12 4.109 12 5v14c0 .891-1.077 1.337-1.707.707L5.586 15z" fill="none" stroke="currentColor" stroke-width="1.5"/><path d="M17 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2" fill="none" stroke="currentColor" stroke-width="1.5"/></svg>{/if}
			</button>
			<button class="sr-icon-btn {micOn ? 'btn-on' : ''}" title="میکروفون" onclick={toggleMic}>
				{#if micOn}<svg viewBox="0 0 24 24" fill="currentColor"><path d="M19 11a7 7 0 01-7 7m0 0a7 7 0 01-7-7m7 7v4m0 0H8m4 0h4m-4-8a3 3 0 01-3-3V5a3 3 0 116 0v6a3 3 0 01-3 3z" fill="none" stroke="currentColor" stroke-width="1.5"/></svg>{:else}<svg viewBox="0 0 24 24" fill="currentColor"><path d="M5.586 15H4a1 1 0 01-1-1v-4a1 1 0 011-1h1.586l4.707-4.707C10.923 3.663 12 4.109 12 5v14c0 .891-1.077 1.337-1.707.707L5.586 15z" fill="none" stroke="currentColor" stroke-width="1.5"/><path d="M17 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2" fill="none" stroke="currentColor" stroke-width="1.5"/></svg>{/if}
			</button>
			<button class="sr-icon-btn {webcamOn ? 'btn-on' : ''}" title="وبکم" onclick={toggleWebcam}>
				{#if webcamOn}<svg viewBox="0 0 24 24" fill="currentColor"><path d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" fill="none" stroke="currentColor" stroke-width="1.5"/></svg>{:else}<svg viewBox="0 0 24 24" fill="currentColor"><path d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" fill="none" stroke="currentColor" stroke-width="1.5"/><path d="M17 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2" fill="none" stroke="currentColor" stroke-width="1.5"/></svg>{/if}
			</button>
			<button class="sr-icon-btn {screenShareOn ? 'btn-on' : ''}" title="اشتراک‌گذاری صفحه" onclick={() => screenShareOn = !screenShareOn}>
				<svg viewBox="0 0 24 24" fill="currentColor"><path d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" fill="none" stroke="currentColor" stroke-width="1.5"/></svg>
			</button>
			<button class="sr-icon-btn" title="تخته" onclick={() => whiteboardOn = !whiteboardOn}>
				<svg viewBox="0 0 24 24" fill="currentColor"><path d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" fill="none" stroke="currentColor" stroke-width="1.5"/></svg>
			</button>
			<button class="sr-icon-btn" title="فایل‌ها" onclick={() => filesOn = !filesOn}>
				<svg viewBox="0 0 24 24" fill="currentColor"><path d="M7 4v16M17 4v16M3 8h4m10 0h4M3 12h18M3 16h4m10 0h4M4 20h16a1 1 0 001-1V5a1 1 0 00-1-1H4a1 1 0 00-1 1v14a1 1 0 001 1z" fill="none" stroke="currentColor" stroke-width="1.5"/></svg>
			</button>
			<button class="sr-icon-btn {handRaised ? 'btn-on' : ''}" title="بالا بردن دست" onclick={() => handRaised = !handRaised}>
				<svg viewBox="0 0 24 24" fill="currentColor"><path d="M7 11.5V14m0-2.5v-6a1.5 1.5 0 113 0m-3 6a1.5 1.5 0 00-3 0v2a7.5 7.5 0 0015 0v-5a1.5 1.5 0 00-3 0m-6-3V11m0-5.5v-1a1.5 1.5 0 013 0v1m0 0V11m0-5.5a1.5 1.5 0 013 0v3m0 0V11" fill="none" stroke="currentColor" stroke-width="1.5"/></svg>
			</button>
		</div>

		<!-- Layout: Sidebar + Mainbar -->
		<div class="sr-layout">
			<!-- Sidebar -->
			{#if showUsers || showChat}
			<aside class="sr-sidebar">
				{#if showUsers}
				<!-- Users Block -->
				<div class="sr-block" style="flex:1;min-height:0;">
					<div class="sr-block-header">
						<div style="position:relative;">
							<button class="sr-icon-btn sm" onclick={() => showUsersMenu = !showUsersMenu}>
								<svg viewBox="0 0 24 24" fill="currentColor"><circle cx="12" cy="5" r="1.5"/><circle cx="12" cy="12" r="1.5"/><circle cx="12" cy="19" r="1.5"/></svg>
							</button>
							{#if showUsersMenu}
								<ul class="sr-menu-items" style="top:100%;right:0;margin-top:4px;">
									<li>نمایش کاربران</li>
									<li>پایین آوردن دست‌ها</li>
									<li>رفع مسدودی همه</li>
									<li>حضور و غیاب</li>
									<li>دریافت لیست حاضرین</li>
									<div class="separator"></div>
									<li onclick={() => showUsersMenu = false}>بستن</li>
								</ul>
							{/if}
						</div>
						<div class="block-title">کاربران</div>
						{#if participants.filter(p => p.handRaised).length > 0}
							<span style="font-size:11px;color:#d7911d;">✋ {participants.filter(p => p.handRaised).length}</span>
						{/if}
						<span style="font-size:11px;color:#888;">{participants.length}</span>
						<div class="block-icon">
							<svg viewBox="0 0 24 24" fill="currentColor" width="18" height="18"><path d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z" fill="none" stroke="currentColor" stroke-width="1.5"/></svg>
						</div>
					</div>
					<div class="sr-block-content" style="padding:7px;">
						{#each participants as p (p.id)}
							<div class="sr-user-row">
								<div class="user-icon">
									<svg viewBox="0 0 24 24" fill="currentColor"><path d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" fill="none" stroke="currentColor" stroke-width="1.5"/></svg>
								</div>
								<div class="user-nickname" title={p.name}>{p.name}{p.isLocal ? ' (شما)' : ''}</div>
								<div class="user-action">
									{#if isTeacherOrAdmin && !p.isLocal}
										<button class="{p.hasAudio ? 'btn-on' : ''}" title="میکروفون"><svg viewBox="0 0 24 24" fill="currentColor"><path d="M19 11a7 7 0 01-7 7m0 0a7 7 0 01-7-7m7 7v4m0 0H8m4 0h4m-4-8a3 3 0 01-3-3V5a3 3 0 116 0v6a3 3 0 01-3 3z" fill="none" stroke="currentColor" stroke-width="1.5"/></svg></button>
										<button class="{p.hasVideo ? 'btn-on' : ''}" title="وبکم"><svg viewBox="0 0 24 24" fill="currentColor"><path d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" fill="none" stroke="currentColor" stroke-width="1.5"/></svg></button>
										<button title="صفحه"><svg viewBox="0 0 24 24" fill="currentColor"><path d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" fill="none" stroke="currentColor" stroke-width="1.5"/></svg></button>
										<button title="تخته"><svg viewBox="0 0 24 24" fill="currentColor"><path d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" fill="none" stroke="currentColor" stroke-width="1.5"/></svg></button>
										<button title="دست"><svg viewBox="0 0 24 24" fill="currentColor"><path d="M7 11.5V14m0-2.5v-6a1.5 1.5 0 113 0m-3 6a1.5 1.5 0 00-3 0v2a7.5 7.5 0 0015 0v-5a1.5 1.5 0 00-3 0m-6-3V11m0-5.5v-1a1.5 1.5 0 013 0v1m0 0V11m0-5.5a1.5 1.5 0 013 0v3m0 0V11" fill="none" stroke="currentColor" stroke-width="1.5"/></svg></button>
									{/if}
								</div>
							</div>
						{/each}
						{#if participants.length === 0}
							<p style="text-align:center;color:#888;font-size:12px;padding:16px;">هنوز کسی متصل نیست</p>
						{/if}
					</div>
				</div>
				{/if}

				{#if showChat}
				<!-- Chat Block -->
				<div class="sr-block" style="flex-shrink:0;max-height:50%;">
					<div class="sr-block-header">
						<div style="position:relative;">
							<button class="sr-icon-btn sm" onclick={() => showChatMenu = !showChatMenu}>
								<svg viewBox="0 0 24 24" fill="currentColor"><circle cx="12" cy="5" r="1.5"/><circle cx="12" cy="12" r="1.5"/><circle cx="12" cy="19" r="1.5"/></svg>
							</button>
							{#if showChatMenu}
								<ul class="sr-menu-items" style="bottom:100%;right:0;margin-bottom:4px;">
									<li>نمایش بزرگتر</li>
									<li>غیر فعال سازی چت</li>
									<li>حالت خصوصی</li>
									<li>پاک کردن همه پیام‌ها</li>
									<li>ذخیره تمام پیام‌ها</li>
									<div class="separator"></div>
									<li onclick={() => showChatMenu = false}>بستن</li>
								</ul>
							{/if}
						</div>
						<div class="block-title">پیام‌ها</div>
						<div class="block-icon">
							<svg viewBox="0 0 24 24" fill="currentColor" width="18" height="18"><path d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" fill="none" stroke="currentColor" stroke-width="1.5"/></svg>
						</div>
					</div>
					<div class="sr-chat-msgs">
						{#each chatMessages as msg (msg.id)}
							<div class="sr-chat-msg {msg.isOwn ? 'own' : 'other'}" oncontextmenu={(e) => handleContextMenu(e, msg)}>
								{#if !msg.isOwn}<div class="msg-sender">{msg.sender}</div>{/if}
								<div>{msg.content}</div>
								<div class="msg-time">{msg.time}</div>
							</div>
						{/each}
					</div>
					<div class="sr-chat-input" style="position:relative;">
						{#if emojiPickerOpen}
							<div class="sr-emoji-list">
								{#each ['😃','😄','😊','😁','😂','😅','😉','😜','😍','😘','😏','😒','😞','😩','😢','😭','😤','😡','😲','😨','😱','🙏','👍','👎','👏','👋','👌','✌️','❤️','🌹'] as emoji}
									<button onclick={() => { chatInput += emoji; emojiPickerOpen = false; }}>{emoji}</button>
								{/each}
							</div>
						{/if}
						<input class="input" bind:value={chatInput} placeholder="پیام خود را وارد کنید" onkeydown={(e) => { if (e.key === 'Enter') sendChatMessage(); }} />
						<button class="sr-icon-btn sm" onclick={() => emojiPickerOpen = !emojiPickerOpen}>
							<svg viewBox="0 0 24 24" fill="currentColor"><path d="M14.828 14.828a4 4 0 01-5.656 0M9 10h.01M15 10h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" fill="none" stroke="currentColor" stroke-width="1.5"/></svg>
						</button>
						<button class="sr-icon-btn sm" onclick={sendChatMessage}>
							<svg viewBox="0 0 24 24" fill="currentColor"><path d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8" fill="none" stroke="currentColor" stroke-width="1.5"/></svg>
						</button>
					</div>
				</div>
				{/if}
			</aside>
			{/if}

			<!-- Mainbar -->
			<div class="sr-mainbar">
				{#if !connected}
					<div style="position:absolute;inset:0;display:flex;flex-direction:column;align-items:center;justify-content:center;">
						<div style="width:80px;height:80px;border-radius:50%;background:#272b35;display:flex;align-items:center;justify-content:center;margin-bottom:12px;">
							<span style="font-size:32px;font-weight:700;color:#888;">{$auth.user?.display_name?.charAt(0) || '?'}</span>
						</div>
						<p style="color:#888;font-size:13px;margin-bottom:12px;">آماده پیوستن</p>
						<button onclick={joinRoom} style="padding:10px 24px;background:#23b9d7;color:#fff;border:none;border-radius:8px;font-size:13px;font-weight:600;cursor:pointer;font-family:'Vazirmatn',system-ui,sans-serif;">پیوستن</button>
					</div>
				{/if}
				<div bind:this={remoteContainer} class="absolute inset-0 {gridCols} gap-1 p-1 auto-rows-fr"></div>
				<div style="position:absolute;bottom:80px;left:12px;width:144px;height:112px;border-radius:4px;overflow:hidden;{webcamOn ? 'border:1px solid #3a3e47;' : 'display:none;'}background:#0a0a1a;">
					<video bind:this={localVideoEl} autoplay muted playsinline style="width:100%;height:100%;object-fit:cover;"></video>
				</div>
			</div>
		</div>
	</div>

	<!-- Trademark -->
	<div class="sr-trademark">
		<a href="https://www.skyroom.online/" target="_blank">© آی‌روم</a>
	</div>
</div>

<!-- Context Menu -->
{#if contextMenu.show && contextMenu.message}
	<div class="sr-msg-context" style="left:{contextMenu.x}px;top:{contextMenu.y}px;">
		<li onclick={() => { replyTo = contextMenu.message!; contextMenu.show = false; }}>↩ پاسخ</li>
		{#if contextMenu.message.isOwn}
			<li onclick={() => { editingMessage = contextMenu.message!; editContent = contextMenu.message!.content; contextMenu.show = false; }}>✏️ ویرایش</li>
			<li style="color:#e05252;" onclick={() => { chatMessages = chatMessages.filter(m => m.id !== contextMenu.message!.id); contextMenu.show = false; }}>🗑️ حذف</li>
		{/if}
		{#if isTeacherOrAdmin}
			<li onclick={() => { chatMessages = chatMessages.map(m => m.id === contextMenu.message!.id ? { ...m, isPinned: !m.isPinned } : m); contextMenu.show = false; }}>📌 {contextMenu.message.isPinned ? 'برداشتن سنجاق' : 'سنجاق کردن'}</li>
		{/if}
	</div>
{/if}

{:else}
	<div style="display:flex;align-items:center;justify-content:center;height:100vh;background:#1e2028;"><p style="color:#888;">جلسه یافت نشد</p></div>
{/if}
