<script lang="ts">
	import { api } from '$lib/api';
	import { onMount } from 'svelte';
	import type { Ticket, TicketMessage } from '$lib/types';

	let tickets = $state<Ticket[]>([]);
	let total = $state(0);
	let currentPage = $state(1);
	let search = $state('');
	let statusFilter = $state('all');
	let priorityFilter = $state('all');
	let loading = $state(true);

	const perPage = 15;

	let selectedTicket = $state<Ticket | null>(null);
	let showDetail = $state(false);
	let replyContent = $state('');
	let sendingReply = $state(false);

	onMount(() => loadTickets());

	async function loadTickets() {
		loading = true;
		const params: Record<string, string> = { page: String(currentPage), per_page: String(perPage) };
		if (search) params.search = search;
		if (statusFilter !== 'all') params.status = statusFilter;
		if (priorityFilter !== 'all') params.priority = priorityFilter;
		const res = await api.get<{ items: Ticket[]; total: number }>('/admin/tickets', params);
		if (res.success && res.data) {
			tickets = res.data.items || [];
			total = res.data.total;
		}
		loading = false;
	}

	function searchTickets() {
		currentPage = 1;
		loadTickets();
	}

	async function openTicket(ticket: Ticket) {
		selectedTicket = ticket;
		const res = await api.get<Ticket>(`/tickets/${ticket.id}`);
		if (res.success && res.data) {
			selectedTicket = res.data;
		}
		showDetail = true;
	}

	async function sendReply() {
		if (!selectedTicket || !replyContent.trim()) return;
		sendingReply = true;
		const res = await api.post(`/tickets/${selectedTicket.id}/reply`, { content: replyContent.trim() });
		if (res.success) {
			replyContent = '';
			await openTicket(selectedTicket);
		}
		sendingReply = false;
	}

	async function closeTicket() {
		if (!selectedTicket) return;
		const res = await api.post(`/tickets/${selectedTicket.id}/close`);
		if (res.success) {
			showDetail = false;
			selectedTicket = null;
			await loadTickets();
		}
	}

	function toPersian(n: number): string {
		return n.toLocaleString('fa-IR');
	}

	function formatDate(dateStr: string): string {
		return new Date(dateStr).toLocaleDateString('fa-IR', { year: 'numeric', month: '2-digit', day: '2-digit' });
	}

	const statusLabels: Record<string, string> = { open: 'باز', answered: 'پاسخ داده شده', closed: 'بسته شده' };
	const statusColors: Record<string, string> = { open: 'bg-green-100 text-green-700', answered: 'bg-blue-100 text-blue-700', closed: 'bg-gray-100 text-gray-600' };
	const priorityLabels: Record<string, string> = { low: 'کم', normal: 'عادی', high: 'زیاد', urgent: 'فوری' };
	const priorityColors: Record<string, string> = { low: 'bg-gray-100 text-gray-600', normal: 'bg-blue-100 text-blue-700', high: 'bg-orange-100 text-orange-700', urgent: 'bg-red-100 text-red-700' };
</script>

