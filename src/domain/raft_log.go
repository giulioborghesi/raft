package domain

// abstractRaftLog specifies the interface to be exposed by a log in Raft
type abstractRaftLog interface {
	appendEntries(int64, int64) bool
}

// mockRaftLog implements a mock log to be used for unit testing purposes
type mockRaftLog struct {
	value bool
}

func (l *mockRaftLog) appendEntries(int64, int64) bool {
	return l.value
}
