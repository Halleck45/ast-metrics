#!/bin/bash
# Legacy script for dataset generation and balancing
# Note: This script is kept for backward compatibility
# Consider using train_pipeline.bash or running steps manually

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"

echo "Generating dataset..."
cd "$PROJECT_ROOT" || exit 1
go run cmd/dev/ai_dataset.go --output=ai/training/classifier/dataset/samples.csv ./samples/

echo "Balancing dataset..."
python3 ai/training/classifier/0-balance_dataset.py \
    ai/training/classifier/dataset/samples.csv \
    --output=ai/training/classifier/dataset/samples.csv \
    --strategy=language

echo "Dataset generation and balancing complete!"