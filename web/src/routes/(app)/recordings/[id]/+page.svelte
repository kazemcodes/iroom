<script lang="ts">
	import { page } from '$app/state';
	import { api } from '$lib/api';
	import { onMount } from 'svelte';
	import type { Recording, Session } from '$lib/types';
	import { toPersianNum, toPersianDateTime } from '$lib/utils/persian';

	let recordings = $state<Recording[]>([]);
	let session = $state<Session | null>(null);
	let loading = $state(true);
	let activeRecording = $state<Recording | null>(null);

	const sessionId = $derived(page.params.id);

	onMount(async () => {
		const [recRes, sessRes] = await Promise.all([
			api.get<Recording[]>(`/sessions/${sessionId}/recordings`),
			api.get<Session>(`/sessions/${sessionId}`)
		]);
		if (recRes.success && recRes.data) recordings = Array.isArray(recRes.data) ? recRes.data : [];
		if (sessRes.success) session = sessRes.data!;
		loading = false;
	});

	function formatDurationPersian(secs: number) {
		const m = Math.floor(secs / 60);
		const s = secs % 60;
		return toPersianNum(`${m}:${s.toString().padStart(2, '0')}`);
	}

	function formatFileSize(bytes: number): string {
		if (bytes < 1024) {
			return `${toPersianNum(bytes)} بایت`;
		}
		if (bytes < 1048576) {
			// < 1 MB: show in KB
			const kb = Math.round(bytes / 1024);
			return `${toPersianNum(kb)} کیلوبایت`;
		}
		if (bytes < 104857600) {
			// 1-100 MB: show in MB with one decimal
			const mb = (bytes / 1048576).toFixed(1);
			return `${toPersianNum(mb)} مگابایت`;
		}
		// > 100 MB: show in MB as integer
		const mb = Math.round(bytes / 1048576);
		return `${toPersianNum(mb)} مگابایت`;
	}

	function formatDatePersian(d: string) {
		return toPersianDateTime(d);
	}
</script>

<div class="space-y-5">
	<div class="flex items-center gap-3">
		<a href="/sessions" class="sky-btn-icon">
			<svg width="20" height="20" fill="none" stroke="currentColor" stroke-width="1.75" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M15 6l-6 6 6 6"/></svg>
		</a>
		<div>
			<h1 class="sky-page-title">ضبط‌های جلسه</h1>
			<p class="sky-page-subtitle">{session?.title || ''} — {toPersianNum(recordings.length)} ضبط</p>
		</div>
	</div>

	{#if loading}
		<div class="flex items-center justify-center py-16"><svg class="sky-spinner lg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="color: var(--color-crystal-clear);"><path d="M12 2v4M12 18v4M4.93 4.93l2.83 2.83M16.24 16.24l2.83 2.83M2 12h4M18 12h4M4.93 19.07l2.83-2.83M16.24 7.76l2.83-2.83"/></svg></div>
	{:else if activeRecording}
		<!-- Player -->
		<div class="sky-card overflow-hidden">
			<div class="sky-card-header">
				<h2>{activeRecording.filename}</h2>
				<button onclick={() => activeRecording = null} class="sky-btn sky-btn-ghost" style="padding: 0.25rem 0.75rem; font-size: 12px;">بازگشت</button>
			</div>
			<div style="background: #000;">
				<video controls autoplay class="w-full max-h-[70vh]" src="{api.getBaseUrl()}/recordings/{activeRecording.id}/download"></video>
			</div>
			<div class="p-4 text-sm" style="color: var(--color-mystic-sea);">
				<span>{formatDatePersian(activeRecording.created_at)}</span>
				<span class="mx-2">•</span>
				<span>{formatFileSize(activeRecording.filesize)}</span>
				{#if activeRecording.duration > 0}
					<span class="mx-2">•</span>
					<span>{formatDurationPersian(activeRecording.duration)}</span>
				{/if}
			</div>
		</div>
	{:else if recordings.length === 0}
		<div class="sky-card"><div class="sky-empty">
			<div class="sky-empty-icon"><svg width="48" height="48" fill="none" stroke="currentColor" stroke-width="1" viewBox="0 0 24 24" style="color: var(--color-muted-mountain);"><circle cx="12" cy="12" r="10"/><circle cx="12" cy="12" r="3"/></svg></div>
			<p class="sky-empty-title">هنوز ضبطی ثبت نشده</p>
			<p class="sky-empty-desc">ضبط‌ها پس از پایان جلسه اینجا نمایش داده می‌شوند</p>
		</div></div>
	{:else}
		<div class="space-y-3">
			{#each recordings as rec}
				<div class="sky-list-card p-4 flex items-center justify-between cursor-pointer" onclick={() => activeRecording = rec} role="button" tabindex="0">
					<div class="flex items-center gap-4">
						<div class="w-12 h-12 rounded-xl flex items-center justify-center" style="background: rgba(224,82,82,0.12);">
							<svg width="24" height="24" fill="none" stroke="var(--color-fiery-passion)" stroke-width="1.5" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z"/><path stroke-linecap="round" stroke-linejoin="round" d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/></svg>
						</div>
						<div>
							<p class="font-semibold" style="color: var(--color-midnight-sky);">{rec.filename}</p>
							<p class="text-sm" style="color: var(--color-mystic-sea);">{formatDatePersian(rec.created_at)} • {formatFileSize(rec.filesize)}</p>
						</div>
					</div>
					<div class="flex items-center gap-3">
						{#if rec.duration > 0}<span class="text-xs" style="color: var(--color-moonlit-mist);">{formatDurationPersian(rec.duration)}</span>{/if}
						<span class="sky-badge {rec.status === 'ready' ? 'sky-badge-success' : 'sky-badge-warning'}">{rec.status === 'ready' ? 'آماده' : 'پردازش'}</span>
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>
