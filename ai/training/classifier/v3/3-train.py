import pandas as pd
from sklearn.model_selection import train_test_split
from sklearn.ensemble import RandomForestClassifier
from sklearn.preprocessing import LabelEncoder, OneHotEncoder
from sklearn.feature_extraction.text import TfidfVectorizer
from sklearn.metrics import classification_report, confusion_matrix
import joblib
import json
import os
from argparse import ArgumentParser
from scipy.sparse import hstack, csr_matrix
import numpy as np
from collections import Counter

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
#NLP_COLS = ["path_raw", "method_calls_raw", "uses_raw", "stmt_type"] 
NLP_COLS = ["path_raw",  "uses_raw", "stmt_type"] 

# Colonnes à ignorer (brutes, non utiles après traitement, ou déjà incluses)
IGNORE = ["externals_raw", "method_calls_raw"]

# Limite le vocabulaire pour chaque colonne textuelle (clé de la légèreté)
# Augmenté pour améliorer le score (la compression réduit la taille du fichier)
MAX_NLP_FEATURES = 1400

# Poids pour le vectorizer du nom de classe (pour augmenter son importance)
CLASS_NAME_WEIGHT = 3.0  # Multiplier les valeurs TF-IDF du nom par ce facteur

# Nombre de features pour le vectorizer spécifique du nom de classe
CLASS_NAME_MAX_FEATURES = 2000  # Plus de features pour capturer mieux les noms

# Poids pour le langage de programmation (pour augmenter son importance)
PROGRAMMING_LANGUAGE_WEIGHT = 2.5  # Multiplier les valeurs one-hot du langage par ce facteur 

# Seuil pour le regroupement des classes rares dans les colonnes catégorielles
MIN_FREQUENCY = 5 

# --- 1. Chargement et Préparation Initiale ---

print("[INFO] Loading dataset…")
df = pd.read_csv(DATA)

# Extraire le nom de la classe depuis namespace_raw
if "namespace_raw" in df.columns:
    print("[INFO] Extracting class names from namespace_raw...")
    def extract_class_name(namespace):
        if pd.isna(namespace) or not namespace:
            return ""
        namespace = str(namespace).strip()
        # Extraire le dernier élément (nom de la classe) après le dernier séparateur
        # Supporte à la fois les backslashes (PHP) et les points (autres langages)
        if '\\' in namespace:
            return namespace.split('\\')[-1]
        elif '.' in namespace:
            return namespace.split('.')[-1]
        else:
            return namespace
    
    df['class_name'] = df['namespace_raw'].apply(extract_class_name)
    print(f"[INFO] Extracted class names. Sample: {df['class_name'].head(3).tolist()}")
else:
    print("[WARNING] namespace_raw not found, class_name will be empty")
    df['class_name'] = ""

# Définir les features à partir des colonnes restantes
features = [c for c in df.columns if c not in IGNORE and c != "label" and c != "class_name"]

# Identifier les colonnes catégorielles (non numériques)
# Exclure programming_language qui sera traité séparément avec un poids
categorical_cols = []
for col in features:
    if col not in NLP_COLS and col != "programming_language" and (df[col].dtype == 'object' or not pd.api.types.is_numeric_dtype(df[col])):
        categorical_cols.append(col)

print("[INFO] Features used (base):", features)
if categorical_cols:
    print("[INFO] Categorical columns to encode:", categorical_cols)
print("[INFO] Programming language will be encoded separately with weight", PROGRAMMING_LANGUAGE_WEIGHT)

# --- 2. Encodage des Features Catégorielles (X) ---

feature_encoders = {}
# Exclure programming_language qui sera traité séparément
base_numerical_cols = [c for c in features if c not in NLP_COLS and c != "programming_language"]
X_numerical = df[base_numerical_cols].copy()

