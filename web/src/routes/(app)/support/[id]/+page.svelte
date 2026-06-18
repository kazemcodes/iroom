<script lang="ts">
	import { api } from '$lib/api';
	import { page } from '$app/state';
	import { onMount } from 'svelte';
	import type { Ticket, TicketMessage } from '$lib/types';
	import { toPersianNum, toPersianDateTime } from '$lib/utils/persian';

	let ticket = $state<Ticket | null>(null);
	let loading = $state(true);
	let replyText = $state('');
	let replying = $state(false);
	let selectedFile = $state<File | null>(null);
	let fileInput = $state<HTMLInputElement | null>(null);

	const ticketId = $derived(Number(page.params.id));

	onMount(() => loadTicket());

	async function loadTicket() {
		loading = true;
		const res = await api.get<{ ticket: Ticket; messages: TicketMessage[] }>(`/tickets/${ticketId}`);
		if (res.success && res.data) {
			ticket = { ...res.data.ticket, messages: res.data.messages };
		}
		loading = false;
	}

	async function sendReply() {
		if ((!replyText.trim() && !selectedFile) || !ticket) return;
		replying = true;
		
		let res;
		if (selectedFile) {
			const formData = new FormData();
			formData.append('content', replyText);
			formData.append('file', selectedFile);
			res = await api.postFormData<TicketMessage>(`/tickets/${ticketId}/reply`, formData);
		} else {
			res = await api.post<TicketMessage>(`/tickets/${ticketId}/reply`, {
				content: replyText
			});
		}
		
		if (res.success && res.data && ticket.messages) {
			ticket.messages = [...ticket.messages, res.data];
			replyText = '';
			selectedFile = null;
		}
		replying = false;
	}

	async function closeTicket() {
		if (!ticket) return;
		const res = await api.post(`/tickets/${ticketId}/close`);
		if (res.success) {
			ticket.status = 'closed';
		}
	}

	function triggerFilePicker() {
		fileInput?.click();
	}

	function handleFileSelect(event: Event) {
		const input = event.target as HTMLInputElement;
		if (input.files && input.files.length > 0) {
			selectedFile = input.files[0];
		}
	}

	function removeSelectedFile() {
		selectedFile = null;
		if (fileInput) fileInput.value = '';
	}

	function formatFileSize(bytes: number): string {
		if (bytes < 1024) return toPersianNum(bytes) + ' B';
		if (bytes < 1048576) return toPersianNum((bytes / 1024).toFixed(1)) + ' KB';
		return toPersianNum((bytes / 1048576).toFixed(1)) + ' MB';
	}

	function formatDate(d: string) {
		if (!d) return '';
		return toPersianDateTime(d);
	}

	const statusLabels: Record<string, string> = { open: 'باز', answered: 'پاسخ داده شده', closed: 'بسته شده' };
	const statusBadge: Record<string, string> = { open: 'sky-badge sky-badge-success', answered: 'sky-badge sky-badge-info', closed: 'sky-badge sky-badge-default' };
	const priorityLabels: Record<string, string> = { low: 'کم', normal: 'عادی', high: 'زیاد', urgent: 'فوری' };
	const priorityBadge: Record<string, string> = { low: 'sky-badge sky-badge-default', normal: 'sky-badge sky-badge-info', high: 'sky-badge sky-badge-warning', urgent: 'sky-badge sky-badge-danger' };
	const categoryLabels: Record<string, string> = { 'عمومی': 'عمومی', 'فنی': 'فنی', 'مالی': 'مالی' };
</script>

