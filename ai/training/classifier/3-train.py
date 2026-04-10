#!/usr/bin/env python3
"""
Model Training Script (Hashing TF-IDF + RandomForest)

- Keeps RandomForestClassifier (same model family)
- Replaces TfidfVectorizer (vocabulary) by HashingVectorizer + TfidfTransformer (no vocabulary)
- Keeps weights:
  - CLASS_NAME_WEIGHT
  - PROGRAMMING_LANGUAGE_WEIGHT
- Produces artifacts:
  - model.pkl (joblib, compressed)
  - encoders.joblib (feature_encoders + vectorizers + programming_language_encoder)
  - features.json (metadata)

Usage:
  python 3-train.py dataset/final_dataset.csv build/php/model.pkl build/php/features.json --encoders-out=build/php/encoders.joblib

Note:
  - Export script must expect vectorizers to be sklearn Pipeline(hv->tfidf).
"""

import os
import json
from argparse import ArgumentParser
from collections import Counter

import numpy as np
import pandas as pd
import joblib

from scipy.sparse import hstack, csr_matrix

from sklearn.model_selection import train_test_split
from sklearn.ensemble import RandomForestClassifier
from sklearn.preprocessing import LabelEncoder, OneHotEncoder
from sklearn.metrics import classification_report, confusion_matrix, f1_score
from sklearn.model_selection import StratifiedKFold, cross_val_score

from sklearn.pipeline import Pipeline
from sklearn.feature_extraction.text import HashingVectorizer, TfidfTransformer


# --- Configuration et Arguments ---

parser = ArgumentParser()
parser.add_argument("data", help="Input dataset CSV")
parser.add_argument("model_out", help="Output model file (RandomForest only)")
parser.add_argument("meta_out", help="Output metadata file (features and label mapping)")
parser.add_argument(
    "--encoders-out",
    help="Output encoders and vectorizers file (defaults to encoders.joblib)",
    default="encoders.joblib",
)
args = parser.parse_args()

DATA = args.data
MODEL_OUT = args.model_out
META_OUT = args.meta_out
ENCODERS_OUT = args.encoders_out

# --- Constantes ---

# Colonnes Textuelles (NLP léger)
NLP_COLS = ["path_raw", "uses_raw", "stmt_type"]

# Poids par colonne NLP (uses_raw = imports, très discriminant; stmt_type = 2 valeurs, faible signal)
NLP_WEIGHTS = {"path_raw": 1.5, "uses_raw": 2.0, "stmt_type": 0.5}

# Colonnes à ignorer
IGNORE = ["externals_raw", "method_calls_raw"]

# Features mortes (100% zéros dans le dataset PHP) ou quasi-mortes (<0.1% non-zéros)
DEAD_FEATURES = [
    "nb_comments", "nb_extends", "nb_implements", "nb_traits", "count_elseif",
    "count_case", "count_switch",
]

# Hashing bins (remplace max_features/vocab)
# 2^15 = 32768 bins : bon compromis entre collision et taille pour ~34k samples
N_HASH = 1 << 15  # 32768

# Poids
CLASS_NAME_WEIGHT = 3.0
PROGRAMMING_LANGUAGE_WEIGHT = 1.0  # 1.0 pour modèle mono-langage (PHP)

# Seuil pour classes rares (categorical cols)
MIN_FREQUENCY = 5

# RandomForest hyperparameters (tuned for 63-class imbalanced classification)
RF_N_ESTIMATORS = 200
RF_MAX_DEPTH = None       # arbres profonds pour max qualité
RF_MAX_LEAF_NODES = None  # pas de limite
RF_MIN_SAMPLES_SPLIT = 5
RF_MIN_SAMPLES_LEAF = 2   # feuilles fines pour capturer les classes rares
RF_MAX_FEATURES = "sqrt"

RANDOM_STATE = 42


