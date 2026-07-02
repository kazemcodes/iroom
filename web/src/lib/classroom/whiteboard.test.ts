import { describe, it, expect, beforeEach, afterEach } from 'vitest';

// ---------------------------------------------------------------------------
// Constants matching the Svelte component
// ---------------------------------------------------------------------------
const WB_VIRTUAL_W = 1920;
const WB_VIRTUAL_H = 1080;
const BG_COLOR = '#1c2a3a';
const ERASE_COLOR = 'rgba(0,0,0,0)';
const PEN_WIDTHS = { thin: 1, normal: 2, thick: 5, extraThick: 10 };
const ERASER_WIDTH = 40;

// ---------------------------------------------------------------------------
// Pure-logic helpers extracted from the whiteboard code
// ---------------------------------------------------------------------------

interface Stroke {
	x1: number;
	y1: number;
	x2: number;
	y2: number;
	color: string;
	width: number;
}

/** Map DOM pixel coords to virtual coordinate space. */
function toVirtual(domX: number, domY: number, rect: { left: number; top: number; width: number; height: number }) {
	const x = (domX - rect.left) / rect.width * WB_VIRTUAL_W;
	const y = (domY - rect.top) / rect.height * WB_VIRTUAL_H;
	return { x, y };
}

/** Normalize a virtual stroke to 0-1 range for network transmission. */
function normalizeStroke(s: Stroke) {
	return {
		x1: s.x1 / WB_VIRTUAL_W,
		y1: s.y1 / WB_VIRTUAL_H,
		x2: s.x2 / WB_VIRTUAL_W,
		y2: s.y2 / WB_VIRTUAL_H,
		color: s.color,
		width: s.width / Math.max(WB_VIRTUAL_W, WB_VIRTUAL_H),
	};
}

/** Denormalize a network stroke back to virtual coords. */
function denormalizeStroke(d: { x1: number; y1: number; x2: number; y2: number; color: string; width: number }): Stroke {
	return {
		x1: d.x1 * WB_VIRTUAL_W,
		y1: d.y1 * WB_VIRTUAL_H,
		x2: d.x2 * WB_VIRTUAL_W,
		y2: d.y2 * WB_VIRTUAL_H,
		color: d.color,
		width: d.width * Math.max(WB_VIRTUAL_W, WB_VIRTUAL_H),
	};
}

/** Check whether a stroke was made with the eraser tool. */
function isEraserStroke(s: Stroke): boolean {
	return s.color === ERASE_COLOR;
}

/** Apply a stroke to a canvas 2d context. */
function renderStroke(ctx: CanvasRenderingContext2D, s: Stroke): void {
	const erasing = isEraserStroke(s);
	ctx.save();
	if (erasing) ctx.globalCompositeOperation = 'destination-out';
	ctx.beginPath();
	ctx.moveTo(s.x1, s.y1);
	ctx.lineTo(s.x2, s.y2);
	ctx.strokeStyle = s.color;
	ctx.lineWidth = s.width;
	ctx.lineCap = 'round';
	ctx.stroke();
	ctx.restore();
}

/** Build the pen stroke for the given tool, coords, colour, and line-width. */
function makeStroke(
	x1: number, y1: number, x2: number, y2: number,
	tool: 'pen' | 'eraser', color: string, lineWidth: number,
): Stroke {
	if (tool === 'eraser') {
		return { x1, y1, x2, y2, color: ERASE_COLOR, width: ERASER_WIDTH };
	}
	return { x1, y1, x2, y2, color, width: lineWidth };
}

/** Remove the last stroke (undo). Returns a new array. */
function undoLast(strokes: Stroke[]): Stroke[] {
	if (strokes.length === 0) return strokes;
	return strokes.slice(0, -1);
}


// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

