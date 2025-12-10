#!/usr/bin/env python3
import joblib
import sys

encoders = joblib.load('ai/training/classifier/v3/build/php/encoders.joblib')
vec = encoders['vectorizers']['class_name']

print(f"Vocabulary size: {len(vec.vocabulary_)}")
print(f"\nChecking common class names:")
test_words = ['repository', 'entity', 'challenge', 'organization', 'githubEvent', 'event']
for word in test_words:
    lower = word.lower()
    print(f"  '{word}' (as '{lower}'): {'YES' if lower in vec.vocabulary_ else 'NO'}")

print(f"\nFirst 30 vocabulary items:")
for i, (word, idx) in enumerate(sorted(vec.vocabulary_.items(), key=lambda x: x[1])[:30]):
    print(f"  {idx}: {word}")

# Test transformation
print(f"\nTest TF-IDF transformation:")
test_names = ['Repository', 'Entity', 'Challenge', 'GithubEvent']
for name in test_names:
    result = vec.transform([name.lower()])
    non_zero = result.nonzero()[1]
    print(f"  '{name}' -> {len(non_zero)} non-zero features")
    if len(non_zero) > 0:
        print(f"    Indices: {non_zero[:5]}")
