package namer

import (
	"strings"
	"unicode"
)

func lastAfterBackslash(s string) string {
	if i := strings.LastIndexByte(s, '\\'); i >= 0 && i+1 < len(s) {
		return s[i+1:]
	}
	return s
}

// splitIdentifier splits "HTTPClientFactory2" into ["HTTP", "Client", "Factory", "2"].
func splitIdentifier(s string) []string {
	if s == "" {
		return nil
	}

	var parts []string
	var b strings.Builder
	b.Grow(len(s))

	flush := func() {
		if b.Len() > 0 {
			parts = append(parts, b.String())
			b.Reset()
		}
	}

	runes := []rune(s)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		prev := rune(0)
		next := rune(0)
		if i > 0 {
			prev = runes[i-1]
		}
		if i+1 < len(runes) {
			next = runes[i+1]
		}

		if i > 0 {
			switch {
			case isLetter(prev) && isDigit(r):
				flush()
			case isDigit(prev) && isLetter(r):
				flush()
			case unicode.IsLower(prev) && unicode.IsUpper(r):
				flush()
			case unicode.IsUpper(prev) && unicode.IsUpper(r) && unicode.IsLower(next):
				flush()
			}
		}

		b.WriteRune(r)
	}
	flush()

	return parts
}

func title(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	for i := 1; i < len(r); i++ {
		r[i] = unicode.ToLower(r[i])
	}
	return string(r)
}

func isDigit(r rune) bool { return r >= '0' && r <= '9' }
func isLetter(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || unicode.IsLetter(r)
}
