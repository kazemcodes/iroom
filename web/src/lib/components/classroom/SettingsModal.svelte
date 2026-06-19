<!--
  SettingsModal — Classroom settings with 3 tabs.
  
  Tabs:
    - عمومی (General): Webcam display, mirror, desktop, echo prevention
    - صدا و تصویر (Audio/Video): Microphone/webcam device selection, quality
    - اعلان‌ها (Notifications): Toggle notifications for various events
  
  Triggered from the hamburger menu → تنظیمات.
  
  Props:
    onClose: Callback to close the modal
-->
<script lang="ts">
	import { onMount } from 'svelte';

	let { onClose }: { onClose: () => void } = $props();

	let activeTab = $state<'general' | 'audio' | 'notifications'>('general');

	let generalSettings = $state({
		showSelfWebcam: true,
		mirrorWebcam: false,
		showSelfDesktop: false,
		preventEcho: true,
		sendSystemAudio: false,
	});

	let audioSettings = $state({
		micDevice: '',
		micQuality: 'خوب',
		webcamDevice: '',
		webcamQuality: 'خوب',
	});

	let notifSettings = $state({
		userJoin: true,
		userLeave: true,
		handRaise: true,
		kicked: true,
		newMessage: true,
		muteToggle: true,
		serverDisconnect: true,
	});

	let micDevices = $state<string[]>([]);
	let webcamDevices = $state<string[]>([]);

	onMount(async () => {
		try {
			const devices = await navigator.mediaDevices.enumerateDevices();
			micDevices = devices.filter(d => d.kind === 'audioinput').map(d => d.label || `میکروفون ${micDevices.length + 1}`);
			webcamDevices = devices.filter(d => d.kind === 'videoinput').map(d => d.label || `وبکم ${webcamDevices.length + 1}`);
			if (micDevices.length > 0) audioSettings.micDevice = micDevices[0];
			if (webcamDevices.length > 0) audioSettings.webcamDevice = webcamDevices[0];
		} catch (e) {
			micDevices = ['پیش‌فرض'];
			webcamDevices = ['پیش‌فرض'];
		}
	});
</script>