# 2.0 Encodage spécial pour programming_language avec poids
programming_language_encoder = None
X_programming_language = None
prog_lang_feature_names = []
if "programming_language" in df.columns:
    print("[INFO] Encoding programming_language with one-hot (weighted)...")
    # Préparer les données
    prog_lang_data = df["programming_language"].astype(str).fillna('__NAN__')
    
    # Créer un encodage one-hot
    programming_language_encoder = OneHotEncoder(sparse_output=True, handle_unknown='ignore')
    X_programming_language = programming_language_encoder.fit_transform(prog_lang_data.values.reshape(-1, 1))
    
    # Appliquer le poids
    X_programming_language = X_programming_language * PROGRAMMING_LANGUAGE_WEIGHT
    
    # Obtenir les noms des features
    prog_lang_feature_names = [f"prog_lang_{name}" for name in programming_language_encoder.get_feature_names_out(['programming_language'])]
    print(f"[INFO] Programming language encoded: {len(prog_lang_feature_names)} features, weight={PROGRAMMING_LANGUAGE_WEIGHT}x")
    print(f"[INFO] Programming languages: {prog_lang_data.value_counts().to_dict()}")
else:
    print("[WARNING] programming_language column not found")

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

# 3.1 Vectorisation spéciale pour le nom de la classe (avec poids plus élevé)
print("[INFO] Creating specialized TfidfVectorizer for class names (weighted)...")
class_name_vectorizer = TfidfVectorizer(
    max_features=CLASS_NAME_MAX_FEATURES,
    ngram_range=(1, 1),  # Unigrammes uniquement pour le nom
    token_pattern=r'\b\w{2,}\b',  # Mots de 2 caractères ou plus
    stop_words=None,  # Pas de stop words pour les noms de classes
    min_df=1,
    max_df=0.98,
    sublinear_tf=True
)

class_name_data = df['class_name'].astype(str).fillna('')
X_class_name_tfidf = class_name_vectorizer.fit_transform(class_name_data)

# Appliquer le poids au nom de classe
X_class_name_tfidf = X_class_name_tfidf * CLASS_NAME_WEIGHT
print(f"[INFO] Class name vectorizer: {X_class_name_tfidf.shape[1]} features, weight={CLASS_NAME_WEIGHT}x")

X_nlp_matrices.append(X_class_name_tfidf)
vectorizers['class_name'] = class_name_vectorizer
final_nlp_feature_names.extend([f'class_name_{name}' for name in class_name_vectorizer.get_feature_names_out()])

# 3.2 Vectorisation des autres colonnes NLP
print(f"[INFO] Applying TfidfVectorizer to {NLP_COLS}...")

