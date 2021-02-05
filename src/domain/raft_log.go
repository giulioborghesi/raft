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

	// entries return the log entries at or following the specified index,
	// together with the term of the log entry at the previous index
	entries(int64) ([]*service.LogEntry, int64)

	// nextIndex returns the index of the last log entry plus one
	nextIndex() int64
}

// raftLog implements the abstractRaftLog interface. Log entries are stored
// in memory for fast access as well as on permanent storage
type raftLog struct {
	e []*service.LogEntry
}

func (l *raftLog) appendEntry(entry *service.LogEntry) int64 {
	if entry != nil {
		l.e = append(l.e, entry)
	}
	return int64(len(l.e)) - 1
}

func (l *raftLog) appendEntries(entries []*service.LogEntry,
	prevEntryTerm int64, prevEntryIndex int64) bool {
	// Must replace all log entries
	if prevEntryIndex == invalidLogID {
		l.e = entries
		return true
	}

	// Previous log entry was removed
	if (prevEntryIndex >= int64(len(l.e))) ||
		(l.e[prevEntryIndex].EntryTerm != prevEntryTerm) {
		return false
	}

	// Log entry found: remove obsolete entries and append new ones
	l.e = l.e[:prevEntryIndex+1]
	l.e = append(l.e, entries...)
	return true
}

func (l *raftLog) entryTerm(idx int64) int64 {
	if idx == invalidLogID {
		return initialTerm
	}

	// Log entry exists, return its term
	if idx < int64(len(l.e)) {
		return l.e[idx].EntryTerm
	}

	// Log entry does not exist, return ID for invalid term
	return invalidTermID
}

func (l *raftLog) entries(idx int64) ([]*service.LogEntry, int64) {
	if idx < 0 {
		panic(fmt.Sprintf("negative log indices not allowed"))
	}

	var entries []*service.LogEntry = nil
	if idx < int64(len(l.e)) {
		entries = l.e[idx:]
	}

	prevIdx := idx - 1
	return entries, l.entryTerm(prevIdx)
}

func (l *raftLog) nextIndex() int64 {
	return int64(len(l.e))
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

func (l *mockRaftLog) entries(idx int64) ([]*service.LogEntry, int64) {
	return nil, 0
}

func (l *mockRaftLog) nextIndex() int64 {
	return 0
}
