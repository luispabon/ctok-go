"""Record formatted Go output and generator hashes after offline generation."""
import hashlib
import json
from pathlib import Path

root = Path("/repo")
manifest_path = root / "testdata/unicode/manifest.json"
manifest = json.loads(manifest_path.read_text())
for name, path in {
    "unicode_tables.go": root / "unicode_tables.go",
    "tools/reference/generate.py": Path("/work/generate.py"),
    "tools/reference/finish_manifest.py": Path(__file__),
}.items():
    manifest["artifacts"][name] = hashlib.sha256(path.read_bytes()).hexdigest()
manifest_path.write_text(json.dumps(manifest, indent=2, sort_keys=True) + "\n")
