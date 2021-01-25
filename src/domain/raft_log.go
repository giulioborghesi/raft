package domain

const (
	committed = iota
	appended
	lost
	invalid
)

// logEntry augments a log entry with the term the entry was added to the log
type logEntry struct {
	entryTerm int64
	payload   string
}

// logEntryStatus represents the status of a log entry. A log entry is
// committed if has been applied to the replicated state machine; is
// appended if it has been appended to the log, but has not been applied
// yet; and is lost if it cannot be found
type logEntryStatus int64

// abstractRaftLog specifies the interface to be exposed by a log in Raft
type abstractRaftLog interface {
	appendEntry(*logEntry) int64

	appendEntries([]*logEntry, int64, int64) bool

	// entryTerm returns the term of the entry with the specified index
	entryTerm(int64) int64

	// nextIndex returns the index of the last log entry plus one
	nextIndex() int64
}

// mockRaftLog implements a mock log to be used for unit testing purposes
type mockRaftLog struct {
	value bool
}

func (l *mockRaftLog) appendEntry(_ *logEntry) int64 {
	return 1
}

func (l *mockRaftLog) appendEntries(_ []*logEntry, _ int64, _ int64) bool {
	return l.value
}

func (l *mockRaftLog) entryTerm(idx int64) int64 {
	return 0
}

func (l *mockRaftLog) nextIndex() int64 {
	return 1
}
