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
	const statusBadge: Record<string, string> = { open: 'sky-badge sky-badge-success', answered: 'sky-badge sky-badge-info', closed: 'sky-badge sky-badge-default' };
	const priorityLabels: Record<string, string> = { low: 'کم', normal: 'عادی', high: 'زیاد', urgent: 'فوری' };
	const priorityBadge: Record<string, string> = { low: 'sky-badge sky-badge-default', normal: 'sky-badge sky-badge-info', high: 'sky-badge sky-badge-warning', urgent: 'sky-badge sky-badge-danger' };
</script>

<div class="space-y-5">
	<div class="flex items-center justify-between">
		<div>
			<h1 class="sky-page-title">مدیریت تیکت‌ها</h1>
			<p class="sky-page-subtitle">{toPersian(total)} تیکت</p>
		</div>
	</div>

	<div class="flex items-center gap-3 flex-wrap">
		<div class="sky-search flex-1 min-w-[200px]">
			<div class="sky-search-icon"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg></div>
			<input type="text" bind:value={search} onkeydown={(e) => e.key === 'Enter' && searchTickets()} class="sky-input" style="padding-right: 2.5rem;" placeholder="جستجوی عنوان تیکت..." />
		</div>
		<select bind:value={statusFilter} onchange={() => { currentPage = 1; loadTickets(); }} class="sky-input" style="width:auto;min-width:140px;">
			<option value="all">همه وضعیت‌ها</option>
			<option value="open">باز</option>
			<option value="answered">پاسخ داده شده</option>
			<option value="closed">بسته شده</option>
		</select>
		<select bind:value={priorityFilter} onchange={() => { currentPage = 1; loadTickets(); }} class="sky-input" style="width:auto;min-width:140px;">
			<option value="all">همه اولویت‌ها</option>
			<option value="low">کم</option>
			<option value="normal">عادی</option>
			<option value="high">زیاد</option>
			<option value="urgent">فوری</option>
		</select>
	</div>

	{#if loading}
		<div class="flex items-center justify-center py-16"><svg class="sky-spinner lg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="color: var(--color-crystal-clear);"><path d="M12 2v4M12 18v4M4.93 4.93l2.83 2.83M16.24 16.24l2.83 2.83M2 12h4M18 12h4M4.93 19.07l2.83-2.83M16.24 7.76l2.83-2.83"/></svg></div>
	{:else if tickets.length === 0}
		<div class="sky-card"><div class="sky-empty"><p class="sky-empty-desc">تیکتی یافت نشد</p></div></div>
	{:else}
		<div class="sky-card overflow-hidden">
			<table class="sky-table">
				<thead><tr><th>عنوان</th><th>کاربر</th><th>دسته‌بندی</th><th>اولویت</th><th>وضعیت</th><th>تاریخ</th></tr></thead>
				<tbody>
					{#each tickets as ticket}
						<tr class="cursor-pointer" onclick={() => openTicket(ticket)}>
							<td class="font-semibold">{ticket.title}</td>
							<td style="color: var(--color-mystic-sea);">{ticket.user_display_name}</td>
							<td style="color: var(--color-mystic-sea);">{ticket.category}</td>
							<td><span class="{priorityBadge[ticket.priority]}">{priorityLabels[ticket.priority]}</span></td>
							<td><span class="{statusBadge[ticket.status]}">{statusLabels[ticket.status]}</span></td>
							<td style="color: var(--color-moonlit-mist);" dir="ltr">{formatDate(ticket.created_at)}</td>
						</tr>
					{/each}
				</tbody>
			</table>
			{#if total > perPage}
				<div class="px-5 py-3 flex items-center justify-between text-sm" style="border-top: 1px solid var(--color-zen-garden); color: var(--color-mystic-sea);">
					<span>{toPersian(total)} تیکت</span>
					<div class="sky-pagination">
						<button class="sky-page-btn" disabled={currentPage <= 1} onclick={() => { currentPage--; loadTickets(); }}>قبلی</button>
						<span class="sky-page-btn" style="cursor:default;">{toPersian(currentPage)}/{toPersian(Math.ceil(total / perPage))}</span>
						<button class="sky-page-btn" disabled={currentPage >= Math.ceil(total / perPage)} onclick={() => { currentPage++; loadTickets(); }}>بعدی</button>
					</div>
				</div>
			{/if}
		</div>
	{/if}
</div>

{#if showDetail && selectedTicket}
	<div class="modal-overlay" role="button" tabindex="-1" onclick={() => showDetail = false}>
		<div class="modal-content flex flex-col" style="max-width: 42rem; max-height: 80vh;" onclick={(e) => e.stopPropagation()}>
			<div class="sky-modal-header">
				<div>
					<h2>{selectedTicket.title}</h2>
					<div class="flex items-center gap-2 mt-1">
						<span class="{priorityBadge[selectedTicket.priority]}">{priorityLabels[selectedTicket.priority]}</span>
						<span class="{statusBadge[selectedTicket.status]}">{statusLabels[selectedTicket.status]}</span>
						<span class="text-xs" style="color: var(--color-mystic-sea);">{selectedTicket.user_display_name}</span>
					</div>
				</div>
				<button onclick={() => showDetail = false} class="sky-btn-icon">
					<svg width="18" height="18" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
				</button>
			</div>
			<div class="flex-1 overflow-y-auto px-6 py-4 space-y-3">
				{#each selectedTicket.messages || [] as msg}
					<div class="rounded-lg p-3" style="background: {msg.is_admin ? 'var(--color-polar-ice)' : 'var(--color-secret-glow)'}; border: 1px solid var(--color-zen-garden);">
						<div class="flex items-center gap-2 mb-1">
							<span class="text-xs font-bold" style="color: {msg.is_admin ? 'var(--color-crystal-clear)' : 'var(--color-midnight-sky)'};">{msg.user_display_name}</span>
							{#if msg.is_admin}<span class="sky-badge sky-badge-info" style="font-size:10px;padding:1px 8px;">مدیر</span>{/if}
							<span class="text-[10px]" style="color: var(--color-moonlit-mist);" dir="ltr">{new Date(msg.created_at).toLocaleString('fa-IR')}</span>
						</div>
						<p class="text-sm whitespace-pre-wrap" style="color: var(--color-ocean-wave);">{msg.content}</p>
					</div>
				{/each}
				{#if !selectedTicket.messages?.length}
					<p class="text-center text-sm py-4" style="color: var(--color-moonlit-mist);">هنوز پیامی ثبت نشده</p>
				{/if}
			</div>
			{#if selectedTicket.status !== 'closed'}
				<div class="px-6 py-4 space-y-3" style="border-top: 1px solid var(--color-zen-garden);">
					<textarea bind:value={replyContent} rows="3" class="sky-input resize-none" placeholder="پاسخ خود را بنویسید..."></textarea>
					<div class="flex items-center justify-between">
						<button onclick={closeTicket} class="sky-btn sky-btn-outline" style="color: var(--color-fiery-passion); border-color: rgba(224,82,82,0.3);">بستن تیکت</button>
						<button onclick={sendReply} disabled={!replyContent.trim() || sendingReply} class="sky-btn sky-btn-primary">
							{sendingReply ? 'در حال ارسال...' : 'ارسال پاسخ'}
						</button>
					</div>
				</div>
			{/if}
		</div>
	</div>
{/if}
