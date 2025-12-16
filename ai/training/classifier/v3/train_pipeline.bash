#!/bin/bash
set -e

# Default values
LANGUAGE=""
SOURCE=""
COUNT="20000"

# Flags to skip certain steps
SKIP_DATASET_GEN=true
SKIP_LABELIZE=true

# Parse arguments
for i in "$@"; do
  case $i in
    --language=*)
      LANGUAGE="${i#*=}"
      shift # past argument=value
      ;;
    --source=*)
      SOURCE="${i#*=}"
      shift # past argument=value
      ;;
    *)
      # unknown option
      echo "Unknown option: $i"
      exit 1
      ;;
  esac
done

# Validation
if [ -z "$LANGUAGE" ] || [ -z "$SOURCE" ]; then
    echo "Usage: ./train_pipeline.bash --language=<lang> --source=<path-to-source>"
    exit 1
fi

LANGUAGE=$(echo "$LANGUAGE" | tr '[:upper:]' '[:lower:]')
ALLOWED_LANGUAGES=("php" "go" "rust" "python")

if [[ ! " ${ALLOWED_LANGUAGES[@]} " =~ " ${LANGUAGE} " ]]; then
    echo "Error: Language must be one of: ${ALLOWED_LANGUAGES[*]}"
    exit 1
fi

if [ ! -d "$SOURCE" ]; then
    echo "Error: Source directory not found: $SOURCE"
    exit 1
fi

# Paths
BASE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DATASET_DIR="$BASE_DIR/dataset"
CLASSIFIED_DIR="$BASE_DIR/classified_output/$LANGUAGE"
BUILD_DIR="$BASE_DIR/build/$LANGUAGE"

# Ensure directories exist
mkdir -p "$DATASET_DIR"
mkdir -p "$CLASSIFIED_DIR"
mkdir -p "$BUILD_DIR"

# Files
DATASET_CSV="$DATASET_DIR/dataset_${LANGUAGE}.csv"
BALANCED_CSV="$DATASET_DIR/dataset_${LANGUAGE}_balanced.csv"
MERGED_CSV="$DATASET_DIR/final_dataset_${LANGUAGE}.csv"
MODEL_OUT="$BUILD_DIR/model.pkl"
FEATURES_OUT="$BUILD_DIR/features.json"
ENCODERS_OUT="$BUILD_DIR/encoders.joblib"

echo "=============================================="
echo "Starting Training Pipeline for $LANGUAGE"
echo "Source: $SOURCE"
echo "=============================================="
echo ""

# 1. Generate Dataset
echo "[1/5] Generating dataset from source..."
# make it absolute
DATASET_REAL_PATH=$(cd "$SOURCE" && pwd)
cd $BASE_DIR/../../../../
pwd
CMD="go run cmd/dev/ai_dataset.go --output=$DATASET_CSV $DATASET_REAL_PATH"
echo "Running: $CMD"
if [ "$SKIP_DATASET_GEN" = true ]; then
    echo "Skipping dataset generation as per configuration."
else
    $CMD
fi
cd $BASE_DIR

# 2. Balance Dataset
# echo ""
# echo "[2/5] Balancing dataset..."
# CMD="python 0-balance_dataset.py $DATASET_CSV --output=$BALANCED_CSV"
# echo "Running: $CMD"
# $CMD

# 3. Labelize
echo ""
echo "[3/5] Labelizing dataset..."
# Default count 
# CMD="python 1-labelize_remote.py --count=$COUNT --output-dir=$CLASSIFIED_DIR $BALANCED_CSV"
CMD="python 1-labelize_remote.py --count=$COUNT --output-dir=$CLASSIFIED_DIR $DATASET_CSV"
if [ "$SKIP_LABELIZE" = true ]; then
    echo "Skipping labelization as per configuration."
else
    echo "Running: $CMD"
    $CMD
fi

# 4. Merge
echo ""
echo "[4/5] Merging labels..."
LABELED_CSV="$CLASSIFIED_DIR/classified_c4.csv"

if [ ! -f "$LABELED_CSV" ]; then
    echo "Error: Labeled CSV not found: $LABELED_CSV"
    exit 1
fi

#CMD="python 2-merge_dataset.py $BALANCED_CSV $LABELED_CSV $MERGED_CSV"
CMD="python 2-merge_dataset.py $DATASET_CSV $LABELED_CSV $MERGED_CSV"
echo "Running: $CMD"
$CMD

# 5. Train
echo ""
echo "[5/5] Training model..."
CMD="python 3-train.py $MERGED_CSV $MODEL_OUT $FEATURES_OUT --encoders-out=$ENCODERS_OUT"
echo "Running: $CMD"
$CMD

# 6. Export Model to JSON
echo ""
echo "[6/6] Exporting model to JSON for Go runtime..."
MODEL_JSON="$BUILD_DIR/model.json"
CMD="python 5-export.py $MODEL_OUT $ENCODERS_OUT $FEATURES_OUT $MODEL_JSON"
echo "Running: $CMD"
$CMD

echo ""
echo "=============================================="
echo "Pipeline Finished Successfully!"
echo "Model: $MODEL_OUT"
echo "Model JSON: $MODEL_JSON"
echo "Features: $FEATURES_OUT"
echo "Encoders: $ENCODERS_OUT"
echo "=============================================="
