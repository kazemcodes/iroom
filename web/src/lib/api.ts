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

		if (res.status === 401 && browser) {
			localStorage.removeItem('access_token');
			localStorage.removeItem('refresh_token');
			localStorage.removeItem('user');
			if (token && window.location.pathname !== '/') {
				window.location.href = '/';
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

		if (res.status === 401 && browser) {
			localStorage.removeItem('access_token');
			localStorage.removeItem('refresh_token');
			localStorage.removeItem('user');
			if (token && window.location.pathname !== '/') {
				window.location.href = '/';
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
