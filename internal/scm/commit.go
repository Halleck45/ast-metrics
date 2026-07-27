package scm

type Commit struct {
	Hash string
	// Author is the display name recorded by git (%an)
	Author string
	// Email is the author address (%ae). It is only used to resolve an avatar,
	// and stays empty on repositories whose log does not expose it.
	Email     string
	Timestamp int
	Files     []string
}
