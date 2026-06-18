<!--
  UsersPanel — Displays connected participants in the classroom.
  
  Features:
    - Groups users by role (owner > admin > teacher > presenter > student)
    - Shows role-colored avatars
    - Displays hand-raise indicator
    - Shows total participant count
    - Empty state when no one is connected
  
  Props:
    participants: Array of connected participants
    currentUserRole: Current user's role (for permission checks)
    onClose: Callback when panel should close
-->
<script lang="ts">
	import type { Participant, UserRole } from '$lib/classroom/types';
	import { ROLE_LABELS } from '$lib/classroom/types';

	let {
		participants = [],
		currentUserRole = 'student',
		onMuteAudio,
		onMuteVideo,
		onKick,
		onClose,
	}: {
		participants: Participant[];
		currentUserRole: string;
		onMuteAudio?: (id: string) => void;
		onMuteVideo?: (id: string) => void;
		onKick?: (id: string) => void;
		onClose: () => void;
	} = $props();

	const handsRaisedCount = $derived(participants.filter(p => p.handRaised).length);

	const groupedParticipants = $derived.by(() => {
		const groups: Record<string, Participant[]> = {
			owner: [], admin: [], teacher: [], presenter: [], user: [],
		};
		for (const p of participants) {
			const role = p.role as UserRole;
			if (groups[role]) groups[role].push(p);
			else groups.user.push(p);
		}
		return groups;
	});

	const roleOrder: UserRole[] = ['owner', 'admin', 'teacher', 'presenter', 'user'];
	const isAdmin = $derived(currentUserRole === 'admin' || currentUserRole === 'owner');
</script>

<div class="users-panel">
	<div class="users-header">
		<div class="flex items-center gap-2">
			<svg width="16" height="16" style="fill:#8a8a96;"><use xlink:href="#shape_group"></use></svg>
			<span style="font-size:0.75rem;color:#8a8a96;">کاربران</span>
			<span style="font-size:0.65rem;padding:1px 6px;border-radius:10px;background:rgba(255,255,255,0.06);color:#e0e0e6;">{participants.length}</span>
			{#if handsRaisedCount > 0}
				<span style="font-size:0.65rem;padding:1px 6px;border-radius:10px;background:rgba(245,158,11,0.2);color:#f59e0b;">✋ {handsRaisedCount}</span>
			{/if}
		</div>
		<button onclick={onClose} class="close-btn" aria-label="بستن">
			<svg width="14" height="14"><use xlink:href="#shape_clear"></use></svg>
		</button>
	</div>

	<div class="users-list">
		{#each roleOrder as role}
			{#if groupedParticipants[role].length > 0}
				{#each groupedParticipants[role] as p (p.id)}
					<div class="user-row">
						<div class="user-avatar" class:role-owner={role === 'owner'} class:role-admin={role === 'admin'} class:role-teacher={role === 'teacher'} class:role-presenter={role === 'presenter'}>
							<svg width="14" height="14" style="fill:currentColor;"><use xlink:href="#shape_person"></use></svg>
						</div>
						<div class="user-info">
							<span class="user-name">{p.name}</span>
							<span class="user-role">{ROLE_LABELS[role as UserRole] || role}</span>
						</div>
						{#if p.handRaised}
							<span style="font-size:0.7rem;">✋</span>
						{/if}
					</div>
				{/each}
			{/if}
		{/each}

		{#if participants.length === 0}
			<div style="padding:24px;text-align:center;">
				<p style="font-size:0.7rem;color:#8a8a96;">هنوز کسی متصل نیست</p>
			</div>
		{/if}
	</div>
</div>

<style>
	.users-panel { display: flex; flex-direction: column; height: 100%; }
	.users-header { display: flex; align-items: center; justify-content: space-between; padding: 10px 12px; border-bottom: 1px solid rgba(255,255,255,0.06); }
	.close-btn { background: none; border: none; cursor: pointer; padding: 4px; display: flex; align-items: center; justify-content: center; border-radius: 4px; }
	.close-btn svg { fill: #8a8a96; }
	.close-btn:hover { background: rgba(255,255,255,0.06); }
	.close-btn:hover svg { fill: #e0e0e6; }
	.users-list { flex: 1; overflow-y: auto; padding: 4px; }
	.user-row { display: flex; align-items: center; gap: 8px; padding: 6px 8px; border-radius: 6px; transition: background 0.12s; }
	.user-row:hover { background: rgba(255,255,255,0.04); }
	.user-avatar { width: 26px; height: 26px; border-radius: 50%; display: flex; align-items: center; justify-content: center; background: rgba(255,255,255,0.08); color: #8a8a96; flex-shrink: 0; }
	.role-owner { background: rgba(245,158,11,0.2); color: #f59e0b; }
	.role-admin { background: rgba(239,68,68,0.2); color: #ef4444; }
	.role-teacher { background: rgba(139,92,246,0.2); color: #8b5cf6; }
	.role-presenter { background: rgba(35,185,215,0.2); color: #23b9d7; }
	.user-info { flex: 1; min-width: 0; display: flex; flex-direction: column; }
	.user-name { font-size: 0.75rem; font-weight: 600; color: #e0e0e6; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
	.user-role { font-size: 0.65rem; color: #8a8a96; }
</style>
