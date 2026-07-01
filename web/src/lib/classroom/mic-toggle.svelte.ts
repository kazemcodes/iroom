/**
 * Mic Toggle Controller
 *
 * Extracted from the room page to make mic toggle logic testable.
 * Handles: UI button toggles, browser-initiated microphone stops, and
 * WS command synchronization — without depending on WebSocket or MediaClient instances.
 */

export interface MicToggleDeps {
	/** Send a command over WebSocket (e.g. mic_on, mic_off). */
	sendCommand: (cmd: 'mic_on' | 'mic_off') => void;
	/** Call mediaClient.toggleAudio() — returns true on success. */
	toggleAudio: () => Promise<boolean>;
	/** Whether a mediaClient instance is available. */
	hasMediaClient: boolean;
}

/**
 * Creates a mic toggle controller.
 *
 * The caller owns the `micOn` reactive state; this controller returns
 * the target state after each toggle so the caller can update it.
 */
export function createMicToggle(deps: MicToggleDeps) {
	let _toggling = $state(false);

	function sendCmd(cmd: 'mic_on' | 'mic_off') {
		deps.sendCommand(cmd);
	}

	return {
		/** Whether a toggle operation is in progress (prevents re-entrance). */
		get toggling(): boolean {
			return _toggling;
		},

		/**
		 * Toggle mic on/off.
		 * @param micOn current mic state
		 * @returns true if turned on, false if turned off, null if no-op or rolled back
		 */
		async toggle(micOn: boolean): Promise<boolean | null> {
			if (_toggling) return null;
			_toggling = true;
			try {
				if (deps.hasMediaClient) {
					const targetOn = !micOn;
					sendCmd(targetOn ? 'mic_on' : 'mic_off');
					await new Promise((r) => setTimeout(r, 100));
					const ok = await deps.toggleAudio();
					if (!ok) {
						// Roll back the WS command
						sendCmd(!targetOn ? 'mic_on' : 'mic_off');
						return null;
					}
					return targetOn;
				} else {
					const newState = !micOn;
					sendCmd(newState ? 'mic_on' : 'mic_off');
					return newState;
				}
			} finally {
				_toggling = false;
			}
		},

		/**
		 * Handle browser-initiated microphone stop.
		 * Sends mic_off WS command. Caller should also set micOn = false.
		 */
		onEnded(): void {
			sendCmd('mic_off');
		},
	};
}
