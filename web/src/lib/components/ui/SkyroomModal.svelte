<script lang="ts">
	let {
		open = $bindable(false),
		title = '',
		size = 'md',
		onClose,
	}: {
		open: boolean;
		title: string;
		size?: 'sm' | 'md' | 'lg';
		onClose: () => void;
	} = $props();

	const sizeClasses = { sm: 'max-w-sm', md: 'max-w-md', lg: 'max-w-lg' };
</script>

{#if open}
	<div class="fixed inset-0 z-50 flex items-center justify-center p-4" onclick={onClose}>
		<div class="fixed inset-0" style="background: rgba(0, 0, 0, 0.5);"></div>
		<div class="relative w-full {sizeClasses[size]} rounded-xl overflow-hidden shadow-2xl" style="background: white;" onclick={(e) => e.stopPropagation()}>
			<!-- Header -->
			<div class="flex items-center justify-between px-5 py-4" style="border-bottom: 1px solid #e0e4eb;">
				<h2 class="text-base font-semibold" style="color: #1c293a;">{title}</h2>
				<button onclick={onClose} class="text-[#94a3b8] hover:text-[#1c293a] p-1">
					<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" /></svg>
				</button>
			</div>

			<!-- Content -->
			<div class="px-5 py-4">
				<slot name="content" />
			</div>

			<!-- Footer -->
			<slot name="footer" />
		</div>
	</div>
{/if}
