<script lang="ts">
	import { toasts, type Toast } from '$lib/stores/toast';

	const borderColors: Record<Toast['type'], string> = {
		success: 'border-r-green-500',
		error: 'border-r-red-500',
		info: 'border-r-blue-500',
		warning: 'border-r-amber-500'
	};

	const bgColors: Record<Toast['type'], string> = {
		success: 'bg-green-50',
		error: 'bg-red-50',
		info: 'bg-blue-50',
		warning: 'bg-amber-50'
	};

	const textColors: Record<Toast['type'], string> = {
		success: 'text-green-800',
		error: 'text-red-800',
		info: 'text-blue-800',
		warning: 'text-amber-800'
	};

	const icons: Record<Toast['type'], string> = {
		success: 'M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z',
		error: 'M12 9v3.75m9-.75a9 9 0 11-18 0 9 9 0 0118 0zm-9 3.75h.008v.008H12v-.008z',
		info: 'M11.25 11.25l.041-.02a.75.75 0 011.063.852l-.708 2.836a.75.75 0 001.063.853l.041-.021M21 12a9 9 0 11-18 0 9 9 0 0118 0zm-9-3.75h.008v.008H12V8.25z',
		warning: 'M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126zM12 15.75h.007v.008H12v-.008z'
	};
</script>

{#if $toasts.length > 0}
	<div class="fixed bottom-6 left-1/2 -translate-x-1/2 z-50 flex flex-col gap-3 pointer-events-none">
		{#each $toasts as toast (toast.id)}
			<div
				class="pointer-events-auto flex items-center gap-3 px-5 py-3 rounded-xl shadow-lg border-r-4 backdrop-blur-sm {bgColors[toast.type]} {borderColors[toast.type]} {textColors[toast.type]} animate-slide-up"
				role="alert"
			>
				<svg class="w-5 h-5 shrink-0" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" d={icons[toast.type]} />
				</svg>
				<span class="text-sm font-medium">{toast.message}</span>
				<button
					onclick={() => toasts.removeToast(toast.id)}
					class="ml-2 shrink-0 opacity-60 hover:opacity-100 transition-opacity"
					aria-label="بستن"
				>
					<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
						<path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
					</svg>
				</button>
			</div>
		{/each}
	</div>
{/if}

<style>
	@keyframes slide-up {
		from {
			opacity: 0;
			transform: translateY(1rem);
		}
		to {
			opacity: 1;
			transform: translateY(0);
		}
	}

	.animate-slide-up {
		animation: slide-up 0.3s ease-out;
	}
</style>
