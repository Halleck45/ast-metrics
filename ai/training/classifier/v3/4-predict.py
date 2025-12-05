# predict_batch.py
import sys
import os
import csv
import pandas as pd
import joblib
import json
import numpy as np
from argparse import ArgumentParser
from sklearn.preprocessing import LabelEncoder, OneHotEncoder
from scipy.sparse import hstack

parser = ArgumentParser()
parser.add_argument("samples", help="Input samples CSV")
parser.add_argument("model", help="Input model file (RandomForest only)")
parser.add_argument("features", help="Input features JSON")
parser.add_argument("--encoders", help="Input encoders file (defaults to encoders.joblib)", 
                    default="encoders.joblib")
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

print("[INFO] Loading features…")
features_data = json.load(open(FEATURES, "r"))
if isinstance(features_data, dict):
    final_feature_names = features_data.get('final_feature_names', [])
    categorical_cols = features_data.get('categorical_cols', [])
    nlp_cols = features_data.get('nlp_cols', [])
    # Compatibilité avec l'ancien format
    if not final_feature_names:
        final_feature_names = features_data.get('features', [])
else:
    # Compatibilité avec l'ancien format
    final_feature_names = features_data
    categorical_cols = []
    nlp_cols = []

print("[INFO] Loading encoders from:", ENCODERS)
# Charger les encoders depuis un fichier séparé
encoders_data = {}
if os.path.exists(ENCODERS):
    encoders_data = joblib.load(ENCODERS)
    feature_encoders = encoders_data.get('feature_encoders', {})
    vectorizers = encoders_data.get('vectorizers', {})
    programming_language_encoder = encoders_data.get('programming_language_encoder', None)
else:
    # Compatibilité avec l'ancien format
    print("[WARNING] Encoders file not found, trying legacy format…")
    try:
        model_data = joblib.load(MODEL)
        if isinstance(model_data, dict) and 'model' in model_data:
            model = model_data['model']
            encoders_data = model_data.get('encoders', {})
            feature_encoders = encoders_data.get('feature_encoders', {})
            vectorizers = encoders_data.get('vectorizers', {})
        else:
            feature_encoders = {}
            vectorizers = {}
            programming_language_encoder = None
    except:
        feature_encoders = {}
        vectorizers = {}
        programming_language_encoder = None
    if not feature_encoders and not vectorizers and programming_language_encoder is None:
        print("[WARNING] No encoders found, categorical columns will not be encoded")

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

# Extraire le nom de la classe depuis namespace_raw (comme dans le script d'entraînement)
if "namespace_raw" in df.columns:
    print("[INFO] Extracting class names from namespace_raw...")
    def extract_class_name(namespace):
        if pd.isna(namespace) or not namespace:
            return ""
        namespace = str(namespace).strip()
        # Extraire le dernier élément (nom de la classe) après le dernier séparateur
        if '\\' in namespace:
            return namespace.split('\\')[-1]
        elif '.' in namespace:
            return namespace.split('.')[-1]
        else:
            return namespace
    
    df['class_name'] = df['namespace_raw'].apply(extract_class_name)
else:
    print("[WARNING] namespace_raw not found, class_name will be empty")
    df['class_name'] = ""

print("[INFO] Preparing dataset…")

# Identifier les colonnes de base nécessaires (sans les features NLP vectorisées)
base_features = [c for c in final_feature_names if not any(c.startswith(nlp_col + '_') for nlp_col in nlp_cols)]
base_features = [c for c in base_features if not c.startswith('class_name_')]  # Exclure les features du nom de classe
base_features = [c for c in base_features if not c.startswith('prog_lang_')]  # Exclure les features du langage de programmation
base_features = [c for c in base_features if c not in nlp_cols and c != 'class_name' and c != 'programming_language']

# Vérifier les colonnes manquantes (namespace_raw est optionnel car on peut extraire class_name)
required_cols = base_features + nlp_cols
# namespace_raw n'est pas requis si class_name est déjà présent
if 'namespace_raw' in required_cols and 'class_name' in df.columns:
    required_cols = [c for c in required_cols if c != 'namespace_raw']
    
missing = [c for c in required_cols if c not in df.columns]
if missing:
    print("[ERROR] Missing required columns:", missing)
    sys.exit(1)

# Préparer les features numériques et catégorielles
# IMPORTANT: Maintenir l'ordre exact des colonnes comme dans l'entraînement
base_numerical_cols = [c for c in base_features if c not in nlp_cols]

# Vérifier que toutes les colonnes sont présentes
missing_cols = [c for c in base_numerical_cols if c not in df.columns]
if missing_cols:
    print(f"[ERROR] Missing columns in dataset: {missing_cols}")
    sys.exit(1)

# Créer X_numerical dans l'ordre exact de base_numerical_cols
X_numerical = df[base_numerical_cols].copy()

# Encoder le langage de programmation avec poids (comme dans l'entraînement)
X_programming_language = None
if programming_language_encoder is not None and "programming_language" in df.columns:
    print("[INFO] Encoding programming_language with one-hot (weighted)...")
    prog_lang_data = df["programming_language"].astype(str).fillna('__NAN__')
    X_programming_language = programming_language_encoder.transform(prog_lang_data.values.reshape(-1, 1))
    
    # Appliquer le poids
    programming_language_weight = features_data.get('programming_language_weight', 2.5)
    X_programming_language = X_programming_language * programming_language_weight
    print(f"[INFO] Programming language encoded with weight {programming_language_weight}x")
