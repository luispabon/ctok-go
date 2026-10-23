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
