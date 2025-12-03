# merge_dataset.py
import pandas as pd
from argparse import ArgumentParser

# SAMPLES = "dataset/samples.csv"
# LABELS = "classified_output/classified_c4.csv"
# OUT = "dataset/final_dataset.csv"

parser = ArgumentParser()
parser.add_argument("csv", help="Input CSV")
parser.add_argument("labels", help="Labels CSV")
parser.add_argument("out", help="Output CSV")
args = parser.parse_args()

SAMPLES = args.csv
LABELS = args.labels
OUT = args.out

print("[INFO] Loading…")
df_samples = pd.read_csv(SAMPLES)
df_labels = pd.read_csv(LABELS)

# Normalisation des colonnes
df_samples.rename(columns={
    "stmt_name": "class",
    "file_path": "file"
}, inplace=True)

df_labels.rename(columns={
    "class": "class",
    "file": "file",
    "label": "label"
}, inplace=True)

print("[INFO] Merging…")
df = df_samples.merge(df_labels, on=["class", "file"], how="inner")

print("[INFO] Removing UNKNOWN…")
df = df[df["label"] != "UNKNOWN"]

print("[INFO] Saving", OUT)
df.to_csv(OUT, index=False)

print("[OK] final_dataset.csv ready.")
