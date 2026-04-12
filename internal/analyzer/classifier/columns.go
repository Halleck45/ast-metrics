package classifier

import "fmt"

// IMPORTANT: must match the order returned by FeatureExtractor.ExtractClassMetrics
var extractorRowColumns = []string{
	"stmt_name",                // 0
	"stmt_type",                // 1
	"file_path",                // 2
	"method_calls_raw",         // 3
	"uses_raw",                 // 4
	"namespace_raw",            // 5
	"path_raw",                 // 6
	"class_loc",                // 7
	"logical_loc",              // 8
	"comment_loc",              // 9
	"nb_comments",              // 10
	"nb_methods",               // 11
	"nb_extends",               // 12
	"nb_implements",            // 13
	"nb_traits",                // 14
	"count_if",                 // 15
	"count_elseif",             // 16
	"count_else",               // 17
	"count_case",               // 18
	"count_switch",             // 19
	"count_loop",               // 20
	"nb_external_dependencies", // 21
	"depth_estimate",           // 22
	"nb_method_calls",          // 23
	"nb_getters",               // 24
	"nb_setters",               // 25
	"nb_attributes",            // 26
	"nb_unique_operators",      // 27
	"programming_language",     // 28
	"cyclomatic_complexity",    // 29
}

var colIndex = func() map[string]int {
	m := make(map[string]int, len(extractorRowColumns))
	for i, c := range extractorRowColumns {
		m[c] = i
	}
	return m
}()

func columnIndex(col string) int {
	i, ok := colIndex[col]
	if !ok {
		return -1
	}
	return i
}

func mustRowGet(row []string, col string) (string, error) {
	i := columnIndex(col)
	if i < 0 || i >= len(row) {
		return "", fmt.Errorf("row missing column %q", col)
	}
	return row[i], nil
}
