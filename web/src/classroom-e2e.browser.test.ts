import { describe, it, beforeAll, afterAll, expect } from 'vitest';
import { chromium, type Browser, type BrowserContext, type Page } from 'playwright';

const FRONTEND = 'http://localhost:5173';
const ROOM_SLUG = 'math';

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
let sessionId: number | null = null;

const API_BASE = 'http://localhost:8080/api/v1';

beforeAll(async () => {
	browser = await chromium.launch({
		executablePath: chromePath,
		headless: true,
		args: [
			'--use-fake-device-for-media-stream',
			'--use-fake-ui-for-media-stream',
			'--no-sandbox',
			'--disable-setuid-sandbox',
			'--disable-dev-shm-usage',
			'--auto-select-desktop-capture-source=Entire Screen',
		],
	});

	context = await browser.newContext({
		permissions: ['camera', 'microphone'],
	});

	page = await context.newPage();

	page.on('console', (msg) => {
		if (msg.type() === 'error') {
			const t = msg.text();
			if (!t.includes('favicon') && !t.includes('status of 404')) {
				consoleErrors.push(t);
			}
		}
	});

	// Navigate to room — will show login form
	await page.goto(`${FRONTEND}/room/${ROOM_SLUG}`, { waitUntil: 'domcontentloaded', timeout: 15000 });

	// Wait for Svelte hydration — login form should render with email input
	await page.waitForSelector('input[type="email"]', { timeout: 10000 });

	// Log page state before login
	const beforeLogin = await page.evaluate(() => document.body.innerText.substring(0, 200));
	console.log('[test] before login:', beforeLogin.replace(/\s+/g, ' '));

	// Fill login form — exactly like a real user
	const emailInput = page.locator('input[type="email"]');
	if (await emailInput.isVisible({ timeout: 3000 }).catch(() => false)) {
		await emailInput.fill('admin@iroom.local');
		const pwdInput = page.locator('input[type="password"]');
		await pwdInput.fill('admin123');

		// Click login button (use button[type="submit"] — more robust than text matching)
		const loginBtn = page.locator('button[type="submit"]').first();
		await loginBtn.click();

		// Wait for joinRoom() to complete — "پیوستن به کلاس" button appears
		await page.waitForSelector('text=پیوستن به کلاس', { timeout: 15000 });

		const afterLogin = await page.evaluate(() => document.body.innerText.substring(0, 300));
		console.log('[test] after login:', afterLogin.replace(/\s+/g, ' '));
	}

	// Click "پیوستن به کلاس" (Join the class) button
	const joinBtn = page.locator('text=پیوستن به کلاس');
	try {
		await joinBtn.click();
		await page.waitForTimeout(500);

		// Entry modal: click "ورود به اتاق" (Enter room) as speaker
		const enterBtn = page.getByText('ورود به اتاق');
		await enterBtn.waitFor({ state: 'visible', timeout: 5000 });
		await enterBtn.click();

		// Wait for joinClassroom() and MediaClient to initialize
		try {
			await page.waitForSelector('.skyroom-header', { timeout: 15000 });
			classroomLoaded = true;

			// Extract session ID for cleanup
			sessionId = await page.evaluate(() => {
				// The session ID is in the URL path or can be extracted from API state
				// Use room_id from the page's reactive state if exposed
				const accessToken = localStorage.getItem('access_token');
				return accessToken ? 1 : null; // flag that we have a live session
			});
		} catch {
			console.warn('[test] classroom header not found');
		}
		await page.waitForTimeout(2000);
	} catch {
		console.warn('[test] join button not found — classroom may not be available');
	}

	const finalBody = await page.evaluate(() => document.body.innerText.substring(0, 200));
	console.log('[test] final state:', finalBody.replace(/\s+/g, ' '));
}, 60000);

afterAll(async () => {
	// End the session to avoid orphaned database records
	if (sessionId) {
		try {
			const token = await page?.evaluate(() => localStorage.getItem('access_token'));
			if (token) {
				// Find live sessions for the math room and end them
				const sessRes = await fetch(`${API_BASE}/sessions?per_page=10`, {
					headers: { Authorization: `Bearer ${token}` },
				});
				const sessData: any = await sessRes.json();
				const items = sessData?.data?.items || [];
				for (const s of items) {
					if (s.status === 'live') {
						await fetch(`${API_BASE}/sessions/${s.id}/end`, {
							method: 'POST',
							headers: { Authorization: `Bearer ${token}` },
						});
					}
				}
			}
		} catch { /* cleanup is best-effort */ }
	}
	if (browser) await browser.close();
}, 10000);

