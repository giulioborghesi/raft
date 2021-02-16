package datasources

import (
	"bufio"
	"encoding/binary"
	"math"
	"os"

	"github.com/giulioborghesi/raft-implementation/src/service"
	"github.com/giulioborghesi/raft-implementation/src/utils"
)

// readLogEntries reads a slice of log entries from a file
func readLogEntries(r *bufio.Reader) ([]*service.LogEntry, []int64, int64) {
	entries := make([]*service.LogEntry, 0)
	offsets := make([]int64, 0)

	var rBytes int64 = 0
	for {
		// Read entry term
		var entryTerm int64
		tBytes, err := utils.ReadValue(r, &entryTerm)
		if err != nil {
			break
		}

		// Read entry payload
		payload, pBytes, err := utils.ReadString(r)
		if err != nil {
			break
		}

		// Read and validate checksum
		var checksum [28]byte
		cBytes, err := utils.ReadValue(r, &checksum)
		if err != nil {
			break
		}

		actualChecksum := utils.Int64StringCheckSum(payload, entryTerm)
		if checksum != actualChecksum {
			break
		}

		// Update bytes read, log entries and file offsets
		rBytes += (tBytes + pBytes + cBytes)
		entries = append(entries, &service.LogEntry{EntryTerm: entryTerm,
			Payload: payload})
		offsets = append(offsets, rBytes)
	}
	return entries, offsets, rBytes
}

// writeLogEntryToFile writes a slice of log entries to stable storage
/*func writeLogEntriesToFile(f *os.File, entries []*service.LogEntry) error {
	w := bufio.NewWriter(f)
	for _, entry := range entries {
		// Write entry term
		if _, err := writeValue(w, entry.EntryTerm); err != nil {
			return err
		}

		// Write payload
		slen := int64(len([]byte(entry.Payload)))
		if _, err := writeValue(w, slen); err != nil {
			return err
		}

		if _, err := writeValue(w, []byte(entry.Payload)); err != nil {
			return err
		}

		// Write checksum
		checkSum := utils.Int64StringCheckSum(entry.Payload, entry.EntryTerm)
		if _, err := writeValue(w, checkSum); err != nil {
			return err
		}
	}

	// Flush changes to disk and return
	return w.Flush()
}*/

// AbstractLogDao defines the interface of a log DAO object
type AbstractLogDao interface {
	// AppendEntry allows a client to append a log entry to the log and store
	// it on durable storage
	AppendEntry(*service.LogEntry) error

	// AppendEntries appends multiple log entries to the log and store them on
	// durable storage. Entries following the previous log index are dropped
	AppendEntries([]*service.LogEntry, int64) error
}

// logDao implements the AbstractLogDao interface using an os file to store log
// entries on durable storage
type logDao struct {
	f       *os.File
	buffer  []byte
	offset  int64
	offsets []int64
}

func MakeLogFromFile(f *os.File) (AbstractLogDao, []*service.LogEntry, error) {
	return nil, nil, nil
}

func (l *logDao) AppendEntry(entry *service.LogEntry) error {
	return l.AppendEntries([]*service.LogEntry{entry}, math.MaxInt64)
}

func (l *logDao) AppendEntries(entries []*service.LogEntry,
	prevEntryIndex int64) error {
	// Remove old entries if applicable
	if prevEntryIndex < int64(len(l.offsets)) {
		bytes := l.offsets[prevEntryIndex]
		if err := l.f.Truncate(bytes); err != nil {
			return err
		}
		l.offsets = l.offsets[:prevEntryIndex]
	}

	var bytes int64 = 0
	offsets := make([]int64, 0)

	w := bufio.NewWriter(l.f)
	for _, entry := range entries {
		// Write entry term
		nv := binary.PutVarint(l.buffer, entry.EntryTerm)
		l.buffer = l.buffer[:nv]
		nl, err := w.Write(l.buffer)
		if err != nil {
			return err
		}

		// Write payload
		np, err := w.WriteString(entry.Payload)
		if err != nil {
			return err
		}

		// Accumulate bytes written and new offsets
		offsets = append(offsets, l.offset+bytes)
		bytes += int64(nl + np)
	}

	// Flush changes to disk
	if err := w.Flush(); err != nil {
		return err
	}

	// Update offsets and return
	l.offset += bytes
	l.offsets = append(l.offsets, offsets...)
	return nil
}
