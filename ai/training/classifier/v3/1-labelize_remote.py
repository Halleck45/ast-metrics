import csv
import os
import sys
import random
import json
from argparse import ArgumentParser
from openai import OpenAI
from typing import List, Dict


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
# OPENAI CLIENT
# ------------------------------------------
api_key = os.environ.get("OPENAI_API_KEY")
if not api_key:
    print("[ERROR] OPENAI_API_KEY environment variable not set")
    sys.exit(1)

client = OpenAI(api_key=api_key)
MODEL = "gpt-4o"
TEMPERATURE = 0.2


# ------------------------------------------
# LOAD LABEL FAMILIES
# Each CSV in labels/ represents one family
# ------------------------------------------
def load_labels(path):
    labels = []
    with open(path, encoding="utf-8") as f:
        for row in csv.reader(f):
            if row and row[0] != "Label":
                if len(row) > 1:
                    labels.append(row[0].strip() + " # " + row[1].strip())
                else:
                    labels.append(row[0].strip())
    return labels


label_families = {
    fname[:-4]: load_labels(os.path.join(LABELS_DIR, fname))
    for fname in os.listdir(LABELS_DIR)
    if fname.endswith(".csv")
}

# ------------------------------------------
# BUILD JSON SCHEMA
# ------------------------------------------
def build_json_schema(labels: List[str]) -> Dict:
    """Build a JSON schema with enum for labels"""
    return {
        "type": "object",
        "properties": {
            "classifications": {
                "type": "array",
                "items": {
                    "type": "object",
                    "properties": {
                        "file_index": {
                            "type": "integer",
                            "description": "The position of the file in the list. Starts at 0."
                        },
                        "label": {
                            "type": "string",
                            "enum": labels,
                            "description": "The classification label selected from the available labels"
                        }
                    },
                    "required": ["file_index", "label"],
                    "additionalProperties": False
                }
            }
        },
        "required": ["classifications"],
        "additionalProperties": False
    }


# ------------------------------------------
# CODE SNIPPET READING
# ------------------------------------------
def read_snippet(filepath, max_lines=60):
    """Read the first max_lines of a file"""
    if not filepath or not os.path.exists(filepath):
        return ""
    try:
        with open(filepath, 'r', encoding='utf-8', errors='ignore') as f:
            lines = f.readlines()
        return ''.join(lines[:max_lines])
    except:
        return ""


def detect_language(filepath):
    """Detect programming language from file extension"""
    if not filepath:
        return ""
    ext = os.path.splitext(filepath)[1].lower()
    lang_map = {
        '.php': 'php',
        '.py': 'python',
        '.js': 'javascript',
        '.ts': 'typescript',
        '.java': 'java',
        '.go': 'go',
        '.rs': 'rust',
        '.cpp': 'cpp',
        '.c': 'c',
        '.cs': 'csharp',
        '.rb': 'ruby',
    }
    return lang_map.get(ext, '')


# ------------------------------------------
# PROMPTS
# ------------------------------------------
def build_system_prompt(labels: List[str]) -> str:
    """Build system prompt with label list"""
    return f"""
We want to classify code sourcefiles.

For example, 
- component:data_access:repository for DTO or repository.
- component:messaging:subscriber for message handler.
- component:logic:domain_service for domain logic service, or business logic.
- utility:helper:component for helper or utility classes.

This will help us to train an AI model to classify code sourcefiles. Your decision should be really accurate.

All the classes or components are Open Source. you already have access to all the code and knowledge about them.
use this knowledge to make the best possible classification. snippets will be provided for each class.

For each file, select the most appropriate label.
"""

def build_user_prompt(batch: List[tuple]) -> str:
    """Build user prompt for a batch of classes with code snippets"""
    items = []
    file_index = 0
    for cname, fname in batch:
        snippet = read_snippet(fname)
        lang = detect_language(fname)
        lang_tag = f"```{lang}\n" if lang else "```\n"
        items.append(
            f"File index: {file_index}\nClass: {cname}\nFile: {fname}\nCode:\n{lang_tag}{snippet}\n```"
        )
        file_index += 1
    return "Classify the following classes:\n\n" + "\n\n---\n\n".join(items)


# ------------------------------------------
# CLASSIFIER FUNCTION (BATCH)
# ------------------------------------------
def classify_batch(batch: List[tuple], labels: List[str], system_prompt: str) -> List[Dict]:
    """
    Classify a batch of classes (up to 100) in one API call.
    Returns a list of dicts with 'class', 'file' and 'label' keys.
    """
    user_prompt = build_user_prompt(batch)

    #print("user prompt", user_prompt)
    
    try:
        response = client.chat.completions.create(
            model=MODEL,
            messages=[
                {"role": "system", "content": system_prompt},
                {"role": "user", "content": user_prompt},
            ],
            response_format={
                "type": "json_schema",
                "json_schema": {
                    "name": "classification_responsev2",
                    "strict": True,
                    "schema": build_json_schema(labels)
                }
            },
            temperature=TEMPERATURE,
        )
        
        content = response.choices[0].message.content
        print("got response", content)
        result = json.loads(content)
        
        # Create a mapping from file_index to label for easy lookup
        classifications = {item["file_index"]: item["label"] for item in result.get("classifications", [])}
        
        # Return results in the same order as input batch
        results = []
        for index, (cname, fname) in enumerate(batch):
            label = classifications.get(index, "UNKNOWN")
            results.append({"class": cname, "file": fname, "label": label})
        return results
        
    except Exception as e:
        print(f"[ERROR] Batch classification failed: {e}")
        # Return UNKNOWN for all items in case of error
        return [
            {"class": cname, "file": fname, "label": "UNKNOWN"}
            for cname, fname in batch
        ]


# ------------------------------------------
# READ INPUT CSV AND SAMPLE ROWS
# ------------------------------------------
with open(INPUT_CSV, encoding="utf-8") as f:
    all_rows = [row for row in csv.reader(f) if row]

if MAX_COUNT < len(all_rows):
    rows = random.sample(all_rows, MAX_COUNT)
else:
    rows = all_rows

print(f"[INFO] Using {len(rows)} sampled rows (requested {MAX_COUNT}).")

# ------------------------------------------
# PROCESS EACH FAMILY
# ------------------------------------------
BATCH_SIZE = 100

for family, labels in label_families.items():
    print(f"\n[INFO] Processing family: {family} ({len(labels)} labels)")
    
    system_prompt = build_system_prompt(labels)
    out_path = os.path.join(OUTPUT_DIR, f"classified_{family}.csv")
    
    with open(out_path, "w", newline="", encoding="utf-8") as out_f:
        writer = csv.writer(out_f)
        writer.writerow(["class", "file", "label"])
        
        # Process in batches of 100
        for batch_start in range(0, len(rows), BATCH_SIZE):
            batch_end = min(batch_start + BATCH_SIZE, len(rows))
            batch_rows = rows[batch_start:batch_end]
            
            # Prepare batch: (class_name, file_name)
            batch = []
            for row in batch_rows:
                cname = row[0].strip()
                fname = row[2].strip() if len(row) > 2 else ""
                batch.append((cname, fname))
            
            # Classify batch
            print(f"[{family}] Processing batch {batch_start//BATCH_SIZE + 1} ({batch_start+1}-{batch_end}/{len(rows)})...")
            results = classify_batch(batch, labels, system_prompt)
            
            # Write results
            for result in results:
                writer.writerow([result["class"], result["file"], result["label"]])
            
            print(f"[{family}] Batch {batch_start//BATCH_SIZE + 1} completed")
    
    print("[OK] Saved:", out_path)


print("\n[ALL FAMILIES DONE]")

