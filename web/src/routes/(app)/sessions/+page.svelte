<script lang="ts">
	import { auth, isAdmin, isTeacher } from '$lib/stores';
	import { api } from '$lib/api';
	import { onMount } from 'svelte';
	import type { Session } from '$lib/types';
	import ConfirmModal from '$lib/components/ConfirmModal.svelte';
	import { classroomWindow } from '$lib/classroom/ClassroomWindow';
	import { toPersianNum, toPersianDateTime } from '$lib/utils/persian';

	let sessions = $state<Session[]>([]);
	let loading = $state(true);
	let search = $state('');
	let filter = $state<'all' | 'scheduled' | 'live' | 'ended'>('all');

	let currentPage = $state(1);
	let totalSessions = $state(0);
	let perPage = $state(10);
	let perPageInitialized = $state(false);

	$effect(() => {
		if (!perPageInitialized) {
			perPageInitialized = true;
			return;
		}
		perPage;
		currentPage = 1;
		loadSessions();
	});

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
		const res = await api.post<{ livekit_room?: string }>(`/sessions/${id}/start`);
		if (res.success) {
			sessions = sessions.map(s => s.id === id ? { ...s, status: 'live', livekit_room: res.data?.livekit_room } as any : s);
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
		return toPersianDateTime(d);
	}

	const filtered = $derived(
		filter === 'all' ? sessions : sessions.filter(s => s.status === filter)
	);

	const statusLabels: Record<string, string> = { scheduled: 'برنامه‌ریزی شده', live: 'در حال برگزاری', ended: 'پایان یافته' };
	const statusClasses: Record<string, string> = { scheduled: 'badge-info', live: 'badge-success', ended: 'badge' };
</script>

