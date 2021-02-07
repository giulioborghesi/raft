package datasources

// ServerStateDao defines the interface of a server state DAO object, which is
// used to access and persist to disk the term number and the ID of the server
// that has received the server's vote in the current term
type ServerStateDao interface {
	// CurrentTerm returns the ID of the current term
	CurrentTerm() int64

	// UpdateTerm updates the current term ID and persists its value to disk
	UpdateTerm(int64) error

	// UpdateTermVotedFor updates the current term ID and the ID of the server
	// that has received this server's vote in the current term
	UpdateTermVotedFor(int64, int64) error

	// UpdateVotedFor updates the ID of the server that has received this
	// server's vote in the current term
	UpdateVotedFor(int64) error

	// VotedFor returns the current term ID and the ID of the server that has
	// received this server's vote in the current term
	VotedFor() (int64, int64)
}

// MakeInMemoryServerStateDao creates an instance of inMemoryServerStateDao for
// use in unit tests outside of this package
func MakeInMemoryServerStateDao(currentTerm,
	votedFor int64) *inMemoryServerStateDao {
	return &inMemoryServerStateDao{currentTerm: currentTerm, votedFor: votedFor}
}

// inMemoryServerStateDao implements an in-memory version of the ServerStateDao
// interface. This implementation is useful for testing code
type inMemoryServerStateDao struct {
	currentTerm int64
	votedFor    int64
}

func (s *inMemoryServerStateDao) CurrentTerm() int64 {
	return s.currentTerm
}

func (s *inMemoryServerStateDao) UpdateTerm(serverTerm int64) error {
	s.currentTerm = serverTerm
	s.votedFor = -1
	return nil
}

func (s *inMemoryServerStateDao) UpdateTermVotedFor(serverTerm int64,
	serverID int64) error {
	s.currentTerm = serverTerm
	s.votedFor = serverID
	return nil
}

func (s *inMemoryServerStateDao) UpdateVotedFor(serverID int64) error {
	s.votedFor = serverID
	return nil
}

func (s *inMemoryServerStateDao) VotedFor() (int64, int64) {
	return s.currentTerm, s.votedFor
}
