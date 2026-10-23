# Vendored NFC implementation

Derived from `golang.org/x/text/unicode/norm` v0.42.0, source commit
`fafe4a06967e06550e69ee42787d9902845d2a3f`, under the adjacent BSD license.
Only NFC is implemented, through the String method of the single-form Form type.
NFKC data remains because the generator shares its encoding. Runtime dependencies
are standard library packages. Generated tables are frozen rather than compiler-selected.

The composition and Hangul algorithms and packed property decoder are retained
from x/text. Segment storage changes from fixed UTF-8 byte/position arrays to
growable compact rune entries. Stable counting sort replaces per-rune insertion
sort, with CCC-zero entries as fixed barriers. This preserves canonical order
without quadratic behavior on long segments. An allocation-free measurement pass
sizes segment entries, unordered-run scratch and output before normalization.
The main entries use 8 bytes each; stable-sort scratch packs each scalar, CCC and
backward flag into 4 bytes. A pre-sized strings.Builder avoids output growth and
the final string copy. Storage is local to the call, with doubling only as a
correctness fallback if a measured bound is too small. No stream-safe CGJ is inserted. Input validity is handled by
the owning ctok package before calling String.

The retained implementation intentionally does not expose streaming readers,
transforms, boundary APIs or compatibility normalization. Data shared with NFKC
remains in the generated tables to avoid rewriting the generator's encoding.

Upstream source SHA-256:

- composition.go: `326e91fea709726d884e475db9b7a61ad17829e32d795f182cdcfcfdf64d9bf6`
- normalize.go: `f1363700e71a35f9c966492f10dc6f3405cd824f2cc03b08ee64b24503d56f39`
- forminfo.go: `b53ecff9ab5a83fd45598f54dfc9ea03849a3bcf4617ef02cf43225c71acc1e3`
- input.go: `965b431790bb139543d71a9c497920ef7d9a15af417456a2bbd0cdb629330e8d`
- trie.go: `d87793d558251ee8824954f0b7bc5564803e4c9d59a8c4eebb4c3c5cfbd19492`

## Reproduce tables

`tools/reference/ucd.json` pins the official Unicode 15.1 UCD archive using a
pre-existing Fedora SHA-512 pin, plus SHA-256 values for extracted inputs.
Clone x/text at the recorded commit into a scratch directory. In its
`internal/gen/gen.go`, add `{"15.1.0", "go1.1"}` to the version/tag registry.
Extract the four pinned files into `DATA/15.1.0/ucd/` under that scratch clone.
Run from its `unicode/norm/` directory, offline:

```sh
GOTOOLCHAIN=local GOPROXY=off UNICODE_VERSION=15.1.0 \
  go run maketables.go triegen.go -unicode=15.1.0 \
  -tables=all -url=http://127.0.0.1:1
```

The executor also runs this command in a network-disabled Docker container with
explicit scratch/output mounts and a read-only Go toolchain/module cache.
Any `FAILED` or `FAIL F:` generator diagnostic is a failure even at exit 0.
Compared to the shipped 15.0 table, the raw output changes only Version and its
build tag. Run `tools/reference/finish_norm.py` on the generated table to remove
the build tag and streaming-only MaxTransformChunkSize constant. The resulting
SHA-256 must equal `generated_table_sha256` in `ucd.json`. The runtime has no
x/text module dependency. Unicode data is covered by the root UNICODE-LICENSE.

The retained, patched NFC source file hashes are recorded in
`tools/reference/norm_patch.json`; these identify the documented vendor adaptation
separately from unchanged generated data.