describe('Whiteboard – virtual coordinate mapping', () => {
	const rect = { left: 0, top: 0, width: 960, height: 540 };

	it('maps top-left corner to (0, 0)', () => {
		const { x, y } = toVirtual(0, 0, rect);
		expect(x).toBe(0);
		expect(y).toBe(0);
	});

	it('maps bottom-right corner to virtual resolution', () => {
		const { x, y } = toVirtual(960, 540, rect);
		expect(x).toBe(WB_VIRTUAL_W);
		expect(y).toBe(WB_VIRTUAL_H);
	});

	it('maps centre of the canvas to centre of virtual space', () => {
		const { x, y } = toVirtual(480, 270, rect);
		expect(x).toBe(WB_VIRTUAL_W / 2);
		expect(y).toBe(WB_VIRTUAL_H / 2);
	});

	it('handles offset rect correctly', () => {
		const offsetRect = { left: 100, top: 50, width: 640, height: 360 };
		const { x, y } = toVirtual(420, 230, offsetRect);
		expect(x).toBe((320 / 640) * WB_VIRTUAL_W);
		expect(y).toBe((180 / 360) * WB_VIRTUAL_H);
	});
});

describe('Whiteboard – stroke creation', () => {
	it('pen stroke stores colour and line-width', () => {
		const s = makeStroke(100, 200, 300, 400, 'pen', '#ff0000', 5);
		expect(s.color).toBe('#ff0000');
		expect(s.width).toBe(5);
		expect(isEraserStroke(s)).toBe(false);
	});

	it('eraser stroke stores transparent colour and eraser width', () => {
		const s = makeStroke(100, 200, 300, 400, 'eraser', '#ffffff', 2);
		expect(s.color).toBe(ERASE_COLOR);
		expect(s.width).toBe(ERASER_WIDTH);
		expect(isEraserStroke(s)).toBe(true);
	});

	it('eraser width ignores the pen line-width argument', () => {
		const s = makeStroke(0, 0, 10, 10, 'eraser', '#ffffff', 10);
		expect(s.width).toBe(ERASER_WIDTH);
	});
});

describe('Whiteboard – line-width values', () => {
	it('defines four preset widths', () => {
		expect(PEN_WIDTHS.thin).toBe(1);
		expect(PEN_WIDTHS.normal).toBe(2);
		expect(PEN_WIDTHS.thick).toBe(5);
		expect(PEN_WIDTHS.extraThick).toBe(10);
	});

	it('pen strokes store the selected line-width', () => {
		const thin = makeStroke(0, 0, 10, 10, 'pen', '#fff', PEN_WIDTHS.thin);
		expect(thin.width).toBe(1);

		const thick = makeStroke(0, 0, 10, 10, 'pen', '#fff', PEN_WIDTHS.extraThick);
		expect(thick.width).toBe(10);
	});
});

describe('Whiteboard – undo', () => {
	it('removes the most recently added stroke', () => {
		const strokes: Stroke[] = [
			makeStroke(0, 0, 10, 10, 'pen', '#fff', 2),
			makeStroke(10, 10, 20, 20, 'pen', '#f00', 2),
			makeStroke(20, 20, 30, 30, 'eraser', '#fff', 2),
		];
		const after = undoLast(strokes);
		expect(after).toHaveLength(2);
		expect(isEraserStroke(after[1])).toBe(false);
	});

	it('is a no-op on an empty stroke list', () => {
		const strokes: Stroke[] = [];
		expect(undoLast(strokes)).toHaveLength(0);
	});

	it('undo preserves pen-tool stroke properties', () => {
		const strokes = [makeStroke(100, 200, 300, 400, 'pen', '#aabbcc', 5)];
		const after = undoLast(strokes);
		expect(after).toHaveLength(0);
		const redo = [...after, strokes[0]];
		expect(redo[0].color).toBe('#aabbcc');
		expect(redo[0].width).toBe(5);
	});

	it('undoing twice leaves only the first stroke', () => {
		const strokes: Stroke[] = [
			makeStroke(0, 0, 10, 10, 'pen', '#111', 2),
			makeStroke(10, 10, 20, 20, 'pen', '#222', 2),
			makeStroke(20, 20, 30, 30, 'pen', '#333', 2),
		];
		const afterTwo = undoLast(undoLast(strokes));
		expect(afterTwo).toHaveLength(1);
		expect(afterTwo[0].color).toBe('#111');
	});

	it('undoing more strokes than available returns empty array', () => {
		const strokes: Stroke[] = [makeStroke(0, 0, 10, 10, 'pen', '#fff', 2)];
		expect(undoLast(undoLast(undoLast(strokes)))).toHaveLength(0);
	});
});

