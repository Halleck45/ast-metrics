package ruleset

import pb "github.com/ast-metrics/ast-metrics/pb"

// lineOf returns the 1-based start line of a location, or 0 when the location
// is unavailable. Rules use it to anchor a violation to the offending class or
// method, so downstream reports (e.g. SARIF) can point at the exact line.
func lineOf(loc *pb.StmtLocationInFile) int {
	if loc == nil {
		return 0
	}
	return int(loc.GetStartLine())
}
