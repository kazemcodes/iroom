/**
 * API Client — Centralized HTTP request handler for the IRoom frontend.
 *
 * Features:
 *   - Auto-attaches JWT token from localStorage to all requests
 *   - Returns typed responses via APIResponse<T>
 *   - Handles 401 by logging out the user
 *   - Provides convenience methods: api.get(), api.post(), api.put(), api.delete()
 *
 * Usage:
 *   const res = await api.get<User>('/auth/me');
 *   if (res.success) { console.log(res.data); }
 *
 * All requests go to /api/v1/* which is proxied to the Go backend in dev mode.
 */
import { browser } from '$app/environment';
import { auth } from './stores';
import type { APIResponse } from './types';

function getBaseUrl(): string {
	if (!browser) return '';
	return window.location.origin + '/api/v1';
}

function getWsUrl(): string {
	if (!browser) return '';
	const proto = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
	return `${proto}//${window.location.host}`;
}

function getApiUrl(path: string): string {
	return getBaseUrl() + path;
}

function getToken(): string | null {
	if (!browser) return null;
	return localStorage.getItem('access_token');
}

// Mutex for refresh token flow — prevents concurrent refresh attempts
let refreshPromise: Promise<boolean> | null = null;

/**
 * Attempt to refresh the access token using the stored refresh token.
 * Returns true if refresh succeeded, false otherwise.
 * Only one refresh is in flight at a time — concurrent callers wait for the same promise.
 */
async function tryRefreshToken(): Promise<boolean> {
	// If a refresh is already in progress, wait for it
	if (refreshPromise) return refreshPromise;

	refreshPromise = (async () => {
		const refreshToken = localStorage.getItem('refresh_token');
		if (!refreshToken) return false;

		try {
			const res = await fetch(getApiUrl('/auth/refresh'), {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ refresh_token: refreshToken }),
			});

			if (!res.ok) return false;

			const data = await res.json();
			if (data.success && data.data) {
				const tokens = data.data;
				localStorage.setItem('access_token', tokens.access_token);
				localStorage.setItem('refresh_token', tokens.refresh_token);
				auth.updateTokens({
					access_token: tokens.access_token,
					refresh_token: tokens.refresh_token,
				});
				return true;
			}
			return false;
		} catch {
			return false;
		} finally {
			refreshPromise = null;
		}
	})();

	return refreshPromise;
}

async function request<T>(
	method: string,
	path: string,
	body?: any,
	params?: Record<string, string>
): Promise<APIResponse<T>> {
	let url = getApiUrl(path);
	if (params) {
		const qs = new URLSearchParams(params).toString();
		if (qs) url += '?' + qs;
	}

	const headers: Record<string, string> = {
		'Content-Type': 'application/json'
	};

	const token = getToken();
	if (token) {
		headers['Authorization'] = `Bearer ${token}`;
	}

	try {
		const res = await fetch(url, {
			method,
			headers,
			body: body ? JSON.stringify(body) : undefined
		});

		if (res.status === 401 && browser && token) {
			// Try to refresh the token before giving up
			const refreshed = await tryRefreshToken();
			if (refreshed) {
				// Retry the original request with the new token
				const newToken = getToken();
				if (newToken) {
					headers['Authorization'] = `Bearer ${newToken}`;
				}
				try {
					const retryRes = await fetch(url, {
						method,
						headers,
						body: body ? JSON.stringify(body) : undefined
					});
					if (retryRes.ok) {
						const data = await retryRes.json();
						return data;
					}
					// If retry also fails with 401, the new token is also invalid — log out
					if (retryRes.status === 401) {
						localStorage.removeItem('access_token');
						localStorage.removeItem('refresh_token');
						localStorage.removeItem('user');
						if (window.location.pathname !== '/auth' && window.location.pathname !== '/') {
							window.location.href = '/auth';
						}
						return { success: false, error: 'توکن منقضی شده — لطفاً دوباره وارد شوید' };
					}
					// Non-401 error on retry — return the result
					const retryData = await retryRes.json().catch(() => null);
					return retryData || { success: false, error: 'خطا در سرور' };
				} catch (e) {
					return { success: false, error: 'خطا در اتصال به سرور' };
				}
			}

			// Refresh failed — clear auth state and redirect
			localStorage.removeItem('access_token');
			localStorage.removeItem('refresh_token');
			localStorage.removeItem('user');
			if (token && window.location.pathname !== '/auth' && window.location.pathname !== '/') {
				window.location.href = '/auth';
			}
			return { success: false, error: 'توکن منقضی شده' };
		}

		const data = await res.json();
		return data;
	} catch (e) {
		return { success: false, error: 'خطا در اتصال به سرور' };
	}
}

async function postFormData<T>(path: string, formData: FormData): Promise<APIResponse<T>> {
	let url = getApiUrl(path);

	const headers: Record<string, string> = {};
	const token = getToken();
	if (token) {
		headers['Authorization'] = `Bearer ${token}`;
	}

	try {
		const res = await fetch(url, {
			method: 'POST',
			headers,
			body: formData
		});

		if (res.status === 401 && browser && token) {
			// Try to refresh the token before giving up
			const refreshed = await tryRefreshToken();
			if (refreshed) {
				// Retry the original request with the new token
				const newToken = getToken();
				if (newToken) {
					headers['Authorization'] = `Bearer ${newToken}`;
				}
				try {
					const retryRes = await fetch(url, {
						method: 'POST',
						headers,
						body: formData
					});
					if (retryRes.ok) {
						const data = await retryRes.json();
						return data;
					}
					if (retryRes.status === 401) {
						localStorage.removeItem('access_token');
						localStorage.removeItem('refresh_token');
						localStorage.removeItem('user');
						if (window.location.pathname !== '/auth' && window.location.pathname !== '/') {
							window.location.href = '/auth';
						}
						return { success: false, error: 'توکن منقضی شده — لطفاً دوباره وارد شوید' };
					}
					const retryData = await retryRes.json().catch(() => null);
					return retryData || { success: false, error: 'خطا در سرور' };
				} catch (e) {
					return { success: false, error: 'خطا در اتصال به سرور' };
				}
			}

			// Refresh failed — clear auth state and redirect
			localStorage.removeItem('access_token');
			localStorage.removeItem('refresh_token');
			localStorage.removeItem('user');
			if (token && window.location.pathname !== '/auth' && window.location.pathname !== '/') {
				window.location.href = '/auth';
			}
			return { success: false, error: 'توکن منقضی شده' };
		}

		const data = await res.json();
		return data;
	} catch (e) {
		return { success: false, error: 'خطا در اتصال به سرور' };
	}
}

export const api = {
	get: <T>(path: string, params?: Record<string, string>) => request<T>('GET', path, undefined, params),
	post: <T>(path: string, body?: any) => request<T>('POST', path, body),
	postFormData: <T>(path: string, formData: FormData) => postFormData<T>(path, formData),
	put: <T>(path: string, body?: any) => request<T>('PUT', path, body),
	delete: <T>(path: string) => request<T>('DELETE', path),
	getWsUrl: () => getWsUrl(),
	getBaseUrl: () => getBaseUrl(),
};
