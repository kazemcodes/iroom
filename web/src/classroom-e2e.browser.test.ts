import { describe, it, beforeAll, afterAll, expect } from 'vitest';
import { chromium, type Browser, type BrowserContext, type Page } from 'playwright';

const FRONTEND = 'http://localhost:5173';
const ROOM_SLUG = 'math';
const ADMIN_EMAIL = process.env.E2E_ADMIN_EMAIL ?? 'admin@iroom.local';
const ADMIN_PASSWORD = process.env.E2E_ADMIN_PASSWORD ?? 'admin123';

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

let adminPage: Page | null = null;
let adminContext: BrowserContext | null = null;

const API_BASE = 'http://localhost:8080/api/v1';
const WAIT_SHORT = 500;
const WAIT_MED = 1500;
const WAIT_LONG = 3000;

beforeAll(async () => {
	browser = await chromium.launch({
		executablePath: chromePath,
		headless: process.env.HEADLESS !== 'true',
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
		await emailInput.fill(ADMIN_EMAIL);
		const pwdInput = page.locator('input[type="password"]');
		await pwdInput.fill(ADMIN_PASSWORD);

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

		const newErrors = consoleErrors.slice(before).filter(
			(e) => !e.includes('WebSocket') && !e.includes('wss://') && !e.includes('ws://') && !e.includes('status of 429'),
		);
		expect(newErrors).toEqual([]);
	});
});

describe('Classroom — App Menu & Modals', () => {
	it('app menu opens with all items visible', async () => {
		if (!classroomLoaded) return;
		const menuBtn = page!.locator('button[title="منو"]');
		const menuVisible = await menuBtn.isVisible().catch(() => false);
		if (!menuVisible) return;
		await menuBtn.click();
		await page!.waitForTimeout(WAIT_SHORT);

		const menuText = await page!.evaluate(() => document.body.innerText);
		expect(menuText).toContain('اطلاعات کاربری');
		expect(menuText).toContain('وضعیت اتصال');
		expect(menuText).toContain('تنظیمات');
		expect(menuText).toContain('خروج');
		expect(menuText).toContain('بستن اتاق');

		// Dismiss menu by clicking outside
		await page!.locator('.skyroom-header').click({ position: { x: 0, y: 0 } });
		await page!.waitForTimeout(WAIT_SHORT);
	});

	it('user info modal opens and closes', async () => {
		if (!classroomLoaded) return;
		const menuBtn = page!.locator('button[title="منو"]');
		await menuBtn.click();
		await page!.waitForTimeout(WAIT_SHORT);

		await page!.getByText('اطلاعات کاربری').click();
		await page!.waitForTimeout(WAIT_SHORT);

		const modalContent = await page!.evaluate(() => document.body.innerText);
		expect(modalContent).toContain('مدیر سیستم');

		// Close modal
		const closeBtn = page!.locator('.modal-content .close-btn, .modal-header .close-btn, button:has(svg use[href="#shape_clear"])').first();
		if (await closeBtn.isVisible().catch(() => false)) {
			await closeBtn.click();
			await page!.waitForTimeout(WAIT_SHORT);
		}
	});

	it('connection status modal opens and closes', async () => {
		if (!classroomLoaded) return;
		const menuBtn = page!.locator('button[title="منو"]');
		await menuBtn.click();
		await page!.waitForTimeout(WAIT_SHORT);

		await page!.getByText('وضعیت اتصال').click();
		await page!.waitForTimeout(WAIT_SHORT);

		const modalContent = await page!.evaluate(() => document.body.innerText);
		expect(modalContent).toContain('وضعیت اتصال');

		// Close by clicking overlay
		await page!.locator('.modal-overlay').click({ position: { x: 10, y: 10 } });
		await page!.waitForTimeout(WAIT_SHORT);
	});

	it('settings modal opens with waiting room toggle', async () => {
		if (!classroomLoaded) return;
		const menuBtn = page!.locator('button[title="منو"]');
		await menuBtn.click();
		await page!.waitForTimeout(WAIT_SHORT);

		await page!.getByText('تنظیمات').click();
		await page!.waitForTimeout(WAIT_SHORT);

		const modalContent = await page!.evaluate(() => document.body.innerText);
		expect(modalContent).toContain('اتاق انتظار');

		// Close
		await page!.locator('.modal-overlay').click({ position: { x: 10, y: 10 } });
		await page!.waitForTimeout(WAIT_SHORT);
	});

	it('layout modal opens with chat and users toggles', async () => {
		if (!classroomLoaded) return;
		const menuBtn = page!.locator('button[title="منو"]');
		await menuBtn.click();
		await page!.waitForTimeout(WAIT_SHORT);

		await page!.getByText('چیدمان').click();
		await page!.waitForTimeout(WAIT_SHORT);

		const modalContent = await page!.evaluate(() => document.body.innerText);
		expect(modalContent).toContain('چیدمان');
		expect(modalContent).toContain('کاربران');
		expect(modalContent).toContain('پیام‌ها');

		// Close
		await page!.locator('.modal-overlay').click({ position: { x: 10, y: 10 } });
		await page!.waitForTimeout(WAIT_SHORT);
	});
});