<div class="space-y-5">
	<div class="flex items-center justify-between">
		<div class="flex items-center gap-3">
			<a href="/support" class="sky-btn-icon">
				<svg width="20" height="20" fill="none" stroke="currentColor" stroke-width="1.75" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M9 6l6 6-6 6"/></svg>
			</a>
			{#if ticket}
				<div>
					<h1 class="sky-page-title">{ticket.title}</h1>
					<div class="flex items-center gap-2 mt-1">
						<span class="{statusBadge[ticket.status]}">{statusLabels[ticket.status]}</span>
						<span class="{priorityBadge[ticket.priority]}"><span class="dot"></span>{priorityLabels[ticket.priority]}</span>
						<span class="text-xs" style="color: var(--color-moonlit-mist);">{formatDate(ticket.created_at)}</span>
					</div>
				</div>
			{/if}
		</div>
		{#if ticket && ticket.status !== 'closed'}
			<button onclick={closeTicket} class="sky-btn sky-btn-outline" style="color: var(--color-fiery-passion); border-color: rgba(224,82,82,0.3);">بستن تیکت</button>
		{/if}
	</div>

	<!-- Ticket Metadata Card -->
	{#if ticket}
		<div class="sky-card p-4">
			<div class="grid grid-cols-3 gap-4">
				<div>
					<p class="text-xs mb-1" style="color: var(--color-moonlit-mist);">وضعیت</p>
					<span class="{statusBadge[ticket.status]}">{statusLabels[ticket.status]}</span>
				</div>
				<div>
					<p class="text-xs mb-1" style="color: var(--color-moonlit-mist);">اولویت</p>
					<span class="{priorityBadge[ticket.priority]}"><span class="dot"></span>{priorityLabels[ticket.priority]}</span>
				</div>
				<div>
					<p class="text-xs mb-1" style="color: var(--color-moonlit-mist);">دسته‌بندی</p>
					<span class="sky-badge sky-badge-info">{categoryLabels[ticket.category] || ticket.category}</span>
				</div>
			</div>
		</div>
	{/if}

	{#if loading}
		<div class="flex items-center justify-center py-16"><svg class="sky-spinner lg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="color: var(--color-crystal-clear);"><path d="M12 2v4M12 18v4M4.93 4.93l2.83 2.83M16.24 16.24l2.83 2.83M2 12h4M18 12h4M4.93 19.07l2.83-2.83M16.24 7.76l2.83-2.83"/></svg></div>
	{:else if !ticket}
		<div class="sky-card"><div class="sky-empty"><p class="sky-empty-desc">تیکت یافت نشد</p></div></div>
	{:else}
		<!-- Messages Thread -->
		<div class="space-y-3">
			{#each ticket.messages || [] as msg}
				<div class="flex {msg.is_admin ? 'justify-end' : 'justify-start'}">
					<div class="max-w-[70%] rounded-xl p-4" style="background: {msg.is_admin ? 'var(--color-polar-ice)' : 'var(--color-pure)'}; border: 1px solid var(--color-zen-garden);">
						<div class="flex items-center gap-2 mb-1.5">
							<span class="text-xs font-bold" style="color: {msg.is_admin ? 'var(--color-crystal-clear)' : 'var(--color-midnight-sky)'};">{msg.user_display_name}</span>
							{#if msg.is_admin}<span class="sky-badge sky-badge-info" style="font-size:10px;padding:1px 8px;">ادمین</span>{/if}
							<span class="text-[10px]" style="color: var(--color-moonlit-mist);">{formatDate(msg.created_at)}</span>
						</div>
						<p class="text-sm leading-relaxed whitespace-pre-wrap" style="color: var(--color-ocean-wave);">{msg.content}</p>
					</div>
				</div>
			{/each}
		</div>

		<!-- Reply Form -->
		{#if ticket.status !== 'closed'}
			<div class="sky-card p-4">
				<textarea bind:value={replyText} class="sky-input resize-none" rows="3" placeholder="پاسخ خود را بنویسید..."></textarea>

				<input bind:this={fileInput} type="file" class="hidden" onchange={handleFileSelect} />

				{#if selectedFile}
					<div class="mt-3 flex items-center gap-2 px-3 py-2 rounded-lg" style="background: var(--color-polar-ice); border: 1px solid rgba(35,185,215,0.3);">
						<svg width="16" height="16" fill="none" stroke="var(--color-crystal-clear)" stroke-width="2" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M15.172 7l-6.586 6.586a2 2 0 102.828 2.828l6.414-6.586a4 4 0 00-5.656-5.656l-6.415 6.585a6 6 0 108.486 8.486L20.5 13"/></svg>
						<span class="text-xs flex-1 truncate" style="color: var(--color-crystal-clear);">{selectedFile.name}</span>
						<span class="text-xs" style="color: var(--color-mystic-sea);">{formatFileSize(selectedFile.size)}</span>
						<button onclick={removeSelectedFile} class="p-1 rounded transition-colors hover:bg-white/50">
							<svg width="12" height="12" fill="none" stroke="var(--color-crystal-clear)" stroke-width="2" viewBox="0 0 24 24"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
						</button>
					</div>
				{/if}

				<div class="flex justify-between items-center mt-3">
					<button onclick={triggerFilePicker} class="sky-btn-icon" title="پیوست فایل">
						<svg width="20" height="20" fill="none" stroke="currentColor" stroke-width="1.75" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M15.172 7l-6.586 6.586a2 2 0 102.828 2.828l6.414-6.586a4 4 0 00-5.656-5.656l-6.415 6.585a6 6 0 108.486 8.486L20.5 13"/></svg>
					</button>
					<button onclick={sendReply} disabled={(!replyText.trim() && !selectedFile) || replying} class="sky-btn sky-btn-primary">
						{replying ? 'در حال ارسال...' : 'ارسال پاسخ'}
					</button>
				</div>
			</div>
		{/if}
	{/if}
</div>
