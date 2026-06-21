import { sveltekit } from '@sveltejs/kit/vite';
import tailwindcss from '@tailwindcss/vite';
import { defineConfig } from 'vitest/config';

export default defineConfig({
	plugins: [
		tailwindcss(),
		sveltekit({
			compilerOptions: {
				runes: ({ filename }) =>
					filename.split(/[/\\]/).includes('node_modules') ? undefined : true
			}
		})
	],
	server: {
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
		include: ['src/**/*.{test,spec}.{js,ts}']
	}
});
