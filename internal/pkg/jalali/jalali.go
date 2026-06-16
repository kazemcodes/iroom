package pkg

import (
	"fmt"
	"time"
)

func GregorianToJalali(year, month, day int) (int, int, int) {
	gy, gm, gd := year, month, day

	g_d_m := [12]int{0, 31, 59, 90, 120, 151, 181, 212, 243, 273, 304, 334}
	divider := 0
	leap := 0
	jy := 0
	jm := 0
	jd := 0

	divider = (gy - 1600) / 4 - (gy - 1600) / 100 + (gy - 1600) / 400
	leap = 0
	if gy%4 == 0 {
		if gy%100 != 0 || gy%400 == 0 {
			leap = 1
		}
	}

	jy = 979 + 33*divider + 8*(divider/4)
	if gm > 2 {
		jd += leap
	}
	jd += g_d_m[gm-1] + gd - 1

	divider = jy / 100
	jd += 8*(divider-2) + divider/4
	jy += 11*divider + 14 - (divider/4)*8

	for jm = 1; jm < 7; jm++ {
		if jd < 29+(jm-6)/2 {
			break
		}
		jd -= 29 + (jm-6)/2
	}

	if jm > 12 {
		jm -= 12
		jy++
	}

	if jd < 1 {
		jm--
		if jm < 1 {
			jm = 12
			jy--
		}
		jd += 29
	}

	return jy, jm, jd
}

func JalaliMonthName(month int) string {
	names := [13]string{"", "فروردین", "اردیبهشت", "خرداد", "تیر", "مرداد", "شهریور", "مهر", "آبان", "آذر", "دی", "بهمن", "اسفند"}
	if month < 1 || month > 12 {
		return ""
	}
	return names[month]
}

func FormatJalali(t time.Time) string {
	y, m, d := GregorianToJalali(t.Year(), int(t.Month()), t.Day())
	return fmt.Sprintf("%d %s %d", d, JalaliMonthName(m), y)
}
