<script lang="ts">
	let {
		icon = '',
		label = '',
		value = '',
		actionLabel = '',
		onAction,
	}: {
		icon?: string;
		label: string;
		value: string;
		actionLabel?: string;
		onAction?: () => void;
	} = $props();
</script>

<div class="flex items-center gap-3 py-2.5" style="border-bottom: 1px solid #e0e4eb;">
	{#if icon}
		<span class="text-base">{icon}</span>
	{/if}
	<span class="text-[13px] flex-1" style="color: #6790a0;">{label}</span>
	<span class="text-[13px] font-medium" style="color: #1c293a;">{value}</span>
	{#if actionLabel && onAction}
		<button onclick={onAction} class="text-xs font-medium hover:underline" style="color: #23b9d7;">{actionLabel}</button>
	{/if}
</div>
