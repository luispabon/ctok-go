"""Offline public-API oracle for the authored, length-prefixed Go corpus."""
import collections
import concurrent.futures
import gzip
import hashlib
import io
import json
from pathlib import Path
import struct
import sys
import unicodedata

import regex
from ctok.main import tokenize
from ctok.constants import PAD

assert sys.version_info[:3] == (3, 13, 17)
assert unicodedata.unidata_version == "15.1.0"
assert regex.__version__ == "2026.8.31"
COUNT = 250000
CORPUS_HASH = "7abf5b4b6bb0d0825c77694b3dbf761ec61eeb48e0de33070edf6d3c53c13467"
CATEGORIES = ["code", "prose", "scripts", "marks", "markers", "seams", "digits", "scalars", "whitespace", "long"]


def evaluate(raw):
    tokens = tokenize(raw.decode("utf-8", errors="strict"), "4.8")
    assert tokens[:6] == [PAD] * 6
    content = len(tokens[6:])
    assert len(tokens) - content == 6
    return len(tokens), content


def main():
    source = sys.stdin.buffer
    digest = hashlib.sha256()
    inputs = []
    categories = collections.Counter()
    total = 0
    for index in range(COUNT):
        prefix = source.read(4)
        assert len(prefix) == 4, (index, "missing length")
        length, = struct.unpack("<I", prefix)
        assert length <= 100 * 1024, (index, length)
        raw = source.read(length)
        assert len(raw) == length, (index, "truncated case")
        digest.update(prefix)
        digest.update(raw)
        total += length
        inputs.append(raw)
        categories["fixed" if index < 20 else CATEGORIES[index % 10]] += 1
    assert source.read(1) == b"", "trailing corpus bytes"
    assert digest.hexdigest() == CORPUS_HASH, digest.hexdigest()
    expected = io.BytesIO()
    with concurrent.futures.ProcessPoolExecutor(max_workers=8) as pool:
        for framed, content in pool.map(evaluate, inputs, chunksize=128):
            expected.write(struct.pack("<II", framed, content))
    output = Path("/out")
    artifact = gzip.compress(expected.getvalue(), compresslevel=9, mtime=0)
    (output / "counts.bin.gz").write_bytes(artifact)
    script_hash = hashlib.sha256(Path(__file__).read_bytes()).hexdigest()
    manifest = {"schema": 1, "generator_version": 1, "seed": 931, "count": COUNT,
                "corpus_sha256": CORPUS_HASH, "input_bytes": total,
                "categories": dict(sorted(categories.items())),
                "counts_sha256": hashlib.sha256(artifact).hexdigest(),
                "oracle_script_sha256": script_hash,
                "upstream_commit": "434612e23c9d9d42d1c5eba23c4761d653e0279d",
                "python": sys.version.split()[0], "unicodedata": unicodedata.unidata_version,
                "regex": regex.__version__}
    (output / "manifest.json").write_text(json.dumps(manifest, indent=2, sort_keys=True) + "\n")
    print(json.dumps(manifest, sort_keys=True), flush=True)


if __name__ == "__main__":
    main()
