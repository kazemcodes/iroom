/**
 * PionClient — WebRTC client for classroom video/audio.
 *
 * Handles:
 *   - Peer connection setup with STUN server
 *   - Local media stream (camera + microphone)
 *   - SDP offer/answer exchange via HTTP API
 *   - ICE candidate exchange
 *   - Screen sharing
 *   - Clean disconnect
 *
 * Usage:
 *   const client = new PionClient({ roomId, userId, role, displayName });
 *   client.onLocalStream = (stream) => { videoEl.srcObject = stream; };
 *   client.onRemoteStream = (stream, id) => { ... };
 *   await client.connect();
 *   // Later:
 *   client.disconnect();
 *
 * API calls made:
 *   POST /sessions/:id/classroom/offer — Send SDP offer, receive answer
 *   POST /sessions/:id/classroom/candidate — Send ICE candidate
 *   DELETE /sessions/:id/classroom/:userId — Leave room
 */
export interface PionRoomConfig {
	roomId: string;
	userId: string;
	role: string;
	displayName: string;
	userRole: string;
	listener?: boolean;
	stunUrl?: string;
	turnUrl?: string;
	turnUsername?: string;
	turnCredential?: string;
}

export class PionClient {
	private pc: RTCPeerConnection | null = null;
	private roomId: string;
	private userId: string;
	private displayName: string;
	private userRole: string;
	private localStream: MediaStream | null = null;
	private listener: boolean;
	private stunUrl: string | null = null;
	private turnUrl: string | null = null;
	private turnUsername: string | null = null;
	private turnCredential: string | null = null;
	private screenTrack: MediaStreamTrack | null = null;
	private screenStream: MediaStream | null = null;
	private isRenegotiating = false;

	onRemoteStream?: (stream: MediaStream, participantId: string) => void;
	onLocalStream?: (stream: MediaStream) => void;

	constructor(config: PionRoomConfig) {
		this.roomId = config.roomId;
		this.userId = config.userId;
		this.displayName = config.displayName;
		this.userRole = config.userRole || 'student';
		this.listener = config.listener || false;
		this.stunUrl = config.stunUrl || null;
		this.turnUrl = config.turnUrl || null;
		this.turnUsername = config.turnUsername || null;
		this.turnCredential = config.turnCredential || null;
	}

	async connect(): Promise<void> {
		console.log('[Pion] Starting connect...');
		if (!this.listener) {
			this.localStream = await navigator.mediaDevices.getUserMedia({
				video: true,
				audio: true
			});
			console.log('[Pion] Got local stream');
		} else {
			console.log('[Pion] Listener mode — no media');
		}

		if (this.onLocalStream && this.localStream) {
			this.onLocalStream(this.localStream);
		}

		const token = localStorage.getItem('access_token');
		if (!token) throw new Error('No auth token');
		console.log('[Pion] Got token');

		const iceServers: RTCIceServer[] = [];
		if (this.turnUrl && this.turnUsername && this.turnCredential) {
			iceServers.push({
				urls: this.turnUrl,
				username: this.turnUsername,
				credential: this.turnCredential,
			});
		}
		if (this.stunUrl) {
			iceServers.push({ urls: this.stunUrl });
		}
		this.pc = new RTCPeerConnection({ iceServers });
		console.log('[Pion] Created RTCPeerConnection');

		this.localStream?.getTracks().forEach(track => {
			if (this.pc) {
				this.pc.addTrack(track, this.localStream!);
			}
		});

		if (this.listener && this.pc) {
			this.pc.addTransceiver('audio', { direction: 'recvonly' });
			this.pc.addTransceiver('video', { direction: 'recvonly' });
		}

		this.pc.onicecandidate = async (event) => {
			if (event.candidate) {
				console.log('[Pion] Sending ICE candidate');
				await fetch(`/api/v1/sessions/${this.roomId}/classroom/candidate`, {
					method: 'POST',
					headers: {
						'Content-Type': 'application/json',
						'Authorization': `Bearer ${token}`
					},
					body: JSON.stringify({
						candidate: event.candidate.candidate,
						sdp_mid: event.candidate.sdpMid,
						sdp_m_line_index: event.candidate.sdpMLineIndex,
						room_id: this.roomId,
						user_id: this.userId
					})
				});
			}
		};

		// Map to track pending tracks awaiting participant ID from signaling
		const pendingTracks = new Map<string, { stream: MediaStream; track: MediaStreamTrack }>();

		this.pc.ondatachannel = (event) => {
			console.log('[Pion] ondatachannel', { label: event.channel.label });
			const dc = event.channel;
			dc.onmessage = (msg) => {
				try {
					const data = JSON.parse(msg.data);
					console.log('[Pion] datachannel message', data);
					if (data.type === 'track_added') {
						// Store the mapping for future ontrack events
						streamIdToUserId.set(data.track_id, data.user_id);

						// Match pending track by track_id
						const pending = pendingTracks.get(data.track_id);
						if (pending) {
							pendingTracks.delete(data.track_id);
							console.log('[Pion] track_added matched pending track', { trackId: data.track_id, userId: data.user_id });
							if (this.onRemoteStream && pending.stream) {
								this.onRemoteStream(pending.stream, data.user_id);
							}
						} else {
							console.log('[Pion] track_added no pending track for', data.track_id, '- waiting for ontrack');
						}
					}
				} catch {}
			};
		};

		const streamIdToUserId = new Map<string, string>();

		this.pc.ontrack = (event) => {
			const stream = event.streams[0];
			const streamId = stream?.id || 'unknown';
			const trackId = event.track.id;
			console.log('[Pion] ontrack', { trackId, kind: event.track.kind, streamId, hasStream: !!stream });

			// Server sets streamID = sender's userID for direct identification
			const userId = streamIdToUserId.get(trackId) || streamIdToUserId.get(streamId) || (streamId !== 'unknown' ? streamId : null);
			if (userId) {
				console.log('[Pion] ontrack resolved', { userId, source: streamIdToUserId.has(trackId) ? 'signaling' : 'streamId' });
				if (this.onRemoteStream && stream) {
					this.onRemoteStream(stream, userId);
				}
			} else {
				console.log('[Pion] ontrack pending', { trackId, streamId });
				if (stream) {
					pendingTracks.set(trackId, { stream, track: event.track });
					pendingTracks.set(streamId, { stream, track: event.track });
				}
			}
		};

		this.pc.onconnectionstatechange = () => {
			if (this.pc?.connectionState === 'failed' || this.pc?.connectionState === 'closed') {
				this.disconnect();
			}
		};

		const offer = await this.pc.createOffer();
		await this.pc.setLocalDescription(offer);
		console.log('[Pion] Created SDP offer, sending to server...');

		const offerRes = await fetch(`/api/v1/sessions/${this.roomId}/classroom/offer`, {
			method: 'POST',
			headers: {
				'Content-Type': 'application/json',
				'Authorization': `Bearer ${token}`
			},
			body: JSON.stringify({
				sdp: offer.sdp,
				room_id: this.roomId,
				user_id: this.userId,
				name: this.displayName,
				role: this.userRole
			})
		});

		const offerData = await offerRes.json();
		console.log('[Pion] Offer response:', offerData);
		if (!offerData.success) throw new Error(offerData.error || 'Failed to send offer');

		const answer = new RTCSessionDescription({ type: 'answer', sdp: offerData.data.sdp });
		await this.pc.setRemoteDescription(answer);
		console.log('[Pion] Connected!');
	}

