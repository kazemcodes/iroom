import { describe, it, expect, vi, beforeEach } from 'vitest';

// --- Global Media mocks for jsdom ---
// These must be available globally before any imports, since MediaClient
// references MediaStream, MediaRecorder, etc. at the top level.

class MockMediaRecorder {
	state = 'inactive';
	ondataavailable: ((e: { data: Blob }) => void) | null = null;
	onerror: ((e: Event) => void) | null = null;
	onstart: (() => void) | null = null;
	onstop: (() => void) | null = null;
	stream: MediaStream;

	constructor(stream: MediaStream, _options?: MediaRecorderOptions) {
		this.stream = stream;
	}

	start = vi.fn(function(this: MockMediaRecorder, _timeslice?: number) {
		this.state = 'recording';
		if (this.onstart) this.onstart();
	});

	stop = vi.fn(function(this: MockMediaRecorder) {
		this.state = 'inactive';
		if (this.onstop) this.onstop();
	});

	pause = vi.fn();
	resume = vi.fn();
	requestData = vi.fn();

	static isTypeSupported(_mimeType: string): boolean {
		return true;
	}
}

function createMockTrack(kind: string, enabled = true): MediaStreamTrack {
	return {
		kind,
		enabled,
		stop: vi.fn(),
		id: `${kind}-${Math.random()}`,
		label: kind,
		muted: false,
		readyState: 'live',
		onended: null,
		onmute: null,
		onunmute: null,
		contentHint: '',
		applyConstraints: vi.fn(),
		clone: vi.fn(),
		getCapabilities: vi.fn(),
		getConstraints: vi.fn(),
		getSettings: vi.fn(),
		dispatchEvent: vi.fn(),
		addEventListener: vi.fn(),
		removeEventListener: vi.fn(),
	} as unknown as MediaStreamTrack;
}

function createMockStream(tracks: MediaStreamTrack[]): MediaStream {
	return {
		getTracks: vi.fn(() => [...tracks]),
		getVideoTracks: vi.fn(() => tracks.filter(t => t.kind === 'video')),
		getAudioTracks: vi.fn(() => tracks.filter(t => t.kind === 'audio')),
		addTrack: vi.fn(),
		removeTrack: vi.fn(),
		id: 'mock-stream',
		active: true,
		addEventListener: vi.fn(),
		removeEventListener: vi.fn(),
		dispatchEvent: vi.fn(),
		onaddtrack: null,
		onremovetrack: null,
		clone: vi.fn(),
		getTrackById: vi.fn(),
	} as unknown as MediaStream;
}

// Stub globals before any imports
vi.stubGlobal('MediaStream', class {
	private _tracks: MediaStreamTrack[] = [];
	getTracks() { return this._tracks; }
	getVideoTracks() { return this._tracks.filter((t: MediaStreamTrack) => t.kind === 'video'); }
	getAudioTracks() { return this._tracks.filter((t: MediaStreamTrack) => t.kind === 'audio'); }
	addTrack(t: MediaStreamTrack) { this._tracks.push(t); }
	removeTrack(t: MediaStreamTrack) { this._tracks = this._tracks.filter((tr: MediaStreamTrack) => tr !== t); }
	id = 'global-mock-stream';
	active = true;
	addEventListener() {}
	removeEventListener() {}
	dispatchEvent() {}
	clone() { return this; }
	getTrackById() { return null; }
} as any);

vi.stubGlobal('MediaRecorder', MockMediaRecorder as any);

// Mock WebSocket
class MockWebSocket {
	readyState = WebSocket.OPEN;
	send = vi.fn();
	addEventListener = vi.fn();
	removeEventListener = vi.fn();
	close = vi.fn();
}

function createVideoTrackMock(): MediaStreamTrack {
	const track = createMockTrack('video');
	// Allow onended to be set and called
	let _onended: ((this: MediaStreamTrack, ev: Event) => any) | null = null;
	Object.defineProperty(track, 'onended', {
		get() { return _onended; },
		set(fn) { _onended = fn; },
		configurable: true,
	});
	return track;
}

