package slug

import (
	"fmt"
	"strings"
	"time"
	"unicode"
)

func Generate(name string) string {
	s := strings.ToLower(name)
	replacements := map[string]string{
		" ": "-", "‌": "", "۰": "0", "۱": "1", "۲": "2", "۳": "3",
		"۴": "4", "۵": "5", "۶": "6", "۷": "7", "۸": "8", "۹": "9",
	}
	for k, v := range replacements {
		s = strings.ReplaceAll(s, k, v)
	}
	var result []rune
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' {
			result = append(result, r)
		}
	}
	s = string(result)
	for strings.Contains(s, "--") {
		s = strings.ReplaceAll(s, "--", "-")
	}
	s = strings.Trim(s, "-")
	if s == "" {
		s = fmt.Sprintf("room-%d", time.Now().UnixMilli())
	}
	return s
}
