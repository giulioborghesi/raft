package datasources

import (
	"bufio"
	"os"

	"github.com/giulioborghesi/raft-implementation/src/service"
	"github.com/giulioborghesi/raft-implementation/src/utils"
)

// readLogEntries reads a slice of log entries from a buffered io.Reader object
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

// writeLogEntries writes a slice of log entries to a buffered io.Writer object
func writeLogEntries(w *bufio.Writer, offset int64,
	entries []*service.LogEntry) ([]int64, int64, error) {
	offsets := make([]int64, 0)
	wBytes := int64(0)
	for _, entry := range entries {
		// Write entry term
		tBytes, err := utils.WriteValue(w, entry.EntryTerm)
		if err != nil {
			return nil, 0, err
		}

		// Write entry payload
		pBytes, err := utils.WriteString(w, entry.Payload)
		if err != nil {
			return nil, 0, err
		}

		// Write checksum
		checksum := utils.Int64StringCheckSum(entry.Payload, entry.EntryTerm)
		cBytes, err := utils.WriteValue(w, checksum)
		if err != nil {
			return nil, 0, err
		}

		// Update written bytes and offset
		wBytes += (tBytes + pBytes + cBytes)
		offsets = append(offsets, wBytes+offset)
	}
	return offsets, wBytes, nil
}

// AbstractLogDao defines the interface of a log DAO object
type AbstractLogDao interface {
	// AppendEntries appends multiple log entries to the log and store them on
	// durable storage.
	AppendEntries([]*service.LogEntry) error

	// Entries returns the entries at and following the specified log entry
	// index. It is the responsibility of the calling code to ensure that the
	// specified log entry index is valid
	Entries(int64) []*service.LogEntry

	// EntryTerm returns the entry term of the log entry at the specified log
	// entry index. It is the responsibility of the calling code to ensure that
	// the specified log entry index is valid
	EntryTerm(int64) int64

	// LastEntryIndex returns the index of the last entry in the log
	LastEntryIndex() int64

	// RemoveEntries remove the log entries following the specified index. It
	// is the responsibility of the calling code to ensure that the specified
	// log entry index is valid
	RemoveEntries(int64) error
}

// MakeInMemoryLogDao creates an in-memory version of a log dao from a slice
// of log entries
func MakeInMemoryLogDao(entries []*service.LogEntry) AbstractLogDao {
	return &inMemoryLogDao{e: entries}
}

// inMemoryLogDao implements the AbstractLogDao interface using a slice of
// LogEntry to store the
type inMemoryLogDao struct {
	e []*service.LogEntry
}

func (l *inMemoryLogDao) AppendEntries(entries []*service.LogEntry) error {
	l.e = append(l.e, entries...)
	return nil
}

func (l *inMemoryLogDao) Entries(entryIndex int64) []*service.LogEntry {
	return l.e[entryIndex:]
}

func (l *inMemoryLogDao) EntryTerm(entryIndex int64) int64 {
	return l.e[entryIndex].EntryTerm
}

func (l *inMemoryLogDao) LastEntryIndex() int64 {
	return int64(len(l.e)) - 1
}

func (l *inMemoryLogDao) RemoveEntries(firstValidEntryIndex int64) error {
	l.e = l.e[:firstValidEntryIndex+1]
	return nil
}

// MakePersistentLogDao creates a persistent log dao whose log entries are
// initialized from file
func MakePersistentLogDao(logFilePath string) AbstractLogDao {
	// Open or create the log file if it does not exist
	f, err := os.OpenFile(logFilePath, os.O_RDWR|os.O_CREATE, permissions)
	if err != nil {
		panic(err)
	}

	// Read log entries from file
	r := bufio.NewReader(f)
	entries, offsets, offset := readLogEntries(r)
	if err = f.Truncate(offset); err != nil {
		panic(err)
	}

	// Set IO offset to end of file
	offset, err = f.Seek(offset, 0)
	if err != nil {
		panic(err)
	}

	// Create and return persistent log dao
	dao := MakeInMemoryLogDao(entries)
	return &persistentLogDao{f: f, offset: offset, offsets: offsets,
		AbstractLogDao: dao}
}

// persistentLogDao implements the AbstractLogDao interface using an os.File
// file to store the log entries on durable storage
type persistentLogDao struct {
	f       *os.File
	offset  int64
	offsets []int64
	AbstractLogDao
}

func (l *persistentLogDao) AppendEntries(entries []*service.LogEntry) error {
	// Nothing to do if no entry to write is passed
	if len(entries) == 0 {
		return nil
	}

	// Store entries to in-memory storage
	l.AbstractLogDao.AppendEntries(entries)

	// Create a buffered io.Writer object from log file and write entries to it
	w := bufio.NewWriter(l.f)
	offsets, wBytes, err := writeLogEntries(w, l.offset, entries)
	if err != nil {
		return err
	}

	// Flush changes to durable storage
	if err := w.Flush(); err != nil {
		return err
	}

	// Update offsets and return
	l.offset += wBytes
	l.offsets = append(l.offsets, offsets...)
	return nil
}

func (l *persistentLogDao) EntryTerm(entryIndex int64) int64 {
	return l.AbstractLogDao.EntryTerm(entryIndex)
}

func (l *persistentLogDao) RemoveEntries(firstValidEntryIndex int64) error {
	// Remove entries from in-memory storage
	l.AbstractLogDao.RemoveEntries(firstValidEntryIndex)

	// Compute new file size
	oBytes := int64(0)
	if firstValidEntryIndex >= 0 {
		oBytes = l.offsets[firstValidEntryIndex]
	}

	// Truncate file
	if err := l.f.Truncate(oBytes); err != nil {
		return err
	}

	// Set IO offset to end of file
	oBytes, err := l.f.Seek(oBytes, 0)
	if err != nil {
		panic(err)
	}

	// Update offsets and return
	l.offset = oBytes
	l.offsets = l.offsets[:firstValidEntryIndex+1]
	return nil
}
