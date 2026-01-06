#!/usr/bin/env python3
import argparse
import gzip
import io
import json
import struct
from typing import Iterable, Tuple

import gensim.downloader as api


MODEL_NAME = "glove-wiki-gigaword-50"
DIM = 50


def iter_word_vectors(model, limit: int) -> Iterable[Tuple[str, list[float]]]:
    """
    Yield (word, vector) pairs for the first `limit` words in the model vocabulary.
    """
    for i, word in enumerate(model.index_to_key):
        if i >= limit:
            break
        # model[word] is typically a numpy array
        vec = [float(x) for x in model[word]]
        if len(vec) != DIM:
            raise ValueError(f"Unexpected vector dimension for '{word}': {len(vec)} (expected {DIM})")
        yield word, vec


def export_json_gz(output_file: str, model, limit: int, decimals: int) -> int:
    """
    Stream a JSON object to a gzip file without keeping everything in memory.
    Returns number of exported words.
    """
    count = 0
    with gzip.open(output_file, "wt", encoding="utf-8") as f:
        f.write("{")
        first = True
        for word, vec in iter_word_vectors(model, limit):
            if decimals is not None:
                vec = [round(x, decimals) for x in vec]

            if not first:
                f.write(",")
            first = False

            # Proper JSON escaping for keys and array values
            f.write(json.dumps(word, ensure_ascii=False))
            f.write(":")
            f.write(json.dumps(vec, ensure_ascii=False))
            count += 1
        f.write("}")
    return count


def export_bin(output_file: str, model, limit: int) -> int:
    """
    Export a simple binary format that's easy to read from Go.

    Format (little-endian):
    - 4 bytes: magic "W2V1"
    - u32: dim (50)
    - u32: count (number of entries)
    Then repeated count times:
      - u16: word length in bytes (UTF-8)
      - bytes: word (UTF-8)
      - dim * f32: vector values
    """
    # First pass: collect words to know count (still cheap vs storing all vectors)
    words = []
    for word, _ in iter_word_vectors(model, limit):
        words.append(word)

    count = len(words)

    with open(output_file, "wb") as f:
        f.write(b"W2V1")
        f.write(struct.pack("<II", DIM, count))

        for word in words:
            b = word.encode("utf-8")
            if len(b) > 65535:
                raise ValueError(f"Word too long for u16 length: {word[:80]}... ({len(b)} bytes)")
            f.write(struct.pack("<H", len(b)))
            f.write(b)

            vec = [float(x) for x in model[word]]
            f.write(struct.pack("<" + "f" * DIM, *vec))

    return count


def main():
    parser = argparse.ArgumentParser(description="Export a subset of GloVe word vectors.")
    parser.add_argument("--limit", type=int, default=200_000, help="Number of words to export.")
    parser.add_argument(
        "--format",
        choices=["json.gz", "bin"],
        default="json.gz",
        help="Output format. 'json.gz' is portable; 'bin' is compact and Go-friendly.",
    )
    parser.add_argument("--output", type=str, default=None, help="Output file path.")
    parser.add_argument(
        "--decimals",
        type=int,
        default=3,
        help="Rounding decimals for json.gz (ignored for bin). Use e.g. 3. Set to -1 to disable rounding.",
    )
    args = parser.parse_args()

    decimals = None if args.decimals == -1 else args.decimals

    if args.output is None:
        args.output = "vectors.json.gz" if args.format == "json.gz" else "vectors.w2v"

    print(f"Loading model: {MODEL_NAME} ...")
    model = api.load(MODEL_NAME)

    if args.format == "json.gz":
        print(f"Exporting first {args.limit} words to {args.output} (gzip JSON)...")
        count = export_json_gz(args.output, model, args.limit, decimals)
    else:
        print(f"Exporting first {args.limit} words to {args.output} (binary W2V1)...")
        count = export_bin(args.output, model, args.limit)

    print(f"Done. Exported {count} words.")


if __name__ == "__main__":
    main()
