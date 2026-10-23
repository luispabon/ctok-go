"""Record exact normalized V48 streams from authored length-prefixed inputs."""
import gzip
import hashlib
import json
from pathlib import Path
import struct
import sys
import unicodedata

import regex
from ctok.main import _model, normalize, token_count
from ctok.normalize import stream_norm

assert sys.version_info[:3] == (3, 13, 17)
assert unicodedata.unidata_version == "15.1.0"
assert regex.__version__ == "2026.8.31"
model = _model("v4.8")
rows = []
while prefix := sys.stdin.buffer.read(4):
    assert len(prefix) == 4
    length, = struct.unpack("<I", prefix)
    raw = sys.stdin.buffer.read(length)
    assert len(raw) == length
    text = raw.decode("utf-8", errors="strict")
    norm = normalize(text, "4.8")
    rows.append([text, stream_norm(norm, model), len(norm) - len(norm.rstrip("\n")), token_count(text, "4.8")])
assert len(rows) == 5000
artifact = gzip.compress(json.dumps(rows, ensure_ascii=True, separators=(",", ":")).encode(), mtime=0)
Path("/out/streams.json.gz").write_bytes(artifact)
print(len(rows), hashlib.sha256(artifact).hexdigest())
