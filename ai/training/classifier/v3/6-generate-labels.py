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
    
    # Read labels from CSV
    labels = []
    with open(csv_path, 'r', encoding='utf-8') as f:
        reader = csv.reader(f)
        for i, row in enumerate(reader):
            if i == 0:  # Skip header
                continue
            if len(row) > 0:
                label = row[0].strip()
                if label and label != "label" and not label.startswith("#"):
                    labels.append(label)
    
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

// GetLabel returns the label for a given line number (1-indexed).
func GetLabel(lineNumber int) string {
	if label, ok := ClassificationLabels[lineNumber]; ok {
		return label
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