elif "programming_language" in df.columns:
    print("[WARNING] No programming_language encoder found, skipping weighted encoding")

# Encoder les colonnes catégorielles
MIN_FREQUENCY = 10  # Même valeur que dans le script d'entraînement
for col in categorical_cols:
    if col in feature_encoders:
        # Reconstruire le LabelEncoder à partir des classes sauvegardées
        le = LabelEncoder()
        le.classes_ = np.array(feature_encoders[col])
        
        # Préparer les données comme dans l'entraînement
        X_numerical[col] = X_numerical[col].astype(str).fillna('__NAN__')
        
        # Remplacer les valeurs rares et inconnues
        known_classes = set(feature_encoders[col])
        unknown_mask = ~X_numerical[col].isin(known_classes)
        if unknown_mask.any():
            print(f"[WARNING] Found {unknown_mask.sum()} unknown values in '{col}', using '__RARE__'")
            X_numerical.loc[unknown_mask, col] = '__RARE__'
        
        # Vérifier que '__RARE__' est dans les classes connues (sinon le transform échouera)
        if '__RARE__' not in known_classes:
            print(f"[ERROR] '__RARE__' not found in known classes for '{col}'")
            print(f"[ERROR] Known classes: {sorted(list(known_classes))[:10]}...")
            print(f"[ERROR] This means the training data did not have rare values grouped, but prediction data does")
            sys.exit(1)
        
        # Encoder
        X_numerical[col] = le.transform(X_numerical[col])
        
        # Debug: vérifier si toutes les valeurs sont identiques après encodage
        unique_values = X_numerical[col].unique()
        if len(unique_values) == 1:
            print(f"[WARNING] Column '{col}' has only one unique value after encoding: {unique_values[0]}")
    else:
        print(f"[WARNING] No encoder found for categorical column '{col}', skipping encoding")

# Vectoriser les colonnes NLP
X_nlp_matrices = []

# 1. Vectoriser le nom de la classe avec poids (comme dans l'entraînement)
if 'class_name' in vectorizers:
    class_name_vectorizer = vectorizers['class_name']
    class_name_data = df['class_name'].astype(str).fillna('')
    X_class_name_tfidf = class_name_vectorizer.transform(class_name_data)
    
    # Appliquer le poids (récupérer depuis features.json ou utiliser la valeur par défaut)
    class_name_weight = features_data.get('class_name_weight', 3.0)
    X_class_name_tfidf = X_class_name_tfidf * class_name_weight
    X_nlp_matrices.append(X_class_name_tfidf)
    print(f"[INFO] Class name vectorized with weight {class_name_weight}x")
else:
    print("[WARNING] No vectorizer found for 'class_name', skipping")

# 2. Vectoriser les autres colonnes NLP
for col in nlp_cols:
    if col in vectorizers:
        vectorizer = vectorizers[col]
        text_data = df[col].astype(str).fillna('')
        X_col_tfidf = vectorizer.transform(text_data)
        X_nlp_matrices.append(X_col_tfidf)
    else:
        print(f"[WARNING] No vectorizer found for NLP column '{col}', skipping")

# Concaténer les matrices NLP
if X_nlp_matrices:
    X_nlp_combined = hstack(X_nlp_matrices)
else:
    X_nlp_combined = None

# Convertir les features numériques en matrice
X_numerical_sparse = X_numerical.values

# Concaténer toutes les matrices : numériques + langage de programmation + NLP
matrices_to_stack = [X_numerical_sparse]

# Ajouter le langage de programmation si disponible
if X_programming_language is not None:
    matrices_to_stack.append(X_programming_language)

# Ajouter les données NLP si disponibles
if X_nlp_combined is not None:
    matrices_to_stack.append(X_nlp_combined)

# Concaténer tout
X_final = hstack(matrices_to_stack)

# Debug: vérifier la forme et le nombre de features
print(f"[DEBUG] X_final shape: {X_final.shape}")
print(f"[DEBUG] Expected features from metadata: {len(final_feature_names)}")
if hasattr(model, 'n_features_in_'):
    print(f"[DEBUG] Model expects: {model.n_features_in_} features")
    if X_final.shape[1] != model.n_features_in_:
        print(f"[ERROR] Feature count mismatch! Got {X_final.shape[1]}, expected {model.n_features_in_}")
        print(f"[ERROR] This will cause incorrect predictions!")
        sys.exit(1)
else:
    print("[WARNING] Model does not have n_features_in_ attribute, cannot verify feature count")

print("[INFO] Predicting…")
preds = model.predict(X_final)

# Debug: vérifier la distribution des prédictions
unique_preds, counts = np.unique(preds, return_counts=True)
print(f"[DEBUG] Predictions distribution: {len(unique_preds)} unique classes predicted")
if len(unique_preds) == 1:
    print(f"[WARNING] All predictions are the same class: {unique_preds[0]}")
    print(f"[WARNING] This may indicate a problem with feature encoding or model")

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
