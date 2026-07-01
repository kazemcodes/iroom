/**
 * Screen Share Toggle Controller
 *
 * Extracted from the room page to make screen share toggle logic testable.
 * Handles: UI button toggles, browser-initiated screen share stops, and
 * WS command synchronization — without depending on WebSocket or MediaClient instances.
 */

export interface ScreenShareToggleDeps {
	/** Send a command over WebSocket (e.g. screenshare_on, screenshare_off). */
	sendCommand: (cmd: 'screenshare_on' | 'screenshare_off') => void;
	/** Call mediaClient.shareScreen() — throws on failure. */
	shareScreen: () => Promise<void>;
	/** Call mediaClient.stopScreenShare(). */
	stopScreenShare: () => Promise<void>;
	/** Whether a mediaClient instance is available. */
	hasMediaClient: boolean;
}

/**
 * Creates a screen share toggle controller.
 *
 * The caller owns the `screenShareOn` reactive state; this controller returns
 * the target state after each toggle so the caller can update it.
 */
export function createScreenShareToggle(deps: ScreenShareToggleDeps) {
	let _toggling = $state(false);

	function sendCmd(cmd: 'screenshare_on' | 'screenshare_off') {
		deps.sendCommand(cmd);
	}

	return {
		/** Whether a toggle operation is in progress (prevents re-entrance). */
		get toggling(): boolean {
			return _toggling;
		},

		/**
		 * Toggle screen share on/off.
		 * @param screenShareOn current screen share state
		 * @returns true if turned on, false if turned off, null if no-op or error
		 */
		async toggle(screenShareOn: boolean): Promise<boolean | null> {
			if (_toggling) return null;
			_toggling = true;
			try {
				if (deps.hasMediaClient) {
					if (!screenShareOn) {
						// TURN ON
						await deps.shareScreen();
						sendCmd('screenshare_on');
						return true;
					} else {
						// TURN OFF
						await deps.stopScreenShare();
						sendCmd('screenshare_off');
						return false;
					}
				} else {
					const newState = !screenShareOn;
					sendCmd(newState ? 'screenshare_on' : 'screenshare_off');
					return newState;
				}
			} catch (e) {
				// On error (user cancelled or API failure), force state to false
				sendCmd('screenshare_off');
				return false;
			} finally {
				_toggling = false;
			}
		},

		/**
		 * Handle browser-initiated screen share stop.
		 * Sends screenshare_off WS command. Caller should also set screenShareOn = false.
		 */
		onEnded(): void {
			sendCmd('screenshare_off');
		},
	};
}
