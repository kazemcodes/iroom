<script lang="ts">
	let {
		variant = 'default',
		label = '',
		dot = false,
	}: {
		variant?: 'success' | 'warning' | 'danger' | 'info' | 'default';
		label: string;
		dot?: boolean;
	} = $props();

	const colors = {
		success: { bg: '#e6f9f0', text: '#40bf7f' },
		warning: { bg: '#fff4e5', text: '#d7911d' },
		danger: { bg: '#fde8e8', text: '#e05252' },
		info: { bg: '#e6f4ff', text: '#23b9d7' },
		default: { bg: '#f0f2f5', text: '#6790a0' },
	};

	const c = $derived(colors[variant]);
</script>

<span class="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-medium" style="background: {c.bg}; color: {c.text};">
	{#if dot}
		<span class="w-1.5 h-1.5 rounded-full" style="background: {c.text};"></span>
	{/if}
	{label}
</span>