<div class="modal-overlay" onclick={onClose}>
	<div class="modal-content modal-lg" onclick={(e) => e.stopPropagation()}>
		<div class="modal-header">
			<span>تنظیمات</span>
			<button class="close-btn" onclick={onClose}>
				<svg width="20" height="20"><use xlink:href="#shape_clear"></use></svg>
			</button>
		</div>

		<!-- Tabs -->
		<div class="tabs">
			<button class:active={activeTab === 'general'} onclick={() => activeTab = 'general'}>عمومی</button>
			<button class:active={activeTab === 'audio'} onclick={() => activeTab = 'audio'}>صدا و تصویر</button>
			<button class:active={activeTab === 'notifications'} onclick={() => activeTab = 'notifications'}>اعلان‌ها</button>
		</div>

		<div class="modal-body">
			{#if activeTab === 'general'}
				<div class="settings-group">
					<label class="toggle-row">
						<span>نمایش تصویر وبکم خودم</span>
						<input type="checkbox" bind:checked={generalSettings.showSelfWebcam} class="toggle" />
					</label>
					<label class="toggle-row">
						<span>نمایش معکوس تصویر وبکم خودم (حالت آینه)</span>
						<input type="checkbox" bind:checked={generalSettings.mirrorWebcam} class="toggle" />
					</label>
					<label class="toggle-row">
						<span>نمایش تصویر دسکتاپ خودم</span>
						<input type="checkbox" bind:checked={generalSettings.showSelfDesktop} class="toggle" />
					</label>
					<label class="toggle-row">
						<span>جلوگیری از اکوی تصویر دسکتاپ</span>
						<input type="checkbox" bind:checked={generalSettings.preventEcho} class="toggle" />
					</label>
					<label class="toggle-row">
						<span>ارسال صدای سیستم به هنگام اشتراک دسکتاپ</span>
						<input type="checkbox" bind:checked={generalSettings.sendSystemAudio} class="toggle" />
					</label>
				</div>
			{:else if activeTab === 'audio'}
				<div class="settings-group">
					<div class="setting-item">
						<label>میکروفون:</label>
						<select class="setting-select" bind:value={audioSettings.micDevice}>
							{#each micDevices as device}<option value={device}>{device}</option>{/each}
						</select>
					</div>
					<div class="setting-item">
						<label>کیفیت صدا:</label>
						<select class="setting-select" bind:value={audioSettings.micQuality}>
							<option>خوب</option>
							<option>بهترین</option>
							<option>صرفه‌جویی</option>
						</select>
					</div>
					<div class="setting-item">
						<label>وبکم:</label>
						<select class="setting-select" bind:value={audioSettings.webcamDevice}>
							{#each webcamDevices as device}<option value={device}>{device}</option>{/each}
						</select>
					</div>
					<div class="setting-item">
						<label>کیفیت تصویر:</label>
						<select class="setting-select" bind:value={audioSettings.webcamQuality}>
							<option>خوب</option>
							<option>بهترین</option>
							<option>صرفه‌جویی</option>
						</select>
					</div>
				</div>
			{:else}
				<div class="settings-group">
					<label class="toggle-row">
						<span>ورود کاربر به اتاق</span>
						<input type="checkbox" bind:checked={notifSettings.userJoin} class="toggle" />
					</label>
					<label class="toggle-row">
						<span>خروج کاربر از اتاق</span>
						<input type="checkbox" bind:checked={notifSettings.userLeave} class="toggle" />
					</label>
					<label class="toggle-row">
						<span>بالا بردن دست توسط کاربر</span>
						<input type="checkbox" bind:checked={notifSettings.handRaise} class="toggle" />
					</label>
					<label class="toggle-row">
						<span>اخراج شدن از اتاق</span>
						<input type="checkbox" bind:checked={notifSettings.kicked} class="toggle" />
					</label>
					<label class="toggle-row">
						<span>پیام جدید از سوی کاربر</span>
						<input type="checkbox" bind:checked={notifSettings.newMessage} class="toggle" />
					</label>
					<label class="toggle-row">
						<span>قطع صدا یا تصویر ارسالی</span>
						<input type="checkbox" bind:checked={notifSettings.muteToggle} class="toggle" />
					</label>
					<label class="toggle-row">
						<span>قطع ارتباط با سرور</span>
						<input type="checkbox" bind:checked={notifSettings.serverDisconnect} class="toggle" />
					</label>
				</div>
			{/if}
		</div>

		<div class="modal-footer">
			<button class="btn-confirm" onclick={onClose}>تایید</button>
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
		width: 400px;
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

	.tabs {
		display: flex;
		border-bottom: 1px solid rgba(255,255,255,0.08);
		padding: 0 16px;
	}
	.tabs button {
		background: none;
		border: none;
		padding: 10px 14px;
		color: #8a8a96;
		font-size: 0.82rem;
		cursor: pointer;
		border-bottom: 2px solid transparent;
		transition: all 0.15s;
		font-family: inherit;
	}
	.tabs button:hover { color: #e0e0e6; }
	.tabs button.active {
		color: #23b9d7;
		border-bottom-color: #23b9d7;
	}

	.modal-body { padding: 14px 16px; }

	.settings-group { display: flex; flex-direction: column; gap: 2px; }

	.toggle-row {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 8px 0;
		border-bottom: 1px solid rgba(255,255,255,0.04);
		font-size: 0.82rem;
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

	.setting-item {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 8px 0;
		border-bottom: 1px solid rgba(255,255,255,0.04);
		font-size: 0.82rem;
		color: #e0e0e6;
	}
	.setting-item:last-child { border-bottom: none; }

	.setting-select {
		background: #121822;
		color: #e0e0e6;
		border: 1px solid rgba(255,255,255,0.1);
		border-radius: 6px;
		padding: 5px 10px;
		font-size: 0.8rem;
		font-family: inherit;
		min-width: 140px;
	}

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
