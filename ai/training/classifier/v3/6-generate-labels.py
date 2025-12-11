#!/usr/bin/env python3
"""
Generate a Go file containing the classification labels from the CSV file.
This allows the labels to be embedded in the compiled binary.
"""

import csv
import sys
import os

def generate_labels_go(csv_path, output_path):
    """Generate a Go file with embedded labels from CSV."""
    
    # Read labels and descriptions from CSV
    labels = []
    label_descriptions = {}
    with open(csv_path, 'r', encoding='utf-8') as f:
        reader = csv.reader(f)
        for i, row in enumerate(reader):
            if i == 0:  # Skip header
                continue
            if len(row) > 0:
                label = row[0].strip()
                if label and label != "label" and not label.startswith("#"):
                    labels.append(label)
                    # Get description from second column if available
                    description = row[1].strip() if len(row) > 1 and row[1].strip() else ""
                    label_descriptions[label] = description
    
    # Generate Go code
    go_code = '''package classifier

// ClassificationLabels contains all possible classification labels.
// This file is auto-generated from labels/c4.csv - DO NOT EDIT MANUALLY.
// To regenerate: python3 ai/training/classifier/v3/6-generate-labels.py

var ClassificationLabels = map[int]string{
'''
    
    for i, label in enumerate(labels, start=1):
        # Escape quotes in label
        escaped_label = label.replace('\\', '\\\\').replace('"', '\\"')
        go_code += f'\t{i}: "{escaped_label}",\n'
    
    go_code += '''}

// ClassificationDescriptions contains descriptions for each classification label.
var ClassificationDescriptions = map[string]string{
'''
    
    for label in labels:
        # Escape quotes in label and description
        escaped_label = label.replace('\\', '\\\\').replace('"', '\\"')
        description = label_descriptions.get(label, "")
        escaped_description = description.replace('\\', '\\\\').replace('"', '\\"')
        go_code += f'\t"{escaped_label}": "{escaped_description}",\n'
    
    go_code += '''}

// GetLabel returns the label for a given line number (1-indexed).
func GetLabel(lineNumber int) string {
	if label, ok := ClassificationLabels[lineNumber]; ok {
		return label
	}
	return ""
}

// GetDescription returns the description for a given label.
func GetDescription(label string) string {
	if description, ok := ClassificationDescriptions[label]; ok {
		return description
	}
	return ""
}
'''
    
    # Write to file
    with open(output_path, 'w', encoding='utf-8') as f:
        f.write(go_code)
    
    print(f"Generated {output_path} with {len(labels)} labels")

if __name__ == "__main__":
    if len(sys.argv) < 3:
        print("Usage: python3 6-generate-labels.py <labels_csv> <output_go_file>")
        print("Example: python3 6-generate-labels.py labels/c4.csv ../../internal/analyzer/classifier/labels.go")
        sys.exit(1)
    
    csv_path = sys.argv[1]
    output_path = sys.argv[2]
    
    if not os.path.exists(csv_path):
        print(f"Error: CSV file not found: {csv_path}")
        sys.exit(1)
    
    generate_labels_go(csv_path, output_path)
