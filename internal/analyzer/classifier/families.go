package classifier

// Family describes a high-level architectural grouping.
type Family struct {
	Key         string   // unique key: interface, application, domain, infrastructure, core, utility, development
	Description string   // english description
	Color       string   // hex color for UI grouping
	Labels      []string // classification labels belonging to this family
}

var ClassificationFamilies = []Family{
	{
		Key:         "interface",
		Description: "Entry points and presentation components. They expose the system through controllers, views, presenters, and transform data for display.",
		Color:       "#2196F3", // Blue
		Labels: []string{
			"component:interface:controller",
			"component:interface:presenter",
			"component:interface:view",
			"component:interface:view_model",
			"construction:transformer:view",
		},
	},
	{
		Key:         "application",
		Description: "Application orchestration layer. Coordinates use cases, services, workflows, transactions, and message handlers. Contains logic that organizes domain operations.",
		Color:       "#4CAF50", // Green
		Labels: []string{
			"component:app:use_case",
			"component:app:service",
			"component:app:flow_orchestrator",
			"component:app:transaction_manager",
			"component:app:mapper",
			"component:app:converter",
			"component:messaging:handler",
		},
	},
	{
		Key:         "domain",
		Description: "Pure domain logic and business rules: entities, value objects, aggregates, policies, rules, specifications, and domain services.",
		Color:       "#FFC107", // Amber
		Labels: []string{
			"component:domain:entity",
			"component:domain:value_object",
			"component:domain:aggregate_root",
			"component:domain:service",
			"component:domain:rule",
			"component:domain:policy",
			"component:domain:specification",
		},
	},
	{
		Key:         "infrastructure",
		Description: "Technical implementation details: persistence, gateways, external systems, clients, caching, security, configuration, monitoring, logging, and messaging infrastructure.",
		Color:       "#9C27B0", // Purple
		Labels: []string{
			"component:data_access:repository",
			"component:data_access:gateway",
			"component:data_access:mapper",
			"component:data_access:dto",
			"component:messaging:publisher",
			"component:messaging:subscriber",
			"component:messaging:bus",
			"infrastructure:client:http",
			"infrastructure:client:database",
			"infrastructure:client:queue",
			"infrastructure:system:file_io",
			"infrastructure:system:environment",
			"infrastructure:logging:logger",
			"infrastructure:monitoring:metric",
			"infrastructure:security:token_handler",
			"infrastructure:security:authenticator",
			"infrastructure:security:authorizer",
			"infrastructure:cache:manager",
			"infrastructure:configuration:loader",
			"infrastructure:configuration:model",
			"infrastructure:error:handler",
			"framework:internal:infrastructure",
		},
	},
	{
		Key:         "core",
		Description: "Low-level reusable components, algorithms, utilities, and runtime support structures used internally by the system.",
		Color:       "#F44336", // Red
		Labels: []string{
			"component:core:library",
			"component:core:algorithm",
			"component:core:utility",
			"component:core:runtime_support",
		},
	},
	{
		Key:         "utility",
		Description: "General-purpose helpers, validators, serializers, converters and any cross-cutting utility unrelated to domain or application logic.",
		Color:       "#607D8B", // Blue Grey
		Labels: []string{
			"utility:helper:string",
			"utility:helper:date_time",
			"utility:helper:math",
			"utility:helper:component",
			"utility:converter:format",
			"utility:validator:input",
			"utility:validator:model",
			"utility:serialization:serializer",
			"utility:serialization:deserializer",
		},
	},
	{
		Key:         "development",
		Description: "Testing-related artefacts: test cases, fixtures, mocks, and development-only classes.",
		Color:       "#795548", // Brown
		Labels: []string{
			"development:test:case",
			"development:test:fixture",
			"development:test:mock",
		},
	},
}

// GetFamilyForLabel returns the family key for a given label, or "unknown" if not found.
func GetFamilyForLabel(label string) string {
	for _, family := range ClassificationFamilies {
		for _, familyLabel := range family.Labels {
			if familyLabel == label {
				return family.Key
			}
		}
	}
	return "unknown"
}

// GroupByFamilyAndLabel groups predictions by family first, then by label.
// Returns a map: familyKey -> label -> []ClassPrediction
type FamilyGroupedPredictions map[string]map[string][]ClassPrediction

func GroupByFamilyAndLabel(predictions []ClassPrediction) FamilyGroupedPredictions {
	result := make(FamilyGroupedPredictions)

	// Initialize all families in order
	for _, family := range ClassificationFamilies {
		result[family.Key] = make(map[string][]ClassPrediction)
	}
	result["unknown"] = make(map[string][]ClassPrediction)

	// Group predictions
	for _, p := range predictions {
		var label string
		if len(p.Predictions) > 0 {
			label = p.Predictions[0].Label
		} else {
			label = "Unknown"
		}

		familyKey := GetFamilyForLabel(label)
		if result[familyKey] == nil {
			result[familyKey] = make(map[string][]ClassPrediction)
		}
		result[familyKey][label] = append(result[familyKey][label], p)
	}

	return result
}
