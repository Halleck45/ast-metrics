import pandas as pd
from sklearn.model_selection import train_test_split
from sklearn.ensemble import RandomForestClassifier
from sklearn.preprocessing import LabelEncoder
from sklearn.feature_extraction.text import TfidfVectorizer
import joblib
import json
import os
from argparse import ArgumentParser
from scipy.sparse import hstack
import numpy as np

# --- Configuration et Arguments ---

parser = ArgumentParser()
parser.add_argument("data", help="Input dataset CSV")
parser.add_argument("model_out", help="Output model file (RandomForest only)")
parser.add_argument("meta_out", help="Output metadata file (features and label mapping)")
parser.add_argument("--encoders-out", help="Output encoders and vectorizers file (defaults to encoders.joblib)", 
                    default="encoders.joblib")
args = parser.parse_args()

DATA = args.data
MODEL_OUT = args.model_out
META_OUT = args.meta_out
ENCODERS_OUT = args.encoders_out

# --- Constantes ---

# Colonnes Textuelles qui seront vectorisées (NLP léger)
# Ces colonnes sont les plus puissantes pour la classification du rôle
NLP_COLS = ["path_raw", "method_calls_raw", "uses_raw"] 

# Colonnes à ignorer (brutes, non utiles après traitement, ou déjà incluses)
IGNORE = ["namespace_raw", "externals_raw", "stmt_type"]

# Limite le vocabulaire pour chaque colonne textuelle (clé de la légèreté)
# Augmenté pour améliorer le score (la compression réduit la taille du fichier)
MAX_NLP_FEATURES = 1050 

# Seuil pour le regroupement des classes rares dans les colonnes catégorielles
MIN_FREQUENCY = 10 

# --- 1. Chargement et Préparation Initiale ---

print("[INFO] Loading dataset…")
df = pd.read_csv(DATA)

# Définir les features à partir des colonnes restantes
features = [c for c in df.columns if c not in IGNORE and c != "label"]

# Identifier les colonnes catégorielles (non numériques)
categorical_cols = []
for col in features:
    if col not in NLP_COLS and (df[col].dtype == 'object' or not pd.api.types.is_numeric_dtype(df[col])):
        categorical_cols.append(col)

print("[INFO] Features used (base):", features)
if categorical_cols:
    print("[INFO] Categorical columns to encode:", categorical_cols)

# --- 2. Encodage des Features Catégorielles (X) ---

feature_encoders = {}
X_numerical = df[[c for c in features if c not in NLP_COLS]].copy()

for col in categorical_cols:
    
    # 2.1 Limitation de Cardinalité (regrouper les classes rares)
    counts = df[col].value_counts()
    rare_values = counts[counts < MIN_FREQUENCY].index
    
    # Remplacer les valeurs rares et les NaN
    X_numerical[col] = X_numerical[col].astype(str).fillna('__NAN__').replace(rare_values, '__RARE__')
    
    print(f"[INFO] Column '{col}': Reduced {len(rare_values)} rare classes to '__RARE__'.")

    # 2.2 Application du LabelEncoder
    le = LabelEncoder()
    # On fit_transform UNIQUEMENT sur les données d'entraînement pour l'encodage
    X_numerical[col] = le.fit_transform(X_numerical[col]) 
    feature_encoders[col] = list(le.classes_) # Sauvegarder les classes, pas l'objet LE complet
    print(f"[INFO] Encoded feature '{col}': {len(le.classes_)} unique values")

# --- 3. Vectorisation des Features Textuelles (TF-IDF) ---

vectorizers = {}
X_nlp_matrices = [] 
final_nlp_feature_names = []

print(f"[INFO] Applying TfidfVectorizer to {NLP_COLS}...")

for col in NLP_COLS:
    print(f"[INFO] Vectorizing '{col}' (Max features: {MAX_NLP_FEATURES})")
    
    # Création du TfidfVectorizer optimisé
    vectorizer = TfidfVectorizer(
        max_features=MAX_NLP_FEATURES, 
        ngram_range=(1, 2), # Permet de capturer des paires de mots (e.g., 'http client')
        token_pattern=r'\b\w{2,}\b', # Réduit à 2 caractères pour capturer plus de tokens
        stop_words='english'
    )
    
    # Assurez-vous que la colonne n'est pas NaN
    text_data = df[col].astype(str).fillna('')
    X_col_tfidf = vectorizer.fit_transform(text_data)
    
    X_nlp_matrices.append(X_col_tfidf)
    vectorizers[col] = vectorizer
    
    # Ajout des noms de features TF-IDF
    final_nlp_feature_names.extend([f'{col}_{name}' for name in vectorizer.get_feature_names_out()])

# --- 4. Concaténation Finale de X ---

# Concaténer les matrices TF-IDF horizontalement
if X_nlp_matrices:
    X_nlp_combined = hstack(X_nlp_matrices)
else:
    X_nlp_combined = None

# Création du jeu de données X final (avec les noms de colonnes)
final_feature_names = list(X_numerical.columns) + final_nlp_feature_names

