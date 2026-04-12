package analyzer

import (
	"fmt"
	"sort"
	"strings"

	"github.com/halleck45/ast-metrics/internal/analyzer/classifier"
	pb "github.com/halleck45/ast-metrics/pb"
)

// ArchitectureViolation représente une violation de règle d'architecture
type ArchitectureViolation struct {
	FromClass string
	ToClass   string
	FromRole  string
	ToRole    string
	Rule      string
}

// Ambiguity représente une classe ambiguë (plusieurs labels avec scores proches)
type Ambiguity struct {
	ClassName      string
	File           string
	TopLabels      []LabelScore
	AmbiguityScore float64 // Écart entre label 1 et label 2
	Pattern        string  // Pattern détecté (ex: "Entité + Repository")
}

// LabelScore représente un label avec son score
type LabelScore struct {
	Label string
	Score float64
}

// RoleFlow représente un flux entre deux rôles
type RoleFlow struct {
	FromRole string
	ToRole   string
	Count    int
	IsValid  bool // Si le flux respecte les règles d'architecture
}

// ArchitectureMetrics contient toutes les métriques d'architecture
type ArchitectureMetrics struct {
	Violations  []ArchitectureViolation
	Ambiguities []Ambiguity
	RoleFlows   []RoleFlow
}

// AnalyzeArchitecture analyse l'architecture et détecte les violations, ambiguïtés et flux
func AnalyzeArchitecture(aggregate *Aggregated, predictions []classifier.ClassPrediction) *ArchitectureMetrics {
	metrics := &ArchitectureMetrics{
		Violations:  []ArchitectureViolation{},
		Ambiguities: []Ambiguity{},
		RoleFlows:   []RoleFlow{},
	}

	if aggregate == nil || aggregate.Graph == nil {
		return metrics
	}

	// Créer un map classe -> prédictions
	classPredictions := make(map[string]classifier.ClassPrediction)
	for _, pred := range predictions {
		classPredictions[pred.Class] = pred
	}

	// Détecter les violations
	metrics.Violations = detectViolations(aggregate.Graph, classPredictions)

	// Détecter les ambiguïtés
	metrics.Ambiguities = detectAmbiguities(predictions, 0.10) // 10% de seuil

	// Analyser les flux de rôles
	metrics.RoleFlows = analyzeRoleFlows(aggregate.Graph, classPredictions)

	return metrics
}

// detectViolations détecte les violations de règles d'architecture
func detectViolations(graph *pb.Graph, classPredictions map[string]classifier.ClassPrediction) []ArchitectureViolation {
	var violations []ArchitectureViolation

	if graph == nil || graph.Nodes == nil {
		return violations
	}

	for fromClass, node := range graph.Nodes {
		fromPred, hasFromPred := classPredictions[fromClass]
		if !hasFromPred || len(fromPred.Predictions) == 0 {
			continue
		}

		fromRole := fromPred.Predictions[0].Label

		for _, toClass := range node.Edges {
			toPred, hasToPred := classPredictions[toClass]
			if !hasToPred || len(toPred.Predictions) == 0 {
				continue
			}

			toRole := toPred.Predictions[0].Label

			// Vérifier si la dépendance est valide
			if !isValidDependency(fromRole, toRole) {
				violations = append(violations, ArchitectureViolation{
					FromClass: fromClass,
					ToClass:   toClass,
					FromRole:  fromRole,
					ToRole:    toRole,
					Rule:      fmt.Sprintf("%s ne peut pas dépendre de %s", fromRole, toRole),
				})
			}
		}
	}

	return violations
}

// isValidDependency vérifie si une dépendance entre deux rôles est valide
func isValidDependency(fromRole, toRole string) bool {
	fromParts := strings.Split(fromRole, ":")
	toParts := strings.Split(toRole, ":")

	if len(fromParts) < 2 || len(toParts) < 2 {
		return true // Par défaut, on accepte si on ne peut pas parser
	}

	fromCategory := fromParts[1] // Ex: "interface", "domain", "app", "data_access"
	fromLayer := ""
	if len(fromParts) >= 3 {
		fromLayer = fromParts[2]
	}

	toCategory := toParts[1]
	toLayer := ""
	if len(toParts) >= 3 {
		toLayer = toParts[2]
	}

	// Règle 1: component:interface:controller ne peut pas dépendre d'infrastructure:*
	if fromLayer == "controller" && fromCategory == "interface" && strings.HasPrefix(toRole, "infrastructure:") {
		return false
	}

	// Règle 2: component:domain:* ne doit jamais dépendre de component:app:*
	if fromCategory == "domain" && toCategory == "app" {
		return false
	}

	// Règle 3: component:data_access:* ne doit pas appeler component:interface:*
	if fromCategory == "data_access" && toLayer == "interface" {
		return false
	}

	// Règle 4: component:domain:rule ne doit dépendre que de domain:entity ou domain:value_object
	if fromLayer == "rule" && fromCategory == "domain" {
		if toCategory != "domain" || (toLayer != "entity" && toLayer != "value_object") {
			return false
		}
	}

	return true
}

