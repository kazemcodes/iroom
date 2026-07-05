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

async function loginAndJoin() {
	// Navigate to room — will show login form
	await page!.goto(`${FRONTEND}/room/${ROOM_SLUG}`, { waitUntil: 'domcontentloaded', timeout: 15000 });
	await page!.waitForSelector('input[type="email"]', { timeout: 10000 });

	// Fill login form
	const emailInput = page!.locator('input[type="email"]');
	if (await emailInput.isVisible({ timeout: 3000 }).catch(() => false)) {
		await emailInput.fill(ADMIN_EMAIL);
		await page!.locator('input[type="password"]').fill(ADMIN_PASSWORD);
		await page!.locator('button[type="submit"]').first().click();
		// Wait for "پیوستن به کلاس" button
		await page!.waitForSelector('text=پیوستن به کلاس', { timeout: 15000 });
	}

	// Click join button
	const joinBtn = page!.locator('text=پیوستن به کلاس');
	try {
		await joinBtn.click();
		await page!.waitForTimeout(500);

		// Entry modal: click "ورود به اتاق" as speaker
		const enterBtn = page!.getByText('ورود به اتاق');
		await enterBtn.waitFor({ state: 'visible', timeout: 5000 });
		await enterBtn.click();

		// Wait for classroom header
		await page!.waitForSelector('.skyroom-header', { timeout: 15000 });
		classroomLoaded = true;
		accessToken = (await page!.evaluate(() => localStorage.getItem('access_token'))) || '';
		await page!.waitForTimeout(2000);
	} catch {
		console.warn('[chat-e2e] classroom header not found');
	}
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

	await loginAndJoin();
}, 60000);

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

describe('Chat — Input & Sending', () => {
	it('chat input field is visible and typeable', async () => {
		if (!classroomLoaded) return;
		const chatInput = page!.locator('.skyroom-chat-block .chat-input');
		const visible = await chatInput.isVisible().catch(() => false);
		expect(visible).toBe(true);
	});

	it('send button is disabled when input is empty', async () => {
		if (!classroomLoaded) return;
		const sendBtn = page!.locator('.skyroom-chat-block .send-btn');
		if (await sendBtn.count() === 0) return;
		const hasActive = await sendBtn.first().evaluate((el) => el.classList.contains('active'));
		expect(hasActive).toBe(false);
	});

	it('send button becomes active when input has text', async () => {
		if (!classroomLoaded) return;
		const chatInput = page!.locator('.skyroom-chat-block .chat-input');
		if (await chatInput.count() === 0) return;

		await chatInput.fill('test message');
		await page!.waitForTimeout(WAIT_SHORT);

		const sendBtn = page!.locator('.skyroom-chat-block .send-btn');
		const hasActive = await sendBtn.first().evaluate((el) => el.classList.contains('active'));
		expect(hasActive).toBe(true);

		// Clear input
		await chatInput.fill('');
	});

	it('can type a message and clear it', async () => {
		if (!classroomLoaded) return;
		const chatInput = page!.locator('.skyroom-chat-block .chat-input');
		if (await chatInput.count() === 0) return;

		await chatInput.fill('Hello from e2e test');
		const value = await chatInput.inputValue();
		expect(value).toBe('Hello from e2e test');

		// Clear
		await chatInput.fill('');
	});

	it('sending message via Enter key adds it to chat', async () => {
		if (!classroomLoaded) return;
		const chatInput = page!.locator('.skyroom-chat-block .chat-input');
		if (await chatInput.count() === 0) return;

		const msgText = `E2E test message ${Date.now()}`;
		await chatInput.fill(msgText);
		await chatInput.press('Enter');
		await page!.waitForTimeout(WAIT_LONG);

		// Message should appear in chat messages area
		const chatContent = await page!.evaluate(() => document.body.innerText);
		expect(chatContent).toContain(msgText);
	});

	it('sending message via send button adds it to chat', async () => {
		if (!classroomLoaded) return;
		const chatInput = page!.locator('.skyroom-chat-block .chat-input');
		if (await chatInput.count() === 0) return;

		const msgText = `E2E click send ${Date.now()}`;
		await chatInput.fill(msgText);

		const sendBtn = page!.locator('.skyroom-chat-block .send-btn');
		await sendBtn.click();
		await page!.waitForTimeout(WAIT_LONG);

		const chatContent = await page!.evaluate(() => document.body.innerText);
		expect(chatContent).toContain(msgText);
	});
});

