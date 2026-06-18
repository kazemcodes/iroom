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

<div class="flex flex-col h-[calc(100vh-8rem)] bg-white rounded-xl overflow-hidden">
	<!-- Header -->
	<div class="px-5 py-3 border-b flex items-center justify-between shrink-0">
		<div class="flex items-center gap-3">
			<a href="/classes" class="text-gray-400 hover:text-gray-600">
				<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 5l7 7-7 7" /></svg>
			</a>
			<div>
				<h2 class="font-bold text-gray-900">{session?.title || 'چت'}</h2>
				<p class="text-xs text-gray-500">{messages.length} پیام</p>
			</div>
		</div>
		{#if session?.status === 'live'}
			<a href="/classroom/{session?.id}" class="px-3 py-1.5 bg-green-600 text-white text-xs rounded-lg hover:bg-green-700 flex items-center gap-1">
				<span class="w-2 h-2 bg-white rounded-full animate-pulse"></span>
				کلاس زنده
			</a>
		{/if}
	</div>

	{#if loading}
		<div class="flex-1 flex items-center justify-center">
			<div class="animate-spin h-8 w-8 border-4 border-blue-600 border-t-transparent rounded-full"></div>
		</div>
	{:else}
		<!-- Messages -->
		<div bind:this={chatContainer} class="flex-1 overflow-y-auto px-5 py-4 space-y-4">
			{#if messages.length === 0}
				<div class="flex items-center justify-center h-full text-gray-400">
					<p>هنوز پیامی ارسال نشده</p>
				</div>
			{:else}
				{#each messages as msg}
					<div class="flex {isOwnMessage(msg) ? 'justify-end' : 'justify-start'}">
						<div class="max-w-[70%] {isOwnMessage(msg) ? 'bg-blue-50 text-blue-900' : 'bg-gray-100 text-gray-900'} rounded-2xl px-4 py-2.5">
							<p class="text-sm whitespace-pre-wrap break-words">{msg.content}</p>
							<p class="text-[10px] mt-1 {isOwnMessage(msg) ? 'text-blue-600' : 'text-gray-500'}">{formatTime(msg.created_at)}</p>
						</div>
					</div>
				{/each}
			{/if}
		</div>

		<!-- Input -->
		<div class="px-5 py-3 border-t shrink-0">
			<form onsubmit={(e) => { e.preventDefault(); sendMessage(); }} class="flex gap-3">
				<input
					type="text"
					bind:value={newMessage}
					class="flex-1 px-4 py-2.5 border border-gray-200 rounded-xl text-sm focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none bg-gray-50 focus:bg-white"
					placeholder="پیام بنویسید..."
				/>
				<button
					type="submit"
					disabled={!newMessage.trim()}
					class="px-4 py-2.5 bg-blue-600 text-white rounded-xl text-sm font-medium hover:bg-blue-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2"
				>
					<svg class="w-4 h-4 rotate-180" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8" /></svg>
					ارسال
				</button>
			</form>
		</div>
	{/if}
</div>
