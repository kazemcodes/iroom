<script lang="ts">
	let {
		size = 'md',
	}: {
		size?: 'sm' | 'md' | 'lg';
	} = $props();

	const sizes = { sm: 'w-4 h-4', md: 'w-6 h-6', lg: 'w-10 h-10' };
</script>

<svg class="{sizes[size]} animate-spin" viewBox="0 0 24 24" fill="none">
	<circle cx="12" cy="12" r="10" stroke="currentColor" stroke-width="3" class="opacity-25"></circle>
	<path d="M4 12a8 8 0 018-8" stroke="currentColor" stroke-width="3" stroke-linecap="round"></path>
</svg>
