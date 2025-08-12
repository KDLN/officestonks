package utils

import (
	"strings"
	"testing"
)

func TestSanitizeStringRemovesJavascriptProtocol(t *testing.T) {
	input := "Check this link: JavaScript:alert('xss')"
	sanitized := SanitizeString(input)
	if strings.Contains(strings.ToLower(sanitized), "javascript:") {
		t.Fatalf("sanitized string still contains javascript protocol: %q", sanitized)
	}
}

func TestSanitizeStringRemovesDataProtocol(t *testing.T) {
	input := "Embedded data: DATA:text/html;base64,PHNjcmlwdD5hbGVydCgxKTwvc2NyaXB0Pg=="
	sanitized := SanitizeString(input)
	if strings.Contains(strings.ToLower(sanitized), "data:") {
		t.Fatalf("sanitized string still contains data protocol: %q", sanitized)
	}
}
