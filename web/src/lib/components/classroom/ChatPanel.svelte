<!--
  ChatPanel — Real-time chat panel for classroom.
  
  Features:
    - RTL (right-to-left) message alignment
    - Sender name shown for all messages
    - Long-press (500ms) or right-click to open context menu with reply
    - Reply preview bar
    - Auto-scroll to bottom on new messages
    - Dark theme matching Skyroom design
  
  Props:
    messages: Array of chat messages to display
    isAdmin: Whether current user is admin/teacher (affects message styling)
    onSend: Callback when user sends a message
    onClose: Callback when panel should close
-->
<script lang="ts">
	import type { ChatMessage } from '$lib/classroom/types';

	let {
		messages = [],
		isAdmin = false,
		onSend,
		onClose,
	}: {
		messages: ChatMessage[];
		isAdmin: boolean;
		onSend: (content: string) => void;
		onClose: () => void;
	} = $props();

	let chatInput = $state('');
	let replyTo = $state<ChatMessage | null>(null);
	let contextMenu = $state<{ show: boolean; x: number; y: number; message: ChatMessage | null }>({
		show: false, x: 0, y: 0, message: null
	});
	let chatContainer: HTMLDivElement;
	let longPressTimer: ReturnType<typeof setTimeout> | null = null;

	function sendMessage() {
		if (!chatInput.trim()) return;
		onSend(chatInput.trim());
		chatInput = '';
		replyTo = null;
	}

	function handleContextMenu(e: MouseEvent, message: ChatMessage) {
		e.preventDefault();
		contextMenu = { show: true, x: e.clientX, y: e.clientY, message };
	}

	function handleTouchStart(e: TouchEvent, message: ChatMessage) {
		longPressTimer = setTimeout(() => {
			const touch = e.touches[0];
			contextMenu = { show: true, x: touch.clientX, y: touch.clientY, message };
		}, 500);
	}

	function handleTouchEnd() {
		if (longPressTimer) { clearTimeout(longPressTimer); longPressTimer = null; }
	}

	function closeContextMenu() {
		contextMenu.show = false;
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Enter' && !e.shiftKey) {
			e.preventDefault();
			sendMessage();
		}
		if (e.key === 'Escape') {
			replyTo = null;
			closeContextMenu();
		}
	}

	function scrollToBottom() {
		if (chatContainer) {
			chatContainer.scrollTop = chatContainer.scrollHeight;
		}
	}

	$effect(() => {
		if (messages.length) {
			scrollToBottom();
		}
	});
</script>

<svelte:window onclick={closeContextMenu} />

