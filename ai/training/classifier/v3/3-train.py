# train.py
import pandas as pd
from sklearn.model_selection import train_test_split
from sklearn.ensemble import RandomForestClassifier
import joblib
import json
from argparse import ArgumentParser

# DATA = "dataset/final_dataset.csv"
# MODEL_OUT = "model.pkl"
# FEATURES_OUT = "features.json"


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
IGNORE = ["class", "file", "namespace_raw", "externals_raw", 
          "method_calls_raw", "uses_raw", "path_raw", "label", "stmt_type"]

# Garder seulement features numériques
features = [c for c in df.columns if c not in IGNORE]

print("[INFO] Features used:", features)

X = df[features]
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
joblib.dump(model, MODEL_OUT)

print("[INFO] Saving features:", FEATURES_OUT)
with open(FEATURES_OUT, "w") as f:
    json.dump(features, f)

print("[OK] Training finished.")
