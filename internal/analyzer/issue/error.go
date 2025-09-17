package issue

// Severity represents the criticité of a rule outcome
// Values are indicative; when not detected, Unknown is used.
type Severity string

const (
	SeverityUnknown Severity = "unknown"
	SeverityLow     Severity = "low"
	SeverityMedium  Severity = "medium"
	SeverityHigh    Severity = "high"
)

// RequirementError is a structured error produced by rules
// Message: human-readable description (without severity tag prefix)
// Code: a stable code identifying error category (usually rule name)
// Severity: criticité
type RequirementError struct {
	Message  string
	Code     string
	Severity Severity
}
