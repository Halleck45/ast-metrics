#!/usr/bin/env python3
"""
Local LLM Labeling Script

This script uses a local LLM (via llama-cpp-python) to classify code snippets
according to their architectural role. Labels are defined in CSV files in the labels/ directory.

Usage:
    python 1-labelize.py <input_csv> [--count=<n>] [--output-dir=<dir>] [--labels-dir=<dir>]

Requirements:
    - A GGUF model file in models/ directory
    - llama-cpp-python installed
    - Sufficient GPU/CPU resources for inference

Example:
    python 1-labelize.py dataset/samples.csv --count=10000 --output-dir=classified_output
"""
import csv
import os
import sys
import random
from argparse import ArgumentParser
from llama_cpp import Llama


# ------------------------------------------
# ARGS
# ------------------------------------------
parser = ArgumentParser()
parser.add_argument("csv", help="Input CSV (all classes)")
parser.add_argument("--count", type=int, default=1000, help="Max number of rows to classify")
parser.add_argument("--output-dir", help="Output directory for classified CSVs", default="classified_output")
parser.add_argument("--labels-dir", help="Directory containing label definitions", default="labels")
args = parser.parse_args()

INPUT_CSV = args.csv
MAX_COUNT = args.count
LABELS_DIR = args.labels_dir
OUTPUT_DIR = args.output_dir
os.makedirs(OUTPUT_DIR, exist_ok=True)


# ------------------------------------------
# MODEL CONFIG
# ------------------------------------------
MODEL_PATH = "models/Meta-Llama-3-8B-Instruct.Q4_0.gguf"
TEMPERATURE = 0.0

print("[INFO] Loading model...")
if not os.path.exists(MODEL_PATH):
    print(f"[ERROR] Model file not found: {MODEL_PATH}")
    print("[INFO] Please download a GGUF model or update MODEL_PATH in the script")
    sys.exit(1)

try:
    llm = Llama(
        model_path=MODEL_PATH,
        n_ctx=2048,
        n_threads=16,
        n_gpu_layers=999,
        embedding=False,
        chat_format="llama-3",
    )
    print("[INFO] Model ready.")
except Exception as e:
    print(f"[ERROR] Failed to load model: {e}")
    sys.exit(1)


# ------------------------------------------
# LOAD LABEL FAMILIES
# Each CSV in labels/ represents one family
# ------------------------------------------
def load_labels(path):
    print(f"[DEBUG] Loading labels from: {path}")
    labels = []
    with open(path, encoding="utf-8") as f:
        for row in csv.reader(f):
            if row and row[0] != "Label":
                #labels.append(row[0].strip())
                if len(row) > 1:
                    labels.append(row[0].strip() + " # " + row[1].strip())
                else:
                    labels.append(row[0].strip())
    print(f"[DEBUG] Loaded {len(labels)} labels from {path}")
    print(labels)
    return labels


label_families = {
    fname[:-4]: load_labels(os.path.join(LABELS_DIR, fname))
    for fname in os.listdir(LABELS_DIR)
    if fname.endswith(".csv")
}

# ------------------------------------------
# PROMPTS ULTRA COURTS
# ------------------------------------------
def build_system_prompt(labels):
    numbered = "\n".join(f"{i+1} {lbl}" for i, lbl in enumerate(labels))
    return f"Pick the correct label number.\n{numbered}\nReply only with the number."


def build_user_prompt(fullname):
    return f"FILE: {fullname}\nANSWER: "


# ------------------------------------------
# CLASSIFIER FUNCTION
# ------------------------------------------
def classify_one(fullname, system_prompt, labels):
    user_prompt = build_user_prompt(fullname)

    out = llm.create_chat_completion(
        messages=[
            {"role": "system", "content": system_prompt},
            {"role": "user", "content": user_prompt},
        ],
        max_tokens=2,
        temperature=TEMPERATURE,
        stop=["\n", " "],
        # digit bias (optional)
        logit_bias={llm.tokenize(str(d).encode())[1]: 5.0 for d in range(10)}
    )

    ans = out["choices"][0]["message"]["content"].strip()
    ans = ans.split()[0]

    if ans.isdigit() and 1 <= int(ans) <= len(labels):
        return labels[int(ans)-1]

    return "UNKNOWN"


# ------------------------------------------
# READ INPUT CSV AND SAMPLE ROWS
# ------------------------------------------
with open(INPUT_CSV, encoding="utf-8") as f:
    all_rows = [row for row in csv.reader(f) if row]

print(f"[DEBUG] Total rows in CSV: {len(all_rows)}")
if MAX_COUNT < len(all_rows):
    print(f"[DEBUG] Sampling {MAX_COUNT} rows...")
    rows = random.sample(all_rows, MAX_COUNT)
else:
    print(f"[DEBUG] Using all {len(all_rows)} rows...")
    rows = all_rows

print(f"[INFO] Using {len(rows)} sampled rows (requested {MAX_COUNT}).")

# ------------------------------------------
# PROCESS EACH FAMILY
# ------------------------------------------
print("[DEBUG] Starting family processing loop...")
for family, labels in label_families.items():
    print(f"\n[INFO] Processing family: {family}")
    print(f"[DEBUG] Labels for {family}: {len(labels)}")

    system_prompt = build_system_prompt(labels)
    out_path = os.path.join(OUTPUT_DIR, f"classified_{family}.csv")

    with open(out_path, "w", newline="", encoding="utf-8") as out_f:
        writer = csv.writer(out_f)
        writer.writerow(["class", "file", "label"])

        for i, row in enumerate(rows):
            cname = row[0].strip()
            fname = row[2].strip() if len(row) > 2 else ""

            fullname = f"{cname} in file {fname}"

            lbl = classify_one(fullname, system_prompt, labels)
            writer.writerow([cname, fname, lbl])

            if (i+1) % 50 == 0:
                print(f"[{family}] {i+1} done")

    print("[OK] Saved:", out_path)


print("\n[ALL FAMILIES DONE]")
