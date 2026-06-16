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

	function toPersianNumber(n: number): string {
		const persianDigits = ['۰', '۱', '۲', '۳', '۴', '۵', '۶', '۷', '۸', '۹'];
		return String(n).replace(/\d/g, d => persianDigits[Number(d)]);
	}

	function formatTime(d: string | null) {
		if (!d) return '—';
		return new Date(d).toLocaleTimeString('fa-IR', { hour: '2-digit', minute: '2-digit', second: '2-digit' });
	}

	function formatDuration(seconds: number) {
		const h = Math.floor(seconds / 3600);
		const m = Math.floor((seconds % 3600) / 60);
		const s = seconds % 60;
		if (h > 0) return `${toPersianNumber(h)} ساعت و ${toPersianNumber(m)} دقیقه`;
		if (m > 0) return `${toPersianNumber(m)} دقیقه و ${toPersianNumber(s)} ثانیه`;
		return `${toPersianNumber(s)} ثانیه`;
	}

	function formatDate(d: string) {
		if (!d) return '';
		return new Date(d).toLocaleDateString('fa-IR', { year: 'numeric', month: 'long', day: 'numeric', hour: '2-digit', minute: '2-digit' });
	}

	function exportCSV() {
		const headers = ['شرکت‌کننده', 'زمان ورود', 'زمان خروج', 'مدت (ثانیه)', 'آی‌پی'];
		const rows = logs.map(l => [
			l.user_display_name,
			l.joined_at || '',
			l.left_at || '',
			String(l.duration),
			l.ip_address
		]);
		const csvContent = [headers, ...rows].map(r => r.map(c => `"${c}"`).join(',')).join('\n');
		const bom = '\uFEFF';
		const blob = new Blob([bom + csvContent], { type: 'text/csv;charset=utf-8;' });
		const url = URL.createObjectURL(blob);
		const a = document.createElement('a');
		a.href = url;
		a.download = `session-${sessionId}-logs.csv`;
		a.click();
		URL.revokeObjectURL(url);
	}

	const totalDuration = $derived(logs.reduce((sum, l) => sum + l.duration, 0));
	const participantCount = $derived(new Set(logs.map(l => l.user_id)).size);

	const participantStats = $derived(() => {
		const map = new Map<number, { name: string; joinCount: number; totalDuration: number; firstJoin: string | null; lastLeft: string | null }>();
		for (const l of logs) {
			const existing = map.get(l.user_id);
			if (existing) {
				existing.joinCount++;
				existing.totalDuration += l.duration;
				if (l.joined_at && (!existing.firstJoin || l.joined_at < existing.firstJoin)) existing.firstJoin = l.joined_at;
				if (l.left_at && (!existing.lastLeft || l.left_at > existing.lastLeft)) existing.lastLeft = l.left_at;
			} else {
				map.set(l.user_id, {
					name: l.user_display_name,
					joinCount: 1,
					totalDuration: l.duration,
					firstJoin: l.joined_at,
					lastLeft: l.left_at
				});
			}
		}
		return Array.from(map.values());
	});
</script>

