<script lang="ts">
	import type { Participant, UserRole } from '$lib/classroom/types';
	import { ROLE_HIERARCHY, ROLE_LABELS } from '$lib/classroom/types';
	import UserRow from './UserRow.svelte';

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
			owner: [],
			admin: [],
			teacher: [],
			presenter: [],
			user: [],
		};

		for (const p of participants) {
			const role = p.role as UserRole;
			if (role === 'owner') groups.owner.push(p);
			else if (role === 'admin') groups.admin.push(p);
			else if (role === 'teacher') groups.teacher.push(p);
			else if (role === 'presenter') groups.presenter.push(p);
			else groups.user.push(p);
		}

		return groups;
	});

	const roleOrder: UserRole[] = ['owner', 'admin', 'teacher', 'presenter', 'user'];

	const isAdmin = $derived(currentUserRole === 'admin' || currentUserRole === 'owner');
	const isOwner = $derived(currentUserRole === 'owner');
</script>

<div class="w-[220px] flex flex-col shrink-0 border-l" style="background-color: #16213e; border-color: #2a2a4a;">
	<!-- Header -->
	<div class="px-3 py-2.5 border-b flex items-center justify-between" style="border-color: #2a2a4a;">
		<div class="flex items-center gap-2">
			<h3 class="font-bold text-xs text-gray-300">شرکت‌کنندگان ({participants.length})</h3>
			{#if handsRaisedCount > 0}
				<span class="text-[10px] px-1.5 py-0.5 rounded-full bg-yellow-500/20 text-yellow-400 font-medium">
					✋ {handsRaisedCount}
				</span>
			{/if}
		</div>
		<button onclick={onClose} class="text-gray-400 hover:text-white p-1" aria-label="بستن">
			<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" /></svg>
		</button>
	</div>

	<!-- User List -->
	<div class="flex-1 overflow-y-auto px-2 py-2 space-y-3">
		{#each roleOrder as role}
			{#if groupedParticipants[role].length > 0}
				<div>
					<p class="text-[10px] font-bold text-gray-500 px-2 mb-1 uppercase tracking-wider">
						{ROLE_LABELS[role]} ({groupedParticipants[role].length})
					</p>
					{#each groupedParticipants[role] as participant (participant.id)}
						<UserRow
							{participant}
							{isAdmin}
							{isOwner}
							{onMuteAudio}
							{onMuteVideo}
							{onKick}
						/>
					{/each}
				</div>
			{/if}
		{/each}

		{#if participants.length === 0}
			<p class="text-center text-gray-500 text-xs py-3">هنوز کسی متصل نیست</p>
		{/if}
	</div>
</div>
