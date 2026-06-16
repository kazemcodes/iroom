<script lang="ts">
	import { api } from '$lib/api';
	import { onMount } from 'svelte';
	import type { Ticket, TicketMessage } from '$lib/types';

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
		return new Date(d).toLocaleDateString('fa-IR', { year: 'numeric', month: 'long', day: 'numeric', hour: '2-digit', minute: '2-digit' });
	}

	function toPersianNumber(n: number): string {
		const persianDigits = ['۰', '۱', '۲', '۳', '۴', '۵', '۶', '۷', '۸', '۹'];
		return String(n).replace(/\d/g, d => persianDigits[Number(d)]);
	}

	const statusLabels: Record<string, string> = { open: 'باز', answered: 'پاسخ داده شده', closed: 'بسته شده' };
	const statusColors: Record<string, string> = { open: 'bg-green-100 text-green-700', answered: 'bg-blue-100 text-blue-700', closed: 'bg-gray-100 text-gray-500' };
	const priorityLabels: Record<string, string> = { low: 'کم', normal: 'عادی', high: 'زیاد', urgent: 'فوری' };
	const priorityColors: Record<string, string> = { low: 'bg-gray-100 text-gray-500', normal: 'bg-blue-100 text-blue-700', high: 'bg-orange-100 text-orange-700', urgent: 'bg-red-100 text-red-700' };
	const priorityDotColors: Record<string, string> = { low: 'bg-gray-400', normal: 'bg-blue-500', high: 'bg-orange-500', urgent: 'bg-red-500' };
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
			result = result.filter(t => t.title.toLowerCase().includes(q));
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

