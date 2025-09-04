package scm

type Commit struct {
	Hash   string
	Author string
	Timestamp   int
	Files  []string
}
