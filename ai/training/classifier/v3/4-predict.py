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
model_data = joblib.load(MODEL)
model = model_data['model']
encoders = model_data.get('encoders', {})

features_data = json.load(open(FEATURES, "r"))
if isinstance(features_data, dict):
    features = features_data['features']
    categorical_cols = features_data.get('categorical_cols', [])
else:
    # Compatibilité avec l'ancien format
    features = features_data
    categorical_cols = []

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
X = df[features].copy()

# Encoder les colonnes catégorielles avec les encodeurs sauvegardés
for col in categorical_cols:
    if col in encoders:
        le = encoders[col]
        # Gérer les valeurs inconnues (non vues pendant l'entraînement)
        X[col] = X[col].astype(str)
        # Remplacer les valeurs inconnues par la première classe (ou une valeur par défaut)
        unknown_mask = ~X[col].isin(le.classes_)
        if unknown_mask.any():
            print(f"[WARNING] Found {unknown_mask.sum()} unknown values in '{col}', using default encoding")
            # Utiliser la première classe comme valeur par défaut
            X.loc[unknown_mask, col] = le.classes_[0]
        X[col] = le.transform(X[col])
    else:
        print(f"[WARNING] No encoder found for categorical column '{col}', skipping encoding")

print("[INFO] Predicting…")
preds = model.predict(X)

print("\n=== RESULTS ===\n")
for (_, row), label in zip(df.iterrows(), preds):
    cls = row["class"]
    file = row["file"]
    print(f"{cls} | {file} -> {label}")
