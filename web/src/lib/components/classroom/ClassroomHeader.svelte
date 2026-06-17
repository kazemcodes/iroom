<script lang="ts">
	let {
		ownerName = 'مالک',
		roomName = 'اتاق',
		timerSeconds = 0,
	}: {
		ownerName: string;
		roomName: string;
		timerSeconds: number;
	} = $props();

	const formattedTime = $derived.by(() => {
		const m = Math.floor(timerSeconds / 60);
		const s = timerSeconds % 60;
		return `${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`;
	});
</script>

<header class="flex items-center justify-between px-4 h-12 shrink-0" style="background-color: #252540; border-bottom: 1px solid #3a3a5a;">
	<div class="flex items-center gap-2 min-w-0">
		<div class="w-7 h-7 rounded-full bg-[#23b9d7] flex items-center justify-center text-white text-xs font-bold shrink-0">
			{ownerName.charAt(0)}
		</div>
		<div class="flex items-center gap-1 min-w-0">
			<span class="text-sm font-medium text-[#e2e8f0] truncate">{ownerName}</span>
			<span class="text-[#94a3b8] text-sm">:</span>
			<span class="text-sm font-medium text-[#e2e8f0] truncate">{roomName}</span>
		</div>
	</div>
	<div class="flex items-center gap-2 shrink-0">
		<svg class="w-4 h-4 text-[#94a3b8]" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
		<span class="text-[13px] font-mono text-[#94a3b8]">{formattedTime}</span>
	</div>
</header>
