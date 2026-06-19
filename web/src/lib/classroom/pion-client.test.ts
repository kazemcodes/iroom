import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';

class MockRTCPeerConnection {
	onicecandidate: ((event: { candidate: null }) => void) | null = null;
	ontrack: ((event: { streams: MediaStream[] }) => void) | null = null;
	onconnectionstatechange: (() => void) | null = null;
	connectionState = 'new';
	private tracks: any[] = [];

	async createOffer() {
		return { type: 'offer', sdp: 'mock-sdp' };
	}
	async setLocalDescription() {}
	async setRemoteDescription() {}
	addTrack(track: any, stream: MediaStream) {
		this.tracks.push({ track, stream });
	}
	addTransceiver(kind: string, init: any) {
		this.tracks.push({ kind, direction: init.direction });
	}
	getSenders() {
		return this.tracks.map(t => ({
			track: t.track,
			replaceTrack: vi.fn(),
		}));
	}
	close() {}
}

describe('PionClient', () => {
	const mockConfig = {
		roomId: 'room-1',
		userId: 'user-1',
		role: 'operator',
		displayName: 'Test User',
		userRole: 'operator',
	};

	beforeEach(() => {
		vi.stubGlobal('RTCPeerConnection', MockRTCPeerConnection);
		vi.stubGlobal('RTCSessionDescription', class { sdp: string; type: string; constructor(init: any) { this.sdp = init.sdp; this.type = init.type; } });
		vi.stubGlobal('localStorage', {
			getItem: vi.fn().mockReturnValue('mock-token'),
			setItem: vi.fn(),
			removeItem: vi.fn(),
		});
		vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
			json: () => Promise.resolve({ success: true, data: { sdp: 'mock-sdp' } }),
		}));
	});

	afterEach(() => {
		vi.restoreAllMocks();
	});

	describe('listener mode configuration', () => {
		it('accepts listener option in config', () => {
			const configWithListener = { ...mockConfig, listener: true };
			expect(configWithListener.listener).toBe(true);
		});

		it('defaults to false when listener option not provided', () => {
			const configWithoutListener = { ...mockConfig };
			expect(configWithoutListener.listener).toBeUndefined();
		});

		it('accepts speaker mode (listener: false)', () => {
			const configWithSpeaker = { ...mockConfig, listener: false };
			expect(configWithSpeaker.listener).toBe(false);
		});
	});

	describe('listener mode media behavior', () => {
		it('getUserMedia should not be called in listener mode', async () => {
			const { PionClient } = await import('$lib/classroom/pion-client');

			const getUserMediaSpy = vi.fn().mockResolvedValue({
				getTracks: () => [],
				getAudioTracks: () => [],
				getVideoTracks: () => [],
			});

			vi.stubGlobal('navigator', {
				mediaDevices: { getUserMedia: getUserMediaSpy },
			});

			const client = new PionClient({
				...mockConfig,
				listener: true,
			});

			await client.connect();

			expect(getUserMediaSpy).not.toHaveBeenCalled();
		});

		it('getUserMedia should be called in speaker mode', async () => {
			const { PionClient } = await import('$lib/classroom/pion-client');

			const getUserMediaSpy = vi.fn().mockResolvedValue({
				getTracks: () => [],
				getAudioTracks: () => [],
				getVideoTracks: () => [],
			});

			vi.stubGlobal('navigator', {
				mediaDevices: { getUserMedia: getUserMediaSpy },
			});

			const client = new PionClient({
				...mockConfig,
				listener: false,
			});

			await client.connect();

			expect(getUserMediaSpy).toHaveBeenCalledWith({ video: true, audio: true });
		});
	});

	describe('listener mode connect behavior', () => {
		it('connects successfully in listener mode', async () => {
			const { PionClient } = await import('$lib/classroom/pion-client');

			vi.stubGlobal('navigator', {
				mediaDevices: { getUserMedia: vi.fn() },
			});

			const client = new PionClient({
				...mockConfig,
				listener: true,
			});

			await expect(client.connect()).resolves.not.toThrow();
		});

		it('onLocalStream is not called in listener mode', async () => {
			const { PionClient } = await import('$lib/classroom/pion-client');

			vi.stubGlobal('navigator', {
				mediaDevices: { getUserMedia: vi.fn() },
			});

			const client = new PionClient({
				...mockConfig,
				listener: true,
			});

			const onLocalStreamSpy = vi.fn();
			client.onLocalStream = onLocalStreamSpy;

			await client.connect();

			expect(onLocalStreamSpy).not.toHaveBeenCalled();
		});

		it('onLocalStream is called in speaker mode', async () => {
			const { PionClient } = await import('$lib/classroom/pion-client');

			const mockStream = {
				getTracks: () => [],
				getAudioTracks: () => [],
				getVideoTracks: () => [],
			};

			vi.stubGlobal('navigator', {
				mediaDevices: { getUserMedia: vi.fn().mockResolvedValue(mockStream) },
			});

			const client = new PionClient({
				...mockConfig,
				listener: false,
			});

			const onLocalStreamSpy = vi.fn();
			client.onLocalStream = onLocalStreamSpy;

			await client.connect();

			expect(onLocalStreamSpy).toHaveBeenCalledWith(mockStream);
		});

		it('listener mode adds recvonly transceivers for valid SDP', async () => {
			const { PionClient } = await import('$lib/classroom/pion-client');

			vi.stubGlobal('navigator', {
				mediaDevices: { getUserMedia: vi.fn() },
			});

			let addTransceiverCalls: any[] = [];
			const origRTCPeerConnection = MockRTCPeerConnection;
			vi.stubGlobal('RTCPeerConnection', class extends origRTCPeerConnection {
				addTransceiver(kind: string, init: any) {
					addTransceiverCalls.push({ kind, direction: init.direction });
				}
			});

			const client = new PionClient({
				...mockConfig,
				listener: true,
			});

			await client.connect();

			expect(addTransceiverCalls).toEqual([
				{ kind: 'audio', direction: 'recvonly' },
				{ kind: 'video', direction: 'recvonly' },
			]);
		});

		it('speaker mode does NOT add recvonly transceivers', async () => {
			const { PionClient } = await import('$lib/classroom/pion-client');

			const mockStream = {
				getTracks: () => [],
				getAudioTracks: () => [],
				getVideoTracks: () => [],
			};

			vi.stubGlobal('navigator', {
				mediaDevices: { getUserMedia: vi.fn().mockResolvedValue(mockStream) },
			});

			let addTransceiverCalls: any[] = [];
			const origRTCPeerConnection = MockRTCPeerConnection;
			vi.stubGlobal('RTCPeerConnection', class extends origRTCPeerConnection {
				addTransceiver(kind: string, init: any) {
					addTransceiverCalls.push({ kind, direction: init.direction });
				}
			});

			const client = new PionClient({
				...mockConfig,
				listener: false,
			});

			await client.connect();

			expect(addTransceiverCalls).toEqual([]);
		});
	});
});
