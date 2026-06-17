import { writable, derived } from 'svelte/store';
import { browser } from '$app/environment';
import { api } from '$lib/api';

export interface Notification {
	id: number;
	user_id: number;
	type: string;
	title: string;
	message?: string;
	data?: string;
	is_read: boolean;
	created_at: string;
}

function createNotificationStore() {
	const { subscribe, set, update } = writable<Notification[]>([]);

	return {
		subscribe,
		load: async () => {
			if (!browser) return;
			try {
				const res = await api.get<Notification[]>('/notifications');
				if (res.success && res.data) set(res.data);
			} catch {}
		},
		markRead: async (id: number) => {
			await api.post(`/notifications/${id}/read`, {});
			update((list) => list.map((n) => (n.id === id ? { ...n, is_read: true } : n)));
		},
		markAllRead: async () => {
			await api.post('/notifications/read-all', {});
			update((list) => list.map((n) => ({ ...n, is_read: true })));
		},
		add: (n: Notification) => {
			update((list) => [n, ...list]);
		},
		reset: () => set([])
	};
}

export const notifications = createNotificationStore();

export const unreadCount = derived(notifications, ($n) => $n.filter((x) => !x.is_read).length);
