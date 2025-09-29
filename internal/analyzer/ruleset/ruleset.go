package ruleset

type Ruleset interface {
	Category() string
	Description() string
	All() []Rule
	Enabled() []Rule
	IsEnabled() bool
}
