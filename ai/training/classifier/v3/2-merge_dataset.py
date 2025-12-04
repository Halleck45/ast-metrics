# merge_dataset.py
import pandas as pd
import sys
import os
import csv
from argparse import ArgumentParser

parser = ArgumentParser()
parser.add_argument("csv", help="Input CSV")
parser.add_argument("labels", help="Labels CSV")
parser.add_argument("out", help="Output CSV")
parser.add_argument("--labels-def", help="Labels definition CSV (c4.csv)", 
                    default="labels/c4.csv")
args = parser.parse_args()

SAMPLES = args.csv
LABELS = args.labels
OUT = args.out
LABELS_DEF = args.labels_def

# Charger le fichier de définition des labels pour créer le mapping
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
label_to_number = {}
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
                label_to_number[label] = line_number - 1

print(f"[INFO] Loaded {len(label_to_number)} label mappings")

print("[INFO] Loading samples from:", SAMPLES)
df_samples = pd.read_csv(SAMPLES)

# Vérifier que le fichier samples n'est pas vide
if len(df_samples) == 0:
    print("[ERROR] Samples CSV is empty (only header)!")
    print("[INFO] Please ensure the samples CSV contains data rows.")
    sys.exit(1)

print("[INFO] Loading labels from:", LABELS)
df_labels = pd.read_csv(LABELS)

# Filtrer les lignes invalides dans le fichier de labels
# (lignes qui ressemblent à des headers ou commentaires)
print("[INFO] Filtering invalid label rows…")
before_filter = len(df_labels)

# Filtrer les lignes où 'class' ou 'file' ressemblent à des headers
invalid_patterns = [
    "stmt_name", "file_path", "class", "file", 
    "Label", "Label Complet", "#"
]
def is_valid_row(row):
    class_val = str(row.get('class', '')).strip()
    file_val = str(row.get('file', '')).strip()
    label_val = str(row.get('label', '')).strip()
    
    # Ignorer les lignes où class ou file ressemblent à des headers
    if any(pattern.lower() in class_val.lower() for pattern in invalid_patterns):
        return False
    if any(pattern.lower() in file_val.lower() for pattern in invalid_patterns):
        return False
    # Ignorer les lignes vides
    if not class_val or not file_val or not label_val:
        return False
    return True

df_labels = df_labels[df_labels.apply(is_valid_row, axis=1)]
after_filter = len(df_labels)
if before_filter != after_filter:
    print(f"[INFO] Filtered out {before_filter - after_filter} invalid label rows")

# Vérifier que le fichier labels n'est pas vide après filtrage
if len(df_labels) == 0:
    print("[ERROR] Labels CSV is empty after filtering invalid rows!")
    print("[INFO] Please ensure the labels CSV contains valid data rows.")
    print("[INFO] Expected format: class,file,label")
    sys.exit(1)

# Vérifier les colonnes nécessaires dans df_samples
required_samples_cols = ["stmt_name", "file_path"]
missing_samples_cols = [col for col in required_samples_cols if col not in df_samples.columns]
if missing_samples_cols:
    print(f"[ERROR] Missing required columns in samples CSV: {missing_samples_cols}")
    print(f"[INFO] Available columns: {list(df_samples.columns)}")
    sys.exit(1)

# Vérifier les colonnes nécessaires dans df_labels
required_labels_cols = ["class", "file", "label"]
missing_labels_cols = [col for col in required_labels_cols if col not in df_labels.columns]
if missing_labels_cols:
    print(f"[ERROR] Missing required columns in labels CSV: {missing_labels_cols}")
    print(f"[INFO] Available columns: {list(df_labels.columns)}")
    sys.exit(1)

# Normalisation des colonnes du dataset d'entrée
print("[INFO] Normalizing column names…")
df_samples.rename(columns={
    "stmt_name": "class",
    "file_path": "file"
}, inplace=True)

print(f"[INFO] Samples: {len(df_samples)} rows")
print(f"[INFO] Labels: {len(df_labels)} rows")

# Afficher des exemples pour le debug
if len(df_samples) > 0:
    print("\n[DEBUG] Sample examples (first 3 rows):")
    print(df_samples[["class", "file"]].head(3).to_string())
    
