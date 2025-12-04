# train.py
import pandas as pd
from sklearn.model_selection import train_test_split
from sklearn.ensemble import RandomForestClassifier
from sklearn.preprocessing import LabelEncoder
import joblib
import json
import os
from argparse import ArgumentParser

parser = ArgumentParser()
parser.add_argument("data", help="Input dataset CSV")
parser.add_argument("model_out", help="Output model file (RandomForest only)")
parser.add_argument("features_out", help="Output features JSON")
parser.add_argument("--encoders-out", help="Output encoders file (optional, defaults to encoders.pkl)", 
                    default="encoders.pkl")
args = parser.parse_args()

DATA = args.data
MODEL_OUT = args.model_out
FEATURES_OUT = args.features_out
ENCODERS_OUT = args.encoders_out


print("[INFO] Loading dataset…")
df = pd.read_csv(DATA)

# Colonnes à ignorer
IGNORE = ["namespace_raw", "externals_raw", 
          "method_calls_raw", "uses_raw", "path_raw", "label", "stmt_type"]

# Garder seulement features numériques
features = [c for c in df.columns if c not in IGNORE]

# Identifier les colonnes catégorielles (non numériques)
categorical_cols = []
for col in features:
    if df[col].dtype == 'object' or not pd.api.types.is_numeric_dtype(df[col]):
        categorical_cols.append(col)

print("[INFO] Features used:", features)
if categorical_cols:
    print("[INFO] Categorical columns to encode:", categorical_cols)

# Encoder les colonnes catégorielles
encoders = {}
X = df[features].copy()

for col in categorical_cols:
    le = LabelEncoder()
    X[col] = le.fit_transform(X[col].astype(str))
    encoders[col] = le
    print(f"[INFO] Encoded '{col}': {len(le.classes_)} unique values")

y = df["label"]

print("[INFO] Split…")
X_train, X_test, y_train, y_test = train_test_split(
    X, y, test_size=0.20, random_state=42
)

print("[INFO] Training RandomForest (optimized for size)…")
# Hyperparamètres optimisés pour réduire la taille du modèle :
# - n_estimators réduit : 400 -> 100 (réduit la taille de ~75%)
# - max_depth limité : None -> 20 (évite les arbres trop profonds)
# - max_leaf_nodes : limite le nombre de feuilles par arbre
# - min_samples_split : force plus de données avant de splitter
model = RandomForestClassifier(
    n_estimators=100,           # Réduit de 400 à 100 (taille ~75% plus petite)
    max_depth=20,                # Limite la profondeur (au lieu de None)
    max_leaf_nodes=100,          # Limite le nombre de feuilles par arbre
    min_samples_split=10,        # Minimum d'échantillons pour splitter
    min_samples_leaf=4,          # Minimum d'échantillons par feuille
    random_state=42,
    n_jobs=-1                    # Utilise tous les CPU disponibles
)
model.fit(X_train, y_train)

score = model.score(X_test, y_test)
print("[INFO] Score:", score)

print("[INFO] Saving model (RandomForest only):", MODEL_OUT)
# Sauvegarder uniquement le modèle RandomForestClassifier (sans les encoders)
# Les encoders sont volumineux car ils contiennent toutes les classes uniques
joblib.dump(model, MODEL_OUT)

# Calculer la taille réelle du fichier modèle sauvegardé
model_size = os.path.getsize(MODEL_OUT)
print(f"[INFO] Model file size: {model_size / (1024*1024):.2f} MB")

print("[INFO] Saving encoders separately:", ENCODERS_OUT)
# Sauvegarder les encoders dans un fichier séparé
# Cela permet de réduire la taille du fichier modèle principal
joblib.dump(encoders, ENCODERS_OUT)

# Calculer la taille du fichier encoders
encoders_size = os.path.getsize(ENCODERS_OUT)
print(f"[INFO] Encoders file size: {encoders_size / (1024*1024):.2f} MB")
print(f"[INFO] Total size (model + encoders): {(model_size + encoders_size) / (1024*1024):.2f} MB")

print("[INFO] Saving features:", FEATURES_OUT)
features_data = {
    'features': features,
    'categorical_cols': categorical_cols
}
with open(FEATURES_OUT, "w") as f:
    json.dump(features_data, f, indent=2)

print("[OK] Training finished.")
