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

print("[INFO] Loading labels from:", LABELS)
df_labels = pd.read_csv(LABELS)

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

# Merge
print("[INFO] Merging on ['class', 'file']…")
df = df_samples.merge(df_labels, on=["class", "file"], how="inner")

if len(df) == 0:
    print("[WARNING] No matching rows found after merge!")
    print("[INFO] This might indicate a mismatch between sample and label data.")
    print("[INFO] Check that 'class' and 'file' values match between the two CSVs.")
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
