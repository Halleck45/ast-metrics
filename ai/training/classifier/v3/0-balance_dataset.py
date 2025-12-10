#!/usr/bin/env python3
"""
Script pour équilibrer le dataset par langage de programmation.
Prend le dataset original, le sauvegarde avec _full.csv, et crée un dataset équilibré.
"""
import pandas as pd
import sys
import os
from argparse import ArgumentParser
from collections import Counter

parser = ArgumentParser(description="Balance dataset by programming language")
parser.add_argument("input_csv", help="Input dataset CSV file")
parser.add_argument("--output", help="Output balanced dataset CSV (default: input name without extension + .csv)", 
                    default=None)
parser.add_argument("--min-samples", type=int, default=None,
                    help="Minimum number of samples per language (default: use the smallest language count)")
parser.add_argument("--strategy", choices=["language", "label", "both"], default="language",
                    help="Balancing strategy: by language, by label, or both (default: language)")
args = parser.parse_args()

INPUT_CSV = args.input_csv
if args.output:
    OUTPUT_CSV = args.output
else:
    # Créer le nom de sortie en ajoutant .csv si nécessaire
    base_name = os.path.splitext(INPUT_CSV)[0]
    OUTPUT_CSV = base_name + ".csv"

# Nom du fichier complet (avec _full)
FULL_CSV = os.path.splitext(INPUT_CSV)[0] + "_full.csv"

print(f"[INFO] Loading dataset from: {INPUT_CSV}")
df = pd.read_csv(INPUT_CSV)

print(f"[INFO] Dataset shape: {df.shape}")
print(f"[INFO] Saving full dataset to: {FULL_CSV}")
df.to_csv(FULL_CSV, index=False)
print(f"[INFO] Full dataset saved ({len(df)} rows)")

# Vérifier que la colonne programming_language existe
if "programming_language" not in df.columns:
    print("[ERROR] Column 'programming_language' not found in dataset")
    sys.exit(1)

# Afficher la distribution actuelle
print("\n[INFO] Current distribution by programming language:")
lang_counts = df["programming_language"].value_counts()
print(lang_counts)

if args.strategy == "language":
    # Équilibrer par langage de programmation
    print(f"\n[INFO] Balancing by programming language...")
    
    # Déterminer le nombre d'échantillons par langage
    if args.min_samples:
        samples_per_lang = args.min_samples
        print(f"[INFO] Using {samples_per_lang} samples per language (user specified)")
    else:
        # Utiliser le nombre d'échantillons du langage le moins représenté
        samples_per_lang = lang_counts.min()
        print(f"[INFO] Using {samples_per_lang} samples per language (minimum available)")
    
    # Échantillonner équitablement par langage
    balanced_dfs = []
    for lang in lang_counts.index:
        lang_df = df[df["programming_language"] == lang]
        if len(lang_df) >= samples_per_lang:
            # Échantillonner aléatoirement
            sampled = lang_df.sample(n=samples_per_lang, random_state=42)
            balanced_dfs.append(sampled)
            print(f"  {lang}: {len(sampled)} samples (from {len(lang_df)} available)")
        else:
            # Prendre tous les échantillons disponibles
            balanced_dfs.append(lang_df)
            print(f"  {lang}: {len(lang_df)} samples (all available, less than {samples_per_lang})")
    
    df_balanced = pd.concat(balanced_dfs, ignore_index=True)
    
elif args.strategy == "label":
    # Équilibrer par label
    print(f"\n[INFO] Balancing by label...")
    
    if "label" not in df.columns:
        print("[ERROR] Column 'label' not found in dataset")
        sys.exit(1)
    
    label_counts = df["label"].value_counts()
    print(f"[INFO] Label distribution (top 10):")
    print(label_counts.head(10))
    
    if args.min_samples:
        samples_per_label = args.min_samples
        print(f"[INFO] Using {samples_per_label} samples per label (user specified)")
    else:
        samples_per_label = label_counts.min()
        print(f"[INFO] Using {samples_per_label} samples per label (minimum available)")
    
    # Échantillonner équitablement par label
    balanced_dfs = []
    for label in label_counts.index:
        label_df = df[df["label"] == label]
        if len(label_df) >= samples_per_label:
            sampled = label_df.sample(n=samples_per_label, random_state=42)
            balanced_dfs.append(sampled)
        else:
            balanced_dfs.append(label_df)
    
    df_balanced = pd.concat(balanced_dfs, ignore_index=True)
    
elif args.strategy == "both":
    # Équilibrer par combinaison langage + label
    print(f"\n[INFO] Balancing by programming language AND label...")
    
    if "label" not in df.columns:
        print("[ERROR] Column 'label' not found in dataset")
        sys.exit(1)
    
    # Compter les combinaisons langage+label
    combo_counts = df.groupby(["programming_language", "label"]).size()
    print(f"[INFO] Found {len(combo_counts)} unique language+label combinations")
    
    if args.min_samples:
        samples_per_combo = args.min_samples
        print(f"[INFO] Using {samples_per_combo} samples per combination (user specified)")
    else:
        samples_per_combo = combo_counts.min()
        print(f"[INFO] Using {samples_per_combo} samples per combination (minimum available)")
    
    # Échantillonner équitablement par combinaison
    balanced_dfs = []
    for (lang, label), count in combo_counts.items():
        combo_df = df[(df["programming_language"] == lang) & (df["label"] == label)]
        if len(combo_df) >= samples_per_combo:
            sampled = combo_df.sample(n=samples_per_combo, random_state=42)
            balanced_dfs.append(sampled)
        else:
            balanced_dfs.append(combo_df)
    
    df_balanced = pd.concat(balanced_dfs, ignore_index=True)

# Mélanger les données équilibrées
print(f"\n[INFO] Shuffling balanced dataset...")
df_balanced = df_balanced.sample(frac=1, random_state=42).reset_index(drop=True)

# Afficher la nouvelle distribution
print(f"\n[INFO] Balanced dataset shape: {df_balanced.shape}")
print(f"[INFO] New distribution by programming language:")
print(df_balanced["programming_language"].value_counts())

if "label" in df_balanced.columns:
    print(f"\n[INFO] New distribution by label (top 10):")
    print(df_balanced["label"].value_counts().head(10))

# Sauvegarder le dataset équilibré
print(f"\n[INFO] Saving balanced dataset to: {OUTPUT_CSV}")
df_balanced.to_csv(OUTPUT_CSV, index=False)
print(f"[INFO] Balanced dataset saved ({len(df_balanced)} rows)")
print(f"[INFO] Reduction: {len(df)} -> {len(df_balanced)} rows ({len(df_balanced)/len(df)*100:.1f}% of original)")

