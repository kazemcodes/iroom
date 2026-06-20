<script lang="ts">
	import type { ChatMessage } from '$lib/classroom/types';

	let {
		messages = [],
		isAdmin = false,
		disabled = false,
		onSend,
		onClose,
	}: {
		messages: ChatMessage[];
		isAdmin: boolean;
		disabled: boolean;
		onSend: (content: string, replyTo?: { sender: string; content: string }) => void;
		onClose: () => void;
	} = $props();

	let chatInput = $state('');
	let replyTo = $state<ChatMessage | null>(null);
	let chatContainer: HTMLDivElement;

	function sendMessage() {
		if (!chatInput.trim()) return;
		if (replyTo) {
			onSend(chatInput.trim(), { sender: replyTo.sender, content: replyTo.content });
		} else {
			onSend(chatInput.trim());
		}
		chatInput = '';
		replyTo = null;
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Enter' && !e.shiftKey) {
			e.preventDefault();
			sendMessage();
		}
		if (e.key === 'Escape') {
			replyTo = null;
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

<div class="chat-panel">
	<div bind:this={chatContainer} class="chat-messages">
		{#each messages as msg (msg.id)}
			<div class="msg-row">
				{#if msg.replyTo}
					<div class="reply-preview">
						↩ پاسخ به {msg.replyTo.sender}: "{msg.replyTo.content}"
					</div>
				{/if}
				<div class="msg-bubble">
				<div class="msg-bubble-top">
					<p class="msg-sender">{msg.sender}</p>
					<button class="reply-btn" onclick={() => { replyTo = msg; }} title="پاسخ">
						<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M9 17l-5-5 5-5"/><path d="M4 12h11a4 4 0 010 8h-1"/></svg>
					</button>
				</div>
					<p class="msg-text">{msg.content}</p>
					<span class="msg-time">{msg.time}</span>
				</div>
			</div>
		{/each}
	</div>

	{#if replyTo}
		<div class="reply-bar">
			<div class="reply-info">
				<p class="reply-label">↩ پاسخ به {replyTo.sender}</p>
				<p class="reply-text">{replyTo.content}</p>
			</div>
			<button onclick={() => replyTo = null} class="reply-close" title="لغو پاسخ">
				<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M18 6L6 18M6 6l12 12"/></svg>
			</button>
		</div>
	{/if}

	<div class="chat-input-area">
		{#if disabled}
			<p class="chat-disabled-msg">چت غیرفعال شده است</p>
		{:else}
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
		{/if}
	</div>
</div>

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
		align-items: flex-start;
	}

	.reply-preview {
		font-size: 0.65rem;
		color: #8a8a96;
		padding: 2px 8px;
		margin-bottom: 2px;
	}

	.msg-bubble {
		max-width: 100%;
		padding: 6px 10px;
		border-radius: 10px;
		font-size: 0.8rem;
		line-height: 1.5;
		word-wrap: break-word;
		background: rgba(255,255,255,0.06);
		color: #e0e0e6;
		border-bottom-left-radius: 2px;
		position: relative;
		width: 100%;
		box-sizing: border-box;
	}

	.msg-bubble-top {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 8px;
	}

	.reply-btn {
		background: none;
		border: none;
		cursor: pointer;
		padding: 2px;
		border-radius: 4px;
		display: flex;
		align-items: center;
		justify-content: center;
		color: #5a6070;
		opacity: 0;
		transition: opacity 0.15s, color 0.15s;
		flex-shrink: 0;
	}

	.msg-bubble:hover .reply-btn {
		opacity: 1;
	}

	.reply-btn:hover {
		color: #23b9d7;
	}

	.msg-sender {
		font-size: 0.7rem;
		font-weight: 600;
		color: #23b9d7;
		margin: 0 0 2px;
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
		color: #8a8a96;
		flex-shrink: 0;
	}

	.reply-close:hover { color: #e0e0e6; }

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

	.chat-disabled-msg {
		flex: 1;
		text-align: center;
		font-size: 0.75rem;
		color: #5a6070;
		padding: 8px 0;
	}
</style>