<div class="space-y-6">
	{#if ticketDetail}
		<!-- Ticket Detail View -->
		<div class="flex items-center justify-between">
			<div class="flex items-center gap-3">
				<button onclick={() => { ticketDetail = null; selectedTicket = null; }} class="p-2 hover:bg-gray-100 rounded-lg transition-colors">
					<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M15.75 19.5L8.25 12l7.5-7.5" /></svg>
				</button>
				<div>
					<h1 class="text-xl font-bold text-gray-900">{ticketDetail.title}</h1>
					<div class="flex items-center gap-2 mt-1">
						<span class="text-xs px-2.5 py-1 rounded-full font-medium {statusColors[ticketDetail.status]}">{statusLabels[ticketDetail.status]}</span>
						<span class="flex items-center gap-1.5 text-xs px-2.5 py-1 rounded-full font-medium {priorityColors[ticketDetail.priority]}">
							<span class="w-2 h-2 rounded-full {priorityDotColors[ticketDetail.priority]}"></span>
							{priorityLabels[ticketDetail.priority]}
						</span>
						<span class="text-xs text-gray-400">{formatDate(ticketDetail.created_at)}</span>
					</div>
				</div>
			</div>
			{#if ticketDetail.status !== 'closed'}
				<button onclick={closeTicket} class="px-4 py-2 text-sm text-red-600 border border-red-200 rounded-xl hover:bg-red-50 transition-colors font-medium">
					بستن تیکت
				</button>
			{/if}
		</div>

		<!-- Metadata Card -->
		<div class="bg-white border border-gray-200 rounded-xl p-4">
			<div class="grid grid-cols-3 gap-4">
				<div>
					<p class="text-xs text-gray-400 mb-1">وضعیت</p>
					<span class="text-xs px-2.5 py-1 rounded-full font-medium {statusColors[ticketDetail.status]}">{statusLabels[ticketDetail.status]}</span>
				</div>
				<div>
					<p class="text-xs text-gray-400 mb-1">اولویت</p>
					<span class="flex items-center gap-1.5 text-xs px-2.5 py-1 rounded-full font-medium {priorityColors[ticketDetail.priority]}">
						<span class="w-2 h-2 rounded-full {priorityDotColors[ticketDetail.priority]}"></span>
						{priorityLabels[ticketDetail.priority]}
					</span>
				</div>
				<div>
					<p class="text-xs text-gray-400 mb-1">دسته‌بندی</p>
					<span class="text-xs px-2.5 py-1 rounded-full font-medium bg-purple-100 text-purple-700">{categoryLabels[ticketDetail.category] || ticketDetail.category}</span>
				</div>
			</div>
		</div>

		<!-- Messages -->
		{#if detailLoading}
			<div class="flex items-center justify-center py-10">
				<div class="animate-spin h-6 w-6 border-4 border-blue-600 border-t-transparent rounded-full"></div>
			</div>
		{:else}
			<div class="space-y-3">
				{#each ticketDetail.messages || [] as msg}
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
		{/if}

		<!-- Reply Form -->
		{#if ticketDetail.status !== 'closed'}
			<div class="bg-white border border-gray-200 rounded-xl p-4">
				<textarea
					bind:value={replyText}
					class="w-full px-4 py-3 border border-gray-200 rounded-xl text-sm focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none resize-none bg-gray-50"
					rows="3"
					placeholder="پاسخ خود را بنویسید..."
				></textarea>
				<div class="flex justify-end mt-3">
					<button onclick={sendReply} disabled={!replyText.trim() || replying} class="px-5 py-2 bg-blue-600 text-white text-sm rounded-xl hover:bg-blue-700 disabled:opacity-50 transition-colors font-medium">
						{replying ? 'در حال ارسال...' : 'ارسال پاسخ'}
					</button>
				</div>
			</div>
		{/if}
	{:else}
		<!-- Ticket List View -->
		<div class="flex items-center justify-between">
			<div>
				<h1 class="text-2xl font-bold text-gray-900">پشتیبانی</h1>
				<p class="text-gray-500 mt-1">{toPersianNumber(tickets.length)} تیکت</p>
			</div>
			<button onclick={() => showCreate = true} class="px-4 py-2.5 bg-blue-600 text-white text-sm rounded-xl hover:bg-blue-700 transition-colors font-medium flex items-center gap-2">
				<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" /></svg>
				تیکت جدید
			</button>
		</div>

		{#if !loading && tickets.length > 0}
			<!-- Status Count Badges -->
			<div class="flex items-center gap-3">
				<button onclick={() => filterStatus = 'all'} class="px-3 py-1.5 text-xs rounded-full font-medium transition-colors {filterStatus === 'all' ? 'bg-gray-900 text-white' : 'bg-gray-100 text-gray-600 hover:bg-gray-200'}">
					همه ({toPersianNumber(tickets.length)})
				</button>
				<button onclick={() => filterStatus = 'open'} class="px-3 py-1.5 text-xs rounded-full font-medium transition-colors {filterStatus === 'open' ? 'bg-green-600 text-white' : 'bg-green-50 text-green-700 hover:bg-green-100'}">
					باز ({toPersianNumber(statusCounts().open)})
				</button>
				<button onclick={() => filterStatus = 'answered'} class="px-3 py-1.5 text-xs rounded-full font-medium transition-colors {filterStatus === 'answered' ? 'bg-blue-600 text-white' : 'bg-blue-50 text-blue-700 hover:bg-blue-100'}">
					پاسخ داده شده ({toPersianNumber(statusCounts().answered)})
				</button>
				<button onclick={() => filterStatus = 'closed'} class="px-3 py-1.5 text-xs rounded-full font-medium transition-colors {filterStatus === 'closed' ? 'bg-gray-600 text-white' : 'bg-gray-50 text-gray-600 hover:bg-gray-100'}">
					بسته شده ({toPersianNumber(statusCounts().closed)})
				</button>
			</div>

			<!-- Filters Row -->
			<div class="flex items-center gap-3">
				<div class="flex-1 relative">
					<svg class="absolute right-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" /></svg>
					<input bind:value={searchQuery} class="w-full pr-10 pl-4 py-2.5 border border-gray-200 rounded-xl text-sm focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none bg-white" placeholder="جستجو در عنوان..." />
				</div>
				<select bind:value={filterPriority} class="px-3 py-2.5 border border-gray-200 rounded-xl text-sm focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none bg-white">
					<option value="all">همه اولویت‌ها</option>
					<option value="low">کم</option>
					<option value="normal">عادی</option>
					<option value="high">زیاد</option>
					<option value="urgent">فوری</option>
				</select>
				<select bind:value={filterCategory} class="px-3 py-2.5 border border-gray-200 rounded-xl text-sm focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none bg-white">
					<option value="all">همه دسته‌ها</option>
					<option value="general">عمومی</option>
					<option value="technical">فنی</option>
					<option value="financial">مالی</option>
				</select>
			</div>
		{/if}

		{#if loading}
			<div class="flex items-center justify-center py-20">
				<div class="animate-spin h-8 w-8 border-4 border-blue-600 border-t-transparent rounded-full"></div>
			</div>
		{:else if tickets.length === 0}
			<div class="text-center py-20 bg-white rounded-xl border">
				<svg class="w-12 h-12 text-gray-300 mx-auto mb-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M20.25 8.511c.884.284 1.5 1.128 1.5 2.097v4.286c0 1.136-.847 2.1-1.98 2.193-.34.027-.68.052-1.02.072v3.091l-3-3c-1.354 0-2.694-.055-4.02-.163a2.115 2.115 0 01-.825-.242m9.345-8.334a2.126 2.126 0 00-.476-.095 48.64 48.64 0 00-8.048 0c-1.131.094-1.976.94-1.976 2.097v4.286c0 .837.46 1.58 1.155 1.951m9.345-8.334V6.637c0-1.621-1.152-3.026-2.76-3.235A48.455 48.455 0 0011.25 3c-2.115 0-4.198.137-6.24.402-1.608.209-2.76 1.614-2.76 3.235v6.226c0 1.621 1.152 3.026 2.76 3.235.577.075 1.157.14 1.74.194V21l4.155-4.155" /></svg>
				<p class="text-gray-500">تیکتی وجود ندارد</p>
			</div>
		{:else if filteredTickets().length === 0}
			<div class="text-center py-12 bg-white rounded-xl border">
				<p class="text-gray-500">تیکتی با این فیلترها یافت نشد</p>
			</div>
		{:else}
			<div class="space-y-2">
				{#each filteredTickets() as ticket}
					<button onclick={() => viewTicket(ticket)} class="w-full text-right bg-white border border-gray-200 rounded-xl p-4 hover:border-blue-200 hover:shadow-sm transition-all">
						<div class="flex items-center justify-between">
							<div class="flex-1 min-w-0">
								<h3 class="font-bold text-gray-900 truncate">{ticket.title}</h3>
								<p class="text-xs text-gray-400 mt-1">{formatDate(ticket.created_at)} • {categoryLabels[ticket.category] || ticket.category}</p>
							</div>
							<div class="flex items-center gap-2 mr-4">
								<span class="flex items-center gap-1.5 text-xs px-2.5 py-1 rounded-full font-medium {priorityColors[ticket.priority]}">
									<span class="w-2 h-2 rounded-full {priorityDotColors[ticket.priority]}"></span>
									{priorityLabels[ticket.priority]}
								</span>
								<span class="text-xs px-2.5 py-1 rounded-full font-medium {statusColors[ticket.status]}">{statusLabels[ticket.status]}</span>
							</div>
						</div>
					</button>
				{/each}
			</div>
		{/if}

		{#if totalPages > 1}
			<div class="flex items-center justify-between text-sm text-gray-500">
				<span>{totalTickets} تیکت</span>
				<div class="flex gap-1">
					<button disabled={currentPage <= 1} onclick={() => { currentPage--; loadTickets(); }} class="px-3 py-1 border rounded hover:bg-gray-50 disabled:opacity-50">قبلی</button>
					<span class="px-3 py-1">صفحه {currentPage} از {totalPages}</span>
					<button disabled={currentPage >= totalPages} onclick={() => { currentPage++; loadTickets(); }} class="px-3 py-1 border rounded hover:bg-gray-50 disabled:opacity-50">بعدی</button>
				</div>
			</div>
		{/if}

		<!-- Create Ticket Modal -->
		{#if showCreate}
			<div class="fixed inset-0 bg-black/30 backdrop-blur-sm z-50 flex items-center justify-center p-4" onclick={(e) => { if (e.target === e.currentTarget) showCreate = false; }} role="button" tabindex="-1">
				<div class="bg-white rounded-2xl shadow-2xl w-full max-w-lg p-6 space-y-4">
					<h2 class="text-lg font-bold text-gray-900">تیکت جدید</h2>
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">عنوان</label>
						<input bind:value={newTitle} class="w-full px-4 py-2.5 border border-gray-200 rounded-xl text-sm focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none" placeholder="عنوان تیکت" />
					</div>
					<div class="grid grid-cols-2 gap-3">
						<div>
							<label class="block text-sm font-medium text-gray-700 mb-1">دسته‌بندی</label>
							<select bind:value={newCategory} class="w-full px-4 py-2.5 border border-gray-200 rounded-xl text-sm focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none bg-white">
								<option>عمومی</option>
								<option>فنی</option>
								<option>مالی</option>
							</select>
						</div>
						<div>
							<label class="block text-sm font-medium text-gray-700 mb-1">اولویت</label>
							<select bind:value={newPriority} class="w-full px-4 py-2.5 border border-gray-200 rounded-xl text-sm focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none bg-white">
								<option value="low">کم</option>
								<option value="normal">عادی</option>
								<option value="high">زیاد</option>
								<option value="urgent">فوری</option>
							</select>
						</div>
					</div>
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">پیام</label>
						<textarea bind:value={newMessage} class="w-full px-4 py-3 border border-gray-200 rounded-xl text-sm focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none resize-none bg-gray-50" rows="4" placeholder="توضیحات مشکل خود..."></textarea>
					</div>
					<div class="flex justify-end gap-2 pt-2">
						<button onclick={() => showCreate = false} class="px-4 py-2.5 text-sm text-gray-600 hover:bg-gray-100 rounded-xl transition-colors font-medium">انصراف</button>
						<button onclick={createTicket} disabled={!newTitle.trim() || !newMessage.trim() || creating} class="px-5 py-2.5 bg-blue-600 text-white text-sm rounded-xl hover:bg-blue-700 disabled:opacity-50 transition-colors font-medium">
							{creating ? 'در حال ایجاد...' : 'ایجاد تیکت'}
						</button>
					</div>
				</div>
			</div>
		{/if}
	{/if}
</div>
