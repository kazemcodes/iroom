<script lang="ts">
	let {
		variant = 'success',
		message = '',
		duration = 4000,
		onDismiss,
	}: {
		variant?: 'success' | 'error' | 'warning';
		message: string;
		duration?: number;
		onDismiss: () => void;
	} = $props();

	const colors = {
		success: '#40bf7f',
		error: '#e05252',
		warning: '#d7911d',
	};

	const icons = {
		success: '✓',
		error: '✕',
		warning: '⚠',
	};

	let visible = $state(true);

	$effect(() => {
		const timer = setTimeout(() => { visible = false; onDismiss(); }, duration);
		return () => clearTimeout(timer);
	});
</script>

{#if visible}
	<div class="flex items-center gap-3 px-4 py-3 rounded-lg shadow-lg max-w-sm" style="background: white; border-left: 4px solid {colors[variant]}; animation: slideIn 0.3s ease;">
		<span class="text-sm font-bold" style="color: {colors[variant]};">{icons[variant]}</span>
		<p class="text-sm flex-1" style="color: #1c293a;">{message}</p>
		<button onclick={() => { visible = false; onDismiss(); }} class="text-[#94a3b8] hover:text-[#1c293a]">
			<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" /></svg>
		</button>
	</div>
{/if}

<style>
	@keyframes slideIn {
		from { transform: translateX(100%); opacity: 0; }
		to { transform: translateX(0); opacity: 1; }
	}
</style>
