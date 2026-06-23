<script lang="ts">
	import { page } from '$app/state';
	import { api } from '$lib/api';
	import { onMount } from 'svelte';
	import type { User, Room } from '$lib/types';
	import { toPersianNum } from '$lib/utils/persian';

	const roomId = $derived(Number(page.params.id));
	let room = $state<Room | null>(null);
	let users = $state<User[]>([]);
	let allUsers = $state<User[]>([]);
	let loading = $state(true);
	let showAddModal = $state(false);
	let selectedUserId = $state(0);

	onMount(loadData);

	async function loadData() {
		loading = true;
		const [roomRes, usersRes, allRes] = await Promise.all([
			api.get<Room>(`/rooms/${roomId}`),
			api.get<User[]>(`/rooms/${roomId}/users`),
			api.get<{ items: User[] }>('/admin/users', { per_page: '1000' })
		]);
		if (roomRes.success) room = roomRes.data;
		if (usersRes.success && Array.isArray(usersRes.data)) users = usersRes.data;
		if (allRes.success && allRes.data) allUsers = allRes.data.items || [];
		loading = false;
	}

	async function addUser() {
		if (!selectedUserId) return;
		await api.post(`/rooms/${roomId}/users`, { user_id: selectedUserId, role: 'user' });
		selectedUserId = 0;
		showAddModal = false;
		await loadData();
	}

	async function removeUser(userId: number) {
		if (!confirm('آیا از حذف این کاربر اطمینان دارید؟')) return;
		await api.delete(`/rooms/${roomId}/users/${userId}`);
		await loadData();
	}

	const availableUsers = $derived(allUsers.filter(u => !users.find(ur => ur.id === u.id)));
</script>

<div class="space-y-5">
	<div class="flex items-center justify-between">
		<div>
			<h1 class="sky-page-title">مدیریت کاربران اتاق</h1>
			<p class="sky-page-subtitle">{room?.name || ''} — {toPersianNum(users.length)} کاربر</p>
		</div>
		<div class="flex gap-2">
			<a href="/admin/rooms" class="sky-btn sky-btn-secondary">بازگشت</a>
			<button onclick={() => showAddModal = true} class="sky-btn sky-btn-primary">افزودن کاربر</button>
		</div>
	</div>

	{#if loading}
		<div class="flex justify-center py-16"><svg class="sky-spinner lg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="color: var(--color-crystal-clear);"><path d="M12 2v4M12 18v4M4.93 4.93l2.83 2.83M16.24 16.24l2.83 2.83M2 12h4M18 12h4M4.93 19.07l2.83-2.83M16.24 7.76l2.83-2.83"/></svg></div>
	{:else if users.length === 0}
		<div class="sky-card"><div class="sky-empty"><p class="sky-empty-desc">کاربری اختصاص داده نشده</p></div></div>
	{:else}
		<div class="sky-card">
			<table class="sky-table">
				<thead><tr><th>نام</th><th>ایمیل</th><th>عملیات</th></tr></thead>
				<tbody>
					{#each users as u}
						<tr>
							<td class="font-semibold">{u.display_name}</td>
							<td style="color: var(--color-mystic-sea);">{u.email}</td>
							<td>
								<button onclick={() => removeUser(u.id)} class="sky-btn-icon" style="width:32px;height:32px;">
									<svg width="15" height="15" fill="none" stroke="var(--color-fiery-passion)" stroke-width="1.75" viewBox="0 0 24 24"><polyline points="3 6 5 6 21 6"/><path d="M19 6l-1 14a2 2 0 01-2 2H8a2 2 0 01-2-2L5 6"/><path d="M10 11v6M14 11v6"/></svg>
								</button>
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	{/if}
</div>

{#if showAddModal}
	<div class="modal-overlay" onclick={() => showAddModal = false} role="button" tabindex="-1">
		<div class="modal-content" onclick={(e) => e.stopPropagation()}>
			<div class="sky-modal-header">
				<h2>افزودن کاربر</h2>
				<button onclick={() => showAddModal = false} class="sky-btn-icon"><svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg></button>
			</div>
			<div class="sky-modal-body">
				<select bind:value={selectedUserId} class="sky-input">
					<option value={0}>انتخاب کاربر</option>
					{#each availableUsers as u}
						<option value={u.id}>{u.display_name} ({u.email})</option>
					{/each}
				</select>
			</div>
			<div class="sky-modal-footer">
				<button onclick={() => showAddModal = false} class="sky-btn sky-btn-secondary">انصراف</button>
				<button onclick={addUser} disabled={!selectedUserId} class="sky-btn sky-btn-primary">افزودن</button>
			</div>
		</div>
	</div>
{/if}