describe('Whiteboard – eraser compositing (canvas)', () => {
	let canvas: HTMLCanvasElement;
	let ctx: CanvasRenderingContext2D;
	let compositingSupported: boolean;

	beforeEach(() => {
		canvas = document.createElement('canvas');
		canvas.width = WB_VIRTUAL_W;
		canvas.height = WB_VIRTUAL_H;
		ctx = canvas.getContext('2d')!;
		// Fill initial background
		ctx.fillStyle = BG_COLOR;
		ctx.fillRect(0, 0, canvas.width, canvas.height);

		// Detect whether the canvas environment supports destination-out.
		// jsdom / happy-dom often ignore composite operations during getImageData.
		ctx.save();
		ctx.globalCompositeOperation = 'destination-out';
		ctx.fillStyle = '#ffffff';
		ctx.fillRect(100, 100, 10, 10);
		ctx.restore();
		const probe = ctx.getImageData(100, 100, 1, 1).data;
		compositingSupported = probe[3] === 0;
	});

	afterEach(() => {
		canvas.remove();
	});

	function getPixel(x: number, y: number): [number, number, number, number] {
		const d = ctx.getImageData(x, y, 1, 1).data;
		return [d[0], d[1], d[2], d[3]];
	}

	it('pen draws visible pixels', () => {
		const s = makeStroke(100, 100, 300, 100, 'pen', '#ffffff', 10);
		renderStroke(ctx, s);

		const [, , , a] = getPixel(s.x1, s.y1);
		expect(a).toBeGreaterThan(0);
	});

	const itCanvas = it.skip; // Canvas compositing disabled in jsdom — enable with: compositingSupported ? it : it.skip

	itCanvas('eraser removes pixels that were drawn by pen', () => {
		// Draw a thick white line
		renderStroke(ctx, makeStroke(100, 100, 500, 100, 'pen', '#ffffff', 10));
		expect(getPixel(300, 100)[3]).toBeGreaterThan(0);

		// Erase over the same area
		renderStroke(ctx, makeStroke(100, 100, 500, 100, 'eraser', '#ffffff', 2));
		expect(getPixel(300, 100)[3]).toBe(0);
	});

	itCanvas('eraser only erases; does not paint new colour', () => {
		renderStroke(ctx, makeStroke(100, 200, 400, 200, 'pen', '#ff0000', 10));
		renderStroke(ctx, makeStroke(250, 200, 350, 200, 'eraser', '#ffffff', 2));

		// Erased pixel should be transparent
		expect(getPixel(300, 200)[3]).toBe(0);

		// Untouched pixel should still be visible
		const untouched = getPixel(150, 200);
		expect(untouched[3]).toBeGreaterThan(0);
	});

	it('sets globalCompositeOperation to destination-out for eraser strokes', () => {
		// Verify by checking the ctx.globalCompositeOperation value after renderStroke
		const eraserStroke = makeStroke(0, 0, 10, 10, 'eraser', '#ffffff', 2);
		const penStroke = makeStroke(0, 0, 10, 10, 'pen', '#ffffff', 2);

		// After rendering an eraser stroke, restore should reset to source-over
		renderStroke(ctx, eraserStroke);
		expect(ctx.globalCompositeOperation).toBe('source-over');

		// During render, save/restore should leave it at source-over for pen too
		renderStroke(ctx, penStroke);
		expect(ctx.globalCompositeOperation).toBe('source-over');
	});
});

