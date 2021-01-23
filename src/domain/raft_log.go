package domain

// logEntry augments a log entry with the term the entry was added to the log
type logEntry struct {
	entryTerm int64
	payload   string
}

// abstractRaftLog specifies the interface to be exposed by a log in Raft
type abstractRaftLog interface {
	appendEntries([]*logEntry, int64, int64) bool
}

// mockRaftLog implements a mock log to be used for unit testing purposes
type mockRaftLog struct {
	value bool
}

func (l *mockRaftLog) appendEntries(_ []*logEntry, _ int64, _ int64) bool {
	return l.value
}
