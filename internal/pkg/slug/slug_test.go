package slug

import (
	"testing"
)

func TestGenerate_Basic(t *testing.T) {
	got := Generate("Hello World")
	if got != "hello-world" {
		t.Errorf("expected 'hello-world', got '%s'", got)
	}
}

func TestGenerate_PersianDigits(t *testing.T) {
	got := Generate("کلاس ۱۲۳")
	if got != "کلاس-123" {
		t.Errorf("expected 'کلاس-123', got '%s'", got)
	}
}

func TestGenerate_SpecialChars(t *testing.T) {
	got := Generate("Test!@#$%Name")
	if got != "testname" {
		t.Errorf("expected 'testname', got '%s'", got)
	}
}

func TestGenerate_EmptyFallback(t *testing.T) {
	got := Generate("!@#$%")
	if got == "" {
		t.Error("expected non-empty slug for special-char-only input")
	}
	if len(got) < 5 {
		t.Errorf("fallback slug too short: '%s'", got)
	}
}

func TestGenerate_HyphenCollapsing(t *testing.T) {
	got := Generate("a   b")
	if got != "a-b" {
		t.Errorf("expected 'a-b', got '%s'", got)
	}
}

func TestGenerate_TrimHyphens(t *testing.T) {
	got := Generate(" test ")
	if got != "test" {
		t.Errorf("expected 'test', got '%s'", got)
	}
}

func TestGenerate_PersianText(t *testing.T) {
	got := Generate("ریاضی پایه دهم")
	if got == "" {
		t.Error("expected non-empty slug for Persian text")
	}
}
