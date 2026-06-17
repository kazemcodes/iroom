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

<div class="space-y-6">
	<div class="flex items-center gap-3">
		<a href="/sessions" class="text-gray-400 hover:text-gray-600">
			<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 5l7 7-7 7" /></svg>
		</a>
		<div>
			<h1 class="text-2xl font-bold text-gray-900">ضبط‌های جلسه</h1>
			<p class="text-gray-500 mt-1">{session?.title || ''} — {toPersianNum(recordings.length)} ضبط</p>
		</div>
	</div>

	{#if loading}
		<div class="flex items-center justify-center py-20">
			<div class="animate-spin h-8 w-8 border-4 border-blue-600 border-t-transparent rounded-full"></div>
		</div>
	{:else if activeRecording}
		<!-- Player -->
		<div class="bg-white rounded-xl overflow-hidden">
			<div class="p-4 border-b flex items-center justify-between">
				<h2 class="font-bold">{activeRecording.filename}</h2>
				<button onclick={() => activeRecording = null} class="text-sm text-gray-500 hover:text-gray-700">بازگشت</button>
			</div>
			<div class="bg-black">
				<video
					controls
					autoplay
					class="w-full max-h-[70vh]"
					src="{api.getBaseUrl()}/recordings/{activeRecording.id}/download"
				></video>
			</div>
			<div class="p-4 text-sm text-gray-500">
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
		<div class="text-center py-20 bg-white rounded-xl">
			<svg class="w-16 h-16 text-gray-400 mx-auto mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1" d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" />
			</svg>
			<p class="text-gray-500">هنوز ضبطی ثبت نشده</p>
			<p class="text-sm text-gray-400 mt-1">ضبط‌ها پس از پایان جلسه اینجا نمایش داده می‌شوند</p>
		</div>
	{:else}
		<div class="space-y-3">
			{#each recordings as rec}
				<div
					class="bg-white rounded-xl p-4 flex items-center justify-between hover:shadow-md transition-all cursor-pointer"
					onclick={() => activeRecording = rec}
					role="button"
					tabindex="0"
				>
					<div class="flex items-center gap-4">
						<div class="w-12 h-12 bg-red-100 rounded-xl flex items-center justify-center">
							<svg class="w-6 h-6 text-red-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z" />
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
							</svg>
						</div>
						<div>
							<p class="font-medium text-gray-900">{rec.filename}</p>
							<p class="text-sm text-gray-500">{formatDatePersian(rec.created_at)} • {formatFileSize(rec.filesize)}</p>
						</div>
					</div>
					<div class="flex items-center gap-3">
						{#if rec.duration > 0}
							<span class="text-xs text-gray-400">{formatDurationPersian(rec.duration)}</span>
						{/if}
						<span class="text-xs px-2 py-1 rounded-full {rec.status === 'ready' ? 'bg-green-100 text-green-700' : 'bg-yellow-100 text-yellow-700'}">
							{rec.status === 'ready' ? 'آماده' : 'پردازش'}
						</span>
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>
