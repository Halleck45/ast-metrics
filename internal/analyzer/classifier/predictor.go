package classifier

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/halleck45/ast-metrics/internal/engine"
	pb "github.com/halleck45/ast-metrics/pb"
	"github.com/pterm/pterm"
	log "github.com/sirupsen/logrus"
)

type Predictor struct {
	ModelDir string
}

func NewPredictor(modelDir string) *Predictor {
	return &Predictor{ModelDir: modelDir}
}

func (p *Predictor) Predict(files []*pb.File, workdir string) ([]ClassPrediction, error) {
	filesByLang := make(map[string][]*pb.File)
	for _, f := range files {
		if f == nil {
			continue
		}
		if f.ProgrammingLanguage != "" {
			filesByLang[f.ProgrammingLanguage] = append(filesByLang[f.ProgrammingLanguage], f)
		}
	}

	var allPredictions []ClassPrediction
	extractor := NewFeatureExtractor()

	for lang, langFiles := range filesByLang {
		langModelDir := filepath.Join(p.ModelDir, strings.ToLower(lang))
		modelFile := filepath.Join(langModelDir, "model.json")

		model, err := LoadModel(modelFile)
		if err != nil {
			continue
		}

		// Required metadata for hashing path
		if model.Features.HashingNFeatures <= 0 {
			continue
		}
		if len(model.Features.NumericalColsOrder) == 0 {
			continue
		}
		if _, ok := model.Vectorizers["class_name"]; !ok {
			continue
		}
		for _, col := range model.Features.NlpCols {
			if _, ok := model.Vectorizers[col]; !ok {
				continue
			}
		}

		pterm.Info.Printf("Classifying %s classes using local AI", lang)

		// Reusable hashing buffer to avoid per-row allocations
		hashBuf := make([]float64, model.Features.HashingNFeatures)

		for _, pbFile := range langFiles {
			classes := engine.GetClassesInFile(pbFile)
			for _, class := range classes {
				row := extractor.ExtractClassMetrics(class, pbFile)

				features, err := p.rowToFeatures(row, model, hashBuf)
				if err != nil {
					log.Errorf("Failed to convert row to features: %v", err)
					continue
				}

				preds, err := model.Predict(features)
				if err != nil {
					log.Errorf("Failed to predict: %v", err)
					continue
				}

				// Map class_idx -> label string
				for i := range preds {
					key := fmt.Sprintf("%d", preds[i].ClassIndex)
					if label, ok := model.Features.LabelMapping[key]; ok && label != "" {
						preds[i].Label = label
					} else {
						preds[i].Label = fmt.Sprintf("UNKNOWN(class_idx=%d)", preds[i].ClassIndex)
					}
				}

				className := ""
				if class != nil && class.Name != nil {
					className = class.Name.Qualified
					if className == "" {
						className = class.Name.Short
					}
				}

				filePath := ""
				if pbFile != nil {
					filePath = pbFile.ShortPath
					if filePath == "" {
						filePath = pbFile.Path
					}
				}

				allPredictions = append(allPredictions, ClassPrediction{
					File:        filePath,
					Class:       className,
					Predictions: preds,
				})
			}
		}
	}

	return allPredictions, nil
}

