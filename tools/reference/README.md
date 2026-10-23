# Development oracle

This tooling is development-only. Consumers and ordinary Go tests do not use
Docker, Python, credentials or network access. Oracle inputs and dependency hashes
are recorded in `REFERENCE.json`.

## Image

In an empty scratch directory outside this repository, download the recorded
ctok tag archive as `tag.tar.gz` and the recorded regex wheel under its original
filename. Verify both SHA-256 values from `REFERENCE.json` before use. Confirm
that upstream tag v1.3.0 resolves to the recorded commit. Do not execute upstream
code on the host. Do not copy upstream corpora into this repository.

Build using that scratch directory as the context:

```sh
docker build --network none -t ctok-go-oracle:434612e \
  -f /absolute/path/to/ctok-go/tools/reference/Dockerfile /absolute/path/to/scratch
```

If the sandbox cannot write Docker's default buildx activity path, set
`DOCKER_CONFIG` to a writable scratch directory. Do not copy credentials there.

Verify the oracle offline:

```sh
docker run --rm --network none \
  -v /absolute/path/to/ctok-go/tools/reference:/work:ro \
  ctok-go-oracle:434612e python probe.py
```

## Frozen Unicode behavior

Run from the library root. The generated file and output directory must already
exist. Mount only these explicit outputs writable:

```sh
docker run --rm --network none --user "$(id -u):$(id -g)" \
  -v "$PWD/tools/reference:/work:ro" \
  -v /absolute/path/to/verified/ucd:/ucd:ro \
  -v "$PWD/data/v48.bin:/out/data/v48.bin:ro" \
  -v "$PWD/testdata/vocabulary.json:/out/testdata/vocabulary.json:ro" \
  -v "$PWD/unicode_tables.go:/out/unicode_tables.go" \
  -v "$PWD/testdata/unicode:/out/testdata/unicode" \
  ctok-go-oracle:434612e python generate.py
gofmt -w unicode_tables.go
docker run --rm --network none --user "$(id -u):$(id -g)" \
  -v "$PWD/tools/reference:/work:ro" \
  -v "$PWD/unicode_tables.go:/repo/unicode_tables.go:ro" \
  -v "$PWD/testdata/unicode:/repo/testdata/unicode" \
  ctok-go-oracle:434612e python finish_manifest.py
```

The read-only UCD directory contains the pinned NormalizationTest.txt input from
`ucd.json`. The fixture manifest hashes JSON and conformance outputs, generators
and formatted Go output. Repeat generation and require unchanged files before committing.
Normalization table generation and patch verification are recorded separately
with the vendored NFC implementation.

## Runtime vocabulary

Run from the library root after the Unicode fixtures exist. The data and testdata
directories must already exist. This flow hashes the generated vocabulary and
refreshes the generator hash in the shared artifact manifest:

```sh
docker run --rm --network none --user "$(id -u):$(id -g)" \
  -v "$PWD/tools/reference:/work:ro" \
  -v "$PWD/data:/out/data" \
  -v "$PWD/testdata:/out/testdata" \
  ctok-go-oracle:434612e python generate.py vocabulary
```

Repeat and require byte-identical output. The Unicode flow also recomputes the
vocabulary artifact hashes from read-only mounts; missing artifacts fail rather
than silently carrying forward stale recorded hashes. Generate the vocabulary
first when bootstrapping all artifacts, using the shipped Unicode manifest.

## Count parity and marked streams

From the library root, using the same pinned image. Raw uint32 little-endian
lengths followed by UTF-8 bytes preserve CRLF, NUL and all other control bytes.
No JSON/text-mode stdin decoding is involved. The script verifies the complete
corpus SHA-256 before asking upstream's public API for counts. It uses eight
ordered workers. The compressed count pairs and manifest are the only corpus
outputs committed, not the generated inputs.

```sh
set -o pipefail
GOTOOLCHAIN=go1.26.10 go run ./internal/testcorpus/cmd/export | \
  docker run -i --rm --network none --user "$(id -u):$(id -g)" \
    -v "$PWD/tools/reference:/work:ro" \
    -v "$PWD/testdata/parity:/out" \
    ctok-go-oracle:434612e python count_corpus.py
GOTOOLCHAIN=go1.26.10 go run ./internal/testcorpus/cmd/export -limit 5000 | \
  docker run -i --rm --network none --user "$(id -u):$(id -g)" \
    -v "$PWD/tools/reference:/work:ro" \
    -v "$PWD/testdata/parity:/out" \
    ctok-go-oracle:434612e python stream_fixtures.py
```

Repeat both commands and require identical SHA-256 hashes. `gzip` outputs use
mtime=0. `manifest.json` records the corpus version, seed, category counts, total
bytes, oracle pins, script hash and compressed expectation hash. Stream artifacts
and their script are pinned separately in PROVENANCE.md. Corpus generation uses
an explicit integer PRNG and fixed atoms, never toolchain Unicode predicates.

The upstream MIT API edge inputs are separate from the authored corpus. Supply
only the pinned source test file, not any evaluation corpora:

```sh
docker run --rm --network none --user "$(id -u):$(id -g)" \
  -v "$PWD/tools/reference:/work:ro" \
  -v /absolute/path/to/pinned/ctok/tests/test_api.py:/reference_test_api.py:ro \
  -v "$PWD/testdata/parity:/out" \
  ctok-go-oracle:434612e python api_edges.py
```

Repeat and require the api_edges.json hash recorded in PROVENANCE.md. These
inputs are evaluated under V48, whose frame differs from the older families'
expectations in the source test file.
