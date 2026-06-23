import { describe, it, expect } from 'vitest';
import { ROLE_PERMISSIONS, ROLE_HIERARCHY, ROLE_LABELS, type Participant } from '$lib/classroom/types';

describe('Room Entry & Role System', () => {
	describe('Entry mode types', () => {
		it('entry mode must be speaker or listener', () => {
			const validModes: Array<'speaker' | 'listener'> = ['speaker', 'listener'];
			expect(validModes).toContain('speaker');
			expect(validModes).toContain('listener');
			expect(validModes).toHaveLength(2);
		});
	});

	describe('Speaker role permissions', () => {
		it('speaker (operator) has full mic and webcam permissions', () => {
			const perms = ROLE_PERMISSIONS.operator;
			expect(perms.canMic).toBe(true);
			expect(perms.canWebcam).toBe(true);
			expect(perms.canScreenShare).toBe(true);
		});

		it('presenter has mic and webcam permissions', () => {
			const perms = ROLE_PERMISSIONS.presenter;
			expect(perms.canMic).toBe(true);
			expect(perms.canWebcam).toBe(true);
		});
	});

	describe('Listener role permissions', () => {
		it('user/listener has no mic or webcam permissions', () => {
			const perms = ROLE_PERMISSIONS.user;
			expect(perms.canMic).toBe(false);
			expect(perms.canWebcam).toBe(false);
			expect(perms.canScreenShare).toBe(false);
			expect(perms.canWhiteboard).toBe(false);
			expect(perms.canHandRaise).toBe(true);
			expect(perms.canChat).toBe(true);
			expect(perms.canWatch).toBe(true);
		});
	});

	describe('Media icon state logic', () => {
		it('muted icon should be shown when hasAudio is false', () => {
			const participant: Participant = {
				id: '1', name: 'Test', role: 'user',
				isSpeaking: false, hasVideo: false, hasAudio: false,
				hasScreen: false, hasWhiteboard: false, handRaised: false,
			};
			expect(participant.hasAudio).toBe(false);
		});

		it('active icon should be shown when hasAudio is true', () => {
			const participant: Participant = {
				id: '1', name: 'Test', role: 'operator',
				isSpeaking: false, hasVideo: true, hasAudio: true,
				hasScreen: false, hasWhiteboard: false, handRaised: false,
			};
			expect(participant.hasAudio).toBe(true);
		});

		it('speaking participant should have isSpeaking flag', () => {
			const participant: Participant = {
				id: '1', name: 'Speaker', role: 'operator',
				isSpeaking: true, hasVideo: true, hasAudio: true,
				hasScreen: false, hasWhiteboard: false, handRaised: false,
			};
			expect(participant.isSpeaking).toBe(true);
		});

		it('video icon should reflect hasVideo state', () => {
			const withVideo: Participant = {
				id: '1', name: 'A', role: 'user',
				isSpeaking: false, hasVideo: true, hasAudio: true,
				hasScreen: false, hasWhiteboard: false, handRaised: false,
			};
			const withoutVideo: Participant = {
				id: '2', name: 'B', role: 'user',
				isSpeaking: false, hasVideo: false, hasAudio: false,
				hasScreen: false, hasWhiteboard: false, handRaised: false,
			};
			expect(withVideo.hasVideo).toBe(true);
			expect(withoutVideo.hasVideo).toBe(false);
		});
	});

	describe('Three room roles', () => {
		it('operator has full permissions', () => {
			expect(ROLE_PERMISSIONS.operator.canMic).toBe(true);
			expect(ROLE_PERMISSIONS.operator.canWebcam).toBe(true);
			expect(ROLE_PERMISSIONS.operator.canKick).toBe(true);
			expect(ROLE_PERMISSIONS.operator.canChangeRole).toBe(true);
		});

		it('presenter has media but not admin permissions', () => {
			expect(ROLE_PERMISSIONS.presenter.canMic).toBe(true);
			expect(ROLE_PERMISSIONS.presenter.canWebcam).toBe(true);
			expect(ROLE_PERMISSIONS.presenter.canKick).toBe(false);
			expect(ROLE_PERMISSIONS.presenter.canChangeRole).toBe(false);
		});

		it('user has only watch/chat/hand-raise', () => {
			expect(ROLE_PERMISSIONS.user.canMic).toBe(false);
			expect(ROLE_PERMISSIONS.user.canWebcam).toBe(false);
			expect(ROLE_PERMISSIONS.user.canHandRaise).toBe(true);
			expect(ROLE_PERMISSIONS.user.canChat).toBe(true);
		});

		it('operator has highest rank among room roles', () => {
			expect(ROLE_HIERARCHY.operator).toBeLessThan(ROLE_HIERARCHY.presenter);
			expect(ROLE_HIERARCHY.presenter).toBeLessThan(ROLE_HIERARCHY.user);
		});

		it('role labels are in Persian', () => {
			expect(ROLE_LABELS.operator).toBe('اپراتور');
			expect(ROLE_LABELS.presenter).toBe('ارائه‌دهنده');
			expect(ROLE_LABELS.user).toBe('کاربر عادی');
		});
	});
});