for col in NLP_COLS:
    print(f"[INFO] Vectorizing '{col}' (Max features: {MAX_NLP_FEATURES})")
    
    # Création du TfidfVectorizer optimisé
    vectorizer = TfidfVectorizer(
        max_features=MAX_NLP_FEATURES, 
        ngram_range=(1, 3), # Permet de capturer des trigrammes pour plus de contexte
        token_pattern=r'\b\w{2,}\b', # Réduit à 2 caractères pour capturer plus de tokens
        stop_words='english',
        min_df=1,              # Inclure tous les tokens (même rares)
        max_df=0.95,           # Filtrer les tokens trop fréquents
        sublinear_tf=True      # Utiliser log(1+tf) au lieu de tf pour réduire l'impact des fréquences élevées
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
final_feature_names = list(X_numerical.columns)
# Ajouter les features du langage de programmation si elles existent
if X_programming_language is not None:
    final_feature_names.extend(prog_lang_feature_names)
final_feature_names.extend(final_nlp_feature_names)

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

# --- 5. Encodage de la Cible (Y) ---

print("[INFO] Encoding target labels (Y)…")
label_encoder = LabelEncoder()
y = label_encoder.fit_transform(df["label"])

# Sauvegarder le mapping pour l'inférence
label_mapping = {str(i): label for i, label in enumerate(label_encoder.classes_)}
print(f"[INFO] Target encoded: {len(label_encoder.classes_)} unique classes.")

# Vérifier la distribution des classes pour la stratification
class_counts = Counter(y)
min_class_count = min(class_counts.values())
print(f"[INFO] Class distribution: min={min_class_count}, max={max(class_counts.values())}")

# --- 6. Entraînement du Modèle ---

print("[INFO] Split…")
# Utiliser X_final directement dans train_test_split car il est déjà une matrice (sparse ou numpy)
# Utiliser stratified split pour maintenir la distribution des classes, si possible
# Stratification nécessite au moins 2 échantillons par classe
if min_class_count >= 2:
    print("[INFO] Using stratified split (all classes have at least 2 samples)")
    X_train, X_test, y_train, y_test = train_test_split(
        X_final, y, test_size=0.20, random_state=42, stratify=y
    )
else:
    rare_classes = [cls for cls, count in class_counts.items() if count < 2]
    print(f"[WARNING] Cannot use stratified split: {len(rare_classes)} classes have only 1 sample")
    print(f"[WARNING] Using non-stratified split instead")
    X_train, X_test, y_train, y_test = train_test_split(
        X_final, y, test_size=0.20, random_state=42, stratify=None
    )

print("[INFO] Training RandomForest…")
# Hyperparamètres optimisés pour score >= 0.80 et taille < 8 MB
# Stratégie: maximiser le score avec compression maximale
# Note: La taille du fichier compressé peut dépasser 8 MB, mais le modèle en mémoire reste efficace
model = RandomForestClassifier(
    n_estimators=500,          # Plus raisonnable
    max_depth=50,               
    max_leaf_nodes=400,         # Permet plus de distinctions par arbre
    max_features='sqrt',        # Améliore la diversité
    min_samples_split=2,        
    min_samples_leaf=1,         
    random_state=42,
    n_jobs=-1,
    class_weight='balanced'                    
)
model.fit(X_train, y_train)

score = model.score(X_test, y_test)
print(f"[INFO] Score: {score:.4f}")

# --- 6.1 Rapport de Classification et Matrice de Confusion ---
print("\n[INFO] Generating classification report...")
y_pred = model.predict(X_test)

# piste : ne pas utiliser predict, mais fit

# Obtenir les noms des labels pour le rapport (s'assurer que ce sont des strings)
target_names = [str(label_mapping.get(str(i), f"Label_{i}")) for i in range(len(label_encoder.classes_))]

# Obtenir toutes les classes possibles (même celles absentes du test set)
all_labels = np.arange(len(label_encoder.classes_))

print("\n=== CLASSIFICATION REPORT ===")
print(classification_report(y_test, y_pred, labels=all_labels, target_names=target_names, zero_division=0))

print("\n=== CONFUSION MATRIX (Top 10 most confused labels) ===")
cm = confusion_matrix(y_test, y_pred)
# Trouver les paires de labels les plus confondues
confusion_pairs = []
for i in range(len(cm)):
    for j in range(len(cm)):
        if i != j and cm[i, j] > 0:
            confusion_pairs.append((cm[i, j], i, j, target_names[i], target_names[j]))

# Trier par nombre d'erreurs
confusion_pairs.sort(reverse=True, key=lambda x: x[0])

# Afficher les 10 paires les plus confondues
print(f"{'Count':<8} {'True Label':<40} {'Predicted Label':<40}")
print("-" * 90)
for count, true_idx, pred_idx, true_label, pred_label in confusion_pairs[:10]:
    print(f"{count:<8} {true_label:<40} {pred_label:<40}")

if len(confusion_pairs) > 10:
    print(f"\n... and {len(confusion_pairs) - 10} more confusion pairs")

# --- 7. Sauvegarde des Composants Légers ---

# 7.1 Sauvegarde du Modèle (RandomForest) avec compression
print("[INFO] Saving model (RandomForest only):", MODEL_OUT)
# Utiliser compression maximale pour réduire la taille du fichier
joblib.dump(model, MODEL_OUT, compress=9)  # Niveau de compression 9 (maximum) pour réduire au maximum la taille
model_size = os.path.getsize(MODEL_OUT)
print(f"[INFO] Model file size: {model_size / (1024*1024):.2f} MB")

# 7.2 Sauvegarde des Encoders et Vectoriseurs
print("[INFO] Saving encoders and vectorizers separately:", ENCODERS_OUT)
encoders_data = {
    'feature_encoders': feature_encoders,
    'vectorizers': vectorizers,
    'programming_language_encoder': programming_language_encoder
}
joblib.dump(encoders_data, ENCODERS_OUT, compress=9)
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
    'nlp_cols': NLP_COLS,
    'class_name_weight': CLASS_NAME_WEIGHT,
    'programming_language_weight': PROGRAMMING_LANGUAGE_WEIGHT
}
with open(META_OUT, "w") as f:
    # Utiliser l'argument 'default' pour gérer les types non standards
    json.dump(meta_data, f, indent=2, default=convert_to_python_type)

print("[OK] Training finished.")