<div class="space-y-6">
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-bold text-gray-900">مدیریت تیکت‌ها</h1>
			<p class="text-gray-500 mt-1">{toPersian(total)} تیکت</p>
		</div>
	</div>

	<div class="flex items-center gap-3 flex-wrap">
		<input type="text" bind:value={search} onkeydown={(e) => e.key === 'Enter' && searchTickets()} class="flex-1 min-w-[200px] px-4 py-2.5 border rounded-lg text-sm focus:ring-2 focus:ring-blue-500 outline-none bg-white" placeholder="جستجوی عنوان تیکت..." />
		<select bind:value={statusFilter} onchange={() => { currentPage = 1; loadTickets(); }} class="px-4 py-2.5 border rounded-lg text-sm focus:ring-2 focus:ring-blue-500 outline-none bg-white">
			<option value="all">همه وضعیت‌ها</option>
			<option value="open">باز</option>
			<option value="answered">پاسخ داده شده</option>
			<option value="closed">بسته شده</option>
		</select>
		<select bind:value={priorityFilter} onchange={() => { currentPage = 1; loadTickets(); }} class="px-4 py-2.5 border rounded-lg text-sm focus:ring-2 focus:ring-blue-500 outline-none bg-white">
			<option value="all">همه اولویت‌ها</option>
			<option value="low">کم</option>
			<option value="normal">عادی</option>
			<option value="high">زیاد</option>
			<option value="urgent">فوری</option>
		</select>
	</div>

	{#if loading}
		<div class="flex items-center justify-center py-12"><div class="animate-spin h-8 w-8 border-4 border-blue-600 border-t-transparent rounded-full"></div></div>
	{:else if tickets.length === 0}
		<div class="text-center py-20 bg-white rounded-xl">
			<p class="text-gray-500">تیکتی یافت نشد</p>
		</div>
	{:else}
		<div class="bg-white rounded-xl overflow-hidden">
			<table class="w-full text-sm">
				<thead class="bg-gray-50 border-b">
					<tr>
						<th class="px-5 py-3 text-right font-medium text-gray-600">عنوان</th>
						<th class="px-5 py-3 text-right font-medium text-gray-600">کاربر</th>
						<th class="px-5 py-3 text-right font-medium text-gray-600">دسته‌بندی</th>
						<th class="px-5 py-3 text-right font-medium text-gray-600">اولویت</th>
						<th class="px-5 py-3 text-right font-medium text-gray-600">وضعیت</th>
						<th class="px-5 py-3 text-right font-medium text-gray-600">تاریخ</th>
					</tr>
				</thead>
				<tbody class="divide-y">
					{#each tickets as ticket}
						<tr class="hover:bg-gray-50 cursor-pointer" onclick={() => openTicket(ticket)}>
							<td class="px-5 py-3 font-medium">{ticket.title}</td>
							<td class="px-5 py-3 text-gray-500">{ticket.user_display_name}</td>
							<td class="px-5 py-3 text-gray-500">{ticket.category}</td>
							<td class="px-5 py-3"><span class="text-xs px-2 py-1 rounded-full font-medium {priorityColors[ticket.priority]}">{priorityLabels[ticket.priority]}</span></td>
							<td class="px-5 py-3"><span class="text-xs px-2 py-1 rounded-full font-medium {statusColors[ticket.status]}">{statusLabels[ticket.status]}</span></td>
							<td class="px-5 py-3 text-gray-500 text-xs" dir="ltr">{formatDate(ticket.created_at)}</td>
						</tr>
					{/each}
				</tbody>
			</table>
			{#if total > perPage}
				<div class="px-5 py-3 border-t flex items-center justify-between text-sm text-gray-500">
					<span>{toPersian(total)} تیکت</span>
					<div class="flex gap-1">
						<button disabled={currentPage <= 1} onclick={() => { currentPage--; loadTickets(); }} class="px-3 py-1 border rounded hover:bg-gray-50 disabled:opacity-50">قبلی</button>
						<span class="px-3 py-1">صفحه {toPersian(currentPage)} از {toPersian(Math.ceil(total / perPage))}</span>
						<button disabled={currentPage >= Math.ceil(total / perPage)} onclick={() => { currentPage++; loadTickets(); }} class="px-3 py-1 border rounded hover:bg-gray-50 disabled:opacity-50">بعدی</button>
					</div>
				</div>
			{/if}
		</div>
	{/if}
</div>

{#if showDetail && selectedTicket}
	<div class="fixed inset-0 bg-black/40 z-50 flex items-center justify-center p-4" onclick={() => showDetail = false}>
		<div class="bg-white rounded-2xl w-full max-w-2xl max-h-[80vh] shadow-xl flex flex-col" onclick={(e) => e.stopPropagation()}>
			<div class="px-6 py-4 border-b flex items-center justify-between">
				<div>
					<h2 class="font-bold text-lg">{selectedTicket.title}</h2>
					<div class="flex items-center gap-2 mt-1">
						<span class="text-xs px-2 py-0.5 rounded-full font-medium {priorityColors[selectedTicket.priority]}">{priorityLabels[selectedTicket.priority]}</span>
						<span class="text-xs px-2 py-0.5 rounded-full font-medium {statusColors[selectedTicket.status]}">{statusLabels[selectedTicket.status]}</span>
						<span class="text-xs text-gray-500">{selectedTicket.user_display_name}</span>
					</div>
				</div>
				<button onclick={() => showDetail = false} class="text-gray-400 hover:text-gray-600">
					<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" /></svg>
				</button>
			</div>
			<div class="flex-1 overflow-y-auto px-6 py-4 space-y-3">
				{#each selectedTicket.messages || [] as msg}
					<div class="rounded-lg p-3 {msg.is_admin ? 'bg-blue-50 border border-blue-200' : 'bg-gray-50 border border-gray-200'}">
						<div class="flex items-center gap-2 mb-1">
							<span class="text-xs font-bold {msg.is_admin ? 'text-blue-600' : 'text-gray-700'}">{msg.user_display_name}</span>
							{#if msg.is_admin}<span class="text-[10px] px-1.5 py-0.5 bg-blue-100 text-blue-600 rounded">مدیر</span>{/if}
							<span class="text-[10px] text-gray-400" dir="ltr">{new Date(msg.created_at).toLocaleString('fa-IR')}</span>
						</div>
						<p class="text-sm text-gray-700 whitespace-pre-wrap">{msg.content}</p>
					</div>
				{/each}
				{#if !selectedTicket.messages?.length}
					<p class="text-center text-gray-400 text-sm py-4">هنوز پیامی ثبت نشده</p>
				{/if}
			</div>
			{#if selectedTicket.status !== 'closed'}
				<div class="px-6 py-4 border-t space-y-3">
					<textarea bind:value={replyContent} rows="3" class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-blue-500 outline-none resize-none" placeholder="پاسخ خود را بنویسید..."></textarea>
					<div class="flex items-center justify-between">
						<button onclick={closeTicket} class="px-4 py-2 text-sm text-red-600 hover:bg-red-50 rounded-lg border border-red-200">بستن تیکت</button>
						<button onclick={sendReply} disabled={!replyContent.trim() || sendingReply} class="px-4 py-2 bg-blue-600 text-white rounded-lg text-sm font-medium hover:bg-blue-700 disabled:opacity-50">
							{sendingReply ? 'در حال ارسال...' : 'ارسال پاسخ'}
						</button>
					</div>
				</div>
			{/if}
		</div>
	</div>
{/if}
