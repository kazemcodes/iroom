<script lang="ts">
	import { auth } from '$lib/stores';
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';

	let { children } = $props();

	onMount(() => {
		auth.init();
		const unsub = auth.subscribe(($auth) => {
			if (!$auth.isLoggedIn) {
				goto('/auth');
			}
		});
		return unsub;
	});
</script>

{#if $auth.isLoggedIn}
	{@render children()}
{/if}
