<script lang="ts">
	let {
		icon = '📭',
		title = 'هنوز موردی ثبت نشده است',
		description = '',
		actionLabel = '',
		onAction,
	}: {
		icon?: string;
		title?: string;
		description?: string;
		actionLabel?: string;
		onAction?: () => void;
	} = $props();
</script>

<div class="flex flex-col items-center justify-center py-16">
	<span class="text-4xl mb-4">{icon}</span>
	<p class="text-base font-medium mb-1" style="color: #1c293a;">{title}</p>
	{#if description}
		<p class="text-[13px] mb-4" style="color: #6790a0;">{description}</p>
	{/if}
	{#if actionLabel && onAction}
		<button onclick={onAction} class="px-4 py-2 rounded-lg text-sm font-semibold text-white" style="background: #23b9d7;">
			{actionLabel}
		</button>
	{/if}
</div>
