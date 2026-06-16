import { browser } from '$app/environment';
import type { APIResponse } from './types';

const BASE = 'http://localhost:8080/api/v1';

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
	let url = BASE + path;
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
			window.location.href = '/';
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
	put: <T>(path: string, body?: any) => request<T>('PUT', path, body),
	delete: <T>(path: string) => request<T>('DELETE', path)
};
