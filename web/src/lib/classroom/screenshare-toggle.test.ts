import { describe, it, expect, vi, beforeEach } from 'vitest';
import { createScreenShareToggle, type ScreenShareToggleDeps } from '$lib/classroom/screenshare-toggle.svelte';

/**
 * Integration-style tests for the screen share toggle logic.
 *
 * Covers both UI button click paths and browser-initiated screen share stop paths,
 * as well as edge cases like permission guards, re-entrance prevention,
 * and command handling on error.
 */

function makeDeps(overrides: Partial<ScreenShareToggleDeps> = {}): ScreenShareToggleDeps {
	return {
		sendCommand: vi.fn(),
		shareScreen: vi.fn(async () => {}),
		stopScreenShare: vi.fn(async () => {}),
		hasMediaClient: true,
		...overrides,
	};
}

describe('ScreenShareToggle — UI button clicks (toggle)', () => {
	let deps: ReturnType<typeof makeDeps>;

	beforeEach(() => {
		vi.clearAllMocks();
		deps = makeDeps();
	});

	it('calls shareScreen and sends screenshare_on when turning on', async () => {
		const ctrl = createScreenShareToggle(deps);

		const result = await ctrl.toggle(false);

		expect(result).toBe(true);
		expect(deps.shareScreen).toHaveBeenCalledTimes(1);
		expect(deps.stopScreenShare).not.toHaveBeenCalled();
		expect(deps.sendCommand).toHaveBeenCalledWith('screenshare_on');
	});

	it('calls stopScreenShare and sends screenshare_off when turning off', async () => {
		const ctrl = createScreenShareToggle(deps);

		const result = await ctrl.toggle(true);

		expect(result).toBe(false);
		expect(deps.stopScreenShare).toHaveBeenCalledTimes(1);
		expect(deps.shareScreen).not.toHaveBeenCalled();
		expect(deps.sendCommand).toHaveBeenCalledWith('screenshare_off');
	});

	it('toggles without media client (just sends command and returns new state)', async () => {
		deps = makeDeps({ hasMediaClient: false });

		const ctrl = createScreenShareToggle(deps);

		const result = await ctrl.toggle(false);

		expect(result).toBe(true);
		expect(deps.sendCommand).toHaveBeenCalledWith('screenshare_on');
		expect(deps.shareScreen).not.toHaveBeenCalled();
		expect(deps.stopScreenShare).not.toHaveBeenCalled();
	});

	it('toggles without media client from on to off', async () => {
		deps = makeDeps({ hasMediaClient: false });

		const ctrl = createScreenShareToggle(deps);

		const result = await ctrl.toggle(true);

		expect(result).toBe(false);
		expect(deps.sendCommand).toHaveBeenCalledWith('screenshare_off');
	});

	it('sends screenshare_off and returns false when shareScreen throws', async () => {
		deps = makeDeps({
			shareScreen: vi.fn(async () => { throw new Error('permission denied'); }),
		});

		const ctrl = createScreenShareToggle(deps);

		const result = await ctrl.toggle(false);

		expect(result).toBe(false);
		expect(deps.sendCommand).toHaveBeenCalledWith('screenshare_off');
	});

	it('sends screenshare_off and returns false when stopScreenShare throws', async () => {
		deps = makeDeps({
			stopScreenShare: vi.fn(async () => { throw new Error('already stopped'); }),
		});

		const ctrl = createScreenShareToggle(deps);

		const result = await ctrl.toggle(true);

		expect(result).toBe(false);
		expect(deps.sendCommand).toHaveBeenCalledWith('screenshare_off');
	});

	it('prevents re-entrance (toggling guard)', async () => {
		let resolveShare!: () => void;
		const shareDeferred = new Promise<void>((resolve) => {
			resolveShare = resolve;
		});
		deps = makeDeps({
			shareScreen: vi.fn(() => shareDeferred),
		});

		const ctrl = createScreenShareToggle(deps);

		// Start first toggle
		const first = ctrl.toggle(false);
		expect(ctrl.toggling).toBe(true);

		// Try second toggle while first is in progress
		const second = await ctrl.toggle(true);
		expect(second).toBeNull(); // re-entrance prevented

		// Resolve the first toggle
		resolveShare();
		const firstResult = await first;
		expect(firstResult).toBe(true);
		expect(ctrl.toggling).toBe(false);
	});

	it('clears toggling flag even when shareScreen throws', async () => {
		deps = makeDeps({
			shareScreen: vi.fn(async () => { throw new Error('boom'); }),
		});

		const ctrl = createScreenShareToggle(deps);

		const result = await ctrl.toggle(false);
		expect(result).toBe(false);
		expect(ctrl.toggling).toBe(false);
	});
});

