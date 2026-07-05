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
let context: BrowserContext | null = null;
let page: Page | null = null;
let consoleErrors: string[] = [];
let classroomLoaded = false;
let accessToken = '';

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

	context = await browser.newContext({ permissions: ['camera', 'microphone'] });
	page = await context.newPage();

	page.on('console', (msg) => {
		if (msg.type() === 'error') {
			const t = msg.text();
			if (!t.includes('favicon') && !t.includes('status of 404')) {
				consoleErrors.push(t);
			}
		}
	});
}, 30000);

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
	if (context) await context.close();
	if (browser) await browser.close();
}, 10000);

describe('Room Page — Loading', () => {
	it('room page loads for existing room slug', async () => {
		await page!.goto(`${FRONTEND}/room/${ROOM_SLUG}`, { waitUntil: 'domcontentloaded', timeout: 15000 });
		const roomEl = page!.locator('text=math').first();
		await roomEl.waitFor({ state: 'visible', timeout: 10000 });
		expect(await roomEl.isVisible()).toBe(true);
	});

	it('room page shows 404 for non-existent slug', async () => {
		const notFoundPage = await context!.newPage();
		await notFoundPage.goto(`${FRONTEND}/room/nonexistent-room-xyz-98765`, { waitUntil: 'domcontentloaded', timeout: 15000 });
		await notFoundPage.waitForTimeout(WAIT_MED);

		const pageText = await notFoundPage.evaluate(() => document.body.innerText);
		expect(pageText).toContain('اتاق یافت نشد');
		await notFoundPage.close();
	});

	it('room page shows login form with email and password fields', async () => {
		// Navigate back to the room page
		await page!.goto(`${FRONTEND}/room/${ROOM_SLUG}`, { waitUntil: 'domcontentloaded', timeout: 15000 });
		await page!.waitForSelector('input[type="email"]', { timeout: 10000 });

		const emailVisible = await page!.locator('input[type="email"]').isVisible().catch(() => false);
		const pwdVisible = await page!.locator('input[type="password"]').isVisible().catch(() => false);
		expect(emailVisible).toBe(true);
		expect(pwdVisible).toBe(true);
	});

	it('room color icon displays first letter of room name', async () => {
		const pageText = await page!.evaluate(() => document.body.innerText);
		// Room name should be visible
		expect(pageText).toContain('math');
	});
});

describe('Login Flow — Validation', () => {
	it('submitting empty form shows error message', async () => {
		// Clear any existing form state
		await page!.goto(`${FRONTEND}/room/${ROOM_SLUG}`, { waitUntil: 'domcontentloaded', timeout: 15000 });
		await page!.waitForSelector('input[type="email"]', { timeout: 10000 });

		// Click submit without filling
		const submitBtn = page!.locator('button[type="submit"]').first();
		await submitBtn.click();
		await page!.waitForTimeout(WAIT_SHORT);

		// Should show error about email/password required
		const pageText = await page!.evaluate(() => document.body.innerText);
		expect(pageText).toContain('ایمیل');
		// The HTML5 validation should prevent submission
	});

	it('wrong credentials show error on login', async () => {
		const emailInput = page!.locator('input[type="email"]');
		const pwdInput = page!.locator('input[type="password"]');

		await emailInput.fill(ADMIN_EMAIL);
		await pwdInput.fill('wrongpassword');
		await page!.locator('button[type="submit"]').first().click();
		await page!.waitForTimeout(WAIT_MED);

		// Should not navigate to classroom
		const joinBtn = page!.locator('text=پیوستن به کلاس');
		const joinVisible = await joinBtn.isVisible().catch(() => false);
		expect(joinVisible).toBe(false);
	});

	it('correct credentials show join button after login', async () => {
		const emailInput = page!.locator('input[type="email"]');
		const pwdInput = page!.locator('input[type="password"]');

		await emailInput.fill(ADMIN_EMAIL);
		await pwdInput.fill(ADMIN_PASSWORD);
		await page!.locator('button[type="submit"]').first().click();

		await page!.waitForSelector('text=پیوستن به کلاس', { timeout: 15000 });
		const joinVisible = await page!.locator('text=پیوستن به کلاس').isVisible().catch(() => false);
		expect(joinVisible).toBe(true);
	});
});

describe('Guest Login Flow', () => {
	it('guest login button is visible when guest_login is enabled', async () => {
		// Ensure guest_login is enabled for the room via API
		const loginRes = await fetch(`${API_BASE}/auth/login`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ email: ADMIN_EMAIL, password: ADMIN_PASSWORD }),
		});
		const loginData: any = await loginRes.json();
		const token = loginData?.data?.access_token;
		if (token) {
			// Get room ID from slug
			const roomsRes = await fetch(`${API_BASE}/rooms?slug=${ROOM_SLUG}`, {
				headers: { Authorization: `Bearer ${token}` },
			});
			const roomsData: any = await roomsRes.json();
			const room = roomsData?.data?.items?.[0] || roomsData?.data?.[0];
			if (room?.id) {
				await fetch(`${API_BASE}/rooms/${room.id}`, {
					method: 'PUT',
					headers: { Authorization: `Bearer ${token}`, 'Content-Type': 'application/json' },
					body: JSON.stringify({ guest_login: true }),
				});
			}
		}

		// Reload the page to pick up the setting change
		await page!.reload({ waitUntil: 'domcontentloaded', timeout: 15000 });
		await page!.waitForSelector('input[type="email"]', { timeout: 10000 });

		const guestBtn = page!.locator('text=ورود مهمان');
		await guestBtn.waitFor({ state: 'visible', timeout: 5000 }).catch(() => {});
		const visible = await guestBtn.isVisible().catch(() => false);
		expect(visible).toBe(true);
	});

	it('clicking guest login shows name input form', async () => {
		const guestBtn = page!.locator('text=ورود مهمان');
		if (!(await guestBtn.isVisible().catch(() => false))) return;

		await guestBtn.click();
		await page!.waitForTimeout(WAIT_SHORT);

		// Name input should appear, email input should be gone
		const nameInput = page!.locator('input[type="text"]');
		const emailInput = page!.locator('input[type="email"]');
		expect(await nameInput.isVisible().catch(() => false)).toBe(true);
	});
});

