import { describe, it, expect } from 'vitest';
import { ROLE_PERMISSIONS, ROLE_HIERARCHY, ROLE_LABELS, canMuteUser, canKickUser, type Participant } from '$lib/classroom/types';

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

		it('user has limited permissions', () => {
			const perms = ROLE_PERMISSIONS.user;
			expect(perms.canMic).toBe(false);
			expect(perms.canWebcam).toBe(false);
			expect(perms.canScreenShare).toBe(false);
			expect(perms.canWhiteboard).toBe(false);
			expect(perms.canHandRaise).toBe(true);
			expect(perms.canChat).toBe(true);
			expect(perms.canWatch).toBe(true);
			expect(perms.canKick).toBe(false);
			expect(perms.canChangeRole).toBe(false);
		});

		it('presenter has media permissions but not admin', () => {
			const perms = ROLE_PERMISSIONS.presenter;
			expect(perms.canMic).toBe(true);
			expect(perms.canWebcam).toBe(true);
			expect(perms.canScreenShare).toBe(true);
			expect(perms.canWhiteboard).toBe(true);
			expect(perms.canKick).toBe(false);
			expect(perms.canChangeRole).toBe(false);
		});

		it('operator has full permissions', () => {
			const perms = ROLE_PERMISSIONS.operator;
			expect(perms.canKick).toBe(true);
			expect(perms.canChangeRole).toBe(true);
			expect(perms.canCloseRoom).toBe(true);
		});
	});

	describe('canMuteUser', () => {
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

	describe('canKickUser', () => {
		it('owner can kick anyone', () => {
			expect(canKickUser('owner', 'admin')).toBe(true);
			expect(canKickUser('owner', 'user')).toBe(true);
		});

		it('admin can kick lower roles', () => {
			expect(canKickUser('admin', 'user')).toBe(true);
			expect(canKickUser('admin', 'presenter')).toBe(true);
		});

		it('admin cannot kick other admins', () => {
			expect(canKickUser('admin', 'admin')).toBe(false);
		});

		it('presenter cannot kick', () => {
			expect(canKickUser('presenter', 'user')).toBe(false);
		});
	});

	describe('Participant media state', () => {
		it('participant has media properties', () => {
			const p: Participant = {
				id: '1',
				name: 'Test User',
				role: 'user',
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
			expect(ROLE_HIERARCHY.admin).toBeLessThan(ROLE_HIERARCHY.operator);
		});

		it('user has lowest rank', () => {
			expect(ROLE_HIERARCHY.user).toBeGreaterThan(ROLE_HIERARCHY.presenter);
		});
	});
});
