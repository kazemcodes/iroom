<script lang="ts">
	import type { ChatMessage } from '$lib/classroom/types';
	import EmojiPicker from './EmojiPicker.svelte';

	let {
		messages = [],
		isAdmin = false,
		onSend,
		onReply,
		onEdit,
		onDelete,
		onPin,
		onClose,
	}: {
		messages: ChatMessage[];
		isAdmin: boolean;
		onSend: (content: string) => void;
		onReply?: (messageId: string) => void;
		onEdit?: (messageId: string, content: string) => void;
		onDelete?: (messageId: string) => void;
		onPin?: (messageId: string) => void;
		onClose: () => void;
	} = $props();

	let chatInput = $state('');
	let replyTo = $state<ChatMessage | null>(null);
	let editingMessage = $state<ChatMessage | null>(null);
	let editContent = $state('');
	let showEmojiPicker = $state(false);
	let contextMenu = $state<{ show: boolean; x: number; y: number; message: ChatMessage | null }>({
		show: false, x: 0, y: 0, message: null
	});
	let chatContainer: HTMLDivElement;

	function sendMessage() {
		if (!chatInput.trim()) return;
		onSend(chatInput.trim());
		chatInput = '';
		replyTo = null;
	}

	function handleEmojiSelect(emoji: string) {
		chatInput += emoji;
	}

	function startReply(message: ChatMessage) {
		replyTo = message;
		contextMenu.show = false;
	}

	function startEdit(message: ChatMessage) {
		editingMessage = message;
		editContent = message.content;
		contextMenu.show = false;
	}

	function saveEdit() {
		if (editingMessage && editContent.trim()) {
			onEdit?.(editingMessage.id, editContent.trim());
		}
		editingMessage = null;
		editContent = '';
	}

	function cancelEdit() {
		editingMessage = null;
		editContent = '';
	}

	function handleContextMenu(e: MouseEvent, message: ChatMessage) {
		e.preventDefault();
		contextMenu = { show: true, x: e.clientX, y: e.clientY, message };
	}

	function closeContextMenu() {
		contextMenu.show = false;
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Enter' && !e.shiftKey) {
			e.preventDefault();
			if (editingMessage) {
				saveEdit();
			} else {
				sendMessage();
			}
		}
		if (e.key === 'Escape') {
			replyTo = null;
			cancelEdit();
			closeContextMenu();
		}
	}
</script>

<svelte:window onclick={closeContextMenu} />

