<script lang="ts">
	import { onMount } from 'svelte';

	let { show = $bindable() }: { show: boolean } = $props();
	let audioDevices = $state<MediaDeviceInfo[]>([]);
	let videoDevices = $state<MediaDeviceInfo[]>([]);
	let selectedAudio = $state('');
	let selectedVideo = $state('');

	onMount(async () => {
		try {
			const devices = await navigator.mediaDevices.enumerateDevices();
			audioDevices = devices.filter(d => d.kind === 'audioinput');
			videoDevices = devices.filter(d => d.kind === 'videoinput');
		} catch {}
	});
</script>

{#if show}
<div class="modal-overlay" onclick={() => show = false} role="button" tabindex="-1">
	<div class="modal-content p-6" onclick={(e) => e.stopPropagation()}>
		<h3 class="font-bold text-lg mb-4">تنظیمات صدا و تصویر</h3>
		<div class="space-y-4">
			<div>
				<label class="block text-sm font-medium mb-1">میکروفون</label>
				<select bind:value={selectedAudio} class="input-field">
					{#each audioDevices as d}<option value={d.deviceId}>{d.label || 'میکروفون ' + d.deviceId}</option>{/each}
				</select>
			</div>
			<div>
				<label class="block text-sm font-medium mb-1">دوربین</label>
				<select bind:value={selectedVideo} class="input-field">
					{#each videoDevices as d}<option value={d.deviceId}>{d.label || 'دوربین ' + d.deviceId}</option>{/each}
				</select>
			</div>
		</div>
		<div class="flex justify-end mt-6">
			<button onclick={() => show = false} class="btn-primary">بستن</button>
		</div>
	</div>
</div>
{/if}
