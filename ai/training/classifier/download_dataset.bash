#!/bin/bash
set -e

# Download classification labels from HuggingFace
# Repository: https://huggingface.co/datasets/Halleck45/code-classification

BASE_URL="https://huggingface.co/datasets/Halleck45/code-classification/resolve/main"
BASE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LANGUAGES=("php")  # extensible: add "go", "python", "rust" when available

cd "$BASE_DIR"

for lang in "${LANGUAGES[@]}"; do
    OUTPUT_DIR="classified_output/$lang"
    OUTPUT_FILE="$OUTPUT_DIR/classified_roles.csv"
    mkdir -p "$OUTPUT_DIR"

    echo "Downloading $lang labels from HuggingFace..."
    curl -fSL "$BASE_URL/$lang/classified_roles.csv" -o "$OUTPUT_FILE"

    ROWS=$(wc -l < "$OUTPUT_FILE")
    echo "  → $OUTPUT_FILE ($((ROWS - 1)) labels)"
done

echo "Done."
