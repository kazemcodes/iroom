<script lang="ts">
	import { page } from '$app/state';
	import { onMount } from 'svelte';
	import { api } from '$lib/api';
	import { goto } from '$app/navigation';

	const channelId = $derived(page.params.id);

	let classData = $state<any>(null);
	let loading = $state(true);
	let saving = $state(false);

	let formName = $state('');
	let formDescription = $state('');
	let formColor = $state('#3B82F6');
	let formMaxStudents = $state(30);

	onMount(loadClass);

	async function loadClass() {
		loading = true;
		const res = await api.get(`/classes/${channelId}`);
		if (res.success && res.data) {
			classData = res.data;
			formName = classData.name;
			formDescription = classData.description || '';
			formColor = classData.color || '#3B82F6';
			formMaxStudents = classData.max_students || 30;
		}
		loading = false;
	}

	async function save() {
		saving = true;
		const res = await api.put(`/classes/${channelId}`, {
			name: formName, description: formDescription, color: formColor, max_students: formMaxStudents
		});
		saving = false;
		if (res.success) goto('/admin/channels');
	}
</script>

<div class="max-w-2xl mx-auto space-y-5">
	<div class="flex items-center gap-3">
		<a href="/admin/channels" class="sky-btn-icon">
			<svg width="20" height="20" fill="none" stroke="currentColor" stroke-width="1.75" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M15 6l-6 6 6 6"/></svg>
		</a>
		<div>
			<h1 class="sky-page-title">ویرایش کلاس</h1>
			<p class="sky-page-subtitle">{classData?.name || '...'}</p>
		</div>
	</div>

	{#if loading}
		<div class="flex items-center justify-center py-16"><svg class="sky-spinner lg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="color: var(--color-crystal-clear);"><path d="M12 2v4M12 18v4M4.93 4.93l2.83 2.83M16.24 16.24l2.83 2.83M2 12h4M18 12h4M4.93 19.07l2.83-2.83M16.24 7.76l2.83-2.83"/></svg></div>
	{:else}
		<div class="sky-card p-6 space-y-5">
			<div><label class="sky-label">نام کلاس</label><input bind:value={formName} class="sky-input" /></div>
			<div><label class="sky-label">توضیحات</label><textarea bind:value={formDescription} class="sky-input resize-none" rows="3"></textarea></div>
			<div class="grid grid-cols-2 gap-4">
				<div><label class="sky-label">رنگ</label><input type="color" bind:value={formColor} class="w-12 h-12 rounded-lg cursor-pointer" style="border: 1px solid var(--color-zen-garden);" /></div>
				<div><label class="sky-label">حداکثر دانش‌آموز</label><input type="number" bind:value={formMaxStudents} class="sky-input" /></div>
			</div>
			{#if classData?.invite_code}
				<div>
					<label class="sky-label">کد دعوت</label>
					<div class="flex gap-2">
						<input value={classData.invite_code} readonly class="sky-input flex-1 font-mono" style="background: var(--color-eternal-snow);" />
						<button onclick={() => navigator.clipboard.writeText(classData.invite_code)} class="sky-btn sky-btn-secondary">کپی</button>
					</div>
				</div>
			{/if}
			<div class="flex gap-3 pt-4" style="border-top: 1px solid var(--color-zen-garden);">
				<button onclick={save} disabled={saving || !formName.trim()} class="sky-btn sky-btn-primary">{saving ? 'در حال ذخیره...' : 'ذخیره تغییرات'}</button>
				<a href="/admin/channels" class="sky-btn sky-btn-secondary">انصراف</a>
			</div>
		</div>
	{/if}
</div>