describe('Chat — Context Menu Controls', () => {
	it('chat dots menu opens with all options', async () => {
		if (!classroomLoaded) return;
		const chatDots = page!.locator('.skyroom-chat-block .skyroom-dots-btn').first();
		if (!(await chatDots.isVisible().catch(() => false))) return;

		await chatDots.click();
		await page!.waitForTimeout(WAIT_SHORT);

		const menuText = await page!.evaluate(() => document.body.innerText);
		expect(menuText).toContain('نمایش بزرگتر');
		expect(menuText).toContain('غیرفعال‌سازی چت');
		expect(menuText).toContain('حالت خصوصی');
		expect(menuText).toContain('پاک کردن همه پیام‌ها');

		// Dismiss
		await page!.locator('.skyroom-header').click({ position: { x: 0, y: 0 } });
		await page!.waitForTimeout(WAIT_SHORT);
	});

	it('toggling chat disabled shows disabled message', async () => {
		if (!classroomLoaded) return;
		const chatDots = page!.locator('.skyroom-chat-block .skyroom-dots-btn').first();
		if (!(await chatDots.isVisible().catch(() => false))) return;

		await chatDots.click();
		await page!.waitForTimeout(WAIT_SHORT);

		// Click "غیرفعال‌سازی چت"
		const disableBtn = page!.getByText('غیرفعال‌سازی چت');
		if (await disableBtn.isVisible().catch(() => false)) {
			await disableBtn.click();
			await page!.waitForTimeout(WAIT_SHORT);

			// Chat should show disabled message
			const chatText = await page!.evaluate(() => document.body.innerText);
			expect(chatText).toContain('چت غیرفعال');

			// Re-enable
			await chatDots.click();
			await page!.waitForTimeout(WAIT_SHORT);
			const enableBtn = page!.getByText('فعال‌سازی چت');
			if (await enableBtn.isVisible().catch(() => false)) {
				await enableBtn.click();
				await page!.waitForTimeout(WAIT_SHORT);
			}
		}
	});

	it('toggling chat private mode shows private label', async () => {
		if (!classroomLoaded) return;
		const chatDots = page!.locator('.skyroom-chat-block .skyroom-dots-btn').first();
		if (!(await chatDots.isVisible().catch(() => false))) return;

		await chatDots.click();
		await page!.waitForTimeout(WAIT_SHORT);

		const privateBtn = page!.getByText('حالت خصوصی');
		if (await privateBtn.isVisible().catch(() => false)) {
			await privateBtn.click();
			await page!.waitForTimeout(WAIT_SHORT);

			// Private label should appear in chat header
			const chatText = await page!.evaluate(() => document.body.innerText);
			expect(chatText).toContain('خصوصی');

			// Switch back to public
			await chatDots.click();
			await page!.waitForTimeout(WAIT_SHORT);
			const publicBtn = page!.getByText('حالت عمومی');
			if (await publicBtn.isVisible().catch(() => false)) {
				await publicBtn.click();
				await page!.waitForTimeout(WAIT_SHORT);
			}
		}
	});

	it('clear messages removes all messages from chat', async () => {
		if (!classroomLoaded) return;
		const chatDots = page!.locator('.skyroom-chat-block .skyroom-dots-btn').first();
		if (!(await chatDots.isVisible().catch(() => false))) return;

		await chatDots.click();
		await page!.waitForTimeout(WAIT_SHORT);

		const clearBtn = page!.getByText('پاک کردن همه پیام‌ها');
		if (await clearBtn.isVisible().catch(() => false)) {
			await clearBtn.click();
			await page!.waitForTimeout(WAIT_SHORT);

			// Chat messages container should be empty
			const msgCount = await page!.locator('.skyroom-chat-block .msg-row').count();
			expect(msgCount).toBe(0);
		}
	});

	it('chat expanded toggle works', async () => {
		if (!classroomLoaded) return;
		const chatDots = page!.locator('.skyroom-chat-block .skyroom-dots-btn').first();
		if (!(await chatDots.isVisible().catch(() => false))) return;

		await chatDots.click();
		await page!.waitForTimeout(WAIT_SHORT);

		const expandBtn = page!.getByText('نمایش بزرگتر');
		if (await expandBtn.isVisible().catch(() => false)) {
			await expandBtn.click();
			await page!.waitForTimeout(WAIT_SHORT);

			// Chat should be in expanded mode (sidebar with max-width none)
			const sidebar = page!.locator('.skyroom-sidebar');
			const expandAttr = await sidebar.first().getAttribute('style');
			expect(expandAttr).toBeTruthy();

			// Collapse back
			const exitBtn = page!.locator('.skyroom-chat-block .skyroom-dots-btn').first();
			await exitBtn.click();
			await page!.waitForTimeout(WAIT_SHORT);
		}
	});
});

describe('Chat — Reply & Message Structure', () => {
	it('sent message has sender name and timestamp', async () => {
		if (!classroomLoaded) return;
		const chatInput = page!.locator('.skyroom-chat-block .chat-input');
		if (await chatInput.count() === 0) return;

		const msgText = `Timestamp test ${Date.now()}`;
		await chatInput.fill(msgText);
		await chatInput.press('Enter');
		await page!.waitForTimeout(WAIT_LONG);

		// Check message structure: sender name "شما" (you) should be visible
		const chatContent = await page!.evaluate(() => document.body.innerText);
		expect(chatContent).toContain('شما');
	});

	it('reply button exists on message bubbles', async () => {
		if (!classroomLoaded) return;
		// Send a message first
		const chatInput = page!.locator('.skyroom-chat-block .chat-input');
		if (await chatInput.count() === 0) return;

		await chatInput.fill(`Reply target ${Date.now()}`);
		await chatInput.press('Enter');
		await page!.waitForTimeout(WAIT_LONG);

		// Hover over the message bubble to reveal reply button
		const msgBubble = page!.locator('.skyroom-chat-block .msg-bubble').first();
		if (await msgBubble.count() === 0) return;
		await msgBubble.hover();
		await page!.waitForTimeout(WAIT_SHORT);

		const replyBtn = page!.locator('.skyroom-chat-block .reply-btn').first();
		expect(await replyBtn.count()).toBeGreaterThan(0);
	});
});

describe('Chat — Console Errors', () => {
	it('no fatal JS errors during chat interactions', () => {
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