def extract_class_name(namespace):
    if pd.isna(namespace) or not namespace:
        return ""
    namespace = str(namespace).strip()
    if "\\" in namespace:
        return namespace.split("\\")[-1]
    if "." in namespace:
        return namespace.split(".")[-1]
    return namespace


def make_hashing_tfidf_pipeline(n_features: int, ngram_range=(1, 3), sublinear_tf=True) -> Pipeline:
    """
    TF-IDF sans vocabulaire:
    HashingVectorizer -> TfidfTransformer

    Important:
      - alternate_sign=False pour éviter des valeurs négatives (plus stable)
      - norm=None au niveau hashing; normalisation en fin via TfidfTransformer(norm="l2")
    """
    return Pipeline(
        steps=[
            (
                "hv",
                HashingVectorizer(
                    n_features=n_features,
                    alternate_sign=False,
                    norm=None,
                    lowercase=True,
                    token_pattern=r"(?u)\b\w\w+\b",
                    ngram_range=ngram_range,
                    stop_words="english",
                ),
            ),
            (
                "tfidf",
                TfidfTransformer(
                    norm="l2",
                    use_idf=True,
                    smooth_idf=True,
                    sublinear_tf=sublinear_tf,
                ),
            ),
        ]
    )


def make_class_name_pipeline(n_features: int) -> Pipeline:
    # Class names: unigrams only, no stop words
    return Pipeline(
        steps=[
            (
                "hv",
                HashingVectorizer(
                    n_features=n_features,
                    alternate_sign=False,
                    norm=None,
                    lowercase=True,
                    token_pattern=r"(?u)\b\w\w+\b",
                    ngram_range=(1, 1),
                    stop_words=None,
                ),
            ),
            (
                "tfidf",
                TfidfTransformer(
                    norm="l2",
                    use_idf=True,
                    smooth_idf=True,
                    sublinear_tf=True,
                ),
            ),
        ]
    )


def convert_to_python_type(obj):
    if isinstance(obj, np.integer):
        return int(obj)
    if isinstance(obj, np.floating):
        return float(obj)
    raise TypeError(f"Object of type {obj.__class__.__name__} is not JSON serializable")


# --- 1. Chargement et Préparation Initiale ---

print("[INFO] Loading dataset…")
df = pd.read_csv(DATA)

# --- Normalisation des chemins (supprimer le préfixe absolu commun) ---
# Les chemins absolus polluent le TF-IDF (le modèle apprend /home/user/... au lieu des patterns de projet)
if "path_raw" in df.columns:
    paths = df["path_raw"].astype(str).tolist()
    common_prefix = os.path.commonprefix(paths)
    # Ne couper que jusqu'au dernier / pour garder des chemins propres
    if "/" in common_prefix:
        common_prefix = common_prefix[:common_prefix.rfind("/") + 1]
    if len(common_prefix) > 10:
        print(f"[INFO] Stripping common path prefix: '{common_prefix}' ({len(common_prefix)} chars)")
        df["path_raw"] = df["path_raw"].astype(str).str[len(common_prefix):]
        print(f"[INFO] Sample normalized paths: {df['path_raw'].head(3).tolist()}")

# class_name depuis namespace_raw
if "namespace_raw" in df.columns:
    print("[INFO] Extracting class names from namespace_raw...")
    df["class_name"] = df["namespace_raw"].apply(extract_class_name)
    print(f"[INFO] Extracted class names. Sample: {df['class_name'].head(3).tolist()}")
else:
    print("[WARNING] namespace_raw not found, class_name will be empty")
    df["class_name"] = ""

