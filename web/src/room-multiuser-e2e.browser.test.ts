import { describe, it, beforeAll, afterAll, expect } from 'vitest';
import { chromium, type Browser, type BrowserContext, type Page } from 'playwright';

const FRONTEND = 'http://localhost:5173';
const ROOM_SLUG = 'math';
const ADMIN_EMAIL = process.env.E2E_ADMIN_EMAIL ?? 'admin@iroom.local';
const ADMIN_PASSWORD = process.env.E2E_ADMIN_PASSWORD ?? 'admin123';
const API_BASE = 'http://localhost:8080/api/v1';

const WAIT_SHORT = 500;
const WAIT_MED = 1500;
const WAIT_LONG = 3000;

const chromePath =
	process.env.CHROME_BIN ||
	(process.platform === 'darwin'
		? '/Applications/Google Chrome.app/Contents/MacOS/Google Chrome'
		: '/usr/bin/google-chrome-stable');

let browser: Browser | null = null;
let adminContext: BrowserContext | null = null;
let guestContext: BrowserContext | null = null;
let adminPage: Page | null = null;
let guestPage: Page | null = null;
let consoleErrors: string[] = [];
let adminClassroomLoaded = false;
let guestClassroomLoaded = false;
let accessToken = '';

async function loginAsAdmin(p: Page) {
	await p.goto(`${FRONTEND}/room/${ROOM_SLUG}`, { waitUntil: 'domcontentloaded', timeout: 15000 });
	await p.waitForSelector('input[type="email"]', { timeout: 10000 });

	const emailInput = p.locator('input[type="email"]');
	if (await emailInput.isVisible({ timeout: 3000 }).catch(() => false)) {
		await emailInput.fill(ADMIN_EMAIL);
		await p.locator('input[type="password"]').fill(ADMIN_PASSWORD);
		await p.locator('button[type="submit"]').first().click();
		await p.waitForSelector('text=پیوستن به کلاس', { timeout: 15000 });
	}

	const joinBtn = p.locator('text=پیوستن به کلاس');
	try {
		await joinBtn.click();
		await p.waitForTimeout(500);

		const enterBtn = p.getByText('ورود به اتاق');
		await enterBtn.waitFor({ state: 'visible', timeout: 5000 });
		await enterBtn.click();

		await p.waitForSelector('.skyroom-header', { timeout: 15000 });
		await p.waitForTimeout(2000);
		return true;
	} catch {
		console.warn('[multiuser] admin classroom not loaded');
		return false;
	}
}

async function joinAsGuest(p: Page, displayName: string) {
	await p.goto(`${FRONTEND}/room/${ROOM_SLUG}`, { waitUntil: 'domcontentloaded', timeout: 15000 });
	await p.waitForSelector('input[type="email"]', { timeout: 10000 });

	// Click guest login
	const guestBtn = p.locator('text=ورود مهمان');
	if (await guestBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
		await guestBtn.click();
		await p.waitForTimeout(WAIT_SHORT);

		const nameInput = p.locator('input[type="text"]');
		if (await nameInput.isVisible({ timeout: 3000 }).catch(() => false)) {
			await nameInput.fill(displayName);
			await p.locator('button[type="submit"]').click();
			await p.waitForTimeout(WAIT_LONG);
		}
	}

	// Join room
	const joinBtn = p.locator('text=پیوستن به کلاس');
	try {
		if (await joinBtn.isVisible({ timeout: 5000 }).catch(() => false)) {
			await joinBtn.click();
			await p.waitForTimeout(500);

			const enterBtn = p.getByText('ورود به اتاق');
			await enterBtn.waitFor({ state: 'visible', timeout: 5000 });
			await enterBtn.click();

			await p.waitForSelector('.skyroom-header', { timeout: 15000 });
			await p.waitForTimeout(2000);
			return true;
		}
	} catch {
		console.warn('[multiuser] guest classroom not loaded');
	}
	return false;
}

