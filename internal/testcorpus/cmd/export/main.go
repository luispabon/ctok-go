// Command export writes the deterministic corpus as uint32 little-endian byte
// lengths followed by raw UTF-8, for the offline Python oracle.
package main

import (
	"bufio"
	"encoding/binary"
	"flag"
	"fmt"
	"os"

	"github.com/luispabon/ctok-go/internal/testcorpus"
)

func main() {
	limit := flag.Int("limit", testcorpus.Count, "number of corpus cases")
	flag.Parse()
	if *limit < 0 || *limit > testcorpus.Count {
		fmt.Fprintln(os.Stderr, "invalid case limit")
		os.Exit(1)
	}
	out := bufio.NewWriter(os.Stdout)
	for i := range *limit {
		text := testcorpus.Case(i)
		if err := binary.Write(out, binary.LittleEndian, uint32(len(text))); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if _, err := out.WriteString(text); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
	if err := out.Flush(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
