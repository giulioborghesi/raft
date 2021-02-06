package datasources

// testServerStateDao implements the ServerStateDao interface for use in unit
// tests
type testServerStateDao struct {
	currentTerm int64
	votedFor    int64
}

// MakeTestServerStateDao creates an instance of testServerStateDao for use
// outside of this package
func MakeTestServerStateDao(currentTerm, votedFor int64) *testServerStateDao {
	return &testServerStateDao{currentTerm: currentTerm, votedFor: votedFor}
}

func (s *testServerStateDao) CurrentTerm() int64 {
	return s.currentTerm
}

func (s *testServerStateDao) UpdateTerm(serverTerm int64) error {
	s.currentTerm = serverTerm
	s.votedFor = -1
	return nil
}

func (s *testServerStateDao) UpdateTermVotedFor(serverTerm int64,
	serverID int64) error {
	s.currentTerm = serverTerm
	s.votedFor = serverID
	return nil
}

func (s *testServerStateDao) UpdateVotedFor(serverID int64) error {
	s.votedFor = serverID
	return nil
}

func (s *testServerStateDao) VotedFor() (int64, int64) {
	return s.currentTerm, s.votedFor
}
