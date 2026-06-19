import { describe, it, expect, vi, beforeEach } from 'vitest';
import { ROLE_PERMISSIONS, ROLE_HIERARCHY, ROLE_LABELS, canMuteUser, canKickUser, type UserRole, type Participant } from '$lib/classroom/types';

describe('Room Features', () => {
	describe('ROLE_PERMISSIONS', () => {
		it('admin has all permissions', () => {
			const perms = ROLE_PERMISSIONS.admin;
			expect(perms.canMic).toBe(true);
			expect(perms.canWebcam).toBe(true);
			expect(perms.canScreenShare).toBe(true);
			expect(perms.canWhiteboard).toBe(true);
			expect(perms.canHandRaise).toBe(true);
			expect(perms.canChat).toBe(true);
			expect(perms.canWatch).toBe(true);
		});

		it('student has limited permissions', () => {
			const perms = ROLE_PERMISSIONS.student;
			expect(perms.canMic).toBe(false);
			expect(perms.canWebcam).toBe(false);
			expect(perms.canScreenShare).toBe(false);
			expect(perms.canWhiteboard).toBe(false);
			expect(perms.canHandRaise).toBe(true);
			expect(perms.canChat).toBe(true);
			expect(perms.canWatch).toBe(true);
		});

		it('teacher has full permissions', () => {
			const perms = ROLE_PERMISSIONS.teacher;
			expect(perms.canMic).toBe(true);
			expect(perms.canWebcam).toBe(true);
			expect(perms.canScreenShare).toBe(true);
			expect(perms.canWhiteboard).toBe(true);
		});
	});

	describe('canMuteUser', () => {
		it('admin can mute student', () => {
			expect(canMuteUser('admin', 'student')).toBe(true);
		});

		it('admin can mute teacher', () => {
			expect(canMuteUser('admin', 'teacher')).toBe(true);
		});

		it('student cannot mute admin', () => {
			expect(canMuteUser('student', 'admin')).toBe(false);
		});

		it('teacher can mute student', () => {
			expect(canMuteUser('teacher', 'student')).toBe(true);
		});
	});

	describe('canKickUser', () => {
		it('owner can kick anyone', () => {
			expect(canKickUser('owner', 'admin')).toBe(true);
			expect(canKickUser('owner', 'teacher')).toBe(true);
			expect(canKickUser('owner', 'student')).toBe(true);
		});

		it('admin can kick lower roles', () => {
			expect(canKickUser('admin', 'student')).toBe(true);
			expect(canKickUser('admin', 'teacher')).toBe(true);
		});

		it('admin cannot kick other admins', () => {
			expect(canKickUser('admin', 'admin')).toBe(false);
		});

		it('teacher cannot kick', () => {
			expect(canKickUser('teacher', 'student')).toBe(false);
		});
	});

	describe('Participant media state', () => {
		it('participant has media properties', () => {
			const p: Participant = {
				id: '1',
				name: 'Test User',
				role: 'student',
				isSpeaking: false,
				hasVideo: true,
				hasAudio: false,
				hasScreen: false,
				hasWhiteboard: false,
				handRaised: true,
			};
			expect(p.hasVideo).toBe(true);
			expect(p.hasAudio).toBe(false);
			expect(p.handRaised).toBe(true);
		});

		it('participant can be marked as local', () => {
			const p: Participant = {
				id: '1',
				name: 'Local User',
				role: 'admin',
				isSpeaking: false,
				hasVideo: true,
				hasAudio: true,
				hasScreen: false,
				hasWhiteboard: false,
				handRaised: false,
				isLocal: true,
			};
			expect(p.isLocal).toBe(true);
		});
	});

	describe('Role hierarchy', () => {
		it('owner is highest', () => {
			expect(ROLE_HIERARCHY.owner).toBeLessThan(ROLE_HIERARCHY.admin);
			expect(ROLE_HIERARCHY.admin).toBeLessThan(ROLE_HIERARCHY.teacher);
		});

		it('student and user have same rank', () => {
			expect(ROLE_HIERARCHY.student).toBe(ROLE_HIERARCHY.user);
		});
	});
});
