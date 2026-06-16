<script lang="ts">
	import { auth, isAdmin, isTeacher } from '$lib/stores';
	import { api } from '$lib/api';
	import { onMount } from 'svelte';
	import type { Class } from '$lib/types';

	let classes = $state<Class[]>([]);
	let loading = $state(true);
	let showCreate = $state(false);
	let search = $state('');

	let currentPage = $state(1);
	let totalClasses = $state(0);
	const perPage = 12;

	// Form
	let formName = $state('');
	let formDesc = $state('');
	let formColor = $state('#3B82F6');
	let formMaxStudents = $state(30);
	let formLoading = $state(false);
	let formError = $state('');

	const colors = ['#3B82F6', '#10B981', '#F59E0B', '#EF4444', '#8B5CF6', '#EC4899', '#06B6D4'];

	const totalPages = $derived(Math.ceil(totalClasses / perPage));

	onMount(() => loadClasses());

	async function loadClasses() {
		loading = true;
		const params: Record<string, string> = { page: String(currentPage), per_page: String(perPage) };
		if (search) params.search = search;
		const res = await api.get<{ items: Class[]; total: number }>('/classes', params);
		if (res.success && res.data) {
			classes = res.data.items || (Array.isArray(res.data) ? res.data : []);
			totalClasses = res.data.total || classes.length;
		}
		loading = false;
	}

	async function createClass() {
		formLoading = true;
		formError = '';

		const res = await api.post<Class>('/classes', {
			name: formName,
			description: formDesc,
			color: formColor,
			max_students: formMaxStudents
		});

		if (!res.success) {
			formError = res.error || 'خطا در ایجاد کلاس';
			formLoading = false;
			return;
		}

		classes = [res.data!, ...classes];
		showCreate = false;
		formName = '';
		formDesc = '';
		formColor = '#3B82F6';
		formMaxStudents = 30;
		formLoading = false;
	}

	function handleSearch() {
		currentPage = 1;
		loadClasses();
	}
</script>

<div class="space-y-6">
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-bold text-gray-900">کلاس‌ها</h1>
			<p class="text-gray-500 mt-1">{totalClasses} کلاس</p>
		</div>
		{#if $isAdmin || $isTeacher}
			<button
				onclick={() => showCreate = true}
				class="px-4 py-2.5 bg-blue-600 text-white rounded-lg font-medium text-sm hover:bg-blue-700 transition-colors flex items-center gap-2"
			>
				<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
				</svg>
				ایجاد کلاس
			</button>
		{/if}
	</div>

	<!-- Search -->
	<div class="flex gap-3">
		<input
			type="text"
			bind:value={search}
			onkeydown={(e) => e.key === 'Enter' && handleSearch()}
			class="flex-1 px-4 py-2.5 border border-gray-200 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none bg-white text-sm"
			placeholder="جستجوی کلاس..."
		/>
		<button onclick={handleSearch} class="px-4 py-2.5 bg-gray-100 text-gray-700 rounded-lg text-sm hover:bg-gray-200 transition-colors">جستجو</button>
	</div>

	{#if loading}
		<div class="flex items-center justify-center py-20">
			<div class="animate-spin h-8 w-8 border-4 border-blue-600 border-t-transparent rounded-full"></div>
		</div>
	{:else if classes.length === 0}
		<div class="text-center py-20 bg-white rounded-xl border">
			<svg class="w-16 h-16 text-gray-300 mx-auto mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1" d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253" />
			</svg>
			<p class="text-gray-500 text-lg">کلاسی یافت نشد</p>
		</div>
	{:else}
		<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
			{#each classes as cls}
				<a href="/classes/{cls.id}" class="block bg-white rounded-xl border p-5 hover:shadow-lg transition-all group">
					<div class="flex items-start justify-between mb-3">
						<div class="flex items-center gap-3">
							<div class="w-4 h-4 rounded-full shrink-0" style="background-color: {cls.color}"></div>
							<h3 class="font-bold text-gray-900 group-hover:text-blue-600 transition-colors">{cls.name}</h3>
						</div>
					</div>
					{#if cls.description}
						<p class="text-sm text-gray-500 line-clamp-2 mb-4">{cls.description}</p>
					{/if}
					<div class="flex items-center justify-between text-xs text-gray-400 pt-3 border-t">
						<span class="flex items-center gap-1">
							<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z" />
							</svg>
							حداکثر {cls.max_students} نفر
						</span>
					</div>
				</a>
			{/each}
		</div>
	{/if}

	{#if totalPages > 1}
		<div class="flex items-center justify-between text-sm text-gray-500">
			<span>{totalClasses} کلاس</span>
			<div class="flex gap-1">
				<button disabled={currentPage <= 1} onclick={() => { currentPage--; loadClasses(); }} class="px-3 py-1 border rounded hover:bg-gray-50 disabled:opacity-50">قبلی</button>
				<span class="px-3 py-1">صفحه {currentPage} از {totalPages}</span>
				<button disabled={currentPage >= totalPages} onclick={() => { currentPage++; loadClasses(); }} class="px-3 py-1 border rounded hover:bg-gray-50 disabled:opacity-50">بعدی</button>
			</div>
		</div>
	{/if}
</div>

<!-- Create Modal -->
{#if showCreate}
	<div class="fixed inset-0 bg-black/40 z-50 flex items-center justify-center p-4" onclick={() => showCreate = false}>
		<div class="bg-white rounded-2xl w-full max-w-md shadow-xl" onclick={(e) => e.stopPropagation()}>
			<div class="px-6 py-4 border-b">
				<h2 class="font-bold text-lg">ایجاد کلاس جدید</h2>
			</div>
			<div class="px-6 py-4 space-y-4">
				{#if formError}
					<div class="p-3 bg-red-50 text-red-600 rounded-lg text-sm">{formError}</div>
				{/if}

				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">نام کلاس</label>
					<input
						type="text"
						bind:value={formName}
						class="w-full px-4 py-2.5 border border-gray-200 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none text-sm"
						placeholder="مثال: ریاضی پایه دهم"
						required
					/>
				</div>

				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">توضیحات</label>
					<textarea
						bind:value={formDesc}
						class="w-full px-4 py-2.5 border border-gray-200 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none text-sm resize-none"
						rows="2"
						placeholder="توضیحات کلاس..."
					></textarea>
				</div>

				<div>
					<label class="block text-sm font-medium text-gray-700 mb-2">رنگ</label>
					<div class="flex gap-2">
						{#each colors as color}
							<button
								class="w-8 h-8 rounded-full transition-transform {formColor === color ? 'ring-2 ring-offset-2 ring-blue-500 scale-110' : 'hover:scale-105'}"
								style="background-color: {color}"
								onclick={() => formColor = color}
							></button>
						{/each}
					</div>
				</div>

				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">حداکثر دانش‌آموز</label>
					<input
						type="number"
						bind:value={formMaxStudents}
						class="w-full px-4 py-2.5 border border-gray-200 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none text-sm"
						min="1"
						max="200"
					/>
				</div>
			</div>
			<div class="px-6 py-4 border-t flex justify-end gap-3">
				<button onclick={() => showCreate = false} class="px-4 py-2 text-sm text-gray-600 hover:bg-gray-100 rounded-lg transition-colors">انصراف</button>
				<button
					onclick={createClass}
					disabled={formLoading || !formName}
					class="px-4 py-2 bg-blue-600 text-white text-sm rounded-lg font-medium hover:bg-blue-700 transition-colors disabled:opacity-50"
				>
					{formLoading ? 'در حال ایجاد...' : 'ایجاد کلاس'}
				</button>
			</div>
		</div>
	</div>
{/if}
