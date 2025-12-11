package classifier

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/halleck45/ast-metrics/internal/engine"
	pb "github.com/halleck45/ast-metrics/pb"
	log "github.com/sirupsen/logrus"
)

type Predictor struct {
	ModelDir string
}

func NewPredictor(modelDir string, scriptPath string) *Predictor {
	return &Predictor{
		ModelDir: modelDir,
	}
}

func (p *Predictor) Predict(files []*pb.File, workdir string) ([]ClassPrediction, error) {
	// Group files by language
	filesByLang := make(map[string][]*pb.File)
	for _, f := range files {
		if f.ProgrammingLanguage != "" {
			filesByLang[f.ProgrammingLanguage] = append(filesByLang[f.ProgrammingLanguage], f)
		}
	}

	var allPredictions []ClassPrediction
	extractor := NewFeatureExtractor()

	for lang, langFiles := range filesByLang {
		// Check if model exists for this language
		langModelDir := filepath.Join(p.ModelDir, strings.ToLower(lang))
		modelFile := filepath.Join(langModelDir, "model.json")

		// Load the model (will automatically try .gz first)
		model, err := LoadModel(modelFile)
		if err != nil {
			log.Errorf("Failed to load model for %s: %v", lang, err)
			continue
		}

		// Log which file was actually loaded
		if _, err := os.Stat(modelFile + ".gz"); err == nil {
			log.Infof("Loaded compressed model for %s from %s.gz", lang, modelFile)
		} else {
			log.Infof("Loaded model for %s from %s", lang, modelFile)
		}

		// Process each file
		for _, pbFile := range langFiles {
			classes := engine.GetClassesInFile(pbFile)
			for _, class := range classes {
				// Extract features as CSV row
				row := extractor.ExtractClassMetrics(class, pbFile)

				// Convert row to feature vector matching Python's exact order
				features, err := p.rowToFeatures(row, model)
				if err != nil {
					log.Errorf("Failed to convert row to features: %v", err)
					continue
				}

				// Predict
				predictions, err := model.Predict(features)
				if err != nil {
					log.Errorf("Failed to predict: %v", err)
					continue
				}

				// Map class indices to labels using embedded labels
				for i := range predictions {
					classIdxStr := fmt.Sprintf("%d", predictions[i].ClassIndex)
					if csvLine, ok := model.Features.LabelMapping[classIdxStr]; ok {
						label := GetLabel(csvLine)
						if label != "" {
							predictions[i].Label = label
						} else {
							predictions[i].Label = fmt.Sprintf("UNKNOWN(csv_line=%d)", csvLine)
						}
					} else {
						predictions[i].Label = fmt.Sprintf("UNKNOWN(class_idx=%d)", predictions[i].ClassIndex)
					}
				}

				// Get class name and file path
				className := ""
				if class.Name != nil {
					className = class.Name.Qualified
					if className == "" {
						className = class.Name.Short
					}
				}

				filePath := pbFile.ShortPath
				if filePath == "" {
					filePath = pbFile.Path
				}

				allPredictions = append(allPredictions, ClassPrediction{
					File:        filePath,
					Class:       className,
					Predictions: predictions,
				})
			}
		}
	}

	return allPredictions, nil
}

func (p *Predictor) rowToFeatures(row []string, model *RandomForestModel) ([]float64, error) {
	// Row format from ExtractClassMetrics:
	// 0: stmt_name, 1: stmt_type, 2: file_path, 3: method_calls_raw, 4: uses_raw,
	// 5: namespace_raw, 6: path_raw, 7-28: numeric features, 29: programming_language, 30: cyclomatic_complexity

	if len(row) < 30 {
		return nil, fmt.Errorf("row has insufficient columns: %d", len(row))
	}

	// Extract class_name from namespace_raw (like Python does)
	className := extractClassName(row[5])
	namespaceRaw := row[5]
	progLang := row[28] // programming_language at index 28

	// Build feature vector in EXACT Python order:
	// 1. namespace_raw (categorical encoded)
	// 2. Numeric features (indices 7-28 in row)
	// 3. Programming language one-hot (weighted)
	// 4. class_name TF-IDF (weighted)
	// 5. Other NLP TF-IDF (path_raw, uses_raw, stmt_type)

	var allFeatures []float64

	// Step 1: Encode namespace_raw (categorical)
	if encoder, ok := model.Encoders["namespace_raw"]; ok {
		encoded := float64(-1) // Default for unknown
		for i, val := range encoder {
			if val == namespaceRaw {
				encoded = float64(i)
				break
			}
		}
		allFeatures = append(allFeatures, encoded)
	} else {
		// If no encoder, use 0
		allFeatures = append(allFeatures, 0)
	}

	// Step 2: Add numeric features (indices 7-28 in row)
	// These are: class_loc, logical_loc, comment_loc, nb_comments, nb_methods,
	// nb_extends, nb_implements, nb_traits, count_if, count_elseif, count_else,
	// count_case, count_switch, count_loop, nb_external_dependencies, depth_estimate,
	// nb_method_calls, nb_getters, nb_setters, nb_attributes, nb_unique_operators, cyclomatic_complexity
	for i := 7; i <= 28; i++ {
		if i < len(row) {
			val, err := strconv.ParseFloat(row[i], 64)
			if err != nil {
				val = 0
			}
			allFeatures = append(allFeatures, val)
		} else {
			allFeatures = append(allFeatures, 0)
		}
	}

	// Step 3: One-hot encode programming language with weight
	if model.ProgrammingLanguageEncoder != nil {
		oneHot := model.ProgrammingLanguageEncoder.Transform(progLang)
		weight := model.Features.ProgrammingLanguageWeight
		if weight == 0 {
			weight = 2.5 // Default from Python
		}
		for _, val := range oneHot {
			allFeatures = append(allFeatures, val*weight)
		}
	}

	// Step 4: Vectorize class_name with weight
	if vec, ok := model.Vectorizers["class_name"]; ok {
		tfidf := vec.Transform(className)
		weight := model.Features.ClassNameWeight
		if weight == 0 {
			weight = 3.0 // Default from Python
		}
		for _, val := range tfidf {
			allFeatures = append(allFeatures, val*weight)
		}
	}

	// Step 5: Vectorize other NLP columns in order: path_raw, uses_raw, stmt_type
	nlpOrder := []string{"path_raw", "uses_raw", "stmt_type"}
	nlpData := map[string]string{
		"path_raw":  row[6],
		"uses_raw":  row[4],
		"stmt_type": row[1],
	}

	for _, col := range nlpOrder {
		if vec, ok := model.Vectorizers[col]; ok {
			text := nlpData[col]
			if text == "" {
				text = ""
			}
			tfidf := vec.Transform(text)
			for _, val := range tfidf {
				allFeatures = append(allFeatures, val)
			}
		}
	}

	return allFeatures, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func extractClassName(namespace string) string {
	if namespace == "" {
		return ""
	}
	// Extract the last element (class name) after the last separator
	if strings.Contains(namespace, "\\") {
		parts := strings.Split(namespace, "\\")
		return parts[len(parts)-1]
	} else if strings.Contains(namespace, ".") {
		parts := strings.Split(namespace, ".")
		return parts[len(parts)-1]
	}
	return namespace
}
