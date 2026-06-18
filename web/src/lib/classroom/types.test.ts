import { describe, it, expect } from 'vitest';
import { ROLE_PERMISSIONS, ROLE_HIERARCHY, ROLE_LABELS, canMuteUser, canKickUser } from '$lib/classroom/types';

describe('ROLE_HIERARCHY', () => {
	it('owner has highest rank (0)', () => {
		expect(ROLE_HIERARCHY.owner).toBe(0);
	});

	it('student/user have lowest rank', () => {
		expect(ROLE_HIERARCHY.student).toBeGreaterThan(ROLE_HIERARCHY.teacher);
		expect(ROLE_HIERARCHY.user).toBeGreaterThan(ROLE_HIERARCHY.presenter);
	});

	it('operator is between admin and teacher', () => {
		expect(ROLE_HIERARCHY.operator).toBeGreaterThan(ROLE_HIERARCHY.admin);
		expect(ROLE_HIERARCHY.operator).toBeLessThan(ROLE_HIERARCHY.teacher);
	});
});

describe('ROLE_LABELS', () => {
	it('has labels for all roles', () => {
		expect(ROLE_LABELS.owner).toBeDefined();
		expect(ROLE_LABELS.admin).toBeDefined();
		expect(ROLE_LABELS.operator).toBeDefined();
		expect(ROLE_LABELS.teacher).toBeDefined();
		expect(ROLE_LABELS.presenter).toBeDefined();
		expect(ROLE_LABELS.user).toBeDefined();
		expect(ROLE_LABELS.student).toBeDefined();
	});

	it('labels are in Persian', () => {
		expect(ROLE_LABELS.owner).toBe('مالک');
		expect(ROLE_LABELS.admin).toBe('مدیر');
		expect(ROLE_LABELS.operator).toBe('اپراتور');
		expect(ROLE_LABELS.teacher).toBe('مدرس');
	});
});

describe('canKickUser', () => {
	it('owner can kick anyone', () => {
		expect(canKickUser('owner', 'admin')).toBe(true);
		expect(canKickUser('owner', 'student')).toBe(true);
	});

	it('admin can kick lower roles', () => {
		expect(canKickUser('admin', 'student')).toBe(true);
		expect(canKickUser('admin', 'teacher')).toBe(true);
	});

	it('teacher cannot kick', () => {
		expect(canKickUser('teacher', 'student')).toBe(false);
	});
});

describe('ROLE_PERMISSIONS completeness', () => {
	it('all roles have permission entries', () => {
		const roles = ['owner', 'admin', 'operator', 'teacher', 'presenter', 'user', 'student'] as const;
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
});
