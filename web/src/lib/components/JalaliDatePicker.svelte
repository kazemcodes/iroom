<script lang="ts">
	import { toPersianNum } from '$lib/utils/persian';
	import { toJalali, toGregorian, daysInJalaliMonth } from '$lib/utils/jalali';

	let {
		value = $bindable(''),
		label = ''
	}: { value: string; label?: string } = $props();

	// Parse the ISO value into Jalali year/month/day
	function parseIsoToJalali(iso: string): { jy: number; jm: number; jd: number } | null {
		if (!iso) return null;
		const d = new Date(iso + 'T00:00:00');
		if (isNaN(d.getTime())) return null;
		return toJalali(d.getFullYear(), d.getMonth() + 1, d.getDate());
	}

	// Current view state (which month/year the picker is showing)
	let viewJy = $state(new Date().getFullYear() - 621); // rough Jalali year
	let viewJm = $state(1);
	let isOpen = $state(false);

	// Initialize view from value
	$effect(() => {
		const parsed = parseIsoToJalali(value);
		if (parsed) {
			viewJy = parsed.jy;
			viewJm = parsed.jm;
		}
	});

	const MONTH_NAMES = [
		'فروردین',
		'اردیبهشت',
		'خرداد',
		'تیر',
		'مرداد',
		'شهریور',
		'مهر',
		'آبان',
		'آذر',
		'دی',
		'بهمن',
		'اسفند'
	];

	const DAY_NAMES = ['ش', 'ی', 'د', 'س', 'چ', 'پ', 'ج'];

	// Get the day of week (0=Saturday, 1=Sunday, ..., 6=Friday) for the 1st of a Jalali month
	function firstDayOfWeek(jy: number, jm: number): number {
		const greg = toGregorian(jy, jm, 1);
		const d = new Date(greg.gy, greg.gm - 1, greg.gd);
		// JS getDay(): 0=Sunday, 1=Monday, ..., 6=Saturday
		// We want: 0=Saturday, 1=Sunday, ..., 6=Friday
		return (d.getDay() + 1) % 7;
	}

	// Build the calendar grid
	interface CalendarDay {
		day: number;
		isCurrentMonth: boolean;
		iso: string;
		isToday: boolean;
		isSelected: boolean;
	}

	let calendarDays = $state<CalendarDay[]>([]);

	$effect(() => {
		const days: CalendarDay[] = [];
		const today = new Date();
		const todayJalali = toJalali(today.getFullYear(), today.getMonth() + 1, today.getDate());

		const totalDays = daysInJalaliMonth(viewJy, viewJm);
		const startDow = firstDayOfWeek(viewJy, viewJm);

		// Previous month days
		const prevJm = viewJm === 1 ? 12 : viewJm - 1;
		const prevJy = viewJm === 1 ? viewJy - 1 : viewJy;
		const prevTotalDays = daysInJalaliMonth(prevJy, prevJm);

		for (let i = startDow - 1; i >= 0; i--) {
			const d = prevTotalDays - i;
			const g = toGregorian(prevJy, prevJm, d);
			const iso = `${g.gy}-${String(g.gm).padStart(2, '0')}-${String(g.gd).padStart(2, '0')}`;
			days.push({
				day: d,
				isCurrentMonth: false,
				iso,
				isToday: false,
				isSelected: value === iso
			});
		}

		// Current month days
		for (let d = 1; d <= totalDays; d++) {
			const g = toGregorian(viewJy, viewJm, d);
			const iso = `${g.gy}-${String(g.gm).padStart(2, '0')}-${String(g.gd).padStart(2, '0')}`;
			days.push({
				day: d,
				isCurrentMonth: true,
				iso,
				isToday:
					todayJalali.jy === viewJy && todayJalali.jm === viewJm && todayJalali.jd === d,
				isSelected: value === iso
			});
		}

		// Next month days to fill the grid (always show 6 rows = 42 cells)
		const nextJm = viewJm === 12 ? 1 : viewJm + 1;
		const nextJy = viewJy + (viewJm === 12 ? 1 : 0);
		const remaining = 42 - days.length;
		for (let d = 1; d <= remaining; d++) {
			const g = toGregorian(nextJy, nextJm, d);
			const iso = `${g.gy}-${String(g.gm).padStart(2, '0')}-${String(g.gd).padStart(2, '0')}`;
			days.push({
				day: d,
				isCurrentMonth: false,
				iso,
				isToday: false,
				isSelected: value === iso
			});
		}

		calendarDays = days;
	});

	function prevMonth() {
		if (viewJm === 1) {
			viewJm = 12;
			viewJy--;
		} else {
			viewJm--;
		}
	}

	function nextMonth() {
		if (viewJm === 12) {
			viewJm = 1;
			viewJy++;
		} else {
			viewJm++;
		}
	}

	function selectDate(iso: string) {
		value = iso;
		isOpen = false;
	}

	function togglePicker() {
		isOpen = !isOpen;
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') {
			isOpen = false;
		}
	}

	// Display value in Persian
	let displayValue = $derived(() => {
		const parsed = parseIsoToJalali(value);
		if (!parsed) return '';
		return `${toPersianNum(parsed.jy)}/${toPersianNum(String(parsed.jm).padStart(2, '0'))}/${toPersianNum(String(parsed.jd).padStart(2, '0'))}`;
	});

	// Year range for selector
	let yearOptions = $derived(Array.from({ length: 100 }, (_, i) => viewJy - 50 + i));