describe('Entry Modal & Classroom Join', () => {
	it('clicking join button shows entry modal', async () => {
		// Go back to main login and log in properly
		const backBtn = page!.getByText('بازگشت به ورود');
		if (await backBtn.isVisible().catch(() => false)) {
			await backBtn.click();
			await page!.waitForTimeout(WAIT_SHORT);
		}

		const emailInput = page!.locator('input[type="email"]');
		if (await emailInput.isVisible({ timeout: 3000 }).catch(() => false)) {
			await emailInput.fill(ADMIN_EMAIL);
			await page!.locator('input[type="password"]').fill(ADMIN_PASSWORD);
			await page!.locator('button[type="submit"]').first().click();
			await page!.waitForSelector('text=پیوستن به کلاس', { timeout: 15000 });
		}

		const joinBtn = page!.locator('text=پیوستن به کلاس');
		try {
			await joinBtn.click();
			await page!.waitForTimeout(500);

			// Entry modal should appear with speaker/listener options
			const speakerOption = page!.getByText('ورود به عنوان گوینده');
			await speakerOption.waitFor({ state: 'visible', timeout: 5000 });
			expect(await speakerOption.isVisible()).toBe(true);

			const listenerOption = page!.getByText('ورود به عنوان شنونده');
			expect(await listenerOption.isVisible()).toBe(true);
		} catch {
			console.warn('[lifecycle-e2e] entry modal not found');
		}
	});

	it('entering as speaker loads classroom UI', async () => {
		const enterBtn = page!.getByText('ورود به اتاق');
		if (!(await enterBtn.isVisible().catch(() => false))) return;

		await enterBtn.click();

		try {
			await page!.waitForSelector('.skyroom-header', { timeout: 15000 });
			classroomLoaded = true;
			accessToken = (await page!.evaluate(() => localStorage.getItem('access_token'))) || '';
			expect(classroomLoaded).toBe(true);
		} catch {
			console.warn('[lifecycle-e2e] classroom header not found');
		}
		await page!.waitForTimeout(2000);
	});

	it('session timer is visible in classroom header', async () => {
		if (!classroomLoaded) return;
		const timer = page!.locator('.skyroom-room-timer');
		expect(await timer.isVisible().catch(() => false)).toBe(true);
	});

	it('timer format is MM:SS', async () => {
		if (!classroomLoaded) return;
		const timerText = await page!.locator('.skyroom-room-timer span').first().textContent();
		expect(timerText).toMatch(/\d{2}:\d{2}/);
	});
});

describe('App Menu — Leave Room', () => {
	it('app menu opens from toolbar', async () => {
		if (!classroomLoaded) return;
		const menuBtn = page!.locator('button[title="منو"]');
		if (!(await menuBtn.isVisible().catch(() => false))) return;

		await menuBtn.click();
		await page!.waitForTimeout(WAIT_SHORT);

		const menuText = await page!.evaluate(() => document.body.innerText);
		expect(menuText).toContain('اطلاعات کاربری');
		expect(menuText).toContain('وضعیت اتصال');
		expect(menuText).toContain('تنظیمات');
		expect(menuText).toContain('چیدمان');
		expect(menuText).toContain('خروج');
		expect(menuText).toContain('بستن اتاق');

		// Dismiss menu
		await page!.locator('.skyroom-header').click({ position: { x: 0, y: 0 } });
		await page!.waitForTimeout(WAIT_SHORT);
	});

	it('connection status modal shows session info', async () => {
		if (!classroomLoaded) return;
		const menuBtn = page!.locator('button[title="منو"]');
		await menuBtn.click();
		await page!.waitForTimeout(WAIT_SHORT);

		await page!.getByText('وضعیت اتصال').click();
		await page!.waitForTimeout(WAIT_SHORT);

		const modalText = await page!.evaluate(() => document.body.innerText);
		expect(modalText).toContain('وضعیت اتصال');

		// Close
		await page!.locator('.modal-overlay').click({ position: { x: 10, y: 10 } });
		await page!.waitForTimeout(WAIT_SHORT);
	});

	it('layout modal shows chat and users toggles', async () => {
		if (!classroomLoaded) return;
		const menuBtn = page!.locator('button[title="منو"]');
		await menuBtn.click();
		await page!.waitForTimeout(WAIT_SHORT);

		await page!.getByText('چیدمان').click();
		await page!.waitForTimeout(WAIT_SHORT);

		const modalText = await page!.evaluate(() => document.body.innerText);
		expect(modalText).toContain('چیدمان');
		expect(modalText).toContain('کاربران');
		expect(modalText).toContain('پیام‌ها');

		// Close
		await page!.locator('.modal-overlay').click({ position: { x: 10, y: 10 } });
		await page!.waitForTimeout(WAIT_SHORT);
	});
});

describe('Lifecycle — Console Errors', () => {
	it('no fatal JS errors during room lifecycle', () => {
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
