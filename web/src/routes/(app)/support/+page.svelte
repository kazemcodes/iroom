<script lang="ts">
	import { api } from '$lib/api';
	import { onMount } from 'svelte';
	import type { Ticket, TicketMessage } from '$lib/types';
	import { toPersianNum, toPersianDateTime } from '$lib/utils/persian';

	let tickets = $state<Ticket[]>([]);
	let loading = $state(true);
	let showCreate = $state(false);
	let selectedTicket = $state<Ticket | null>(null);
	let ticketDetail = $state<Ticket | null>(null);
	let detailLoading = $state(false);
	let replyText = $state('');
	let replying = $state(false);

	let newTitle = $state('');
	let newCategory = $state('عمومی');
	let newPriority = $state<'low' | 'normal' | 'high' | 'urgent'>('normal');
	let newMessage = $state('');
	let creating = $state(false);

	let filterStatus = $state<string>('all');
	let filterPriority = $state<string>('all');
	let filterCategory = $state<string>('all');
	let searchQuery = $state('');

	let currentPage = $state(1);
	let totalTickets = $state(0);
	const perPage = 20;

	const totalPages = $derived(Math.ceil(totalTickets / perPage));

	onMount(() => loadTickets());

	async function loadTickets() {
		loading = true;
		const params: Record<string, string> = { page: String(currentPage), per_page: String(perPage) };
		const res = await api.get<{ items: Ticket[]; total: number }>('/tickets', params);
		if (res.success && res.data) {
			tickets = res.data.items || (Array.isArray(res.data) ? res.data : []);
			totalTickets = res.data.total || tickets.length;
		}
		loading = false;
	}

	async function createTicket() {
		if (!newTitle.trim() || !newMessage.trim()) return;
		creating = true;
		const res = await api.post<Ticket>('/tickets', {
			title: newTitle,
			category: newCategory,
			priority: newPriority,
			message: newMessage
		});
		if (res.success && res.data) {
			tickets = [res.data, ...tickets];
			showCreate = false;
			newTitle = '';
			newMessage = '';
		}
		creating = false;
	}

	async function viewTicket(ticket: Ticket) {
		selectedTicket = ticket;
		detailLoading = true;
		const res = await api.get<{ ticket: Ticket; messages: TicketMessage[] }>(`/tickets/${ticket.id}`);
		if (res.success && res.data) {
			ticketDetail = { ...res.data.ticket, messages: res.data.messages };
		}
		detailLoading = false;
	}

	async function sendReply() {
		if (!replyText.trim() || !ticketDetail) return;
		replying = true;
		const res = await api.post<TicketMessage>(`/tickets/${ticketDetail.id}/reply`, {
			content: replyText
		});
		if (res.success && res.data && ticketDetail.messages) {
			ticketDetail.messages = [...ticketDetail.messages, res.data];
			replyText = '';
		}
		replying = false;
	}

	async function closeTicket() {
		if (!ticketDetail) return;
		const res = await api.post(`/tickets/${ticketDetail.id}/close`);
		if (res.success) {
			ticketDetail.status = 'closed';
			tickets = tickets.map(t => t.id === ticketDetail!.id ? { ...t, status: 'closed' } : t);
		}
	}

	function formatSize(bytes: number) {
		if (bytes < 1024) return bytes + ' B';
		if (bytes < 1048576) return (bytes / 1024).toFixed(1) + ' KB';
		return (bytes / 1048576).toFixed(1) + ' MB';
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

	const filteredTickets = $derived(() => {
		let result = tickets;
		if (filterStatus !== 'all') {
			result = result.filter(t => t.status === filterStatus);
		}
		if (filterPriority !== 'all') {
			result = result.filter(t => t.priority === filterPriority);
		}
		if (filterCategory !== 'all') {
			const catMap: Record<string, string> = { general: 'عمومی', technical: 'فنی', financial: 'مالی' };
			result = result.filter(t => t.category === (catMap[filterCategory] || filterCategory));
		}
		if (searchQuery.trim()) {
			const q = searchQuery.trim().toLowerCase();
			result = result.filter(t => 
				t.title.toLowerCase().includes(q) || 
				String(t.id).includes(q)
			);
		}
		return result;
	});

	const statusCounts = $derived(() => {
		const counts: Record<string, number> = { open: 0, answered: 0, closed: 0 };
		for (const t of tickets) {
			if (counts[t.status] !== undefined) counts[t.status]++;
		}
		return counts;
	});
</script>

<div class="space-y-5">
	{#if ticketDetail}
		<!-- Ticket Detail View -->
		<div class="flex items-center justify-between">
			<div class="flex items-center gap-3">
				<button onclick={() => { ticketDetail = null; selectedTicket = null; }} class="sky-btn-icon">
					<svg width="20" height="20" fill="none" stroke="currentColor" stroke-width="1.75" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M9 6l6 6-6 6"/></svg>
				</button>
				<div>
					<h1 class="sky-page-title">{ticketDetail.title}</h1>
					<div class="flex items-center gap-2 mt-1">
						<span class="{statusBadge[ticketDetail.status]}">{statusLabels[ticketDetail.status]}</span>
						<span class="{priorityBadge[ticketDetail.priority]}"><span class="dot"></span>{priorityLabels[ticketDetail.priority]}</span>
						<span class="text-xs" style="color: var(--color-moonlit-mist);">{formatDate(ticketDetail.created_at)}</span>
					</div>
				</div>
			</div>
			{#if ticketDetail.status !== 'closed'}
				<button onclick={closeTicket} class="sky-btn sky-btn-outline" style="color: var(--color-fiery-passion); border-color: rgba(224,82,82,0.3);">بستن تیکت</button>
			{/if}
		</div>

		<!-- Metadata Card -->
		<div class="sky-card p-4">
			<div class="grid grid-cols-3 gap-4">
				<div>
					<p class="text-xs mb-1" style="color: var(--color-moonlit-mist);">وضعیت</p>
					<span class="{statusBadge[ticketDetail.status]}">{statusLabels[ticketDetail.status]}</span>
				</div>
				<div>
					<p class="text-xs mb-1" style="color: var(--color-moonlit-mist);">اولویت</p>
					<span class="{priorityBadge[ticketDetail.priority]}"><span class="dot"></span>{priorityLabels[ticketDetail.priority]}</span>
				</div>
				<div>
					<p class="text-xs mb-1" style="color: var(--color-moonlit-mist);">دسته‌بندی</p>
					<span class="sky-badge sky-badge-info">{categoryLabels[ticketDetail.category] || ticketDetail.category}</span>
				</div>
			</div>
		</div>

		<!-- Messages -->
		{#if detailLoading}
			<div class="flex items-center justify-center py-10"><svg class="sky-spinner md" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="color: var(--color-crystal-clear);"><path d="M12 2v4M12 18v4M4.93 4.93l2.83 2.83M16.24 16.24l2.83 2.83M2 12h4M18 12h4M4.93 19.07l2.83-2.83M16.24 7.76l2.83-2.83"/></svg></div>
		{:else}
			<div class="space-y-3">
				{#each ticketDetail.messages || [] as msg}
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
		{/if}

		<!-- Reply Form -->
		{#if ticketDetail.status !== 'closed'}
			<div class="sky-card p-4">
				<textarea bind:value={replyText} class="sky-input resize-none" rows="3" placeholder="پاسخ خود را بنویسید..."></textarea>
				<div class="flex justify-end mt-3">
					<button onclick={sendReply} disabled={!replyText.trim() || replying} class="sky-btn sky-btn-primary">{replying ? 'در حال ارسال...' : 'ارسال پاسخ'}</button>
				</div>
			</div>
		{/if}
	{:else}
		<!-- Ticket List View -->
		<div class="flex items-center justify-between">
			<div>
				<h1 class="sky-page-title">پشتیبانی</h1>
				<p class="sky-page-subtitle">{toPersianNum(tickets.length)} تیکت</p>
			</div>
			<button onclick={() => showCreate = true} class="sky-btn sky-btn-primary flex items-center gap-2">
				<svg width="15" height="15" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
				تیکت جدید
			</button>
		</div>

		{#if !loading && tickets.length > 0}
			<!-- Status filter -->
			<div class="sky-filter-bar w-fit">
				<button onclick={() => filterStatus = 'all'} class="sky-filter-btn {filterStatus === 'all' ? 'active' : ''}">همه ({toPersianNum(tickets.length)})</button>
				<button onclick={() => filterStatus = 'open'} class="sky-filter-btn {filterStatus === 'open' ? 'active' : ''}">باز ({toPersianNum(statusCounts().open)})</button>
				<button onclick={() => filterStatus = 'answered'} class="sky-filter-btn {filterStatus === 'answered' ? 'active' : ''}">پاسخ داده شده ({toPersianNum(statusCounts().answered)})</button>
				<button onclick={() => filterStatus = 'closed'} class="sky-filter-btn {filterStatus === 'closed' ? 'active' : ''}">بسته شده ({toPersianNum(statusCounts().closed)})</button>
			</div>

			<!-- Filters Row -->
			<div class="flex items-center gap-3 flex-wrap">
				<div class="sky-search flex-1 min-w-[200px]">
					<div class="sky-search-icon"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg></div>
					<input bind:value={searchQuery} class="sky-input" placeholder="جستجو در عنوان یا شماره تیکت..." style="padding-right: 2.5rem;" />
				</div>
				<select bind:value={filterPriority} class="sky-input" style="width:auto;">
					<option value="all">همه اولویت‌ها</option>
					<option value="low">کم</option>
					<option value="normal">عادی</option>
					<option value="high">زیاد</option>
					<option value="urgent">فوری</option>
				</select>
				<select bind:value={filterCategory} class="sky-input" style="width:auto;">
					<option value="all">همه دسته‌ها</option>
					<option value="general">عمومی</option>
					<option value="technical">فنی</option>
					<option value="financial">مالی</option>
				</select>
			</div>
		{/if}

		{#if loading}
			<div class="flex items-center justify-center py-16"><svg class="sky-spinner lg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="color: var(--color-crystal-clear);"><path d="M12 2v4M12 18v4M4.93 4.93l2.83 2.83M16.24 16.24l2.83 2.83M2 12h4M18 12h4M4.93 19.07l2.83-2.83M16.24 7.76l2.83-2.83"/></svg></div>
		{:else if tickets.length === 0}
			<div class="sky-card"><div class="sky-empty"><div class="sky-empty-icon"><svg width="48" height="48" fill="none" stroke="currentColor" stroke-width="1" viewBox="0 0 24 24" style="color: var(--color-muted-mountain);"><path stroke-linecap="round" stroke-linejoin="round" d="M21 15a2 2 0 01-2 2H7l-4 4V5a2 2 0 012-2h14a2 2 0 012 2z"/></svg></div><p class="sky-empty-title">تیکتی وجود ندارد</p><p class="sky-empty-desc">برای ارتباط با پشتیبانی تیکت جدید بسازید</p></div></div>
		{:else if filteredTickets().length === 0}
			<div class="sky-card"><div class="sky-empty"><p class="sky-empty-desc">تیکتی با این فیلترها یافت نشد</p></div></div>
		{:else}
			<div class="space-y-2">
				{#each filteredTickets() as ticket}
					<button onclick={() => viewTicket(ticket)} class="sky-list-card w-full text-right p-4">
						<div class="flex items-center justify-between">
							<div class="flex-1 min-w-0">
								<h3 class="font-bold truncate" style="color: var(--color-midnight-sky);">{ticket.title}</h3>
								<p class="text-xs mt-1" style="color: var(--color-moonlit-mist);">{formatDate(ticket.created_at)} • {categoryLabels[ticket.category] || ticket.category}</p>
							</div>
							<div class="flex items-center gap-2 mr-4">
								<span class="{priorityBadge[ticket.priority]}"><span class="dot"></span>{priorityLabels[ticket.priority]}</span>
								<span class="{statusBadge[ticket.status]}">{statusLabels[ticket.status]}</span>
							</div>
						</div>
					</button>
				{/each}
			</div>
		{/if}

		{#if totalPages > 1}
			<div class="flex items-center justify-between text-sm" style="color: var(--color-mystic-sea);">
				<span>{toPersianNum(totalTickets)} تیکت</span>
				<div class="sky-pagination">
					<button class="sky-page-btn" disabled={currentPage <= 1} onclick={() => { currentPage--; loadTickets(); }}>قبلی</button>
					<span class="sky-page-btn" style="cursor:default;">{toPersianNum(currentPage)}/{toPersianNum(totalPages)}</span>
					<button class="sky-page-btn" disabled={currentPage >= totalPages} onclick={() => { currentPage++; loadTickets(); }}>بعدی</button>
				</div>
			</div>
		{/if}

		<!-- Create Ticket Modal -->
		{#if showCreate}
			<div class="modal-overlay" onclick={(e) => { if (e.target === e.currentTarget) showCreate = false; }} role="button" tabindex="-1">
				<div class="modal-content">
					<div class="sky-modal-header">
						<h2>تیکت جدید</h2>
						<button onclick={() => showCreate = false} class="sky-btn-icon"><svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg></button>
					</div>
					<div class="sky-modal-body space-y-4">
						<div><label class="sky-label">عنوان</label><input bind:value={newTitle} class="sky-input" placeholder="عنوان تیکت" /></div>
						<div class="grid grid-cols-2 gap-3">
							<div>
								<label class="sky-label">دسته‌بندی</label>
								<select bind:value={newCategory} class="sky-input"><option>عمومی</option><option>فنی</option><option>مالی</option></select>
							</div>
							<div>
								<label class="sky-label">اولویت</label>
								<select bind:value={newPriority} class="sky-input"><option value="low">کم</option><option value="normal">عادی</option><option value="high">زیاد</option><option value="urgent">فوری</option></select>
							</div>
						</div>
						<div><label class="sky-label">پیام</label><textarea bind:value={newMessage} class="sky-input resize-none" rows="4" placeholder="توضیحات مشکل خود..."></textarea></div>
					</div>
					<div class="sky-modal-footer">
						<button onclick={() => showCreate = false} class="sky-btn sky-btn-secondary">انصراف</button>
						<button onclick={createTicket} disabled={!newTitle.trim() || !newMessage.trim() || creating} class="sky-btn sky-btn-primary">{creating ? 'در حال ایجاد...' : 'ایجاد تیکت'}</button>
					</div>
				</div>
			</div>
		{/if}
	{/if}
</div>
