import sys
import os
import json
import joblib
import numpy as np
from sklearn.ensemble import RandomForestClassifier
from sklearn.feature_extraction.text import TfidfVectorizer

def export_model(model_path, encoders_path, features_path, output_path):
    print(f"Loading model from {model_path}...")
    model = joblib.load(model_path)
    
    print(f"Loading encoders from {encoders_path}...")
    encoders_data = joblib.load(encoders_path)
    
    print(f"Loading features from {features_path}...")
    with open(features_path, 'r') as f:
        features_data = json.load(f)

    export_data = {
        "meta": {
            "type": "RandomForestClassifier",
            "n_features": model.n_features_in_,
            "n_classes": model.n_classes_,
            "classes": model.classes_.tolist() if hasattr(model.classes_, "tolist") else model.classes_,
        },
        "features": features_data,
        "encoders": {},
        "vectorizers": {},
        "trees": []
    }

    # Export Encoders
    feature_encoders = encoders_data.get('feature_encoders', {})
    for col, classes in feature_encoders.items():
        export_data["encoders"][col] = classes.tolist() if hasattr(classes, "tolist") else classes

    # Export Vectorizers (TF-IDF)
    vectorizers = encoders_data.get('vectorizers', {})
    for col, vec in vectorizers.items():
        if isinstance(vec, TfidfVectorizer):
            export_data["vectorizers"][col] = {
                "vocabulary": vec.vocabulary_,
                "idf": vec.idf_.tolist(),
                "norm": vec.norm,
                "use_idf": vec.use_idf,
                "smooth_idf": vec.smooth_idf,
                "sublinear_tf": vec.sublinear_tf,
            }
    
    # Export Programming Language OneHotEncoder
    prog_lang_encoder = encoders_data.get('programming_language_encoder')
    if prog_lang_encoder is not None:
        # OneHotEncoder has categories_ attribute
        if hasattr(prog_lang_encoder, 'categories_'):
            export_data["programming_language_encoder"] = {
                "categories": [cat.tolist() if hasattr(cat, "tolist") else list(cat) for cat in prog_lang_encoder.categories_]
            }

    # Export Trees
    print(f"Exporting {len(model.estimators_)} trees...")
    for i, estimator in enumerate(model.estimators_):
        tree = estimator.tree_
        tree_data = {
            "children_left": tree.children_left.tolist(),
            "children_right": tree.children_right.tolist(),
            "feature": tree.feature.tolist(),
            "threshold": tree.threshold.tolist(),
            "value": [v[0].tolist() for v in tree.value], # value is shape (n_nodes, 1, n_classes)
            "node_count": tree.node_count
        }
        export_data["trees"].append(tree_data)

    print(f"Saving to {output_path}...")
    
    # Custom JSON encoder to handle numpy types
    class NumpyEncoder(json.JSONEncoder):
        def default(self, obj):
            if isinstance(obj, np.integer):
                return int(obj)
            elif isinstance(obj, np.floating):
                return float(obj)
            elif isinstance(obj, np.ndarray):
                return obj.tolist()
            return super(NumpyEncoder, self).default(obj)
    
    # Save as gzipped JSON for much smaller file size
    import gzip
    output_path_gz = output_path + '.gz'
    with gzip.open(output_path_gz, 'wt', encoding='utf-8') as f:
        json.dump(export_data, f, cls=NumpyEncoder)
    print(f"Done. Compressed model saved to {output_path_gz}")
    
    # Also save uncompressed for debugging if needed
    # with open(output_path, 'w') as f:
    #     json.dump(export_data, f, cls=NumpyEncoder)

if __name__ == "__main__":
    if len(sys.argv) < 5:
        print("Usage: python 5-export.py <model_pkl> <encoders_joblib> <features_json> <output_json>")
        sys.exit(1)
    
    export_model(sys.argv[1], sys.argv[2], sys.argv[3], sys.argv[4])