<div class="space-y-6">
	<div class="flex items-center justify-between">
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
		{#if !loading && logs.length > 0}
			<button onclick={exportCSV} class="px-4 py-2 text-sm text-blue-600 border border-blue-200 rounded-xl hover:bg-blue-50 transition-colors font-medium flex items-center gap-2">
				<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" /></svg>
				خروجی CSV
			</button>
		{/if}
	</div>

	<!-- Stats -->
	<div class="grid grid-cols-2 gap-4">
		<div class="bg-white border border-gray-200 rounded-xl p-4">
			<p class="text-xs text-gray-400 mb-1">تعداد شرکت‌کنندگان</p>
			<p class="text-2xl font-bold text-gray-900">{toPersianNumber(participantCount)}</p>
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
		<!-- Participant Summary Cards -->
		<div>
			<h2 class="text-sm font-semibold text-gray-700 mb-3">خلاصه شرکت‌کنندگان</h2>
			<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
				{#each participantStats() as stat}
					<div class="bg-white border border-gray-200 rounded-xl p-4">
						<div class="flex items-center gap-3 mb-3">
							<div class="w-9 h-9 rounded-full bg-blue-100 flex items-center justify-center text-blue-700 text-sm font-bold">
								{stat.name.charAt(0)}
							</div>
							<div class="flex-1 min-w-0">
								<p class="text-sm font-semibold text-gray-900 truncate">{stat.name}</p>
								<p class="text-xs text-gray-400">{toPersianNumber(stat.joinCount)} بار ورود</p>
							</div>
						</div>
						<div class="space-y-1.5 text-xs text-gray-500">
							<div class="flex justify-between">
								<span>مجموع مدت حضور:</span>
								<span class="font-medium text-gray-700">{formatDuration(stat.totalDuration)}</span>
							</div>
							<div class="flex justify-between">
								<span>اولین ورود:</span>
								<span class="font-medium text-gray-700">{formatTime(stat.firstJoin)}</span>
							</div>
							<div class="flex justify-between">
								<span>آخرین خروج:</span>
								<span class="font-medium text-gray-700">{stat.lastLeft ? formatTime(stat.lastLeft) : 'همچنان حاضر'}</span>
							</div>
						</div>
					</div>
				{/each}
			</div>
		</div>

		<!-- Timeline -->
		<div>
			<h2 class="text-sm font-semibold text-gray-700 mb-3">زمان‌بندی ورود و خروج</h2>
			<div class="bg-white border border-gray-200 rounded-xl p-5">
				<div class="relative">
					<div class="absolute right-[15px] top-0 bottom-0 w-0.5 bg-gray-200"></div>
					<div class="space-y-4">
						{#each logs as log, i}
							<div class="relative flex items-start gap-4">
								<div class="relative z-10 flex-shrink-0 w-8 h-8 rounded-full {log.left_at ? 'bg-gray-100 text-gray-500' : 'bg-green-100 text-green-600'} flex items-center justify-center text-xs font-bold">
									{#if log.left_at}
										<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" /></svg>
									{:else}
										<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 16l-4-4m0 0l4-4m-4 4h14m-5 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h7a3 3 0 013 3v1" /></svg>
									{/if}
								</div>
								<div class="flex-1 pb-1">
									<div class="flex items-center gap-2">
										<span class="text-sm font-semibold text-gray-900">{log.user_display_name}</span>
										{#if !log.left_at}
											<span class="text-[10px] px-1.5 py-0.5 rounded-full bg-green-100 text-green-600 font-medium">حاضر</span>
										{/if}
									</div>
									<div class="flex items-center gap-3 mt-1 text-xs text-gray-500">
										<span class="flex items-center gap-1">
											<svg class="w-3 h-3 text-green-500" fill="currentColor" viewBox="0 0 8 8"><circle cx="4" cy="4" r="4" /></svg>
											ورود: {formatTime(log.joined_at)}
										</span>
										<span class="text-gray-300">→</span>
										<span class="flex items-center gap-1">
											<svg class="w-3 h-3 {log.left_at ? 'text-red-400' : 'text-gray-300'}" fill="currentColor" viewBox="0 0 8 8"><circle cx="4" cy="4" r="4" /></svg>
											خروج: {log.left_at ? formatTime(log.left_at) : '—'}
										</span>
										<span class="text-gray-300">•</span>
										<span>{formatDuration(log.duration)}</span>
									</div>
									<p class="text-[10px] text-gray-400 mt-0.5 font-mono">{log.ip_address}</p>
								</div>
							</div>
						{/each}
					</div>
				</div>
			</div>
		</div>

		<!-- Detailed Table -->
		<div class="bg-white border border-gray-200 rounded-xl overflow-hidden">
			<table class="w-full">
				<thead>
					<tr class="border-b border-gray-100">
						<th class="text-right px-5 py-3 text-xs font-semibold text-gray-500">#</th>
						<th class="text-right px-5 py-3 text-xs font-semibold text-gray-500">شرکت‌کننده</th>
						<th class="text-right px-5 py-3 text-xs font-semibold text-gray-500">زمان ورود</th>
						<th class="text-right px-5 py-3 text-xs font-semibold text-gray-500">زمان خروج</th>
						<th class="text-right px-5 py-3 text-xs font-semibold text-gray-500">مدت</th>
						<th class="text-right px-5 py-3 text-xs font-semibold text-gray-500">آی‌پی</th>
					</tr>
				</thead>
				<tbody>
					{#each logs as log, i}
						<tr class="border-b border-gray-50 hover:bg-gray-50 transition-colors">
							<td class="px-5 py-3.5 text-sm text-gray-400">{toPersianNumber(i + 1)}</td>
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