function createAudioTrackMock(): MediaStreamTrack {
	const track = createMockTrack('audio');
	// Allow onended to be set and called
	let _onended: ((this: MediaStreamTrack, ev: Event) => any) | null = null;
	Object.defineProperty(track, 'onended', {
		get() { return _onended; },
		set(fn) { _onended = fn; },
		configurable: true,
	});
	return track;
}

describe('MediaClient webcam', () => {
	beforeEach(() => {
		vi.clearAllMocks();
	});

	it('calls onWebcamEnded when browser revokes camera (video track.onended)', async () => {
		const { MediaClient } = await import('$lib/classroom/media-client');

		const ws = new MockWebSocket() as unknown as WebSocket;
		const client = new MediaClient(ws, 42);

		const videoTrack = createVideoTrackMock();
		const audioTrack = createMockTrack('audio');
		const localStream = createMockStream([videoTrack, audioTrack]);

		vi.stubGlobal('navigator', {
			...navigator,
			mediaDevices: {
				...navigator.mediaDevices,
				getUserMedia: vi.fn(async () => localStream),
				getDisplayMedia: vi.fn(),
			},
		});

		vi.stubGlobal('MediaRecorder', MockMediaRecorder as any);

		const onWebcamEnded = vi.fn();
		client.onWebcamEnded = onWebcamEnded;

		// Start with video + audio
		await client.start(true, true);

		// Verify the video track has an onended handler set
		expect(videoTrack.onended).toBeInstanceOf(Function);

		// Simulate browser revoking camera permission
		const endedHandler = videoTrack.onended! as () => void;
		await endedHandler();

		// Verify the callback was called
		expect(onWebcamEnded).toHaveBeenCalledTimes(1);
		// Verify the track was stopped
		expect(videoTrack.stop).toHaveBeenCalled();
	});

	it('onWebcamEnded not set when starting without video', async () => {
		const { MediaClient } = await import('$lib/classroom/media-client');

		const ws = new MockWebSocket() as unknown as WebSocket;
		const client = new MediaClient(ws, 42);

		const audioTrack = createMockTrack('audio');
		const localStream = createMockStream([audioTrack]);

		vi.stubGlobal('navigator', {
			...navigator,
			mediaDevices: {
				...navigator.mediaDevices,
				getUserMedia: vi.fn(async () => localStream),
				getDisplayMedia: vi.fn(),
			},
		});

		const onWebcamEnded = vi.fn();
		client.onWebcamEnded = onWebcamEnded;

		// Start with audio only (no video)
		await client.start(false, true);

		// Callback should never fire since there's no video track
		expect(onWebcamEnded).not.toHaveBeenCalled();
	});

	it('does not call onWebcamEnded when toggleVideo stops webcam programmatically', async () => {
		const { MediaClient } = await import('$lib/classroom/media-client');

		const ws = new MockWebSocket() as unknown as WebSocket;
		const client = new MediaClient(ws, 42);

		const videoTrack = createVideoTrackMock();
		const audioTrack = createMockTrack('audio');
		const localStream = createMockStream([videoTrack, audioTrack]);

		vi.stubGlobal('navigator', {
			...navigator,
			mediaDevices: {
				...navigator.mediaDevices,
				getUserMedia: vi.fn(async () => localStream),
				getDisplayMedia: vi.fn(),
			},
		});

		const onWebcamEnded = vi.fn();
		client.onWebcamEnded = onWebcamEnded;

		await client.start(true, true);

		// First stub getUserMedia for toggle ON to return a new video track
		const newVideoTrack = createVideoTrackMock();
		const newStream = createMockStream([newVideoTrack]);
		vi.stubGlobal('navigator', {
			...navigator,
			mediaDevices: {
				...navigator.mediaDevices,
				getUserMedia: vi.fn(async () => newStream),
				getDisplayMedia: vi.fn(),
			},
		});

		// Toggle OFF
		await client.toggleVideo();

		// Programmatic stop should NOT fire onWebcamEnded
		expect(onWebcamEnded).not.toHaveBeenCalled();
	});
});

