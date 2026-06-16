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
	<!-- svelte-ignore a11y_click_events_have_key_events -->
	<!-- svelte-ignore a11y_no_static_element_interactions -->
	<div class="fixed inset-0 z-50 flex items-center justify-center p-4" onclick={cancel}>
		<div class="fixed inset-0 bg-black/40 backdrop-blur-sm"></div>
		<div
			class="relative bg-white rounded-2xl w-full max-w-sm shadow-xl animate-slide-up"
			onclick={(e) => e.stopPropagation()}
		>
			<div class="px-6 pt-6 pb-2">
				<h2 class="font-bold text-lg text-gray-900">{title}</h2>
			</div>
			<div class="px-6 pb-6">
				<p class="text-sm text-gray-500 mt-1">{message}</p>
			</div>
			<div class="px-6 py-4 border-t flex justify-end gap-3">
				<button onclick={cancel} class="px-4 py-2 text-sm text-gray-600 hover:bg-gray-100 rounded-lg transition-colors">انصراف</button>
				<button onclick={confirm} class="px-4 py-2 bg-red-600 text-white text-sm rounded-lg font-medium hover:bg-red-700 transition-colors">تایید</button>
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
