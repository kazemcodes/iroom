const POPUP_WIDTH = 1100;
const POPUP_HEIGHT = 700;

interface OpenedPopup {
	window: Window;
	sessionId: string;
	interval: ReturnType<typeof setInterval>;
}

const popups = new Map<string, OpenedPopup>();

function getPopupFeatures(): string {
	const left = Math.round((screen.width - POPUP_WIDTH) / 2);
	const top = Math.round((screen.height - POPUP_HEIGHT) / 2);
	return `width=${POPUP_WIDTH},height=${POPUP_HEIGHT},left=${left},top=${top},popup=yes`;
}

function open(sessionId: string, title: string): Window | null {
	if (popups.has(sessionId)) {
		const existing = popups.get(sessionId)!;
		existing.window.focus();
		return existing.window;
	}

	const url = `/classroom/popup/${sessionId}`;
	const popup = window.open(url, `classroom-${sessionId}`, getPopupFeatures());

	if (!popup) {
		window.open(url, '_blank');
		return null;
	}

	if (title) {
		popup.document.title = title;
	}

	const interval = setInterval(() => {
		if (popup.closed) {
			const entry = popups.get(sessionId);
			if (entry) {
				clearInterval(entry.interval);
				popups.delete(sessionId);
			}
			window.dispatchEvent(
				new CustomEvent('classroom-closed', { detail: { sessionId } })
			);
		}
	}, 500);

	popups.set(sessionId, { window: popup, sessionId, interval });
	return popup;
}

function close(sessionId: string): void {
	const entry = popups.get(sessionId);
	if (entry) {
		clearInterval(entry.interval);
		entry.window.close();
		popups.delete(sessionId);
	}
}

function getAll(): ReadonlyMap<string, OpenedPopup> {
	return popups;
}

export const classroomWindow = { open, close, getAll };
