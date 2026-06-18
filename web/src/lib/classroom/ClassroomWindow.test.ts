import { describe, it, expect, vi, beforeEach } from 'vitest';

const mockWindow = {
	closed: false,
	close: vi.fn(),
	focus: vi.fn(),
};

vi.stubGlobal('window', {
	...window,
	open: vi.fn(() => mockWindow),
	addEventListener: vi.fn(),
	dispatchEvent: vi.fn(),
});

describe('classroomWindow', async () => {
	const { classroomWindow } = await import('$lib/classroom/ClassroomWindow');

	beforeEach(() => {
		vi.clearAllMocks();
		mockWindow.closed = false;
	});

	it('open creates new window', () => {
		const result = classroomWindow.open('1', 'Test Session');
		expect(window.open).toHaveBeenCalledWith('/classroom/popup/1', '_blank');
		expect(result).toBe(mockWindow);
	});

	it('close closes the window', () => {
		classroomWindow.open('1', 'Test Session');
		classroomWindow.close('1');
		expect(mockWindow.close).toHaveBeenCalled();
	});

	it('getAll returns opened tabs', () => {
		classroomWindow.open('1', 'Session 1');
		classroomWindow.open('2', 'Session 2');
		const all = classroomWindow.getAll();
		expect(all.size).toBe(2);
	});
});
