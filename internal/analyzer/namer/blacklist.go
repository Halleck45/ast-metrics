package namer

var blacklist = map[string]struct{}{
	"the": {}, "for": {}, "one": {}, "and": {}, "with": {},
	"from": {}, "this": {}, "that": {}, "not": {}, "implicitly": {},
	"has": {}, "but": {}, "are": {}, "have": {}, "all": {},
}

func isBlacklisted(token string) bool {
	_, ok := blacklist[token]
	return ok
}