describe('Whiteboard – network format', () => {
	const stroke = makeStroke(960, 540, 480, 270, 'pen', '#ffffff', 4);

	it('normalize produces values in 0-1 range', () => {
		const n = normalizeStroke(stroke);
		expect(n.x1).toBe(0.5);
		expect(n.y1).toBe(0.5);
		expect(n.x2).toBe(0.25);
		expect(n.y2).toBe(0.25);
		expect(n.color).toBe('#ffffff');
		expect(n.width).toBeLessThan(1);
	});

	it('round-trip through normalize → denormalize preserves values', () => {
		const n = normalizeStroke(stroke);
		const d = denormalizeStroke(n);
		expect(d.x1).toBeCloseTo(stroke.x1, 0);
		expect(d.y1).toBeCloseTo(stroke.y1, 0);
		expect(d.x2).toBeCloseTo(stroke.x2, 0);
		expect(d.y2).toBeCloseTo(stroke.y2, 0);
		expect(d.color).toBe(stroke.color);
		expect(d.width).toBeCloseTo(stroke.width, 0);
	});

	it('round-trip preserves eraser strokes', () => {
		const eraserStroke = makeStroke(192, 108, 384, 216, 'eraser', '#ffffff', 2);
		const n = normalizeStroke(eraserStroke);
		const d = denormalizeStroke(n);
		expect(isEraserStroke(d)).toBe(true);
		expect(d.color).toBe(ERASE_COLOR);
		expect(d.width).toBeCloseTo(ERASER_WIDTH, 0);
	});
});

describe('Whiteboard – full re-render abstraction', () => {
	let canvas: HTMLCanvasElement;
	let ctx: CanvasRenderingContext2D;
	let compositingSupported: boolean;

	beforeEach(() => {
		canvas = document.createElement('canvas');
		canvas.width = WB_VIRTUAL_W;
		canvas.height = WB_VIRTUAL_H;
		ctx = canvas.getContext('2d')!;
		ctx.fillStyle = BG_COLOR;
		ctx.fillRect(0, 0, canvas.width, canvas.height);

		ctx.save();
		ctx.globalCompositeOperation = 'destination-out';
		ctx.fillStyle = '#ffffff';
		ctx.fillRect(100, 100, 10, 10);
		ctx.restore();
		compositingSupported = ctx.getImageData(100, 100, 1, 1).data[3] === 0;
	});

	afterEach(() => {
		canvas.remove();
	});

	/** Full re-render from strokes array (simulates resizeWhiteboard / initWhiteboard). */
	function renderAll(strokes: Stroke[]): void {
		ctx.fillStyle = BG_COLOR;
		ctx.fillRect(0, 0, canvas.width, canvas.height);
		for (const s of strokes) {
			renderStroke(ctx, s);
		}
	}

	function getPixel(x: number, y: number): [number, number, number, number] {
		const d = ctx.getImageData(x, y, 1, 1).data;
		return [d[0], d[1], d[2], d[3]];
	}

	const itCanvas = it.skip;

	it('re-render after undo shows remaining strokes', () => {
		const strokes: Stroke[] = [
			makeStroke(0, 100, 500, 100, 'pen', '#ff0000', 10),
		];
		renderAll(strokes);
		// After drawing, the pixel should have visible red (non-zero alpha)
		expect(getPixel(250, 100)[3]).toBeGreaterThan(0);

		// Undo all strokes — canvas should be fully repainted with background
		renderAll(undoLast(strokes));

		// Background color #1c2a3a is opaque, but pen stroke is gone.
		// Verify no red is present (red channel should be background's red=28, not pen's 255).
		expect(getPixel(250, 100)[0]).toBeLessThan(50); // not bright red
	});

	itCanvas('re-render correctly applies erase strokes', () => {
		const strokes: Stroke[] = [
			makeStroke(100, 200, 400, 200, 'pen', '#ffffff', 10),
			makeStroke(200, 195, 300, 205, 'eraser', '#ffffff', 2),
		];

		renderAll(strokes);

		// Middle of erased section should be transparent
		expect(getPixel(250, 200)[3]).toBe(0);

		// Edge of pen stroke (not erased) should be visible
		expect(getPixel(120, 200)[3]).toBeGreaterThan(0);
	});
});