<div class="chat-panel">
	<!-- Messages -->
	<div bind:this={chatContainer} class="chat-messages">
		{#each messages as msg (msg.id)}
			<div
				class="msg-row {msg.isOwn ? 'msg-own' : 'msg-other'}"
				oncontextmenu={(e) => handleContextMenu(e, msg)}
				ontouchstart={(e) => handleTouchStart(e, msg)}
				ontouchend={handleTouchEnd}
				ontouchcancel={handleTouchEnd}
			>
				{#if msg.replyTo}
					<div class="reply-preview">
						↩ پاسخ به {msg.replyTo.sender}: "{msg.replyTo.content}"
					</div>
				{/if}
				<div class="msg-bubble {msg.isOwn ? 'bubble-own' : 'bubble-other'}">
					{#if !msg.isOwn}
						<p class="msg-sender">{msg.sender}</p>
					{/if}
					<p class="msg-text">{msg.content}</p>
					<span class="msg-time {msg.isOwn ? 'time-own' : ''}">{msg.time}</span>
				</div>
			</div>
		{/each}
	</div>

	<!-- Context Menu -->
	{#if contextMenu.show && contextMenu.message}
		<div
			class="ctx-menu"
			style="left: {contextMenu.x}px; top: {contextMenu.y}px;"
			role="menu"
		>
			<div class="ctx-item" onclick={() => { replyTo = contextMenu.message; closeContextMenu(); }}>
				↩ پاسخ
			</div>
		</div>
	{/if}

	<!-- Reply Preview -->
	{#if replyTo}
		<div class="reply-bar">
			<div class="reply-info">
				<p class="reply-label">↩ پاسخ به {replyTo.sender}</p>
				<p class="reply-text">{replyTo.content}</p>
			</div>
			<button onclick={() => replyTo = null} class="reply-close">
				<svg width="14" height="14"><use xlink:href="#shape_clear"></use></svg>
			</button>
		</div>
	{/if}

	<!-- Input -->
	<div class="chat-input-area">
		<input
			type="text"
			bind:value={chatInput}
			class="chat-input"
			placeholder="پیام خود را وارد کنید"
			onkeydown={handleKeydown}
		/>
		<button onclick={sendMessage} disabled={!chatInput.trim()} class="send-btn" class:active={chatInput.trim()}>
			<svg width="16" height="16"><use xlink:href="#shape_send"></use></svg>
		</button>
	</div>
</div>

<!-- SVG send icon -->
<svg style="display:none;" xmlns="http://www.w3.org/2000/svg">
	<symbol id="shape_send" viewBox="0 0 24 24"><path d="M2.01 21L23 12 2.01 3 2 10l15 2-15 2z"/></symbol>
</svg>

<style>
	.chat-panel {
		display: flex;
		flex-direction: column;
		height: 100%;
		background: transparent;
	}

	.chat-messages {
		flex: 1;
		overflow-y: auto;
		padding: 8px;
		display: flex;
		flex-direction: column;
		gap: 6px;
	}

	.msg-row {
		display: flex;
		flex-direction: column;
		align-items: flex-end;
	}

	.msg-other {
		align-items: flex-end;
	}

	.msg-own {
		align-items: flex-end;
	}

	.reply-preview {
		font-size: 0.65rem;
		color: #8a8a96;
		padding: 2px 8px;
		margin-bottom: 2px;
	}

	.msg-bubble {
		max-width: 85%;
		padding: 6px 10px;
		border-radius: 10px;
		font-size: 0.8rem;
		line-height: 1.5;
		word-wrap: break-word;
	}

	.bubble-other {
		background: rgba(255,255,255,0.06);
		color: #e0e0e6;
		border-bottom-right-radius: 2px;
	}

	.bubble-own {
		background: #23b9d7;
		color: #fff;
		border-bottom-left-radius: 2px;
	}

	.msg-sender {
		font-size: 0.7rem;
		font-weight: 600;
		color: #23b9d7;
		margin-bottom: 2px;
	}

	.msg-text {
		margin: 0;
	}

	.msg-time {
		font-size: 0.6rem;
		color: rgba(255,255,255,0.4);
		margin-top: 2px;
		display: block;
		text-align: left;
	}

	.time-own {
		color: rgba(255,255,255,0.6);
	}

	.ctx-menu {
		position: fixed;
		background: #1c2a3a;
		border-radius: 8px;
		box-shadow: 0 8px 24px rgba(0,0,0,0.4);
		z-index: 200;
		min-width: 140px;
		padding: 4px 0;
		animation: fadeIn 0.12s ease;
	}

	@keyframes fadeIn { from { opacity: 0; transform: translateY(-4px); } to { opacity: 1; transform: translateY(0); } }

	.ctx-item {
		padding: 8px 14px;
		cursor: pointer;
		color: #e0e0e6;
		font-size: 0.8rem;
		transition: background 0.12s;
	}

	.ctx-item:hover { background: rgba(255,255,255,0.06); }

	.reply-bar {
		display: flex;
		align-items: center;
		padding: 6px 10px;
		background: rgba(35, 185, 215, 0.1);
		border-top: 1px solid rgba(255,255,255,0.06);
		gap: 8px;
	}

	.reply-info { flex: 1; min-width: 0; }
	.reply-label { font-size: 0.7rem; color: #23b9d7; margin: 0 0 2px; }
	.reply-text { font-size: 0.7rem; color: #8a8a96; margin: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }

	.reply-close {
		background: none;
		border: none;
		cursor: pointer;
		padding: 4px;
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.reply-close svg { fill: #8a8a96; }
	.reply-close:hover svg { fill: #e0e0e6; }

	.chat-input-area {
		display: flex;
		align-items: center;
		gap: 6px;
		padding: 8px;
		border-top: 1px solid rgba(255,255,255,0.06);
	}

	.chat-input {
		flex: 1;
		background: rgba(0,0,0,0.2);
		border: 1px solid rgba(255,255,255,0.08);
		border-radius: 8px;
		padding: 8px 12px;
		font-size: 0.8rem;
		color: #e0e0e6;
		font-family: inherit;
		outline: none;
		direction: rtl;
	}

	.chat-input::placeholder { color: #5a6070; }
	.chat-input:focus { border-color: #23b9d7; }

	.send-btn {
		width: 34px;
		height: 34px;
		border-radius: 8px;
		border: none;
		cursor: pointer;
		display: flex;
		align-items: center;
		justify-content: center;
		background: rgba(255,255,255,0.06);
		color: #5a6070;
		transition: all 0.15s;
	}

	.send-btn.active {
		background: #23b9d7;
		color: #fff;
	}

	.send-btn.active:hover { background: #1a9fc0; }

	.send-btn svg { fill: currentColor; }
</style>
