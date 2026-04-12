package classifier

type Prediction struct {
	Label       string  `json:"label"`
	Probability float64 `json:"probability"`
	ClassIndex  int     `json:"-"` // Internal use only, not exported to JSON
}

type ClassPrediction struct {
	File        string       `json:"file"`
	Class       string       `json:"class"`
	Predictions []Prediction `json:"predictions"`
}
