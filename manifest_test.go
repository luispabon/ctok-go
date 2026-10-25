package ctok

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"maps"
	"os"
	"slices"
	"testing"
)

// unicodeArtifacts is the exact, sorted set of files pinned by
// testdata/unicode/manifest.json.
var unicodeArtifacts = []string{
	"data/v48.bin",
	"testdata/unicode/case.json",
	"testdata/unicode/conformance.json.gz",
	"testdata/unicode/nfc.json",
	"testdata/unicode/normalize.json",
	"testdata/unicode/properties.json",
	"testdata/vocabulary.json",
	"tools/reference/finish_manifest.py",
	"tools/reference/generate.py",
	"unicode_tables.go",
}

type unicodeManifest struct {
	Artifacts        map[string]string `json:"artifacts"`
	GeneratorVersion int               `json:"generator_version"`
	Python           string            `json:"python"`
	Unicodedata      string            `json:"unicodedata"`
	Regex            string            `json:"regex"`
}

type referenceFile struct {
	GeneratorVersion int `json:"generator_version"`
	Oracle           struct {
		Python      string `json:"python"`
		Unicodedata string `json:"unicodedata"`
		Regex       string `json:"regex"`
	} `json:"oracle"`
}

func TestUnicodeManifest(t *testing.T) {
	var manifest unicodeManifest
	readJSONFile(t, "testdata/unicode/manifest.json", &manifest)
	var reference referenceFile
	readJSONFile(t, "REFERENCE.json", &reference)

	if got := slices.Sorted(maps.Keys(manifest.Artifacts)); !slices.Equal(got, unicodeArtifacts) {
		t.Fatalf("manifest artifacts %v, want %v", got, unicodeArtifacts)
	}
	for _, path := range unicodeArtifacts {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		sum := sha256.Sum256(data)
		if got := hex.EncodeToString(sum[:]); got != manifest.Artifacts[path] {
			t.Errorf("%s sha256 %s, manifest %s", path, got, manifest.Artifacts[path])
		}
	}

	if manifest.GeneratorVersion != reference.GeneratorVersion {
		t.Errorf("generator_version %d, REFERENCE.json %d", manifest.GeneratorVersion, reference.GeneratorVersion)
	}
	if manifest.Python != reference.Oracle.Python {
		t.Errorf("python %s, REFERENCE.json %s", manifest.Python, reference.Oracle.Python)
	}
	if manifest.Unicodedata != reference.Oracle.Unicodedata {
		t.Errorf("unicodedata %s, REFERENCE.json %s", manifest.Unicodedata, reference.Oracle.Unicodedata)
	}
	if manifest.Regex != reference.Oracle.Regex {
		t.Errorf("regex %s, REFERENCE.json %s", manifest.Regex, reference.Oracle.Regex)
	}
}

func readJSONFile(t *testing.T, path string, v any) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if err := json.Unmarshal(data, v); err != nil {
		t.Fatalf("decode %s: %v", path, err)
	}
}