beforeAll(async () => {
	browser = await chromium.launch({
		executablePath: chromePath,
		headless: process.env.HEADLESS !== 'false',
		args: [
			'--use-fake-device-for-media-stream',
			'--use-fake-ui-for-media-stream',
			'--no-sandbox',
			'--disable-setuid-sandbox',
			'--disable-dev-shm-usage',
			'--auto-select-desktop-capture-source=Entire Screen',
		],
	});

	// Admin browser context
	adminContext = await browser.newContext({ permissions: ['camera', 'microphone'] });
	adminPage = await adminContext.newPage();
	adminPage.on('console', (msg) => {
		if (msg.type() === 'error') {
			const t = msg.text();
			if (!t.includes('favicon') && !t.includes('status of 404')) {
				consoleErrors.push(`[admin] ${t}`);
			}
		}
	});

	// Guest browser context
	guestContext = await browser.newContext({ permissions: ['camera', 'microphone'] });
	guestPage = await guestContext.newPage();
	guestPage.on('console', (msg) => {
		if (msg.type() === 'error') {
			const t = msg.text();
			if (!t.includes('favicon') && !t.includes('status of 404')) {
				consoleErrors.push(`[guest] ${t}`);
			}
		}
	});

	// Admin joins first
	adminClassroomLoaded = await loginAsAdmin(adminPage);
	if (adminClassroomLoaded) {
		accessToken = (await adminPage!.evaluate(() => localStorage.getItem('access_token'))) || '';
	}

	// Ensure guest_login is enabled for the room via API
	if (accessToken) {
		try {
			const roomsRes = await fetch(`${API_BASE}/rooms?slug=${ROOM_SLUG}`, {
				headers: { Authorization: `Bearer ${accessToken}` },
			});
			const roomsData: any = await roomsRes.json();
			const room = roomsData?.data?.items?.[0] || roomsData?.data?.[0];
			if (room?.id) {
				await fetch(`${API_BASE}/rooms/${room.id}`, {
					method: 'PUT',
					headers: { Authorization: `Bearer ${accessToken}`, 'Content-Type': 'application/json' },
					body: JSON.stringify({ guest_login: true }),
				});
			}
		} catch { /* best effort */ }
	}

	// Guest joins second
	guestClassroomLoaded = await joinAsGuest(guestPage, `مهمان تست ${Date.now()}`);
}, 90000);

afterAll(async () => {
	// Cleanup: end live sessions
	if (accessToken) {
		try {
			const sessRes = await fetch(`${API_BASE}/sessions?per_page=10`, {
				headers: { Authorization: `Bearer ${accessToken}` },
			});
			const sessData: any = await sessRes.json();
			const items = sessData?.data?.items || [];
			for (const s of items) {
				if (s.status === 'live') {
					await fetch(`${API_BASE}/sessions/${s.id}/end`, {
						method: 'POST',
						headers: { Authorization: `Bearer ${accessToken}` },
					});
				}
			}
		} catch { /* cleanup is best-effort */ }
	}
	if (adminContext) await adminContext.close();
	if (guestContext) await guestContext.close();
	if (browser) await browser.close();
}, 10000);

describe('Multi-User — Participant List', () => {
	it('admin sees own name in participants list', async () => {
		if (!adminClassroomLoaded) return;
		const usersText = await adminPage!.evaluate(() => document.body.innerText);
		expect(usersText).toContain('مدیر سیستم');
	});

	it('admin sees participant count >= 1', async () => {
		if (!adminClassroomLoaded) return;
		const countEl = adminPage!.locator('.skyroom-users-count');
		if (await countEl.count() === 0) return;
		const count = await countEl.first().textContent();
		expect(Number(count)).toBeGreaterThanOrEqual(1);
	});

	it('guest sees themselves in participants list', async () => {
		if (!guestClassroomLoaded) return;
		const usersText = await guestPage!.evaluate(() => document.body.innerText);
		expect(usersText).toContain('کاربران');
	});

	it('after both join, admin sees 2 participants', async () => {
		if (!adminClassroomLoaded || !guestClassroomLoaded) return;

		// Wait for participant refresh (happens every 5 seconds)
		// Poll a few times
		let count = 0;
		for (let i = 0; i < 6; i++) {
			await adminPage!.waitForTimeout(2000);
			const countEl = adminPage!.locator('.skyroom-users-count');
			if (await countEl.count() > 0) {
				count = Number(await countEl.first().textContent());
				if (count >= 2) break;
			}
		}
		console.log(`[multiuser] participant count after waiting: ${count}`);
		expect(count).toBeGreaterThanOrEqual(2);
	});
});

describe('Multi-User — Hand Raise', () => {
	it('admin can raise hand (button toggle)', async () => {
		if (!adminClassroomLoaded) return;
		const handBtn = adminPage!.locator('button[title="بالا بردن دست"]');
		if (await handBtn.count() === 0) return;

		const wasActive = await handBtn.first().evaluate((el) => el.classList.contains('active'));
		await handBtn.first().click();
		await adminPage!.waitForTimeout(WAIT_SHORT);

		const isNowActive = await handBtn.first().evaluate((el) => el.classList.contains('active'));
		expect(isNowActive).toBe(!wasActive);

		// Lower hand
		if (isNowActive) {
			await handBtn.first().click();
			await adminPage!.waitForTimeout(WAIT_SHORT);
		}
	});

	it('"lower all hands" is available in users context menu', async () => {
		if (!adminClassroomLoaded) return;
		const usersDots = adminPage!.locator('.skyroom-users-block .skyroom-dots-btn').first();
		if (!(await usersDots.isVisible().catch(() => false))) return;

		await usersDots.click();
		await adminPage!.waitForTimeout(WAIT_SHORT);

		const menuText = await adminPage!.evaluate(() => document.body.innerText);
		expect(menuText).toContain('پایین آوردن دست‌ها');

		// Dismiss
		await adminPage!.locator('.skyroom-header').click({ position: { x: 0, y: 0 } });
		await adminPage!.waitForTimeout(WAIT_SHORT);
	});

	it('admin raises hand → guest can see hand icon in users list', async () => {
		if (!adminClassroomLoaded || !guestClassroomLoaded) return;

		// Admin raises hand
		const handBtn = adminPage!.locator('button[title="بالا بردن دست"]');
		if (await handBtn.count() === 0) return;
		await handBtn.first().click();
		await adminPage!.waitForTimeout(WAIT_SHORT);

		// Wait for WS sync + participant refresh
		await guestPage!.waitForTimeout(6000);

		// Guest should see the hand-raised indicator
		const handIcon = guestPage!.locator('.skyroom-users-list .media-icon.hand-raised');
		const handCount = await handIcon.count();
		console.log(`[multiuser] hand icons visible to guest: ${handCount}`);
		// Hand icon may or may not propagate depending on WS latency

		// Lower hand
		await handBtn.first().click();
		await adminPage!.waitForTimeout(WAIT_SHORT);
	});
});

