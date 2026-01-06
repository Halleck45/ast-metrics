package namer

import "strings"

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
		return title(bestNS)
	}
	return ""
}