describe('MediaClient screen share', () => {
	beforeEach(() => {
		vi.clearAllMocks();
	});

	it('calls onScreenShareEnded when browser stops screen share (track.onended)', async () => {
		const { MediaClient } = await import('$lib/classroom/media-client');

		const ws = new MockWebSocket() as unknown as WebSocket;
		const client = new MediaClient(ws, 42);

		// Set up a simple local stream (needed for shareScreen)
		const localTracks = [createMockTrack('audio')];
		const localStream = createMockStream(localTracks);
		(client as any).localStream = localStream;

		// Track the onended value set by shareScreen
		const screenTrack = createMockTrack('video');
		let onendedValue: any = null;
		const onendedSpy = vi.fn();

		const trackProxy = new Proxy(screenTrack, {
			set(target, prop, value) {
				if (prop === 'onended') {
					onendedValue = value;
					onendedSpy(value);
				}
				return Reflect.set(target, prop, value);
			},
		});

		vi.stubGlobal('navigator', {
			...navigator,
			mediaDevices: {
				...navigator.mediaDevices,
				getDisplayMedia: vi.fn(async () =>
					createMockStream([trackProxy as unknown as MediaStreamTrack])
				),
				getUserMedia: vi.fn(),
			},
		});

		// Spy on stopScreenShare
		const stopScreenShareSpy = vi.spyOn(client, 'stopScreenShare');

		// Set up the callback
		const onScreenShareEnded = vi.fn();
		client.onScreenShareEnded = onScreenShareEnded;

		// Start screen share
		await client.shareScreen();

		// Verify onended was set
		expect(onendedSpy).toHaveBeenCalled();
		expect(onendedValue).toBeInstanceOf(Function);

		// Simulate the browser stopping the screen share
		await onendedValue!();

		// Verify stopScreenShare was called
		expect(stopScreenShareSpy).toHaveBeenCalled();

		// Verify onScreenShareEnded callback was called
		expect(onScreenShareEnded).toHaveBeenCalledTimes(1);
	});

	it('stopScreenShare cleans up screen stream and recorder', async () => {
		const { MediaClient } = await import('$lib/classroom/media-client');

		const ws = new MockWebSocket() as unknown as WebSocket;
		const client = new MediaClient(ws, 42);

		const localTracks = [createMockTrack('audio')];
		const localStream = createMockStream(localTracks);
		(client as any).localStream = localStream;

		const screenTrack = createMockTrack('video');
		const screenStream = createMockStream([screenTrack]);
		(client as any).screenStream = screenStream;

		const screenRecorder = new MockMediaRecorder(screenStream as unknown as MediaStream);
		screenRecorder.state = 'recording';
		(client as any).screenRecorder = screenRecorder;

		const onScreenStream = vi.fn();
		client.onScreenStream = onScreenStream;

		client.stopScreenShare();

		expect(screenTrack.stop).toHaveBeenCalled();
		expect((client as any).screenStream).toBeNull();
		expect(onScreenStream).toHaveBeenCalledWith(null);
		expect((client as any).screenRecorder).toBeNull();
	});

	it('does not call onScreenShareEnded when stopScreenShare is called directly (programmatic stop)', async () => {
		const { MediaClient } = await import('$lib/classroom/media-client');

		const ws = new MockWebSocket() as unknown as WebSocket;
		const client = new MediaClient(ws, 42);

		const localTracks = [createMockTrack('audio')];
		const localStream = createMockStream(localTracks);
		(client as any).localStream = localStream;

		const screenTrack = createMockTrack('video');
		const screenStream = createMockStream([screenTrack]);
		(client as any).screenStream = screenStream;

		const onScreenShareEnded = vi.fn();
		client.onScreenShareEnded = onScreenShareEnded;

		// Calling stopScreenShare directly should NOT fire the callback
		// (the callback is only for browser-initiated stops)
		client.stopScreenShare();

		expect(onScreenShareEnded).not.toHaveBeenCalled();
	});
});

