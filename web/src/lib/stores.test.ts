import { describe, it, expect, vi, beforeEach } from 'vitest';

const mockLocalStorage = {
	getItem: vi.fn(),
	setItem: vi.fn(),
	removeItem: vi.fn(),
	clear: vi.fn(),
};

vi.stubGlobal('localStorage', mockLocalStorage);

vi.mock('$app/environment', () => ({
	browser: true
}));

describe('auth store', async () => {
	const { auth, isAdmin } = await import('$lib/stores');

	beforeEach(() => {
		vi.clearAllMocks();
		mockLocalStorage.getItem.mockReturnValue(null);
		auth.logout();
	});

	it('login stores tokens and user', () => {
		const user = { id: 1, email: 'test@test.com', display_name: 'Test', role: 'admin' as const, phone: '', is_active: true, created_at: '', updated_at: '' };
		const tokens = { access_token: 'access123', refresh_token: 'refresh123' };

		auth.login(user, tokens);

		expect(mockLocalStorage.setItem).toHaveBeenCalledWith('access_token', 'access123');
		expect(mockLocalStorage.setItem).toHaveBeenCalledWith('refresh_token', 'refresh123');
		expect(mockLocalStorage.setItem).toHaveBeenCalledWith('user', JSON.stringify(user));
	});

	it('logout clears tokens', () => {
		auth.logout();

		expect(mockLocalStorage.removeItem).toHaveBeenCalledWith('access_token');
		expect(mockLocalStorage.removeItem).toHaveBeenCalledWith('refresh_token');
		expect(mockLocalStorage.removeItem).toHaveBeenCalledWith('user');
	});

	it('init restores from localStorage', () => {
		const user = { id: 1, email: 'test@test.com', display_name: 'Test', role: 'operator' as const, phone: '', is_active: true, created_at: '', updated_at: '' };
		mockLocalStorage.getItem.mockImplementation((key: string) => {
			if (key === 'access_token') return 'token123';
			if (key === 'user') return JSON.stringify(user);
			return null;
		});

		auth.init();

		let state: any;
		auth.subscribe(s => state = s)();
		expect(state.isLoggedIn).toBe(true);
		expect(state.user?.role).toBe('operator');
	});

	it('isAdmin derived store', () => {
		const user = { id: 1, email: 'test@test.com', display_name: 'Test', role: 'admin' as const, phone: '', is_active: true, created_at: '', updated_at: '' };
		auth.login(user, { access_token: 't', refresh_token: 't' });

		let result = false;
		isAdmin.subscribe(v => result = v)();
		expect(result).toBe(true);
	});
});

describe('ROLE_PERMISSIONS', async () => {
	const { ROLE_PERMISSIONS } = await import('$lib/classroom/types');

	it('operator has all permissions', () => {
		const p = ROLE_PERMISSIONS.operator;
		expect(p.canMic).toBe(true);
		expect(p.canWebcam).toBe(true);
		expect(p.canScreenShare).toBe(true);
		expect(p.canWhiteboard).toBe(true);
		expect(p.canHandRaise).toBe(true);
		expect(p.canChat).toBe(true);
		expect(p.canKick).toBe(true);
	});

	it('presenter has mic/webcam/screen/whiteboard', () => {
		const p = ROLE_PERMISSIONS.presenter;
		expect(p.canMic).toBe(true);
		expect(p.canWebcam).toBe(true);
		expect(p.canScreenShare).toBe(true);
		expect(p.canWhiteboard).toBe(true);
		expect(p.canKick).toBe(false);
	});

	it('user has only hand raise and chat', () => {
		const p = ROLE_PERMISSIONS.user;
		expect(p.canMic).toBe(false);
		expect(p.canWebcam).toBe(false);
		expect(p.canScreenShare).toBe(false);
		expect(p.canWhiteboard).toBe(false);
		expect(p.canHandRaise).toBe(true);
		expect(p.canChat).toBe(true);
	});
});

describe('canMuteUser', async () => {
	const { canMuteUser } = await import('$lib/classroom/types');

	it('admin can mute user', () => {
		expect(canMuteUser('admin', 'user')).toBe(true);
	});

	it('user cannot mute admin', () => {
		expect(canMuteUser('user', 'admin')).toBe(false);
	});

	it('operator can mute presenter', () => {
		expect(canMuteUser('operator', 'presenter')).toBe(true);
	});
});
