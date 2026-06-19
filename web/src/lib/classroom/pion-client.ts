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
}

export class PionClient {
	private pc: RTCPeerConnection | null = null;
	private roomId: string;
	private userId: string;
	private displayName: string;
	private userRole: string;
	private localStream: MediaStream | null = null;
	private listener: boolean;

	onRemoteStream?: (stream: MediaStream, participantId: string) => void;
	onLocalStream?: (stream: MediaStream) => void;

	constructor(config: PionRoomConfig) {
		this.roomId = config.roomId;
		this.userId = config.userId;
		this.displayName = config.displayName;
		this.userRole = config.userRole || 'student';
		this.listener = config.listener || false;
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

		this.pc = new RTCPeerConnection({
			iceServers: [{ urls: 'stun:stun.l.google.com:19302' }]
		});
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

		this.pc.ontrack = (event) => {
			const id = event.streams[0]?.id || 'unknown';
			if (this.onRemoteStream) {
				this.onRemoteStream(event.streams[0], id);
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
		const screenTrack = screenStream.getVideoTracks()[0];
		if (this.pc && screenTrack) {
			const sender = this.pc.getSenders().find(s => s.track?.kind === 'video');
			if (sender) {
				await sender.replaceTrack(screenTrack);
			}
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
