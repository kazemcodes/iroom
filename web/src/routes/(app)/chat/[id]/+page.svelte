<script lang="ts">
	import { page } from '$app/state';
	import { auth } from '$lib/stores';
	import { api } from '$lib/api';
	import { onMount, onDestroy, tick } from 'svelte';
	import type { Message, Session } from '$lib/types';

	let messages = $state<Message[]>([]);
	let session = $state<Session | null>(null);
	let newMessage = $state('');
	let loading = $state(true);
	let chatContainer: HTMLDivElement;
	let ws: WebSocket | null = null;

	const sessionId = $derived(page.params.id);

	onMount(async () => {
		await loadData();
		connectWS();
	});

	onDestroy(() => { ws?.close(); });

	async function loadData() {
		loading = true;
		const [sessionRes, messagesRes] = await Promise.all([
			api.get<Session>(`/sessions/${sessionId}`),
			api.get<Message[]>(`/sessions/${sessionId}/messages`)
		]);

		if (sessionRes.success) session = sessionRes.data!;
		if (messagesRes.success && messagesRes.data) messages = Array.isArray(messagesRes.data) ? messagesRes.data : [];
		loading = false;
		await tick();
		scrollToBottom();
	}

	function connectWS() {
		const token = localStorage.getItem('access_token');
		if (!token) return;

		const wsProto = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
		ws = new WebSocket(`${wsProto}//${window.location.host}/ws/sessions/${sessionId}?token=${token}`);

		ws.onmessage = async (event) => {
			const data = JSON.parse(event.data);
			if (data.type === 'message') {
				messages = [...messages, data.message];
				await tick();
				scrollToBottom();
			}
		};

		ws.onclose = () => {
			setTimeout(connectWS, 3000);
		};
	}

	async function sendMessage() {
		if (!newMessage.trim()) return;

		const content = newMessage.trim();
		newMessage = '';

		if (ws?.readyState === WebSocket.OPEN) {
			ws.send(JSON.stringify({ type: 'message', content }));
		} else {
			await api.post(`/sessions/${sessionId}/messages`, { content });
			await loadData();
		}
	}

	function scrollToBottom() {
		if (chatContainer) {
			chatContainer.scrollTop = chatContainer.scrollHeight;
		}
	}

	function formatTime(d: string) {
		return new Date(d).toLocaleTimeString('fa-IR', { hour: '2-digit', minute: '2-digit' });
	}

	function isOwnMessage(msg: Message) {
		return msg.user_id === $auth.user?.id;
	}
</script>

<div class="flex flex-col h-[calc(100vh-8rem)] sky-card overflow-hidden">
	<!-- Header -->
	<div class="px-5 py-3 flex items-center justify-between shrink-0" style="border-bottom: 1px solid var(--color-zen-garden);">
		<div class="flex items-center gap-3">
			<a href="/classes" class="sky-btn-icon">
				<svg width="20" height="20" fill="none" stroke="currentColor" stroke-width="1.75" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M15 6l-6 6 6 6"/></svg>
			</a>
			<div>
				<h2 class="font-bold" style="color: var(--color-midnight-sky);">{session?.title || 'چت'}</h2>
				<p class="text-xs" style="color: var(--color-mystic-sea);">{messages.length} پیام</p>
			</div>
		</div>
		{#if session?.status === 'live'}
			<a href="/classroom/{session?.id}" class="sky-btn flex items-center gap-1.5" style="background: var(--color-lush-meadow); color: white; padding: 0.4rem 0.85rem; font-size: 12px;">
				<span class="w-2 h-2 bg-white rounded-full animate-pulse"></span>
				کلاس زنده
			</a>
		{/if}
	</div>

	{#if loading}
		<div class="flex-1 flex items-center justify-center"><svg class="sky-spinner lg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="color: var(--color-crystal-clear);"><path d="M12 2v4M12 18v4M4.93 4.93l2.83 2.83M16.24 16.24l2.83 2.83M2 12h4M18 12h4M4.93 19.07l2.83-2.83M16.24 7.76l2.83-2.83"/></svg></div>
	{:else}
		<!-- Messages -->
		<div bind:this={chatContainer} class="flex-1 overflow-y-auto px-5 py-4 space-y-4">
			{#if messages.length === 0}
				<div class="flex items-center justify-center h-full" style="color: var(--color-moonlit-mist);">
					<p>هنوز پیامی ارسال نشده</p>
				</div>
			{:else}
				{#each messages as msg}
					<div class="flex {isOwnMessage(msg) ? 'justify-end' : 'justify-start'}">
						<div class="max-w-[70%] rounded-2xl px-4 py-2.5" style="background: {isOwnMessage(msg) ? 'var(--color-polar-ice)' : 'var(--color-secret-glow)'};">
							<p class="text-sm whitespace-pre-wrap break-words" style="color: {isOwnMessage(msg) ? 'var(--color-ocean-wave)' : 'var(--color-midnight-sky)'};">{msg.content}</p>
							<p class="text-[10px] mt-1" style="color: {isOwnMessage(msg) ? 'var(--color-crystal-clear)' : 'var(--color-moonlit-mist)'};">{formatTime(msg.created_at)}</p>
						</div>
					</div>
				{/each}
			{/if}
		</div>

		<!-- Input -->
		<div class="px-5 py-3 shrink-0" style="border-top: 1px solid var(--color-zen-garden);">
			<form onsubmit={(e) => { e.preventDefault(); sendMessage(); }} class="flex gap-3">
				<input type="text" bind:value={newMessage} class="sky-input flex-1" placeholder="پیام بنویسید..." />
				<button type="submit" disabled={!newMessage.trim()} class="sky-btn sky-btn-primary flex items-center gap-2">
					<svg width="16" height="16" class="rotate-180" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8"/></svg>
					ارسال
				</button>
			</form>
		</div>
	{/if}
</div>
