#!/usr/bin/env python3
"""
Model Export Script

Exports a trained RandomForest model to JSON format for use in Go runtime.
The exported model is compressed with gzip for smaller file size.

Supports:
- Legacy sklearn TfidfVectorizer (vocabulary-based)
- Hashing TF-IDF (dict bundles stored in encoders.joblib)

Usage:
    python 5-export.py <model_pkl> <encoders_joblib> <features_json> <output_json>
"""

import sys
import json
import gzip
import joblib
import numpy as np
from sklearn.feature_extraction.text import TfidfVectorizer
from sklearn.feature_extraction.text import TfidfVectorizer
from sklearn.pipeline import Pipeline
from sklearn.feature_extraction.text import HashingVectorizer, TfidfTransformer

def _to_list(x):
    if x is None:
        return None
    if hasattr(x, "tolist"):
        return x.tolist()
    return list(x)


class NumpyEncoder(json.JSONEncoder):
    def default(self, obj):
        if isinstance(obj, np.integer):
            return int(obj)
        if isinstance(obj, np.floating):
            return float(obj)
        if isinstance(obj, np.ndarray):
            return obj.tolist()
        return super().default(obj)

def export_hashing_pipeline(col: str, pipe: Pipeline):
    """
    Supporte: Pipeline( ('hv', HashingVectorizer), ('tfidf', TfidfTransformer) )
    """
    if not hasattr(pipe, "named_steps"):
        raise RuntimeError(f"[export] vectorizer '{col}': pipeline has no named_steps")

    hv = pipe.named_steps.get("hv")
    tfidf = pipe.named_steps.get("tfidf")

    if not isinstance(hv, HashingVectorizer) or not isinstance(tfidf, TfidfTransformer):
        raise RuntimeError(
            f"[export] vectorizer '{col}': unsupported Pipeline steps. "
            f"Expected hv=HashingVectorizer + tfidf=TfidfTransformer, got hv={type(hv)} tfidf={type(tfidf)}"
        )

    if not hasattr(tfidf, "idf_") or tfidf.idf_ is None:
        raise RuntimeError(
            f"[export] vectorizer '{col}': TfidfTransformer is not fitted (idf_ missing). "
            f"Make sure you called pipe.fit_transform(...) during training."
        )

    n_features = int(hv.n_features)
    idf_list = _to_list(tfidf.idf_)
    if len(idf_list) != n_features:
        raise RuntimeError(
            f"[export] vectorizer '{col}': idf length mismatch (len(idf)={len(idf_list)} vs n_features={n_features})"
        )

    # stop_words peut être 'english' ou None
    stop_words = hv.stop_words
    if stop_words not in (None, "english"):
        # Pour éviter d'embarquer une liste énorme : on refuse explicitement
        raise RuntimeError(
            f"[export] vectorizer '{col}': unsupported stop_words={stop_words!r}. "
            f"Use stop_words=None or 'english'."
        )

    return {
        "type": "hashing_tfidf",
        "n_features": n_features,
        "idf": idf_list,
        "norm": tfidf.norm,
        "use_idf": bool(tfidf.use_idf),
        "smooth_idf": bool(tfidf.smooth_idf),
        "sublinear_tf": bool(tfidf.sublinear_tf),

        "alternate_sign": bool(hv.alternate_sign),
        "lowercase": bool(hv.lowercase),
        "token_pattern": hv.token_pattern,
        "ngram_range": list(hv.ngram_range),
        "stop_words": stop_words,
    }


