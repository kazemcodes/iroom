<script lang="ts">
	import { auth } from '$lib/stores';
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';

	onMount(() => {
		auth.init();
		const unsub = auth.subscribe(($auth) => {
			if ($auth.isLoggedIn) {
				goto('/dashboard');
			}
		});
		return unsub;
	});
</script>

<div class="min-h-screen flex items-center justify-center bg-gradient-to-br from-blue-50 to-indigo-100">
	<div class="text-center">
		<div class="w-16 h-16 bg-blue-600 rounded-2xl flex items-center justify-center text-white font-bold text-2xl mx-auto mb-4">آ</div>
		<h1 class="text-2xl font-bold text-gray-900">آی‌روم</h1>
		<p class="text-gray-500 mt-1">در حال بارگذاری...</p>
	</div>
</div>
