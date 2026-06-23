import { describe, it, expect } from 'vitest';
import { ROLE_PERMISSIONS, ROLE_HIERARCHY, ROLE_LABELS, canMuteUser, canKickUser } from '$lib/classroom/types';

describe('ROLE_HIERARCHY', () => {
	it('owner has highest rank (0)', () => {
		expect(ROLE_HIERARCHY.owner).toBe(0);
	});

	it('user has lowest rank', () => {
		expect(ROLE_HIERARCHY.user).toBeGreaterThan(ROLE_HIERARCHY.presenter);
	});

	it('operator is between admin and presenter', () => {
		expect(ROLE_HIERARCHY.operator).toBeGreaterThan(ROLE_HIERARCHY.admin);
		expect(ROLE_HIERARCHY.operator).toBeLessThan(ROLE_HIERARCHY.presenter);
	});
});

describe('ROLE_LABELS', () => {
	it('has labels for all roles', () => {
		expect(ROLE_LABELS.owner).toBeDefined();
		expect(ROLE_LABELS.admin).toBeDefined();
		expect(ROLE_LABELS.operator).toBeDefined();
		expect(ROLE_LABELS.presenter).toBeDefined();
		expect(ROLE_LABELS.user).toBeDefined();
	});

	it('labels are in Persian', () => {
		expect(ROLE_LABELS.owner).toBe('مالک');
		expect(ROLE_LABELS.admin).toBe('مدیر');
		expect(ROLE_LABELS.operator).toBe('اپراتور');
		expect(ROLE_LABELS.presenter).toBe('ارائه‌دهنده');
		expect(ROLE_LABELS.user).toBe('کاربر عادی');
	});
});

describe('canKickUser', () => {
	it('owner can kick anyone', () => {
		expect(canKickUser('owner', 'admin')).toBe(true);
		expect(canKickUser('owner', 'user')).toBe(true);
	});

	it('admin can kick lower roles', () => {
		expect(canKickUser('admin', 'user')).toBe(true);
		expect(canKickUser('admin', 'presenter')).toBe(true);
	});

	it('presenter cannot kick', () => {
		expect(canKickUser('presenter', 'user')).toBe(false);
	});
});

describe('ROLE_PERMISSIONS completeness', () => {
	it('all roles have permission entries', () => {
		const roles = ['owner', 'admin', 'operator', 'presenter', 'user'] as const;
		for (const role of roles) {
			expect(ROLE_PERMISSIONS[role]).toBeDefined();
			expect(typeof ROLE_PERMISSIONS[role].canMic).toBe('boolean');
			expect(typeof ROLE_PERMISSIONS[role].canWebcam).toBe('boolean');
			expect(typeof ROLE_PERMISSIONS[role].canScreenShare).toBe('boolean');
			expect(typeof ROLE_PERMISSIONS[role].canWhiteboard).toBe('boolean');
			expect(typeof ROLE_PERMISSIONS[role].canHandRaise).toBe('boolean');
			expect(typeof ROLE_PERMISSIONS[role].canChat).toBe('boolean');
			expect(typeof ROLE_PERMISSIONS[role].canWatch).toBe('boolean');
		}
	});

	it('user cannot mic or webcam', () => {
		expect(ROLE_PERMISSIONS.user.canMic).toBe(false);
		expect(ROLE_PERMISSIONS.user.canWebcam).toBe(false);
	});

	it('presenter can mic and webcam', () => {
		expect(ROLE_PERMISSIONS.presenter.canMic).toBe(true);
		expect(ROLE_PERMISSIONS.presenter.canWebcam).toBe(true);
	});

	it('user cannot kick or change roles', () => {
		expect(ROLE_PERMISSIONS.user.canKick).toBe(false);
		expect(ROLE_PERMISSIONS.user.canChangeRole).toBe(false);
	});

	it('operator has full permissions', () => {
		expect(ROLE_PERMISSIONS.operator.canKick).toBe(true);
		expect(ROLE_PERMISSIONS.operator.canChangeRole).toBe(true);
		expect(ROLE_PERMISSIONS.operator.canCloseRoom).toBe(true);
	});
});