def export_model(model_path, encoders_path, features_path, output_path):
    print(f"Loading model: {model_path}")
    model = joblib.load(model_path)

    print(f"Loading encoders: {encoders_path}")
    encoders_data = joblib.load(encoders_path)

    print(f"Loading features metadata: {features_path}")
    with open(features_path, "r") as f:
        features_data = json.load(f)

    export_data = {
        "meta": {
            "type": "RandomForestClassifier",
            "n_features": int(getattr(model, "n_features_in_", 0)),
            "n_classes": int(getattr(model, "n_classes_", 0)),
            "classes": _to_list(getattr(model, "classes_", [])),
        },
        "features": features_data,  # keep as-is
        "encoders": {},
        "vectorizers": {},
        "programming_language_encoder": None,
        "trees": [],
    }

    # --- Export Encoders (categorical label encoders stored as classes lists)
    feature_encoders = encoders_data.get("feature_encoders", {})
    for col, classes in feature_encoders.items():
        export_data["encoders"][col] = _to_list(classes)

    # --- Export Vectorizers
    # We support 2 shapes:
    # 1) Legacy: sklearn TfidfVectorizer (vocabulary-based)
    # 2) Hashing TF-IDF: dict bundle {"type":"hashing_tfidf", "n_features":..., "idf":..., ...}
    # --- Export Vectorizers
    vectorizers = encoders_data.get("vectorizers", {})
    for col, vec in vectorizers.items():
        # NEW: sklearn Pipeline(HashingVectorizer -> TfidfTransformer)
        if isinstance(vec, Pipeline):
            export_data["vectorizers"][col] = export_hashing_pipeline(col, vec)
            continue

        # Hashing TF-IDF bundle (dict) (still supported)
        if isinstance(vec, dict) and vec.get("type") == "hashing_tfidf":
            n_features = int(vec.get("n_features", 0))
            idf = vec.get("idf", None)
            if n_features <= 0:
                raise RuntimeError(f"[export] vectorizer '{col}': hashing_tfidf missing/invalid n_features")
            if idf is None:
                raise RuntimeError(f"[export] vectorizer '{col}': hashing_tfidf missing idf")
            idf_list = _to_list(idf)
            if len(idf_list) != n_features:
                raise RuntimeError(
                    f"[export] vectorizer '{col}': idf length mismatch "
                    f"(len(idf)={len(idf_list)} vs n_features={n_features})"
                )

            export_data["vectorizers"][col] = {
                "type": "hashing_tfidf",
                "n_features": n_features,
                "idf": idf_list,
                "norm": vec.get("norm", "l2"),
                "use_idf": bool(vec.get("use_idf", True)),
                "smooth_idf": bool(vec.get("smooth_idf", True)),
                "sublinear_tf": bool(vec.get("sublinear_tf", True)),
                "alternate_sign": bool(vec.get("alternate_sign", False)),
                "lowercase": bool(vec.get("lowercase", True)),
                "token_pattern": vec.get("token_pattern", r"(?u)\b\w\w+\b"),
                "ngram_range": vec.get("ngram_range", [1, 3]),
                "stop_words": vec.get("stop_words", "english"),
            }
            continue

        # Legacy sklearn TfidfVectorizer (vocabulary-based)
        if isinstance(vec, TfidfVectorizer):
            if not hasattr(vec, "vocabulary_") or vec.vocabulary_ is None:
                raise RuntimeError(
                    f"[export] vectorizer '{col}': TfidfVectorizer has no vocabulary_. "
                    f"You cannot remove vocabulary at export time. "
                    f"Switch training to hashing pipelines."
                )

            export_data["vectorizers"][col] = {
                "type": "tfidf_vocab",
                "vocabulary": vec.vocabulary_,
                "idf": _to_list(vec.idf_),
                "norm": vec.norm,
                "use_idf": bool(vec.use_idf),
                "smooth_idf": bool(vec.smooth_idf),
                "sublinear_tf": bool(vec.sublinear_tf),
            }
            continue

        raise RuntimeError(
            f"[export] vectorizer '{col}': unsupported type={type(vec)}. "
            f"Expected Pipeline(HashingVectorizer->TfidfTransformer), dict(type='hashing_tfidf', ...) or sklearn TfidfVectorizer."
        )


    # --- Export Programming Language OneHotEncoder (sklearn OneHotEncoder)
    prog_lang_encoder = encoders_data.get("programming_language_encoder")
    if prog_lang_encoder is not None:
        if hasattr(prog_lang_encoder, "categories_"):
            export_data["programming_language_encoder"] = {
                "categories": [
                    _to_list(cat) for cat in prog_lang_encoder.categories_
                ]
            }

    # --- Export Trees (optimized: skip value for internal nodes, round leaf values)
    print(f"Exporting {len(model.estimators_)} trees (optimized)...")
    total_nodes = 0
    total_leaves = 0
    for estimator in model.estimators_:
        tree = estimator.tree_
        children_left = tree.children_left
        children_right = tree.children_right
        n_nodes = int(tree.node_count)
        total_nodes += n_nodes

        # Optimize value arrays: only store for leaf nodes, round to 4 decimals
        optimized_values = []
        for i in range(n_nodes):
            if children_left[i] == -1 and children_right[i] == -1:
                # Leaf node: store rounded values, zero out tiny probabilities
                vals = tree.value[i][0]
                rounded = [round(float(v), 4) if abs(float(v)) > 0.0005 else 0.0 for v in vals]
                optimized_values.append(rounded)
                total_leaves += 1
            else:
                # Internal node: store empty array (Go never reads value for internal nodes)
                optimized_values.append([])

        # Round thresholds to 6 decimal places
        thresholds = [round(float(t), 6) for t in tree.threshold]

        export_data["trees"].append({
            "children_left": _to_list(children_left),
            "children_right": _to_list(children_right),
            "feature": _to_list(tree.feature),
            "threshold": thresholds,
            "value": optimized_values,
            "node_count": n_nodes,
        })
    print(f"[INFO] Exported {total_nodes} nodes ({total_leaves} leaves) across {len(model.estimators_)} trees")

    # --- Save gzipped JSON
    out_gz = output_path + ".gz"
    print(f"Saving gzipped JSON model: {out_gz}")
    with gzip.open(out_gz, "wt", encoding="utf-8") as f:
        json.dump(export_data, f, cls=NumpyEncoder)

    print("[OK] Done.")


if __name__ == "__main__":
    if len(sys.argv) < 5:
        print("Usage: python 5-export.py <model_pkl> <encoders_joblib> <features_json> <output_json>")
        sys.exit(1)

    export_model(sys.argv[1], sys.argv[2], sys.argv[3], sys.argv[4])
