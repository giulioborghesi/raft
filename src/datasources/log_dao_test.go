package datasources

import (
	"bufio"
	"bytes"
	"testing"

	"github.com/giulioborghesi/raft-implementation/src/service"
)

const (
	testLogFile = "/root/testLogFile.bin"
)

func TestReadWriteLogEntries(t *testing.T) {
	// Create log entries
	expectedEntries := []*service.LogEntry{{EntryTerm: 0, Payload: "giulio"},
		{EntryTerm: 2, Payload: "candice"}, {EntryTerm: 4, Payload: "nell"}}
	expectedOffsets := []int64{50, 101, 149}

	// Create buffer
	buffer := &bytes.Buffer{}
	w := bufio.NewWriter(buffer)

	// Write entries to buffered io.Writer object
	offsets, wBytes, err := writeLogEntries(w, 0, expectedEntries)
	if err != nil {
		t.Fatalf("%v", err)
	}

	// Compare results against expected ones
	expectedBytes := offsets[len(offsets)-1]
	if wBytes != expectedBytes {
		t.Fatalf(bytesMismatchErrFmt, expectedBytes, wBytes)
	}

	for i := range offsets {
		if expectedOffsets[i] != offsets[i] {
			t.Fatalf(offsetMismatchErrFmt, i, expectedOffsets[i], offsets[i])
		}
	}

	// Flush changes to writer object
	err = w.Flush()
	if err != nil {
		t.Fatalf("%v", err)
	}

	// Read entries from buffered io.Reader object
	r := bufio.NewReader(buffer)
	entries, offsets, rBytes := readLogEntries(r)

	// Compare results against expected ones
	if rBytes != wBytes {
		t.Fatalf(bytesMismatchErrFmt, wBytes, rBytes)
	}

	for i := range offsets {
		if expectedOffsets[i] != offsets[i] {
			t.Fatalf(offsetMismatchErrFmt, i, expectedOffsets[i], offsets[i])
		}
	}

	for i := range entries {
		if entries[i].Payload != expectedEntries[i].Payload {
			t.Fatalf(entryMismatchErrFmt, i, expectedEntries[i].Payload,
				entries[i].Payload)
		}

		if entries[i].EntryTerm != expectedEntries[i].EntryTerm {
			t.Fatalf(entryMismatchErrFmt, i, expectedEntries[i].EntryTerm,
				entries[i].EntryTerm)
		}
	}
}
