# ctok-go

Pure-Go, local, count-only reconstruction of the pinned ctok V48 family.
Unofficial and unaffiliated with Anthropic. MIT licensed, derived from
[Sander Land's ctok](https://github.com/sanderland/ctok) v1.3.0.

## Install and use

Requires Go 1.26 or later. No runtime dependencies, cgo, subprocesses, credentials
or network access. Until the first release, install the reviewed development
commit from `main`:

```sh
go get github.com/luispabon/ctok-go@main
```

```go
package main

import (
    "fmt"
    "log"

    ctok "github.com/luispabon/ctok-go"
)

func main() {
    counter, err := ctok.New(ctok.V48)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(counter.CountContent("hello")) // 1
    fmt.Println(counter.Count("hello"))        // 7
}
```

Only explicit `V48` (`"4.8"`) is supported. `New` returns
`ErrUnsupportedFamily` for other values. Construct counters with `New`; the zero
value is not usable. Initialized counters share immutable data and support
concurrent calls. Scratch storage belongs to each call.

## Count semantics and limits

- `CountContent(text)` excludes message framing.
- `Count(text)` always equals `CountContent(text) + 6`, modeling one text-only
  user message with the pinned single-message frame. Empty or stripped content
  counts as 0 content tokens and 6 framed tokens.
- Counts are not additive across concatenation. To predict appended text, use
  `CountContent(base + payload) - CountContent(base)`.
- Each contiguous invalid UTF-8 byte run becomes one U+FFFD before NFC and the
  reference's control, space and private-use normalization.
- This is not an encode/decode API and does not expose token boundaries.
- Exact agreement with the pinned reconstruction does not imply exact provider
  billing. Requests, tool schemas, caches, system framing and session overhead
  need separate accounting. No model-name routing or automatic future-family
  selection is provided. Model-specific Opus behavior remains tracked in
  [issue #1](https://github.com/luispabon/ctok-go/issues/1).

## Correctness and reproducibility

Runtime data and Unicode behavior are embedded and frozen. The development oracle
is ctok v1.3.0 at `434612e23c9d9d42d1c5eba23c4761d653e0279d`, Python
3.13.17, unicodedata 15.1.0 and regex 2026.8.31 in a digest-pinned Docker image.
The regex properties intentionally follow that release's Unicode 17 data while
NFC/category/combining behavior follows Unicode 15.1. See [PROVENANCE.md](PROVENANCE.md),
[REFERENCE.json](REFERENCE.json), and [generation commands](tools/reference/README.md).

Ordinary tests require only Go and committed fixtures, not Python or a network.
They regenerate 250,000 deterministic authored UTF-8 inputs (32,438,088 bytes)
and check framed and content counts against compressed oracle expectations.
Another 5,000 fixtures check exact marked streams and trailing-newline counts.
Tests also cover exhaustive Unicode scalar behavior, official NFC conformance,
long combining runs, byte fallback, malformed data and concurrent calls. No
separately licensed evaluation corpora are included.

```sh
go test ./...
go test -race ./...
go vet ./...
CGO_ENABLED=0 go build ./...
go test -run '^$' -bench . -benchmem ./...
```

Verified on Go 1.26.10 and 1.27.3, linux/amd64. CI checks Go 1.26.x and the
current stable Go release. These are supported-toolchain evidence, not a claim
of universal exactness on every compiler or provider model.

## Performance

Development measurements: AMD Ryzen 9 7940HS, linux/amd64, Go 1.26.10. Single
unprofiled runs, not service-level guarantees:

| Operation / input | Time | Allocated |
| --- | ---: | ---: |
| Cold initialization | 10.4 ms | 1.1 MiB retained shared heap |
| 1 MiB ASCII code | 122 ms | 7.4 MiB |
| 1 MiB mixed Unicode | 92 ms | 3.3 MiB |
| 1 MiB short words | 182 ms | 10.0 MiB |
| 1 MiB mixed expanding nonstarters | 213 ms | 22.0 MiB |

Recorded development gates: cold initialization <=150 ms, retained shared heap
<=48 MiB, mixed throughput >=2 MiB/s, 1 MiB count <=1 s, allocation
<=24 bytes per input byte plus 1 MiB. Adversarial allocation tests also measure
4 MiB inputs and require a slope <=24 bytes per input byte. Input is never
truncated. Resource tests are machine-sensitive; race-instrumented timing is not
normal runtime performance. Detailed evidence and storage adaptations are in
[PROVENANCE.md](PROVENANCE.md).

## License and notices

[MIT](LICENSE), preserving Sander Land's copyright and Luis Pabon's contributions.
Vendored normalization retains Go Authors BSD/PATENTS notices; Unicode data has
its separate notice. See [NOTICE](NOTICE) and [UNICODE-LICENSE](UNICODE-LICENSE).