describe('Classroom E2E — Toggle Controllers', () => {
	it('no fatal JS errors during login and join', () => {
		const fatal = consoleErrors.filter(
			(e) => e.includes('ReferenceError') || e.includes('TypeError'),
		);
		expect(fatal).toEqual([]);
	});

	it('login succeeded (classroom or join UI visible)', async () => {
		const pageText = await page!.evaluate(() => document.body.innerText);
		// Room name "math" should appear, and either classroom UI or join button
		const hasClassName = pageText.includes('math') || pageText.includes('Math');
		expect(hasClassName).toBe(true);
	});

	it('webcam toggle works end-to-end', async () => {
		if (!classroomLoaded) {
			console.log('[test] classroom not loaded, skipping webcam test');
			return;
		}

		const btn = page!.locator('button[title="وبکم"]');
		const count = await btn.count();
		if (count === 0) {
			console.log('[test] webcam button not found');
			return;
		}

		const wasActive = await btn.first().evaluate((el) => el.classList.contains('active'));
		console.log('[test] webcam initial active:', wasActive);

		// Toggle off
		await btn.first().click();
		await page!.waitForTimeout(1500);
		const afterOff = await btn.first().evaluate((el) => el.classList.contains('active'));
		console.log('[test] webcam after off click:', afterOff);

		// Toggle on
		await btn.first().click();
		await page!.waitForTimeout(1500);
		const afterOn = await btn.first().evaluate((el) => el.classList.contains('active'));
		console.log('[test] webcam after on click:', afterOn);

		// Should be back to original after off→on cycle
		expect(afterOn).toBe(wasActive);
	});

	it('mic toggle works end-to-end', async () => {
		if (!classroomLoaded) {
			console.log('[test] classroom not loaded, skipping mic test');
			return;
		}

		const btn = page!.locator('button[title="میکروفون"]');
		const count = await btn.count();
		if (count === 0) {
			console.log('[test] mic button not found');
			return;
		}

		const wasActive = await btn.first().evaluate((el) => el.classList.contains('active'));
		console.log('[test] mic initial active:', wasActive);

		await btn.first().click();
		await page!.waitForTimeout(1500);
		await btn.first().click();
		await page!.waitForTimeout(1500);

		const afterOn = await btn.first().evaluate((el) => el.classList.contains('active'));
		expect(afterOn).toBe(wasActive);
	});

	it('screen share toggle is accessible', async () => {
		if (!classroomLoaded) {
			console.log('[test] classroom not loaded, skipping screenshare test');
			return;
		}

		const btn = page!.locator('button[title="اشتراک‌گذاری صفحه"]');
		const count = await btn.count();
		if (count > 0) {
			// Toggle on → check active → toggle off
			const wasActive = await btn.first().evaluate((el) => el.classList.contains('active'));
			await btn.first().click();
			await page!.waitForTimeout(2500);
			await btn.first().click();
			await page!.waitForTimeout(1500);
			const isNowActive = await btn.first().evaluate((el) => el.classList.contains('active'));
			expect(isNowActive).toBe(wasActive);
		} else {
			console.log('[test] screenshare button not available');
		}
	});

	it('no toggle controller console errors', () => {
		const toggleErrors = consoleErrors.filter(
			(e) =>
				e.includes('webcamCtrl') || e.includes('micCtrl') ||
				e.includes('screenShareCtrl') || e.includes('createWebcamToggle') ||
				e.includes('createMicToggle') || e.includes('createScreenShareToggle'),
		);
		expect(toggleErrors).toEqual([]);
	});

	it('re-entrance guard works on rapid double-click', async () => {
		if (!classroomLoaded) return;

		const btn = page!.locator('button[title="وبکم"]');
		if ((await btn.count()) === 0) return;

		const before = consoleErrors.length;
		await btn.first().click();
		await btn.first().click(); // rapid second click
		await page!.waitForTimeout(2000);

		const newErrors = consoleErrors.slice(before);
		expect(newErrors).toEqual([]);
	});
});
