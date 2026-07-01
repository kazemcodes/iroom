import { describe, it, expect, vi, beforeEach } from 'vitest';
import { createMicToggle, type MicToggleDeps } from '$lib/classroom/mic-toggle.svelte';

/**
 * Integration-style tests for the mic toggle logic.
 *
 * Covers both UI button click paths and browser-initiated microphone stop paths,
 * as well as edge cases like permission guards, re-entrance prevention,
 * and command rollback on failure.
 */

function makeDeps(overrides: Partial<MicToggleDeps> = {}): MicToggleDeps {
	return {
		sendCommand: vi.fn(),
		toggleAudio: vi.fn(async () => true),
		hasMediaClient: true,
		...overrides,
	};
}

describe('MicToggle — UI button clicks (toggle)', () => {
	let deps: ReturnType<typeof makeDeps>;

	beforeEach(() => {
		vi.clearAllMocks();
		deps = makeDeps();
	});

	it('sends mic_on and calls toggleAudio when turning on', async () => {
		const ctrl = createMicToggle(deps);

		const result = await ctrl.toggle(false);

		expect(result).toBe(true);
		expect(deps.sendCommand).toHaveBeenNthCalledWith(1, 'mic_on');
		expect(deps.toggleAudio).toHaveBeenCalledTimes(1);
	});

	it('sends mic_off and calls toggleAudio when turning off', async () => {
		const ctrl = createMicToggle(deps);

		const result = await ctrl.toggle(true);

		expect(result).toBe(false);
		expect(deps.sendCommand).toHaveBeenNthCalledWith(1, 'mic_off');
		expect(deps.toggleAudio).toHaveBeenCalledTimes(1);
	});

	it('toggles without media client (just sends command and returns new state)', async () => {
		deps = makeDeps({ hasMediaClient: false });

		const ctrl = createMicToggle(deps);

		const result = await ctrl.toggle(false);

		expect(result).toBe(true);
		expect(deps.sendCommand).toHaveBeenCalledWith('mic_on');
		expect(deps.toggleAudio).not.toHaveBeenCalled();
	});

	it('toggles without media client from on to off', async () => {
		deps = makeDeps({ hasMediaClient: false });

		const ctrl = createMicToggle(deps);

		const result = await ctrl.toggle(true);

		expect(result).toBe(false);
		expect(deps.sendCommand).toHaveBeenCalledWith('mic_off');
	});

	it('rolls back WS command when toggleAudio fails', async () => {
		deps = makeDeps({ toggleAudio: vi.fn(async () => false) });

		const ctrl = createMicToggle(deps);

		const result = await ctrl.toggle(false);

		expect(result).toBeNull();
		// First: mic_on was sent
		expect(deps.sendCommand).toHaveBeenNthCalledWith(1, 'mic_on');
		// Then: rollback to mic_off
		expect(deps.sendCommand).toHaveBeenNthCalledWith(2, 'mic_off');
		expect(deps.toggleAudio).toHaveBeenCalledTimes(1);
	});

	it('rolls back WS command when toggleAudio fails (off→on attempt)', async () => {
		deps = makeDeps({ toggleAudio: vi.fn(async () => false) });

		const ctrl = createMicToggle(deps);

		const result = await ctrl.toggle(true);

		expect(result).toBeNull();
		expect(deps.sendCommand).toHaveBeenNthCalledWith(1, 'mic_off');
		expect(deps.sendCommand).toHaveBeenNthCalledWith(2, 'mic_on');
	});

	it('prevents re-entrance (toggling guard)', async () => {
		// toggleAudio returns a promise that we control
		let resolveToggle!: (value: boolean) => void;
		const toggleDeferred = new Promise<boolean>((resolve) => {
			resolveToggle = resolve;
		});
		deps = makeDeps({
			toggleAudio: vi.fn(() => toggleDeferred),
		});

		const ctrl = createMicToggle(deps);

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

	it('clears toggling flag even when toggleAudio throws', async () => {
		deps = makeDeps({
			toggleAudio: vi.fn(async () => { throw new Error('boom'); }),
		});

		const ctrl = createMicToggle(deps);

		await expect(ctrl.toggle(false)).rejects.toThrow('boom');
		expect(ctrl.toggling).toBe(false);
	});

	it('waits 100ms between sending command and calling toggleAudio', async () => {
		const start = Date.now();
		deps = makeDeps({
			toggleAudio: vi.fn(async () => {
				const elapsed = Date.now() - start;
				// Should be at least ~100ms since the setTimeout
				return elapsed >= 90;
			}),
		});

		const ctrl = createMicToggle(deps);

		const result = await ctrl.toggle(false);
		expect(result).toBe(true);
	});
});

describe('MicToggle — browser-initiated microphone stop (onEnded)', () => {
	let deps: ReturnType<typeof makeDeps>;

	beforeEach(() => {
		vi.clearAllMocks();
		deps = makeDeps();
	});

	it('sends mic_off WS command', () => {
		const ctrl = createMicToggle(deps);

		ctrl.onEnded();

		expect(deps.sendCommand).toHaveBeenCalledTimes(1);
		expect(deps.sendCommand).toHaveBeenCalledWith('mic_off');
	});

	it('does not call toggleAudio (browser already killed the track)', () => {
		const ctrl = createMicToggle(deps);

		ctrl.onEnded();

		expect(deps.toggleAudio).not.toHaveBeenCalled();
	});

	it('does not set toggling flag (not a programmatic toggle)', () => {
		const ctrl = createMicToggle(deps);

		ctrl.onEnded();

		expect(ctrl.toggling).toBe(false);
	});

	it('works even without media client (just sends command)', () => {
		deps = makeDeps({ hasMediaClient: false });

		const ctrl = createMicToggle(deps);

		ctrl.onEnded();

		expect(deps.sendCommand).toHaveBeenCalledWith('mic_off');
	});

	it('is idempotent — safe to call multiple times', () => {
		const ctrl = createMicToggle(deps);

		ctrl.onEnded();
		ctrl.onEnded();
		ctrl.onEnded();

		expect(deps.sendCommand).toHaveBeenCalledTimes(3);
		expect(deps.toggleAudio).not.toHaveBeenCalled();
	});
});

describe('MicToggle — integration scenarios', () => {
	it('full cycle: turn on via button → browser kills mic', async () => {
		const sentCommands: string[] = [];
		const deps: MicToggleDeps = {
			sendCommand: (cmd) => sentCommands.push(cmd),
			toggleAudio: vi.fn(async () => true),
			hasMediaClient: true,
		};

		const ctrl = createMicToggle(deps);

		// User clicks mic button to turn on
		let micOn = false;
		const result = await ctrl.toggle(micOn);
		expect(result).toBe(true);
		micOn = result!;
		expect(micOn).toBe(true);

		// Browser revokes microphone permission
		ctrl.onEnded();
		micOn = false;

		expect(sentCommands).toEqual(['mic_on', 'mic_off']);
		expect(deps.toggleAudio).toHaveBeenCalledTimes(1);
	});

	it('full cycle: turn on → turn off via button', async () => {
		const sentCommands: string[] = [];
		const deps: MicToggleDeps = {
			sendCommand: (cmd) => sentCommands.push(cmd),
			toggleAudio: vi.fn(async () => true),
			hasMediaClient: true,
		};

		const ctrl = createMicToggle(deps);

		// Turn on
		let micOn = false;
		let result = await ctrl.toggle(micOn);
		expect(result).toBe(true);
		micOn = result!;

		// Turn off
		result = await ctrl.toggle(micOn);
		expect(result).toBe(false);
		micOn = result!;
		expect(micOn).toBe(false);

		expect(sentCommands).toEqual(['mic_on', 'mic_off']);
		expect(deps.toggleAudio).toHaveBeenCalledTimes(2);
	});

	it('button toggle while browser-killed re-enables mic', async () => {
		const sentCommands: string[] = [];
		const deps: MicToggleDeps = {
			sendCommand: (cmd) => sentCommands.push(cmd),
			toggleAudio: vi.fn(async () => true),
			hasMediaClient: true,
		};

		const ctrl = createMicToggle(deps);

		// Browser killed the microphone
		ctrl.onEnded();
		let micOn = false;

		// User clicks to re-enable
		const result = await ctrl.toggle(micOn);
		micOn = result!;

		expect(sentCommands).toEqual(['mic_off', 'mic_on']);
		expect(micOn).toBe(true);
	});

	it('toggleAudio failure during browser-stop recovery is handled', async () => {
		const sentCommands: string[] = [];
		const deps: MicToggleDeps = {
			sendCommand: (cmd) => sentCommands.push(cmd),
			toggleAudio: vi.fn(async () => false), // fails
			hasMediaClient: true,
		};

		const ctrl = createMicToggle(deps);

		// Browser killed the microphone
		ctrl.onEnded();
		let micOn = false;

		// User tries to re-enable — toggleAudio fails, rollback
		const result = await ctrl.toggle(micOn);
		expect(result).toBeNull();
		expect(sentCommands).toEqual(['mic_off', 'mic_on', 'mic_off']);
	});
});
