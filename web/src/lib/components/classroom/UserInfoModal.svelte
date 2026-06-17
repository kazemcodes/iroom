<script lang="ts">
	let { onClose }: { onClose: () => void } = $props();

	const info = $derived({
		id: Math.floor(Math.random() * 90000) + 10000,
		username: localStorage.getItem('user') ? JSON.parse(localStorage.getItem('user')!).email : '-',
		displayName: localStorage.getItem('user') ? JSON.parse(localStorage.getItem('user')!).display_name : '-',
		role: localStorage.getItem('user') ? JSON.parse(localStorage.getItem('user')!).role : '-',
		os: navigator.platform,
		browser: navigator.userAgent.split(' ').pop() || '-',
		version: '1.0.0',
		entryTime: new Date().toLocaleTimeString('fa-IR'),
	});
</script>

<div class="modal-overlay" onclick={onClose}>
	<div class="modal-content" onclick={(e) => e.stopPropagation()}>
		<div class="modal-header">
			<span>اطلاعات کاربری</span>
			<button class="close-btn" onclick={onClose}>
				<svg width="20" height="20"><use xlink:href="#shape_clear"></use></svg>
			</button>
		</div>
		<div class="modal-body">
			<div class="info-row"><span class="label">شناسه</span><span class="value">{info.id}</span></div>
			<div class="info-row"><span class="label">نام کاربری</span><span class="value" dir="ltr">{info.username}</span></div>
			<div class="info-row"><span class="label">نام نمایشی</span><span class="value">{info.displayName}</span></div>
			<div class="info-row"><span class="label">نقش</span><span class="value">{info.role}</span></div>
			<div class="info-row"><span class="label">سیستم عامل</span><span class="value">{info.os}</span></div>
			<div class="info-row"><span class="label">مرورگر</span><span class="value" dir="ltr">{info.browser}</span></div>
			<div class="info-row"><span class="label">نگارش برنامه</span><span class="value">{info.version}</span></div>
			<div class="info-row"><span class="label">زمان ورود</span><span class="value">{info.entryTime}</span></div>
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
		width: 360px;
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
	.modal-body { padding: 12px 16px 16px; }
	.info-row {
		display: flex;
		justify-content: space-between;
		padding: 8px 0;
		border-bottom: 1px solid rgba(255,255,255,0.04);
		font-size: 0.85rem;
	}
	.info-row:last-child { border-bottom: none; }
	.label { color: #8a8a96; }
	.value { color: #e0e0e6; font-weight: 500; }
</style>
