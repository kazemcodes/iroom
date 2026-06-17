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

<div class="max-w-2xl mx-auto space-y-6">
	<div class="flex items-center gap-3">
		<a href="/admin/channels" class="p-2 text-gray-400 hover:text-gray-600 rounded-lg hover:bg-gray-100">
			<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" /></svg>
		</a>
		<div>
			<h1 class="text-2xl font-bold text-gray-900">ویرایش کلاس</h1>
			<p class="text-gray-500 mt-1">{classData?.name || '...'}</p>
		</div>
	</div>

	{#if loading}
		<div class="flex items-center justify-center py-20">
			<div class="animate-spin h-8 w-8 border-4 border-blue-600 border-t-transparent rounded-full"></div>
		</div>
	{:else}
		<div class="bg-white rounded-xl p-6 space-y-5">
			<div>
				<label class="block text-sm font-medium text-gray-700 mb-1">نام کلاس</label>
				<input bind:value={formName} class="w-full px-4 py-2.5 border rounded-lg text-sm focus:ring-2 focus:ring-blue-500 outline-none" />
			</div>
			<div>
				<label class="block text-sm font-medium text-gray-700 mb-1">توضیحات</label>
				<textarea bind:value={formDescription} class="w-full px-4 py-2.5 border rounded-lg text-sm focus:ring-2 focus:ring-blue-500 outline-none" rows="3"></textarea>
			</div>
			<div class="grid grid-cols-2 gap-4">
				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">رنگ</label>
					<input type="color" bind:value={formColor} class="w-12 h-12 rounded-lg border cursor-pointer" />
				</div>
				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">حداکثر دانش‌آموز</label>
					<input type="number" bind:value={formMaxStudents} class="w-full px-4 py-2.5 border rounded-lg text-sm focus:ring-2 focus:ring-blue-500 outline-none" />
				</div>
			</div>
			{#if classData?.invite_code}
				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">کد دعوت</label>
					<div class="flex gap-2">
						<input value={classData.invite_code} readonly class="flex-1 px-4 py-2.5 border rounded-lg text-sm bg-gray-50 font-mono" />
						<button onclick={() => navigator.clipboard.writeText(classData.invite_code)} class="px-4 py-2.5 bg-gray-100 text-gray-700 rounded-lg text-sm hover:bg-gray-200">کپی</button>
					</div>
				</div>
			{/if}
			<div class="flex gap-3 pt-4 border-t">
				<button onclick={save} disabled={saving || !formName.trim()} class="px-6 py-2.5 bg-blue-600 text-white rounded-lg text-sm font-medium hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed">
					{saving ? 'در حال ذخیره...' : 'ذخیره تغییرات'}
				</button>
				<a href="/admin/channels" class="px-6 py-2.5 bg-gray-100 text-gray-700 rounded-lg text-sm font-medium hover:bg-gray-200">انصراف</a>
			</div>
		</div>
	{/if}
</div>
