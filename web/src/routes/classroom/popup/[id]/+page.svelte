<script lang="ts">
	import { page } from '$app/state';
	import { auth } from '$lib/stores';
	import { api } from '$lib/api';
	import { onMount, onDestroy } from 'svelte';
	import { PionClient } from '$lib/classroom/pion-client';
	import type { UserRole, Participant, ChatMessage } from '$lib/classroom/types';
	import { ROLE_PERMISSIONS } from '$lib/classroom/types';
	import ChatPanel from '$lib/components/classroom/ChatPanel.svelte';
	import UsersPanel from '$lib/components/classroom/UsersPanel.svelte';
	import AppMenu from '$lib/components/classroom/AppMenu.svelte';
	import UserInfoModal from '$lib/components/classroom/UserInfoModal.svelte';
	import ConnectionStatusModal from '$lib/components/classroom/ConnectionStatusModal.svelte';
	import SettingsModal from '$lib/components/classroom/SettingsModal.svelte';
	import LayoutModal from '$lib/components/classroom/LayoutModal.svelte';

	let session = $state<any>(null);
	let loading = $state(true);

	let pion = $state<PionClient | null>(null);
	let connected = $state(false);
	let audioOn = $state(true);
	let micOn = $state(false);
	let webcamOn = $state(false);
	let screenShareOn = $state(false);
	let handRaised = $state(false);
	let elapsedSeconds = $state(0);
	let timerInterval: ReturnType<typeof setInterval> | null = null;

	let showUsersPanel = $state(true);
	let showChatPanel = $state(true);
	let showAppMenu = $state(false);
	let showUsersMenu = $state(false);
	let showChatMenu = $state(false);
	let showModal = $state<'userInfo' | 'connection' | 'settings' | 'layout' | null>(null);
	let joinNotification = $state<{ name: string; show: boolean }>({ name: '', show: false });
	let participants = $state<Participant[]>([]);
	let chatMessages = $state<ChatMessage[]>([]);
	let localVideoEl: HTMLVideoElement;
	let remoteContainer: HTMLDivElement;
	let chatWs: WebSocket | null = null;

	const sessionId = $derived(page.params.id!);
	const currentUserRole = $derived(($auth.user?.role || 'student') as UserRole);
	const perms = $derived(ROLE_PERMISSIONS[currentUserRole] || ROLE_PERMISSIONS.student);
	const isPresenterOrAbove = $derived(['owner', 'admin', 'operator', 'teacher', 'presenter'].includes(currentUserRole));

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

	onMount(async () => {
		const res = await api.get(`/sessions/${sessionId}`);
		if (res.success) session = res.data!;
		loading = false;
		if (session?.status === 'live') {
			startTimer();
		}
		connectChatWs();
		fetchParticipants();
	});
	onDestroy(() => { disconnect(); if (timerInterval) clearInterval(timerInterval); });

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

		const { room_id, user_id, role } = joinRes.data as { room_id: string; user_id: string; role: string };
		try {
			pion = new PionClient({
				roomId: String(room_id), userId: String(user_id), role,
				displayName: $auth.user?.display_name || 'کاربر',
			});
			pion.onLocalStream = (stream) => {
				if (localVideoEl) localVideoEl.srcObject = stream;
			};
			pion.onRemoteStream = (stream, participantId) => {
				if (remoteContainer) {
					const el = document.createElement('video');
					el.id = `track-${participantId}`;
					el.autoplay = true;
					el.playsInline = true;
					el.className = 'w-full h-full object-cover rounded-lg';
					remoteContainer.appendChild(el);
					el.srcObject = stream;
				}
			};
			await pion.connect();
			connected = true;
			startTimer();
			startParticipantRefresh();
			showJoinNotification($auth.user?.display_name || 'کاربر');
		} catch (e: any) {
			console.error('Join failed:', e);
			alert('خطا در اتصال به اتاق: ' + (e?.message || e));
		}
	}

	function disconnect() {
		if (pion) { pion.disconnect(); pion = null; }
		connected = false;
		if (timerInterval) { clearInterval(timerInterval); timerInterval = null; }
		if (participantInterval) { clearInterval(participantInterval); participantInterval = null; }
	}

	function startTimer() {
		timerInterval = setInterval(() => { elapsedSeconds++; }, 1000);
	}

	let participantInterval: ReturnType<typeof setInterval> | null = null;

	function startParticipantRefresh() {
		participantInterval = setInterval(() => { fetchParticipants(); }, 5000);
	}

	function toggleMic() {
		if (!perms.canMic) return;
		if (pion) { pion.toggleAudio(); }
		micOn = !micOn;
	}

	function toggleWebcam() {
		if (!perms.canWebcam) return;
		if (pion) { pion.toggleVideo(); }
		webcamOn = !webcamOn;
	}

	function toggleScreenShare() {
		if (!perms.canScreenShare) return;
		if (pion && !screenShareOn) { pion.shareScreen(); }
		screenShareOn = !screenShareOn;
	}

	function toggleHand() {
		if (!perms.canHandRaise) return;
		handRaised = !handRaised;
	}

	function sendChatMessage(text: string) {
		if (!perms.canChat) return;
		if (!chatWs || chatWs.readyState !== WebSocket.OPEN) return;
		chatWs.send(JSON.stringify({ type: 'message', content: text }));
	}

	function leaveRoom() { disconnect(); window.close(); }

	function showJoinNotification(name: string) {
		joinNotification = { name, show: true };
		setTimeout(() => { joinNotification = { name: '', show: false }; }, 3000);
	}

	async function fetchParticipants() {
		if (!session) return;
		try {
			const res = await api.get<any[]>(`/sessions/${sessionId}/classroom/participants`);
			if (res.success && Array.isArray(res.data)) {
				participants = res.data.map((p: any) => ({
					id: p.id,
					name: p.name,
					role: 'student' as UserRole,
					isSpeaking: false,
					hasVideo: false,
					hasAudio: true,
					hasScreen: false,
					hasWhiteboard: false,
					handRaised: false,
				}));
			}
		} catch (e) {}
	}

	async function startSession() {
		if (!session) return;
		const res = await api.post(`/sessions/${sessionId}/start`);
		if (res.success) {
			session = { ...session, status: 'live' };
			startTimer();
		}
	}
