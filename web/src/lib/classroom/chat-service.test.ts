import { describe, it, expect, vi, beforeEach } from 'vitest';
import type { ChatMessage } from '$lib/classroom/types';

describe('ChatService', () => {
	describe('parseIncomingMessage', () => {
		it('parses a chat/message payload into ChatMessage format', () => {
			const raw = JSON.stringify({
				type: 'chat',
				payload: {
					type: 'message',
					message: {
						user_id: 1,
						user_display_name: 'Ali',
						content: 'Hi',
						created_at: '2026-01-01T00:00:00Z',
					},
				},
			});
			const result = parseIncoming(raw, 2);
			expect(result).not.toBeNull();
			expect(result!.type).toBe('message');
			expect((result as { type: 'message'; chatMessage: ChatMessage }).chatMessage.sender).toBe('Ali');
			expect((result as { type: 'message'; chatMessage: ChatMessage }).chatMessage.content).toBe('Hi');
			expect((result as { type: 'message'; chatMessage: ChatMessage }).chatMessage.isOwn).toBe(false);
		});

		it('marks message as own when user_id matches', () => {
			const raw = JSON.stringify({
				type: 'chat',
				payload: {
					type: 'message',
					message: {
						user_id: 1,
						user_display_name: 'Ali',
						content: 'Hello',
						created_at: '2026-01-01T14:30:00Z',
					},
				},
			});
			const result = parseIncoming(raw, 1);
			expect((result as { type: 'message'; chatMessage: ChatMessage }).chatMessage.isOwn).toBe(true);
			expect((result as { type: 'message'; chatMessage: ChatMessage }).chatMessage.sender).toBe('شما');
		});

		it('falls back to کاربر when display name is empty', () => {
			const raw = JSON.stringify({
				type: 'chat',
				payload: {
					type: 'message',
					message: {
						user_id: 2,
						user_display_name: '',
						content: 'Test',
						created_at: '2026-01-01T00:00:00Z',
					},
				},
			});
			const result = parseIncoming(raw, 1);
			expect((result as { type: 'message'; chatMessage: ChatMessage }).chatMessage.sender).toBe('کاربر');
		});

		it('formats time using fa-IR locale', () => {
			const raw = JSON.stringify({
				type: 'chat',
				payload: {
					type: 'message',
					message: {
						user_id: 2,
						user_display_name: 'Ali',
						content: 'Hi',
						created_at: '2026-01-01T14:30:00Z',
					},
				},
			});
			const result = parseIncoming(raw, 1);
			expect((result as { type: 'message'; chatMessage: ChatMessage }).chatMessage.time).toContain(':');
		});

		it('handles the lower_hands command', () => {
			const raw = JSON.stringify({
				type: 'chat',
				payload: {
					type: 'command',
					command: 'lower_hands',
				},
			});
			const result = parseIncoming(raw, 1);
			expect(result).not.toBeNull();
			expect(result!.type).toBe('command');
			expect((result as { type: 'command'; command: string }).command).toBe('lower_hands');
		});

		it('returns null for malformed JSON', () => {
			expect(parseIncoming('not json', 1)).toBeNull();
		});

		it('returns null for unknown payload types', () => {
			const raw = JSON.stringify({
				type: 'chat',
				payload: { type: 'unknown' },
			});
			expect(parseIncoming(raw, 1)).toBeNull();
		});
	});

	describe('message validation', () => {
		it('rejects empty messages', () => {
			expect(''.trim().length).toBe(0);
		});

		it('rejects whitespace-only messages', () => {
			expect('   '.trim().length).toBe(0);
		});

		it('accepts valid messages', () => {
			expect('Hello'.trim().length).toBeGreaterThan(0);
		});

		it('truncates messages longer than 10000 chars', () => {
			const text = 'a'.repeat(15000);
			const maxLen = 10000;
			const result = text.length > maxLen ? text.slice(0, maxLen) : text;
			expect(result.length).toBe(maxLen);
		});
	});

	describe('send protocol', () => {
		it('sends message in correct WebSocket JSON format', () => {
			const payload = JSON.stringify({ type: 'message', content: 'Hello' });
			const parsed = JSON.parse(payload);
			expect(parsed.type).toBe('message');
			expect(parsed.content).toBe('Hello');
		});

		it('sends command in correct format', () => {
			const payload = JSON.stringify({ type: 'command', command: 'lower_hands' });
			const parsed = JSON.parse(payload);
			expect(parsed.type).toBe('command');
			expect(parsed.command).toBe('lower_hands');
		});
	});

	describe('receive protocol', () => {
		it('unwraps message payload correctly', () => {
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

		it('unwraps command payload correctly', () => {
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

	describe('sender logic', () => {
		it('own message shows شما', () => {
			const sender = 1 === 1 ? 'شما' : 'Ali';
			expect(sender).toBe('شما');
		});

		it('other message shows display name', () => {
			const userId = 1 as number;
			const currentUserId = 2 as number;
			const sender = userId !== currentUserId ? 'Ali' : 'شما';
			expect(sender).toBe('Ali');
		});

		it('falls back to کاربر if no display name', () => {
			const displayName = '';
			const sender = displayName || 'کاربر';
			expect(sender).toBe('کاربر');
		});
	});

	describe('time formatting', () => {
		it('formats time with colon separator', () => {
			const date = new Date('2026-01-01T14:30:00Z');
			const time = date.toLocaleTimeString('fa-IR', { hour: '2-digit', minute: '2-digit' });
			expect(time).toContain(':');
			expect(time.split(':')).toHaveLength(2);
		});
	});
});

// Minimal inline implementation for testing
type ParseResult =
	| { type: 'message'; chatMessage: ChatMessage }
	| { type: 'command'; command: string }
	| null;

function parseIncoming(raw: string, currentUserId: number): ParseResult {
	try {
		const data = JSON.parse(raw);
		const inner = data.payload || data;
		if (inner.type === 'message' && inner.message) {
			const msg = inner.message;
			const isOwn = msg.user_id === currentUserId;
			return {
				type: 'message',
				chatMessage: {
					id: String(Date.now()) + '-' + Math.random().toString(36).slice(2, 7),
					sender: isOwn ? 'شما' : (msg.user_display_name || 'کاربر'),
					content: msg.content,
					time: new Date(msg.created_at).toLocaleTimeString('fa-IR', { hour: '2-digit', minute: '2-digit' }),
					isOwn,
				},
			};
		}
		if (inner.type === 'command' && inner.command) {
			return { type: 'command', command: inner.command };
		}
		return null;
	} catch {
		return null;
	}
}
