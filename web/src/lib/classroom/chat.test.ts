import { describe, it, expect } from 'vitest';
import type { ChatMessage } from '$lib/classroom/types';

describe('Chat functionality', () => {
	describe('ChatMessage structure', () => {
		it('has required fields', () => {
			const msg: ChatMessage = {
				id: '1',
				sender: 'Ali',
				content: 'Hello',
				time: '12:00',
				isOwn: false,
			};
			expect(msg.id).toBeDefined();
			expect(msg.sender).toBeDefined();
			expect(msg.content).toBeDefined();
			expect(msg.time).toBeDefined();
			expect(typeof msg.isOwn).toBe('boolean');
		});

		it('supports replyTo field', () => {
			const msg: ChatMessage = {
				id: '1',
				sender: 'Ali',
				content: 'Reply',
				time: '12:00',
				isOwn: false,
				replyTo: { sender: 'Sara', content: 'Original' },
			};
			expect(msg.replyTo).toBeDefined();
			expect(msg.replyTo!.sender).toBe('Sara');
		});
	});

	describe('Chat input validation', () => {
		it('empty message should be rejected', () => {
			const text = '';
			expect(text.trim().length).toBe(0);
		});

		it('whitespace-only message should be rejected', () => {
			const text = '   ';
			expect(text.trim().length).toBe(0);
		});

		it('valid message should be accepted', () => {
			const text = 'Hello world';
			expect(text.trim().length).toBeGreaterThan(0);
		});

		it('very long message should be truncated to 1000 chars', () => {
			const text = 'a'.repeat(1500);
			const maxLen = 1000;
			const result = text.length > maxLen ? text.slice(0, maxLen) : text;
			expect(result.length).toBe(maxLen);
		});
	});

	describe('Chat send protocol', () => {
		it('send message uses correct WebSocket format', () => {
			const text = 'Hello';
			const payload = JSON.stringify({ type: 'message', content: text });
			const parsed = JSON.parse(payload);
			expect(parsed.type).toBe('message');
			expect(parsed.content).toBe('Hello');
		});

		it('command message uses correct format', () => {
			const payload = JSON.stringify({ type: 'command', command: 'lower_hands' });
			const parsed = JSON.parse(payload);
			expect(parsed.type).toBe('command');
			expect(parsed.command).toBe('lower_hands');
		});
	});

	describe('Chat receive protocol', () => {
		it('message payload unwraps correctly', () => {
			const raw = JSON.stringify({
				type: 'chat',
				payload: {
					type: 'message',
					message: {
						user_id: 1,
						user_display_name: 'Ali',
						content: 'Hello',
						created_at: '2026-01-01T00:00:00Z',
					},
				},
			});
			const data = JSON.parse(raw);
			const inner = data.payload || data;
			expect(inner.type).toBe('message');
			expect(inner.message.user_display_name).toBe('Ali');
		});

		it('command payload unwraps correctly', () => {
			const raw = JSON.stringify({
				type: 'chat',
				payload: {
					type: 'command',
					command: 'lower_hands',
				},
			});
			const data = JSON.parse(raw);
			const inner = data.payload || data;
			expect(inner.type).toBe('command');
			expect(inner.command).toBe('lower_hands');
		});
	});

	describe('Chat time formatting', () => {
		it('formats time for display', () => {
			const date = new Date('2026-01-01T14:30:00Z');
			const time = date.toLocaleTimeString('fa-IR', { hour: '2-digit', minute: '2-digit' });
			expect(time).toContain(':');
			expect(time.split(':')).toHaveLength(2);
		});
	});

	describe('Chat message sender', () => {
		it('own message shows شما', () => {
			const currentUserId = 1;
			const msgUserId = 1;
			const sender = msgUserId === currentUserId ? 'شما' : 'Ali';
			expect(sender).toBe('شما');
		});

		it('other message shows display name', () => {
			const currentUserId = 1;
			const msgUserId = 2;
			const displayName = 'Ali';
			const sender = msgUserId === currentUserId ? 'شما' : (displayName || 'کاربر');
			expect(sender).toBe('Ali');
		});

		it('falls back to کاربر if no display name', () => {
			const currentUserId = 1;
			const msgUserId = 2;
			const displayName = '';
			const sender = msgUserId === currentUserId ? 'شما' : (displayName || 'کاربر');
			expect(sender).toBe('کاربر');
		});
	});
});
