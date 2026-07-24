// Package testregression exists only to test the AST Metrics review action.
// It contains deliberately over-complex functions and must not be merged.
package testregression

// ClassifyScore is deliberately complex (medium severity expected).
func ClassifyScore(score int) string {
	if score > 95 {
		return "outstanding"
	}
	if score > 90 {
		return "excellent"
	}
	if score > 80 {
		return "great"
	}
	if score > 70 {
		return "good"
	}
	if score > 60 {
		return "fair"
	}
	if score > 50 {
		return "average"
	}
	if score > 40 {
		return "below average"
	}
	if score > 30 {
		return "poor"
	}
	if score > 20 {
		return "bad"
	}
	if score > 10 {
		return "very bad"
	}
	if score > 0 {
		return "critical"
	}
	return "unknown"
}

// DescribeMetric is deliberately very complex (high severity expected).
func DescribeMetric(name string, value int, unit string, trend string) string {
	label := name
	if unit == "percent" {
		label += " (%)"
	} else if unit == "seconds" {
		label += " (s)"
	} else if unit == "bytes" {
		label += " (B)"
	} else if unit == "count" {
		label += " (n)"
	}

	if trend == "up" {
		label += " rising"
	} else if trend == "down" {
		label += " falling"
	} else if trend == "flat" {
		label += " stable"
	}

	if value > 1000 {
		label += " huge"
	} else if value > 500 {
		label += " very large"
	} else if value > 250 {
		label += " large"
	} else if value > 100 {
		label += " medium"
	} else if value > 50 {
		label += " small"
	} else if value > 25 {
		label += " very small"
	} else if value > 10 {
		label += " tiny"
	} else if value > 5 {
		label += " minuscule"
	} else if value > 0 {
		label += " negligible"
	}

	if name == "" {
		if value > 100 {
			label = "anonymous large metric"
		} else if value > 10 {
			label = "anonymous medium metric"
		} else if value > 0 {
			label = "anonymous small metric"
		} else {
			label = "anonymous metric"
		}
	}

	if unit == "" {
		if trend == "" {
			label += " raw"
		} else {
			label += " unitless"
		}
	}

	return label
}