# --- Features dérivées (ratios discriminants pour l'architecture) ---
DERIVED_FEATURES = {
    "getter_ratio":          lambda d: d["nb_getters"] / d["nb_methods"].replace(0, 1),
    "setter_ratio":          lambda d: d["nb_setters"] / d["nb_methods"].replace(0, 1),
    "loc_per_method":        lambda d: d["class_loc"] / d["nb_methods"].replace(0, 1),
    "complexity_per_method": lambda d: d["cyclomatic_complexity"] / d["nb_methods"].replace(0, 1),
    "attribute_per_method":  lambda d: d["nb_attributes"] / d["nb_methods"].replace(0, 1),
    "comment_ratio":         lambda d: d["comment_loc"] / d["class_loc"].replace(0, 1),
    "method_call_density":   lambda d: d["nb_method_calls"] / d["nb_methods"].replace(0, 1),
    "dep_per_method":        lambda d: d["nb_external_dependencies"] / d["nb_methods"].replace(0, 1),
}

print("[INFO] Computing derived features...")
for name, fn in DERIVED_FEATURES.items():
    df[name] = fn(df).clip(0, 100)
print(f"[INFO] Added {len(DERIVED_FEATURES)} derived features: {list(DERIVED_FEATURES.keys())}")

# Features (base)
features = [c for c in df.columns if c not in IGNORE and c not in DEAD_FEATURES and c != "label" and c != "class_name"]

# Cat cols (hors NLP + prog lang)
categorical_cols = []
for col in features:
    if col not in NLP_COLS and col != "programming_language" and (df[col].dtype == "object" or not pd.api.types.is_numeric_dtype(df[col])):
        categorical_cols.append(col)

print("[INFO] Features used (base):", features)
if categorical_cols:
    print("[INFO] Categorical columns to encode:", categorical_cols)
print("[INFO] Programming language weight:", PROGRAMMING_LANGUAGE_WEIGHT)
print("[INFO] Hashing bins:", N_HASH)

# --- 2. Encodage des Features Catégorielles (X) ---

feature_encoders = {}
base_numerical_cols = [c for c in features if c not in NLP_COLS and c != "programming_language"]
X_numerical = df[base_numerical_cols].copy()

# 2.0 programming_language one-hot (weighted)
programming_language_encoder = None
X_programming_language = None
prog_lang_feature_names = []  # plus utile en hashing, mais on le garde pour debug/meta

if "programming_language" in df.columns:
    print("[INFO] Encoding programming_language with one-hot (weighted)...")
    prog_lang_data = df["programming_language"].astype(str).fillna("__NAN__")

    programming_language_encoder = OneHotEncoder(sparse_output=True, handle_unknown="ignore")
    X_programming_language = programming_language_encoder.fit_transform(prog_lang_data.values.reshape(-1, 1))
    X_programming_language = X_programming_language * PROGRAMMING_LANGUAGE_WEIGHT

    prog_lang_feature_names = [
        f"prog_lang_{name}" for name in programming_language_encoder.get_feature_names_out(["programming_language"])
    ]
    print(f"[INFO] Programming language encoded: {len(prog_lang_feature_names)} features, weight={PROGRAMMING_LANGUAGE_WEIGHT}x")
else:
    print("[WARNING] programming_language column not found")

# 2.1 other categorical cols with LabelEncoder (and rare bucket)
for col in categorical_cols:
    counts = df[col].value_counts()
    rare_values = counts[counts < MIN_FREQUENCY].index

    X_numerical[col] = X_numerical[col].astype(str).fillna("__NAN__").replace(rare_values, "__RARE__")
    print(f"[INFO] Column '{col}': Reduced {len(rare_values)} rare classes to '__RARE__'.")

    le = LabelEncoder()
    X_numerical[col] = le.fit_transform(X_numerical[col])
    feature_encoders[col] = list(le.classes_)
    print(f"[INFO] Encoded feature '{col}': {len(le.classes_)} unique values")

# --- 3. Vectorisation des Features Textuelles (Hashing TF-IDF) ---

vectorizers = {}
X_nlp_matrices = []

