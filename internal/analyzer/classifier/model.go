package classifier

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"math"
	"os"
	"sort"
	"strings"
)

// Model structures
type RandomForestModel struct {
	Meta                       ModelMeta                  `json:"meta"`
	Features                   FeaturesMetadata           `json:"features"`
	Encoders                   map[string][]string        `json:"encoders"`
	Vectorizers                map[string]TfidfVectorizer `json:"vectorizers"`
	ProgrammingLanguageEncoder *OneHotEncoder             `json:"programming_language_encoder"`
	Trees                      []DecisionTree             `json:"trees"`
}

type ModelMeta struct {
	Type      string `json:"type"`
	NFeatures int    `json:"n_features"`
	NClasses  int    `json:"n_classes"`
	Classes   []int  `json:"classes"`
}

type FeaturesMetadata struct {
	FinalFeatureNames         []string       `json:"final_feature_names"`
	CategoricalCols           []string       `json:"categorical_cols"`
	NlpCols                   []string       `json:"nlp_cols"`
	LabelMapping              map[string]int `json:"label_mapping"`
	ClassNameWeight           float64        `json:"class_name_weight"`
	ProgrammingLanguageWeight float64        `json:"programming_language_weight"`
}

type TfidfVectorizer struct {
	Vocabulary  map[string]int `json:"vocabulary"`
	Idf         []float64      `json:"idf"`
	Norm        string         `json:"norm"`
	UseIdf      bool           `json:"use_idf"`
	SmoothIdf   bool           `json:"smooth_idf"`
	SublinearTf bool           `json:"sublinear_tf"`
}

type OneHotEncoder struct {
	Categories [][]string `json:"categories"`
}

type DecisionTree struct {
	ChildrenLeft  []int       `json:"children_left"`
	ChildrenRight []int       `json:"children_right"`
	Feature       []int       `json:"feature"`
	Threshold     []float64   `json:"threshold"`
	Value         [][]float64 `json:"value"`
	NodeCount     int         `json:"node_count"`
}

// LoadModel loads the Random Forest model from a JSON file (supports .json.gz)
func LoadModel(path string) (*RandomForestModel, error) {
	// Try .gz version first
	gzPath := path + ".gz"
	if _, err := os.Stat(gzPath); err == nil {
		return loadModelGzip(gzPath)
	}

	// Fall back to uncompressed
	return loadModelPlain(path)
}

func loadModelGzip(path string) (*RandomForestModel, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return nil, err
	}
	defer gzReader.Close()

	data, err := io.ReadAll(gzReader)
	if err != nil {
		return nil, err
	}

	var model RandomForestModel
	if err := json.Unmarshal(data, &model); err != nil {
		return nil, err
	}

	return &model, nil
}

func loadModelPlain(path string) (*RandomForestModel, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var model RandomForestModel
	if err := json.Unmarshal(data, &model); err != nil {
		return nil, err
	}

	return &model, nil
}

// TransformTfidf transforms text using TF-IDF vectorization
func (v *TfidfVectorizer) Transform(text string) []float64 {
	// Tokenize using Python's default pattern: \b\w\w+\b (words with 2+ chars)
	// For simplicity, we'll use a basic approach: lowercase + split on non-alphanumeric
	text = strings.ToLower(text)

	// Extract tokens (alphanumeric sequences of 2+ characters)
	var tokens []string
	currentToken := ""
	for _, ch := range text {
		if (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') {
			currentToken += string(ch)
		} else {
			if len(currentToken) >= 2 {
				tokens = append(tokens, currentToken)
			}
			currentToken = ""
		}
	}
	if len(currentToken) >= 2 {
		tokens = append(tokens, currentToken)
	}

	// Count term frequencies
	termFreq := make(map[string]int)
	for _, token := range tokens {
		termFreq[token]++
	}

	// Create feature vector
	features := make([]float64, len(v.Vocabulary))

	for term, count := range termFreq {
		if idx, ok := v.Vocabulary[term]; ok && idx < len(features) {
			tf := float64(count)

			// Apply sublinear TF if enabled
			if v.SublinearTf {
				tf = 1.0 + math.Log(tf)
			}

			// Apply IDF if enabled
			if v.UseIdf && idx < len(v.Idf) {
				tf *= v.Idf[idx]
			}

			features[idx] = tf
		}
	}

	// Apply normalization
	if v.Norm == "l2" {
		norm := 0.0
		for _, val := range features {
			norm += val * val
		}
		norm = math.Sqrt(norm)
		if norm > 0 {
			for i := range features {
				features[i] /= norm
			}
		}
	}

	return features
}

// PredictTree predicts using a single decision tree
func (t *DecisionTree) Predict(features []float64) []float64 {
	node := 0

	// Traverse the tree
	for {
		// Check if we're at a leaf node
		if t.ChildrenLeft[node] == -1 && t.ChildrenRight[node] == -1 {
			// Return the class probabilities at this leaf
			return t.Value[node]
		}

		// Get the feature to split on
		featureIdx := t.Feature[node]
		if featureIdx < 0 || featureIdx >= len(features) {
			// Invalid feature index, return current node value
			return t.Value[node]
		}

		// Decide which child to visit
		if features[featureIdx] <= t.Threshold[node] {
			node = t.ChildrenLeft[node]
		} else {
			node = t.ChildrenRight[node]
		}

		// Safety check to prevent infinite loops
		if node < 0 || node >= t.NodeCount {
			break
		}
	}

	// Fallback: return uniform distribution
	nClasses := len(t.Value[0])
	uniform := make([]float64, nClasses)
	for i := range uniform {
		uniform[i] = 1.0 / float64(nClasses)
	}
	return uniform
}

// Predict performs Random Forest prediction
func (m *RandomForestModel) Predict(features []float64) ([]Prediction, error) {
	nTrees := len(m.Trees)
	nClasses := m.Meta.NClasses

	// Aggregate predictions from all trees
	classSums := make([]float64, nClasses)

	for _, tree := range m.Trees {
		probs := tree.Predict(features)
		for i, prob := range probs {
			if i < nClasses {
				classSums[i] += prob
			}
		}
	}

	// Average the probabilities
	for i := range classSums {
		classSums[i] /= float64(nTrees)
	}

	// Get top 3 predictions
	type scoredClass struct {
		classIdx int
		score    float64
	}

	scored := make([]scoredClass, nClasses)
	for i := 0; i < nClasses; i++ {
		scored[i] = scoredClass{classIdx: i, score: classSums[i]}
	}

	// Sort by score descending
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})

	// Take top 3
	topN := 3
	if topN > len(scored) {
		topN = len(scored)
	}

	predictions := make([]Prediction, 0, topN)
	for i := 0; i < topN; i++ {
		if scored[i].score < 0.01 {
			continue
		}
		predictions = append(predictions, Prediction{
			Label:       "", // Will be filled by caller using label mapping
			Probability: scored[i].score,
			ClassIndex:  scored[i].classIdx,
		})
	}

	return predictions, nil
}

// Transform performs one-hot encoding
func (e *OneHotEncoder) Transform(value string) []float64 {
	if e == nil || len(e.Categories) == 0 {
		return []float64{}
	}

	// For now, we assume single feature (categories[0])
	categories := e.Categories[0]
	result := make([]float64, len(categories))

	// Find the value in categories and set corresponding position to 1
	for i, cat := range categories {
		if cat == value {
			result[i] = 1.0
			break
		}
	}

	return result
}