</script>

<div class="skyroom-col" style="background-color: var(--bg-color); color: var(--text-color); font-family: var(--font-family); font-size: var(--font-size);">
	{#if loading}
		<div style="display:flex;align-items:center;justify-content:center;flex:1;">
			<div class="animate-spin" style="width:32px;height:32px;border:3px solid #23b9d7;border-top-color:transparent;border-radius:50%;"></div>
		</div>
	{:else if session}
		<!-- Header: Room name + Timer -->
		<header class="skyroom-header">
			<div class="skyroom-row" style="gap:8px;min-width:0;">
				<div style="width:28px;height:28px;border-radius:6px;background:var(--accent);display:flex;align-items:center;justify-content:center;flex-shrink:0;">
					<span style="font-size:12px;font-weight:700;color:#fff;">{($auth.user?.display_name || 'م').charAt(0)}</span>
				</div>
				<span style="font-weight:600;font-size:var(--font-size);white-space:nowrap;overflow:hidden;text-overflow:ellipsis;">{$auth.user?.display_name || 'مالک'}</span>
				<span style="color:var(--inactive);">:</span>
				<span style="font-weight:500;font-size:var(--font-size);white-space:nowrap;overflow:hidden;text-overflow:ellipsis;">{session.title}</span>
			</div>
			<div style="flex:1;"></div>
			<div class="skyroom-room-timer">
				<svg width="16" height="16" style="fill:currentColor;"><use xlink:href="#shape_access_time"></use></svg>
				<span style="font-family:monospace;font-size:var(--font-size-sm);">{formattedTime}</span>
			</div>
		</header>

		<!-- Workspace -->
		<div id="workspace" class="skyroom-col" style="flex:1;overflow:hidden;">
			<!-- Room nav: Right (users, chat, hamburger) | Left (toolbar) -->
			<div class="skyroom-room-nav" style="position:relative;">
				<!-- Right: Hamburger, Chat, Users (first in code = rightmost in RTL) -->
				<div class="skyroom-row" style="flex-shrink:0;gap:4px;">
					<button class="skyroom-icon-square" title="منو"
						onclick={() => showAppMenu = !showAppMenu}
					>
						<svg width="18" height="18"><use xlink:href="#shape_menu"></use></svg>
					</button>
					<button class="skyroom-icon-square" class:active={showChatPanel} title="پیام‌ها"
						onclick={() => showChatPanel = !showChatPanel}
					>
						<svg width="18" height="18"><use xlink:href="#shape_chat"></use></svg>
					</button>
					<button class="skyroom-icon-square" class:active={showUsersPanel} title="کاربران"
						onclick={() => showUsersPanel = !showUsersPanel}
					>
						<svg width="18" height="18"><use xlink:href="#shape_group"></use></svg>
					</button>
				</div>
				<div style="flex:1;"></div>
				<!-- Left: Hand, Files, Whiteboard, Screen, Webcam, Mic, Speaker (last in code = leftmost in RTL) -->
				<div class="skyroom-row" style="flex-shrink:0;gap:4px;">
					{#if perms.canHandRaise}
						<button class="skyroom-icon-square" class:active={handRaised} title="بالا بردن دست"
							onclick={toggleHand}
						>
							<svg width="18" height="18"><use xlink:href="#shape_hand"></use></svg>
						</button>
					{/if}
					<button class="skyroom-icon-square" title="فایل‌ها">
						<svg width="18" height="18"><use xlink:href="#shape_slideshow"></use></svg>
					</button>
					{#if perms.canWhiteboard}
						<button class="skyroom-icon-square" title="تخته">
							<svg width="18" height="18"><use xlink:href="#shape_brush"></use></svg>
						</button>
					{/if}
					{#if perms.canScreenShare}
						<button class="skyroom-icon-square" class:active={screenShareOn} title="اشتراک‌گذاری صفحه"
							onclick={toggleScreenShare}
						>
							<svg width="18" height="18"><use xlink:href="#shape_laptop"></use></svg>
						</button>
					{/if}
					{#if perms.canWebcam}
						<button class="skyroom-icon-square" class:active={webcamOn} title="وبکم"
							onclick={toggleWebcam}
						>
							<svg width="18" height="18"><use xlink:href={webcamOn ? '#shape_videocam' : '#shape_videocamoff'}></use></svg>
						</button>
					{/if}
					{#if perms.canMic}
						<button class="skyroom-icon-square" class:active={micOn} title="میکروفون"
							onclick={toggleMic}
						>
							<svg width="18" height="18"><use xlink:href={micOn ? '#shape_mic' : '#shape_mic_off'}></use></svg>
						</button>
					{/if}
					<button class="skyroom-icon-square" class:active={audioOn} title="خروجی صدا"
						onclick={() => audioOn = !audioOn}
					>
						<svg width="18" height="18"><use xlink:href={audioOn ? '#shape_volume_up' : '#shape_volume_off'}></use></svg>
					</button>
				</div>
			</div>

			<!-- Layout: Sidebar + Mainbar -->
			<div class="skyroom-layout">
				<!-- Sidebar -->
				{#if showUsersPanel || showChatPanel}
					<div class="skyroom-sidebar">
						{#if showUsersPanel}
							<div class="skyroom-block skyroom-users-block">
								<div class="skyroom-block-header">
									<div class="skyroom-block-title">
										<div class="skyroom-block-title-content">کاربران</div>
									</div>
									<span class="skyroom-users-count">{participants.length}</span>
									<div style="position:relative;">
										<button class="skyroom-dots-btn" onclick={(e) => { e.stopPropagation(); showUsersMenu = !showUsersMenu; }}>
											<svg width="16" height="16"><use xlink:href="#shape_more_vert"></use></svg>
										</button>
										{#if showUsersMenu}
											<div class="skyroom-context-menu" onclick={(e) => e.stopPropagation()}>
												<div class="ctx-item" onclick={() => showUsersMenu = false}>نمایش کاربران</div>
												<div class="ctx-item" onclick={() => showUsersMenu = false}>پایین آوردن دست‌ها</div>
												<div class="ctx-item" onclick={() => showUsersMenu = false}>رفع مسدودی همه</div>
												<div class="ctx-item" onclick={() => showUsersMenu = false}>حضور و غیاب</div>
												<div class="ctx-item" onclick={() => showUsersMenu = false}>دریافت لیست حاضرین</div>
												<div class="ctx-separator"></div>
												<div class="ctx-item" onclick={() => showUsersMenu = false}>بستن</div>
											</div>
										{/if}
									</div>
								</div>
								<div class="skyroom-block-content">
									<div class="skyroom-users-list-wrapper">
										<div class="skyroom-users-list">
											{#each participants as p}
												<div class="skyroom-user-row">
													<div class="skyroom-user-icon">
														<svg width="24" height="24" style="vertical-align:middle;fill:var(--text-color);width:16px;height:16px;display:inline-block;">
															<use xlink:href="#shape_person"></use>
														</svg>
													</div>
													<div class="skyroom-user-nickname">{p.name}</div>
												</div>
											{/each}
										</div>
									</div>
								</div>
							</div>
						{/if}
						{#if showChatPanel}
							<div class="skyroom-block skyroom-chat-block" style="flex:1;min-height:0;">
								<div class="skyroom-block-header">
									<div class="skyroom-block-title">
										<div class="skyroom-block-title-content">پیام‌ها</div>
									</div>
									<div style="position:relative;">
										<button class="skyroom-dots-btn" onclick={(e) => { e.stopPropagation(); showChatMenu = !showChatMenu; }}>
											<svg width="16" height="16"><use xlink:href="#shape_more_vert"></use></svg>
										</button>
										{#if showChatMenu}
											<div class="skyroom-context-menu" onclick={(e) => e.stopPropagation()}>
												<div class="ctx-item" onclick={() => showChatMenu = false}>نمایش بزرگتر</div>
												<div class="ctx-item" onclick={() => showChatMenu = false}>غیرفعال‌سازی چت</div>
												<div class="ctx-item" onclick={() => showChatMenu = false}>حالت خصوصی</div>
												<div class="ctx-item" onclick={() => showChatMenu = false}>پاک کردن همه پیام‌ها</div>
												<div class="ctx-item" onclick={() => showChatMenu = false}>ذخیره تمام پیام‌ها</div>
												<div class="ctx-item" onclick={() => showChatMenu = false}>کاربران ساکت‌شده</div>
												<div class="ctx-separator"></div>
												<div class="ctx-item" onclick={() => showChatMenu = false}>بستن</div>
											</div>
										{/if}
									</div>
								</div>
								<div class="skyroom-block-content" style="flex:1;min-height:0;">
									<ChatPanel messages={chatMessages} isAdmin={perms.canMic} onSend={sendChatMessage} onClose={() => showChatPanel = false} />
								</div>
							</div>
						{/if}
					</div>
				{/if}

				<!-- Mainbar -->
				<div class="skyroom-mainbar">
					{#if !connected}
						<div style="display:flex;flex-direction:column;align-items:center;justify-content:center;height:100%;gap:16px;">
							<div style="width:80px;height:80px;border-radius:50%;background:var(--block-bg-light);display:flex;align-items:center;justify-content:center;box-shadow:0 4px 20px rgba(0,0,0,0.3);">
								<span style="font-size:32px;font-weight:700;color:var(--accent);">{$auth.user?.display_name?.charAt(0) || '?'}</span>
							</div>
							<p style="color:var(--text-secondary);font-size:var(--font-size);">{$auth.user?.display_name || 'کاربر'}</p>
							{#if session?.status === 'live'}
								<button onclick={joinRoom} class="skyroom-btn">پیوستن به کلاس</button>
							{:else if session?.status === 'scheduled'}
								<p style="color:var(--inactive);font-size:var(--font-size-sm);">جلسه هنوز شروع نشده</p>
								{#if isPresenterOrAbove}
									<button onclick={startSession} class="skyroom-btn" style="background:#f59e0b;">شروع جلسه</button>
								{/if}
							{:else}
								<p style="color:var(--danger);font-size:var(--font-size-sm);">جلسه به پایان رسیده</p>
							{/if}
						</div>
					{/if}
					<div bind:this={remoteContainer} class="absolute inset-0 {gridCols} gap-1 p-1 auto-rows-fr"></div>
					<div class="absolute bottom-4 left-3 w-36 h-28 rounded overflow-hidden {webcamOn ? 'border border-[#3a3a5a]' : 'hidden'}">
						<video bind:this={localVideoEl} autoplay muted playsinline class="w-full h-full object-cover"></video>
					</div>
				</div>
			</div>
		</div>
	{:else}
		<div style="display:flex;align-items:center;justify-content:center;flex:1;"><p style="color:#b0b0b6;">جلسه یافت نشد</p></div>
	{/if}

	<!-- App Menu -->
	{#if showAppMenu}
		<AppMenu
			userRole={currentUserRole}
			onUserInfo={() => showModal = 'userInfo'}
			onConnectionStatus={() => showModal = 'connection'}
			onSettings={() => showModal = 'settings'}
			onLayout={() => showModal = 'layout'}
			onLeave={leaveRoom}
			onCloseRoom={leaveRoom}
			onDismiss={() => showAppMenu = false}
		/>
	{/if}

	<!-- Join Notification Toast -->
	{#if joinNotification.show}
		<div class="join-toast">
			<svg width="16" height="16" style="fill:#23b9d7;"><use xlink:href="#shape_group"></use></svg>
			<span>{joinNotification.name} به کلاس پیوست</span>
		</div>
	{/if}

	<!-- Modals -->
	{#if showModal === 'userInfo'}
		<UserInfoModal onClose={() => showModal = null} />
	{:else if showModal === 'connection'}
		<ConnectionStatusModal onClose={() => showModal = null} />
	{:else if showModal === 'settings'}
		<SettingsModal onClose={() => showModal = null} />
	{:else if showModal === 'layout'}
		<LayoutModal
			showUsers={showUsersPanel} showChat={showChatPanel}
			onToggleUsers={() => showUsersPanel = !showUsersPanel}
			onToggleChat={() => showChatPanel = !showChatPanel}
			onClose={() => showModal = null}
		/>
	{/if}
</div>

<!-- SVG Icons (hidden, referenced by use xlink:href) -->
<svg style="display:none;" xmlns="http://www.w3.org/2000/svg">
	<symbol id="shape_access_time" viewBox="0 0 24 24"><path d="M11.99 2C6.47 2 2 6.48 2 12s4.47 10 9.99 10C17.52 22 22 17.52 22 12S17.52 2 11.99 2zM12 20c-4.42 0-8-3.58-8-8s3.58-8 8-8 8 3.58 8 8-3.58 8-8 8zm.5-13H11v6l5.25 3.15.75-1.23-4.5-2.67z"/></symbol>
	<symbol id="shape_group" viewBox="0 0 24 24"><path d="M16 11c1.66 0 2.99-1.34 2.99-3S17.66 5 16 5c-1.66 0-3 1.34-3 3s1.34 3 3 3zm-8 0c1.66 0 2.99-1.34 2.99-3S9.66 5 8 5C6.34 5 5 6.34 5 8s1.34 3 3 3zm0 2c-2.33 0-7 1.17-7 3.5V19h14v-2.5c0-2.33-4.67-3.5-7-3.5zm8 0c-.29 0-.62.02-.97.05 1.16.84 1.97 1.97 1.97 3.45V19h6v-2.5c0-2.33-4.67-3.5-7-3.5z"/></symbol>
	<symbol id="shape_chat" viewBox="0 0 24 24"><path d="M20 2H4c-1.1 0-1.99.9-1.99 2L2 22l4-4h14c1.1 0 2-.9 2-2V4c0-1.1-.9-2-2-2zm-2 12H6v-2h12v2zm0-3H6V9h12v2zm0-3H6V6h12v2z"/></symbol>
	<symbol id="shape_volume_up" viewBox="0 0 24 24"><path d="M3 9v6h4l5 5V4L7 9H3zm13.5 3c0-1.77-1.02-3.29-2.5-4.03v8.05c1.48-.73 2.5-2.25 2.5-4.02zM14 3.23v2.06c2.89.86 5 3.54 5 6.71s-2.11 5.85-5 6.71v2.06c4.01-.91 7-4.49 7-8.77s-2.99-7.86-7-8.77z"/></symbol>
	<symbol id="shape_volume_off" viewBox="0 0 24 24"><path d="M16.5 12c0-1.77-1.02-3.29-2.5-4.03v2.21l2.45 2.45c.03-.2.05-.41.05-.63zm2.5 0c0 .94-.2 1.82-.54 2.64l1.51 1.51C20.63 14.91 21 13.5 21 12c0-4.28-2.99-7.86-7-8.77v2.06c2.89.86 5 3.54 5 6.71zM4.27 3L3 4.27 7.73 9H3v6h4l5 5v-6.73l4.25 4.25c-.67.52-1.42.93-2.25 1.18v2.06c1.38-.31 2.63-.95 3.69-1.81L19.73 21 21 19.73l-9-9L4.27 3zM12 4L9.91 6.09 12 8.18V4z"/></symbol>
	<symbol id="shape_mic" viewBox="0 0 24 24"><path d="M12 14c1.66 0 2.99-1.34 2.99-3L15 5c0-1.66-1.34-3-3-3S9 3.34 9 5v6c0 1.66 1.34 3 3 3zm5.3-3c0 3-2.54 5.1-5.3 5.1S6.7 14 6.7 11H5c0 3.41 2.72 6.23 6 6.72V21h2v-3.28c3.28-.48 6-3.3 6-6.72h-1.7z"/></symbol>
	<symbol id="shape_mic_off" viewBox="0 0 24 24"><path d="M19 11h-1.7c0 .74-.16 1.43-.43 2.05l1.23 1.23c.56-.98.9-2.09.9-3.28zm-4.02.17c0-.06.02-.11.02-.17V5c0-1.66-1.34-3-3-3S9 3.34 9 5v.18l5.98 5.99zM4.27 3L3 4.27l6.01 6.01V11c0 1.66 1.33 3 2.99 3 .22 0 .44-.03.65-.08l1.66 1.66c-.71.33-1.5.52-2.31.52-2.76 0-5.3-2.1-5.3-5.1H5c0 3.41 2.72 6.23 6 6.72V21h2v-3.28c.91-.13 1.77-.45 2.54-.9L19.73 21 21 19.73 4.27 3z"/></symbol>
	<symbol id="shape_videocam" viewBox="0 0 24 24"><path d="M17 10.5V7c0-.55-.45-1-1-1H4c-.55 0-1 .45-1 1v10c0 .55.45 1 1 1h12c.55 0 1-.45 1-1v-3.5l4 4v-11l-4 4z"/></symbol>
	<symbol id="shape_videocamoff" viewBox="0 0 24 24"><path d="M21 6.5l-4 4V7c0-.55-.45-1-1-1H9.82L21 17.18V6.5zM3.27 2L2 3.27 4.73 6H4c-.55 0-1 .45-1 1v10c0 .55.45 1 1 1h12c.21 0 .39-.08.54-.18L19.73 21 21 19.73 3.27 2z"/></symbol>
	<symbol id="shape_laptop" viewBox="0 0 24 24"><path d="M20 18c1.1 0 1.99-.9 1.99-2L22 6c0-1.1-.9-2-2-2H4c-1.1 0-2 .9-2 2v10c0 1.1.9 2 2 2H0v2h24v-2h-4zM4 6h16v10H4V6z"/></symbol>
	<symbol id="shape_brush" viewBox="0 0 24 24"><path d="M7 14c-1.66 0-3 1.34-3 3 0 1.31-1.16 2-2 2 .92 1.22 2.49 2 4 2 2.21 0 4-1.79 4-4 0-1.66-1.34-3-3-3zm13.71-9.37l-1.34-1.34a.996.996 0 00-1.41 0L9 12.25 11.75 15l8.96-8.96a.996.996 0 000-1.41z"/></symbol>
	<symbol id="shape_slideshow" viewBox="0 0 24 24"><path d="M19 3H5c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h14c1.1 0 2-.9 2-2V5c0-1.1-.9-2-2-2zm0 16H5V5h14v14zM7 10h2v7H7zm4-3h2v10h-2zm4 6h2v4h-2z"/></symbol>
	<symbol id="shape_hand" viewBox="0 0 24 24"><path d="M21 7c0-1.38-1.12-2.5-2.5-2.5-.17 0-.34.02-.5.05V4c0-1.38-1.12-2.5-2.5-2.5-.23 0-.46.03-.67.09C14.46.66 13.56 0 12.5 0c-1.23 0-2.25.89-2.46 2.06C9.87 2.02 9.69 2 9.5 2 8.12 2 7 3.12 7 4.5v5.89c-.34-.31-.76-.55-1.22-.67C4.56 9.56 3.28 10.33 3 11.58V20c0 1.1.9 2 2 2h14c1.1 0 2-.9 2-2V7z"/></symbol>
	<symbol id="shape_person" viewBox="0 0 24 24"><path d="M12 12c2.21 0 4-1.79 4-4s-1.79-4-4-4-4 1.79-4 4 1.79 4 4 4zm0 2c-2.67 0-8 1.34-8 4v2h16v-2c0-2.66-5.33-4-8-4z"/></symbol>
	<symbol id="shape_menu" viewBox="0 0 24 24"><path d="M3 18h18v-2H3v2zm0-5h18v-2H3v2zm0-7v2h18V6H3z"/></symbol>
	<symbol id="shape_more_vert" viewBox="0 0 24 24"><path d="M12 8c1.1 0 2-.9 2-2s-.9-2-2-2-2 .9-2 2 .9 2 2 2zm0 2c-1.1 0-2 .9-2 2s.9 2 2 2 2-.9 2-2-.9-2-2-2zm0 6c-1.1 0-2 .9-2 2s.9 2 2 2 2-.9 2-2-.9-2-2-2z"/></symbol>
	<symbol id="shape_web" viewBox="0 0 24 24"><path d="M20 4H4c-1.1 0-2 .9-2 2v12c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V6c0-1.1-.9-2-2-2zm0 14H4V6h16v12z"/></symbol>
	<symbol id="shape_exit" viewBox="0 0 24 24"><path d="M10.09 15.59L11.5 17l5-5-5-5-1.41 1.41L12.67 11H3v2h9.67l-2.58 2.59zM19 3H5c-1.11 0-2 .9-2 2v4h2V5h14v14H5v-4H3v4c0 1.1.89 2 2 2h14c1.1 0 2-.9 2-2V5c0-1.1-.9-2-2-2z"/></symbol>
	<symbol id="shape_power_settings_new" viewBox="0 0 24 24"><path d="M13 3h-2v10h2V3zm4.83 2.17l-1.42 1.42C17.99 7.86 19 9.81 19 12c0 3.87-3.13 7-7 7s-7-3.13-7-7c0-2.19 1.01-4.14 2.58-5.42L6.17 5.17C4.23 6.82 3 9.26 3 12c0 4.97 4.03 9 9 9s9-4.03 9-9c0-2.74-1.23-5.18-3.17-6.83z"/></symbol>
	<symbol id="shape_clear" viewBox="0 0 24 24"><path d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z"/></symbol>
	<symbol id="shape_info_outline" viewBox="0 0 24 24"><path d="M11 7h2v2h-2zm0 4h2v6h-2zm1-9C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm0 18c-4.41 0-8-3.59-8-8s3.59-8 8-8 8 3.59 8 8-3.59 8-8 8z"/></symbol>
	<symbol id="shape_network_check" viewBox="0 0 24 24"><path d="M24 8h-3V6c0-1.1-.9-2-2-2H5c-1.1 0-2 .9-2 2v12c0 1.1.9 2 2 2h14c1.1 0 2-.9 2-2v-3h3l-3-3V8zm-5 8H5V6h14v10z"/></symbol>
	<symbol id="shape_settings" viewBox="0 0 24 24"><path d="M19.14 12.94c.04-.3.06-.61.06-.94 0-.32-.02-.64-.07-.94l2.03-1.58a.49.49 0 00.12-.61l-1.92-3.32a.49.49 0 00-.59-.22l-2.39.96c-.5-.38-1.03-.7-1.62-.94l-.36-2.54a.484.484 0 00-.48-.41h-3.84c-.24 0-.43.17-.47.41l-.36 2.54c-.59.24-1.13.57-1.62.94l-2.39-.96c-.22-.08-.47 0-.59.22L2.74 8.87c-.12.21-.08.47.12.61l2.03 1.58c-.05.3-.07.62-.07.94s.02.64.07.94l-2.03 1.58a.49.49 0 00-.12.61l1.92 3.32c.12.22.37.29.59.22l2.39-.96c.5.38 1.03.7 1.62.94l.36 2.54c.05.24.24.41.48.41h3.84c.24 0 .44-.17.47-.41l.36-2.54c.59-.24 1.13-.56 1.62-.94l2.39.96c.22.08.47 0 .59-.22l1.92-3.32c.12-.22.07-.47-.12-.61l-2.01-1.58zM12 15.6c-1.98 0-3.6-1.62-3.6-3.6s1.62-3.6 3.6-3.6 3.6 1.62 3.6 3.6-1.62 3.6-3.6 3.6z"/></symbol>
</svg>

<style>
	@import url('https://cdn.jsdelivr.net/npm/vazirmatn@33.0.0/Vazirmatn-font-face.css');

	:root {
		--bg-color: #121822;
		--text-color: #e0e0e6;
		--text-secondary: #8a8a96;
		--font-family: "Vazirmatn", "Estedad-VF", Tahoma, sans-serif;
		--font-size: .875rem;
		--font-size-sm: .75rem;
		--font-size-xs: .7rem;
		--space: 5px;
		--space-sm: 2px;
		--block-bg: #1c2a3a;
		--block-bg-light: #233348;
		--accent: #23b9d7;
		--accent-glow: rgba(35, 185, 215, 0.15);
		--danger: #e05252;
		--inactive: #5a6070;
		--radius: 10px;
		--radius-sm: 6px;
	}

	.skyroom-col {
		display: flex;
		flex-direction: column;
		height: 100vh;
		background-color: var(--bg-color);
		color: var(--text-color);
		font-family: var(--font-family);
		font-size: var(--font-size);
	}

	.skyroom-row { display: flex; flex-direction: row; align-items: center; }

	.skyroom-header {
		display: flex;
		align-items: center;
		flex-shrink: 0;
		background-color: var(--block-bg);
		padding: 8px 12px;
		border-bottom: 1px solid rgba(0,0,0,0.3);
	}

	.skyroom-room-name { max-width: 100%; }

	.skyroom-room-timer {
		display: flex;
		align-items: center;
		gap: 6px;
		font-size: var(--font-size-sm);
		color: var(--text-secondary);
	}

	.skyroom-room-nav {
		background-color: var(--block-bg);
		border-radius: var(--radius);
		min-height: 44px;
		margin: 6px 8px 10px;
		padding: 6px 10px;
		display: flex;
		align-items: center;
		justify-content: space-between;
	}

	.skyroom-layout {
		display: flex;
		flex: 1;
		overflow: hidden;
		gap: 6px;
		padding: 0 8px 8px;
	}

	.skyroom-sidebar {
		flex-grow: 1;
		min-width: 260px;
		max-width: 320px;
		display: flex;
		flex-direction: column;
		gap: 6px;
	}

	.skyroom-mainbar {
		flex: 1;
		position: relative;
	}

	.skyroom-block {
		background-color: var(--block-bg);
		border-radius: var(--radius);
		min-height: 120px;
		padding: 0 6px 6px;
		display: flex;
		flex-direction: column;
		overflow: hidden;
	}

	.skyroom-block-header {
		display: flex;
		align-items: center;
		padding: 8px 4px 4px;
		gap: 6px;
	}

	.skyroom-block-title {
		padding: 0;
		font-size: var(--font-size-sm);
		flex: 1;
		color: var(--text-secondary);
		font-weight: 500;
	}

	.skyroom-block-title-content { line-height: 1.6; }
	.skyroom-block-icon { flex-shrink: 0; opacity: 0.5; }
	.skyroom-block-content { flex: 1; min-height: 0; overflow: hidden; }

	.skyroom-dots-btn {
		background: none;
		border: none;
		cursor: pointer;
		padding: 4px;
		border-radius: 4px;
		display: flex;
		align-items: center;
		justify-content: center;
	}
	.skyroom-dots-btn svg { fill: #8a8a96; }
	.skyroom-dots-btn:hover svg { fill: #e0e0e6; }
	.skyroom-dots-btn:hover { background: rgba(255,255,255,0.06); }

	.skyroom-context-menu {
		position: absolute;
		top: 100%;
		left: 0;
		background: #1c2a3a;
		border-radius: 8px;
		box-shadow: 0 8px 24px rgba(0,0,0,0.4);
		z-index: 100;
		min-width: 180px;
		padding: 4px 0;
		animation: menuFadeIn 0.12s ease;
	}

	@keyframes menuFadeIn {
		from { opacity: 0; transform: translateY(-4px); }
		to { opacity: 1; transform: translateY(0); }
	}

	.ctx-item {
		padding: 8px 14px;
		cursor: pointer;
		color: #e0e0e6;
		font-size: 0.8rem;
		transition: background 0.12s;
	}
	.ctx-item:hover { background: rgba(255,255,255,0.06); }

	.ctx-separator {
		height: 1px;
		background: rgba(255,255,255,0.08);
		margin: 3px 0;
	}

	.skyroom-users-count {
		font-size: var(--font-size-xs);
		color: var(--text-secondary);
		background: rgba(255,255,255,0.06);
		padding: 1px 6px;
		border-radius: 10px;
	}

	.skyroom-users-list-wrapper {
		overflow-y: auto;
		flex: 1;
	}

	.skyroom-users-list { padding: 4px 2px; }

	.skyroom-user-row {
		display: flex;
		align-items: center;
		gap: 6px;
		padding: 5px 6px;
		border-radius: var(--radius-sm);
		transition: background 0.15s;
	}

	.skyroom-user-row:hover { background: rgba(255,255,255,0.04); }
	.skyroom-user-icon { flex-shrink: 0; opacity: 0.4; }

	.skyroom-user-nickname {
		text-overflow: ellipsis;
		white-space: nowrap;
		overflow: hidden;
		font-weight: 600;
		color: var(--text-color);
		font-size: var(--font-size-sm);
	}

	.skyroom-icon-square {
		width: 36px;
		height: 36px;
		border-radius: var(--radius-sm);
		border: none;
		cursor: pointer;
		display: flex;
		align-items: center;
		justify-content: center;
		background: rgba(255,255,255,0.05);
		color: var(--inactive);
		transition: all 0.2s ease;
	}

	.skyroom-icon-square:hover {
		background: rgba(255,255,255,0.1);
		color: var(--text-color);
	}

	.skyroom-icon-square.active {
		background: var(--accent);
		color: #fff;
		box-shadow: 0 0 12px var(--accent-glow);
	}

	.skyroom-icon-square.active:hover {
		background: #1a9fc0;
	}

	.skyroom-icon-square svg {
		fill: currentColor;
	}

	.skyroom-btn {
		background-color: var(--accent);
		padding: 10px 28px;
		border-radius: var(--radius-sm);
		border: none;
		cursor: pointer;
		font-family: var(--font-family);
		font-size: var(--font-size);
		font-weight: 600;
		color: #fff;
		transition: all 0.2s ease;
	}

	.skyroom-btn:hover {
		background: #1a9fc0;
		transform: translateY(-1px);
		box-shadow: 0 4px 12px rgba(35, 185, 215, 0.3);
	}

	.join-toast {
		position: fixed;
		top: 60px;
		left: 50%;
		transform: translateX(-50%);
		background: #1c2a3a;
		border: 1px solid rgba(35, 185, 215, 0.3);
		color: #e0e0e6;
		padding: 8px 16px;
		border-radius: 8px;
		font-size: 0.8rem;
		display: flex;
		align-items: center;
		gap: 8px;
		z-index: 150;
		box-shadow: 0 4px 20px rgba(0,0,0,0.3);
		animation: toastSlide 0.3s ease;
	}

	@keyframes toastSlide {
		from { opacity: 0; transform: translateX(-50%) translateY(-12px); }
		to { opacity: 1; transform: translateX(-50%) translateY(0); }
	}
</style>