<div class="space-y-6">
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-bold" style="color: var(--sky-text-primary);">جلسات</h1>
			<p style="color: var(--sky-text-secondary);">{toPersianNum(totalSessions)} جلسه</p>
		</div>
	</div>

	<div class="flex flex-wrap gap-3">
		<div class="flex gap-1 p-1 rounded-lg" style="background: var(--sky-bg-dark);">
			{#each [['all', 'همه'], ['scheduled', 'برنامه‌ریزی شده'], ['live', 'در حال برگزاری'], ['ended', 'پایان یافته']] as [val, label]}
				<button
					class="px-3 py-1.5 rounded-md text-xs font-medium transition-all"
					style={filter === val ? 'background: var(--sky-bg-input); color: var(--sky-accent-blue);' : 'color: var(--sky-text-secondary);'}
					onclick={() => filter = val as any}
				>{label}</button>
			{/each}
		</div>
		<input
			type="text"
			bind:value={search}
			onkeydown={(e) => e.key === 'Enter' && loadSessions()}
			class="input-field flex-1 min-w-[200px]"
			placeholder="جستجو..."
		/>
	</div>

	{#if loading}
		<div class="flex items-center justify-center py-20">
			<div class="animate-spin h-8 w-8 border-4 border-blue-500 border-t-transparent rounded-full"></div>
		</div>
	{:else if filtered.length === 0}
		<div class="text-center py-20 card">
			<p style="color: var(--sky-text-secondary);">جلسه‌ای یافت نشد</p>
		</div>
	{:else}
		<div class="space-y-3">
			{#each filtered as s}
				<div class="card p-5">
					<div class="flex items-center justify-between">
						<div class="flex items-center gap-4">
							<div class="w-12 h-12 rounded-xl flex items-center justify-center" style="background: {s.status === 'live' ? 'rgba(0, 210, 106, 0.15)' : 'var(--sky-bg-input)'};">
								{#if s.status === 'live'}
									<div class="relative">
										<svg class="w-6 h-6" style="color: var(--sky-accent-green);" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" />
										</svg>
										<span class="absolute -top-1 -right-1 w-3 h-3 bg-red-500 rounded-full animate-pulse"></span>
									</div>
								{:else}
									<svg class="w-6 h-6" style="color: var(--sky-text-secondary);" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" />
									</svg>
								{/if}
							</div>
							<div>
								<h3 class="font-bold" style="color: var(--sky-text-primary);">{s.title}</h3>
								<p class="text-sm mt-0.5" style="color: var(--sky-text-secondary);">{formatDate(s.scheduled_at)} • {toPersianNum(s.duration)} دقیقه</p>
							</div>
						</div>
						<div class="flex items-center gap-2">
							<span class="text-xs px-2.5 py-1 rounded-full font-medium {statusClasses[s.status]}">{statusLabels[s.status]}</span>
							{#if s.status === 'live'}
								<button onclick={() => classroomWindow.open(String(s.id), s.title)} class="btn-primary" style="padding: 0.375rem 0.75rem; font-size: 0.75rem; background: linear-gradient(135deg, #10b981, #059669);">پیوستن</button>
							{/if}
							{#if ($isAdmin || $isTeacher) && s.status === 'scheduled'}
								<button onclick={() => startSession(s.id)} class="btn-primary" style="padding: 0.375rem 0.75rem; font-size: 0.75rem; background: linear-gradient(135deg, #10b981, #059669);">شروع</button>
							{/if}
							{#if ($isAdmin || $isTeacher) && s.status === 'live'}
								<button onclick={() => endSession(s.id)} class="btn-danger" style="padding: 0.375rem 0.75rem; font-size: 0.75rem;">پایان</button>
							{/if}
							<a href="/sessions/{s.id}/logs" class="btn-ghost" style="padding: 0.375rem 0.75rem; font-size: 0.75rem;">جزئیات</a>
							{#if $isAdmin}
								<button onclick={() => confirmDeleteSession(s.id)} class="btn-icon" style="width: 32px; height: 32px;">
									<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" /></svg>
								</button>
							{/if}
						</div>
					</div>
				</div>
			{/each}
		</div>
	{/if}

	{#if totalPages > 0}
		<div class="flex flex-col sm:flex-row items-center justify-between gap-4 text-sm" style="color: var(--sky-text-secondary);">
			<div class="flex items-center gap-3">
				<span>{toPersianNum(totalSessions)} جلسه</span>
				<div class="flex items-center gap-1">
					<span class="text-xs">نمایش:</span>
					<select bind:value={perPage} class="input-field" style="width: auto; padding: 0.25rem 0.5rem; font-size: 0.75rem;">
						<option value={10}>{toPersianNum(10)}</option>
						<option value={25}>{toPersianNum(25)}</option>
						<option value={50}>{toPersianNum(50)}</option>
					</select>
				</div>
			</div>
			<div class="flex items-center gap-1">
				<button disabled={currentPage <= 1} onclick={() => { currentPage--; loadSessions(); }} class="btn-ghost" style="padding: 0.375rem 0.75rem; font-size: 0.75rem;">قبلی</button>
				{#each Array.from({ length: totalPages }, (_, i) => i + 1) as p}
					<button
						onclick={() => { currentPage = p; loadSessions(); }}
						class="w-8 h-8 rounded-lg text-xs font-medium transition-all"
						style={currentPage === p ? 'background: var(--sky-primary); color: white;' : 'border: 1px solid var(--sky-border); color: var(--sky-text-secondary);'}
					>{toPersianNum(p)}</button>
				{/each}
				<button disabled={currentPage >= totalPages} onclick={() => { currentPage++; loadSessions(); }} class="btn-ghost" style="padding: 0.375rem 0.75rem; font-size: 0.75rem;">بعدی</button>
			</div>
			<span class="text-xs">صفحه {toPersianNum(currentPage)} از {toPersianNum(totalPages)}</span>
		</div>
	{/if}
</div>

<ConfirmModal bind:show={showDeleteConfirm} title="حذف جلسه" message="آیا از حذف این جلسه اطمینان دارید؟" onConfirm={() => deleteSession(deleteTargetId)} onCancel={() => {}} />
