"""Adapt MIT test_api.py input vectors to the explicit V48 family."""
import ast
import hashlib
import json
from pathlib import Path
import sys
import unicodedata

import regex
from ctok.main import _model, normalize, token_count
from ctok.normalize import stream_norm

assert sys.version_info[:3] == (3, 13, 17)
assert unicodedata.unidata_version == "15.1.0"
assert regex.__version__ == "2026.8.31"
source = Path("/reference_test_api.py").read_bytes()
source_hash = hashlib.sha256(source).hexdigest()
wanted = {"MARK_ROWS", "APOSTROPHE_ROWS", "CONTRACTION_ROWS", "PIECE_ROWS", "SYMBOL_ROWS"}
texts = []
groups = {}
for node in ast.parse(source).body:
    if isinstance(node, ast.Assign) and isinstance(node.targets[0], ast.Name):
        name = node.targets[0].id
        if name in wanted:
            rows = ast.literal_eval(node.value)
            groups[name] = len(rows)
            texts.extend(row[0] if name == "MARK_ROWS" else row[1] for row in rows)
assert set(groups) == wanted
assert groups["MARK_ROWS"] == 48
texts.extend(["\u0897", "h\u0897b", "x \u0363The x", "x \u0363HELLO x",
              "ทํางาน", "ทำงาน", "นํ้า", "น้ํา", "ﬁ", "１", "Ⅻ", "①", "²",
              "hello, world", "123", "日本", " a", "a", "hello", "hello" + "\n" * 29,
              "\uf8ff", "\x01", "\ue000\x01\x85", "a\uf0b7b", "a\ue000\uf8ffb",
              "\U000f0000", "\U000f0000\U000f0001", "a\U00090095a", "aa"])
model = _model("v4.8")
output = []
for text in texts:
    norm = normalize(text, "4.8")
    output.append([text, stream_norm(norm, model), len(norm) - len(norm.rstrip("\n")), token_count(text, "4.8")])
artifact = {"source_sha256": source_hash, "groups": groups, "rows": output}
path = Path("/out/api_edges.json")
path.write_text(json.dumps(artifact, ensure_ascii=True, indent=2, sort_keys=True) + "\n")
print(len(output), source_hash, hashlib.sha256(path.read_bytes()).hexdigest())