describe('ScreenShareToggle — browser-initiated screen share stop (onEnded)', () => {
	let deps: ReturnType<typeof makeDeps>;

	beforeEach(() => {
		vi.clearAllMocks();
		deps = makeDeps();
	});

	it('sends screenshare_off WS command', () => {
		const ctrl = createScreenShareToggle(deps);

		ctrl.onEnded();

		expect(deps.sendCommand).toHaveBeenCalledTimes(1);
		expect(deps.sendCommand).toHaveBeenCalledWith('screenshare_off');
	});

	it('does not call shareScreen or stopScreenShare (browser already killed the track)', () => {
		const ctrl = createScreenShareToggle(deps);

		ctrl.onEnded();

		expect(deps.shareScreen).not.toHaveBeenCalled();
		expect(deps.stopScreenShare).not.toHaveBeenCalled();
	});

	it('does not set toggling flag (not a programmatic toggle)', () => {
		const ctrl = createScreenShareToggle(deps);

		ctrl.onEnded();

		expect(ctrl.toggling).toBe(false);
	});

	it('works even without media client (just sends command)', () => {
		deps = makeDeps({ hasMediaClient: false });

		const ctrl = createScreenShareToggle(deps);

		ctrl.onEnded();

		expect(deps.sendCommand).toHaveBeenCalledWith('screenshare_off');
	});

	it('is idempotent — safe to call multiple times', () => {
		const ctrl = createScreenShareToggle(deps);

		ctrl.onEnded();
		ctrl.onEnded();
		ctrl.onEnded();

		expect(deps.sendCommand).toHaveBeenCalledTimes(3);
		expect(deps.shareScreen).not.toHaveBeenCalled();
		expect(deps.stopScreenShare).not.toHaveBeenCalled();
	});
});

describe('ScreenShareToggle — integration scenarios', () => {
	it('full cycle: turn on via button → browser stops share', async () => {
		const sentCommands: string[] = [];
		const deps: ScreenShareToggleDeps = {
			sendCommand: (cmd) => sentCommands.push(cmd),
			shareScreen: vi.fn(async () => {}),
			stopScreenShare: vi.fn(async () => {}),
			hasMediaClient: true,
		};

		const ctrl = createScreenShareToggle(deps);

		// User clicks screen share button to turn on
		let screenOn = false;
		const result = await ctrl.toggle(screenOn);
		expect(result).toBe(true);
		screenOn = result!;
		expect(screenOn).toBe(true);

		// Browser stops screen share (user clicks Chrome's "Stop sharing")
		ctrl.onEnded();
		screenOn = false;

		expect(sentCommands).toEqual(['screenshare_on', 'screenshare_off']);
		expect(deps.shareScreen).toHaveBeenCalledTimes(1);
		expect(deps.stopScreenShare).not.toHaveBeenCalled();
	});

	it('full cycle: turn on → turn off via button', async () => {
		const sentCommands: string[] = [];
		const deps: ScreenShareToggleDeps = {
			sendCommand: (cmd) => sentCommands.push(cmd),
			shareScreen: vi.fn(async () => {}),
			stopScreenShare: vi.fn(async () => {}),
			hasMediaClient: true,
		};

		const ctrl = createScreenShareToggle(deps);

		// Turn on
		let screenOn = false;
		let result = await ctrl.toggle(screenOn);
		expect(result).toBe(true);
		screenOn = result!;

		// Turn off via button
		result = await ctrl.toggle(screenOn);
		expect(result).toBe(false);
		screenOn = result!;
		expect(screenOn).toBe(false);

		expect(sentCommands).toEqual(['screenshare_on', 'screenshare_off']);
		expect(deps.shareScreen).toHaveBeenCalledTimes(1);
		expect(deps.stopScreenShare).toHaveBeenCalledTimes(1);
	});

	it('button toggle after browser-stop re-enables screen share', async () => {
		const sentCommands: string[] = [];
		const deps: ScreenShareToggleDeps = {
			sendCommand: (cmd) => sentCommands.push(cmd),
			shareScreen: vi.fn(async () => {}),
			stopScreenShare: vi.fn(async () => {}),
			hasMediaClient: true,
		};

		const ctrl = createScreenShareToggle(deps);

		// Browser killed the screen share
		ctrl.onEnded();
		let screenOn = false;

		// User clicks to re-enable
		const result = await ctrl.toggle(screenOn);
		screenOn = result!;

		expect(sentCommands).toEqual(['screenshare_off', 'screenshare_on']);
		expect(screenOn).toBe(true);
	});

	it('shareScreen failure forces state to false and sends screenshare_off', async () => {
		const sentCommands: string[] = [];
		const deps: ScreenShareToggleDeps = {
			sendCommand: (cmd) => sentCommands.push(cmd),
			shareScreen: vi.fn(async () => { throw new Error('cancelled'); }),
			stopScreenShare: vi.fn(async () => {}),
			hasMediaClient: true,
		};

		const ctrl = createScreenShareToggle(deps);

		const result = await ctrl.toggle(false);

		expect(result).toBe(false);
		expect(sentCommands).toEqual(['screenshare_off']);
		expect(deps.shareScreen).toHaveBeenCalledTimes(1);
	});
});
