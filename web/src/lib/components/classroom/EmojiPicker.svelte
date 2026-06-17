<script lang="ts">
	let {
		onSelect,
		onClose,
	}: {
		onSelect: (emoji: string) => void;
		onClose: () => void;
	} = $props();

	const emojis = [
		'😃', '😄', '😊', '😁', '😂', '😅', '😉', '😜', '😍',
		'😏', '😒', '😞', '😩', '😢', '😭', '😤', '😡', '😲',
		'😨', '😱', '🙏', '👍', '👎', '👏', '👋', '👌', '✌️',
		'❤️', '🌹',
	];

	function handleSelect(emoji: string) {
		onSelect(emoji);
		onClose();
	}
</script>

<div class="absolute bottom-full left-0 mb-2 p-2 rounded-lg grid grid-cols-9 gap-0.5" style="background: #ffffff; border: 1px solid #e0e4eb; border-radius: 8px; box-shadow: 0 4px 20px rgba(0,0,0,0.3);">
	{#each emojis as emoji}
		<button
			onclick={() => handleSelect(emoji)}
			class="w-7 h-7 flex items-center justify-center rounded text-base hover:bg-[#f0f2f5] transition-colors"
		>
			{emoji}
		</button>
	{/each}
</div>
