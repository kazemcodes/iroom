<script lang="ts">
	import { auth, isAdmin, isTeacher } from '$lib/stores';
	import { api } from '$lib/api';
	import { onMount } from 'svelte';
	import type { Session } from '$lib/types';
	import ConfirmModal from '$lib/components/ConfirmModal.svelte';
	import { classroomWindow } from '$lib/classroom/ClassroomWindow';

	let sessions = $state<Session[]>([]);
	let loading = $state(true);
	let search = $state('');
	let filter = $state<'all' | 'scheduled' | 'live' | 'ended'>('all');

	let currentPage = $state(1);
	let totalSessions = $state(0);
	const perPage = 10;

	let showDeleteConfirm = $state(false);
	let deleteTargetId = $state(0);

	const totalPages = $derived(Math.ceil(totalSessions / perPage));

	onMount(() => loadSessions());

	async function loadSessions() {
		loading = true;
		const params: Record<string, string> = { page: String(currentPage), per_page: String(perPage) };
		if (search) params.search = search;
		const res = await api.get<{ items: Session[]; total: number }>('/sessions', params);
		if (res.success && res.data) {
			sessions = res.data.items || (Array.isArray(res.data) ? res.data : []);
			totalSessions = res.data.total || sessions.length;
		}
		loading = false;
	}

	async function startSession(id: number) {
		const res = await api.post(`/sessions/${id}/start`);
		if (res.success) {
			sessions = sessions.map(s => s.id === id ? { ...s, status: 'live', livekit_room: res.data!.livekit_room } : s);
		}
	}

	async function endSession(id: number) {
		const res = await api.post(`/sessions/${id}/end`);
		if (res.success) {
			sessions = sessions.map(s => s.id === id ? { ...s, status: 'ended' } : s);
		}
	}

	async function deleteSession(id: number) {
		const res = await api.delete(`/sessions/${id}`);
		if (res.success) sessions = sessions.filter(s => s.id !== id);
	}

	function confirmDeleteSession(id: number) {
		deleteTargetId = id;
		showDeleteConfirm = true;
	}

	function formatDate(d: string) {
		if (!d) return '';
		return new Date(d).toLocaleDateString('fa-IR', { year: 'numeric', month: 'long', day: 'numeric', hour: '2-digit', minute: '2-digit' });
	}

	const filtered = $derived(
		filter === 'all' ? sessions : sessions.filter(s => s.status === filter)
	);

	const statusLabels: Record<string, string> = { scheduled: 'برنامه‌ریزی شده', live: 'در حال برگزاری', ended: 'پایان یافته' };
	const statusColors: Record<string, string> = { scheduled: 'bg-blue-100 text-blue-700', live: 'bg-green-100 text-green-700', ended: 'bg-gray-100 text-gray-500' };
</script>

