# Code Classifier Training

This directory contains scripts and tools for training a machine learning model to classify code snippets according to their architectural role.

## Overview

The training pipeline consists of several steps:
1. **Dataset Generation**: Extract code features from source files
2. **Dataset Balancing**: Balance the dataset by programming language or labels
3. **Labeling**: Classify code snippets using LLM (local or remote)
4. **Merging**: Merge labels with the original dataset
5. **Training**: Train a RandomForest classifier
6. **Export**: Export the model to JSON format for Go runtime

## Installation

### Python Dependencies

Install required Python packages:

```bash
pip install -r requirements.txt
```

Or manually:

```bash
pip install numpy pandas scikit-learn sentence-transformers llama-cpp-python openai
```

### Local LLM Model (Optional)

If you want to use a local LLM for labeling instead of OpenAI API, download a GGUF model:

```bash
mkdir -p models
curl -L "https://huggingface.co/TheBloke/Mistral-7B-Instruct-v0.2-GGUF/resolve/main/mistral-7b-instruct-v0.2.Q4_K_M.gguf" -o ./models/mistral-7b-instruct-v0.2.Q4_K_M.gguf
```

> **Note**: Labeling with a local model can be slow. The Mistral 7B instruct model works well for this task, but you can also use a remote OpenAI API (see `1-labelize_remote.py`).

## Quick Start

### Automated Pipeline

The easiest way to train a model is to use the automated pipeline:

```bash
bash train_pipeline.bash --language=php --source=../../../samples/php
```

Supported languages: `php`, `go`, `rust`, `python`

### Manual Steps

If you prefer to run each step manually:

#### 1. Generate Dataset

Prepare the dataset from source code:

```bash
go run cmd/dev/ai_dataset.go --output=ai/training/classifier/dataset/samples.csv ./samples/
```

Or use your own source directory:

```bash
go run cmd/dev/ai_dataset.go --output=<path-to-output-csv> <path-to-examples-folder>
```

#### 2. Balance Dataset (Optional)

Balance the dataset by programming language or labels:

```bash
python 0-balance_dataset.py dataset/samples.csv --output=dataset/samples_balanced.csv --strategy=language
```

Strategies:
- `language`: Balance by programming language
- `label`: Balance by classification labels
- `both`: Balance by both language and labels

#### 3. Label the Dataset

Label the dataset using a local LLM:

```bash
python 1-labelize.py --count=10000 dataset/samples.csv
```

Or use OpenAI API (requires `OPENAI_API_KEY` environment variable):

```bash
python 1-labelize_remote.py --count=10000 dataset/samples.csv
```

> **Note**: Labeling can take a while depending on the number of samples and your hardware/API rate limits.

#### 4. Merge Labels

Merge the labels with the original dataset:

```bash
python 2-merge_dataset.py dataset/samples.csv classified_output/classified_c4.csv dataset/final_dataset.csv
```

#### 5. Train the Model

Train the RandomForest classifier:

```bash
python 3-train.py dataset/final_dataset.csv build/php/model.pkl build/php/features.json --encoders-out=build/php/encoders.joblib
```

#### 6. Export Model

Export the model to JSON format for Go runtime:

```bash
python 5-export.py build/php/model.pkl build/php/encoders.joblib build/php/features.json build/php/model.json
```

#### 7. Generate Labels Go File

After updating labels, regenerate the Go code:

```bash
python 6-generate-labels.py labels/c4.csv ../../internal/analyzer/classifier/labels.go
```

## Prediction

Make predictions on new code samples:

```bash
python 4-predict.py dataset/test_samples.csv build/php/model.pkl build/php/features.json
```

For JSON output:

```bash
python 4-predict.py dataset/test_samples.csv build/php/model.pkl build/php/features.json --json-output
```

## Scripts Overview

| Script | Purpose |
|--------|---------|
| `0-balance_dataset.py` | Balance dataset by language/labels |
| `0-dataset.bash` | Generate and balance dataset (legacy) |
| `1-labelize.py` | Label dataset using local LLM |
| `1-labelize_remote.py` | Label dataset using OpenAI API |
| `2-merge_dataset.py` | Merge labels with original dataset |
| `3-train.py` | Train RandomForest classifier |
| `4-predict.py` | Make predictions on new samples |
| `5-export.py` | Export model to JSON format |
| `6-generate-labels.py` | Generate Go code from label definitions |
| `train_pipeline.bash` | Automated training pipeline |

## Dataset Format

The input CSV should contain the following columns:
- `stmt_name`: Class/function name
- `file_path`: Path to the source file
- `namespace_raw`: Full namespace/package path
- `programming_language`: Programming language (PHP, Go, Python, Rust, etc.)
- Additional feature columns (automatically extracted by `ai_dataset.go`)

The labeled CSV should contain:
- `class`: Class/function name (matches `stmt_name`)
- `file`: File path (matches `file_path`)
- `label`: Classification label

## Labels

Labels are defined in `labels/c4.csv`. Each label can have an optional description in the second column.

To add or modify labels:
1. Edit `labels/c4.csv`
2. Retrain the model
3. Regenerate the Go labels file: `python 6-generate-labels.py labels/c4.csv ../../internal/analyzer/classifier/labels.go`

## Model Architecture

The classifier uses:
- **RandomForest** with 500 estimators
- **TF-IDF vectorization** for text features (class names, paths, etc.)
- **One-hot encoding** for categorical features
- **Weighted features** for class names and programming languages

The model is optimized for:
- Accuracy (target: >80%)
- Model size (compressed: <8MB)
- Inference speed

## Output Files

After training, the following files are generated in `build/<language>/`:
- `model.pkl`: Trained RandomForest model (compressed)
- `model.json.gz`: Model exported for Go runtime (compressed JSON)
- `features.json`: Feature metadata and label mappings
- `encoders.joblib`: Encoders and vectorizers (compressed)

## Troubleshooting

### Missing columns error
Ensure your dataset CSV contains all required columns. Run `ai_dataset.go` to generate a properly formatted dataset.

### Label mismatch
Check that the labels in your classified CSV match the labels defined in `labels/c4.csv`.

### Model size too large
Reduce `MAX_NLP_FEATURES` in `3-train.py` or use more aggressive compression.

### Low accuracy
- Increase training data
- Balance the dataset better
- Adjust hyperparameters in `3-train.py`

## Contributing

When contributing to this training pipeline:
1. Follow Python PEP 8 style guidelines
2. Add docstrings to new functions
3. Update this README if you add new features
4. Test your changes with a small dataset first
