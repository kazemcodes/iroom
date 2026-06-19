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
			<svg class="sky-spinner lg mx-auto mb-4" viewBox="0 0 24 24" fill="none" stroke="#23b9d7" stroke-width="2" stroke-linecap="round"><path d="M12 2v4M12 18v4M4.93 4.93l2.83 2.83M16.24 16.24l2.83 2.83M2 12h4M18 12h4M4.93 19.07l2.83-2.83M16.24 7.76l2.83-2.83"/></svg>
			<p style="color: #94a3b8; font-size: 14px;">در حال بارگذاری...</p>
		</div>
	{:else if error}
		<div class="text-center">
			<div class="w-16 h-16 rounded-full bg-[#252540] flex items-center justify-center mx-auto mb-4">
				<svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="#e05252" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><line x1="15" y1="9" x2="9" y2="15"/><line x1="9" y1="9" x2="15" y2="15"/></svg>
			</div>
			<p style="color: #e05252; font-size: 14px;">{error}</p>
			<a href="/" style="color: #23b9d7; font-size: 13px; text-decoration: none;">بازگشت به صفحه اصلی</a>
		</div>
	{/if}
</div>
