<script lang="ts">
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { api } from '$lib/api';
	import { auth } from '$lib/stores';
	import { onMount } from 'svelte';
	import type { User, Tokens, Room } from '$lib/types';

	let room = $state<Room | null>(null);
	let loading = $state(true);
	let displayName = $state('');
	let email = $state('');
	let password = $state('');
	let actionLoading = $state(false);
	let error = $state('');
	let showGuestForm = $state(false);

	const slug = $derived(page.params.slug!);
	const isLoggedIn = $derived($auth.isLoggedIn);

	onMount(async () => {
		// If already logged in, try to join room directly
		if (isLoggedIn) {
			await joinRoom();
			return;
		}

		// Check for guest cookie first
		const guestData = getGuestCookie(slug);
		if (guestData) {
			try {
				localStorage.setItem('access_token', guestData.accessToken);
				localStorage.setItem('refresh_token', guestData.refreshToken);
				const meRes = await api.get<User>('/auth/me');
				if (meRes.success && meRes.data) {
					auth.login(meRes.data, {
						access_token: guestData.accessToken,
						refresh_token: guestData.refreshToken,
					});
					await joinRoom();
					return;
				}
			} catch {}
			displayName = guestData.displayName;
			showGuestForm = true;
		}

		const res = await api.get<any>('/rooms/slug/' + slug);
		if (res.success && res.data) {
			room = res.data.room;
		}
		loading = false;
	});

	async function joinRoom() {
		const res = await api.get<any>('/rooms/slug/' + slug);
		if (!res.success || !res.data) {
			loading = false;
			return;
		}
		room = res.data.room;

		// Find or create session
		let foundSession = null;
		const sessRes = await api.get<any>('/sessions', { per_page: '50' });
		if (sessRes.success && sessRes.data?.items) {
			foundSession = sessRes.data.items.find((s: any) => s.class_id === room!.id && s.status === 'live');
		}

		if (!foundSession) {
			const createRes = await api.post<any>('/sessions', {
				title: room!.name,
				class_id: room!.id,
				scheduled_at: new Date().toISOString(),
				duration: 120,
			});
			if (createRes.success && createRes.data) {
				await api.post(`/sessions/${createRes.data.id}/start`);
				foundSession = createRes.data;
			}
		}

		if (foundSession) {
			// Store session ID and show classroom
			roomId = foundSession.id;
			const sessionRes = await api.get(`/sessions/${foundSession.id}`);
			if (sessionRes.success) currentSession = sessionRes.data!;
			loading = false;
			if (currentSession?.status === 'live') startTimer();
			connectChatWs();
			fetchParticipants();
		} else {
			loading = false;
		}
	}

	function getGuestCookie(roomSlug: string): { displayName: string; accessToken: string; refreshToken: string } | null {
		if (typeof document === 'undefined') return null;
		const name = `iroom_guest_${roomSlug}=`;
		const decodedCookie = decodeURIComponent(document.cookie);
		const ca = decodedCookie.split(';');
		for (let c of ca) {
			c = c.trim();
			if (c.indexOf(name) === 0) {
				try { return JSON.parse(c.substring(name.length)); } catch { return null; }
			}
		}
		return null;
	}

	function setGuestCookie(roomSlug: string, data: { displayName: string; accessToken: string; refreshToken: string }) {
		const expires = new Date(Date.now() + 30 * 24 * 60 * 60 * 1000).toUTCString();
		document.cookie = `iroom_guest_${roomSlug}=${encodeURIComponent(JSON.stringify(data))}; expires=${expires}; path=/; SameSite=Lax`;
	}

	async function handleLogin() {
		if (!email || !password) { error = 'ایمیل و رمز عبور الزامی است'; return; }
		actionLoading = true; error = '';
		const res = await api.post<{ user: User; tokens: Tokens }>('/auth/login', { email, password });
		if (!res.success) { error = res.error || 'خطا در ورود'; actionLoading = false; return; }
		auth.login(res.data!.user, res.data!.tokens);
		await joinRoom();
	}

	async function handleGuestJoin() {
		if (!displayName.trim()) { error = 'لطفاً نام خود را وارد کنید'; return; }
		actionLoading = true; error = '';
		const res = await api.post<{ user: User; tokens: Tokens }>('/auth/room-guest-login', {
			room_slug: slug,
			display_name: displayName.trim()
		});
		if (!res.success) { error = res.error || 'خطا در ورود'; actionLoading = false; return; }
		auth.login(res.data!.user, res.data!.tokens);
		setGuestCookie(slug, {
			displayName: displayName.trim(),
			accessToken: res.data!.tokens.access_token,
			refreshToken: res.data!.tokens.refresh_token,
		});
		await joinRoom();
	}

	// --- Classroom State ---
	import { onMount as onClassroomMount, onDestroy } from 'svelte';
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

	let roomId = $state<number | null>(null);
	let currentSession = $state<any>(null);
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
	let menuPos = $state<{ top: number; left: number }>({ top: 0, left: 0 });
	let participants = $state<Participant[]>([]);
	let chatMessages = $state<ChatMessage[]>([]);
	let localVideoEl: HTMLVideoElement;
	let remoteContainer: HTMLDivElement;
	let chatWs: WebSocket | null = null;

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

	function handleClickOutside() { showUsersMenu = false; showChatMenu = false; showAppMenu = false; }

	function connectChatWs() {
		const token = localStorage.getItem('access_token');
		if (!token || !roomId) return;
		const proto = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
		chatWs = new WebSocket(`${proto}//${window.location.host}/ws/sessions/${roomId}?token=${token}`);
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

	async function joinClassroom() {
		if (!roomId) return;
		const joinRes = await api.get(`/sessions/${roomId}/classroom`);
		if (!joinRes.success || !joinRes.data) return;
		const { room_id, user_id, role } = joinRes.data;
		try {
			pion = new PionClient({
				roomId: String(room_id), userId: String(user_id), role,
				displayName: $auth.user?.display_name || 'کاربر',
			});
			pion.onLocalStream = (stream) => { if (localVideoEl) localVideoEl.srcObject = stream; };
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
		}
	}

	function disconnect() {
		if (pion) { pion.disconnect(); pion = null; }
		connected = false;
		if (timerInterval) { clearInterval(timerInterval); timerInterval = null; }
		if (participantInterval) { clearInterval(participantInterval); participantInterval = null; }
	}

	function startTimer() { timerInterval = setInterval(() => { elapsedSeconds++; }, 1000); }
	let participantInterval: ReturnType<typeof setInterval> | null = null;
	function startParticipantRefresh() { participantInterval = setInterval(() => { fetchParticipants(); }, 5000); }

	function toggleMic() { if (!perms.canMic) return; if (pion) pion.toggleAudio(); micOn = !micOn; }
	function toggleWebcam() { if (!perms.canWebcam) return; if (pion) pion.toggleVideo(); webcamOn = !webcamOn; }
	function toggleScreenShare() { if (!perms.canScreenShare) return; if (pion && !screenShareOn) pion.shareScreen(); screenShareOn = !screenShareOn; }
	function toggleHand() { if (!perms.canHandRaise) return; handRaised = !handRaised; }
	function sendChatMessage(text: string) {
		if (!perms.canChat || !chatWs || chatWs.readyState !== WebSocket.OPEN) return;
		chatWs.send(JSON.stringify({ type: 'message', content: text }));
	}
	function leaveRoom() { disconnect(); auth.logout(); goto('/'); }
	function positionMenu(e: MouseEvent) { const rect = (e.currentTarget as HTMLElement).getBoundingClientRect(); menuPos = { top: rect.bottom + 4, left: rect.left }; }
	function showJoinNotification(name: string) { joinNotification = { name, show: true }; setTimeout(() => { joinNotification = { name: '', show: false }; }, 3000); }

	async function fetchParticipants() {
		if (!roomId) return;
		try {
			const res = await api.get<any[]>(`/sessions/${roomId}/classroom/participants`);
			if (res.success && Array.isArray(res.data)) {
				participants = res.data.map((p: any) => ({
					id: p.id, name: p.name, role: 'student' as UserRole,
					isSpeaking: false, hasVideo: false, hasAudio: true, hasScreen: false, hasWhiteboard: false, handRaised: false,
				}));
			}
		} catch (e) {}
	}

	async function startSession() {
		if (!roomId) return;
		const res = await api.post(`/sessions/${roomId}/start`);
		if (res.success) {
			currentSession = { ...currentSession, status: 'live' };
			startTimer();
		}
	}
</script>

<svelte:window onclick={handleClickOutside} />

{#if !isLoggedIn || !currentSession}
	<!-- Join Page -->
	<div class="min-h-screen flex items-center justify-center px-4" style="background: linear-gradient(135deg, #0b1120 0%, #1a1a2e 50%, #0d1b2a 100%);">
		<div class="w-full max-w-[400px] rounded-2xl p-6" style="background: #16213e; border: 1px solid #2a2a4a;">
			{#if loading}
				<div class="text-center py-8"><div class="animate-spin w-8 h-8 border-3 border-[#23b9d7] border-t-transparent rounded-full mx-auto"></div></div>
			{:else if room}
				<div class="text-center mb-6">
					<div class="w-14 h-14 rounded-xl mx-auto mb-3 flex items-center justify-center text-white font-bold text-xl" style="background: {room.color};">{room.name.charAt(0)}</div>
					<h1 class="text-xl font-bold" style="color: #eaeaea;">{room.name}</h1>
					{#if room.description}<p class="text-sm mt-1" style="color: #94a3b8;">{room.description}</p>{/if}
				</div>
				{#if error}<div class="mb-4 px-4 py-3 rounded-lg text-sm text-center" style="background: rgba(224,82,82,0.08); color: #e05252;">{error}</div>{/if}
				{#if !showGuestForm}
					<form onsubmit={(e) => { e.preventDefault(); handleLogin(); }} class="space-y-3">
						<div><label class="block text-xs font-medium mb-1.5" style="color: #94a3b8;">ایمیل</label><input type="email" bind:value={email} placeholder="ایمیل" required class="w-full px-4 py-2.5 rounded-lg text-sm outline-none" style="border: 1px solid #2a2a4a; color: #eaeaea; background: #0f3460;" /></div>
						<div><label class="block text-xs font-medium mb-1.5" style="color: #94a3b8;">رمز عبور</label><input type="password" bind:value={password} placeholder="رمز عبور" required class="w-full px-4 py-2.5 rounded-lg text-sm outline-none" style="border: 1px solid #2a2a4a; color: #eaeaea; background: #0f3460;" /></div>
						<button type="submit" disabled={actionLoading} class="w-full py-2.5 rounded-lg text-sm font-semibold text-white" style="background: #23b9d7;">{actionLoading ? 'در حال ورود...' : 'ورود'}</button>
					</form>
					{#if room.guest_login_enabled}
						<div class="mt-4 pt-4" style="border-top: 1px solid #2a2a4a;">
							<button onclick={() => showGuestForm = true} class="w-full py-2.5 rounded-lg text-sm font-semibold" style="background: rgba(255,255,255,0.08); color: #eaeaea; border: 1px solid #2a2a4a;">ورود مهمان</button>
						</div>
					{/if}
				{:else}
					<form onsubmit={(e) => { e.preventDefault(); handleGuestJoin(); }} class="space-y-3">
						<div><label class="block text-xs font-medium mb-1.5" style="color: #94a3b8;">نام شما</label><input type="text" bind:value={displayName} placeholder="نام خود را وارد کنید" dir="auto" required class="w-full px-4 py-2.5 rounded-lg text-sm outline-none" style="border: 1px solid #2a2a4a; color: #eaeaea; background: #0f3460;" /></div>
						<button type="submit" disabled={actionLoading} class="w-full py-2.5 rounded-lg text-sm font-semibold text-white" style="background: #23b9d7;">{actionLoading ? 'در حال پیوستن...' : 'پیوستن به اتاق'}</button>
					</form>
					<div class="mt-3 text-center"><button onclick={() => { showGuestForm = false; error = ''; }} class="text-xs" style="color: #6790a0;">بازگشت به ورود</button></div>
				{/if}
			{:else}
				<div class="text-center py-8"><p style="color: #e05252;">اتاق یافت نشد</p></div>
			{/if}
		</div>
	</div>
{:else}
	<!-- Classroom -->
	<div class="skyroom-col" style="background-color: var(--bg-color); color: var(--text-color); font-family: var(--font-family); font-size: var(--font-size);">
		{#if !connected}
			<div style="display:flex;flex-direction:column;align-items:center;justify-content:center;flex:1;gap:16px;">
				<div style="width:80px;height:80px;border-radius:50%;background:var(--block-bg-light);display:flex;align-items:center;justify-content:center;box-shadow:0 4px 20px rgba(0,0,0,0.3);">
					<span style="font-size:32px;font-weight:700;color:var(--accent);">{$auth.user?.display_name?.charAt(0) || '?'}</span>
				</div>
				<p style="color:var(--text-secondary);font-size:var(--font-size);">{$auth.user?.display_name || 'کاربر'}</p>
				{#if currentSession?.status === 'live'}
					<button onclick={joinClassroom} class="skyroom-btn">پیوستن به کلاس</button>
				{:else if currentSession?.status === 'scheduled'}
					<p style="color:var(--inactive);font-size:var(--font-size-sm);">جلسه هنوز شروع نشده</p>
					{#if isPresenterOrAbove}<button onclick={startSession} class="skyroom-btn" style="background:#f59e0b;">شروع جلسه</button>{/if}
				{:else}
					<p style="color:var(--danger);font-size:var(--font-size-sm);">جلسه به پایان رسیده</p>
				{/if}
			</div>
		{:else}
			<header class="skyroom-header">
				<div class="skyroom-row" style="gap:8px;min-width:0;">
					<div style="width:28px;height:28px;border-radius:6px;background:var(--accent);display:flex;align-items:center;justify-content:center;flex-shrink:0;"><span style="font-size:12px;font-weight:700;color:#fff;">{($auth.user?.display_name || 'م').charAt(0)}</span></div>
					<span style="font-weight:600;font-size:var(--font-size);white-space:nowrap;overflow:hidden;text-overflow:ellipsis;">{$auth.user?.display_name || 'مالک'}</span>
					<span style="color:var(--inactive);">:</span>
					<span style="font-weight:500;font-size:var(--font-size);white-space:nowrap;overflow:hidden;text-overflow:ellipsis;">{currentSession?.title}</span>
				</div>
				<div style="flex:1;"></div>
				<div class="skyroom-room-timer"><svg width="16" height="16" style="fill:currentColor;"><use xlink:href="#shape_access_time"></use></svg><span style="font-family:monospace;font-size:var(--font-size-sm);">{formattedTime}</span></div>
			</header>
			<div id="workspace" class="skyroom-col" style="flex:1;overflow:hidden;">
				<div class="skyroom-room-nav" style="position:relative;">
					<div class="skyroom-row" style="flex-shrink:0;gap:4px;">
						<button class="skyroom-icon-square" title="منو" onclick={() => showAppMenu = !showAppMenu}><svg width="18" height="18"><use xlink:href="#shape_menu"></use></svg></button>
						<button class="skyroom-icon-square" class:active={showChatPanel} title="پیام‌ها" onclick={() => showChatPanel = !showChatPanel}><svg width="18" height="18"><use xlink:href="#shape_chat"></use></svg></button>
						<button class="skyroom-icon-square" class:active={showUsersPanel} title="کاربران" onclick={() => showUsersPanel = !showUsersPanel}><svg width="18" height="18"><use xlink:href="#shape_group"></use></svg></button>
					</div>
					<div style="flex:1;"></div>
					<div class="skyroom-row" style="flex-shrink:0;gap:4px;">
						{#if perms.canHandRaise}<button class="skyroom-icon-square" class:active={handRaised} title="بالا بردن دست" onclick={toggleHand}><svg width="18" height="18"><use xlink:href="#shape_hand"></use></svg></button>{/if}
						{#if perms.canScreenShare}<button class="skyroom-icon-square" class:active={screenShareOn} title="اشتراک‌گذاری صفحه" onclick={toggleScreenShare}><svg width="18" height="18"><use xlink:href="#shape_laptop"></use></svg></button>{/if}
						{#if perms.canWebcam}<button class="skyroom-icon-square" class:active={webcamOn} title="وبکم" onclick={toggleWebcam}><svg width="18" height="18"><use xlink:href={webcamOn ? '#shape_videocam' : '#shape_videocamoff'}></use></svg></button>{/if}
						{#if perms.canMic}<button class="skyroom-icon-square" class:active={micOn} title="میکروفون" onclick={toggleMic}><svg width="18" height="18"><use xlink:href={micOn ? '#shape_mic' : '#shape_mic_off'}></use></svg></button>{/if}
						<button class="skyroom-icon-square" class:active={audioOn} title="خروجی صدا" onclick={() => audioOn = !audioOn}><svg width="18" height="18"><use xlink:href={audioOn ? '#shape_volume_up' : '#shape_volume_off'}></use></svg></button>
					</div>
				</div>
				<div class="skyroom-layout">
					{#if showUsersPanel || showChatPanel}
						<div class="skyroom-sidebar">
							{#if showUsersPanel}
								<div class="skyroom-block skyroom-users-block">
									<div class="skyroom-block-header">
										<div class="skyroom-block-title"><div class="skyroom-block-title-content">کاربران</div></div>
										<span class="skyroom-users-count">{participants.length}</span>
									</div>
									<div class="skyroom-block-content">
										<div class="skyroom-users-list-wrapper"><div class="skyroom-users-list">
											{#each participants as p}<div class="skyroom-user-row"><div class="skyroom-user-icon"><svg width="24" height="24" style="vertical-align:middle;fill:var(--text-color);width:16px;height:16px;display:inline-block;"><use xlink:href="#shape_person"></use></svg></div><div class="skyroom-user-nickname">{p.name}</div></div>{/each}
										</div></div>
									</div>
								</div>
							{/if}
							{#if showChatPanel}
								<div class="skyroom-block skyroom-chat-block" style="flex:1;min-height:0;">
									<div class="skyroom-block-header"><div class="skyroom-block-title"><div class="skyroom-block-title-content">پیام‌ها</div></div></div>
									<div class="skyroom-block-content" style="flex:1;min-height:0;"><ChatPanel messages={chatMessages} isAdmin={perms.canMic} onSend={sendChatMessage} onClose={() => showChatPanel = false} /></div>
								</div>
							{/if}
						</div>
					{/if}
					<div class="skyroom-mainbar">
						<div bind:this={remoteContainer} style="position:absolute;inset:0;display:grid;{gridCols};gap:4px;padding:4px;"></div>
						<div class="absolute bottom-4 left-3 w-36 h-28 rounded overflow-hidden border border-[#3a3a5a]"><video bind:this={localVideoEl} autoplay muted playsinline class="w-full h-full object-cover"></video></div>
					</div>
				</div>
			</div>
		{/if}
		{#if showAppMenu}<AppMenu userRole={currentUserRole} onUserInfo={() => showModal = 'userInfo'} onConnectionStatus={() => showModal = 'connection'} onSettings={() => showModal = 'settings'} onLayout={() => showModal = 'layout'} onLeave={leaveRoom} onCloseRoom={leaveRoom} onDismiss={() => showAppMenu = false} />{/if}
		{#if joinNotification.show}<div class="join-toast"><svg width="16" height="16" style="fill:#23b9d7;"><use xlink:href="#shape_group"></use></svg><span>{joinNotification.name} به کلاس پیوست</span></div>{/if}
		{#if showModal === 'userInfo'}<UserInfoModal onClose={() => showModal = null} />
		{:else if showModal === 'connection'}<ConnectionStatusModal onClose={() => showModal = null} />
		{:else if showModal === 'settings'}<SettingsModal onClose={() => showModal = null} />
		{:else if showModal === 'layout'}<LayoutModal showUsers={showUsersPanel} showChat={showChatPanel} onToggleUsers={() => showUsersPanel = !showUsersPanel} onToggleChat={() => showChatPanel = !showChatPanel} onClose={() => showModal = null} />{/if}
	</div>
{/if}

<!-- SVG Icons -->
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
	<symbol id="shape_exit" viewBox="0 0 24 24"><path d="M10.09 15.59L11.5 17l5-5-5-5-1.41 1.41L12.67 11H3v2h9.67l-2.58 2.59zM19 3H5c-1.11 0-2 .9-2 2v4h2V5h14v14H5v-4H3v4c0 1.1.89 2 2 2h14c1.1 0 2-.9 2-2V5c0-1.1-.9-2-2-2z"/></symbol>
</svg>

<style>
	@import url('https://cdn.jsdelivr.net/npm/vazirmatn@33.0.0/Vazirmatn-font-face.css');
	:root { --bg-color: #121822; --text-color: #e0e0e6; --text-secondary: #8a8a96; --font-family: "Vazirmatn", "Estedad-VF", Tahoma, sans-serif; --font-size: .875rem; --font-size-sm: .75rem; --block-bg: #1c2a3a; --block-bg-light: #233348; --accent: #23b9d7; --accent-glow: rgba(35, 185, 215, 0.15); --danger: #e05252; --inactive: #5a6070; --radius: 10px; --radius-sm: 6px; }
	.skyroom-col { display: flex; flex-direction: column; height: 100vh; background-color: var(--bg-color); color: var(--text-color); font-family: var(--font-family); font-size: var(--font-size); }
	.skyroom-row { display: flex; flex-direction: row; align-items: center; }
	.skyroom-header { display: flex; align-items: center; flex-shrink: 0; background-color: var(--block-bg); padding: 8px 12px; border-bottom: 1px solid rgba(0,0,0,0.3); }
	.skyroom-room-timer { display: flex; align-items: center; gap: 6px; font-size: var(--font-size-sm); color: var(--text-secondary); }
	.skyroom-room-nav { background-color: var(--block-bg); border-radius: var(--radius); min-height: 44px; margin: 6px 8px 10px; padding: 6px 10px; display: flex; align-items: center; justify-content: space-between; }
	.skyroom-layout { display: flex; flex: 1; overflow: hidden; gap: 6px; padding: 0 8px 8px; }
	.skyroom-sidebar { flex-grow: 1; min-width: 260px; max-width: 320px; display: flex; flex-direction: column; gap: 6px; }
	.skyroom-mainbar { flex: 1; position: relative; }
	.skyroom-block { background-color: var(--block-bg); border-radius: var(--radius); min-height: 120px; padding: 0 6px 6px; display: flex; flex-direction: column; overflow: hidden; }
	.skyroom-block-header { display: flex; align-items: center; padding: 8px 4px 4px; gap: 6px; }
	.skyroom-block-title { padding: 0; font-size: var(--font-size-sm); flex: 1; color: var(--text-secondary); font-weight: 500; }
	.skyroom-block-title-content { line-height: 1.6; }
	.skyroom-block-content { flex: 1; min-height: 0; overflow: hidden; }
	.skyroom-dots-btn { background: none; border: none; cursor: pointer; padding: 4px; border-radius: 4px; display: flex; align-items: center; justify-content: center; }
	.skyroom-dots-btn svg { fill: #8a8a96; }
	.skyroom-dots-btn:hover svg { fill: #e0e0e6; }
	.skyroom-users-count { font-size: var(--font-size-xs, .7rem); color: var(--text-secondary); background: rgba(255,255,255,0.06); padding: 1px 6px; border-radius: 10px; }
	.skyroom-users-list-wrapper { overflow-y: auto; flex: 1; }
	.skyroom-users-list { padding: 4px 2px; }
	.skyroom-user-row { display: flex; align-items: center; gap: 6px; padding: 5px 6px; border-radius: var(--radius-sm); }
	.skyroom-user-icon { flex-shrink: 0; opacity: 0.4; }
	.skyroom-user-nickname { text-overflow: ellipsis; white-space: nowrap; overflow: hidden; font-weight: 600; color: var(--text-color); font-size: var(--font-size-sm); }
	.skyroom-icon-square { width: 36px; height: 36px; border-radius: var(--radius-sm); border: none; cursor: pointer; display: flex; align-items: center; justify-content: center; background: rgba(255,255,255,0.05); color: var(--inactive); transition: all 0.2s ease; }
	.skyroom-icon-square:hover { background: rgba(255,255,255,0.1); color: var(--text-color); }
	.skyroom-icon-square.active { background: var(--accent); color: #fff; box-shadow: 0 0 12px var(--accent-glow); }
	.skyroom-icon-square svg { fill: currentColor; }
	.skyroom-btn { background-color: var(--accent); padding: 10px 28px; border-radius: var(--radius-sm); border: none; cursor: pointer; font-family: var(--font-family); font-size: var(--font-size); font-weight: 600; color: #fff; transition: all 0.2s ease; }
	.skyroom-btn:hover { background: #1a9fc0; transform: translateY(-1px); box-shadow: 0 4px 12px rgba(35, 185, 215, 0.3); }
	.join-toast { position: fixed; top: 60px; left: 50%; transform: translateX(-50%); background: #1c2a3a; border: 1px solid rgba(35, 185, 215, 0.3); color: #e0e0e6; padding: 8px 16px; border-radius: 8px; font-size: 0.8rem; display: flex; align-items: center; gap: 8px; z-index: 150; box-shadow: 0 4px 20px rgba(0,0,0,0.3); }
</style>
