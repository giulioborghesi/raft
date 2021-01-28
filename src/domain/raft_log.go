package domain

import (
	"github.com/giulioborghesi/raft-implementation/src/datasources"
	"github.com/giulioborghesi/raft-implementation/src/service"
)

const (
	committed = iota
	appended
	lost
	invalid
)

// logEntryStatus represents the status of a log entry. A log entry is
// committed if has been applied to the replicated state machine; is
// appended if it has been appended to the log, but has not been applied
// yet; and is lost if it cannot be found
type logEntryStatus int64

// abstractRaftLog specifies the interface to be exposed by a log in Raft
type abstractRaftLog interface {
	appendEntry(*service.LogEntry) int64

	appendEntries([]*service.LogEntry, int64, int64) bool

	// entryTerm returns the term of the entry with the specified index
	entryTerm(int64) int64

	// nextIndex returns the index of the last log entry plus one
	nextIndex() int64
}

// raftLog implements the abstractRaftLog interface. Log entries are stored
// in memory for fast access as well as on permanent storage
type raftLog struct {
	entries []*service.LogEntry
}

func (l *raftLog) appendEntry(entry *service.LogEntry) int64 {
	if entry != nil {
		l.entries = append(l.entries, entry)
	}
	return int64(len(l.entries))
}

func (l *raftLog) appendEntries(entries []*service.LogEntry,
	prevEntryTerm int64, prevEntryIndex int64) bool {
	// Must replace all log entries
	if prevEntryIndex == 0 {
		l.entries = entries
		return true
	}

	// Previous log entry does not exist
	if (prevEntryIndex > int64(len(l.entries))) ||
		(l.entries[prevEntryIndex-1].EntryTerm != prevEntryTerm) {
		return false
	}

	// Remove obsolete entries and append new ones
	l.entries = l.entries[:prevEntryIndex]
	l.entries = append(l.entries, entries...)
	return false
}

func (l *raftLog) entryTerm(idx int64) int64 {
	// Log entries numbering starts from 1
	if idx == 0 {
		return 0
	}

	// Log entry exists, return its term
	if idx <= int64(len(l.entries)) {
		return l.entries[idx-1].EntryTerm
	}

	// Log entry does not exist, return ID for invalid term
	return invalidTermID
}

func (l *raftLog) nextIndex() int64 {
	return int64(len(l.entries))
}

// persistentRaftLog implements a log that stores entries in memory and on
// durable storage
type persistentRaftLog struct {
	raftLog
	dao datasources.AbstractLogDao
}

func (l *persistentRaftLog) appendEntry(entry *service.LogEntry) int64 {
	entryIndex := l.raftLog.appendEntry(entry)
	if err := l.dao.AppendEntry(entry); err != nil {
		panic(err)
	}
	return entryIndex
}

func (l *persistentRaftLog) appendEntries(entries []*service.LogEntry,
	prevEntryTerm int64, prevEntryIndex int64) bool {
	if success := l.raftLog.appendEntries(entries, prevEntryTerm,
		prevEntryIndex); !success {
		return success
	}

	if err := l.dao.AppendEntries(entries, prevEntryIndex); err != nil {
		panic(err)
	}
	return true
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

func (l *mockRaftLog) nextIndex() int64 {
	return 1
}