// rowToFeatures builds the final feature vector in the exact same block order as Python training:
//
// [A] X_numerical (numeric + categorical label-encoded) in order model.Features.NumericalColsOrder
// [B] programming_language one-hot (weighted)
// [C] class_name hashing tf-idf (weighted) size = HashingNFeatures
// [D] NLP cols hashing tf-idf in model.Features.NlpCols order (weight=1.0 each) size = HashingNFeatures
func (p *Predictor) rowToFeatures(row []string, model *RandomForestModel, hashBuf []float64) ([]float64, error) {
	if model == nil {
		return nil, fmt.Errorf("model is nil")
	}
	meta := model.Features
	nHash := meta.HashingNFeatures
	if nHash <= 0 {
		return nil, fmt.Errorf("hashing_n_features missing/invalid")
	}
	if len(meta.NumericalColsOrder) == 0 {
		return nil, fmt.Errorf("numerical_cols_order missing")
	}
	if len(hashBuf) != nHash {
		return nil, fmt.Errorf("hashBuf size mismatch: got=%d expected=%d", len(hashBuf), nHash)
	}

	// Sizes and pre-allocation
	nNumerical := len(meta.NumericalColsOrder)
	nProgLang := 0
	if model.ProgrammingLanguageEncoder != nil && len(model.ProgrammingLanguageEncoder.Categories) > 0 {
		nProgLang = len(model.ProgrammingLanguageEncoder.Categories[0])
	}
	total := nNumerical + nProgLang + (1+len(meta.NlpCols))*nHash
	features := make([]float64, total)

	off := 0

	// Fast categorical membership
	catSet := make(map[string]struct{}, len(meta.CategoricalCols))
	for _, c := range meta.CategoricalCols {
		catSet[c] = struct{}{}
	}

	// [A] X_numerical
	for _, col := range meta.NumericalColsOrder {
		// Try derived feature first (computed from other raw columns)
		if val, ok := computeDerivedFeature(row, col); ok {
			features[off] = val
			off++
			continue
		}

		raw := rowValue(row, col)

		// categorical
		if _, isCat := catSet[col]; isCat {
			enc, ok := model.Encoders[col]
			if !ok {
				features[off] = -1
				off++
				continue
			}
			features[off] = float64(indexOfString(enc, raw, -1))
			off++
			continue
		}

		// numeric
		if raw == "" {
			features[off] = 0
			off++
			continue
		}
		v, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			v = 0
		}
		features[off] = v
		off++
	}

	// [B] programming_language one-hot
	if nProgLang > 0 {
		lang := rowValue(row, "programming_language")
		oneHot := model.ProgrammingLanguageEncoder.Transform(lang)

		w := meta.ProgrammingLanguageWeight
		if w == 0 {
			w = 2.5
		}
		if len(oneHot) != nProgLang {
			return nil, fmt.Errorf("programming_language onehot size mismatch: got=%d expected=%d", len(oneHot), nProgLang)
		}
		for i := 0; i < nProgLang; i++ {
			features[off+i] = oneHot[i] * w
		}
		off += nProgLang
	}

	// [C] class_name hashing (derived from namespace_raw)
	namespace := rowValue(row, "namespace_raw")
	className := extractClassName(namespace)

	vecCN, ok := model.Vectorizers["class_name"]
	if !ok {
		return nil, fmt.Errorf("missing vectorizer: class_name")
	}
	wcn := meta.ClassNameWeight
	if wcn == 0 {
		wcn = 3.0
	}
	if err := vecCN.TransformHashingTfidf(className, hashBuf, wcn); err != nil {
		return nil, fmt.Errorf("class_name hashing: %w", err)
	}
	copy(features[off:off+nHash], hashBuf)
	off += nHash

	// [D] NLP cols hashing in training order (with per-column weights from training)
	for _, col := range meta.NlpCols {
		vec, ok := model.Vectorizers[col]
		if !ok {
			return nil, fmt.Errorf("missing vectorizer: %s", col)
		}
		text := rowValue(row, col)
		nlpWeight := 1.0
		if meta.NlpWeights != nil {
			if w, ok := meta.NlpWeights[col]; ok && w > 0 {
				nlpWeight = w
			}
		}
		if err := vec.TransformHashingTfidf(text, hashBuf, nlpWeight); err != nil {
			return nil, fmt.Errorf("%s hashing: %w", col, err)
		}
		copy(features[off:off+nHash], hashBuf)
		off += nHash
	}

	// sanity
	if model.Meta.NFeatures > 0 && len(features) != model.Meta.NFeatures {
		return nil, fmt.Errorf("feature length mismatch: got=%d expected=%d", len(features), model.Meta.NFeatures)
	}

	return features, nil
}

// rowValue reads a column by name from the extractor row using columnIndex(). Unknown => "".
func rowValue(row []string, col string) string {
	i := columnIndex(col)
	if i < 0 || i >= len(row) {
		return ""
	}
	return row[i]
}

func indexOfString(haystack []string, needle string, defaultIndex int) int {
	for i, v := range haystack {
		if v == needle {
			return i
		}
	}
	return defaultIndex
}

// computeDerivedFeature computes a derived ratio feature from raw row values.
// Returns the value and true if the column is a known derived feature, (0, false) otherwise.
func computeDerivedFeature(row []string, name string) (float64, bool) {
	switch name {
	case "getter_ratio":
		return safeRatio(rowValue(row, "nb_getters"), rowValue(row, "nb_methods")), true
	case "setter_ratio":
		return safeRatio(rowValue(row, "nb_setters"), rowValue(row, "nb_methods")), true
	case "loc_per_method":
		return safeRatio(rowValue(row, "class_loc"), rowValue(row, "nb_methods")), true
	case "complexity_per_method":
		return safeRatio(rowValue(row, "cyclomatic_complexity"), rowValue(row, "nb_methods")), true
	case "attribute_per_method":
		return safeRatio(rowValue(row, "nb_attributes"), rowValue(row, "nb_methods")), true
	case "comment_ratio":
		return safeRatio(rowValue(row, "comment_loc"), rowValue(row, "class_loc")), true
	case "method_call_density":
		return safeRatio(rowValue(row, "nb_method_calls"), rowValue(row, "nb_methods")), true
	case "dep_per_method":
		return safeRatio(rowValue(row, "nb_external_dependencies"), rowValue(row, "nb_methods")), true
	}
	return 0, false
}

func safeRatio(numStr, denomStr string) float64 {
	num, _ := strconv.ParseFloat(numStr, 64)
	denom, _ := strconv.ParseFloat(denomStr, 64)
	if denom == 0 {
		denom = 1
	}
	v := num / denom
	if v > 100 {
		v = 100
	}
	if v < 0 {
		v = 0
	}
	return v
}

func extractClassName(namespace string) string {
	if namespace == "" {
		return ""
	}
	if strings.Contains(namespace, "\\") {
		parts := strings.Split(namespace, "\\")
		return parts[len(parts)-1]
	}
	if strings.Contains(namespace, ".") {
		parts := strings.Split(namespace, ".")
		return parts[len(parts)-1]
	}
	return namespace
}
