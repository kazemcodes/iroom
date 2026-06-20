/**
 * MediaClient — WebSocket-based video/audio for classroom.
 *
 * Uses MediaRecorder (VP8+Opus encoding) → WebSocket binary → MediaSource playback.
 * Chrome-only. No WebRTC, no STUN/TURN, no ICE.
 *
 * Binary message format: [userId: uint32 BE][WebM chunk]
 */

const MIME_TYPE = 'video/webm;codecs=vp8,opus';
const VIDEO_BITRATE = 250000;
const AUDIO_BITRATE = 32000;
const CHUNK_INTERVAL_MS = 200;

export class MediaClient {
	private ws: WebSocket;
	private userId: number;
	private localStream: MediaStream | null = null;
	private recorder: MediaRecorder | null = null;
	private videoEnabled = true;
	private audioEnabled = true;

	private remoteEntries = new Map<string, {
		ms: MediaSource;
		video: HTMLVideoElement;
		sb: SourceBuffer | null;
		ready: boolean;
		pendingChunks: ArrayBuffer[];
	}>();

	onLocalStream?: (stream: MediaStream) => void;
	onRemoteStream?: (stream: MediaStream, participantId: string) => void;
	onRemoteStreamEnd?: (participantId: string) => void;

	constructor(ws: WebSocket, userId: number) {
		this.ws = ws;
		this.userId = userId;
	}

	async start(video = true, audio = true): Promise<void> {
		this.localStream = await navigator.mediaDevices.getUserMedia({ video, audio });
		if (this.onLocalStream) this.onLocalStream(this.localStream);

		if (!MediaRecorder.isTypeSupported(MIME_TYPE)) {
			console.error('[Media] MIME type not supported:', MIME_TYPE);
			return;
		}

		this.recorder = new MediaRecorder(this.localStream, {
			mimeType: MIME_TYPE,
			videoBitsPerSecond: VIDEO_BITRATE,
		});

		this.recorder.ondataavailable = async (e) => {
			if (e.data.size > 0 && this.ws.readyState === WebSocket.OPEN) {
				const chunk = await e.data.arrayBuffer();
				const header = new ArrayBuffer(4);
				new DataView(header).setUint32(0, this.userId, false);
				const msg = new Uint8Array(4 + chunk.byteLength);
				msg.set(new Uint8Array(header), 0);
				msg.set(new Uint8Array(chunk), 4);
				this.ws.send(msg.buffer);
			}
		};

		this.recorder.onerror = (e) => {
			console.error('[Media] Recorder error:', e);
		};

		this.recorder.onstop = () => {
			console.log('[Media] Recorder stopped');
		};

		this.recorder.start(CHUNK_INTERVAL_MS);
		console.log('[Media] Started recording', { video, audio, mimeType: MIME_TYPE });
	}

	handleBinaryMessage(data: ArrayBuffer): void {
		if (data.byteLength < 4) return;
		const senderId = new DataView(data).getUint32(0, false).toString();
		if (senderId === this.userId.toString()) return;

		const chunk = data.slice(4);

		let entry = this.remoteEntries.get(senderId);
		if (!entry) {
			entry = this.createRemoteEntry(senderId);
			this.remoteEntries.set(senderId, entry);
		}

		if (entry.ready && entry.sb && !entry.sb.updating) {
			try {
				entry.sb.appendBuffer(new Uint8Array(chunk));
			} catch (e) {
				console.warn('[Media] appendBuffer error:', e);
			}
		} else {
			entry.pendingChunks.push(chunk);
		}
	}