describe('MediaClient audio', () => {
	beforeEach(() => {
		vi.clearAllMocks();
	});

	it('calls onMicEnded when browser revokes microphone (audio track.onended)', async () => {
		const { MediaClient } = await import('$lib/classroom/media-client');

		const ws = new MockWebSocket() as unknown as WebSocket;
		const client = new MediaClient(ws, 42);

		const videoTrack = createVideoTrackMock();
		const audioTrack = createAudioTrackMock();
		const localStream = createMockStream([videoTrack, audioTrack]);

		vi.stubGlobal('navigator', {
			...navigator,
			mediaDevices: {
				...navigator.mediaDevices,
				getUserMedia: vi.fn(async () => localStream),
				getDisplayMedia: vi.fn(),
			},
		});

		vi.stubGlobal('MediaRecorder', MockMediaRecorder as any);

		const onMicEnded = vi.fn();
		client.onMicEnded = onMicEnded;

		// Start with video + audio
		await client.start(true, true);

		// Verify the audio track has an onended handler set
		expect(audioTrack.onended).toBeInstanceOf(Function);

		// Simulate browser revoking microphone permission
		const endedHandler = audioTrack.onended! as () => void;
		await endedHandler();

		// Verify the callback was called
		expect(onMicEnded).toHaveBeenCalledTimes(1);
		// Verify the track was stopped
		expect(audioTrack.stop).toHaveBeenCalled();
	});

	it('onMicEnded not set when starting without audio', async () => {
		const { MediaClient } = await import('$lib/classroom/media-client');

		const ws = new MockWebSocket() as unknown as WebSocket;
		const client = new MediaClient(ws, 42);

		const videoTrack = createVideoTrackMock();
		const localStream = createMockStream([videoTrack]);

		vi.stubGlobal('navigator', {
			...navigator,
			mediaDevices: {
				...navigator.mediaDevices,
				getUserMedia: vi.fn(async () => localStream),
				getDisplayMedia: vi.fn(),
			},
		});

		const onMicEnded = vi.fn();
		client.onMicEnded = onMicEnded;

		// Start with video only (no audio)
		await client.start(true, false);

		// Callback should never fire since there's no audio track
		expect(onMicEnded).not.toHaveBeenCalled();
	});

	it('does not call onMicEnded when toggleAudio stops mic programmatically', async () => {
		const { MediaClient } = await import('$lib/classroom/media-client');

		const ws = new MockWebSocket() as unknown as WebSocket;
		const client = new MediaClient(ws, 42);

		const videoTrack = createVideoTrackMock();
		const audioTrack = createAudioTrackMock();
		const localStream = createMockStream([videoTrack, audioTrack]);

		vi.stubGlobal('navigator', {
			...navigator,
			mediaDevices: {
				...navigator.mediaDevices,
				getUserMedia: vi.fn(async () => localStream),
				getDisplayMedia: vi.fn(),
			},
		});

		const onMicEnded = vi.fn();
		client.onMicEnded = onMicEnded;

		await client.start(true, true);

		// First stub getUserMedia for toggle ON to return a new audio track
		const newAudioTrack = createAudioTrackMock();
		const newStream = createMockStream([newAudioTrack]);
		vi.stubGlobal('navigator', {
			...navigator,
			mediaDevices: {
				...navigator.mediaDevices,
				getUserMedia: vi.fn(async () => newStream),
				getDisplayMedia: vi.fn(),
			},
		});

		// Toggle OFF
		await client.toggleAudio();

		// Programmatic stop should NOT fire onMicEnded
		expect(onMicEnded).not.toHaveBeenCalled();
	});
});
