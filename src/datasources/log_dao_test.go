package datasources

import (
	"bufio"
	"bytes"
	"os"
	"testing"

	"github.com/giulioborghesi/raft-implementation/src/service"
	"github.com/giulioborghesi/raft-implementation/src/utils"
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

	utils.ValidateResult(offsets[len(offsets)-1], wBytes, t)
	utils.ValidateResults(expectedOffsets, offsets, t)

	// Flush changes to writer object
	err = w.Flush()
	if err != nil {
		t.Fatalf("%v", err)
	}

	// Read entries from buffered io.Reader object
	r := bufio.NewReader(buffer)
	entries, offsets, rBytes := readLogEntries(r)

	utils.ValidateResult(wBytes, rBytes, t)
	utils.ValidateResults(expectedOffsets, offsets, t)
	utils.ValidateLogs(expectedEntries, entries, t)
}

func TestMakePersistentLogDao(t *testing.T) {
	// Create log entries
	expectedEntries := []*service.LogEntry{{EntryTerm: 0, Payload: "giulio"},
		{EntryTerm: 2, Payload: "candice"}, {EntryTerm: 4, Payload: "nell"}}
	expectedOffsets := []int64{50, 101, 149}

	// Create buffered writer
	f, err := os.OpenFile(testLogFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC,
		permissions)
	if err != nil {
		t.Fatalf("%v", err)
	}
	w := bufio.NewWriter(f)

	// Write entries to buffered io.Writer object
	_, wBytes, err := writeLogEntries(w, 0, expectedEntries)
	if err != nil {
		t.Fatalf("%v", err)
	}

	// Flush changes to file and close it
	if err := w.Flush(); err != nil {
		t.Fatalf("%v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("%v", err)
	}

	// Create persistent log dao
	dao := MakePersistentLogDao(testLogFile)

	utils.ValidateResult(wBytes, dao.(*persistentLogDao).offset, t)
	utils.ValidateResults(expectedOffsets, dao.(*persistentLogDao).offsets, t)
	utils.ValidateLogs(expectedEntries, dao.Entries(0), t)

	// Append a new entry to the log
	newEntry := &service.LogEntry{EntryTerm: 11, Payload: "monster"}
	if err = dao.AppendEntries([]*service.LogEntry{newEntry}); err != nil {
		t.Fatalf("%v", err)
	}

	tBytes := wBytes + 51
	expectedEntries = append(expectedEntries, newEntry)
	expectedOffsets = append(expectedOffsets, tBytes)

	utils.ValidateResult(tBytes, dao.(*persistentLogDao).offset, t)
	utils.ValidateResults(expectedOffsets, dao.(*persistentLogDao).offsets, t)
	utils.ValidateLogs(expectedEntries, dao.Entries(0), t)

	// Remove entries from log
	if err = dao.RemoveEntries(1); err != nil {
		t.Fatalf("%v", err)
	}

	tBytes = expectedOffsets[1]
	expectedEntries = expectedEntries[:2]
	expectedOffsets = expectedOffsets[:2]

	utils.ValidateResult(tBytes, dao.(*persistentLogDao).offset, t)
	utils.ValidateResults(expectedOffsets, dao.(*persistentLogDao).offsets, t)
	utils.ValidateLogs(expectedEntries, dao.Entries(0), t)

	// Close log file
	if err = dao.(*persistentLogDao).f.Close(); err != nil {
		t.Fatalf("%v", err)
	}

	// Re-open log file
	f, err = os.OpenFile(testLogFile, os.O_RDWR|os.O_CREATE, permissions)
	if err != nil {
		t.Fatalf("%v", err)
	}
	r := bufio.NewReader(f)

	// Read entries from log file
	entries, offsets, offset := readLogEntries(r)

	utils.ValidateResult(tBytes, offset, t)
	utils.ValidateResults(expectedOffsets, offsets, t)
	utils.ValidateLogs(expectedEntries, entries, t)

	// Close log file and delete it
	if err = f.Close(); err != nil {
		t.Fatalf("%v", err)
	}

	if err = os.Remove(testLogFile); err != nil {
		t.Fatalf("%v", err)
	}
}
