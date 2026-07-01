import { describe, it, expect, vi, beforeEach } from 'vitest';
import { createWebcamToggle, type WebcamToggleDeps } from '$lib/classroom/webcam-toggle.svelte';

/**
 * Integration-style tests for the webcam toggle logic.
 *
 * Covers both UI button click paths and browser-initiated camera stop paths,
 * as well as edge cases like permission guards, re-entrance prevention,
 * and command rollback on failure.
 */

function makeDeps(overrides: Partial<WebcamToggleDeps> = {}): WebcamToggleDeps {
	return {
		sendCommand: vi.fn(),
		toggleVideo: vi.fn(async () => true),
		hasMediaClient: true,
		...overrides,
	};
}

describe('WebcamToggle — UI button clicks (toggle)', () => {
	let deps: ReturnType<typeof makeDeps>;

	beforeEach(() => {
		vi.clearAllMocks();
		deps = makeDeps();
	});

	it('sends webcam_on and calls toggleVideo when turning on', async () => {
		const ctrl = createWebcamToggle(deps);

		const result = await ctrl.toggle(false);

		expect(result).toBe(true);
		expect(deps.sendCommand).toHaveBeenNthCalledWith(1, 'webcam_on');
		expect(deps.toggleVideo).toHaveBeenCalledTimes(1);
	});

	it('sends webcam_off and calls toggleVideo when turning off', async () => {
		const ctrl = createWebcamToggle(deps);

		const result = await ctrl.toggle(true);

		expect(result).toBe(false);
		expect(deps.sendCommand).toHaveBeenNthCalledWith(1, 'webcam_off');
		expect(deps.toggleVideo).toHaveBeenCalledTimes(1);
	});

	it('toggles without media client (just sends command and returns new state)', async () => {
		deps = makeDeps({ hasMediaClient: false });

		const ctrl = createWebcamToggle(deps);

		const result = await ctrl.toggle(false);

		expect(result).toBe(true);
		expect(deps.sendCommand).toHaveBeenCalledWith('webcam_on');
		expect(deps.toggleVideo).not.toHaveBeenCalled();
	});

	it('toggles without media client from on to off', async () => {
		deps = makeDeps({ hasMediaClient: false });

		const ctrl = createWebcamToggle(deps);

		const result = await ctrl.toggle(true);

		expect(result).toBe(false);
		expect(deps.sendCommand).toHaveBeenCalledWith('webcam_off');
	});

	it('rolls back WS command when toggleVideo fails', async () => {
		deps = makeDeps({ toggleVideo: vi.fn(async () => false) });

		const ctrl = createWebcamToggle(deps);

		const result = await ctrl.toggle(false);

		expect(result).toBeNull();
		// First: webcam_on was sent
		expect(deps.sendCommand).toHaveBeenNthCalledWith(1, 'webcam_on');
		// Then: rollback to webcam_off
		expect(deps.sendCommand).toHaveBeenNthCalledWith(2, 'webcam_off');
		expect(deps.toggleVideo).toHaveBeenCalledTimes(1);
	});

	it('rolls back WS command when toggleVideo fails (off→on attempt)', async () => {
		deps = makeDeps({ toggleVideo: vi.fn(async () => false) });

		const ctrl = createWebcamToggle(deps);

		const result = await ctrl.toggle(true);

		expect(result).toBeNull();
		expect(deps.sendCommand).toHaveBeenNthCalledWith(1, 'webcam_off');
		expect(deps.sendCommand).toHaveBeenNthCalledWith(2, 'webcam_on');
	});

	it('prevents re-entrance (toggling guard)', async () => {
		// toggleVideo returns a promise that we control
		let resolveToggle!: (value: boolean) => void;
		const toggleDeferred = new Promise<boolean>((resolve) => {
			resolveToggle = resolve;
		});
		deps = makeDeps({
			toggleVideo: vi.fn(() => toggleDeferred),
		});

		const ctrl = createWebcamToggle(deps);

		// Start first toggle
		const first = ctrl.toggle(false);
		expect(ctrl.toggling).toBe(true);

		// Try second toggle while first is in progress
		const second = await ctrl.toggle(true);
		expect(second).toBeNull(); // re-entrance prevented

		// Resolve the first toggle
		resolveToggle(true);
		const firstResult = await first;
		expect(firstResult).toBe(true);
		expect(ctrl.toggling).toBe(false);
	});

	it('clears toggling flag even when toggleVideo throws', async () => {
		deps = makeDeps({
			toggleVideo: vi.fn(async () => { throw new Error('boom'); }),
		});

		const ctrl = createWebcamToggle(deps);

		await expect(ctrl.toggle(false)).rejects.toThrow('boom');
		expect(ctrl.toggling).toBe(false);
	});

	it('waits 100ms between sending command and calling toggleVideo', async () => {
		const start = Date.now();
		deps = makeDeps({
			toggleVideo: vi.fn(async () => {
				const elapsed = Date.now() - start;
				// Should be at least ~100ms since the setTimeout
				return elapsed >= 90;
			}),
		});

		const ctrl = createWebcamToggle(deps);

		const result = await ctrl.toggle(false);
		expect(result).toBe(true);
	});
});

