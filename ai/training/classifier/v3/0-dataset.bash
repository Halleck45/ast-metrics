#!/bin/bash
echo "Generating dataset..."
go run cmd/dev/ai_dataset.go --output=ai/training/classifier/v3/dataset/samples.csv ./samples/

echo "Balancing dataset..."
python3 ai/training/classifier/v3/0-balance_dataset.py \
    ai/training/classifier/v3/dataset/samples.csv \
    --output=ai/training/classifier/v3/dataset/samples.csv \
    --strategy=language

echo "Dataset generation and balancing complete!"