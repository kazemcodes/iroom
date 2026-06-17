<script lang="ts">
	let {
		show = $bindable(false),
		title = '',
		message = '',
		onConfirm,
		onCancel
	}: {
		show: boolean;
		title: string;
		message: string;
		onConfirm: () => void;
		onCancel: () => void;
	} = $props();

	function confirm() {
		show = false;
		onConfirm();
	}

	function cancel() {
		show = false;
		onCancel();
	}
</script>

{#if show}
	<div class="fixed inset-0 z-50 flex items-center justify-center p-4" onclick={cancel}>
		<div class="fixed inset-0 bg-black/40 backdrop-blur-sm"></div>
		<div
			class="modal-content"
			onclick={(e) => e.stopPropagation()}
		>
			<div class="px-6 pt-6 pb-2">
				<h2 class="font-bold text-lg" style="color: var(--sky-text-primary);">{title}</h2>
			</div>
			<div class="px-6 pb-6">
				<p class="text-sm mt-1" style="color: var(--sky-text-secondary);">{message}</p>
			</div>
			<div class="px-6 py-4 flex justify-end gap-3" style="border-top: 1px solid var(--sky-border);">
				<button onclick={cancel} class="btn-ghost">انصراف</button>
				<button onclick={confirm} class="btn-danger">تایید</button>
			</div>
		</div>
	</div>
{/if}

<style>
	@keyframes slide-up {
		from { transform: translateY(20px); opacity: 0; }
		to { transform: translateY(0); opacity: 1; }
	}
	.animate-slide-up {
		animation: slide-up 0.2s ease-out;
	}
</style>