describe('WebcamToggle — browser-initiated camera stop (onEnded)', () => {
	let deps: ReturnType<typeof makeDeps>;

	beforeEach(() => {
		vi.clearAllMocks();
		deps = makeDeps();
	});

	it('sends webcam_off WS command', () => {
		const ctrl = createWebcamToggle(deps);

		ctrl.onEnded();

		expect(deps.sendCommand).toHaveBeenCalledTimes(1);
		expect(deps.sendCommand).toHaveBeenCalledWith('webcam_off');
	});

	it('does not call toggleVideo (browser already killed the track)', () => {
		const ctrl = createWebcamToggle(deps);

		ctrl.onEnded();

		expect(deps.toggleVideo).not.toHaveBeenCalled();
	});

	it('does not set toggling flag (not a programmatic toggle)', () => {
		const ctrl = createWebcamToggle(deps);

		ctrl.onEnded();

		expect(ctrl.toggling).toBe(false);
	});

	it('works even without media client (just sends command)', () => {
		deps = makeDeps({ hasMediaClient: false });

		const ctrl = createWebcamToggle(deps);

		ctrl.onEnded();

		expect(deps.sendCommand).toHaveBeenCalledWith('webcam_off');
	});

	it('is idempotent — safe to call multiple times', () => {
		const ctrl = createWebcamToggle(deps);

		ctrl.onEnded();
		ctrl.onEnded();
		ctrl.onEnded();

		expect(deps.sendCommand).toHaveBeenCalledTimes(3);
		expect(deps.toggleVideo).not.toHaveBeenCalled();
	});
});

describe('WebcamToggle — integration scenarios', () => {
	it('full cycle: turn on via button → browser kills camera', async () => {
		const sentCommands: string[] = [];
		const deps: WebcamToggleDeps = {
			sendCommand: (cmd) => sentCommands.push(cmd),
			toggleVideo: vi.fn(async () => true),
			hasMediaClient: true,
		};

		const ctrl = createWebcamToggle(deps);

		// User clicks webcam button to turn on
		let webcamOn = false;
		const result = await ctrl.toggle(webcamOn);
		expect(result).toBe(true);
		webcamOn = result!;
		expect(webcamOn).toBe(true);

		// Browser revokes camera permission
		ctrl.onEnded();
		webcamOn = false;

		expect(sentCommands).toEqual(['webcam_on', 'webcam_off']);
		expect(deps.toggleVideo).toHaveBeenCalledTimes(1);
	});

	it('full cycle: turn on → turn off via button', async () => {
		const sentCommands: string[] = [];
		const deps: WebcamToggleDeps = {
			sendCommand: (cmd) => sentCommands.push(cmd),
			toggleVideo: vi.fn(async () => true),
			hasMediaClient: true,
		};

		const ctrl = createWebcamToggle(deps);

		// Turn on
		let webcamOn = false;
		let result = await ctrl.toggle(webcamOn);
		expect(result).toBe(true);
		webcamOn = result!;

		// Turn off
		result = await ctrl.toggle(webcamOn);
		expect(result).toBe(false);
		webcamOn = result!;
		expect(webcamOn).toBe(false);

		expect(sentCommands).toEqual(['webcam_on', 'webcam_off']);
		expect(deps.toggleVideo).toHaveBeenCalledTimes(2);
	});

	it('button toggle while browser-killed does not re-enable webcam', async () => {
		const sentCommands: string[] = [];
		const deps: WebcamToggleDeps = {
			sendCommand: (cmd) => sentCommands.push(cmd),
			toggleVideo: vi.fn(async () => true),
			hasMediaClient: true,
		};

		const ctrl = createWebcamToggle(deps);

		// Browser killed the camera
		ctrl.onEnded();
		let webcamOn = false;

		// User clicks to re-enable
		const result = await ctrl.toggle(webcamOn);
		webcamOn = result!;

		expect(sentCommands).toEqual(['webcam_off', 'webcam_on']);
		expect(webcamOn).toBe(true);
	});

	it('toggleVideo failure during browser-stop recovery is handled', async () => {
		const sentCommands: string[] = [];
		const deps: WebcamToggleDeps = {
			sendCommand: (cmd) => sentCommands.push(cmd),
			toggleVideo: vi.fn(async () => false), // fails
			hasMediaClient: true,
		};

		const ctrl = createWebcamToggle(deps);

		// Browser killed the camera
		ctrl.onEnded();
		let webcamOn = false;

		// User tries to re-enable — toggleVideo fails, rollback
		const result = await ctrl.toggle(webcamOn);
		expect(result).toBeNull();
		expect(sentCommands).toEqual(['webcam_off', 'webcam_on', 'webcam_off']);
	});
});
