"""Generate frozen Unicode behavior and normalization fixtures with the pinned oracle."""
import gzip
import hashlib
import json
import random
import struct
from pathlib import Path
import sys
import unicodedata as ud

import regex
from ctok.normalize import classify, mark_case, nfc, _is_upper, _is_lower, _lower, _ALPHABETIC, _MARK

VERSION = 1
ROOT = Path("/out")


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def main():
    assert sys.version_info[:3] == (3, 13, 17)
    assert ud.unidata_version == "15.1.0"
    assert regex.__version__ == "2026.8.31"
    classes = {"wordy": 0, "hard": 1, "digit": 2, "punct": 3, "space": 4}
    ranges = []
    mappings = []
    nfcs = []
    normalize_changes = []
    previous = None
    for cp in range(0x110000):
        if 0xD800 <= cp <= 0xDFFF:
            continue
        char = chr(cp)
        cat = ud.category(char)
        flags = (int(bool(_ALPHABETIC.fullmatch(char)))
                 | (int(bool(_MARK.fullmatch(char))) << 1)
                 | (int(_is_upper(char)) << 2)
                 | (int(_is_lower(char)) << 3)
                 | (int(cat[0] in "LM" or char in "\u1c89\u1c8a\ua7cb\u0264") << 4))
        key = (classes[classify(char)], flags, ud.combining(char), cat)
        if previous == key and ranges[-1][1] + 1 == cp:
            ranges[-1][1] = cp
        else:
            ranges.append([cp, cp, *key])
        previous = key
        lower = _lower(char)
        if lower != char:
            mappings.append([cp, lower])
        normalized = ud.normalize("NFC", char)
        if normalized != char:
            nfcs.append([cp, normalized])
        normalized = nfc(char, fold_quotes=False)
        if normalized != char:
            normalize_changes.append([cp, normalized])
    tables = {"version": VERSION, "ranges": ranges, "lower": mappings}
    target = ROOT / "testdata/unicode"
    target.mkdir(parents=True, exist_ok=True)
    (target / "properties.json").write_text(json.dumps(tables, ensure_ascii=True, separators=(",", ":")) + "\n")
    sequences = ["", "A\u030a", "\u1100\u1161\u11a8", "\u0301\u0323",
                 "a" + "\u0301" * 31, "a" + "\u0301" * 100,
                 "a\u0315\u0300", "\u0344", "\u1c89", "\u0897\u0300"]
    # Exercise all canonical two-part decompositions, both composition and ordering.
    for cp in range(0x110000):
        if 0xD800 <= cp <= 0xDFFF:
            continue
        decomposition = ud.decomposition(chr(cp))
        if decomposition and not decomposition.startswith("<"):
            sequences.append("".join(chr(int(n, 16)) for n in decomposition.split()))
    # Long runs exercise Python NFC without the stream-safe CGJ limit.
    for size in (10000, 500000):
        sequences.append("a" + "\u0301" * size)
        sequences.append("a" + "\u0315\u0300" * (size // 2))
    # Sweep all scalar+mark interactions to distinguish future Unicode tables.
    # Only store changed interactions; unchanged expectations are tested separately.
    interactions = []
    for cp in range(0x110000):
        if 0xD800 <= cp <= 0xDFFF:
            continue
        char = chr(cp)
        for text in ("a" + char + "\u0300", "a\u0315" + char):
            normalized = ud.normalize("NFC", text)
            if normalized != text:
                interactions.append([cp, text, normalized])
    fixture = {"scalar_changes": nfcs, "sequences": [[s, ud.normalize("NFC", s)] for s in sequences],
               "interactions": interactions}
    cases = ["", "Hello", "HELLO", "Σ", "ΟΣ", "Ος", "İ", "ẞ", "Abİ", "ϴ", "Θϴ",
             "\u1c89abc", "\ua7cbabc", "WiFi", "A\u0301", "\u0301Ab", "Aヲ", "AΣ", "Σa"]
    cases.extend(chr(cp) + "abc" for cp in range(0x110000)
                 if not 0xD800 <= cp <= 0xDFFF and _is_upper(chr(cp)))
    (target / "case.json").write_text(json.dumps([[s, mark_case(s, allcaps_min=None)] for s in cases],
                                               ensure_ascii=True, separators=(",", ":")) + "\n")
    samples = ["", "A\u030a", "\x00\x01\t\n\r", "\u0e4d\u0e32", "\u0ecd\u0eb2",
               "a\ue000b", "\ufdd0\ufdd4", "\u3000\u00a0", "‘“x”’"]
    (target / "normalize.json").write_text(json.dumps(
        {"scalar_changes": normalize_changes, "sequences": [[s, nfc(s, fold_quotes=False)] for s in samples]},
        ensure_ascii=True, separators=(",", ":")) + "\n")
    conformance_source = Path("/ucd/NormalizationTest.txt")
    assert digest(conformance_source) == "871238e37e3be0696ec2bd0891119a041b052da1a84485eda05a5438724b223e"
    conformance = []
    for line in conformance_source.read_text().splitlines():
        line = line.split("#", 1)[0].strip()
        if not line or line.startswith("@"):
            continue
        columns = ["".join(chr(int(cp, 16)) for cp in field.split()) for field in line.split(";")[:5]]
        assert len(columns) == 5
        for i, text in enumerate(columns):
            want = columns[1] if i < 3 else columns[3]
            assert ud.normalize("NFC", text) == want
            conformance.append([text, want])
    assert len(conformance) > 90000
    (target / "conformance.json.gz").write_bytes(gzip.compress(
        json.dumps(conformance, ensure_ascii=True, separators=(",", ":")).encode() + b"\n", mtime=0))
    (target / "nfc.json").write_text(json.dumps(fixture, ensure_ascii=True, separators=(",", ":")) + "\n")
    go = ["// Code generated by tools/reference/generate.py; DO NOT EDIT.", "",
          "package ctok", "", "var unicodeRanges = [...]unicodeRange{"]
    for lo, hi, cls, flags, combining, category in ranges:
        go.append(f'\t{{0x{lo:X}, 0x{hi:X}, {cls}, {flags}, {combining}, "{category}"}},')
    go.extend(["}", "", "var lowerMappings = map[rune]string{"])
    for cp, lower in mappings:
        go.append(f'\t0x{cp:X}: {json.dumps(lower, ensure_ascii=False)},')
    go.extend(["}", ""])
    (ROOT / "unicode_tables.go").write_text("\n".join(go))
    manifest = {"generator_version": VERSION, "python": sys.version.split()[0],
                "unicodedata": ud.unidata_version, "regex": regex.__version__,
                "artifacts": {str(p.relative_to(ROOT)): digest(p) for p in
                              [target / "properties.json", target / "nfc.json", target / "case.json",
                               target / "normalize.json", target / "conformance.json.gz"]}}
    for name in ("data/v48.bin", "testdata/vocabulary.json"):
        manifest["artifacts"][name] = digest(ROOT / name)
    (target / "manifest.json").write_text(json.dumps(manifest, indent=2, sort_keys=True) + "\n")


def vocabulary():
    from ctok.constants import MARKER_GLYPHS
    from ctok.engine import char_cost, frame_tail, min_vocab_tile
    from ctok.main import _document, _model
    from ctok.normalize import stream_norm

    assert sys.version_info[:3] == (3, 13, 17)
    assert ud.unidata_version == "15.1.0"
    assert regex.__version__ == "2026.8.31"
    doc = _document("pieces_v4_7.json")
    groups = {key: len(value) for key, value in doc["tokens"].items()}
    assert sum(groups.values()) == 15283
    assert groups["bytes_fallback"] == 468 and groups["markers"] == 3
    model = _model("v4.8")
    pieces = sorted(model.vocab)
    prefixes = sorted(bytes.fromhex(p) for p in doc["tokens"]["bytes_fallback"])
    assert all(pieces) and all(prefixes)
    payload = struct.pack("<II", len(pieces), len(prefixes))
    for value in [p.encode() for p in pieces] + prefixes:
        payload += struct.pack("<I", len(value)) + value
    blob = b"CTOK48\x00\x01" + hashlib.sha256(payload).digest() + payload
    (ROOT / "data/v48.bin").write_bytes(blob)
    rng = random.Random(931)
    scalars = [0, 0x7F, 0x80, 0x7FF, 0x800, 0xFFFF, 0x10000, 0x10FFFF,
               *range(0xFDD0, 0xFDD5), *range(0xE000, 0xE005)]
    scalars += [rng.randrange(0x110000) for _ in range(2048)]
    floors = [[chr(cp), char_cost(model, chr(cp))] for cp in scalars
              if not 0xD800 <= cp <= 0xDFFF]
    samples = ["", "Hello world", "x't x's x'll x're x've", "a\n\n", "\ufdd0\ufdd1\ufdd2\ufdd3\ufdd4",
               "𐀀🙂🫨", "\u0301a\u0315\u0300", "a" * 4096, "\n" * 4096]
    alphabet = "abcXYZ012_' \t\nαΣİẞ中ไทย🙂\ufdd0\ufdd4"
    samples += ["".join(rng.choice(alphabet) for _ in range(rng.randrange(100))) for _ in range(1000)]
    marked = []
    for text in samples:
        stream = stream_norm(nfc(text, fold_quotes=False), model)
        cost, _ = min_vocab_tile(stream, model.trie,
                                lambda at: 1 if stream[at] in MARKER_GLYPHS else char_cost(model, stream[at]))
        marked.append([stream, cost])
    fixture = {"schema": 1, "groups": groups, "expanded_count": len(pieces),
               "blob_sha256": hashlib.sha256(blob).hexdigest(),
               "payload_sha256": hashlib.sha256(payload).hexdigest(),
               "pieces_sha256": hashlib.sha256(json.dumps(pieces, ensure_ascii=True,
                     separators=(",", ":")).encode()).hexdigest(),
               "unit_count": len(model.unit_pieces), "floors": floors, "marked": marked,
               "tails": [[n, len(frame_tail(n, model))] for n in [*range(257), 1000, 10000]],
               "contractions": [cn + "\ufdd1" for cn in doc["tokens"]["contractions"]]}
    (ROOT / "testdata/vocabulary.json").write_text(json.dumps(fixture, ensure_ascii=True,
                                                               separators=(",", ":")) + "\n")
    manifest_path = ROOT / "testdata/unicode/manifest.json"
    manifest = json.loads(manifest_path.read_text())
    for path in (ROOT / "data/v48.bin", ROOT / "testdata/vocabulary.json"):
        manifest["artifacts"][str(path.relative_to(ROOT))] = digest(path)
    manifest["artifacts"]["tools/reference/generate.py"] = digest(Path(__file__))
    manifest_path.write_text(json.dumps(manifest, indent=2, sort_keys=True) + "\n")
    print(json.dumps({k: fixture[k] for k in ("expanded_count", "unit_count", "blob_sha256",
                                              "payload_sha256", "pieces_sha256")}))


if __name__ == "__main__":
    if sys.argv[1:] == ["vocabulary"]:
        vocabulary()
    else:
        assert not sys.argv[1:]
        main()
