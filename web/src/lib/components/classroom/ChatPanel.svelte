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

<div class="flex flex-col h-full" style="background-color: #252540;">
	<!-- Block Header -->
	<div class="flex items-center justify-between px-3 py-2.5 shrink-0" style="border-bottom: 1px solid #3a3a5a;">
		<div class="flex items-center gap-2">
			<svg class="w-4 h-4 text-[#94a3b8]" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" /></svg>
			<span class="text-xs font-medium text-[#94a3b8]">پیام‌ها</span>
		</div>
		<button onclick={onClose} class="text-[#94a3b8] hover:text-[#e2e8f0] p-1" aria-label="بستن">
			<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" /></svg>
		</button>
	</div>

	<!-- Pinned Messages -->
	{#if messages.some(m => m.isPinned)}
		<div class="px-3 py-2 shrink-0" style="background-color: rgba(215, 145, 29, 0.15); border-bottom: 1px solid #3a3a5a;">
			<p class="text-[10px] text-[#d7911d] font-medium mb-1">📌 پیام‌های سنجاق شده</p>
			{#each messages.filter(m => m.isPinned) as pinned}
				<p class="text-[11px] text-[#94a3b8] truncate">"{pinned.content}" — {pinned.sender}</p>
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
				{#if msg.replyTo}
					<div class="text-[10px] text-[#94a3b8] mb-0.5 px-2">
						↩ پاسخ به {msg.replyTo.sender}: "{msg.replyTo.content}"
					</div>
				{/if}

				<div class="max-w-[85%] {msg.isOwn ? 'bg-[#23b9d7] text-white' : 'bg-[#3a3a5a] text-[#e2e8f0]'} rounded-xl px-3 py-2 relative">
					{#if !msg.isOwn}
						<p class="text-[11px] font-medium text-[#23b9d7] mb-0.5">{msg.sender}</p>
					{/if}

					{#if editingMessage?.id === msg.id}
						<textarea
							bind:value={editContent}
							class="w-full bg-[#1a1a2e] text-[13px] text-[#e2e8f0] rounded-lg p-2 outline-none resize-none border border-[#23b9d7]"
							rows="2"
							onkeydown={handleKeydown}
						></textarea>
						<div class="flex gap-1 mt-1">
							<button onclick={saveEdit} class="text-[10px] px-2 py-0.5 bg-[#23b9d7] text-white rounded">ذخیره</button>
							<button onclick={cancelEdit} class="text-[10px] px-2 py-0.5 bg-[#3a3a5a] text-[#e2e8f0] rounded">لغو</button>
						</div>
					{:else}
						<p class="text-[13px] whitespace-pre-wrap break-words leading-relaxed">{msg.content}</p>
					{/if}

					<div class="flex items-center gap-1.5 mt-0.5">
						<span class="text-[10px] {msg.isOwn ? 'text-white/70' : 'text-[#94a3b8]'}">{msg.time}</span>
						{#if msg.isPinned}<span class="text-[10px]">📌</span>{/if}
						{#if msg.isEdited}<span class="text-[10px] {msg.isOwn ? 'text-white/70' : 'text-[#94a3b8]'}">(ویرایش شده)</span>{/if}
					</div>

					<!-- Hover actions -->
					<div class="hidden group-hover:flex absolute -left-16 top-0 items-center gap-0.5">
						<button onclick={() => startReply(msg)} class="p-1 rounded hover:bg-white/10 text-[#94a3b8] hover:text-[#e2e8f0]" title="پاسخ">
							<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 10h10a8 8 0 018 8v2M3 10l6 6m-6-6l6-6" /></svg>
						</button>
						{#if msg.isOwn}
							<button onclick={() => startEdit(msg)} class="p-1 rounded hover:bg-white/10 text-[#94a3b8] hover:text-[#e2e8f0]" title="ویرایش">
								<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" /></svg>
							</button>
							<button onclick={() => onDelete?.(msg.id)} class="p-1 rounded hover:bg-white/10 text-[#94a3b8] hover:text-[#e05252]" title="حذف">
								<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" /></svg>
							</button>
						{/if}
						{#if isAdmin}
							<button onclick={() => onPin?.(msg.id)} class="p-1 rounded hover:bg-white/10 text-[#94a3b8] hover:text-[#d7911d]" title={msg.isPinned ? 'برداشتن سنجاق' : 'سنجاق کردن'}>
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
			class="fixed z-50 w-40 rounded-lg overflow-hidden"
			style="background: #ffffff; border: 1px solid #e0e4eb; border-radius: 8px; box-shadow: 0 4px 20px rgba(0,0,0,0.3); left: {contextMenu.x}px; top: {contextMenu.y}px;"
			role="menu"
		>
			<button onclick={() => startReply(contextMenu.message!)} class="w-full px-3 py-2 text-[13px] text-[#1c293a] hover:bg-[#f0f2f5] flex items-center gap-2">
				↩ پاسخ
			</button>
			{#if contextMenu.message.isOwn}
				<button onclick={() => startEdit(contextMenu.message!)} class="w-full px-3 py-2 text-[13px] text-[#1c293a] hover:bg-[#f0f2f5] flex items-center gap-2">
					✏️ ویرایش
				</button>
				<button onclick={() => { onDelete?.(contextMenu.message!.id); closeContextMenu(); }} class="w-full px-3 py-2 text-[13px] text-[#e05252] hover:bg-[#fde8e8] flex items-center gap-2">
					🗑️ حذف
				</button>
			{/if}
			{#if isAdmin}
				<button onclick={() => { onPin?.(contextMenu.message!.id); closeContextMenu(); }} class="w-full px-3 py-2 text-[13px] text-[#1c293a] hover:bg-[#f0f2f5] flex items-center gap-2">
					📌 {contextMenu.message.isPinned ? 'برداشتن سنجاق' : 'سنجاق کردن'}
				</button>
			{/if}
		</div>
	{/if}

	<!-- Reply Preview -->
	{#if replyTo}
		<div class="px-3 py-2 shrink-0" style="background-color: rgba(35, 185, 215, 0.1); border-top: 1px solid #3a3a5a;">
			<div class="flex items-center justify-between">
				<p class="text-[10px] text-[#23b9d7]">↩ پاسخ به {replyTo.sender}</p>
				<button onclick={() => replyTo = null} class="text-[#94a3b8] hover:text-[#e2e8f0]">
					<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" /></svg>
				</button>
			</div>
			<p class="text-[11px] text-[#94a3b8] truncate">{replyTo.content}</p>
		</div>
	{/if}

	<!-- Input -->
	<div class="px-3 py-2 shrink-0 relative" style="border-top: 1px solid #3a3a5a;">
		{#if showEmojiPicker}
			<EmojiPicker onSelect={handleEmojiSelect} onClose={() => showEmojiPicker = false} />
		{/if}

		<div class="flex gap-1.5 items-end">
			<button onclick={() => showEmojiPicker = !showEmojiPicker} class="p-1.5 text-[#94a3b8] hover:text-[#e2e8f0] shrink-0" title="ایموجی">
				<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M14.828 14.828a4 4 0 01-5.656 0M9 10h.01M15 10h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
			</button>
			<input
				type="text"
				bind:value={chatInput}
				class="flex-1 px-3 py-2 rounded-lg text-[13px] focus:ring-1 focus:ring-[#23b9d7] outline-none"
				style="background-color: #1a1a2e; border: 1px solid #3a3a5a; color: #e2e8f0;"
				placeholder="پیام خود را وارد کنید"
				onkeydown={handleKeydown}
			/>
			<button onclick={sendMessage} disabled={!chatInput.trim()} class="p-2 rounded-lg shrink-0 transition-colors {chatInput.trim() ? 'bg-[#23b9d7] text-white hover:bg-[#1a9ad4]' : 'bg-[#3a3a5a] text-[#94a3b8]'}">
				<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8" /></svg>
			</button>
		</div>
	</div>
</div>
