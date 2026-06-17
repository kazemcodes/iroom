<script lang="ts">
	import type { Participant, UserRole } from '$lib/classroom/types';
	import { ROLE_LABELS } from '$lib/classroom/types';
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

<div class="flex flex-col h-full" style="background-color: #252540;">
	<!-- Block Header -->
	<div class="flex items-center justify-between px-3 py-2.5 shrink-0" style="border-bottom: 1px solid #3a3a5a;">
		<div class="flex items-center gap-2">
			<svg class="w-4 h-4 text-[#94a3b8]" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z" /></svg>
			<span class="text-xs font-medium text-[#94a3b8]">کاربران</span>
			<span class="text-[10px] px-1.5 py-0.5 rounded-full bg-[#3a3a5a] text-[#e2e8f0] font-medium">{participants.length}</span>
			{#if handsRaisedCount > 0}
				<span class="text-[10px] px-1.5 py-0.5 rounded-full bg-[#d7911d]/20 text-[#d7911d] font-medium">✋ {handsRaisedCount}</span>
			{/if}
		</div>
		<button onclick={onClose} class="text-[#94a3b8] hover:text-[#e2e8f0] p-1" aria-label="بستن">
			<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" /></svg>
		</button>
	</div>

	<!-- User List -->
	<div class="flex-1 overflow-y-auto">
		{#each roleOrder as role}
			{#if groupedParticipants[role].length > 0}
				<div>
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
			<div class="px-4 py-8 text-center">
				<p class="text-xs text-[#94a3b8]">هنوز کسی متصل نیست</p>
			</div>
		{/if}
	</div>
</div>
