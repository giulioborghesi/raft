package domain

import (
	"fmt"
	"sync"
	"time"

	"github.com/giulioborghesi/raft-implementation/src/clients"
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

	// ApplyCommandAsync applies a command to the state machine asynchronously.
	// It is invoked by an external client of the Raft service
	ApplyCommandAsync(string) (string, int64, error)

	// CommandStatus checks the status of a command that was previously sent to
	// the replicated state machine. It is invoked by an external client of the
	// Raft service
	CommandStatus(string) (LogEntryStatus, int64, error)

	// entryInfo returns the current leader term and the term of the log entry
	// with the specified index
	entryInfo(int64) (int64, int64)

	// entries returns a slice of the log entries starting from the specified
	// index, together with the current term and the previous entry term
	entries(int64) ([]*service.LogEntry, int64, int64)

	// lastModified returns the timestamp of last server update
	lastModified() time.Time

	// processAppendEntryEvent processes an event generated while trying to
	// append a log entry to a remote server log
	processAppendEntryEvent(int64, int64, int64)

	// RequestVote handles an incoming RequestVote RPC call
	RequestVote(int64, int64, int64, int64) (int64, bool)

	// sendHeartbeat exposes an endpoint to send heartbeats to followers
	sendHeartbeat(time.Duration)

	// StartElection initiates an election if the time elapsed since the last
	// server update exceeds the election timeout
	StartElection(time.Duration)
}

// MakeRaftService creates and return an instance of a Raft server
func MakeRaftService(s *serverState,
	serversInfo []*clients.ServerInfo) AbstractRaftService {
	// Initialize Raft service
	service := &raftService{state: s}
	service.roles[follower] = &followerRole{}

	// Create entry replicators and vote requestors
	ers := make([]abstractEntryReplicator, 0, len(serversInfo))
	vrs := make([]abstractVoteRequestor, 0, len(serversInfo))

	for _, serverInfo := range serversInfo {
		conn := clients.Connect(serverInfo.Address)
		client := clients.MakeRaftClient(conn, serverInfo.ServerID)

		er := makeEntryReplicator(serverInfo.ServerID, client, service)
		vr := makeVoteRequestor(client)

		ers = append(ers, er)
		vrs = append(vrs, vr)
	}

	// Finalize candidate and leader initialization
	clusterSize := len(serversInfo) + 1
	service.roles[candidate] = makeCandidateRole(vrs)
	service.roles[leader] = makeLeaderRole(ers, s.leaderID, clusterSize)
	return service
}

// raftService implements the AbstractRaftService interface
type raftService struct {
	sync.Mutex
	state *serverState
	roles map[int]serverRole
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

	// Append entry to log if possible
	return s.roles[s.state.role].appendEntry(entries, serverTerm,
		serverID, prevLogTerm, prevLogIndex, commitIndex, s.state)
}

func (s *raftService) ApplyCommandAsync(payload string) (string,
	int64, error) {
	s.Lock()
	defer s.Unlock()

	// Try appending entry to log and return
	commitIndex := s.state.targetCommitIndex
	return s.roles[s.state.role].appendNewEntry(payload, commitIndex, s.state)
}

func (s *raftService) CommandStatus(key string) (LogEntryStatus,
	int64, error) {
	commitIndex := s.state.targetCommitIndex
	return s.roles[s.state.role].entryStatus(key, commitIndex, s.state)
}

func (s *raftService) entryInfo(entryIndex int64) (int64, int64) {
	s.Lock()
	defer s.Unlock()

	return s.state.currentTerm(), s.state.log.entryTerm(entryIndex)
}

func (s *raftService) entries(entryIndex int64) ([]*service.LogEntry, int64,
	int64) {
	s.Lock()
	defer s.Unlock()

	entries, prevEntryTerm := s.state.log.entries(entryIndex)
	return entries, s.state.currentTerm(), prevEntryTerm
}

func (s *raftService) lastModified() time.Time {
	// Lock access to server state
	s.Lock()
	defer s.Unlock()

	// Return time of last modification
	return s.state.lastModified
}

func (s *raftService) processAppendEntryEvent(
	appendTerm int64, matchIndex int64, serverID int64) {
	// Lock access to server state
	s.Lock()
	defer s.Unlock()

	s.roles[s.state.role].processAppendEntryEvent(appendTerm, matchIndex,
		serverID, s.state)
}

// RequestVote implements the RequestVote RPC call
func (s *raftService) RequestVote(remoteServerTerm int64, remoteServerID int64,
	lastEntryTerm int64, lastEntryIndex int64) (int64, bool) {
	// Lock access to server state
	s.Lock()
	defer s.Unlock()

	// Grant vote if possible
	return s.roles[s.state.role].requestVote(remoteServerTerm,
		remoteServerID, lastEntryTerm, lastEntryIndex, s.state)
}

func (s *raftService) sendHeartbeat(to time.Duration) {
	// Lock access to server state
	s.Lock()
	defer s.Unlock()

	commitIndex := s.state.targetCommitIndex
	s.roles[s.state.role].sendHeartbeat(to, commitIndex, s.state)
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
	asyncResults := s.roles[s.state.role].startElection(s.state)
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
	if success := s.roles[s.state.role].finalizeElection(electionTerm,
		results, s.state); success {
		commitIndex := s.state.targetCommitIndex
		s.roles[s.state.role].sendHeartbeat(time.Duration(0), commitIndex,
			s.state)
	}
}

// UnimplementedRaftService provides a basic implementation of a Raft service
// that can be specialized by its user. It is useful for unit tests, where only
// the behavior of a few methods need to be changed
type UnimplementedRaftService struct{}

func (s *UnimplementedRaftService) AppendEntry([]*service.LogEntry, int64,
	int64, int64, int64, int64) (int64, bool) {
	panic(fmt.Sprintf(notImplementedErrFmt, "AppendEntry"))
}

func (s *UnimplementedRaftService) ApplyCommandAsync(string) (string,
	int64, error) {
	panic(fmt.Sprintf(notImplementedErrFmt, "ApplyCommandAsync"))
}

func (s *UnimplementedRaftService) CommandStatus(string) (LogEntryStatus,
	int64, error) {
	panic(fmt.Sprintf(notImplementedErrFmt, "CommandStatus"))
}

func (s *UnimplementedRaftService) entryInfo(int64) (int64, int64) {
	panic(fmt.Sprintf(notImplementedErrFmt, "entryInfo"))
}

func (s *UnimplementedRaftService) entries(int64) ([]*service.LogEntry,
	int64, int64) {
	panic(fmt.Sprintf(notImplementedErrFmt, "entries"))
}

func (s *UnimplementedRaftService) lastModified() time.Time {
	panic(fmt.Sprintf(notImplementedErrFmt, "lastModified"))
}

func (s *UnimplementedRaftService) processAppendEntryEvent(_, _, _ int64) {
	panic(fmt.Sprintf(notImplementedErrFmt, "processAppendEntryEvent"))
}

func (s *UnimplementedRaftService) RequestVote(int64, int64, int64,
	int64) (int64, bool) {
	panic(fmt.Sprintf(notImplementedErrFmt, "RequestVote"))
}

func (s *UnimplementedRaftService) sendHeartbeat(time.Duration) {
	panic(fmt.Sprintf(notImplementedErrFmt, "sendHeartbeat"))
}

func (s *UnimplementedRaftService) StartElection(time.Duration) {
	panic(fmt.Sprintf(notImplementedErrFmt, "StartElection"))
}
