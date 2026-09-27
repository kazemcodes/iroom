import { sveltekit } from '@sveltejs/kit/vite';
import tailwindcss from '@tailwindcss/vite';
import { defineConfig } from 'vitest/config';

// NOTE: do not pass inline options to sveltekit() here. When the plugin
// receives an options object, SvelteKit ignores svelte.config.js entirely —
// including the adapter — and the build silently produces no index.html.
export default defineConfig({
	plugins: [tailwindcss(), sveltekit()],
	server: {
		host: '0.0.0.0',
		allowedHosts: true,
		proxy: {
			'/api': 'http://localhost:8080',
			'/ws': {
				target: 'ws://localhost:8080',
				ws: true
			},
			'/uploads': 'http://localhost:8080',
			'/recordings': 'http://localhost:8080'
		}
	},
	test: {
		environment: 'jsdom',
		include: ['src/**/*.{test,spec}.{js,ts}'],
		exclude: ['src/**/*.browser.test.{js,ts}']
	}
});
