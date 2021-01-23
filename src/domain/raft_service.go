package domain

import (
	"sync"
	"time"

	"github.com/giulioborghesi/raft-implementation/src/service"
)

// AbstractRaftService specifies the interface exposed by a type that
// implements a Raft service. AbstractRaftService exposes public methods for
// handling AppendEntry and RequestVote RPC calls, as well as non-public methods
// to access and modify the server state
type AbstractRaftService interface {
	// AppendEntry handles an incoming AppendEntry RPC call
	AppendEntry([]*service.LogEntry, int64, int64, int64, int64,
		int64) (int64, bool)

	// entryInfo returns the current leader term and the term of the log entry
	// with the specified index
	entryInfo(int64) (int64, int64)

	// entries returns a slice of the log entries starting from the specified
	// index, together with the current term and the previous entry term
	entries(int64) ([]*logEntry, int64, int64)

	// lastModified returns the timestamp of last server update
	lastModified() time.Time

	// processAppendEntryEvent processes an event generated while trying to
	// append a log entry to a remote server log
	processAppendEntryEvent(int64, int64, int64)

	// RequestVote handles an incoming RequestVote RPC call
	RequestVote(remoteServerTerm int64, remoteServerID int64) (int64, bool)

	// sendHeartbeat exposes an endpoint to send heartbeats to followers
	sendHeartbeat(time.Duration)

	// StartElection initiates an election if the time elapsed since the last
	// server update exceeds the election timeout
	StartElection(time.Duration)
}

// raftService implements the AbstractRaftService interface
type raftService struct {
	sync.Mutex
	state         *serverState
	remoteServers []string
	roles         map[int]serverRole
}

func (s *raftService) AppendEntry(entries []*service.LogEntry,
	serverTerm int64, serverID int64, prevLogTerm int64, prevLogIndex int64,
	commitIndex int64) (int64, bool) {
	// Lock access to server state
	s.Lock()
	defer s.Unlock()

	// Only a follower can respond to an AppendEntry request
	if ok := s.roles[s.state.role].prepareAppend(serverTerm,
		serverID, s.state); !ok {
		return s.state.currentTerm(), false
	}

	// Unrmarshall log entries
	newEntries := make([]*logEntry, 0, len(entries))
	for _, e := range entries {
		newEntry := logEntry{entryTerm: e.EntryTerm, payload: e.Payload}
		newEntries = append(newEntries, &newEntry)
	}

	// Append entry to log if possible
	return s.roles[s.state.role].appendEntry(newEntries, serverTerm,
		serverID, prevLogTerm, prevLogIndex, commitIndex, s.state)
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

func (s *raftService) lastModified() time.Time {
	// Lock access to server state
	s.Lock()
	defer s.Unlock()

	// Return time of last modification
	return s.state.lastModified
}

func (s *raftService) processAppendEntryEvent(
	serverTerm int64, matchIndex int64, serverID int64) {
	// Lock access to server state
	s.Lock()
	defer s.Unlock()

	s.roles[s.state.role].processAppendEntryEvent(serverTerm, matchIndex,
		serverID, s.state)
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

func (s *raftService) sendHeartbeat(to time.Duration) {
	// Lock access to server state
	s.Lock()
	defer s.Unlock()

	s.roles[s.state.role].sendHeartbeat(to, s.state)
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
