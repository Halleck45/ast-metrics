# train.py
import pandas as pd
from sklearn.model_selection import train_test_split
from sklearn.ensemble import RandomForestClassifier
from sklearn.preprocessing import LabelEncoder
import joblib
import json
from argparse import ArgumentParser

parser = ArgumentParser()
parser.add_argument("data", help="Input dataset CSV")
parser.add_argument("model_out", help="Output model file")
parser.add_argument("features_out", help="Output features JSON")
args = parser.parse_args()

DATA = args.data
MODEL_OUT = args.model_out
FEATURES_OUT = args.features_out


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

print("[INFO] Training RandomForest…")
model = RandomForestClassifier(
    n_estimators=400,
    max_depth=None,
    random_state=42
)
model.fit(X_train, y_train)

score = model.score(X_test, y_test)
print("[INFO] Score:", score)

print("[INFO] Saving model:", MODEL_OUT)
# Sauvegarder le modèle et les encodeurs ensemble
model_data = {
    'model': model,
    'encoders': encoders
}
joblib.dump(model_data, MODEL_OUT)

print("[INFO] Saving features:", FEATURES_OUT)
features_data = {
    'features': features,
    'categorical_cols': categorical_cols
}
with open(FEATURES_OUT, "w") as f:
    json.dump(features_data, f, indent=2)

print("[OK] Training finished.")
