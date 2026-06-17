<script lang="ts">
	// @ts-nocheck
	import { page } from '$app/state';
	import { auth } from '$lib/stores';
	import { api } from '$lib/api';
	import { onMount, onDestroy } from 'svelte';
	import { formatDuration, toPersianNum } from '$lib/utils/persian';
	import { classroomWindow } from '$lib/classroom/ClassroomWindow';
	import type { Session } from '$lib/types';

	let session = $state<Session | null>(null);
	let loading = $state(true);

	const sessionId = $derived(page.params.id!);

	onMount(async () => {
		loading = true;
		const res = await api.get<Session>(`/sessions/${sessionId}`);
		if (res.success) session = res.data!;
		loading = false;
	});

	function openClassroom() {
		if (!session) return;
		classroomWindow.open(String(session.id), session.title);
	}

	function formatDate(d: string) {
		if (!d) return '';
		return new Date(d).toLocaleDateString('fa-IR', { year: 'numeric', month: 'long', day: 'numeric', hour: '2-digit', minute: '2-digit' });
	}
</script>

<div class="min-h-screen bg-gray-50">
	{#if loading}
		<div class="flex items-center justify-center py-20">
			<div class="animate-spin h-8 w-8 border-4 border-blue-600 border-t-transparent rounded-full"></div>
		</div>
	{:else if session}
		<div class="max-w-2xl mx-auto py-10 px-4">
			<!-- Session info card -->
			<div class="bg-white rounded-2xl shadow-sm overflow-hidden" style="box-shadow: 0 1px 3px rgba(0,0,0,0.06), 0 1px 2px rgba(0,0,0,0.04);">
				<div class="p-6">
					<div class="flex items-start justify-between">
						<div>
							<h1 class="text-xl font-bold text-gray-900">{session.title}</h1>
							<p class="text-sm text-gray-500 mt-1">{formatDate(session.scheduled_at)} — {session.duration} دقیقه</p>
						</div>
						{#if session.status === 'live'}
							<span class="flex items-center gap-1.5 text-xs text-green-600 bg-green-50 px-3 py-1.5 rounded-full font-medium">
								<span class="w-2 h-2 bg-green-500 rounded-full animate-pulse"></span>
								در حال برگزاری
							</span>
						{:else if session.status === 'scheduled'}
							<span class="text-xs text-blue-600 bg-blue-50 px-3 py-1.5 rounded-full font-medium">برنامه‌ریزی شده</span>
						{:else}
							<span class="text-xs text-gray-500 bg-gray-100 px-3 py-1.5 rounded-full font-medium">پایان یافته</span>
						{/if}
					</div>

					<div class="mt-6">
						{#if session.status === 'live'}
							<button onclick={openClassroom} class="w-full py-3 bg-blue-600 text-white rounded-xl font-medium hover:bg-blue-700 transition-colors text-center">
								ورود به کلاس
							</button>
						{:else if session.status === 'scheduled'}
							<p class="text-sm text-gray-500 text-center py-3">جلسه هنوز شروع نشده</p>
						{:else}
							<p class="text-sm text-gray-500 text-center py-3">جلسه به پایان رسیده</p>
						{/if}
					</div>
				</div>
			</div>
		</div>
	{:else}
		<div class="flex-1 flex items-center justify-center py-20"><p class="text-gray-400">جلسه یافت نشد</p></div>
	{/if}
</div>
