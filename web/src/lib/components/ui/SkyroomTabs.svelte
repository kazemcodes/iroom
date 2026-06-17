<script lang="ts">
	let {
		tabs = [],
		activeTab = '',
		onSelect,
	}: {
		tabs: { id: string; label: string }[];
		activeTab: string;
		onSelect: (id: string) => void;
	} = $props();
</script>

<div class="flex gap-6" style="border-bottom: 1px solid #e0e4eb;">
	{#each tabs as tab}
		<button
			onclick={() => onSelect(tab.id)}
			class="py-2.5 text-sm font-medium transition-colors relative {activeTab === tab.id ? 'text-[#23b9d7]' : 'text-[#94a3b8] hover:text-[#1c293a]'}"
		>
			{tab.label}
			{#if activeTab === tab.id}
				<div class="absolute bottom-0 left-0 right-0 h-0.5 rounded-full" style="background: #23b9d7;"></div>
			{/if}
		</button>
	{/each}
</div>
