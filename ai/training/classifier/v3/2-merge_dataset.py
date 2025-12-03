# merge_dataset.py
import pandas as pd
import sys
from argparse import ArgumentParser

parser = ArgumentParser()
parser.add_argument("csv", help="Input CSV")
parser.add_argument("labels", help="Labels CSV")
parser.add_argument("out", help="Output CSV")
args = parser.parse_args()

SAMPLES = args.csv
LABELS = args.labels
OUT = args.out

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

print(f"[INFO] Final dataset: {len(df)} rows")

# Sauvegarder
print("[INFO] Saving to:", OUT)
df.to_csv(OUT, index=False)

print("[OK] final_dataset.csv ready.")