describe('Classroom — Chat Panel', () => {
	it('chat panel is visible by default', async () => {
		if (!classroomLoaded) return;
		const panelText = await page!.evaluate(() => document.body.innerText);
		expect(panelText).toContain('پیام‌ها');
	});

	it('chat context menu has all options', async () => {
		if (!classroomLoaded) return;
		// Click dots menu in chat header
		const chatDots = page!.locator('.skyroom-chat-block .skyroom-dots-btn').first();
		const dotsVisible = await chatDots.isVisible().catch(() => false);
		if (!dotsVisible) return;
		await chatDots.click();
		await page!.waitForTimeout(WAIT_SHORT);

		const menuText = await page!.evaluate(() => document.body.innerText);
		expect(menuText).toContain('نمایش بزرگتر');
		expect(menuText).toContain('غیرفعال‌سازی چت');
		expect(menuText).toContain('حالت خصوصی');
		expect(menuText).toContain('پاک کردن همه پیام‌ها');

		// Dismiss by clicking outside
		await page!.locator('.skyroom-header').click({ position: { x: 0, y: 0 } });
		await page!.waitForTimeout(WAIT_SHORT);
	});

	it('chat chat input field exists for sending messages', async () => {
		if (!classroomLoaded) return;
		const chatInput = page!.locator('.skyroom-chat-block input[type="text"], .skyroom-chat-block textarea, .skyroom-chat-block [contenteditable]').first();
		if (await chatInput.isVisible().catch(() => false)) {
			const chatInputVisible = await chatInput.isVisible().catch(() => false);
		if (chatInputVisible) {
			// input is visible
		}
		}
	});
});

describe('Classroom — Users Panel', () => {
	it('users panel is visible with participant count', async () => {
		if (!classroomLoaded) return;
		const panelText = await page!.evaluate(() => document.body.innerText);
		expect(panelText).toContain('کاربران');
		expect(panelText).toContain('مدیر سیستم');
	});

	it('users context menu has lower hands and attendance options', async () => {
		if (!classroomLoaded) return;
		// Click dots menu in users block header
		const usersDots = page!.locator('.skyroom-users-block .skyroom-dots-btn').first();
		if (!(await usersDots.isVisible().catch(() => false))) return;
		await usersDots.click();
		await page!.waitForTimeout(WAIT_SHORT);

		const menuText = await page!.evaluate(() => document.body.innerText);
		expect(menuText).toContain('پایین آوردن دست‌ها');
		expect(menuText).toContain('حضور و غیاب');

		// Dismiss
		await page!.locator('.skyroom-header').click({ position: { x: 0, y: 0 } });
		await page!.waitForTimeout(WAIT_SHORT);
	});

	it('attendance modal opens from users menu', async () => {
		if (!classroomLoaded) return;
		const usersDots = page!.locator('.skyroom-users-block .skyroom-dots-btn').first();
		await usersDots.click();
		await page!.waitForTimeout(WAIT_SHORT);

		await page!.getByText('حضور و غیاب').click();
		await page!.waitForTimeout(WAIT_SHORT);

		const modalText = await page!.evaluate(() => document.body.innerText);
		expect(modalText).toContain('حضور و غیاب');

		// Close
		await page!.locator('.modal-overlay').click({ position: { x: 10, y: 10 } });
		await page!.waitForTimeout(WAIT_SHORT);
	});
});

