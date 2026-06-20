/**
 * MediaClient — WebSocket-based video/audio for classroom.
 *
 * Uses MediaRecorder (VP8+Opus encoding) → WebSocket binary → MediaSource playback.
 * Chrome-only. No WebRTC, no STUN/TURN, no ICE.
 *
 * Binary message format: [userId: uint32 BE][WebM chunk]
 *
 * Adaptive quality: bitrate adjusts based on participant count.
 */

const MIME_TYPE = 'video/webm;codecs=vp8,opus';

interface QualityTier {
	maxUsers: number;
	videoBitrate: number;
	audioBitrate: number;
	width: number;
	height: number;
	frameRate: number;
}

const QUALITY_TIERS: QualityTier[] = [
	{ maxUsers: 5,   videoBitrate: 1_500_000, audioBitrate: 64_000, width: 1280, height: 720,  frameRate: 30 },
	{ maxUsers: 15,  videoBitrate: 800_000,   audioBitrate: 48_000, width: 960,  height: 540,  frameRate: 24 },
	{ maxUsers: 30,  videoBitrate: 400_000,   audioBitrate: 32_000, width: 640,  height: 360,  frameRate: 20 },
	{ maxUsers: 60,  videoBitrate: 250_000,   audioBitrate: 24_000, width: 480,  height: 270,  frameRate: 15 },
	{ maxUsers: 999, videoBitrate: 128_000,   audioBitrate: 16_000, width: 320,  height: 180,  frameRate: 10 },
];

const CHUNK_INTERVAL_MS = 200;

function getTier(userCount: number): QualityTier {
	for (const tier of QUALITY_TIERS) {
		if (userCount <= tier.maxUsers) return tier;
	}
	return QUALITY_TIERS[QUALITY_TIERS.length - 1];
}

export class MediaClient {
	private ws: WebSocket;
	private userId: number;
	private localStream: MediaStream | null = null;
	private recorder: MediaRecorder | null = null;
	private videoEnabled = true;
	private audioEnabled = true;
	private currentTier: QualityTier = QUALITY_TIERS[3];
	private participantCount = 1;
	private localConstraints: { video: boolean; audio: boolean } = { video: true, audio: true };

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
	onQualityChange?: (tier: QualityTier, userCount: number) => void;

	constructor(ws: WebSocket, userId: number) {
		this.ws = ws;
		this.userId = userId;
	}

	async start(video = true, audio = true): Promise<void> {
		this.localConstraints = { video, audio };
		const tier = getTier(this.participantCount);
		this.currentTier = tier;

		const constraints: MediaStreamConstraints = {
			video: video ? { width: { ideal: tier.width }, height: { ideal: tier.height }, frameRate: { ideal: tier.frameRate } } : false,
			audio: audio ? { echoCancellation: true, noiseSuppression: true, autoGainControl: true } : false,
		};

		this.localStream = await navigator.mediaDevices.getUserMedia(constraints);
		if (this.onLocalStream) this.onLocalStream(this.localStream);

		this.startRecorder();
		console.log('[Media] Started', { tier: tier.width + 'p', videoBitrate: tier.videoBitrate, audioBitrate: tier.audioBitrate });
	}

	private startRecorder(): void {
		if (!this.localStream) return;
		if (!MediaRecorder.isTypeSupported(MIME_TYPE)) {
			console.error('[Media] MIME type not supported:', MIME_TYPE);
			return;
		}

		const tier = this.currentTier;

		this.recorder = new MediaRecorder(this.localStream, {
			mimeType: MIME_TYPE,
			videoBitsPerSecond: tier.videoBitrate,
			audioBitsPerSecond: tier.audioBitrate,
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

	updateParticipantCount(count: number): void {
		if (count === this.participantCount) return;
		this.participantCount = count;

		const newTier = getTier(count);
		if (newTier.videoBitrate !== this.currentTier.videoBitrate) {
			console.log('[Media] Quality change', {
				from: this.currentTier.width + 'p@' + (this.currentTier.videoBitrate / 1000) + 'kbps',
				to: newTier.width + 'p@' + (newTier.videoBitrate / 1000) + 'kbps',
				users: count,
			});
			this.currentTier = newTier;
			this.restartWithNewConstraints();
			if (this.onQualityChange) this.onQualityChange(newTier, count);
		}
	}

	private async restartWithNewConstraints(): Promise<void> {
		if (!this.localStream || !this.localConstraints.video) return;

		const tier = this.currentTier;
		const newConstraints: MediaStreamConstraints = {
			video: { width: { ideal: tier.width }, height: { ideal: tier.height }, frameRate: { ideal: tier.frameRate } },
			audio: this.localConstraints.audio ? { echoCancellation: true, noiseSuppression: true, autoGainControl: true } : false,
		};

		try {
			const newStream = await navigator.mediaDevices.getUserMedia(newConstraints);
			const oldTracks = this.localStream.getVideoTracks();
			oldTracks.forEach(t => { t.stop(); this.localStream!.removeTrack(t); });
			newStream.getVideoTracks().forEach(t => this.localStream!.addTrack(t));

			if (this.localVideoCallback) this.localVideoCallback(this.localStream);

			if (this.recorder && this.recorder.state !== 'inactive') {
				this.recorder.stop();
			}
			this.startRecorder();
		} catch (e) {
			console.warn('[Media] Failed to resize:', e);
		}
	}

	private localVideoCallback?: (stream: MediaStream) => void;

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

	private createRemoteEntry(senderId: string) {
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
				entry.sb.onupdateend = () => this.drainQueue(entry);
				entry.ready = true;
				this.drainQueue(entry);
			} catch (e) {
				console.error('[Media] SourceBuffer error:', e);
			}
		});

		video.onloadeddata = () => {
			try {
				const capFn = (video as any).captureStream || (video as any).mozCaptureStream;
				if (capFn) {
					const stream = capFn.call(video);
					if (stream && this.onRemoteStream) {
						this.onRemoteStream(stream, senderId);
					}
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

		screenTrack.onended = () => this.stopScreenShare();

		if (this.recorder.state !== 'inactive') this.recorder.stop();

		this.recorder = new MediaRecorder(composedStream, {
			mimeType: MIME_TYPE,
			videoBitsPerSecond: this.currentTier.videoBitrate * 2,
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

		this.recorder.start(CHUNK_INTERVAL_MS);
	}

	stopScreenShare(): void {
		if (!this.localStream) return;
		if (this.recorder && this.recorder.state !== 'inactive') this.recorder.stop();
		this.startRecorder();
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
