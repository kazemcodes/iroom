package validate

import (
	"net/mail"
	"strings"
)

type Errors map[string]string

func ValidateEmail(email string) string {
	if strings.TrimSpace(email) == "" {
		return "ایمیل الزامی است"
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return "ایمیل نامعتبر است"
	}
	return ""
}

func ValidatePassword(password string) string {
	if len(password) < 6 {
		return "رمز عبور باید حداقل ۶ کاراکتر باشد"
	}
	return ""
}

func ValidateRequired(value, fieldName string) string {
	if strings.TrimSpace(value) == "" {
		return fieldName + " الزامی است"
	}
	return ""
}
