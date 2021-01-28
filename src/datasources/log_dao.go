package datasources

import (
	"bufio"
	"encoding/binary"
	"math"
	"os"

	"github.com/giulioborghesi/raft-implementation/src/service"
)

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
