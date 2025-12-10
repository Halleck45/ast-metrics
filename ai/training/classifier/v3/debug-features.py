#!/usr/bin/env python3
"""
Debug script to show the exact feature vector construction for a specific class.
This helps identify differences between Python and Go implementations.
"""

import sys
import json
import joblib
import pandas as pd
import numpy as np
from sklearn.feature_extraction.text import TfidfVectorizer

def debug_features(csv_path, model_path, features_path, encoders_path, class_name):
    """Show detailed feature construction for a specific class."""
    
    # Load model and features
    print(f"Loading model from {model_path}...")
    model = joblib.load(model_path)
    
    print(f"Loading features from {features_path}...")
    with open(features_path, 'r') as f:
        features_data = json.load(f)
    
    print(f"Loading encoders from {encoders_path}...")
    encoders_data = joblib.load(encoders_path)
    
    # Load CSV
    print(f"Loading CSV from {csv_path}...")
    df = pd.read_csv(csv_path)
    
    # Find the class
    if 'stmt_name' in df.columns:
        df.rename(columns={'stmt_name': 'class'}, inplace=True)
    
    matching = df[df['class'].str.contains(class_name, case=False, na=False)]
    if len(matching) == 0:
        print(f"No class found matching '{class_name}'")
        return
    
    row = matching.iloc[0]
    print(f"\n=== Found class: {row['class']} ===")
    print(f"File: {row.get('file_path', 'N/A')}")
    
    # Extract class_name from namespace_raw
    namespace = row.get('namespace_raw', '')
    if namespace:
        if '\\' in namespace:
            extracted_class = namespace.split('\\')[-1]
        elif '.' in namespace:
            extracted_class = namespace.split('.')[-1]
        else:
            extracted_class = namespace
    else:
        extracted_class = ''
    
    print(f"Extracted class_name: '{extracted_class}'")
    print(f"Programming language: {row.get('programming_language', 'N/A')}")
    
    # Get feature info
    final_feature_names = features_data.get('final_feature_names', [])
    categorical_cols = features_data.get('categorical_cols', [])
    nlp_cols = features_data.get('nlp_cols', [])
    class_name_weight = features_data.get('class_name_weight', 3.0)
    prog_lang_weight = features_data.get('programming_language_weight', 2.5)
    
    print(f"\n=== Feature Configuration ===")
    print(f"Total features: {len(final_feature_names)}")
    print(f"Categorical columns: {categorical_cols}")
    print(f"NLP columns: {nlp_cols}")
    print(f"Class name weight: {class_name_weight}")
    print(f"Programming language weight: {prog_lang_weight}")
    
    # Show first 50 feature names
    print(f"\n=== First 50 Feature Names ===")
    for i, name in enumerate(final_feature_names[:50]):
        print(f"{i}: {name}")
    
    # Show numeric values
    print(f"\n=== Numeric Features (first 25) ===")
    numeric_cols = [c for c in df.columns if c not in categorical_cols + nlp_cols + ['class', 'file_path', 'stmt_type']]
    for col in numeric_cols[:25]:
        if col in row:
            print(f"{col}: {row[col]}")
    
    # Show categorical encoding
    print(f"\n=== Categorical Encoding ===")
    feature_encoders = encoders_data.get('feature_encoders', {})
    for col in categorical_cols:
        if col in feature_encoders and col in row:
            classes = feature_encoders[col]
            value = row[col]
            if value in classes:
                encoded = list(classes).index(value)
                print(f"{col}: '{value}' -> {encoded}")
            else:
                print(f"{col}: '{value}' -> NOT IN ENCODER")
    
    # Show programming language one-hot
    print(f"\n=== Programming Language One-Hot ===")
    prog_lang_encoder = encoders_data.get('programming_language_encoder')
    if prog_lang_encoder and hasattr(prog_lang_encoder, 'categories_'):
        categories = prog_lang_encoder.categories_[0]
        prog_lang = row.get('programming_language', '')
        print(f"Categories: {list(categories)}")
        print(f"Value: '{prog_lang}'")
        if prog_lang in categories:
            idx = list(categories).index(prog_lang)
            print(f"One-hot index: {idx} (weight: {prog_lang_weight})")
    
    # Show NLP features
    print(f"\n=== NLP Features ===")
    vectorizers = encoders_data.get('vectorizers', {})
    
    # Class name
    if 'class_name' in vectorizers:
        vec = vectorizers['class_name']
        print(f"class_name: '{extracted_class}'")
        print(f"  Vocabulary size: {len(vec.vocabulary_)}")
        print(f"  Weight: {class_name_weight}")
        if extracted_class:
            tokens = extracted_class.lower().split()
            print(f"  Tokens: {tokens}")
            for token in tokens:
                if token in vec.vocabulary_:
                    print(f"    '{token}' -> index {vec.vocabulary_[token]}")
    
    # Other NLP
    for col in nlp_cols:
        if col in vectorizers and col in row:
            vec = vectorizers[col]
            value = str(row[col]) if pd.notna(row[col]) else ''
            print(f"{col}: '{value[:50]}...'")
            print(f"  Vocabulary size: {len(vec.vocabulary_)}")

if __name__ == "__main__":
    if len(sys.argv) < 6:
        print("Usage: python3 debug-features.py <csv> <model.pkl> <features.json> <encoders.joblib> <class_name_pattern>")
        print("Example: python3 debug-features.py dataset/pulse.csv build/php/model.pkl build/php/features.json build/php/encoders.joblib GithubEvent")
        sys.exit(1)
    
    debug_features(sys.argv[1], sys.argv[2], sys.argv[3], sys.argv[4], sys.argv[5])
