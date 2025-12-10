import json
import sys

# Load features.json
with open('ai/training/classifier/v3/build/php/features.json', 'r') as f:
    data = json.load(f)

print("=== Features Metadata ===")
print(f"Categorical columns: {data.get('categorical_cols', [])}")
print(f"NLP columns: {data.get('nlp_cols', [])}")
print(f"Total final features: {len(data.get('final_feature_names', []))}")
print(f"Class name weight: {data.get('class_name_weight', 'N/A')}")
print(f"Programming language weight: {data.get('programming_language_weight', 'N/A')}")

print("\n=== First 50 feature names ===")
for i, name in enumerate(data.get('final_feature_names', [])[:50]):
    print(f"{i}: {name}")

print("\n=== Encoders ===")
with open('ai/training/classifier/v3/build/php/encoders.joblib', 'rb') as f:
    import joblib
    encoders = joblib.load(f)
    print(f"Feature encoders keys: {list(encoders.get('feature_encoders', {}).keys())}")
    print(f"Vectorizers keys: {list(encoders.get('vectorizers', {}).keys())}")
    if 'programming_language_encoder' in encoders:
        print(f"Programming language encoder exists: {encoders['programming_language_encoder'] is not None}")