if len(df_labels) > 0:
    print("\n[DEBUG] Label examples (first 3 rows):")
    print(df_labels[["class", "file", "label"]].head(3).to_string())

# Vérifier les types et valeurs uniques pour le debug
print("\n[DEBUG] Checking data types and sample values…")
print(f"  Samples 'class' dtype: {df_samples['class'].dtype}")
print(f"  Labels 'class' dtype: {df_labels['class'].dtype}")
print(f"  Samples 'file' dtype: {df_samples['file'].dtype}")
print(f"  Labels 'file' dtype: {df_labels['file'].dtype}")

# Normaliser les types de données pour faciliter le merge
df_samples['class'] = df_samples['class'].astype(str).str.strip()
df_samples['file'] = df_samples['file'].astype(str).str.strip()
df_labels['class'] = df_labels['class'].astype(str).str.strip()
df_labels['file'] = df_labels['file'].astype(str).str.strip()

# Merge
print("\n[INFO] Merging on ['class', 'file']…")
df = df_samples.merge(df_labels, on=["class", "file"], how="inner")

if len(df) == 0:
    print("[ERROR] No matching rows found after merge!")
    print("\n[DEBUG] Diagnostic information:")
    print(f"  Unique 'class' values in samples: {df_samples['class'].nunique()}")
    print(f"  Unique 'class' values in labels: {df_labels['class'].nunique()}")
    print(f"  Unique 'file' values in samples: {df_samples['file'].nunique()}")
    print(f"  Unique 'file' values in labels: {df_labels['file'].nunique()}")
    
    # Vérifier s'il y a des valeurs communes
    common_classes = set(df_samples['class'].unique()) & set(df_labels['class'].unique())
    common_files = set(df_samples['file'].unique()) & set(df_labels['file'].unique())
    print(f"  Common 'class' values: {len(common_classes)}")
    print(f"  Common 'file' values: {len(common_files)}")
    
    if len(common_classes) > 0 and len(common_files) > 0:
        print("\n[DEBUG] Sample of common values:")
        sample_class = list(common_classes)[:3]
        sample_file = list(common_files)[:3]
        print(f"  Sample classes: {sample_class}")
        print(f"  Sample files: {sample_file}")
    
    print("\n[INFO] This might indicate a mismatch between sample and label data.")
    print("[INFO] Check that 'class' and 'file' values match between the two CSVs.")
    print("[INFO] Note: Both 'class' and 'file' must match simultaneously for a merge.")
    sys.exit(1)

print(f"[INFO] Merged dataset: {len(df)} rows")

# Supprimer les UNKNOWN
print("[INFO] Removing UNKNOWN labels…")
before_unknown = len(df)
df = df[df["label"] != "UNKNOWN"]
removed = before_unknown - len(df)
if removed > 0:
    print(f"[INFO] Removed {removed} rows with UNKNOWN label")

if len(df) == 0:
    print("[WARNING] No rows remaining after removing UNKNOWN labels!")
    sys.exit(1)

# Convertir les labels en numéros
print("[INFO] Converting labels to numbers…")
def label_to_num(label_str):
    label_str = str(label_str).strip()
    # Extraire uniquement la partie avant "#" si elle existe (format: "label # description")
    if "#" in label_str:
        label_str = label_str.split("#")[0].strip()
    if label_str in label_to_number:
        return label_to_number[label_str]
    else:
        print(f"[WARNING] Unknown label '{label_str}', skipping")
        return None

df["label"] = df["label"].apply(label_to_num)
df = df.dropna(subset=["label"])
df["label"] = df["label"].astype(int)

if len(df) == 0:
    print("[WARNING] No rows remaining after label conversion!")
    sys.exit(1)

# Retirer les colonnes "class" et "file" du CSV final
print("[INFO] Removing 'class' and 'file' columns from final dataset…")
df_final = df.drop(columns=["class", "file"])

print(f"[INFO] Final dataset: {len(df_final)} rows")

# Sauvegarder
print("[INFO] Saving to:", OUT)
df_final.to_csv(OUT, index=False)

print("[OK] final_dataset.csv ready.")
