<script lang="ts">
	let {
		onSelect,
		onClose,
	}: {
		onSelect: (emoji: string) => void;
		onClose: () => void;
	} = $props();

	const emojis = [
		'😀', '😃', '😄', '😁', '😆', '😅', '🤣', '😂', '🙂', '🙃',
		'😉', '😊', '😇', '🥰', '😍', '🤩', '😘', '😗', '😚', '😙',
		'🥲', '😋', '😛', '😜', '🤪', '😝', '🤑', '🤗', '🤭', '🫢',
		'🤫', '🤔', '🫡', '🤐', '🤨', '😐', '😑', '😶', '🫥', '😏',
		'😒', '🙄', '😬', '🤥', '😌', '😔', '😪', '🤤', '😴', '😷',
		'🤒', '🤕', '🤢', '🤮', '🥵', '🥶', '🥴', '😵', '🤯', '🤠',
		'🥳', '🥸', '😎', '🤓', '🧐', '😕', '🫤', '😟', '🙁', '😮',
		'😯', '😲', '😳', '🥺', '🥹', '😦', '😧', '😨', '😰', '😥',
		'😢', '😭', '😱', '😖', '😣', '😞', '😓', '😩', '😫', '🥱',
		'😤', '😡', '😠', '🤬', '😈', '👿', '💀', '☠️', '💩', '🤡',
		'👋', '🤚', '🖐️', '✋', '🖖', '🫱', '🫲', '🫳', '🫴', '👌',
		'🤌', '🤏', '✌️', '🤞', '🫰', '🤟', '🤘', '🤙', '👈', '👉',
		'👍', '👎', '✊', '👊', '🤛', '🤜', '👏', '🙌', '🫶', '👐',
		'❤️', '🧡', '💛', '💚', '💙', '💜', '🖤', '🤍', '🤎', '💔',
		'❤️‍🔥', '❤️‍🩹', '💕', '💞', '💓', '💗', '💖', '💘', '💝', '💟',
		'🔥', '⭐', '🌟', '✨', '💫', '🎉', '🎊', '🎈', '🎁', '🏆',
		'✅', '❌', '⭕', '❗', '❓', '💯', '🔔', '📌', '📎', '🔗',
	];

	function handleSelect(emoji: string) {
		onSelect(emoji);
		onClose();
	}
</script>

<div class="absolute bottom-full left-0 mb-2 w-72 max-h-48 overflow-y-auto rounded-xl shadow-2xl p-2 grid grid-cols-8 gap-1" style="background-color: #1e1e3a; border: 1px solid #2a2a4a;">
	{#each emojis as emoji}
		<button
			onclick={() => handleSelect(emoji)}
			class="w-8 h-8 flex items-center justify-center rounded-lg hover:bg-white/10 transition-colors text-lg"
		>
			{emoji}
		</button>
	{/each}
</div>
