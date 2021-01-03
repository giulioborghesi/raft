package domain

// serverRole defines the interface for a server's role in RAFT. There exists
// three roles: follower, candidate and leader. Each role implements a distinct
// behavior for the following endpoints: requestVote, appendEntry and
// startElection
type serverRole interface {
	// appendEntry implements the logic used to determine whether a server
	// should append a log entry sent by the current leader to its log
	appendEntry(int64, int64, *serverState) (int64, bool)

	// requestVote implements the logic used to determine whether a server
	// should grant its vote to an external server for the current term
	requestVote(int64, int64, *serverState) (int64, bool)

	// startElection starts an election. Only candidates can start an election
	// and be elected: calling this method on followers and leaders should
	// automatically return false
	startElection(int64, int64) bool
}
