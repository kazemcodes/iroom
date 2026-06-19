<!--
  AttendanceModal — Shows attendance list for the classroom.
  
  Displays each participant with:
    - Name and role
    - Audio/video status (SVG icons)
    - Hand raised status
    - Connected status
  
  Props:
    participants: Current connected participants
    onClose: Callback to close the modal
-->
<script lang="ts">
	import type { Participant } from '$lib/classroom/types';
	import { ROLE_LABELS } from '$lib/classroom/types';

	let { participants = [], onClose }: { participants: Participant[]; onClose: () => void } = $props();

	const presentCount = $derived(participants.length);
</script>

<div class="modal-overlay" onclick={onClose}>
	<div class="modal-content" onclick={(e) => e.stopPropagation()}>
		<div class="modal-header">
			<div class="header-left">
				<svg width="18" height="18" style="fill:#8a8a96;"><use xlink:href="#shape_check"></use></svg>
				<span>حضور و غیاب</span>
			</div>
			<button class="close-btn" onclick={onClose}>
				<svg width="18" height="18"><use xlink:href="#shape_clear"></use></svg>
			</button>
		</div>

		<div class="modal-body">
			<div class="summary-row">
				<span class="summary-label">تعداد حاضرین</span>
				<span class="summary-value">{presentCount}</span>
			</div>

			{#if participants.length === 0}
				<div class="empty-state">
					<svg width="32" height="32" style="fill:#5a6070;"><use xlink:href="#shape_person"></use></svg>
					<p>هنوز کسی متصل نیست</p>
				</div>
			{:else}
				<div class="attendance-list">
					{#each participants as p}
						<div class="attendance-row">
							<div class="att-avatar" class:online={true}>
								<svg width="14" height="14"><use xlink:href="#shape_person"></use></svg>
							</div>
							<div class="att-info">
								<span class="att-name">{p.name}</span>
								<span class="att-role">{ROLE_LABELS[p.role] || p.role}</span>
							</div>
							<div class="att-media">
								<span class="att-media-icon" class:off={!p.hasAudio}>
									<svg width="12" height="12"><use xlink:href={p.hasAudio ? '#shape_mic' : '#shape_mic_off'}></use></svg>
								</span>
								<span class="att-media-icon" class:off={!p.hasVideo}>
									<svg width="12" height="12"><use xlink:href={p.hasVideo ? '#shape_videocam' : '#shape_videocamoff'}></use></svg>
								</span>
								{#if p.handRaised}
									<span class="att-media-icon hand">
										<svg width="12" height="12" style="fill:#f59e0b;"><use xlink:href="#shape_hand"></use></svg>
									</span>
								{/if}
							</div>
							<div class="att-status online">
								<div class="status-dot"></div>
								<span>متصل</span>
							</div>
						</div>
					{/each}
				</div>
			{/if}
		</div>

		<div class="modal-footer">
			<button class="btn-confirm" onclick={onClose}>بستن</button>
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
		width: 420px;
		max-width: 90vw;
		max-height: 80vh;
		display: flex;
		flex-direction: column;
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
	.header-left { display: flex; align-items: center; gap: 8px; }
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
	.modal-body { padding: 0; overflow-y: auto; flex: 1; }
	.summary-row {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 12px 16px;
		border-bottom: 1px solid rgba(255,255,255,0.06);
	}
	.summary-label { color: #8a8a96; font-size: 0.82rem; }
	.summary-value { color: #23b9d7; font-weight: 700; font-size: 0.9rem; }
	.empty-state {
		padding: 40px 16px;
		text-align: center;
		color: #5a6070;
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 8px;
	}
	.empty-state p { font-size: 0.82rem; }
	.attendance-list { padding: 4px 0; }
	.attendance-row {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 8px 16px;
		transition: background 0.12s;
	}
	.attendance-row:hover { background: rgba(255,255,255,0.03); }
	.att-avatar {
		width: 30px;
		height: 30px;
		border-radius: 50%;
		display: flex;
		align-items: center;
		justify-content: center;
		background: rgba(255,255,255,0.08);
		color: #8a8a96;
		flex-shrink: 0;
	}
	.att-avatar.online { background: rgba(64,191,127,0.15); color: #40bf7f; }
	.att-avatar svg { fill: currentColor; }
	.att-info { flex: 1; min-width: 0; display: flex; flex-direction: column; }
	.att-name { font-size: 0.82rem; font-weight: 600; color: #e0e0e6; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
	.att-role { font-size: 0.7rem; color: #8a8a96; }
	.att-media { display: flex; align-items: center; gap: 4px; flex-shrink: 0; }
	.att-media-icon { display: flex; align-items: center; }
	.att-media-icon svg { fill: #40bf7f; }
	.att-media-icon.off svg { fill: #5a6070; opacity: 0.5; }
	.att-media-icon.hand svg { fill: #f59e0b; }
	.att-status {
		display: flex;
		align-items: center;
		gap: 5px;
		font-size: 0.7rem;
		color: #8a8a96;
		flex-shrink: 0;
	}
	.att-status.online { color: #40bf7f; }
	.status-dot { width: 6px; height: 6px; border-radius: 50%; background: #5a6070; }
	.att-status.online .status-dot { background: #40bf7f; box-shadow: 0 0 6px rgba(64,191,127,0.4); }
	.modal-footer {
		display: flex;
		justify-content: flex-start;
		padding: 12px 16px;
		border-top: 1px solid rgba(255,255,255,0.08);
	}
	.btn-confirm {
		background: #23b9d7;
		color: #fff;
		border: none;
		padding: 8px 24px;
		border-radius: 6px;
		font-size: 0.85rem;
		font-weight: 600;
		cursor: pointer;
		font-family: inherit;
		transition: background 0.15s;
	}
	.btn-confirm:hover { background: #1a9fc0; }
</style>
