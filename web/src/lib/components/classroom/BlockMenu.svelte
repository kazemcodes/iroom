<script lang="ts">
	let {
		items = [],
		onClose,
	}: {
		items: { icon: string; label: string; action: () => void; separator?: boolean }[];
		onClose: () => void;
	} = $props();

	function handleClick(item: { action: () => void }) {
		item.action();
		onClose();
	}
</script>

<div class="py-1" style="min-width: 180px;">
	{#each items as item}
		{#if item.separator}
			<div class="h-px my-1" style="background-color: #e0e4eb;"></div>
		{:else}
			<button onclick={() => handleClick(item)} class="w-full px-4 py-2 flex items-center gap-3 text-[13px] text-[#1c293a] hover:bg-[#f0f2f5] transition-colors">
				<span class="text-sm">{item.icon}</span>
				{item.label}
			</button>
		{/if}
	{/each}
</div>
