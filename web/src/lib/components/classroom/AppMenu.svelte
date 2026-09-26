<!--
  AppMenu — Hamburger menu for classroom controls.
  
  Shows:
    - اطلاعات کاربری (User Info)
    - وضعیت اتصال (Connection Status)
    - تنظیمات (Settings)
    - چیدمان (Layout) — operator only
    - خروج (Leave)
    - بستن اتاق (Close Room) — operator only
  
  Props:
    userRole: Current user's role in the classroom
    onUserInfo, onConnectionStatus, onSettings, onLayout: Callbacks for menu items
    onLeave, onCloseRoom: Callbacks for exit actions
    onDismiss: Called when menu should close
-->
<script lang="ts">
	import type { UserRole } from '$lib/classroom/types';

	let {
		userRole = 'user',
		onUserInfo,
		onConnectionStatus,
		onSettings,
		onLayout,
		onLeave,
		onCloseRoom,
		onDismiss,
	}: {
		userRole: UserRole;
		onUserInfo: () => void;
		onConnectionStatus: () => void;
		onSettings: () => void;
		onLayout: () => void;
		onLeave: () => void;
		onCloseRoom: () => void;
		onDismiss: () => void;
	} = $props();

	const isOperator = $derived(['owner', 'admin', 'operator'].includes(userRole));

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') onDismiss();
	}
</script>

<svelte:window onkeydown={handleKeydown} />

<div class="app-menu">
	<ul>
		<li onclick={() => { onUserInfo(); onDismiss(); }}>
			<svg width="20" height="20"><use xlink:href="#shape_info_outline"></use></svg>
			<span>اطلاعات کاربری</span>
		</li>
		<li onclick={() => { onConnectionStatus(); onDismiss(); }}>
			<svg width="20" height="20"><use xlink:href="#shape_network_check"></use></svg>
			<span>وضعیت اتصال</span>
		</li>
		<li onclick={() => { onSettings(); onDismiss(); }}>
			<svg width="20" height="20"><use xlink:href="#shape_settings"></use></svg>
			<span>تنظیمات</span>
		</li>
		{#if isOperator}
			<li onclick={() => { onLayout(); onDismiss(); }}>
				<svg width="20" height="20"><use xlink:href="#shape_web"></use></svg>
				<span>چیدمان</span>
			</li>
		{/if}
		<li class="separator"></li>
		<li onclick={() => { onLeave(); onDismiss(); }}>
			<svg width="20" height="20"><use xlink:href="#shape_exit"></use></svg>
			<span>خروج</span>
		</li>
		{#if isOperator}
			<li class="danger" onclick={() => { onCloseRoom(); onDismiss(); }}>
				<svg width="20" height="20"><use xlink:href="#shape_power_settings_new"></use></svg>
				<span>بستن اتاق</span>
			</li>
		{/if}
	</ul>
</div>

<style>
	.app-menu {
		position: absolute;
		top: 48px;
		right: 8px;
		background: #1c2a3a;
		border-radius: 10px;
		box-shadow: 0 8px 32px rgba(0,0,0,0.4);
		z-index: 100;
		min-width: 200px;
		padding: 6px 0;
		animation: menuFadeIn 0.15s ease;
	}

	@keyframes menuFadeIn {
		from { opacity: 0; transform: translateY(-8px); }
		to { opacity: 1; transform: translateY(0); }
	}

	ul {
		list-style: none;
		margin: 0;
		padding: 0;
	}

	li {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 10px 16px;
		cursor: pointer;
		color: #e0e0e6;
		font-size: 0.875rem;
		transition: background 0.15s;
	}

	li:hover {
		background: rgba(255,255,255,0.06);
	}

	li svg {
		fill: #8a8a96;
		flex-shrink: 0;
	}

	li:hover svg {
		fill: #23b9d7;
	}

	.separator {
		height: 1px;
		background: rgba(255,255,255,0.08);
		margin: 4px 0;
		padding: 0;
		cursor: default;
	}

	.separator:hover {
		background: rgba(255,255,255,0.08);
	}

	.danger {
		color: #e05252;
	}

	.danger svg {
		fill: #e05252;
	}

	.danger:hover {
		background: rgba(224, 82, 82, 0.1);
	}
</style>
