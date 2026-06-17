<script lang="ts">
	import type { Participant } from '$lib/classroom/types';

	let {
		participants = [],
		columns = 2,
	}: {
		participants: Participant[];
		columns: number;
	} = $props();

	const gridClass = $derived(columns === 2 ? 'grid-cols-2' : columns === 3 ? 'grid-cols-3' : 'grid-cols-4');
</script>

<div class="grid {gridClass} gap-1 p-1 h-full" style="background-color: #1a1a2e;">
	{#each participants as p (p.id)}
		<div class="relative rounded overflow-hidden" style="background-color: #0a0a1a; aspect-ratio: 16/9;">
			<!-- Video placeholder -->
			<div class="w-full h-full flex items-center justify-center">
				<div class="w-12 h-12 rounded-full flex items-center justify-center text-lg font-bold text-white" style="background-color: #3a3a5a;">
					{p.name.charAt(0)}
				</div>
			</div>

			<!-- Name label -->
			<div class="absolute bottom-0 left-0 right-0 px-2 py-1" style="background: linear-gradient(transparent, rgba(0,0,0,0.7));">
				<p class="text-[11px] text-white truncate">{p.name}</p>
			</div>

			<!-- Muted indicator -->
			{#if !p.hasAudio}
				<div class="absolute top-1 right-1 w-5 h-5 rounded-full flex items-center justify-center" style="background-color: rgba(224, 82, 82, 0.8);">
					<svg class="w-3 h-3 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5.586 15H4a1 1 0 01-1-1v-4a1 1 0 011-1h1.586l4.707-4.707C10.923 3.663 12 4.109 12 5v14c0 .891-1.077 1.337-1.707.707L5.586 15z" /><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2" /></svg>
				</div>
			{/if}

			<!-- Speaking indicator -->
			{#if p.isSpeaking}
				<div class="absolute inset-0 rounded" style="border: 2px solid #23b9d7; box-shadow: 0 0 8px rgba(35, 185, 215, 0.5);"></div>
			{/if}
		</div>
	{/each}
</div>
