<script lang="ts">
	/**
	 * Classroom Slug Route — /user-name/class-name/
	 *
	 * This is the Skyroom-style URL for joining a class.
	 * Resolves the slug to a session ID and redirects to the classroom popup.
	 */
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import { api } from '$lib/api';

	const teacherSlug = $derived(page.params.teacherSlug);
	const classSlug = $derived(page.params.classSlug);

	let error = $state('');
	let loading = $state(true);

	onMount(async () => {
		try {
			// Resolve the slug to get the class ID
			const res = await api.get<any>(`/classes/slug/${classSlug}`);
			if (res.success && res.data) {
				// Redirect to the classroom popup
				goto(`/classroom/popup/${res.data.id}`, { replaceState: true });
			} else {
				error = 'کلاس یافت نشد';
				loading = false;
			}
		} catch (e) {
			error = 'خطا در بارگذاری کلاس';
			loading = false;
		}
	});
</script>

<div class="min-h-screen flex items-center justify-center" style="background: linear-gradient(135deg, #0b1120 0%, #1a1a2e 50%, #0d1b2a 100%);">
	{#if loading}
		<div class="text-center">
			<div class="animate-spin h-8 w-8 border-4 border-[#23b9d7] border-t-transparent rounded-full mx-auto mb-4"></div>
			<p style="color: #94a3b8; font-size: 14px;">در حال بارگذاری...</p>
		</div>
	{:else if error}
		<div class="text-center">
			<div class="w-16 h-16 rounded-full bg-[#252540] flex items-center justify-center mx-auto mb-4">
				<span class="text-2xl">❌</span>
			</div>
			<p style="color: #e05252; font-size: 14px;">{error}</p>
			<a href="/" style="color: #23b9d7; font-size: 13px; text-decoration: none;">بازگشت به صفحه اصلی</a>
		</div>
	{/if}
</div>
