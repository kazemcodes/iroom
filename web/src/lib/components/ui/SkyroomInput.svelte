<script lang="ts">
	let {
		label = '',
		type = 'text',
		placeholder = '',
		value = $bindable(''),
		error = '',
		helpText = '',
		required = false,
		disabled = false,
		dir = 'auto',
	}: {
		label?: string;
		type?: string;
		placeholder?: string;
		value: string;
		error?: string;
		helpText?: string;
		required?: boolean;
		disabled?: boolean;
		dir?: string;
	} = $props();
</script>

<div>
	{#if label}
		<label class="block text-sm font-medium mb-1" style="color: #1c293a;">
			{label}
			{#if required}<span class="text-[#e05252]">*</span>{/if}
		</label>
	{/if}
	<input
		{type}
		{placeholder}
		{dir}
		{disabled}
		bind:value
		class="w-full px-3 py-2 rounded-lg text-sm outline-none transition-all {error ? 'border-[#e05252]' : 'border-[#e0e4eb] focus:border-[#23b9d7]'} {disabled ? 'bg-[#f9fafb] cursor-not-allowed' : 'bg-white'}"
		style="border: 1px solid; color: #1c293a; {error ? 'border-color: #e05252;' : ''}"
	/>
	{#if error}
		<p class="text-xs mt-1" style="color: #e05252;">{error}</p>
	{:else if helpText}
		<p class="text-xs mt-1" style="color: #6790a0;">{helpText}</p>
	{/if}
</div>
