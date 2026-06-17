<script lang="ts">
	let { onClose }: { onClose: () => void } = $props();

	let stats = $state({
		duration: '00:00',
		streamsDown: 0,
		streamsUp: 0,
		trafficDown: '0 B',
		trafficUp: '0 B',
		speedDown: '0 bps',
		speedUp: '0 bps',
		quality: '-',
		protocol: 'UDP',
		latency: 0,
		jitter: 0,
		packetLoss: '0%',
	});

	let startTime = Date.now();
	let timerInterval: ReturnType<typeof setInterval>;

	function updateStats() {
		const elapsed = Math.floor((Date.now() - startTime) / 1000);
		const m = Math.floor(elapsed / 60);
		const s = elapsed % 60;
		stats.duration = `${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`;
	}

	import { onMount, onDestroy } from 'svelte';
	onMount(() => { timerInterval = setInterval(updateStats, 1000); });
	onDestroy(() => { clearInterval(timerInterval); });
</script>

<div class="modal-overlay" onclick={onClose}>
	<div class="modal-content" onclick={(e) => e.stopPropagation()}>
		<div class="modal-header">
			<span>وضعیت اتصال</span>
			<button class="close-btn" onclick={onClose}>
				<svg width="20" height="20"><use xlink:href="#shape_clear"></use></svg>
			</button>
		</div>
		<div class="modal-body">
			<div class="info-row"><span class="label">مدت زمان اتصال</span><span class="value ltr">{stats.duration}</span></div>
			<div class="info-row"><span class="label">تعداد استریم ↓↑</span><span class="value ltr">{stats.streamsDown} | {stats.streamsUp}</span></div>
			<div class="info-row"><span class="label">ترافیک مصرفی ↓↑</span><span class="value ltr">{stats.trafficDown} | {stats.trafficUp}</span></div>
			<div class="info-row"><span class="label">سرعت تبادل داده ↓↑</span><span class="value ltr">{stats.speedDown} | {stats.speedUp}</span></div>
			<div class="info-row"><span class="label">کیفیت اتصال از ۱۰ ↓↑</span><span class="value ltr">{stats.quality} | {stats.quality}</span></div>
			<div class="info-row"><span class="label">پروتکل ↓↑</span><span class="value ltr">{stats.protocol} | {stats.protocol}</span></div>
			<div class="info-row"><span class="label">تأخیر ↓↑</span><span class="value ltr">{stats.latency} ms | {stats.latency} ms</span></div>
			<div class="info-row"><span class="label">جیتر ↓↑</span><span class="value ltr">{stats.jitter} ms | {stats.jitter} ms</span></div>
			<div class="info-row"><span class="label">پکت لاس ↓↑</span><span class="value ltr">{stats.packetLoss} | {stats.packetLoss}</span></div>
		</div>
	</div>
</div>

<style>
	.modal-overlay {
		position: fixed;
		inset: 0;
		background: rgba(0,0,0,0.5);
		display: flex;
		align-items: center;
		justify-content: center;
		z-index: 200;
		animation: fadeIn 0.15s ease;
	}
	@keyframes fadeIn { from { opacity: 0; } to { opacity: 1; } }
	.modal-content {
		background: #1c2a3a;
		border-radius: 12px;
		width: 380px;
		max-width: 90vw;
		box-shadow: 0 12px 40px rgba(0,0,0,0.5);
		animation: slideUp 0.2s ease;
	}
	@keyframes slideUp { from { transform: translateY(12px); opacity: 0; } to { transform: translateY(0); opacity: 1; } }
	.modal-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 14px 16px;
		border-bottom: 1px solid rgba(255,255,255,0.08);
		font-weight: 600;
		font-size: 0.9rem;
		color: #e0e0e6;
	}
	.close-btn {
		background: none;
		border: none;
		cursor: pointer;
		padding: 4px;
		border-radius: 6px;
		display: flex;
		align-items: center;
		justify-content: center;
	}
	.close-btn svg { fill: #8a8a96; }
	.close-btn:hover svg { fill: #e0e0e6; }
	.modal-body { padding: 8px 16px 16px; }
	.info-row {
		display: flex;
		justify-content: space-between;
		padding: 6px 0;
		border-bottom: 1px solid rgba(255,255,255,0.04);
		font-size: 0.82rem;
	}
	.info-row:last-child { border-bottom: none; }
	.label { color: #8a8a96; }
	.value { color: #e0e0e6; font-weight: 500; }
	.ltr { direction: ltr; font-family: monospace; font-size: 0.8rem; }
</style>
