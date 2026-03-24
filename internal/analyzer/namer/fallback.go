package namer

import "strings"

var genericSegments = map[string]bool{
	"http": true, "common": true, "util": true, "utils": true,
	"base": true, "core": true, "internal": true, "lib": true,
	"shared": true, "api": true, "src": true, "service": true,
	"services": true, "model": true, "models": true,
}

func fallbackFromNamespaces(classNames []string) string {
	counter := map[string]int{}

	for _, name := range classNames {
		parts := strings.Split(name, "\\")
		if len(parts) > 2 {
			counter[parts[len(parts)-2]]++
		}
	}

	bestNS := ""
	bestCount := 0
	for ns, c := range counter {
		if c > bestCount {
			bestCount = c
			bestNS = ns
		}
	}

	if bestCount > 1 && bestNS != "" {
		// If the best segment covers all classes and is a generic term,
		// skip it so the word2vec path can produce a better name.
		if bestCount == len(classNames) && len(counter) == 1 {
			if genericSegments[strings.ToLower(bestNS)] {
				return ""
			}
		}
		return title(bestNS)
	}
	return ""
}