<div class="w-[280px] flex flex-col shrink-0 border-l" style="background-color: #16213e; border-color: #2a2a4a;">
	<!-- Header -->
	<div class="px-3 py-2.5 border-b flex items-center justify-between" style="border-color: #2a2a4a;">
		<h3 class="font-bold text-xs text-gray-300">گفتگو</h3>
		<button onclick={onClose} class="text-gray-400 hover:text-white p-1" aria-label="بستن">
			<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" /></svg>
		</button>
	</div>

	<!-- Pinned Messages -->
	{#if messages.some(m => m.isPinned)}
		<div class="px-3 py-2 border-b bg-yellow-900/20" style="border-color: #2a2a4a;">
			<p class="text-[10px] text-yellow-400 font-bold mb-1">📌 پیام‌های سنجاق شده</p>
			{#each messages.filter(m => m.isPinned) as pinned}
				<p class="text-xs text-gray-300 truncate">"{pinned.content}" — {pinned.sender}</p>
			{/each}
		</div>
	{/if}

	<!-- Messages -->
	<div bind:this={chatContainer} class="flex-1 overflow-y-auto px-3 py-2 space-y-2">
		{#each messages as msg (msg.id)}
			<div
				class="group relative {msg.isOwn ? 'flex flex-col items-end' : ''}"
				oncontextmenu={(e) => handleContextMenu(e, msg)}
				role="article"
			>
				<!-- Reply reference -->
				{#if msg.replyTo}
					<div class="text-[10px] text-gray-500 mb-0.5 px-2">
						↩ پاسخ به {msg.replyTo.sender}: "{msg.replyTo.content}"
					</div>
				{/if}

				<!-- Message bubble -->
				<div class="max-w-[85%] {msg.isOwn ? 'bg-blue-600/30 text-gray-200' : 'bg-gray-700/50 text-gray-200'} rounded-2xl px-3 py-2 relative">
					{#if !msg.isOwn}
						<p class="text-[10px] font-bold text-blue-400 mb-0.5">{msg.sender}</p>
					{/if}

					{#if editingMessage?.id === msg.id}
						<textarea
							bind:value={editContent}
							class="w-full bg-gray-800 text-sm text-white rounded-lg p-2 outline-none resize-none border border-blue-500"
							rows="2"
							onkeydown={handleKeydown}
						></textarea>
						<div class="flex gap-1 mt-1">
							<button onclick={saveEdit} class="text-[10px] px-2 py-0.5 bg-blue-600 text-white rounded">ذخیره</button>
							<button onclick={cancelEdit} class="text-[10px] px-2 py-0.5 bg-gray-600 text-white rounded">لغو</button>
						</div>
					{:else}
						<p class="text-sm whitespace-pre-wrap break-words">{msg.content}</p>
					{/if}

					<div class="flex items-center gap-1.5 mt-0.5">
						<span class="text-[10px] text-gray-500">{msg.time}</span>
						{#if msg.isPinned}<span class="text-[10px]">📌</span>{/if}
						{#if msg.isEdited}<span class="text-[10px] text-gray-500">(ویرایش شده)</span>{/if}
					</div>

					<!-- Hover actions -->
					<div class="hidden group-hover:flex absolute -left-16 top-0 items-center gap-0.5">
						<button onclick={() => startReply(msg)} class="p-1 rounded hover:bg-gray-700 text-gray-500 hover:text-white" title="پاسخ">
							<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 10h10a8 8 0 018 8v2M3 10l6 6m-6-6l6-6" /></svg>
						</button>
						{#if msg.isOwn}
							<button onclick={() => startEdit(msg)} class="p-1 rounded hover:bg-gray-700 text-gray-500 hover:text-white" title="ویرایش">
								<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" /></svg>
							</button>
							<button onclick={() => onDelete?.(msg.id)} class="p-1 rounded hover:bg-gray-700 text-gray-500 hover:text-red-400" title="حذف">
								<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" /></svg>
							</button>
						{/if}
						{#if isAdmin}
							<button onclick={() => onPin?.(msg.id)} class="p-1 rounded hover:bg-gray-700 text-gray-500 hover:text-yellow-400" title={msg.isPinned ? 'برداشتن سنجاق' : 'سنجاق کردن'}>
								📌
							</button>
						{/if}
					</div>
				</div>
			</div>
		{/each}
	</div>

	<!-- Context Menu -->
	{#if contextMenu.show && contextMenu.message}
		<div
			class="fixed z-50 w-40 rounded-xl shadow-2xl py-1 overflow-hidden"
			style="background-color: #1e1e3a; border: 1px solid #2a2a4a; left: {contextMenu.x}px; top: {contextMenu.y}px;"
			role="menu"
		>
			<button onclick={() => startReply(contextMenu.message!)} class="w-full px-3 py-2 text-xs text-gray-300 hover:bg-white/5 flex items-center gap-2">
				↩ پاسخ
			</button>
			{#if contextMenu.message.isOwn}
				<button onclick={() => startEdit(contextMenu.message!)} class="w-full px-3 py-2 text-xs text-gray-300 hover:bg-white/5 flex items-center gap-2">
					✏️ ویرایش
				</button>
				<button onclick={() => { onDelete?.(contextMenu.message!.id); closeContextMenu(); }} class="w-full px-3 py-2 text-xs text-red-400 hover:bg-red-600/10 flex items-center gap-2">
					🗑️ حذف
				</button>
			{/if}
			{#if isAdmin}
				<button onclick={() => { onPin?.(contextMenu.message!.id); closeContextMenu(); }} class="w-full px-3 py-2 text-xs text-gray-300 hover:bg-white/5 flex items-center gap-2">
					📌 {contextMenu.message.isPinned ? 'برداشتن سنجاق' : 'سنجاق کردن'}
				</button>
			{/if}
		</div>
	{/if}

	<!-- Reply Preview -->
	{#if replyTo}
		<div class="px-3 py-2 border-t bg-blue-900/20" style="border-color: #2a2a4a;">
			<div class="flex items-center justify-between">
				<p class="text-[10px] text-blue-400">↩ پاسخ به {replyTo.sender}</p>
				<button onclick={() => replyTo = null} class="text-gray-500 hover:text-white">
					<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" /></svg>
				</button>
			</div>
			<p class="text-xs text-gray-400 truncate">{replyTo.content}</p>
		</div>
	{/if}

	<!-- Input -->
	<div class="px-3 py-2 border-t relative" style="border-color: #2a2a4a;">
		{#if showEmojiPicker}
			<EmojiPicker onSelect={handleEmojiSelect} onClose={() => showEmojiPicker = false} />
		{/if}

		<div class="flex gap-1.5 items-end">
			<button onclick={() => showEmojiPicker = !showEmojiPicker} class="p-2 text-gray-400 hover:text-white shrink-0" title="ایموجی">
				😀
			</button>
			<textarea
				bind:value={chatInput}
				class="flex-1 px-2.5 py-1.5 rounded-lg text-xs focus:ring-1 focus:ring-blue-500 outline-none resize-none"
				style="background-color: #2a2a4a;"
				placeholder="پیام..."
				rows="1"
				onkeydown={handleKeydown}
			></textarea>
			<button onclick={sendMessage} class="px-2.5 py-1.5 bg-blue-600 rounded-lg text-xs hover:bg-blue-700 shrink-0">
				<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8" /></svg>
			</button>
		</div>
	</div>
</div>
