package analyzer

// Suggestion represents a structured recommendation generated from metrics.
// It replaces the old plain string suggestions.
// Fields are designed to be displayed in reports and to carry enough metadata
// for potential future formatting/logic.
type Suggestion struct {
    // Summary is a short actionable sentence (e.g., "Introduce façade for community X").
    Summary string
    // Location is the subject (community name/ID, file path, class or method name).
    Location string
    // Why explains the metrics and thresholds that led to the suggestion.
    Why string
    // DetailedExplanation provides one or more paragraphs with concrete guidance.
    DetailedExplanation string
}
