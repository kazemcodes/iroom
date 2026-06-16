import { writable, derived } from 'svelte/store';
import { browser } from '$app/environment';
import type { User } from './types';

function createAuthStore() {
	const { subscribe, set, update } = writable<{
		user: User | null;
		isLoggedIn: boolean;
	}>({
		user: null,
		isLoggedIn: false
	});

	return {
		subscribe,
		login: (user: User, tokens: { access_token: string; refresh_token: string }) => {
			if (browser) {
				localStorage.setItem('access_token', tokens.access_token);
				localStorage.setItem('refresh_token', tokens.refresh_token);
				localStorage.setItem('user', JSON.stringify(user));
			}
			set({ user, isLoggedIn: true });
		},
		logout: () => {
			if (browser) {
				localStorage.removeItem('access_token');
				localStorage.removeItem('refresh_token');
				localStorage.removeItem('user');
			}
			set({ user: null, isLoggedIn: false });
		},
		init: () => {
			if (browser) {
				const token = localStorage.getItem('access_token');
				const savedUser = localStorage.getItem('user');
				if (token && savedUser) {
					set({ user: JSON.parse(savedUser), isLoggedIn: true });
				}
			}
		}
	};
}

export const auth = createAuthStore();

export const isAdmin = derived(auth, ($auth) => $auth.user?.role === 'admin');
export const isTeacher = derived(auth, ($auth) => $auth.user?.role === 'teacher');
export const isStudent = derived(auth, ($auth) => $auth.user?.role === 'student');

export const sidebarOpen = writable(true);
