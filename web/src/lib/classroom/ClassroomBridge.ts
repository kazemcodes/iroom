export type BridgeMessageType =
	| 'classroom-closed'
	| 'chat-notification'
	| 'session-ended'
	| 'participant-joined'
	| 'participant-left'
	| 'toggle-mute'
	| 'toggle-video';

export interface BridgeMessage {
	type: BridgeMessageType;
	payload?: unknown;
}

const CHANNEL_NAME = 'classroom-bridge';

let channel: BroadcastChannel | null = null;
let messageHandler: ((event: MessageEvent<BridgeMessage>) => void) | null = null;

export function initBridge(onMessage?: (msg: BridgeMessage) => void): void {
	if (channel) return;

	channel = new BroadcastChannel(CHANNEL_NAME);
	messageHandler = (event: MessageEvent<BridgeMessage>) => {
		onMessage?.(event.data);
	};
	channel.addEventListener('message', messageHandler);
}

export function sendBridge(msg: BridgeMessage): void {
	channel?.postMessage(msg);
}

export function closeBridge(): void {
	if (messageHandler && channel) {
		channel.removeEventListener('message', messageHandler);
	}
	channel?.close();
	channel = null;
	messageHandler = null;
}