	private createRemoteEntry(senderId: string): {
		ms: MediaSource;
		video: HTMLVideoElement;
		sb: SourceBuffer | null;
		ready: boolean;
		pendingChunks: ArrayBuffer[];
	} {
		const ms = new MediaSource();
		const video = document.createElement('video');
		video.autoplay = true;
		video.playsInline = true;
		video.muted = true;
		video.style.display = 'none';
		document.body.appendChild(video);
		video.src = URL.createObjectURL(ms);

		const entry = { ms, video, sb: null as SourceBuffer | null, ready: false, pendingChunks: [] as ArrayBuffer[] };

		ms.addEventListener('sourceopen', () => {
			try {
				entry.sb = ms.addSourceBuffer(MIME_TYPE);
				entry.sb.mode = 'sequence';
				entry.sb.onupdateend = () => {
					this.drainQueue(entry);
				};
				entry.ready = true;
				this.drainQueue(entry);
			} catch (e) {
				console.error('[Media] SourceBuffer error:', e);
			}
		});

		ms.addEventListener('error', (e) => {
			console.error('[Media] MediaSource error:', e);
		});

		video.onloadeddata = () => {
			try {
				const stream = video.captureStream ? (video as any).captureStream() : (video as any).mozCaptureStream();
				if (stream && this.onRemoteStream) {
					this.onRemoteStream(stream, senderId);
				}
			} catch (e) {
				console.warn('[Media] captureStream failed:', e);
			}
		};

		return entry;
	}

	private drainQueue(entry: { sb: SourceBuffer | null; pendingChunks: ArrayBuffer[]; ready: boolean }): void {
		if (!entry.ready || !entry.sb || entry.sb.updating) return;
		while (entry.pendingChunks.length > 0) {
			if (entry.sb.updating) break;
			const chunk = entry.pendingChunks.shift()!;
			try {
				entry.sb.appendBuffer(new Uint8Array(chunk));
			} catch (e) {
				console.warn('[Media] appendBuffer error in drain:', e);
				break;
			}
		}
	}

	toggleVideo(): void {
		if (!this.localStream) return;
		this.videoEnabled = !this.videoEnabled;
		this.localStream.getVideoTracks().forEach(t => { t.enabled = this.videoEnabled; });
	}

	toggleAudio(): void {
		if (!this.localStream) return;
		this.audioEnabled = !this.audioEnabled;
		this.localStream.getAudioTracks().forEach(t => { t.enabled = this.audioEnabled; });
	}

	async shareScreen(): Promise<void> {
		const screenStream = await navigator.mediaDevices.getDisplayMedia({ video: true });
		const screenTrack = screenStream.getVideoTracks()[0];

		if (!this.localStream || !this.recorder) return;

		const composedStream = new MediaStream([
			screenTrack,
			...this.localStream.getAudioTracks()
		]);

		screenTrack.onended = () => {
			this.stopScreenShare();
		};

		this.recorder.stop();
		this.startRecorder(composedStream, VIDEO_BITRATE * 2);
	}

	stopScreenShare(): void {
		if (!this.localStream || !this.recorder) return;

		this.recorder.stop();
		this.startRecorder(this.localStream, VIDEO_BITRATE);
	}

	private startRecorder(stream: MediaStream, bitrate: number): void {
		this.recorder = new MediaRecorder(stream, {
			mimeType: MIME_TYPE,
			videoBitsPerSecond: bitrate,
		});

		this.recorder.ondataavailable = async (e) => {
			if (e.data.size > 0 && this.ws.readyState === WebSocket.OPEN) {
				const chunk = await e.data.arrayBuffer();
				const header = new ArrayBuffer(4);
				new DataView(header).setUint32(0, this.userId, false);
				const msg = new Uint8Array(4 + chunk.byteLength);
				msg.set(new Uint8Array(header), 0);
				msg.set(new Uint8Array(chunk), 4);
				this.ws.send(msg.buffer);
			}
		};

		this.recorder.onerror = (e) => {
			console.error('[Media] Recorder error:', e);
		};

		this.recorder.start(CHUNK_INTERVAL_MS);
	}

	stop(): void {
		if (this.recorder && this.recorder.state !== 'inactive') {
			this.recorder.stop();
		}
		if (this.localStream) {
			this.localStream.getTracks().forEach(t => t.stop());
			this.localStream = null;
		}
		this.remoteEntries.forEach((entry) => {
			entry.video.src = '';
			URL.revokeObjectURL(entry.video.src);
			entry.video.remove();
		});
		this.remoteEntries.clear();
	}
}