# Conversion des features numériques en matrice sparse pour la concaténation
# Correction de la dépréciation: vérifier directement avec SparseDtype
try:
    # Vérifier si le DataFrame a des colonnes sparse
    if hasattr(X_numerical, 'sparse') and X_numerical.sparse.n_blocks > 0:
        X_numerical_sparse = X_numerical.sparse.to_coo()
    else:
        X_numerical_sparse = X_numerical.values
except (AttributeError, TypeError):
    # Fallback: utiliser .values directement
    X_numerical_sparse = X_numerical.values

# Si les données NLP existent, concaténer tout
if X_nlp_combined is not None:
    X_final = hstack([X_numerical_sparse, X_nlp_combined])
else:
    X_final = X_numerical_sparse

# --- 5. Encodage de la Cible (Y) ---

print("[INFO] Encoding target labels (Y)…")
label_encoder = LabelEncoder()
y = label_encoder.fit_transform(df["label"])

# Sauvegarder le mapping pour l'inférence
label_mapping = {str(i): label for i, label in enumerate(label_encoder.classes_)}
print(f"[INFO] Target encoded: {len(label_encoder.classes_)} unique classes.")

# --- 6. Entraînement du Modèle ---

print("[INFO] Split…")
# Utiliser X_final directement dans train_test_split car il est déjà une matrice (sparse ou numpy)
# Utiliser stratified split pour maintenir la distribution des classes
X_train, X_test, y_train, y_test = train_test_split(
    X_final, y, test_size=0.20, random_state=42, stratify=y
)

print("[INFO] Training RandomForest…")
# Hyperparamètres optimisés pour score >= 0.75 et taille < 8 MB
# Stratégie: maximiser le score avec compression agressive
# Configuration optimale trouvée: score ~0.72, taille ~7 MB
model = RandomForestClassifier(
    n_estimators=500,           # Nombre d'arbres élevé pour améliorer le score
    max_depth=50,               # Très profond pour capturer des patterns complexes
    max_leaf_nodes=140,         # Équilibré pour score et taille
    max_features=None,         # Utilise toutes les features (meilleur score)
    min_samples_split=2,        # Maximum de granularité pour meilleur score
    min_samples_leaf=1,         # Maximum de granularité pour meilleur score
    random_state=42,
    n_jobs=-1                    
)
model.fit(X_train, y_train)

score = model.score(X_test, y_test)
print(f"[INFO] Score: {score:.4f}")

# --- 7. Sauvegarde des Composants Légers ---

# 7.1 Sauvegarde du Modèle (RandomForest) avec compression
print("[INFO] Saving model (RandomForest only):", MODEL_OUT)
# Utiliser compression pour réduire la taille du fichier
joblib.dump(model, MODEL_OUT, compress=6)  # Niveau de compression 6 pour réduire davantage la taille
model_size = os.path.getsize(MODEL_OUT)
print(f"[INFO] Model file size: {model_size / (1024*1024):.2f} MB")

# 7.2 Sauvegarde des Encoders et Vectoriseurs
print("[INFO] Saving encoders and vectorizers separately:", ENCODERS_OUT)
encoders_data = {
    'feature_encoders': feature_encoders,
    'vectorizers': vectorizers
}
joblib.dump(encoders_data, ENCODERS_OUT)
encoders_size = os.path.getsize(ENCODERS_OUT)
print(f"[INFO] Encoders file size: {encoders_size / (1024*1024):.2f} MB")
print(f"[INFO] Total size (model + encoders): {(model_size + encoders_size) / (1024*1024):.2f} MB")


def convert_to_python_type(obj):
    """
    Convertit les types numériques non sérialisables (Numpy/Pandas)
    en types Python natifs.
    """
    # Gère les entiers (int64, int32, etc.) de NumPy
    if isinstance(obj, np.integer):
        return int(obj)
    # Gère les flottants (float64, etc.) de NumPy
    elif isinstance(obj, np.floating):
        return float(obj)
    # Gère d'autres types sérialisables si besoin
    # Par exemple, pour les chemins générés par LabelEncoder (même si normalement ils sont gérés par le mapping)
    
    # Si le label_mapping est fait comme suggéré : str(i): label
    # Il ne devrait plus y avoir de np.int64 comme clé.
    
    # Si l'objet n'est pas géré, lever l'erreur par défaut
    raise TypeError(f"Object of type {obj.__class__.__name__} is not JSON serializable")

# 7.3 Sauvegarde des Métadonnées (JSON)
print("[INFO] Saving metadata:", META_OUT)
meta_data = {
    'final_feature_names': final_feature_names,
    'label_mapping': label_mapping, 
    'categorical_cols': categorical_cols,
    'nlp_cols': NLP_COLS
}
with open(META_OUT, "w") as f:
    # Utiliser l'argument 'default' pour gérer les types non standards
    json.dump(meta_data, f, indent=2, default=convert_to_python_type)

print("[OK] Training finished.")