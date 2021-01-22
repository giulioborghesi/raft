package domain

import "time"

const (
	roleErrCallFmt = "%s should not be called when the server is a %s"
)

// serverRole defines the interface for a server's role in Raft. There exists
// three roles in Raft: follower, candidate and leader. Each role implements a
// distinct behavior for each method specified in the interface
type serverRole interface {
	// appendEntry implements the logic used to determine whether a server
	// should append a log entry sent by the current leader to its log
	appendEntry(int64, int64, int64, int64, *serverState) (int64, bool)

	// finalizeElection processes the results of an election and handles the
	// possible transitions from candidate state to either leader or follower
	// states. Only candidates can finalize an election: calling this method on
	// followers and leaders should result in a panic
	finalizeElection(int64, []requestVoteResult, *serverState)

	// makeCandidate implements the role transition logic to candidate
	makeCandidate(time.Duration, *serverState) bool

	// prepareAppend prepares the server to respond to an AppendEntry RPC: it
	// changes the server role to follower, if possible, and updates the server
	// term and the leader ID
	prepareAppend(int64, int64, *serverState) bool

	// processAppendEntryEvent processes events generated while trying to
	// append entries to the log of a remote server
	processAppendEntryEvent(int64, int64, int64, *serverState)

	// requestVote implements the logic used to determine whether a server
	// should grant its vote to an external server or not
	requestVote(int64, int64, *serverState) (int64, bool)

	// startElection starts an election. Only candidates can start an election
	// and be elected: a panic occurs if leaders and followers call this method
	startElection(servers []string, s *serverState) []chan requestVoteResult
}