# 3.1 class_name hashing tf-idf (weighted)
print("[INFO] Creating hashing TF-IDF pipeline for class_name (weighted)...")
class_name_pipe = make_class_name_pipeline(n_features=N_HASH)
class_name_data = df["class_name"].astype(str).fillna("")
X_class_name_tfidf = class_name_pipe.fit_transform(class_name_data)

# Apply weight
X_class_name_tfidf = X_class_name_tfidf * CLASS_NAME_WEIGHT
print(f"[INFO] class_name hashing TF-IDF: {X_class_name_tfidf.shape[1]} bins, weight={CLASS_NAME_WEIGHT}x")

X_nlp_matrices.append(X_class_name_tfidf)
vectorizers["class_name"] = class_name_pipe

# 3.2 other NLP cols hashing tf-idf (with per-column weights)
print(f"[INFO] Applying hashing TF-IDF to {NLP_COLS} with weights {NLP_WEIGHTS}...")
for col in NLP_COLS:
    weight = NLP_WEIGHTS.get(col, 1.0)
    print(f"[INFO] Vectorizing '{col}' (hash bins: {N_HASH}, weight: {weight}x)")
    pipe = make_hashing_tfidf_pipeline(n_features=N_HASH, ngram_range=(1, 3), sublinear_tf=True)

    text_data = df[col].astype(str).fillna("")
    X_col = pipe.fit_transform(text_data)

    if weight != 1.0:
        X_col = X_col * weight

    X_nlp_matrices.append(X_col)
    vectorizers[col] = pipe

# --- 4. Concaténation Finale de X ---

if X_nlp_matrices:
    X_nlp_combined = hstack(X_nlp_matrices)
else:
    X_nlp_combined = None

# numeric -> sparse
try:
    if hasattr(X_numerical, "sparse") and X_numerical.sparse.n_blocks > 0:
        X_numerical_sparse = X_numerical.sparse.to_coo()
    else:
        X_numerical_sparse = csr_matrix(X_numerical.values)
except Exception:
    X_numerical_sparse = csr_matrix(X_numerical.values)

matrices_to_stack = [X_numerical_sparse]

if X_programming_language is not None:
    matrices_to_stack.append(X_programming_language)

if X_nlp_combined is not None:
    matrices_to_stack.append(X_nlp_combined)

X_final = hstack(matrices_to_stack).tocsr()
print(f"[INFO] Final X shape: {X_final.shape}")

# --- 4.5 Fusion des labels ultra-rares ---

# Labels avec < 20 samples sont fusionnés vers leur label parent le plus proche
LABEL_MERGES = {
    54: 55,  # utility:helper:math (2 samples) → utility:helper:component
    60: 59,  # utility:serialization:deserializer (17) → utility:serialization:serializer
}
if LABEL_MERGES:
    before_count = df["label"].nunique()
    df["label"] = df["label"].replace(LABEL_MERGES)
    after_count = df["label"].nunique()
    print(f"[INFO] Merged ultra-rare labels: {before_count} → {after_count} unique labels")

# --- 5. Encodage de la Cible (Y) ---

# Load label names from roles.csv to map numeric IDs back to string labels
LABELS_DEF = os.path.join(os.path.dirname(os.path.abspath(__file__)), "labels", "roles.csv")
number_to_label_name = {}
if os.path.exists(LABELS_DEF):
    import csv as csv_mod
    with open(LABELS_DEF, "r", encoding="utf-8") as f:
        reader = csv_mod.reader(f)
        for line_num, row in enumerate(reader, start=1):
            if line_num == 1:
                continue  # skip header
            if row and row[0].strip() and not row[0].startswith("#"):
                number_to_label_name[line_num - 1] = row[0].strip()
    print(f"[INFO] Loaded {len(number_to_label_name)} label name mappings from roles.csv")
else:
    print(f"[WARNING] Labels definition not found: {LABELS_DEF}")

print("[INFO] Encoding target labels (Y)…")
label_encoder = LabelEncoder()
y = label_encoder.fit_transform(df["label"].astype(str))