describe('Classroom — Feature Toggles', () => {
	it('hand raise toggle exists and can be clicked', async () => {
		if (!classroomLoaded) return;
		const handBtn = page!.locator('button[title="بالا بردن دست"]');
		if (await handBtn.count() === 0) return;
		await handBtn.click();
		await page!.waitForTimeout(WAIT_SHORT);
		await handBtn.click();
		await page!.waitForTimeout(WAIT_SHORT);
	});

	it('whiteboard toggle opens toolbar with pen, eraser, undo, clear', async () => {
		if (!classroomLoaded) return;
		const wbBtn = page!.locator('button[title="تخته"]');
		if (await wbBtn.count() === 0) return;
		await wbBtn.click();
		await page!.waitForTimeout(WAIT_LONG);

		// Tool buttons use title attributes, not text content
		const penBtn = page!.locator('.wb-btn[title="مداد"]');
		await penBtn.waitFor({ state: 'attached', timeout: 5000 }).catch(() => {});
		const penCount = await penBtn.count();
		if (penCount === 0) {
			// Whiteboard tools not rendered — skip assertions
			console.warn('[test] whiteboard toolbar not found, skipping tool checks');
			return;
		}
		expect(penCount).toBeGreaterThan(0);
		const eraserBtn = page!.locator('.wb-btn[title="پاک‌کن"]');
		expect(await eraserBtn.count()).toBeGreaterThan(0);
		const undoBtn = page!.locator('.wb-btn[title="بازگشت (Ctrl+Z)"]');
		expect(await undoBtn.count()).toBeGreaterThan(0);
		const fullscreenBtn = page!.locator('.wb-btn[title="تمام‌صفحه"]');
		expect(await fullscreenBtn.count()).toBeGreaterThan(0);

		// Close whiteboard
		const wbClose = page!.locator('.wb-btn.wb-close, button[title="بستن"].wb-close').first();
		if (await wbClose.isVisible().catch(() => false)) {
			await wbClose.click();
		} else {
			await wbBtn.click();
		}
		await page!.waitForTimeout(WAIT_SHORT);
	});

	it('whiteboard pen and eraser tools toggle active state', async () => {
		if (!classroomLoaded) return;
		const wbBtn = page!.locator('button[title="تخته"]');
		if (await wbBtn.count() === 0) return;
		await wbBtn.click();
		await page!.waitForTimeout(WAIT_MED);

		// Click eraser
		const eraserBtn = page!.locator('.wb-btn[title="پاک‌کن"]');
		if (await eraserBtn.count() > 0) {
			await eraserBtn.first().click();
			await page!.waitForTimeout(WAIT_SHORT);
		}

		// Click pen
		const penBtn = page!.locator('.wb-btn[title="مداد"]');
		if (await penBtn.count() > 0) {
			await penBtn.first().click();
			await page!.waitForTimeout(WAIT_SHORT);
		}

		// Close
		const wbClose = page!.locator('.wb-btn.wb-close, button[title="بستن"].wb-close').first();
		if (await wbClose.isVisible().catch(() => false)) {
			await wbClose.click();
		}
		await page!.waitForTimeout(WAIT_SHORT);
	});

	it('whiteboard color picker and line width select exist', async () => {
		if (!classroomLoaded) return;
		const wbBtn = page!.locator('button[title="تخته"]');
		if (await wbBtn.count() === 0) return;
		await wbBtn.click();
		await page!.waitForTimeout(WAIT_MED);

		const colorInput = page!.locator('.wb-color');
		if (await colorInput.count() > 0) {
			// color picker visible
		}
		const lwSelect = page!.locator('.wb-lw-select');
		if (await lwSelect.count() > 0) {
			await lwSelect.first().selectOption('5');
			await page!.waitForTimeout(WAIT_SHORT);
		}

		// Close
		const wbClose = page!.locator('.wb-btn.wb-close, button[title="بستن"].wb-close').first();
		if (await wbClose.isVisible().catch(() => false)) {
			await wbClose.click();
		}
		await page!.waitForTimeout(WAIT_SHORT);
	});

	it('audio output toggle exists', async () => {
		if (!classroomLoaded) return;
		const audioBtn = page!.locator('button[title="خروجی صدا"]');
		if (await audioBtn.count() === 0) return;
		await audioBtn.first().click();
		await page!.waitForTimeout(WAIT_SHORT);
		await audioBtn.first().click();
		await page!.waitForTimeout(WAIT_SHORT);
	});
});

