<!--
  Root Layout — App entry point.
  
  Responsibilities:
    - Imports global CSS (TailwindCSS)
    - Initializes auth store from localStorage
    - Renders child routes
-->
<script lang="ts">
	import '../app.css';
	import { auth } from '$lib/stores';
	import { onMount } from 'svelte';
	import { dev } from '$app/environment';

	let { children } = $props();

	onMount(() => {
		auth.init();

		// Service worker: register only in production. In development we actively
		// unregister any existing worker and clear caches, because the cache-first
		// strategy in static/sw.js otherwise serves stale UI and hides live edits.
		if ('serviceWorker' in navigator) {
			if (dev) {
				navigator.serviceWorker.getRegistrations().then((regs) => {
					regs.forEach((reg) => reg.unregister());
				});
				if (window.caches) {
					caches.keys().then((keys) => keys.forEach((k) => caches.delete(k)));
				}
			} else {
				window.addEventListener('load', () => {
					navigator.serviceWorker.register('/sw.js', { scope: '/' }).catch(() => {});
				});
			}
		}
	});
</script>

{@render children()}
