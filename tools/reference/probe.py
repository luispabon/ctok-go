"""Verify pinned oracle assumptions before generating any tables."""
import hashlib
import json
import platform
import sys
import unicodedata

import regex
from ctok.main import _document, token_count

assert sys.version_info[:3] == (3, 13, 17), sys.version
assert unicodedata.unidata_version == "15.1.0"
assert regex.__version__ == "2026.8.31", regex.__version__
payload = ("STEINER CONTEXT PROBE REFERENCE DATA 0123456789 abcdefghijklmnopqrstuvwxyz. " * 200)[:10000]
assert hashlib.sha256(payload.encode()).hexdigest() == "31d2500f97354cdfad22b5e6acb757c0569965048f7a17fb427cbaf28817c53f"
doc = _document("pieces_v4_7.json")
groups = {key: len(value) for key, value in doc["tokens"].items()}
assert sum(groups.values()) == 15283, groups
assert groups["bytes_fallback"] == 468, groups
assert groups["markers"] == 3, groups
print(json.dumps({"python": platform.python_version(), "unicode": unicodedata.unidata_version,
                  "regex": regex.__version__, "architecture": platform.machine(),
                  "groups": groups, "metadata": doc["meta"]}, sort_keys=True), flush=True)
delta = token_count("Reply only OK." + payload, "4.8") - token_count("Reply only OK.", "4.8")
print("issue_delta", delta, flush=True)
assert delta == 6446, delta
