# Code Classifier Training

This directory contains the ML pipeline that trains a RandomForest classifier to assign architectural roles to code classes (e.g., `component:domain:entity`, `infrastructure:client:http`). The trained model is exported as compressed JSON for Go runtime inference.

## Dataset

Training labels are hosted on HuggingFace: [Halleck45/code-classification](https://huggingface.co/datasets/Halleck45/code-classification)

Download labels:
```bash
bash download_dataset.bash
```

## Quick Start

### Automated Pipeline

```bash
pip install -r requirements.txt
bash train_pipeline.bash --language=php --source=../../../samples/php
```

Labels are automatically downloaded from HuggingFace. Use `--local-labels=<path>` to use a local labels CSV instead.

Supported languages: `php` (more coming)

### Manual Steps

```bash
# 1. Generate dataset from source code (run from project root)
go run cmd/dev/ai_dataset.go --output=ai/training/classifier/dataset/samples.csv ./samples/

# 2. Download labels from HuggingFace (or use local LLM labeling as alternative)
bash download_dataset.bash
# Alternative: python 1-labelize_remote.py --count=10000 dataset/samples.csv  # OpenAI API

# 3. Merge labels with features
python 2-merge_dataset.py dataset/samples.csv classified_output/php/classified_roles.csv dataset/final_dataset.csv

# 5. Train the model
python 3-train.py dataset/final_dataset.csv build/php/model.pkl build/php/features.json --encoders-out=build/php/encoders.joblib

# 6. Export model to JSON for Go runtime
python 5-export.py build/php/model.pkl build/php/encoders.joblib build/php/features.json build/php/model.json

# 7. Copy to embedded models directory
cp build/php/model.json.gz ../../internal/analyzer/classifier/models/php/model.json.gz

# 8. (Optional) Regenerate Go labels file
python 6-generate-labels.py labels/roles.csv ../../internal/analyzer/classifier/labels.go
```

## Pipeline Architecture

| Script | Purpose |
|--------|---------|
| `0-balance_dataset.py` | Balance dataset by language/labels |
| `1-labelize.py` | Label dataset using local GGUF model |
| `1-labelize_remote.py` | Label dataset using OpenAI API |
| `2-merge_dataset.py` | Merge LLM labels with feature dataset, convert to numeric IDs |
| `3-train.py` | Train RandomForest + cross-validation + F1 metrics |
| `4-predict.py` | Run predictions on new samples |
| `5-export.py` | Export model to gzipped JSON (optimized: leaf-only values) |
| `6-generate-labels.py` | Code-generate `labels.go` from `labels/roles.csv` |
| `train_pipeline.bash` | Automated end-to-end pipeline |

## Model Details

- **Algorithm**: RandomForest (200 trees, unlimited depth, `class_weight="balanced"`)
- **Text features**: Hashing TF-IDF (32,768 bins) for `path_raw`, `uses_raw`, `stmt_type`, `class_name`
- **Feature weights**: `class_name` 3x, `uses_raw` 2x, `path_raw` 1.5x, `stmt_type` 0.5x, `programming_language` 1x
- **Derived features**: 8 ratios (getter_ratio, loc_per_method, complexity_per_method, etc.)
- **Path normalization**: Common path prefix is stripped during training to prevent learning absolute paths
- **Labels**: 64 hierarchical roles defined in `labels/roles.csv` (line number = numeric ID)
- **Current metrics**: F1-macro ~0.57, accuracy ~0.58 on PHP holdout set
- **Model size**: ~8.5 MB gzipped (constraint: <30 MB for distribution)

### Go Inference Compatibility

The exported JSON must match Go inference in `internal/analyzer/classifier/`:
- **Hash function**: Murmur3-32 with `abs(int32(hash)) % n_features` (matches sklearn)
- **Stop words**: Full sklearn English list (318 words) in `tokenize.go`
- **N-grams**: Generated from `ngram_range` field in vectorizer config
- **Feature vector order**: numerical → programming_language one-hot → class_name TF-IDF → NLP cols TF-IDF
- **Derived features**: computed in `predictor.go:computeDerivedFeature()` from raw extraction columns

## Labels

Defined in `labels/roles.csv`. Hierarchical format: `category:subcategory:role`.

Families (used in report grouping):
- **interface**: controllers, views, presenters
- **application**: services, use cases, handlers, mappers
- **domain**: entities, value objects, aggregates, rules, specifications
- **infrastructure**: repositories, clients, security, cache, logging
- **core**: libraries, algorithms, utilities
- **utility**: helpers, validators, serializers
- **development**: test cases, fixtures, mocks

## Output Files

Generated in `build/<language>/`:
- `model.pkl` — sklearn RandomForest (joblib compressed)
- `encoders.joblib` — feature encoders + TF-IDF pipelines
- `features.json` — metadata (column order, label mapping, weights, config)
- `model.json.gz` — final export for Go (trees with leaf-only values, ~8.5 MB)
