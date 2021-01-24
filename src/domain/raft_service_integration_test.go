package domain

import (
	"math"
	"testing"
	"time"

	server_state_dao "github.com/giulioborghesi/raft-implementation/src/datasources"
)

const (
	testServiceIntegrationLeaderID      = 1
	testServiceIntegrationOtherServerID = 2
)

func createMockRaftService(vr []abstractVoteRequestor,
	er []abstractEntryReplicator) *raftService {
	// Create a server state instance
	dao := server_state_dao.MakeTestServerStateDao()
	log := &mockRaftLog{value: true}
	s := makeServerState(dao, log, testServiceIntegrationLeaderID)

	// Create and initialize Raft service
	service := &raftService{state: s, roles: make(map[int]serverRole)}
	service.roles[follower] = &followerRole{}
	service.roles[candidate] = &candidateRole{voteRequestors: vr}
	service.roles[leader] = &leaderRole{replicators: er,
		matchIndices: []int64{math.MaxInt64, 0}}
	return service
}

func TestLeaderAppendEntry(t *testing.T) {
	// Create service
	s := createMockRaftService(nil, nil)
	s.state.updateTerm(s.state.currentTerm() + 1)
	s.state.role = leader

	// Store initial term
	initialTerm := s.state.currentTerm()

	// Server receives append entry request from previous leader
	newTerm, success := s.AppendEntry(nil, initialTerm-1,
		testServiceIntegrationOtherServerID, 0, 0, 0)

	if newTerm != initialTerm {
		t.Fatalf("invalid term returned by AppendEntry: "+
			"expected: %d, actual: %d", initialTerm, newTerm)
	}

	if success {
		t.Fatalf("call to AppendEntry was expected to fail")
	}

	// Server receives append entry request from new leader
	newTerm, success = s.AppendEntry(nil, initialTerm+1,
		testServiceIntegrationOtherServerID, 0, 0, 0)

	if newTerm != initialTerm+1 {
		t.Fatalf("invalid term returned by AppendEntry: "+
			"expected: %d, actual: %d", initialTerm+1, newTerm)
	}

	if !success {
		t.Fatalf("call to AppendEntry was expected to succeed")
	}

}

func TestLeaderElection(t *testing.T) {
	// Create service
	vr := []abstractVoteRequestor{&mockVoteRequestor{maxCount: 1},
		&mockVoteRequestor{maxCount: 2}}
	s := createMockRaftService(vr, nil)

	// Store initial term
	initialTerm := s.state.currentTerm()

	// Force new election. Election is expected to fail
	d := time.Since(s.lastModified())
	s.StartElection(d)

	if s.state.role != candidate {
		t.Fatalf("server expected to transition to candidate")
	}

	if s.state.leaderID != invalidServerID {
		t.Fatalf("unexpected leader ID: expected: %d, actual: %d",
			invalidServerID, s.state.leaderID)
	}

	if s.state.currentTerm() != (initialTerm + 1) {
		t.Fatalf("invalid term: expected: %d, actual: %d",
			initialTerm, s.state.currentTerm())
	}

	// Force another election. Election is expected to succeed
	d = time.Since(s.lastModified())
	s.StartElection(d)

	if s.state.role != leader {
		t.Fatalf("server expected to transition to leader")
	}

	if s.state.leaderID != testServiceIntegrationLeaderID {
		t.Fatalf("unexpected leader ID: expected: %d, actual: %d",
			testServiceIntegrationLeaderID, s.state.leaderID)
	}

	if s.state.currentTerm() != (initialTerm + 2) {
		t.Fatalf("invalid term: expected: %d, actual: %d",
			initialTerm, s.state.currentTerm())
	}
}