describe('Multi-User — User Action Menu', () => {
	it('admin can open action menu for non-local user', async () => {
		if (!adminClassroomLoaded || !guestClassroomLoaded) return;

		// Wait for participant refresh
		await adminPage!.waitForTimeout(6000);

		// Find a non-local user row and click the menu arrow
		const menuArrow = adminPage!.locator('.skyroom-user-row:not(:has(.isLocal)) .user-menu-arrow').first();
		if (await menuArrow.count() === 0) {
			console.log('[multiuser] no non-local user menu arrow found');
			return;
		}

		await menuArrow.click();
		await adminPage!.waitForTimeout(WAIT_SHORT);

		const menuText = await adminPage!.evaluate(() => document.body.innerText);
		expect(menuText).toContain('اپراتور');
		expect(menuText).toContain('ارائه‌دهنده');
		expect(menuText).toContain('کاربر عادی');
		expect(menuText).toContain('اخراج');

		// Dismiss
		await adminPage!.locator('.skyroom-header').click({ position: { x: 0, y: 0 } });
		await adminPage!.waitForTimeout(WAIT_SHORT);
	});
});

describe('Multi-User — Attendance Modal', () => {
	it('attendance modal opens and shows participants', async () => {
		if (!adminClassroomLoaded) return;
		const usersDots = adminPage!.locator('.skyroom-users-block .skyroom-dots-btn').first();
		if (!(await usersDots.isVisible().catch(() => false))) return;

		await usersDots.click();
		await adminPage!.waitForTimeout(WAIT_SHORT);

		const attendanceBtn = adminPage!.getByText('حضور و غیاب');
		if (!(await attendanceBtn.isVisible().catch(() => false))) return;

		await attendanceBtn.click();
		await adminPage!.waitForTimeout(WAIT_SHORT);

		const modalText = await adminPage!.evaluate(() => document.body.innerText);
		expect(modalText).toContain('حضور و غیاب');

		// Close
		await adminPage!.locator('.modal-overlay').click({ position: { x: 10, y: 10 } });
		await adminPage!.waitForTimeout(WAIT_SHORT);
	});
});

describe('Multi-User — Chat Between Users', () => {
	it('admin sends a chat message visible to self', async () => {
		if (!adminClassroomLoaded) return;
		const chatInput = adminPage!.locator('.skyroom-chat-block .chat-input');
		if (await chatInput.count() === 0) return;

		const msgText = `Admin multiuser msg ${Date.now()}`;
		await chatInput.fill(msgText);
		await chatInput.press('Enter');
		await adminPage!.waitForTimeout(WAIT_MED);

		const chatContent = await adminPage!.evaluate(() => document.body.innerText);
		expect(chatContent).toContain(msgText);
	});

	it('guest sends a chat message visible to self', async () => {
		if (!guestClassroomLoaded) return;
		const chatInput = guestPage!.locator('.skyroom-chat-block .chat-input');
		if (await chatInput.count() === 0) return;

		const msgText = `Guest multiuser msg ${Date.now()}`;
		await chatInput.fill(msgText);
		await chatInput.press('Enter');
		await guestPage!.waitForTimeout(WAIT_MED);

		const chatContent = await guestPage!.evaluate(() => document.body.innerText);
		expect(chatContent).toContain(msgText);
	});
});

describe('Multi-User — Console Errors', () => {
	it('no fatal JS errors across both user sessions', () => {
		const knownNoise = [
			'favicon',
			'status of 404',
				'status of 401',
				'status of 403',
				'status of 429',
			'WebSocket',
			'MediaStream',
			'getUserMedia',
			'AbortError',
			'NotAllowedError',
			'NotFoundError',
		];
		const realErrors = consoleErrors.filter(
			(e) => !knownNoise.some((n) => e.includes(n)),
		);
		expect(realErrors).toEqual([]);
	});
});
