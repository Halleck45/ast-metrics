# predict_batch.py
import sys
import os
import csv
import pandas as pd
import joblib
import json
from argparse import ArgumentParser

parser = ArgumentParser()
parser.add_argument("samples", help="Input samples CSV")
parser.add_argument("model", help="Input model file (RandomForest only)")
parser.add_argument("features", help="Input features JSON")
parser.add_argument("--encoders", help="Input encoders file (defaults to encoders.pkl)", 
                    default="encoders.pkl")
parser.add_argument("--labels-def", help="Labels definition CSV (c4.csv)", 
                    default="labels/c4.csv")
args = parser.parse_args()
SAMPLES = args.samples
MODEL = args.model
FEATURES = args.features
ENCODERS = args.encoders
LABELS_DEF = args.labels_def

# Charger le fichier de définition des labels pour créer le mapping inverse
print("[INFO] Loading labels definition from:", LABELS_DEF)
if not os.path.exists(LABELS_DEF):
    # Essayer depuis le répertoire du script
    script_dir = os.path.dirname(os.path.abspath(__file__))
    labels_def_path = os.path.join(script_dir, LABELS_DEF)
    if os.path.exists(labels_def_path):
        LABELS_DEF = labels_def_path
    else:
        print(f"[ERROR] Labels definition file not found: {LABELS_DEF}")
        sys.exit(1)

# Lire le CSV ligne par ligne pour obtenir le numéro de ligne exact
number_to_label = {}
line_number = 0
with open(LABELS_DEF, 'r', encoding='utf-8') as f:
    reader = csv.reader(f)
    for row in reader:
        line_number += 1
        # Ignorer la ligne 1 (header) et les lignes qui sont des headers
        if line_number == 1:
            continue
        if len(row) > 0:
            label = row[0].strip()
            # Ignorer les headers et lignes vides
            if label and label != "Label" and label != "Label Complet" and not label.startswith("#"):
                # Le numéro correspond au numéro de ligne dans le fichier (ligne 2 = numéro 1)
                # car ligne 1 est le header, donc on soustrait 1
                number_to_label[line_number - 1] = label

print(f"[INFO] Loaded {len(number_to_label)} label mappings")

print("[INFO] Loading model…")
# Charger le modèle (RandomForest uniquement)
model = joblib.load(MODEL)

print("[INFO] Loading encoders from:", ENCODERS)
# Charger les encoders depuis un fichier séparé
if os.path.exists(ENCODERS):
    encoders = joblib.load(ENCODERS)
else:
    # Compatibilité avec l'ancien format (modèle contenant model_data)
    print("[WARNING] Encoders file not found, trying legacy format…")
    try:
        model_data = joblib.load(MODEL)
        if isinstance(model_data, dict) and 'model' in model_data:
            model = model_data['model']
            encoders = model_data.get('encoders', {})
        else:
            encoders = {}
    except:
        encoders = {}
    if not encoders:
        print("[WARNING] No encoders found, categorical columns will not be encoded")

print("[INFO] Loading features…")
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
# Garder les colonnes "class" et "file" pour l'affichage, même si elles ne sont pas dans les features
has_class = "class" in df.columns or "stmt_name" in df.columns
has_file = "file" in df.columns or "file_path" in df.columns

if "stmt_name" in df.columns:
    df.rename(columns={"stmt_name": "class"}, inplace=True)
if "file_path" in df.columns:
    df.rename(columns={"file_path": "file"}, inplace=True)

print("[INFO] Preparing dataset…")
missing = [c for c in features if c not in df.columns]
if missing:
    print("[ERROR] Missing required columns:", missing)
    sys.exit(1)

# On ne garde que les colonnes nécessaires pour la prédiction
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
for (idx, row), label_num in zip(df.iterrows(), preds):
    # Convertir le numéro de label en string
    label_str = number_to_label.get(int(label_num), f"UNKNOWN({label_num})")
    
    # Afficher class et file si disponibles
    if has_class and "class" in df.columns:
        cls = row["class"]
    else:
        cls = "N/A"
    
    if has_file and "file" in df.columns:
        file = row["file"]
    else:
        file = "N/A"
    
    print(f"{cls} | {file} => {label_str}")
