<script lang="ts">
	import { api } from '$lib/api';
	import { page } from '$app/state';
	import { onMount } from 'svelte';
	import type { SessionLog, Session } from '$lib/types';
	import { toPersianNum, toPersianDateTime } from '$lib/utils/persian';

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

	function formatDuration(seconds: number) {
		const h = Math.floor(seconds / 3600);
		const m = Math.floor((seconds % 3600) / 60);
		const s = seconds % 60;
		if (h > 0) return `${toPersianNum(h)} ساعت و ${toPersianNum(m)} دقیقه`;
		if (m > 0) return `${toPersianNum(m)} دقیقه و ${toPersianNum(s)} ثانیه`;
		return `${toPersianNum(s)} ثانیه`;
	}

	function formatDurationMinutes(seconds: number): string {
		const m = Math.round(seconds / 60);
		return toPersianNum(m);
	}

	function formatTime(d: string | null) {
		if (!d) return '—';
		return toPersianDateTime(new Date(d)).split(' - ')[1];
	}

	function formatDate(d: string) {
		if (!d) return '';
		return toPersianDateTime(new Date(d));
	}

	// Compute session start/end from all logs
	const sessionBounds = $derived(() => {
		if (logs.length === 0) return { start: 0, end: 0 };
		const allTimes = logs.flatMap(l => [l.joined_at, l.left_at].filter(Boolean).map(d => new Date(d!).getTime()));
		if (allTimes.length === 0) return { start: 0, end: 0 };
		return { start: Math.min(...allTimes), end: Math.max(...allTimes) };
	});

	const sessionDurationMs = $derived(() => {
		const b = sessionBounds();
		return b.end - b.start;
	});

	// Peak concurrent participants
	const peakConcurrent = $derived(() => {
		if (logs.length === 0) return 0;
		const events: { time: number; delta: number }[] = [];
		for (const l of logs) {
			if (l.joined_at) events.push({ time: new Date(l.joined_at).getTime(), delta: 1 });
			if (l.left_at) events.push({ time: new Date(l.left_at).getTime(), delta: -1 });
		}
		events.sort((a, b) => a.time - b.time);
		let current = 0;
		let peak = 0;
		for (const e of events) {
			current += e.delta;
			if (current > peak) peak = current;
		}
		return peak;
	});

	// Average attendance time
	const avgAttendance = $derived(() => {
		if (logs.length === 0) return 0;
		const uniqueParticipants = new Set(logs.map(l => l.user_id)).size;
		if (uniqueParticipants === 0) return 0;
		return Math.round(logs.reduce((sum, l) => sum + l.duration, 0) / uniqueParticipants);
	});

	function exportCSV() {
		const headers = ['نام', 'زمان ورود', 'زمان خروج', 'مدت زمان (دقیقه)'];
		const rows = logs.map(l => [
			l.user_display_name,
			l.joined_at ? formatDate(l.joined_at) : '',
			l.left_at ? formatDate(l.left_at) : '',
			formatDurationMinutes(l.duration)
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
		return Array.from(map.values()).sort((a, b) => {
			if (!a.firstJoin && !b.firstJoin) return 0;
			if (!a.firstJoin) return 1;
			if (!b.firstJoin) return -1;
			return new Date(a.firstJoin).getTime() - new Date(b.firstJoin).getTime();
		});
	});

	// Timeline bar data
	const timelineBars = $derived(() => {
		const bounds = sessionBounds();
		const totalMs = bounds.end - bounds.start;
		if (totalMs <= 0) return [];
		return participantStats().map(p => {
			const joinMs = p.firstJoin ? new Date(p.firstJoin).getTime() : bounds.start;
			const leaveMs = p.lastLeft ? new Date(p.lastLeft).getTime() : bounds.end;
			const leftPercent = ((joinMs - bounds.start) / totalMs) * 100;
			const widthPercent = ((leaveMs - joinMs) / totalMs) * 100;
			return {
				...p,
				leftPercent: Math.max(0, leftPercent),
				widthPercent: Math.max(0, Math.min(100 - leftPercent, widthPercent))
			};
		});
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

	{#if loading}
		<div class="flex items-center justify-center py-20">
			<div class="animate-spin h-8 w-8 border-4 border-blue-600 border-t-transparent rounded-full"></div>
		</div>
	{:else if logs.length === 0}
		<div class="text-center py-20 bg-white rounded-xl border">
			<p class="text-gray-500">لاگی ثبت نشده است</p>
		</div>
	{:else}
		<!-- Session Summary -->
		<div class="bg-white border border-gray-200 rounded-xl p-5">
			<h2 class="text-sm font-semibold text-gray-700 mb-4">خلاصه جلسه</h2>
			<div class="grid grid-cols-2 md:grid-cols-4 gap-4">
				<div class="text-center p-3 bg-blue-50 rounded-lg">
					<p class="text-xs text-blue-600 mb-1">تعداد شرکت‌کنندگان</p>
					<p class="text-2xl font-bold text-blue-700">{toPersianNum(participantCount)}</p>
				</div>
				<div class="text-center p-3 bg-green-50 rounded-lg">
					<p class="text-xs text-green-600 mb-1">مدت زمان جلسه</p>
					<p class="text-2xl font-bold text-green-700">{formatDuration(Math.round(sessionDurationMs() / 1000))}</p>
				</div>
				<div class="text-center p-3 bg-purple-50 rounded-lg">
					<p class="text-xs text-purple-600 mb-1">میانگین زمان حضور</p>
					<p class="text-2xl font-bold text-purple-700">{formatDuration(avgAttendance())}</p>
				</div>
				<div class="text-center p-3 bg-orange-50 rounded-lg">
					<p class="text-xs text-orange-600 mb-1">بیشترین حضور همزمان</p>
					<p class="text-2xl font-bold text-orange-700">{toPersianNum(peakConcurrent())}</p>
				</div>
			</div>
		</div>

		<!-- Visual Timeline -->
		<div class="bg-white border border-gray-200 rounded-xl p-5">
			<h2 class="text-sm font-semibold text-gray-700 mb-4">زمان‌بندی بصری شرکت‌کنندگان</h2>
			{#if timelineBars().length > 0 && sessionDurationMs() > 0}
				<div class="space-y-3">
					{#each timelineBars() as bar}
						<div class="flex items-center gap-3">
							<div class="w-24 flex-shrink-0 text-xs font-medium text-gray-700 truncate text-left" title={bar.name}>
								{bar.name}
							</div>
							<div class="flex-1 relative h-8 bg-gray-100 rounded-full overflow-hidden">
								<div
									class="absolute top-0 h-full bg-gradient-to-r from-green-400 to-green-500 rounded-full flex items-center justify-center"
									style="left: {bar.leftPercent}%; width: {bar.widthPercent}%"
								>
									{#if bar.widthPercent > 15}
										<span class="text-[10px] text-white font-medium px-1">
											{formatDuration(bar.totalDuration)}
										</span>
									{/if}
								</div>
							</div>
							<div class="w-40 flex-shrink-0 text-[10px] text-gray-500 text-left">
								<span class="text-green-600">↑ {formatTime(bar.firstJoin)}</span>
								<span class="mx-1">|</span>
								<span class="text-red-500">↓ {bar.lastLeft ? formatTime(bar.lastLeft) : 'حاضر'}</span>
							</div>
						</div>
					{/each}
				</div>
				<div class="mt-4 flex items-center justify-between text-[10px] text-gray-400 px-24">
					<span>شروع: {formatTime(logs[0]?.joined_at ?? null)}</span>
					<span>پایان: {formatTime(participantStats()[participantStats().length - 1]?.lastLeft ?? null)}</span>
				</div>
			{:else}
				<p class="text-sm text-gray-400 text-center py-4">داده‌ای برای نمایش وجود ندارد</p>
			{/if}
		</div>

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
								<p class="text-xs text-gray-400">{toPersianNum(stat.joinCount)} بار ورود</p>
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

		<!-- Event Timeline -->
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
							<td class="px-5 py-3.5 text-sm text-gray-400">{toPersianNum(i + 1)}</td>
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
