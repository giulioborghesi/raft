package domain

import (
	"fmt"

	"github.com/giulioborghesi/raft-implementation/src/datasources"
	"github.com/giulioborghesi/raft-implementation/src/service"
)

const (
	committed = iota
	appended
	lost
	invalid
	unknown
)

// isRemoteLogCurrent returns true if the remote log is up-to-date,
// false otherwise
func isRemoteLogCurrent(l abstractRaftLog, lastEntryTerm int64,
	lastEntryIndex int64) bool {
	entryIndex := l.nextIndex() - 1
	entryTerm := l.entryTerm(entryIndex)

	if (lastEntryTerm > entryTerm) || ((lastEntryTerm == entryTerm) &&
		(lastEntryIndex >= entryIndex)) {
		return true
	}
	return false
}

// LogEntryStatus represents the status of a log entry. A log entry is
// committed if has been applied to the replicated state machine; is
// appended if it has been appended to the log, but has not been applied
// yet; and is lost if it cannot be found
type LogEntryStatus int64

// abstractRaftLog specifies the interface to be exposed by a log in Raft
type abstractRaftLog interface {
	appendEntry(*service.LogEntry) int64

	appendEntries([]*service.LogEntry, int64, int64) bool

	// entryTerm returns the term of the entry with the specified index
	entryTerm(int64) int64

	// entries return the log entries at or following the specified index,
	// together with the term of the log entry at the previous index
	entries(int64) ([]*service.LogEntry, int64)

	// nextIndex returns the index of the last log entry plus one
	nextIndex() int64
}

// raftLog implements the abstractRaftLog interface. Log entries are stored
// in memory for fast access as well as on permanent storage
type raftLog struct {
	datasources.AbstractLogDao
}

func (l *raftLog) appendEntry(entry *service.LogEntry) int64 {
	if err := l.AppendEntries([]*service.LogEntry{entry}); err != nil {
		panic(err)
	}
	return l.LastEntryIndex()
}

func (l *raftLog) appendEntries(entries []*service.LogEntry,
	prevEntryTerm int64, prevEntryIndex int64) bool {
	// Nothing to do if previous entry index exceeds log length
	if prevEntryIndex > l.LastEntryIndex() {
		return false
	}

	// Logs do not match: remove entries from prevEntryIndex and return false
	if prevEntryTerm != l.entryTerm(prevEntryIndex) {
		if err := l.RemoveEntries(prevEntryIndex - 1); err != nil {
			panic(err)
		}
		return false
	}

	// Logs match: remove entries following prevEntryIndex
	if err := l.RemoveEntries(prevEntryIndex); err != nil {
		panic(err)
	}

	// Append new entries and return true if no error occurred
	if err := l.AppendEntries(entries); err != nil {
		panic(err)
	}
	return true
}

func (l *raftLog) entryTerm(idx int64) int64 {
	// Nothing to do if entry index is -1
	if idx == invalidLogEntryIndex {
		return initialTerm
	}

	// Log entry exists, return its term
	if idx <= l.LastEntryIndex() {
		return l.EntryTerm(idx)
	}

	// Log entry does not exist, return invalid term ID
	return invalidTerm
}

func (l *raftLog) entries(entryIndex int64) ([]*service.LogEntry, int64) {
	if entryIndex < 0 {
		panic(fmt.Sprintf(invalidIndexErrFmt, entryIndex))
	}

	var entries []*service.LogEntry = nil
	if entryIndex <= l.LastEntryIndex() {
		entries = l.Entries(entryIndex)
	}

	prevEntryIndex := entryIndex - 1
	return entries, l.entryTerm(prevEntryIndex)
}

func (l *raftLog) nextIndex() int64 {
	return l.LastEntryIndex() + 1
}

// mockRaftLog implements a mock log to be used for unit testing purposes
type mockRaftLog struct {
	value bool
}

func (l *mockRaftLog) appendEntry(_ *service.LogEntry) int64 {
	return 1
}

func (l *mockRaftLog) appendEntries(_ []*service.LogEntry, _ int64,
	_ int64) bool {
	return l.value
}

func (l *mockRaftLog) entryTerm(idx int64) int64 {
	return 0
}

func (l *mockRaftLog) entries(idx int64) ([]*service.LogEntry, int64) {
	return nil, 0
}

func (l *mockRaftLog) nextIndex() int64 {
	return 0
}