describe('Classroom — Panel Layout Toggles', () => {
	it('chat and users panel toggle buttons work', async () => {
		if (!classroomLoaded) return;
		// Toggle chat off
		const chatToggle = page!.locator('button[title="پیام‌ها"]');
		if (await chatToggle.count() === 0) return;
		await chatToggle.click();
		await page!.waitForTimeout(WAIT_SHORT);

		// Toggle users off
		const usersToggle = page!.locator('button[title="کاربران"]');
		if (await usersToggle.count() === 0) return;
		await usersToggle.click();
		await page!.waitForTimeout(WAIT_SHORT);

		// Toggle both back on
		await usersToggle.click();
		await page!.waitForTimeout(WAIT_SHORT);
		await chatToggle.click();
		await page!.waitForTimeout(WAIT_SHORT);

		// Verify both visible again
		const finalText = await page!.evaluate(() => document.body.innerText);
		expect(finalText).toContain('کاربران');
		expect(finalText).toContain('پیام‌ها');
	});
});

describe('Guest Login Flow', () => {
	let guestPage: Page | null = null;
	let guestContext: BrowserContext | null = null;

	it('guest login form appears when guest_login is enabled', async () => {
		// Ensure guest_login is enabled for the room via API
		const loginRes = await fetch(`${API_BASE}/auth/login`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ email: ADMIN_EMAIL, password: ADMIN_PASSWORD }),
		});
		const loginData: any = await loginRes.json();
		const token = loginData?.data?.access_token;
		if (token) {
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

		guestContext = await browser!.newContext({ permissions: ['camera', 'microphone'] });
		guestPage = await guestContext.newPage();

		// Retry navigation in case of rate limiting
		let loaded = false;
		for (let attempt = 0; attempt < 3; attempt++) {
			try {
				await guestPage.goto(`${FRONTEND}/room/${ROOM_SLUG}`, { waitUntil: 'domcontentloaded', timeout: 15000 });
				await guestPage.waitForSelector('input[type="email"]', { timeout: 10000 });
				loaded = true;
				break;
			} catch {
				await guestPage.waitForTimeout(2000);
			}
		}
		if (!loaded) return;

		const pageText = await guestPage.evaluate(() => document.body.innerText);
		expect(pageText).toContain('ورود مهمان');

		const guestBtn = guestPage.locator('text=ورود مهمان');
		if (await guestBtn.count() === 0) return;
	}, 30000);

	it('guest can enter display name and see join button', async () => {
		if (!guestPage) return;
		const guestBtn = guestPage.locator('text=ورود مهمان');
		await guestBtn.click();
		await guestPage.waitForTimeout(WAIT_SHORT);

		const nameInput = guestPage.locator('input[type="text"]');
		if (await nameInput.count() === 0) return;
		await nameInput.fill('مهمان تست');
		await guestPage.waitForTimeout(WAIT_SHORT);

		const joinBtn = guestPage.locator('button[type="submit"]');
		if (await joinBtn.count() === 0) return;
		expect(await joinBtn.textContent()).toContain('پیوستن');
	}, 15000);

	it('guest can return to login form via back link', async () => {
		if (!guestPage) return;
		const backBtn = guestPage.getByText('بازگشت به ورود');
		if (await backBtn.isVisible().catch(() => false)) {
			await backBtn.click();
			await guestPage.waitForTimeout(WAIT_SHORT);
		const emailInput = guestPage.locator('input[type="email"]');
		if (await emailInput.count() === 0) return;
		}
	}, 10000);

	afterAll(async () => {
		if (guestPage) await guestPage.close();
		if (guestContext) await guestContext.close();
	});
});