<div class="space-y-6">
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-bold text-gray-900">جلسات</h1>
			<p class="text-gray-500 mt-1">{totalSessions} جلسه</p>
		</div>
	</div>

	<!-- Filters -->
	<div class="flex flex-wrap gap-3">
		<div class="flex gap-1 bg-gray-100 p-1 rounded-lg">
			{#each [['all', 'همه'], ['scheduled', 'برنامه‌ریزی شده'], ['live', 'در حال برگزاری'], ['ended', 'پایان یافته']] as [val, label]}
				<button
					class="px-3 py-1.5 rounded-md text-xs font-medium transition-all {filter === val ? 'bg-white text-blue-600 shadow-sm' : 'text-gray-500 hover:text-gray-700'}"
					onclick={() => filter = val as any}
				>{label}</button>
			{/each}
		</div>
		<input
			type="text"
			bind:value={search}
			onkeydown={(e) => e.key === 'Enter' && loadSessions()}
			class="flex-1 min-w-[200px] px-4 py-2 border border-gray-200 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none bg-white"
			placeholder="جستجو..."
		/>
	</div>

	{#if loading}
		<div class="flex items-center justify-center py-20">
			<div class="animate-spin h-8 w-8 border-4 border-blue-600 border-t-transparent rounded-full"></div>
		</div>
	{:else if filtered.length === 0}
		<div class="text-center py-20 bg-white rounded-xl border">
			<p class="text-gray-500">جلسه‌ای یافت نشد</p>
		</div>
	{:else}
		<div class="space-y-3">
			{#each filtered as s}
				<div class="bg-white rounded-xl border p-5">
					<div class="flex items-center justify-between">
						<div class="flex items-center gap-4">
							<div class="w-12 h-12 rounded-xl flex items-center justify-center {s.status === 'live' ? 'bg-green-100' : 'bg-gray-100'}">
								{#if s.status === 'live'}
									<div class="relative">
										<svg class="w-6 h-6 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" />
										</svg>
										<span class="absolute -top-1 -right-1 w-3 h-3 bg-red-500 rounded-full animate-pulse"></span>
									</div>
								{:else}
									<svg class="w-6 h-6 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" />
									</svg>
								{/if}
							</div>
							<div>
								<h3 class="font-bold text-gray-900">{s.title}</h3>
								<p class="text-sm text-gray-500 mt-0.5">{formatDate(s.scheduled_at)} • {s.duration} دقیقه</p>
							</div>
						</div>
						<div class="flex items-center gap-2">
							<span class="text-xs px-2.5 py-1 rounded-full font-medium {statusColors[s.status]}">{statusLabels[s.status]}</span>
							{#if s.status === 'live'}
								<button onclick={() => classroomWindow.open(String(s.id), s.title)} class="px-3 py-1.5 bg-green-600 text-white text-xs rounded-lg hover:bg-green-700 transition-colors">پیوستن</button>
							{/if}
							{#if ($isAdmin || $isTeacher) && s.status === 'scheduled'}
								<button onclick={() => startSession(s.id)} class="px-3 py-1.5 bg-green-600 text-white text-xs rounded-lg hover:bg-green-700">شروع</button>
							{/if}
							{#if ($isAdmin || $isTeacher) && s.status === 'live'}
								<button onclick={() => endSession(s.id)} class="px-3 py-1.5 bg-red-600 text-white text-xs rounded-lg hover:bg-red-700">پایان</button>
							{/if}
						<a href="/sessions/{s.id}/logs" class="px-3 py-1.5 text-xs text-gray-600 border border-gray-200 rounded-lg hover:bg-gray-50 transition-colors font-medium">
							جزئیات
						</a>
						{#if $isAdmin}
							<button onclick={() => confirmDeleteSession(s.id)} class="p-1.5 text-gray-400 hover:text-red-600 rounded-lg hover:bg-red-50 transition-colors">
								<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" /></svg>
							</button>
						{/if}
						</div>
					</div>
				</div>
			{/each}
		</div>
	{/if}

	{#if totalPages > 1}
		<div class="flex items-center justify-between text-sm text-gray-500">
			<span>{totalSessions} جلسه</span>
			<div class="flex gap-1">
				<button disabled={currentPage <= 1} onclick={() => { currentPage--; loadSessions(); }} class="px-3 py-1 border rounded hover:bg-gray-50 disabled:opacity-50">قبلی</button>
				<span class="px-3 py-1">صفحه {currentPage} از {totalPages}</span>
				<button disabled={currentPage >= totalPages} onclick={() => { currentPage++; loadSessions(); }} class="px-3 py-1 border rounded hover:bg-gray-50 disabled:opacity-50">بعدی</button>
			</div>
		</div>
	{/if}
</div>

<ConfirmModal bind:show={showDeleteConfirm} title="حذف جلسه" message="آیا از حذف این جلسه اطمینان دارید؟" onConfirm={() => deleteSession(deleteTargetId)} onCancel={() => {}} />
