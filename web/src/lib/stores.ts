/**
 * Svelte Stores — Global state management for the IRoom frontend.
 *
 * Stores:
 *   - auth: User authentication state (login/logout/init)
 *   - isAdmin / isTeacher / isStudent: Derived role checks
 *   - sidebarOpen: Sidebar toggle state
 *
 * Auth flow:
 *   1. auth.login(user, tokens) — stores tokens in localStorage
 *   2. auth.init() — restores state from localStorage on page load
 *   3. auth.logout() — clears tokens and resets state
 *
 * Usage:
 *   import { auth, isAdmin } from '$lib/stores';
 *   auth.subscribe(state => { if (state.isLoggedIn) ... });
 */
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

export const sidebarOpen = writable(true);