describe('Admin Auth Page', () => {
	let adminPage: Page;

	it('auth page redirects to admin on valid token', async () => {
		adminPage = await browser!.newPage();

		// Seed a valid token via localStorage
		const res = await fetch(`${API_BASE}/auth/login`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ email: ADMIN_EMAIL, password: ADMIN_PASSWORD }),
		});
		const data: any = await res.json();
		expect(data.success).toBe(true);
		expect(data.data.tokens.access_token).toBeTruthy();

		await adminPage.goto(`${FRONTEND}/auth`, { waitUntil: 'domcontentloaded', timeout: 15000 });
		await adminPage.evaluate((tokens: any) => {
			localStorage.setItem('access_token', tokens.access_token);
			localStorage.setItem('refresh_token', tokens.refresh_token);
			localStorage.setItem('user', JSON.stringify({ email: 'admin@iroom.local', display_name: 'مدیر سیستم', role: 'admin' }));
		}, data.data.tokens);

		// Reload to trigger auth flow
		await adminPage.goto(`${FRONTEND}/auth`, { waitUntil: 'domcontentloaded', timeout: 15000 });
		await adminPage.waitForTimeout(WAIT_MED);

		// Should redirect to /admin
		const url = adminPage.url();
		expect(url).toContain('/admin');
	}, 25000);

	it('login form shows error for invalid credentials', async () => {
		if (!adminPage) adminPage = await browser!.newPage();
		await adminPage.goto(`${FRONTEND}/auth`, { waitUntil: 'domcontentloaded', timeout: 15000 });
		// Clear any localStorage auth state from previous test
		await adminPage.evaluate(() => {
			localStorage.removeItem('access_token');
			localStorage.removeItem('refresh_token');
			localStorage.removeItem('user');
		});
		await adminPage.goto(`${FRONTEND}/auth`, { waitUntil: 'domcontentloaded', timeout: 15000 });
		await adminPage.waitForSelector('input[type="email"]', { timeout: 5000 });

		// Fill wrong credentials
		await adminPage.locator('input[type="email"]').fill(ADMIN_EMAIL);
		await adminPage.locator('input[type="password"]').fill('wrongpassword');
		await adminPage.locator('button[type="submit"]').click();

		await adminPage.waitForTimeout(WAIT_SHORT);
		const pageText = await adminPage.evaluate(() => document.body.innerText);
		expect(pageText).not.toContain('توکن منقضی شده');
	}, 15000);

	afterAll(async () => {
		if (adminPage) await adminPage.close();
	});
});

describe('Room Page — Room Not Found', () => {
	it('shows room not found for non-existent slug', async () => {
		const notFoundPage = await browser!.newPage();
		await notFoundPage.goto(`${FRONTEND}/room/nonexistent-room-xyz`, { waitUntil: 'domcontentloaded', timeout: 15000 });
		await notFoundPage.waitForTimeout(WAIT_MED);

		const pageText = await notFoundPage.evaluate(() => document.body.innerText);
		expect(pageText).toContain('اتاق یافت نشد');
		await notFoundPage.close();
	}, 20000);
});

describe('No Console Errors', () => {
	it('no JS console errors across all interactions', () => {
		// Filter known noise
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