# Map sklearn class index → label name string (not number)
# label_encoder.classes_ contains the sorted unique values of df["label"] as strings of numbers
label_mapping = {}
for i, cls_str in enumerate(label_encoder.classes_):
    num = int(cls_str)
    if num in number_to_label_name:
        label_mapping[str(i)] = number_to_label_name[num]
    else:
        label_mapping[str(i)] = cls_str  # fallback to number
        print(f"[WARNING] No label name for number {num}")
print(f"[INFO] Target encoded: {len(label_encoder.classes_)} unique classes.")
print(f"[INFO] Sample label mapping: {dict(list(label_mapping.items())[:5])}")

class_counts = Counter(y)
min_class_count = min(class_counts.values())
print(f"[INFO] Class distribution: min={min_class_count}, max={max(class_counts.values())}")

# --- 6. Entraînement du Modèle ---

print("[INFO] Split…")
if min_class_count >= 2:
    print("[INFO] Using stratified split")
    X_train, X_test, y_train, y_test = train_test_split(
        X_final, y, test_size=0.20, random_state=RANDOM_STATE, stratify=y
    )
else:
    rare_classes = [cls for cls, count in class_counts.items() if count < 2]
    print(f"[WARNING] Non-stratified split: {len(rare_classes)} classes have only 1 sample")
    X_train, X_test, y_train, y_test = train_test_split(
        X_final, y, test_size=0.20, random_state=RANDOM_STATE, stratify=None
    )

print("[INFO] Training RandomForest (class_weight='balanced' handles imbalance)…")
model = RandomForestClassifier(
    n_estimators=RF_N_ESTIMATORS,
    max_depth=RF_MAX_DEPTH,
    max_leaf_nodes=RF_MAX_LEAF_NODES,
    max_features=RF_MAX_FEATURES,
    min_samples_split=RF_MIN_SAMPLES_SPLIT,
    min_samples_leaf=RF_MIN_SAMPLES_LEAF,
    random_state=RANDOM_STATE,
    n_jobs=-1,
    class_weight="balanced",
)
model.fit(X_train, y_train)

score = model.score(X_test, y_test)
print(f"[INFO] Accuracy: {score:.4f}")

# --- 6.1 Cross-validation (on training data) ---

print("\n[INFO] Running 5-fold cross-validation on training set...")
cv = StratifiedKFold(n_splits=5, shuffle=True, random_state=RANDOM_STATE)
cv_scores = cross_val_score(
    RandomForestClassifier(
        n_estimators=RF_N_ESTIMATORS, max_depth=RF_MAX_DEPTH,
        max_leaf_nodes=RF_MAX_LEAF_NODES, max_features=RF_MAX_FEATURES,
        min_samples_split=RF_MIN_SAMPLES_SPLIT, min_samples_leaf=RF_MIN_SAMPLES_LEAF,
        random_state=RANDOM_STATE, n_jobs=-1, class_weight="balanced",
    ),
    X_train, y_train, cv=cv, scoring="f1_macro", n_jobs=1,
)
print(f"[INFO] CV F1-macro: {cv_scores.mean():.4f} ± {cv_scores.std():.4f}")

# --- 6.2 Rapport de Classification et Matrice de Confusion ---
print("\n[INFO] Generating classification report...")
y_pred = model.predict(X_test)

# F1 scores on holdout
f1_macro = f1_score(y_test, y_pred, average="macro", zero_division=0)
f1_weighted = f1_score(y_test, y_pred, average="weighted", zero_division=0)
print(f"[INFO] Holdout F1-macro:    {f1_macro:.4f}")
print(f"[INFO] Holdout F1-weighted: {f1_weighted:.4f}")

target_names = [str(label_mapping.get(str(i), f"Label_{i}")) for i in range(len(label_encoder.classes_))]
all_labels = np.arange(len(label_encoder.classes_))

