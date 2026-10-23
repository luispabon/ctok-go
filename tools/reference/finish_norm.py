"""Apply the two recorded NFC-only table edits to maintained generator output."""
import hashlib
from pathlib import Path
import sys

source = Path(sys.argv[1]).read_text()
assert 'Version = "15.1.0"' in source
assert "//go:build go1.1\n\n" in source
source = source.replace("//go:build go1.1\n\n", "", 1)
chunk_constant = """
	// MaxTransformChunkSize indicates the maximum number of bytes that Transform
	// may need to write atomically for any Form. Making a destination buffer at
	// least this size ensures that Transform can always make progress and that
	// the user does not need to grow the buffer on an ErrShortDst.
	MaxTransformChunkSize = 35 + maxNonStarters*4
"""
assert chunk_constant in source
source = source.replace(chunk_constant, "", 1)
Path(sys.argv[2]).write_text(source)
print(hashlib.sha256(source.encode()).hexdigest())
