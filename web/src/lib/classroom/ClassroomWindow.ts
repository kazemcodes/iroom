/**
 * ClassroomWindow — Manages classroom popup windows.
 *
 * Opens classroom sessions in separate browser windows (like Skyroom).
 * Tracks open windows and detects when they close.
 *
 * Usage:
 *   import { classroomWindow } from '$lib/classroom/ClassroomWindow';
 *   classroomWindow.open('session-123', 'Math Class');
 *   classroomWindow.close('session-123');
 *   const openTabs = classroomWindow.getAll(); // Map of open sessions
 *
 * Events:
 *   Dispatches 'classroom-closed' CustomEvent when a popup window closes.
 */
interface OpenedTab {
	window: Window;
	sessionId: string;
	interval: ReturnType<typeof setInterval>;
}

const tabs = new Map<string, OpenedTab>();

function open(sessionId: string, title: string): Window | null {
	if (tabs.has(sessionId)) {
		const existing = tabs.get(sessionId)!;
		existing.window.focus();
		return existing.window;
	}

	const url = `/classroom/popup/${sessionId}`;
	const tab = window.open(url, `_blank`);

	if (!tab) {
		return null;
	}

	const interval = setInterval(() => {
		if (tab.closed) {
			const entry = tabs.get(sessionId);
			if (entry) {
				clearInterval(entry.interval);
				tabs.delete(sessionId);
			}
			window.dispatchEvent(
				new CustomEvent('classroom-closed', { detail: { sessionId } })
			);
		}
	}, 500);

	tabs.set(sessionId, { window: tab, sessionId, interval });
	return tab;
}

function close(sessionId: string): void {
	const entry = tabs.get(sessionId);
	if (entry) {
		clearInterval(entry.interval);
		entry.window.close();
		tabs.delete(sessionId);
	}
}

function getAll(): ReadonlyMap<string, OpenedTab> {
	return tabs;
}

export const classroomWindow = { open, close, getAll };
