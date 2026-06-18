/**
 * Persian Utilities — Formatting functions for Iranian/Persian locale.
 *
 * Functions:
 *   - toPersianNum(n): Convert digits to Persian (e.g. 123 → ۱۲۳)
 *   - toPersianDate(date): Format as YYYY/MM/DD (e.g. ۱۴۰۳/۰۱/۱۵)
 *   - toPersianDateTime(date): Format as YYYY/MM/DD HH:MM
 *   - toPersianDuration(seconds): Format as HH:MM:SS
 *
 * Note: These use JS Date formatting, NOT actual Jalali calendar.
 * For true Jalali conversion, use date-fns-jalali package.
 */
const PERSIAN_DIGITS = ['۰', '۱', '۲', '۳', '۴', '۵', '۶', '۷', '۸', '۹'];

export function toPersianNum(n: number | string): string {
	return String(n).replace(/\d/g, (d) => PERSIAN_DIGITS[parseInt(d)]);
}

function toJsDate(date: Date | string): Date {
	return typeof date === 'string' ? new Date(date) : date;
}

export function toPersianDate(date: Date | string): string {
	const d = toJsDate(date);
	const y = toPersianNum(d.getFullYear());
	const m = toPersianNum(String(d.getMonth() + 1).padStart(2, '0'));
	const day = toPersianNum(String(d.getDate()).padStart(2, '0'));
	return `${y}/${m}/${day}`;
}

export function toPersianDateTime(date: Date | string): string {
	const d = toJsDate(date);
	const datePart = toPersianDate(d);
	const timePart = toPersianTime(d);
	return `${datePart} - ${timePart}`;
}

export function toPersianTime(date: Date | string): string {
	const d = toJsDate(date);
	const h = toPersianNum(String(d.getHours()).padStart(2, '0'));
	const m = toPersianNum(String(d.getMinutes()).padStart(2, '0'));
	const s = toPersianNum(String(d.getSeconds()).padStart(2, '0'));
	return `${h}:${m}:${s}`;
}

export function formatDuration(seconds: number): string {
	const h = Math.floor(seconds / 3600);
	const m = Math.floor((seconds % 3600) / 60);
	const s = Math.floor(seconds % 60);
	return toPersianNum(`${String(h).padStart(2, '0')}:${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`);
}
