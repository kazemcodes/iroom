/**
 * CI-like browser test for the track.onended verification page.
 *
 * Uses Playwright + Vitest to launch headless Chromium with fake media
 * devices (where supported), run the test suite at /track-ended-test.html,
 * and assert the page loads and JS executes without errors.
 *
 * Note: Fake media device flags (`--use-fake-device-for-media-stream`)
 * depend on the browser version and platform. Tests verify page behavior
 * rather than strict pass/fail of every individual track.onended assertion.
 */

import { describe, it, beforeAll, afterAll, expect } from 'vitest';
import { chromium, type Browser, type Page } from 'playwright';
import { createServer, type Server } from 'node:http';
import { readFileSync, existsSync } from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const staticDir = path.resolve(__dirname, '..', 'static');

let actualPort = 0;
let BASE = '';

let server: Server | null = null;
let browser: Browser | null = null;
let page: Page | null = null;

beforeAll(async () => {
	server = createServer((req, res) => {
		const urlPath = req.url === '/' ? '/track-ended-test.html' : (req.url || '/');
		const filePath = path.join(staticDir, urlPath);

		if (existsSync(filePath)) {
			const ext = path.extname(filePath);
			const mimeTypes: Record<string, string> = {
				'.html': 'text/html',
				'.js': 'application/javascript',
				'.css': 'text/css',
			};
			res.writeHead(200, { 'Content-Type': mimeTypes[ext] || 'text/plain' });
			res.end(readFileSync(filePath));
		} else {
			res.writeHead(404);
			res.end('Not found');
		}
	});

	await new Promise<void>((resolve, reject) => {
		server!.on('error', reject);
		server!.listen(0, () => {
			const addr = server!.address();
			if (addr && typeof addr === 'object') actualPort = addr.port;
			BASE = `http://localhost:${actualPort}`;
			resolve();
		});
	});

	const chromePath = process.env.CHROME_BIN ||
		(process.platform === 'darwin'
			? '/Applications/Google Chrome.app/Contents/MacOS/Google Chrome'
			: '/usr/bin/google-chrome-stable');

	browser = await chromium.launch({
		executablePath: chromePath,
		headless: true,
		args: [
			'--no-sandbox',
			'--disable-setuid-sandbox',
			'--disable-dev-shm-usage',
			'--use-fake-device-for-media-stream',
			'--use-fake-ui-for-media-stream',
			'--auto-select-desktop-capture-source=Entire Screen',
		],
	});

	const context = await browser.newContext({
		permissions: ['camera', 'microphone'],
	});

	page = await context.newPage();

	// Collect browser console errors for diagnostics
	const browserErrors: string[] = [];
	page.on('console', (msg) => {
		if (msg.type() === 'error' && !msg.text().includes('404') && !msg.text().includes('favicon')) {
			browserErrors.push(msg.text());
		}
	});

	// Navigate once for all tests
	await page.goto(`${BASE}/track-ended-test.html`, {
		waitUntil: 'domcontentloaded',
		timeout: 10000,
	});

	// Click "Run All Tests" once so all subsequent tests can inspect results
	await page.locator('#runAllBtn').click();
	const summary = page.locator('#summary');
	await summary.waitFor({ state: 'visible', timeout: 20000 });

	// Store for assertions
	(page as any).__browserErrors = browserErrors;
}, 25000);

afterAll(async () => {
	if (browser) await browser.close();
	if (server) await new Promise<void>((resolve) => server!.close(() => resolve()));
}, 10000);

describe('track-ended-test.html — browser E2E', () => {
	it('page loads with correct title', async () => {
		const title = await page!.title();
		expect(title).toContain('track.onended');
	});

	it('all test rows resolved (none stuck in pending state)', async () => {
		const pendingRows = page!.locator('.test-row.pending');
		const pendingCount = await pendingRows.count();
		expect(pendingCount).toBe(0);
	});

	it('page JavaScript executed without unhandled errors', async () => {
		const logText = await page!.locator('#log').textContent();
		expect(logText).not.toContain('ReferenceError');
		expect(logText).not.toContain('TypeError');
	});

	it('at least 8 out of 24 tests pass (fake media is environment-dependent)', async () => {
		const passRows = page!.locator('.test-row.pass');
		const passCount = await passRows.count();
		console.log(`Browser test results: ${passCount}/24 tests passed`);

		// Pattern tests (onended not re-triggered) and screen-share cancellation
		// should always pass regardless of fake media support.
		expect(passCount).toBeGreaterThanOrEqual(8);
	});

	it('MediaClient pattern-4 test passes: direct stopXxx() does NOT re-trigger onended', async () => {
		const pattern4 = page!.locator('#pattern-4');
		const class4 = await pattern4.getAttribute('class');
		expect(class4).toContain('pass');
	});

	it('no unexpected browser console errors', async () => {
		const errors: string[] = (page as any).__browserErrors || [];
		expect(errors).toHaveLength(0);
	});

	it('page summary displays pass/fail counts', async () => {
		const summary = page!.locator('#summary');
		const text = await summary.textContent();
		expect(text).toContain('Passed:');
		expect(await summary.isVisible()).toBe(true);
	});
});
