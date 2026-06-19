<script lang="ts">
	import { api } from '$lib/api';
	import { auth } from '$lib/stores';
	import { onMount } from 'svelte';
	import type { Room, User } from '$lib/types';
	import ConfirmModal from '$lib/components/ConfirmModal.svelte';

	const toPersian = (n: number) => String(n).replace(/[0-9]/g, d => '۰۱۲۳۴۵۶۷۸۹'[+d]);

	let rooms = $state<Room[]>([]);
	let currentUser = $state<any>(null);
	auth.subscribe(s => { currentUser = s.user; });
	let userCounts = $state<Record<number, number>>({});
	let loading = $state(true);

	let searchQuery = $state('');

	// Create
	let showCreateModal = $state(false);
	let newRoom = $state({ name: '', description: '', color: '#3B82F6' });
	let createLoading = $state(false);

	// Edit
	let showEditModal = $state(false);
	let editingRoom = $state<Room | null>(null);
	let editForm = $state({ name: '', description: '', color: '#3B82F6', guest_login_enabled: true });
	let editLoading = $state(false);

	// Delete
	let showDeleteConfirm = $state(false);
	let deleteTargetId = $state(0);

	// Room settings
	let showSettingsModal = $state(false);
	let settingsRoom = $state<Room | null>(null);
	let roomSettings = $state({
		max_users: 50,
		recording_enabled: true,
		allow_student_video: false,
		allow_student_audio: true,
		allow_student_screen_share: false,
		allow_student_whiteboard: false,
		allow_student_chat: true,
		session_auto_end_minutes: 120,
		waiting_room_enabled: false,
	});
	let settingsLoading = $state(false);
	let settingsSaving = $state(false);

	onMount(() => loadRooms());

	async function loadRooms() {
		loading = true;
		const res = await api.get<{ items: Room[]; total: number }>('/rooms', { per_page: '100' });
		if (res.success && res.data) {
			rooms = res.data.items || [];
		}
		loading = false;

		for (const room of rooms) {
			const usersRes = await api.get<User[]>(`/rooms/${room.id}/users`);
			if (usersRes.success && Array.isArray(usersRes.data)) {
				userCounts[room.id] = usersRes.data.length;
			}
		}
	}

	function formatDate(d: string) {
		if (!d) return '';
		return new Date(d).toLocaleDateString('fa-IR', { year: 'numeric', month: 'long', day: 'numeric' });
	}

	let filteredRooms = $derived(rooms.filter(room => {
		const matchesSearch = !searchQuery || room.name.toLowerCase().includes(searchQuery.toLowerCase());
		return matchesSearch;
	}));

	const colorOptions = ['#3B82F6', '#10B981', '#F59E0B', '#EF4444', '#8B5CF6', '#EC4899', '#06B6D4', '#F97316'];

	// Create
	async function createRoom() {
		createLoading = true;
		const res = await api.post('/rooms', newRoom);
		if (res.success && res.data) {
			rooms = [res.data as Room, ...rooms];
			showCreateModal = false;
			newRoom = { name: '', description: '', color: '#3B82F6' };
		}
		createLoading = false;
	}

	// Edit
	function openEdit(room: Room) {
		editingRoom = room;
		editForm = { name: room.name, description: room.description || '', color: room.color || '#3B82F6', guest_login_enabled: room.guest_login_enabled };
		showEditModal = true;
	}

	async function saveEdit() {
		if (!editingRoom) return;
		editLoading = true;
		const res = await api.put(`/rooms/${editingRoom.id}`, editForm);
		if (res.success) {
			rooms = rooms.map(r => r.id === editingRoom!.id ? { ...r, ...editForm } : r);
			showEditModal = false;
			editingRoom = null;
		}
		editLoading = false;
	}

	// Delete
	async function deleteRoom(id: number) {
		const res = await api.delete(`/rooms/${id}`);
		if (res.success) rooms = rooms.filter(r => r.id !== id);
	}

	function confirmDeleteRoom(id: number) { deleteTargetId = id; showDeleteConfirm = true; }

	function copyLink(slug: string) {
		navigator.clipboard.writeText(`${window.location.origin}/room/${slug}`);
	}

	// Settings
	async function openSettings(room: Room) {
		settingsRoom = room;
		settingsLoading = true;
		showSettingsModal = true;

		const res = await api.get<any>(`/rooms/${room.id}/settings`);
		if (res.success && res.data) {
			roomSettings = { ...roomSettings, ...res.data };
		}
		settingsLoading = false;
	}

	async function saveSettings() {
		if (!settingsRoom) return;
		settingsSaving = true;
		await api.put(`/rooms/${settingsRoom.id}/settings`, roomSettings);
		settingsSaving = false;
		showSettingsModal = false;
	}
