package domain

import (
	"sync"
	"time"
)

type raftService struct {
	sync.Mutex
	state         *serverState
	remoteServers []string
	roles         map[int]serverRole
}

// AppendEntry implements the AppendEntry RPC call. Firstly, RaftService will
// try changing the server state to follower, since only a follower can append
// a log entry to its log. If the transition to follower is successful, the
// service will then forward the request to the underlying server state object,
// and return the result of the forwarding call to the invoking RPC handler
func (s *raftService) AppendEntry(remoteServerTerm int64,
	remoteServerID int64) (int64, bool) {
	// Lock access to server state
	s.Lock()
	defer s.Unlock()

	// Only a follower can respond to an AppendEntry request
	if ok := s.roles[s.state.role].prepareAppend(remoteServerTerm,
		remoteServerID, s.state); !ok {
		return s.state.currentTerm(), false
	}

	// Append entry to log if possible
	return s.roles[s.state.role].appendEntry(remoteServerTerm,
		remoteServerID, s.state)
}

func (s *raftService) entryInfo(entryIndex int64) (int64, int64) {
	s.Lock()
	defer s.Unlock()

	term := s.state.currentTerm()
	if entryIndex > 0 {
		return term, invalidTermID
	}
	return s.state.currentTerm(), 0
}

func (s *raftService) makeFollower(leaderTerm int64) {
	s.Lock()
	defer s.Unlock()

	term := s.state.currentTerm()
	if term < leaderTerm {
		s.state.updateTerm(leaderTerm)
		s.state.role = follower
	}
}

func (s *raftService) lastModified() time.Time {
	// Lock access to server state
	s.Lock()
	defer s.Unlock()

	// Return time of last modification
	return s.state.lastModified
}

// RequestVote implements the RequestVote RPC call
func (s *raftService) RequestVote(remoteServerTerm int64,
	remoteServerID int64) (int64, bool) {
	// Lock access to server state
	s.Lock()
	defer s.Unlock()

	// Grant vote if possible
	return s.roles[s.state.role].requestVote(remoteServerTerm,
		remoteServerID, s.state)
}

func (s *raftService) StartElection(to time.Duration) {
	// Lock access to server state
	s.Lock()

	// Only a candidate server can start an election
	if ok := s.roles[s.state.role].makeCandidate(to, s.state); !ok {
		s.Unlock()
		return
	}
	electionTerm := s.state.currentTerm()

	// Request votes asynchronously from remote servers
	asyncResults := s.roles[s.state.role].startElection(s.remoteServers,
		s.state)
	s.Unlock()

	// Wait for RPC calls to complete (or fail due to timeout)
	results := make([]requestVoteResult, 0)
	for _, asyncResult := range asyncResults {
		results = append(results, <-asyncResult)
	}

	// Lock access to server state again
	s.Lock()
	defer s.Unlock()

	// Finalize election
	s.roles[s.state.role].finalizeElection(electionTerm, results, s.state)
}
