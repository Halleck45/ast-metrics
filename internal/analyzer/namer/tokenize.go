package namer

import (
	"strings"
)

func (n *Namer) tokenize(classNames []string) []string {
	out := make([]string, 0, len(classNames)*2)
	for _, full := range classNames {
		className := lastAfterBackslash(full)
		for _, tok := range splitIdentifier(className) {
			t := strings.ToLower(tok)
			if n.isValidToken(t) {
				out = append(out, t)
			}
		}
	}
	return out
}

func (n *Namer) isValidToken(t string) bool {
	if len(t) < 3 {
		return false
	}
	if isBlacklisted(t) {
		return false
	}
	_, ok := n.vectors[t]
	return ok
}
