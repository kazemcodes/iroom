<script lang="ts">
	import { api } from '$lib/api';
	import { page } from '$app/state';
	import { onMount } from 'svelte';
	import type { SessionLog, Session } from '$lib/types';

	let logs = $state<SessionLog[]>([]);
	let session = $state<Session | null>(null);
	let loading = $state(true);

	const sessionId = $derived(Number(page.params.id));

	onMount(async () => {
		await Promise.all([loadSession(), loadLogs()]);
	});

	async function loadSession() {
		const res = await api.get<Session>(`/sessions/${sessionId}`);
		if (res.success && res.data) {
			session = res.data;
		}
	}

	async function loadLogs() {
		loading = true;
		const res = await api.get<SessionLog[]>(`/sessions/${sessionId}/logs`);
		if (res.success && res.data) {
			logs = Array.isArray(res.data) ? res.data : [];
		}
		loading = false;
	}

	function formatTime(d: string | null) {
		if (!d) return '—';
		return new Date(d).toLocaleTimeString('fa-IR', { hour: '2-digit', minute: '2-digit', second: '2-digit' });
	}

	function formatDuration(seconds: number) {
		const h = Math.floor(seconds / 3600);
		const m = Math.floor((seconds % 3600) / 60);
		const s = seconds % 60;
		if (h > 0) return `${h} ساعت و ${m} دقیقه`;
		if (m > 0) return `${m} دقیقه و ${s} ثانیه`;
		return `${s} ثانیه`;
	}

	function formatDate(d: string) {
		if (!d) return '';
		return new Date(d).toLocaleDateString('fa-IR', { year: 'numeric', month: 'long', day: 'numeric', hour: '2-digit', minute: '2-digit' });
	}

	const totalDuration = $derived(logs.reduce((sum, l) => sum + l.duration, 0));
	const participantCount = $derived(new Set(logs.map(l => l.user_id)).size);
</script>

<div class="space-y-6">
	<div class="flex items-center gap-3">
		<a href="/sessions" class="p-2 hover:bg-gray-100 rounded-lg transition-colors">
			<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M15.75 19.5L8.25 12l7.5-7.5" /></svg>
		</a>
		<div>
			<h1 class="text-xl font-bold text-gray-900">گزارش جلسه</h1>
			{#if session}
				<p class="text-sm text-gray-500 mt-0.5">{session.title} • {formatDate(session.scheduled_at)}</p>
			{/if}
		</div>
	</div>

	<!-- Stats -->
	<div class="grid grid-cols-2 gap-4">
		<div class="bg-white border border-gray-200 rounded-xl p-4">
			<p class="text-xs text-gray-400 mb-1">تعداد شرکت‌کنندگان</p>
			<p class="text-2xl font-bold text-gray-900">{participantCount}</p>
		</div>
		<div class="bg-white border border-gray-200 rounded-xl p-4">
			<p class="text-xs text-gray-400 mb-1">مدت زمان کل</p>
			<p class="text-2xl font-bold text-gray-900">{formatDuration(totalDuration)}</p>
		</div>
	</div>

	{#if loading}
		<div class="flex items-center justify-center py-20">
			<div class="animate-spin h-8 w-8 border-4 border-blue-600 border-t-transparent rounded-full"></div>
		</div>
	{:else if logs.length === 0}
		<div class="text-center py-20 bg-white rounded-xl border">
			<p class="text-gray-500">لاگی ثبت نشده است</p>
		</div>
	{:else}
		<div class="bg-white border border-gray-200 rounded-xl overflow-hidden">
			<table class="w-full">
				<thead>
					<tr class="border-b border-gray-100">
						<th class="text-right px-5 py-3 text-xs font-semibold text-gray-500">شرکت‌کننده</th>
						<th class="text-right px-5 py-3 text-xs font-semibold text-gray-500">زمان ورود</th>
						<th class="text-right px-5 py-3 text-xs font-semibold text-gray-500">زمان خروج</th>
						<th class="text-right px-5 py-3 text-xs font-semibold text-gray-500">مدت</th>
						<th class="text-right px-5 py-3 text-xs font-semibold text-gray-500">آی‌پی</th>
					</tr>
				</thead>
				<tbody>
					{#each logs as log}
						<tr class="border-b border-gray-50 hover:bg-gray-50 transition-colors">
							<td class="px-5 py-3.5 text-sm font-medium text-gray-800">{log.user_display_name}</td>
							<td class="px-5 py-3.5 text-sm text-gray-500">{formatTime(log.joined_at)}</td>
							<td class="px-5 py-3.5 text-sm text-gray-500">
								{#if log.left_at}
									{formatTime(log.left_at)}
								{:else}
									<span class="text-green-600 font-medium">همچنان حاضر</span>
								{/if}
							</td>
							<td class="px-5 py-3.5 text-sm text-gray-500">{formatDuration(log.duration)}</td>
							<td class="px-5 py-3.5 text-sm text-gray-400 font-mono">{log.ip_address}</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	{/if}
</div>
