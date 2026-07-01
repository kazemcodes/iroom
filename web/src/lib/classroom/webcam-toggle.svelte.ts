/**
 * Webcam Toggle Controller
 *
 * Extracted from the room page to make webcam toggle logic testable.
 * Handles: UI button toggles, browser-initiated camera stops, and
 * WS command synchronization — without depending on WebSocket or MediaClient instances.
 */

export interface WebcamToggleDeps {
	/** Send a command over WebSocket (e.g. webcam_on, webcam_off). */
	sendCommand: (cmd: 'webcam_on' | 'webcam_off') => void;
	/** Call mediaClient.toggleVideo() — returns true on success. */
	toggleVideo: () => Promise<boolean>;
	/** Whether a mediaClient instance is available. */
	hasMediaClient: boolean;
}

/**
 * Creates a webcam toggle controller.
 *
 * The caller owns the `webcamOn` reactive state; this controller returns
 * the target state after each toggle so the caller can update it.
 */
export function createWebcamToggle(deps: WebcamToggleDeps) {
	let _toggling = $state(false);

	function sendCmd(cmd: 'webcam_on' | 'webcam_off') {
		deps.sendCommand(cmd);
	}

	return {
		/** Whether a toggle operation is in progress (prevents re-entrance). */
		get toggling(): boolean {
			return _toggling;
		},

		/**
		 * Toggle webcam on/off.
		 * @param webcamOn current webcam state
		 * @returns true if turned on, false if turned off, null if no-op or rolled back
		 */
		async toggle(webcamOn: boolean): Promise<boolean | null> {
			if (_toggling) return null;
			_toggling = true;
			try {
				if (deps.hasMediaClient) {
					const targetOn = !webcamOn;
					sendCmd(targetOn ? 'webcam_on' : 'webcam_off');
					await new Promise((r) => setTimeout(r, 100));
					const ok = await deps.toggleVideo();
					if (!ok) {
						// Roll back the WS command
						sendCmd(!targetOn ? 'webcam_on' : 'webcam_off');
						return null;
					}
					return targetOn;
				} else {
					const newState = !webcamOn;
					sendCmd(newState ? 'webcam_on' : 'webcam_off');
					return newState;
				}
			} finally {
				_toggling = false;
			}
		},

		/**
		 * Handle browser-initiated camera stop.
		 * Sends webcam_off WS command. Caller should also set webcamOn = false.
		 */
		onEnded(): void {
			sendCmd('webcam_off');
		},
	};
}
