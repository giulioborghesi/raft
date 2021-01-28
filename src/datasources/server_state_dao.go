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
