/**
 * Toast Store — Global notification system for the IRoom frontend.
 *
 * Shows temporary notification messages (success, error, info, warning).
 * Auto-dismisses after duration. Can be manually dismissed.
 *
 * Usage:
 *   import { toasts } from '$lib/stores/toast';
 *   toasts.addToast('ذخیره شد', 'success');
 *   toasts.addToast('خطا در اتصال', 'error', 6000);
 *   toasts.removeToast(id);
 *
 * Types:
 *   - success: Green notification
 *   - error: Red notification
 *   - info: Blue notification
 *   - warning: Yellow notification
 */
import { writable } from 'svelte/store';

export type ToastType = 'success' | 'error' | 'info' | 'warning';

export interface Toast {
	id: number;
	message: string;
	type: ToastType;
	duration: number;
}

let nextId = 0;

function createToastStore() {
	const { subscribe, update } = writable<Toast[]>([]);

	return {
		subscribe,
		addToast(message: string, type: ToastType = 'info', duration = 4000) {
			const id = nextId++;
			update((t) => [...t, { id, message, type, duration }]);
			if (duration > 0) {
				setTimeout(() => {
					update((t) => t.filter((toast) => toast.id !== id));
				}, duration);
			}
			return id;
		},
		removeToast(id: number) {
			update((t) => t.filter((toast) => toast.id !== id));
		}
	};
}

export const toasts = createToastStore();