print("\n=== CLASSIFICATION REPORT ===")
print(classification_report(y_test, y_pred, labels=all_labels, target_names=target_names, zero_division=0))

print("\n=== CONFUSION MATRIX (Top 10 most confused labels) ===")
cm = confusion_matrix(y_test, y_pred)
confusion_pairs = []
for i in range(len(cm)):
    for j in range(len(cm)):
        if i != j and cm[i, j] > 0:
            confusion_pairs.append((cm[i, j], i, j, target_names[i], target_names[j]))

confusion_pairs.sort(reverse=True, key=lambda x: x[0])

print(f"{'Count':<8} {'True Label':<40} {'Predicted Label':<40}")
print("-" * 90)
for count, true_idx, pred_idx, true_label, pred_label in confusion_pairs[:10]:
    print(f"{count:<8} {true_label:<40} {pred_label:<40}")

if len(confusion_pairs) > 10:
    print(f"\n... and {len(confusion_pairs) - 10} more confusion pairs")

# --- 7. Sauvegarde des Composants ---

print("[INFO] Saving model (RandomForest only):", MODEL_OUT)
joblib.dump(model, MODEL_OUT, compress=9)
model_size = os.path.getsize(MODEL_OUT)
print(f"[INFO] Model file size: {model_size / (1024*1024):.2f} MB")

print("[INFO] Saving encoders and vectorizers separately:", ENCODERS_OUT)
encoders_data = {
    "feature_encoders": feature_encoders,
    "vectorizers": vectorizers,  # pipelines hv->tfidf
    "programming_language_encoder": programming_language_encoder,
}
joblib.dump(encoders_data, ENCODERS_OUT, compress=9)
encoders_size = os.path.getsize(ENCODERS_OUT)
print(f"[INFO] Encoders file size: {encoders_size / (1024*1024):.2f} MB")
print(f"[INFO] Total size (model + encoders): {(model_size + encoders_size) / (1024*1024):.2f} MB")

# Meta: plus de final_feature_names (pas de vocab)
print("[INFO] Saving metadata:", META_OUT)
# Build hashing-safe final_feature_names (synthetic names, but stable order)
final_feature_names = []

# 1) Base numerical + categorical-encoded columns (exact order used in X_numerical)
# X_numerical.columns is the truth here
final_feature_names.extend(list(X_numerical.columns))

# 2) Programming language one-hot columns (exact order produced by encoder)
if X_programming_language is not None:
    final_feature_names.extend(prog_lang_feature_names)

# 3) class_name hashing tf-idf block
# One feature name per hashing bin to keep exact ordering reproducible in Go
final_feature_names.extend([f"class_name__hash_{i}" for i in range(N_HASH)])

# 4) NLP hashing tf-idf blocks in the exact NLP_COLS order
for col in NLP_COLS:
    final_feature_names.extend([f"{col}__hash_{i}" for i in range(N_HASH)])

# sanity: must match X_final shape
assert len(final_feature_names) == X_final.shape[1], (len(final_feature_names), X_final.shape[1])
meta_data = {
    "final_feature_names": [], # empty to optimize size
    "label_mapping": label_mapping,
    "categorical_cols": categorical_cols,
    "nlp_cols": NLP_COLS,
    "class_name_weight": CLASS_NAME_WEIGHT,
    "programming_language_weight": PROGRAMMING_LANGUAGE_WEIGHT,

    # IMPORTANT: new fields for Go alignment without final_feature_names
    "numerical_cols_order": list(X_numerical.columns),
    "hashing_n_features": int(N_HASH),
    "derived_features": list(DERIVED_FEATURES.keys()),
    "nlp_weights": NLP_WEIGHTS,
}
with open(META_OUT, "w", encoding="utf-8") as f:
    json.dump(meta_data, f, indent=2, default=convert_to_python_type, ensure_ascii=False)

print("[OK] Training finished.")
