# Model for naming things

Small CLI script that exports a subset of pre-trained GloVe word embeddings to a JSON file.

This is **not** a generative AI model (LLM). It uses a **pre-trained embedding model** (GloVe) that maps each word to a fixed-size numeric vector (here: 50 dimensions). These vectors are commonly used for similarity search, clustering, simple NLP features, etc.

## Requirements

+ Python 3.9+ recommended
+ gensim

Install dependencies:

```bash
pip install -r requirements.txt
```

## Usage

Basic usage (default: 200000 words, output vectors.json):

```bash
python export_vectors.py
```

Limit the number of exported words:

```bash
python export_vectors.py --limit 50000
```

Choose a different output file:

```bash
python export_vectors.py --limit 50000 --output vectors.w2v
```

Choose a different format (`bin` or `json.gz`):

```bash
python export_vectors.py --limit 50000 --format bin
# or
python export_vectors.py --limit 50000 --format json.gz
```

Number of decimal places in the output (default: 3):

```bash
python export_vectors.py --limit 50000 --decimals 4
```

### Building for AST-Metrics

```bash
python export_vectors.py --limit 50000 --output ../../internal/analyzer/namer/vectors.w2v --decimals 3 --format bin
```