// detectAmbiguities détecte les classes ambiguës (plusieurs labels avec scores proches)
func detectAmbiguities(predictions []classifier.ClassPrediction, threshold float64) []Ambiguity {
	var ambiguities []Ambiguity

	for _, pred := range predictions {
		if len(pred.Predictions) < 2 {
			continue
		}

		// Calculer l'écart entre le premier et le deuxième label
		ambiguityScore := pred.Predictions[0].Probability - pred.Predictions[1].Probability

		if ambiguityScore < threshold {
			// Classe ambiguë détectée
			topLabels := make([]LabelScore, 0, len(pred.Predictions))
			for _, p := range pred.Predictions {
				if len(topLabels) >= 3 {
					break
				}
				topLabels = append(topLabels, LabelScore{
					Label: p.Label,
					Score: p.Probability,
				})
			}

			pattern := detectPattern(topLabels)

			ambiguities = append(ambiguities, Ambiguity{
				ClassName:      pred.Class,
				File:           pred.File,
				TopLabels:      topLabels,
				AmbiguityScore: ambiguityScore,
				Pattern:        pattern,
			})
		}
	}

	// Trier par score d'ambiguïté (plus ambigu = plus petit écart)
	sort.Slice(ambiguities, func(i, j int) bool {
		return ambiguities[i].AmbiguityScore < ambiguities[j].AmbiguityScore
	})

	return ambiguities
}

// detectPattern détecte le pattern architectural d'une classe ambiguë
func detectPattern(labels []LabelScore) string {
	if len(labels) < 2 {
		return "Unknown"
	}

	label1 := labels[0].Label
	label2 := labels[1].Label

	parts1 := strings.Split(label1, ":")
	parts2 := strings.Split(label2, ":")

	if len(parts1) < 3 || len(parts2) < 3 {
		return fmt.Sprintf("%s + %s", label1, label2)
	}

	layer1 := parts1[2]
	category1 := parts1[1]
	layer2 := parts2[2]
	category2 := parts2[1]

	// Pattern: Entité + Repository
	if (layer1 == "entity" && layer2 == "repository") || (layer1 == "repository" && layer2 == "entity") {
		return "Entité qui contient du code de persistance"
	}

	// Pattern: Service + Infrastructure
	if (category1 == "app" && category2 == "infrastructure") || (category1 == "infrastructure" && category2 == "app") {
		return "Service qui fait de l'infrastructure"
	}

	// Pattern: Handler + Orchestration
	if (layer1 == "handler" && category2 == "app") || (layer1 == "controller" && category2 == "app") {
		return "Handler qui orchestre trop de logique"
	}

	// Pattern générique
	return fmt.Sprintf("%s + %s", layer1, layer2)
}

// analyzeRoleFlows analyse les flux entre rôles
func analyzeRoleFlows(graph *pb.Graph, classPredictions map[string]classifier.ClassPrediction) []RoleFlow {
	flows := make(map[string]map[string]int) // fromRole -> toRole -> count

	if graph == nil || graph.Nodes == nil {
		return []RoleFlow{}
	}

	for fromClass, node := range graph.Nodes {
		fromPred, hasFromPred := classPredictions[fromClass]
		if !hasFromPred || len(fromPred.Predictions) == 0 {
			continue
		}

		fromRole := fromPred.Predictions[0].Label

		for _, toClass := range node.Edges {
			toPred, hasToPred := classPredictions[toClass]
			if !hasToPred || len(toPred.Predictions) == 0 {
				continue
			}

			toRole := toPred.Predictions[0].Label

			if flows[fromRole] == nil {
				flows[fromRole] = make(map[string]int)
			}
			flows[fromRole][toRole]++
		}
	}

	// Convertir en liste
	var roleFlows []RoleFlow
	for fromRole, toRoles := range flows {
		for toRole, count := range toRoles {
			isValid := isValidDependency(fromRole, toRole)
			roleFlows = append(roleFlows, RoleFlow{
				FromRole: fromRole,
				ToRole:   toRole,
				Count:    count,
				IsValid:  isValid,
			})
		}
	}

	return roleFlows
}

// GetRoleCategory extrait la catégorie d'un label
func GetRoleCategory(label string) string {
	parts := strings.Split(label, ":")
	if len(parts) >= 2 {
		return parts[1]
	}
	return "unknown"
}

// GetRoleLayer extrait la couche d'un label
func GetRoleLayer(label string) string {
	parts := strings.Split(label, ":")
	if len(parts) >= 3 {
		return parts[2]
	}
	if len(parts) >= 2 {
		return parts[1]
	}
	return "unknown"
}