</script>

<div class="space-y-5">
	<div class="flex items-center justify-between">
		<div>
			<h1 style="font-size:1.5rem;font-weight:700;color:var(--color-midnight-sky);">مدیریت اتاق‌ها</h1>
			<p style="font-size:0.875rem;color:var(--color-mystic-sea);margin-top:4px;">{toPersian(rooms.length)} اتاق</p>
		</div>
		<button onclick={() => showCreateModal = true} class="sky-btn sky-btn-primary">
			<svg width="16" height="16" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" /></svg>
			اتاق جدید
		</button>
	</div>

	<!-- Search & View Toggle -->
	<div class="flex items-center gap-3 flex-wrap">
		<div class="sky-search flex-1 min-w-[200px]">
			<div class="sky-search-icon"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg></div>
			<input type="text" bind:value={searchQuery} class="sky-input" style="padding-right: 2.5rem;" placeholder="جستجو بر اساس نام اتاق..." />
		</div>
	</div>

	{#if loading}
		<div class="flex items-center justify-center py-16"><svg class="sky-spinner lg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="color: var(--color-crystal-clear);"><path d="M12 2v4M12 18v4M4.93 4.93l2.83 2.83M16.24 16.24l2.83 2.83M2 12h4M18 12h4M4.93 19.07l2.83-2.83M16.24 7.76l2.83-2.83"/></svg></div>
	{:else if filteredRooms.length === 0}
		<div class="sky-card"><div class="sky-empty"><div class="sky-empty-icon"><svg width="48" height="48" fill="none" stroke="currentColor" stroke-width="1" viewBox="0 0 24 24" style="color: var(--color-muted-mountain);"><path d="M2 12s3-7 10-7 10 7 10 7-3 7-10 7-10-7-10-7z"/><circle cx="12" cy="12" r="3"/></svg></div><p class="sky-empty-title">اتاقی یافت نشد</p><p class="sky-empty-desc">اولین اتاق خود را ایجاد کنید</p></div></div>
	{:else}
		<div class="sky-card overflow-hidden">
			<table class="sky-table">
				<thead><tr><th>اتاق</th><th>توضیحات</th><th>کاربران</th><th>مهمان</th><th>تاریخ</th><th>عملیات</th></tr></thead>
				<tbody>
					{#each filteredRooms as room}
						{@const userCount = userCounts[room.id] || 0}
						<tr>
							<td>
								<div class="flex items-center gap-3">
									<div style="width:36px;height:36px;border-radius:8px;display:flex;align-items:center;justify-content:center;color:white;font-weight:700;font-size:14px;background:{room.color};flex-shrink:0;">{room.name.charAt(0)}</div>
									<span class="font-semibold" style="color:var(--color-midnight-sky);">{room.name}</span>
								</div>
							</td>
							<td style="color:var(--color-mystic-sea);max-width:200px;" class="truncate">{room.description || '—'}</td>
							<td style="color:var(--color-mystic-sea);">{toPersian(userCount)}</td>
							<td><span class="sky-badge {room.guest_login_enabled ? 'sky-badge-success' : 'sky-badge-default'}" style="font-size:11px;">{room.guest_login_enabled ? 'فعال' : 'غیرفعال'}</span></td>
							<td style="color:var(--color-mystic-sea);font-size:12px;">{formatDate(room.created_at)}</td>
							<td>
								<div class="flex items-center gap-1">
									<a href="/room/{room.slug}" target="_blank" class="sky-btn sky-btn-primary" style="font-size:11px;padding:0.3rem 0.5rem;">ورود</a>
									<button onclick={() => openSettings(room)} class="sky-btn sky-btn-ghost" style="font-size:11px;">تنظیمات</button>
									<button onclick={() => openEdit(room)} class="sky-btn sky-btn-ghost" style="font-size:11px;">ویرایش</button>
									<button onclick={() => confirmDeleteRoom(room.id)} class="sky-btn-icon" style="width:28px;height:28px;"><svg width="14" height="14" fill="none" stroke="var(--color-fiery-passion)" stroke-width="1.75" viewBox="0 0 24 24"><polyline points="3 6 5 6 21 6"/><path d="M19 6l-1 14a2 2 0 01-2 2H8a2 2 0 01-2-2L5 6"/><path d="M10 11v6M14 11v6"/></svg></button>
								</div>
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	{/if}
</div>

<!-- Create Modal -->
{#if showCreateModal}
	<div class="modal-overlay" onclick={() => showCreateModal = false} role="button" tabindex="-1">
		<div class="modal-content" onclick={(e) => e.stopPropagation()}>
			<div class="sky-modal-header"><h2>ایجاد اتاق جدید</h2><button onclick={() => showCreateModal = false} class="sky-btn-icon"><svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg></button></div>
			<div class="sky-modal-body space-y-4">
				<div><label class="sky-label">نام اتاق</label><input type="text" bind:value={newRoom.name} class="sky-input" required /></div>
				<div><label class="sky-label">توضیحات</label><textarea bind:value={newRoom.description} class="sky-input resize-none" rows="2"></textarea></div>
				<div><label class="sky-label">رنگ</label><div class="flex gap-2 flex-wrap">{#each colorOptions as color}<button type="button" class="w-8 h-8 rounded-full transition-all" style="background: {color}; outline: {newRoom.color === color ? '3px solid var(--color-zen-garden)' : 'none'}; outline-offset: 2px; transform: scale({newRoom.color === color ? 1.15 : 1});" onclick={() => newRoom.color = color}></button>{/each}</div></div>
			</div>
			<div class="sky-modal-footer"><button onclick={() => showCreateModal = false} class="sky-btn sky-btn-secondary">انصراف</button><button onclick={createRoom} disabled={createLoading || !newRoom.name} class="sky-btn sky-btn-primary">{createLoading ? 'در حال ایجاد...' : 'ایجاد اتاق'}</button></div>
		</div>
	</div>
{/if}

<!-- Edit Modal -->
{#if showEditModal && editingRoom}
	<div class="modal-overlay" onclick={() => showEditModal = false} role="button" tabindex="-1">
		<div class="modal-content" onclick={(e) => e.stopPropagation()}>
			<div class="sky-modal-header"><h2>ویرایش اتاق</h2><button onclick={() => showEditModal = false} class="sky-btn-icon"><svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg></button></div>
			<div class="sky-modal-body space-y-4">
				<div><label class="sky-label">نام اتاق</label><input type="text" bind:value={editForm.name} class="sky-input" required /></div>
				<div><label class="sky-label">توضیحات</label><textarea bind:value={editForm.description} class="sky-input resize-none" rows="2"></textarea></div>
				<div><label class="sky-label">رنگ</label><div class="flex gap-2 flex-wrap">{#each colorOptions as color}<button type="button" class="w-8 h-8 rounded-full transition-all" style="background: {color}; outline: {editForm.color === color ? '3px solid var(--color-zen-garden)' : 'none'}; outline-offset: 2px; transform: scale({editForm.color === color ? 1.15 : 1});" onclick={() => editForm.color = color}></button>{/each}</div></div>
				<div class="flex items-center justify-between">
					<span class="text-sm font-medium" style="color:var(--color-ocean-wave);">ورود مهمان</span>
					<button onclick={() => editForm.guest_login_enabled = !editForm.guest_login_enabled} class="relative w-11 h-6 rounded-full transition-colors" style="background:{editForm.guest_login_enabled ? 'var(--color-crystal-clear)' : 'var(--color-muted-mountain)'};"><span class="absolute top-0.5 right-0.5 w-5 h-5 bg-white rounded-full transition-transform {editForm.guest_login_enabled ? 'translate-x-[-20px]' : ''}"></span></button>
				</div>
			</div>
			<div class="sky-modal-footer"><button onclick={() => showEditModal = false} class="sky-btn sky-btn-secondary">انصراف</button><button onclick={saveEdit} disabled={editLoading} class="sky-btn sky-btn-primary">{editLoading ? 'در حال ذخیره...' : 'ذخیره'}</button></div>
		</div>
	</div>
{/if}

<!-- Settings Modal -->
{#if showSettingsModal && settingsRoom}
	<div class="modal-overlay" onclick={() => showSettingsModal = false} role="button" tabindex="-1">
		<div class="modal-content" style="max-width:500px;" onclick={(e) => e.stopPropagation()}>
			<div class="sky-modal-header"><h2>تنظیمات اتاق — {settingsRoom.name}</h2><button onclick={() => showSettingsModal = false} class="sky-btn-icon"><svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg></button></div>
			<div class="sky-modal-body space-y-4">
				{#if settingsLoading}
					<div class="flex justify-center py-8"><div class="animate-spin w-6 h-6 border-2 border-[var(--color-crystal-clear)] border-t-transparent rounded-full"></div></div>
				{:else}
					<div><label class="sky-label">حداکثر کاربر</label><input type="number" bind:value={roomSettings.max_users} min="2" max="500" class="sky-input" /></div>

					<div class="space-y-2">
						<p class="sky-label">مجوزهای دانش‌آموز</p>
						{#each [
							{ key: 'allow_student_audio', label: 'میکروفون' },
							{ key: 'allow_student_video', label: 'وبکم' },
							{ key: 'allow_student_screen_share', label: 'اشتراک صفحه' },
							{ key: 'allow_student_whiteboard', label: 'تخته' },
							{ key: 'allow_student_chat', label: 'چت' },
						] as opt}
							<div class="flex items-center justify-between py-2">
								<span class="text-sm" style="color:var(--color-ocean-wave);">{opt.label}</span>
								<button onclick={() => (roomSettings as any)[opt.key] = !(roomSettings as any)[opt.key]} class="relative w-11 h-6 rounded-full transition-colors" style="background:{(roomSettings as any)[opt.key] ? 'var(--color-crystal-clear)' : 'var(--color-muted-mountain)'};"><span class="absolute top-0.5 right-0.5 w-5 h-5 bg-white rounded-full transition-transform {(roomSettings as any)[opt.key] ? 'translate-x-[-20px]' : ''}"></span></button>
							</div>
						{/each}
					</div>

					<div class="flex items-center justify-between py-2">
						<span class="text-sm" style="color:var(--color-ocean-wave);">ضبط جلسات</span>
						<button onclick={() => roomSettings.recording_enabled = !roomSettings.recording_enabled} class="relative w-11 h-6 rounded-full transition-colors" style="background:{roomSettings.recording_enabled ? 'var(--color-crystal-clear)' : 'var(--color-muted-mountain)'};"><span class="absolute top-0.5 right-0.5 w-5 h-5 bg-white rounded-full transition-transform {roomSettings.recording_enabled ? 'translate-x-[-20px]' : ''}"></span></button>
					</div>

					<div class="flex items-center justify-between py-2">
						<span class="text-sm" style="color:var(--color-ocean-wave);">اتاق انتظار</span>
						<button onclick={() => roomSettings.waiting_room_enabled = !roomSettings.waiting_room_enabled} class="relative w-11 h-6 rounded-full transition-colors" style="background:{roomSettings.waiting_room_enabled ? 'var(--color-crystal-clear)' : 'var(--color-muted-mountain)'};"><span class="absolute top-0.5 right-0.5 w-5 h-5 bg-white rounded-full transition-transform {roomSettings.waiting_room_enabled ? 'translate-x-[-20px]' : ''}"></span></button>
					</div>

					<div><label class="sky-label">پایان خودکار جلسه (دقیقه)</label><input type="number" bind:value={roomSettings.session_auto_end_minutes} min="30" max="480" class="sky-input" /></div>
				{/if}
			</div>
			<div class="sky-modal-footer">
				<button onclick={() => showSettingsModal = false} class="sky-btn sky-btn-secondary">انصراف</button>
				<button onclick={saveSettings} disabled={settingsSaving || settingsLoading} class="sky-btn sky-btn-primary">{settingsSaving ? 'در حال ذخیره...' : 'ذخیره'}</button>
			</div>
		</div>
	</div>
{/if}

<ConfirmModal bind:show={showDeleteConfirm} title="حذف اتاق" message="آیا از حذف این اتاق اطمینان دارید؟" onConfirm={() => { showDeleteConfirm = false; deleteRoom(deleteTargetId); }} onCancel={() => {}} />
