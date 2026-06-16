/**
 * Jalali (Persian/Shamsi) calendar conversion utilities.
 *
 * Provides bidirectional conversion between Gregorian and Jalali calendars,
 * as well as helpers for days-in-month and leap year checks.
 */

// --- Core conversion constants ---

const g_d_m = [0, 31, 59, 90, 120, 151, 181, 212, 243, 273, 304, 334];

interface GregorianDate {
	gy: number;
	gm: number;
	gd: number;
}

interface JalaliDate {
	jy: number;
	jm: number;
	jd: number;
}

/**
 * Check if a Gregorian year is a leap year.
 */
function isGregorianLeap(gy: number): boolean {
	return (gy % 4 === 0 && gy % 100 !== 0) || gy % 400 === 0;
}

/**
 * Check if a Jalali year is a leap year using the 2820-year cycle algorithm.
 */
export function isJalaliLeap(jy: number): boolean {
	const breaks = [
		1, 5, 9, 13, 17, 22, 26, 30, 34, 38, 43, 47, 51, 55, 59, 63, 67, 71, 75, 79, 83, 88, 92, 96,
		100, 104, 108, 112, 116, 120, 124
	];
	const mod = jy % 128;
	return breaks.includes(mod);
}

/**
 * Get the number of days in a Jalali month (1-12).
 */
export function daysInJalaliMonth(jy: number, jm: number): number {
	if (jm <= 6) return 31;
	if (jm <= 11) return 30;
	return isJalaliLeap(jy) ? 30 : 29;
}

/**
 * Convert a Gregorian date to Jalali.
 */
export function toJalali(gy: number, gm: number, gd: number): JalaliDate {
	const gy2 = gm > 2 ? gy + 1 : gy;
	let days =
		355666 +
		365 * gy +
		Math.floor((gy2 + 3) / 4) -
		Math.floor((gy2 + 99) / 100) +
		Math.floor((gy2 + 399) / 400) +
		gd +
		g_d_m[gm - 1];
	let jy = -1595 + 33 * Math.floor(days / 12053);
	days %= 12053;
	jy += 4 * Math.floor(days / 1461);
	days %= 1461;
	if (days > 365) {
		jy += Math.floor((days - 1) / 365);
		days = (days - 1) % 365;
	}
	let jm: number;
	let jd: number;
	if (days < 186) {
		jm = 1 + Math.floor(days / 31);
		jd = 1 + (days % 31);
	} else {
		jm = 7 + Math.floor((days - 186) / 30);
		jd = 1 + ((days - 186) % 30);
	}
	return { jy, jm, jd };
}

/**
 * Convert a Jalali date to Gregorian.
 */
export function toGregorian(jy: number, jm: number, jd: number): GregorianDate {
	const jyShift = jy + 1595;
	let days =
		-355668 +
		365 * jyShift +
		Math.floor(jyShift / 33) * 8 +
		Math.floor(((jyShift % 33) + 3) / 4) +
		jd +
		(jm <= 7 ? (jm - 1) * 31 : (jm - 7) * 30 + 186);
	let gy = 400 * Math.floor(days / 146097);
	days %= 146097;
	if (days > 36524) {
		gy += 100 * Math.floor(--days / 36524);
		days %= 36524;
		if (days >= 365) days++;
	}
	gy += 4 * Math.floor(days / 1461);
	days %= 1461;
	if (days > 365) {
		gy += Math.floor((days - 1) / 365);
		days = (days - 1) % 365;
	}
	const gd = days + 1;
	const sal_a = [0, 31, 29, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31];
	if (!isGregorianLeap(gy)) sal_a[2] = 28;
	let gm = 1;
	while (gm <= 12 && gd > sal_a[gm]) {
		gm++;
	}
	return { gy, gm, gd: gd - sal_a[gm - 1] };
}

/**
 * Validate a Jalali date.
 */
export function isValidJalaliDate(jy: number, jm: number, jd: number): boolean {
	if (jm < 1 || jm > 12) return false;
	if (jd < 1 || jd > daysInJalaliMonth(jy, jm)) return false;
	return true;
}

/**
 * Get the day of the week for a Jalali date.
 * Returns 0=Saturday, 1=Sunday, ..., 6=Friday
 */
export function jalaliDayOfWeek(jy: number, jm: number, jd: number): number {
	const g = toGregorian(jy, jm, jd);
	const d = new Date(g.gy, g.gm - 1, g.gd);
	// JS getDay(): 0=Sunday, 1=Monday, ..., 6=Saturday
	// We want: 0=Saturday, 1=Sunday, ..., 6=Friday
	return (d.getDay() + 1) % 7;
}
