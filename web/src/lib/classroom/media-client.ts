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

/** Offset added to userId for screen-share streams so receivers can distinguish them from webcam. */
const SCREEN_SHARE_ID_OFFSET = 1_000_000;

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

// WebM EBML header magic bytes — every init segment starts with these 4 bytes
const EBML_HEADER = new Uint8Array([0x1A, 0x45, 0xDF, 0xA3]);

function isInitSegment(chunk: ArrayBuffer): boolean {
	if (chunk.byteLength < 4) return false;
	const v = new DataView(chunk);
	return v.getUint32(0, false) === 0x1A45DFA3;
}

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
	private screenRecorder: MediaRecorder | null = null;
	private screenStream: MediaStream | null = null;
	private videoEnabled = true;
	private audioEnabled = true;
	private currentTier: QualityTier = QUALITY_TIERS[3];
	private participantCount = 1;
	private localConstraints: { video: boolean; audio: boolean } = { video: true, audio: true };

	private remoteEntries = new Map<string, {
		ms: MediaSource;
		url: string;
		video?: HTMLVideoElement;
		sb: SourceBuffer | null;
		ready: boolean;
		hasInit: boolean;
		pendingChunks: ArrayBuffer[];
	}>();

	onLocalStream?: (stream: MediaStream) => void;
	onRemoteStream?: (stream: MediaStream, participantId: string) => void;
	onRemoteStreamEnd?: (participantId: string) => void;
	onQualityChange?: (tier: QualityTier, userCount: number) => void;
	onScreenStream?: (stream: MediaStream | null) => void;

	constructor(ws: WebSocket, userId: number) {
		this.ws = ws;
		this.userId = userId;
	}

	async start(video = true, audio = true): Promise<void> {
		this.localConstraints = { video, audio };
		const tier = getTier(this.participantCount);
		this.currentTier = tier;

		if (!video && !audio) {
			this.localStream = new MediaStream();
			if (this.onLocalStream) this.onLocalStream(this.localStream);
			return;
		}

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
		if (this.localStream.getTracks().length === 0) return;
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
				const isInit = isInitSegment(chunk);
				console.debug('[stream] SEND chunk', { userId: this.userId, size: chunk.byteLength, isInit });
				const header = new ArrayBuffer(4);
				new DataView(header).setUint32(0, this.userId, false);
				const msg = new Uint8Array(4 + chunk.byteLength);
				msg.set(new Uint8Array(header), 0);
				msg.set(new Uint8Array(chunk), 4);
				this.ws.send(msg.buffer);
			}
		};

		this.recorder.onerror = (e) => {
			console.error('[stream] Recorder error:', e);
		};

		this.recorder.onstart = () => {
			const trackKinds = this.localStream ? this.localStream.getTracks().map((t: MediaStreamTrack) => t.kind) : [];
			console.debug('[stream] Recorder started', { userId: this.userId, tracks: trackKinds });
		};
		this.recorder.onstop = () => console.debug('[stream] Recorder stopped', { userId: this.userId });

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
		const rawSenderId = new DataView(data).getUint32(0, false);
		// Skip our own webcam stream (but not our screen share — others receive it)
		if (rawSenderId === this.userId) {
			console.debug('[stream] RECV skip own', { userId: this.userId });
			return;
		}

		const senderId = rawSenderId.toString();
		const chunk = data.slice(4);
		const initCheck = isInitSegment(chunk);

		let entry = this.remoteEntries.get(senderId);
		const isNew = !entry;
		if (!entry) {
			console.debug('[stream] RECV new entry', { senderId, chunkSize: chunk.byteLength, isInit: initCheck });
			entry = this.createRemoteEntry(senderId);
			this.remoteEntries.set(senderId, entry);
		}

		// Drop residual chunks from the old recorder — they're media segments
		// without init headers and would corrupt the fresh SourceBuffer
		if (!entry.hasInit && !initCheck) {
			console.debug('[stream] RECV drop residual', { senderId, chunkSize: chunk.byteLength, isNew, reason: 'no init segment yet' });
			return;
		}

		if (!entry.hasInit && initCheck) {
			console.debug('[stream] RECV init segment', { senderId, chunkSize: chunk.byteLength });
			entry.hasInit = true;
		}

		if (entry.ms.readyState !== 'open') {
			console.debug('[stream] RECV queue pending', { senderId, readyState: entry.ms.readyState });
			entry.pendingChunks.push(chunk);
			return;
		}

		if (entry.ready && entry.sb && !entry.sb.updating) {
			try {
				console.debug('[stream] RECV append', { senderId, chunkSize: chunk.byteLength });
				entry.sb.appendBuffer(new Uint8Array(chunk));
			} catch (e) {
				console.warn('[stream] appendBuffer error:', e, { senderId });
				entry.ready = false;
			}
		} else {
			console.debug('[stream] RECV queue pending (sb busy)', { senderId, ready: entry.ready, hasSb: !!entry.sb, updating: entry.sb?.updating });
			entry.pendingChunks.push(chunk);
		}
	}

	private createRemoteEntry(senderId: string) {
		const ms = new MediaSource();
		const url = URL.createObjectURL(ms);

		console.debug('[stream] createRemoteEntry', { senderId });

		const entry = { ms, url, video: undefined as HTMLVideoElement | undefined, sb: null as SourceBuffer | null, ready: false, hasInit: false, pendingChunks: [] as ArrayBuffer[] };

		ms.addEventListener('sourceopen', () => {
			console.debug('[stream] sourceopen', { senderId, pendingCount: entry.pendingChunks.length });
			try {
				entry.sb = ms.addSourceBuffer(MIME_TYPE);
				entry.sb.mode = 'sequence';
				entry.sb.onupdateend = () => this.drainQueue(entry);
				entry.ready = true;
				this.drainQueue(entry);
			} catch (e) {
				console.error('[stream] SourceBuffer error:', e, { senderId });
			}
		});

		if (this.onRemoteStream) {
			const video = document.createElement('video');
			video.autoplay = true;
			video.playsInline = true;
			video.muted = true;
			video.src = url;
			document.body.appendChild(video);

			video.onloadeddata = () => {
				console.debug('[stream] onloadeddata', { senderId, videoWidth: video.videoWidth, videoHeight: video.videoHeight });
				try {
					const capFn = (video as any).captureStream || (video as any).mozCaptureStream;
					if (capFn) {
						const stream = capFn.call(video);
						if (stream) {
						const tracks = stream.getTracks();
						console.debug('[stream] captureStream OK', { senderId, trackCount: tracks.length, kinds: tracks.map((t: MediaStreamTrack) => t.kind) });
							this.onRemoteStream!(stream, senderId);
						} else {
							console.warn('[stream] captureStream returned null/undefined', { senderId });
						}
					} else {
						console.warn('[stream] captureStream not available', { senderId });
					}
				} catch (e) {
					console.warn('[stream] captureStream failed:', e, { senderId });
				}
			};

			entry.video = video;
		}

		return entry;
	}

	private drainQueue(entry: { ms: MediaSource; sb: SourceBuffer | null; pendingChunks: ArrayBuffer[]; ready: boolean; hasInit: boolean }): void {
		if (!entry.ready || !entry.sb || entry.sb.updating || entry.ms.readyState !== 'open') return;
		while (entry.pendingChunks.length > 0) {
			if (entry.sb.updating || entry.ms.readyState !== 'open') break;
			const chunk = entry.pendingChunks.shift()!;
			// Skip residual chunks from old recorder if we haven't seen an init segment yet
			if (!entry.hasInit && !isInitSegment(chunk)) continue;
			if (!entry.hasInit && isInitSegment(chunk)) entry.hasInit = true;
			try {
				entry.sb.appendBuffer(new Uint8Array(chunk));
			} catch (e) {
				console.warn('[Media] appendBuffer error in drain:', e);
				break;
			}
		}
	}

	async toggleVideo(): Promise<boolean> {
		if (!this.localStream) return false;
		const tracks = this.localStream.getVideoTracks();
		if (tracks.length === 0) {
			// TURN ON: acquire video tracks
			try {
				const tier = this.currentTier;
				const stream = await navigator.mediaDevices.getUserMedia({
					video: { width: { ideal: tier.width }, height: { ideal: tier.height }, frameRate: { ideal: tier.frameRate } },
					audio: false,
				});
				stream.getVideoTracks().forEach(t => this.localStream!.addTrack(t));
				this.videoEnabled = true;
				this.localConstraints.video = true;
				if (this.recorder && this.recorder.state !== 'inactive') {
					console.debug('[stream] toggleVideo stopping old recorder');
					this.recorder.stop();
				}
				this.startRecorder();
				if (this.onLocalStream) this.onLocalStream(this.localStream);
				return true;
			} catch (e) {
				console.error('[Media] Failed to acquire video:', e);
				return false;
			}
		}
		// TURN OFF: stop and remove video tracks so recorder stops sending frames
		this.videoEnabled = false;
		this.localConstraints.video = false;
		tracks.forEach(t => { t.stop(); this.localStream!.removeTrack(t); });
		if (this.recorder && this.recorder.state !== 'inactive') {
			this.recorder.stop();
		}
		if (this.localStream.getAudioTracks().length > 0) {
			this.startRecorder();
		} else {
			this.recorder = null;
		}
		if (this.onLocalStream) this.onLocalStream(this.localStream);
		return true;
	}

	/**
	 * Force-stop video: stops video tracks, stops the recorder, and restarts
	 * recorder with audio-only so remote users stop receiving video frames.
	 * Called when admin disables our webcam.
	 */
	stopVideo(): void {
		if (!this.localStream) return;
		this.videoEnabled = false;
		this.localConstraints.video = false;
		const vt = this.localStream.getVideoTracks();
		vt.forEach(t => { t.stop(); this.localStream!.removeTrack(t); });
		if (this.recorder && this.recorder.state !== 'inactive') {
			this.recorder.stop();
		}
		if (this.localStream.getAudioTracks().length > 0) {
			this.startRecorder();
		} else {
			this.recorder = null;
		}
		if (this.onLocalStream) this.onLocalStream(this.localStream);
	}

	async toggleAudio(): Promise<boolean> {
		if (!this.localStream) return false;
		const tracks = this.localStream.getAudioTracks();
		if (tracks.length === 0) {
			// TURN ON: acquire audio tracks
			try {
				const stream = await navigator.mediaDevices.getUserMedia({
					video: false,
					audio: { echoCancellation: true, noiseSuppression: true, autoGainControl: true },
				});
				stream.getAudioTracks().forEach(t => this.localStream!.addTrack(t));
				this.audioEnabled = true;
				this.localConstraints.audio = true;
				if (this.recorder && this.recorder.state !== 'inactive') {
					console.debug('[stream] toggleAudio stopping old recorder');
					this.recorder.stop();
				}
				this.startRecorder();
				if (this.onLocalStream) this.onLocalStream(this.localStream);
				return true;
			} catch (e) {
				console.error('[Media] Failed to acquire audio:', e);
				return false;
			}
		}
		// TURN OFF: stop and remove audio tracks so recorder stops sending audio frames
		this.audioEnabled = false;
		this.localConstraints.audio = false;
		tracks.forEach(t => { t.stop(); this.localStream!.removeTrack(t); });
		if (this.recorder && this.recorder.state !== 'inactive') {
			this.recorder.stop();
		}
		if (this.localStream.getVideoTracks().length > 0) {
			this.startRecorder();
		} else {
			this.recorder = null;
		}
		if (this.onLocalStream) this.onLocalStream(this.localStream);
		return true;
	}

	async shareScreen(): Promise<void> {
		const screenStream = await navigator.mediaDevices.getDisplayMedia({ video: true });
		const screenTrack = screenStream.getVideoTracks()[0];

		if (!this.localStream) return;

		this.screenStream = screenStream;
		if (this.onScreenStream) this.onScreenStream(screenStream);

		screenTrack.onended = () => this.stopScreenShare();

		// Keep the webcam recorder running — screen share uses a separate ID
		const screenShareId = this.userId + SCREEN_SHARE_ID_OFFSET;

		const composedStream = new MediaStream([
			screenTrack,
			...this.localStream.getAudioTracks()
		]);

		this.screenRecorder = new MediaRecorder(composedStream, {
			mimeType: MIME_TYPE,
			videoBitsPerSecond: this.currentTier.videoBitrate * 2,
		});

		this.screenRecorder.ondataavailable = async (e) => {
			if (e.data.size > 0 && this.ws.readyState === WebSocket.OPEN) {
				const chunk = await e.data.arrayBuffer();
				const header = new ArrayBuffer(4);
				new DataView(header).setUint32(0, screenShareId, false);
				const msg = new Uint8Array(4 + chunk.byteLength);
				msg.set(new Uint8Array(header), 0);
				msg.set(new Uint8Array(chunk), 4);
				this.ws.send(msg.buffer);
			}
		};

		this.screenRecorder.start(CHUNK_INTERVAL_MS);
	}

	stopScreenShare(): void {
		if (!this.localStream) return;
		if (this.screenStream) {
			this.screenStream.getTracks().forEach(t => t.stop());
			this.screenStream = null;
			if (this.onScreenStream) this.onScreenStream(null);
		}
		if (this.screenRecorder && this.screenRecorder.state !== 'inactive') {
			this.screenRecorder.stop();
		}
		this.screenRecorder = null;
	}

	/**
	 * Destroy a remote participant's MediaSource entry so it can be recreated
	 * with the correct track configuration. Call when webcam_on/webcam_off
	 * commands are received, since the sender's recorder may have changed
	 * track composition (audio-only ↔ video+audio).
	 */
	resetRemoteStream(participantId: string): void {
		const entry = this.remoteEntries.get(participantId);
		if (!entry) {
			console.debug('[stream] resetRemoteStream no entry', { participantId });
			return;
		}
		console.debug('[stream] resetRemoteStream', { participantId, hadInit: entry.hasInit });
		if (entry.video) {
			entry.video.src = '';
			entry.video.remove();
		}
		URL.revokeObjectURL(entry.url);
		this.remoteEntries.delete(participantId);
	}

	stop(): void {
		if (this.recorder && this.recorder.state !== 'inactive') {
			this.recorder.stop();
		}
		if (this.screenRecorder && this.screenRecorder.state !== 'inactive') {
			this.screenRecorder.stop();
		}
		this.screenRecorder = null;
		if (this.localStream) {
			this.localStream.getTracks().forEach(t => t.stop());
			this.localStream = null;
		}
		this.remoteEntries.forEach((entry) => {
			if (entry.video) {
				entry.video.src = '';
				entry.video.remove();
			}
			URL.revokeObjectURL(entry.url);
		});
		this.remoteEntries.clear();
	}
}
