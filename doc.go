// Package ctok provides local, count-only reconstruction of the pinned ctok V48
// family. It is unofficial and is not affiliated with Anthropic. Counts match
// the recorded upstream reference, not a promise of exact provider billing.
//
// Construct a Counter with New(V48). CountContent excludes framing; Count adds
// the six-token single-user-message frame. Neither models whole requests,
// tool schemas, caches, or provider-specific session overhead. There is no
// encode, decode, token-boundary, or model-name routing API.
//
// Runtime counting requires no network, credentials, subprocess, or cgo.
// Initialized counters are safe for concurrent use. Normalization replaces each
// contiguous invalid UTF-8 byte run with one U+FFFD before counting.
package ctok
