package domain

import "sync"

type RaftService struct {
	sync.Mutex
	state *serverState
	roles map[int]serverRole
}

// AppendEntry implements the AppendEntry RPC call. Firstly, RaftService will
// try changing the server state to follower, since only a follower can append
// a log entry to its log. If the transition to follower is successful, the
// service will then forward the request to the underlying server state object,
// and return the result of the forwarding call to the invoking RPC handler
func (s *RaftService) AppendEntry(remoteServerTerm int64,
	remoteServerID int64) (int64, bool) {
	// Lock access to server state
	s.Lock()
	defer s.Unlock()

	// Only a follower can respond to an AppendEntry request
	if ok := s.roles[state.role].makeFollower(state.currentTerm(),
		s.state); !ok {
		return s.state.currentTerm(), false
	}

	// Append entry to log if possible
	return s.roles[state.role].appendEntry(remoteServerTerm,
		remoteServerID, s.state)
}

// RequestVote implements the RequestVote RPC call. Firstly, RaftService will
// try changing the server state to follower, since only a follower can cast
// a vote for a remote server (leaders and candidates will always cast their
// vote for themselves). If the transition to follower is successful, the
// service will then ask for a vote on the remote server, and return the
// result of that call to the invoking RPC handler
func (s *RaftService) RequestVote(remoteServerTerm int64,
	remoteServerID int64) (int64, bool) {
	// Lock access to server state
	s.Lock()
	defer s.Unlock()

	// Only a follower can grant its vote to a remote server
	if ok := s.roles[state.role].makeFollower(state.currentTerm(),
		s.state); !ok {
		return s.state.currentTerm(), false
	}

	// Grant vote if possible
	return s.roles[state.role].requestVote(remoteServerTerm,
		remoteServerID, s.state)
}
