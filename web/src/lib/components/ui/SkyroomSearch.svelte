<script lang="ts">
	let {
		placeholder = 'جستجو...',
		value = $bindable(''),
	}: {
		placeholder?: string;
		value: string;
	} = $props();
</script>

<div class="relative">
	<svg class="absolute right-3 top-1/2 -translate-y-1/2 w-4 h-4 text-[#94a3b8]" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" /></svg>
	<input type="text" bind:value {placeholder}
		class="w-full pr-10 pl-4 py-2 rounded-lg text-sm outline-none transition-colors focus:border-[#23b9d7]"
		style="border: 1px solid #e0e4eb; color: #1c293a;"
	/>
	{#if value}
		<button onclick={() => value = ''} class="absolute left-3 top-1/2 -translate-y-1/2 text-[#94a3b8] hover:text-[#1c293a]">
			<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" /></svg>
		</button>
	{/if}
</div>