</script>

<div class="jalali-datepicker" dir="rtl" onkeydown={handleKeydown}>
	{#if label}
		<label class="block text-sm font-semibold text-gray-700 mb-1.5">{label}</label>
	{/if}

	<button
		type="button"
		class="jalali-input"
		onclick={togglePicker}
		aria-haspopup="dialog"
		aria-expanded={isOpen}
	>
		<span class="jalali-input-text" class:placeholder={!value}>
			{displayValue() || 'انتخاب تاریخ'}
		</span>
		<svg class="jalali-icon" viewBox="0 0 20 20" fill="currentColor">
			<path
				fill-rule="evenodd"
				d="M6 2a1 1 0 00-1 1v1H4a2 2 0 00-2 2v10a2 2 0 002 2h12a2 2 0 002-2V6a2 2 0 00-2-2h-1V3a1 1 0 10-2 0v1H7V3a1 1 0 00-1-1zm0 5a1 1 0 000 2h8a1 1 0 100-2H6z"
				clip-rule="evenodd"
			/>
		</svg>
	</button>

	{#if isOpen}
		<div class="jalali-popup" role="dialog" aria-label="انتخاب تاریخ شمسی">
			<!-- Header: Navigation -->
			<div class="jalali-header">
				<button type="button" class="jalali-nav-btn" onclick={nextMonth} aria-label="ماه بعد">
					<svg viewBox="0 0 20 20" fill="currentColor" width="18" height="18">
						<path
							fill-rule="evenodd"
							d="M12.707 5.293a1 1 0 010 1.414L9.414 10l3.293 3.293a1 1 0 01-1.414 1.414l-4-4a1 1 0 010-1.414l4-4a1 1 0 011.414 0z"
							clip-rule="evenodd"
						/>
					</svg>
				</button>

				<div class="jalali-header-center">
					<select
						class="jalali-year-select"
						bind:value={viewJy}
						aria-label="سال"
					>
						{#each yearOptions as y}
							<option value={y}>{toPersianNum(y)}</option>
						{/each}
					</select>
					<span class="jalali-month-name">{MONTH_NAMES[viewJm - 1]}</span>
				</div>

				<button type="button" class="jalali-nav-btn" onclick={prevMonth} aria-label="ماه قبل">
					<svg viewBox="0 0 20 20" fill="currentColor" width="18" height="18">
						<path
							fill-rule="evenodd"
							d="M7.293 14.707a1 1 0 010-1.414L10.586 10 7.293 6.707a1 1 0 011.414-1.414l4 4a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0z"
							clip-rule="evenodd"
						/>
					</svg>
				</button>
			</div>

			<!-- Day names -->
			<div class="jalali-day-names">
				{#each DAY_NAMES as dn}
					<span class="jalali-day-name">{dn}</span>
				{/each}
			</div>

			<!-- Calendar grid -->
			<div class="jalali-grid">
				{#each calendarDays as day, i}
					<button
						type="button"
						class="jalali-day"
						class:other-month={!day.isCurrentMonth}
						class:today={day.isToday}
						class:selected={day.isSelected}
						onclick={() => selectDate(day.iso)}
						aria-label={toPersianNum(day.day)}
					>
						{toPersianNum(day.day)}
					</button>
				{/each}
			</div>

			<!-- Footer -->
			<div class="jalali-footer">
				<button
					type="button"
					class="jalali-today-btn"
					onclick={() => {
						const today = new Date();
						const j = toJalali(today.getFullYear(), today.getMonth() + 1, today.getDate());
						const g = toGregorian(j.jy, j.jm, j.jd);
						selectDate(
							`${g.gy}-${String(g.gm).padStart(2, '0')}-${String(g.gd).padStart(2, '0')}`
						);
					}}
				>
					امروز
				</button>
				<button
					type="button"
					class="jalali-clear-btn"
					onclick={() => {
						value = '';
						isOpen = false;
					}}
				>
					پاک کردن
				</button>
			</div>
		</div>
	{/if}
</div>

<style>
	.jalali-datepicker {
		position: relative;
		font-family: 'Vazirmatn', system-ui, sans-serif;
	}

	.jalali-input {
		width: 100%;
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 0.75rem 1rem;
		border: 1.5px solid #e2e8f0;
		border-radius: 0.75rem;
		font-size: 0.875rem;
		background: #f8fafc;
		cursor: pointer;
		transition: all 0.2s ease;
		outline: none;
		font-family: 'Vazirmatn', system-ui, sans-serif;
	}

	.jalali-input:hover {
		border-color: #4361ee;
	}

	.jalali-input:focus {
		border-color: #4361ee;
		box-shadow: 0 0 0 3px rgba(67, 97, 238, 0.15);
		background: white;
	}

	.jalali-input-text {
		color: #1a1a2e;
	}

	.jalali-input-text.placeholder {
		color: #94a3b8;
	}

	.jalali-icon {
		width: 1.25rem;
		height: 1.25rem;
		color: #64748b;
		flex-shrink: 0;
	}

	.jalali-popup {
		position: absolute;
		top: calc(100% + 4px);
		right: 0;
		z-index: 100;
		width: 300px;
		background: #1a1a2e;
		border-radius: 1rem;
		box-shadow: 0 20px 40px rgba(0, 0, 0, 0.3), 0 0 0 1px rgba(255, 255, 255, 0.05);
		overflow: hidden;
		animation: jalaliFadeIn 0.15s ease;
	}

	@keyframes jalaliFadeIn {
		from {
			opacity: 0;
			transform: translateY(-4px) scale(0.98);
		}
		to {
			opacity: 1;
			transform: translateY(0) scale(1);
		}
	}

	.jalali-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 0.75rem 1rem;
		background: #16213e;
		border-bottom: 1px solid rgba(255, 255, 255, 0.06);
	}

	.jalali-nav-btn {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 2rem;
		height: 2rem;
		border-radius: 0.5rem;
		border: none;
		background: transparent;
		color: #a0aec0;
		cursor: pointer;
		transition: all 0.15s ease;
	}

	.jalali-nav-btn:hover {
		background: rgba(67, 97, 238, 0.2);
		color: #4361ee;
	}

	.jalali-header-center {
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}

	.jalali-month-name {
		font-size: 0.875rem;
		font-weight: 600;
		color: #e2e8f0;
		white-space: nowrap;
	}

	.jalali-year-select {
		background: #2a2a4a;
		color: #e2e8f0;
		border: 1px solid rgba(255, 255, 255, 0.1);
		border-radius: 0.5rem;
		padding: 0.25rem 0.5rem;
		font-size: 0.875rem;
		font-family: 'Vazirmatn', system-ui, sans-serif;
		cursor: pointer;
		outline: none;
	}

	.jalali-year-select:focus {
		border-color: #4361ee;
	}

	.jalali-day-names {
		display: grid;
		grid-template-columns: repeat(7, 1fr);
		padding: 0.5rem 0.5rem 0;
		gap: 2px;
	}

	.jalali-day-name {
		text-align: center;
		font-size: 0.75rem;
		font-weight: 600;
		color: #718096;
		padding: 0.25rem 0;
	}

	.jalali-grid {
		display: grid;
		grid-template-columns: repeat(7, 1fr);
		padding: 0.25rem 0.5rem 0.5rem;
		gap: 2px;
	}

	.jalali-day {
		display: flex;
		align-items: center;
		justify-content: center;
		aspect-ratio: 1;
		border-radius: 0.5rem;
		border: none;
		background: transparent;
		color: #e2e8f0;
		font-size: 0.8125rem;
		font-family: 'Vazirmatn', system-ui, sans-serif;
		cursor: pointer;
		transition: all 0.1s ease;
	}

	.jalali-day:hover {
		background: rgba(67, 97, 238, 0.2);
		color: #fff;
	}

	.jalali-day.other-month {
		color: #4a5568;
	}

	.jalali-day.today {
		border: 1px solid #4361ee;
		color: #4361ee;
		font-weight: 600;
	}

	.jalali-day.selected {
		background: #4361ee;
		color: white;
		font-weight: 600;
	}

	.jalali-day.selected.today {
		border-color: rgba(255, 255, 255, 0.3);
	}

	.jalali-footer {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 0.5rem 1rem 0.75rem;
		border-top: 1px solid rgba(255, 255, 255, 0.06);
	}

	.jalali-today-btn {
		padding: 0.375rem 0.75rem;
		border-radius: 0.5rem;
		border: 1px solid rgba(67, 97, 238, 0.4);
		background: transparent;
		color: #4361ee;
		font-size: 0.75rem;
		font-weight: 600;
		font-family: 'Vazirmatn', system-ui, sans-serif;
		cursor: pointer;
		transition: all 0.15s ease;
	}

	.jalali-today-btn:hover {
		background: rgba(67, 97, 238, 0.15);
	}

	.jalali-clear-btn {
		padding: 0.375rem 0.75rem;
		border-radius: 0.5rem;
		border: none;
		background: transparent;
		color: #718096;
		font-size: 0.75rem;
		font-weight: 500;
		font-family: 'Vazirmatn', system-ui, sans-serif;
		cursor: pointer;
		transition: all 0.15s ease;
	}

	.jalali-clear-btn:hover {
		color: #e53e3e;
	}
</style>
