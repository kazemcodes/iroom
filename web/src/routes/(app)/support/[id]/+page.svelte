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
	const statusColors: Record<string, string> = { open: 'bg-green-100 text-green-700', answered: 'bg-blue-100 text-blue-700', closed: 'bg-gray-100 text-gray-500' };
	const priorityLabels: Record<string, string> = { low: 'کم', normal: 'عادی', high: 'زیاد', urgent: 'فوری' };
	const priorityColors: Record<string, string> = { low: 'bg-gray-100 text-gray-500', normal: 'bg-blue-100 text-blue-700', high: 'bg-orange-100 text-orange-700', urgent: 'bg-red-100 text-red-700' };
	const priorityDotColors: Record<string, string> = { low: 'bg-gray-400', normal: 'bg-blue-500', high: 'bg-orange-500', urgent: 'bg-red-500' };
	const categoryLabels: Record<string, string> = { 'عمومی': 'عمومی', 'فنی': 'فنی', 'مالی': 'مالی' };
</script>

<div class="space-y-6">
	<div class="flex items-center justify-between">
		<div class="flex items-center gap-3">
			<a href="/support" class="p-2 hover:bg-gray-100 rounded-lg transition-colors">
				<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M15.75 19.5L8.25 12l7.5-7.5" /></svg>
			</a>
			{#if ticket}
				<div>
					<h1 class="text-xl font-bold text-gray-900">{ticket.title}</h1>
					<div class="flex items-center gap-2 mt-1">
						<span class="text-xs px-2.5 py-1 rounded-full font-medium {statusColors[ticket.status]}">{statusLabels[ticket.status]}</span>
						<span class="flex items-center gap-1.5 text-xs px-2.5 py-1 rounded-full font-medium {priorityColors[ticket.priority]}">
							<span class="w-2 h-2 rounded-full {priorityDotColors[ticket.priority]}"></span>
							{priorityLabels[ticket.priority]}
						</span>
						<span class="text-xs text-gray-400">{formatDate(ticket.created_at)}</span>
					</div>
				</div>
			{/if}
		</div>
		{#if ticket && ticket.status !== 'closed'}
			<button onclick={closeTicket} class="px-4 py-2 text-sm text-red-600 border border-red-200 rounded-xl hover:bg-red-50 transition-colors font-medium">
				بستن تیکت
			</button>
		{/if}
	</div>

	<!-- Ticket Metadata Card -->
	{#if ticket}
		<div class="bg-white border border-gray-200 rounded-xl p-4">
			<div class="grid grid-cols-3 gap-4">
				<div>
					<p class="text-xs text-gray-400 mb-1">وضعیت</p>
					<span class="text-xs px-2.5 py-1 rounded-full font-medium {statusColors[ticket.status]}">{statusLabels[ticket.status]}</span>
				</div>
				<div>
					<p class="text-xs text-gray-400 mb-1">اولویت</p>
					<span class="flex items-center gap-1.5 text-xs px-2.5 py-1 rounded-full font-medium {priorityColors[ticket.priority]}">
						<span class="w-2 h-2 rounded-full {priorityDotColors[ticket.priority]}"></span>
						{priorityLabels[ticket.priority]}
					</span>
				</div>
				<div>
					<p class="text-xs text-gray-400 mb-1">دسته‌بندی</p>
					<span class="text-xs px-2.5 py-1 rounded-full font-medium bg-purple-100 text-purple-700">{categoryLabels[ticket.category] || ticket.category}</span>
				</div>
			</div>
		</div>
	{/if}

	{#if loading}
		<div class="flex items-center justify-center py-20">
			<div class="animate-spin h-8 w-8 border-4 border-blue-600 border-t-transparent rounded-full"></div>
		</div>
	{:else if !ticket}
		<div class="text-center py-20 bg-white rounded-xl border">
			<p class="text-gray-500">تیکت یافت نشد</p>
		</div>
	{:else}
		<!-- Messages Thread -->
		<div class="space-y-3">
			{#each ticket.messages || [] as msg}
				<div class="flex {msg.is_admin ? 'justify-end' : 'justify-start'}">
					<div class="max-w-[70%] {msg.is_admin ? 'bg-blue-50 border border-blue-100' : 'bg-white border border-gray-200'} rounded-xl p-4">
						<div class="flex items-center gap-2 mb-1.5">
							<span class="text-xs font-bold {msg.is_admin ? 'text-blue-700' : 'text-gray-700'}">{msg.user_display_name}</span>
							{#if msg.is_admin}
								<span class="text-[10px] px-1.5 py-0.5 rounded-full bg-blue-100 text-blue-600 font-medium">ادمین</span>
							{/if}
							<span class="text-[10px] text-gray-400">{formatDate(msg.created_at)}</span>
						</div>
						<p class="text-sm text-gray-800 leading-relaxed whitespace-pre-wrap">{msg.content}</p>
					</div>
				</div>
			{/each}
		</div>

		<!-- Reply Form -->
		{#if ticket.status !== 'closed'}
			<div class="bg-white border border-gray-200 rounded-xl p-4">
				<textarea
					bind:value={replyText}
					class="w-full px-4 py-3 border border-gray-200 rounded-xl text-sm focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none resize-none bg-gray-50"
					rows="3"
					placeholder="پاسخ خود را بنویسید..."
				></textarea>
				
				<!-- File Attachment -->
				<input
					bind:this={fileInput}
					type="file"
					class="hidden"
					onchange={handleFileSelect}
				/>
				
				{#if selectedFile}
					<div class="mt-3 flex items-center gap-2 px-3 py-2 bg-blue-50 border border-blue-200 rounded-lg">
						<svg class="w-4 h-4 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.172 7l-6.586 6.586a2 2 0 102.828 2.828l6.414-6.586a4 4 0 00-5.656-5.656l-6.415 6.585a6 6 0 108.486 8.486L20.5 13" /></svg>
						<span class="text-xs text-blue-700 flex-1 truncate">{selectedFile.name}</span>
						<span class="text-xs text-blue-500">{formatFileSize(selectedFile.size)}</span>
						<button onclick={removeSelectedFile} class="p-1 hover:bg-blue-100 rounded transition-colors">
							<svg class="w-3 h-3 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" /></svg>
						</button>
					</div>
				{/if}
				
				<div class="flex justify-between items-center mt-3">
					<button onclick={triggerFilePicker} class="p-2 text-gray-500 hover:text-blue-600 hover:bg-blue-50 rounded-lg transition-colors" title="پیوست فایل">
						<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.172 7l-6.586 6.586a2 2 0 102.828 2.828l6.414-6.586a4 4 0 00-5.656-5.656l-6.415 6.585a6 6 0 108.486 8.486L20.5 13" /></svg>
					</button>
					<button onclick={sendReply} disabled={(!replyText.trim() && !selectedFile) || replying} class="px-5 py-2 bg-blue-600 text-white text-sm rounded-xl hover:bg-blue-700 disabled:opacity-50 transition-colors font-medium">
						{replying ? 'در حال ارسال...' : 'ارسال پاسخ'}
					</button>
				</div>
			</div>
		{/if}
	{/if}
</div>