	async toggleAudio(): Promise<void> {
		if (this.localStream) {
			this.localStream.getAudioTracks().forEach(t => { t.enabled = !t.enabled; });
		}
	}

	async toggleVideo(): Promise<void> {
		if (this.localStream) {
			this.localStream.getVideoTracks().forEach(t => { t.enabled = !t.enabled; });
		}
	}

	async shareScreen(): Promise<void> {
		const screenStream = await navigator.mediaDevices.getDisplayMedia({ video: true });
		this.screenStream = screenStream;
		this.screenTrack = screenStream.getVideoTracks()[0];
		if (this.pc && this.screenTrack) {
			this.pc.addTrack(this.screenTrack, screenStream);
			this.screenTrack.onended = () => {
				this.stopScreenShare();
			};
			await this.renegotiate();
		}
	}

	async stopScreenShare(): Promise<void> {
		if (this.screenTrack) {
			this.screenTrack.stop();
			this.screenTrack = null;
		}
		this.screenStream = null;

		if (this.pc) {
			const senders = this.pc.getSenders();
			for (const sender of senders) {
				if (sender.track && !this.localStream?.getTracks().includes(sender.track)) {
					this.pc.removeTrack(sender);
				}
			}
			await this.renegotiate();
		}
	}

	private async renegotiate(): Promise<void> {
		if (!this.pc || this.isRenegotiating) return;
		this.isRenegotiating = true;
		try {
			const offer = await this.pc.createOffer();
			await this.pc.setLocalDescription(offer);

			const token = localStorage.getItem('access_token');
			if (!token) throw new Error('No auth token');

			const offerRes = await fetch(`/api/v1/sessions/${this.roomId}/classroom/offer`, {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
					'Authorization': `Bearer ${token}`
				},
				body: JSON.stringify({
					sdp: offer.sdp,
					room_id: this.roomId,
					user_id: this.userId,
					name: this.displayName,
					role: this.userRole
				})
			});

			const offerData = await offerRes.json();
			if (!offerData.success) throw new Error(offerData.error || 'Failed to renegotiate');

			const answer = new RTCSessionDescription({ type: 'answer', sdp: offerData.data.sdp });
			await this.pc.setRemoteDescription(answer);
		} catch (e) {
			console.error('[Pion] Renegotiation failed:', e);
			throw e;
		} finally {
			this.isRenegotiating = false;
		}
	}

	async leave(): Promise<void> {
		const token = localStorage.getItem('access_token');
		if (token) {
			await fetch(`/api/v1/sessions/${this.roomId}/classroom/${this.userId}`, {
				method: 'DELETE',
				headers: { 'Authorization': `Bearer ${token}` }
			}).catch(() => {});
		}
		this.disconnect();
	}

	disconnect(): void {
		if (this.localStream) {
			this.localStream.getTracks().forEach(t => t.stop());
			this.localStream = null;
		}
		if (this.pc) {
			this.pc.close();
			this.pc = null;
		}
	}
}
