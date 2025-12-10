package classifier

// ClassificationLabels contains all possible classification labels.
// This file is auto-generated from labels/c4.csv - DO NOT EDIT MANUALLY.
// To regenerate: python3 ai/training/classifier/v3/6-generate-labels.py

var ClassificationLabels = map[int]string{
	1: "component:interface:controller",
	2: "component:interface:presenter",
	3: "component:interface:view",
	4: "component:interface:view_model",
	5: "component:app:use_case",
	6: "component:app:service",
	7: "component:app:flow_orchestrator",
	8: "component:app:transaction_manager",
	9: "component:app:mapper",
	10: "component:app:converter",
	11: "component:domain:entity",
	12: "component:domain:value_object",
	13: "component:domain:aggregate_root",
	14: "component:domain:service",
	15: "component:domain:rule",
	16: "component:domain:policy",
	17: "component:domain:specification",
	18: "component:data_access:repository",
	19: "component:data_access:gateway",
	20: "component:data_access:mapper",
	21: "component:data_access:dto",
	22: "component:messaging:handler",
	23: "component:messaging:publisher",
	24: "component:messaging:subscriber",
	25: "component:messaging:bus",
	26: "construction:factory:abstract",
	27: "construction:factory:method",
	28: "construction:builder:fluent",
	29: "construction:builder:configurator",
	30: "construction:adapter:external",
	31: "construction:adapter:internal",
	32: "construction:transformer:data",
	33: "construction:transformer:view",
	34: "infrastructure:client:http",
	35: "infrastructure:client:database",
	36: "infrastructure:client:queue",
	37: "infrastructure:system:file_io",
	38: "infrastructure:system:environment",
	39: "infrastructure:logging:logger",
	40: "infrastructure:monitoring:metric",
	41: "infrastructure:security:token_handler",
	42: "infrastructure:security:authenticator",
	43: "infrastructure:security:authorizer",
	44: "infrastructure:cache:manager",
	45: "infrastructure:configuration:loader",
	46: "infrastructure:configuration:model",
	47: "infrastructure:error:handler",
	48: "component:core:library",
	49: "component:core:algorithm",
	50: "component:core:utility",
	51: "component:core:runtime_support",
	52: "utility:helper:string",
	53: "utility:helper:date_time",
	54: "utility:helper:math",
	55: "utility:helper:component",
	56: "utility:converter:format",
	57: "utility:validator:input",
	58: "utility:validator:model",
	59: "utility:serialization:serializer",
	60: "utility:serialization:deserializer",
	61: "framework:internal:infrastructure",
	62: "development:test:case",
	63: "development:test:fixture",
	64: "development:test:mock",
}

// GetLabel returns the label for a given line number (1-indexed).
func GetLabel(lineNumber int) string {
	if label, ok := ClassificationLabels[lineNumber]; ok {
		return label
	}
	return ""
}
