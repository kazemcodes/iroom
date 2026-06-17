<script lang="ts">
	let {
		title = '',
		subtitle = '',
	}: {
		title: string;
		subtitle?: string;
	} = $props();
</script>

<div class="flex items-center justify-between mb-5">
	<div>
		<h1 class="text-lg font-semibold" style="color: #1c293a;">{title}</h1>
		{#if subtitle}
			<p class="text-[13px] mt-0.5" style="color: #6790a0;">{subtitle}</p>
		{/if}
	</div>
	<div class="flex items-center gap-2">
		<slot name="actions" />
	</div>
</div>
