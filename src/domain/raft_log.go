package domain

// logEntry augments a log entry with the term the entry was added to the log
type logEntry struct {
	entryTerm int64
	payload   string
}

// abstractRaftLog specifies the interface to be exposed by a log in Raft
type abstractRaftLog interface {
	appendEntries([]*logEntry, int64, int64) bool

	// nextIndex returns the index of the last log entry plus one
	nextIndex() int64
}

// mockRaftLog implements a mock log to be used for unit testing purposes
type mockRaftLog struct {
	value bool
}

func (l *mockRaftLog) appendEntries(_ []*logEntry, _ int64, _ int64) bool {
	return l.value
}

func (l *mockRaftLog) nextIndex() int64 {
	return 1
}
