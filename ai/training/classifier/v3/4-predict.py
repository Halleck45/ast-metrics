# predict_batch.py
import sys
import pandas as pd
import joblib
import json
from argparse import ArgumentParser
parser = ArgumentParser()
parser.add_argument("samples", help="Input samples CSV")
parser.add_argument("model", help="Input model file")
parser.add_argument("features", help="Input features JSON")
args = parser.parse_args()
SAMPLES = args.samples
MODEL = args.model
FEATURES = args.features

print("[INFO] Loading model and features…")
model = joblib.load(MODEL)
features = json.load(open(FEATURES, "r"))

print("[INFO] Loading samples:", SAMPLES)
df = pd.read_csv(SAMPLES)

# Harmoniser avec merge + train
df.rename(columns={
    "stmt_name": "class",
    "file_path": "file",
}, inplace=True)

print("[INFO] Preparing dataset…")
missing = [c for c in features if c not in df.columns]
if missing:
    print("[ERROR] Missing required columns:", missing)
    sys.exit(1)

# On ne garde que les colonnes nécessaires
X = df[features]

print("[INFO] Predicting…")
preds = model.predict(X)

print("\n=== RESULTS ===\n")
for (_, row), label in zip(df.iterrows(), preds):
    cls = row["class"]
    file = row["file"]
    print(f"{cls} | {file} -> {label}")
