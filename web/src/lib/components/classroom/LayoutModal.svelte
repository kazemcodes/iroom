<script lang="ts">
	let {
		showUsers,
		showChat,
		onToggleUsers,
		onToggleChat,
		onClose,
	}: {
		showUsers: boolean;
		showChat: boolean;
		onToggleUsers: () => void;
		onToggleChat: () => void;
		onClose: () => void;
	} = $props();
</script>

<div class="modal-overlay" onclick={onClose}>
	<div class="modal-content" onclick={(e) => e.stopPropagation()}>
		<div class="modal-header">
			<span>چیدمان</span>
			<button class="close-btn" onclick={onClose}>
				<svg width="20" height="20"><use xlink:href="#shape_clear"></use></svg>
			</button>
		</div>
		<div class="modal-body">
			<label class="toggle-row" onclick={onToggleUsers}>
				<span>کاربران</span>
				<input type="checkbox" checked={showUsers} class="toggle" readonly />
			</label>
			<label class="toggle-row" onclick={onToggleChat}>
				<span>پیام‌ها</span>
				<input type="checkbox" checked={showChat} class="toggle" readonly />
			</label>
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
		width: 300px;
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
	.toggle-row {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 10px 0;
		border-bottom: 1px solid rgba(255,255,255,0.04);
		font-size: 0.85rem;
		color: #e0e0e6;
		cursor: pointer;
	}
	.toggle-row:last-child { border-bottom: none; }
	.toggle {
		width: 36px;
		height: 20px;
		accent-color: #23b9d7;
		cursor: pointer;
	}
</style>
