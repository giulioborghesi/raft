package domain

import (
	"math"
	"testing"
	"time"

	server_state_dao "github.com/giulioborghesi/raft-implementation/src/datasources"
)

func createMockRaftService(vr []abstractVoteRequestor,
	er []abstractEntryReplicator) *raftService {
	// Create and initialize server state
	s := new(serverState)
	s.dao = server_state_dao.MakeTestServerStateDao()
	s.role = follower
	s.commitIndex = 0
	s.serverID = 0
	s.leaderID = 1
	s.lastModified = time.Now()

	// Create and initialize Raft service
	service := &raftService{state: s, roles: make(map[int]serverRole)}
	service.roles[follower] = &followerRole{}
	service.roles[candidate] = &candidateRole{voteRequestors: vr}
	service.roles[leader] = &leaderRole{replicators: er,
		matchIndices: []int64{math.MaxInt64, 0}}
	return service
}

func TestLeaderElection(t *testing.T) {
	// Create service
	vr := []abstractVoteRequestor{&mockVoteRequestor{maxCount: 1}}
	s := createMockRaftService(vr, nil)

	// Store initial term
	initialTerm := s.state.currentTerm()

	// Force new election - election expected to fail
	d := time.Since(s.lastModified())
	s.StartElection(d)

	if s.state.role != candidate {
		t.Fatalf("server expected to transition to candidate")
	}

	if s.state.currentTerm() != (initialTerm + 1) {
		t.Fatalf("invalid term: expected: %d, actual: %d",
			initialTerm, s.state.currentTerm())
	}

	// Force yet a new election - this time election will succeed
	d = time.Since(s.lastModified())
	s.StartElection(d)

	if s.state.role != leader {
		t.Fatalf("server expected to transition to leader")
	}

	if s.state.currentTerm() != (initialTerm + 2) {
		t.Fatalf("invalid term: expected: %d, actual: %d",
			initialTerm, s.state.currentTerm())
	}
}
