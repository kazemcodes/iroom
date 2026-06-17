export interface JanusRoomConfig {
	wsUrl: string;
	roomId: number;
	userId: number;
	role: string;
	displayName: string;
}

export interface Participant {
	id: number;
	display: string;
	audio: boolean;
	video: boolean;
	streams: any[];
}

export class JanusClient {
	private session: any = null;
	private pluginHandle: any = null;
	private roomId: number;
	private userId: number;
	private role: string;
	private displayName: string;
	private wsUrl: string;

	onParticipantJoined?: (p: Participant) => void;
	onParticipantLeft?: (p: Participant) => void;
	onStream?: (stream: MediaStream, participant: Participant) => void;

	constructor(config: JanusRoomConfig) {
		this.wsUrl = config.wsUrl;
		this.roomId = config.roomId;
		this.userId = config.userId;
		this.role = config.role;
		this.displayName = config.displayName;
	}

	async connect(): Promise<void> {
		return new Promise((resolve, reject) => {
			(Janus as any).init({
				debug: 'all',
				callback: () => {
					this.session = new (Janus as any)({
						server: this.wsUrl,
						success: () => resolve(),
						error: (err: any) => reject(err),
					});
				}
			});
		});
	}

	async joinRoom(publishAudio = true, publishVideo = true): Promise<void> {
		return new Promise((resolve, reject) => {
			this.session.attach({
				plugin: "janus.plugin.videoroom",
				success: (pluginHandle: any) => {
					this.pluginHandle = pluginHandle;
					this.pluginHandle.send({
						message: {
							request: "join",
							room: this.roomId,
							ptype: "publisher",
							display: this.displayName,
						},
						success: () => resolve(),
						error: (err: any) => reject(err),
					});
				},
				error: (err: any) => reject(err),
				onmessage: (msg: any, jsep: any) => {
					this.handleMessage(msg, jsep);
				},
				onlocalstream: (stream: MediaStream) => {
					const localVideo = document.getElementById('local-video') as HTMLVideoElement;
					if (localVideo) {
						(Janus as any).attachMediaStream(localVideo, stream);
					}
				},
				onremotestream: (stream: MediaStream) => {
					this.onStream?.(stream, { id: 0, display: '', audio: true, video: true, streams: [] });
				},
			});
		});
	}

	private handleMessage(msg: any, jsep: any) {
		const event = msg["videoroom"];
		if (event) {
			switch (event) {
				case "joined":
					this.handleJoined(msg);
					break;
				case "newPublisher":
					this.handleNewPublisher(msg);
					break;
				case "destroyed":
					break;
				case "event":
					break;
			}
		}
		if (jsep) {
			this.pluginHandle.handleJsep({ jsep });
		}
	}

	private handleJoined(msg: any) {
		const list = msg["publishers"] || [];
		for (const pub of list) {
			this.subscribeToPublisher(pub);
		}
	}

	private handleNewPublisher(msg: any) {
		const pub = msg["publishers"][0];
		this.onParticipantJoined?.({
			id: pub.id,
			display: pub.display,
			audio: pub.audio,
			video: pub.video,
			streams: pub.streams,
		});
		this.subscribeToPublisher(pub);
	}

	private subscribeToPublisher(pub: any) {
		this.pluginHandle.send({
			message: {
				request: "join",
				room: this.roomId,
				ptype: "subscriber",
				feed: pub.id,
			},
		});
	}

	async toggleAudio(): Promise<void> {
		if (!this.pluginHandle) return;
		this.pluginHandle.send({
			message: { request: "configure", audio: true },
		});
	}

	async toggleVideo(): Promise<void> {
		if (!this.pluginHandle) return;
		this.pluginHandle.send({
			message: { request: "configure", video: true },
		});
	}

	async shareScreen(): Promise<void> {
		const stream = await navigator.mediaDevices.getDisplayMedia({ video: true });
		const track = stream.getVideoTracks()[0];
		this.pluginHandle.send({
			message: { request: "configure", video: true },
			jsep: { type: 'offer', sdp: '' },
		});
	}

	async sendChatMessage(text: string): Promise<void> {
		this.pluginHandle.data({
			text: JSON.stringify({ textchat: true, room: this.roomId, message: text }),
		});
	}

	async leave(): Promise<void> {
		if (this.pluginHandle) {
			this.pluginHandle.send({ message: { request: "leave" } });
			this.pluginHandle.hangup();
		}
		if (this.session) {
			this.session.destroy();
		}
	}
}
