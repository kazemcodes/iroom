<script lang="ts">
	let {
		variant = 'primary',
		size = 'md',
		loading = false,
		disabled = false,
		icon = '',
		type = 'button',
		onclick,
		children,
	}: {
		variant?: 'primary' | 'secondary' | 'danger' | 'ghost' | 'outline';
		size?: 'sm' | 'md' | 'lg';
		loading?: boolean;
		disabled?: boolean;
		icon?: string;
		type?: 'button' | 'submit';
		onclick?: () => void;
		children: any;
	} = $props();

	const sizeClasses = {
		sm: 'px-3 py-1.5 text-xs',
		md: 'px-4 py-2 text-sm',
		lg: 'px-5 py-2.5 text-base',
	};

	const variantClasses = {
		primary: 'bg-[#23b9d7] text-white hover:bg-[#1a9ad4]',
		secondary: 'bg-[#f0f2f5] text-[#1c293a] hover:bg-[#e0e4eb]',
		danger: 'bg-[#e05252] text-white hover:bg-[#c44040]',
		ghost: 'bg-transparent text-[#23b9d7] hover:bg-[#e6f9f7]',
		outline: 'bg-white text-[#1c293a] border border-[#e0e4eb] hover:bg-[#f0f2f5]',
	};
</script>

<button
	{type}
	{onclick}
	disabled={disabled || loading}
	class="inline-flex items-center justify-center gap-2 rounded-lg font-semibold transition-all {sizeClasses[size]} {variantClasses[variant]} {disabled || loading ? 'opacity-50 cursor-not-allowed' : ''}"
>
	{#if loading}
		<svg class="w-4 h-4 animate-spin" viewBox="0 0 24 24" fill="none"><circle cx="12" cy="12" r="10" stroke="currentColor" stroke-width="3" class="opacity-25"></circle><path d="M4 12a8 8 0 018-8" stroke="currentColor" stroke-width="3" stroke-linecap="round"></path></svg>
	{:else if icon}
		<span>{icon}</span>
	{/if}
	{@render children()}
</button>
