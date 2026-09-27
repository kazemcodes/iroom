import adapter from '@sveltejs/adapter-static';

/** @type {import('@sveltejs/kit').Config} */
const config = {
	kit: {
		adapter: adapter({
			pages: 'build',
			assets: 'build',
			fallback: 'index.html',
			precompress: false,
			strict: false
		})
	},
	// Enable runes for all first-party code but leave node_modules alone.
	// This used to live in vite.config.ts, but passing inline options to
	// sveltekit() makes SvelteKit ignore this whole file — adapter included.
	compilerOptions: {
		runes: ({ filename }) =>
			filename.split(/[/\\]/).includes('node_modules') ? undefined : true
	}
};

export default config;
