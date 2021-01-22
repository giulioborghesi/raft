package domain

import (
	"fmt"
	"testing"
	"time"

	server_state_dao "github.com/giulioborghesi/raft-implementation/src/datasources"
	"github.com/giulioborghesi/raft-implementation/src/utils"
)

const (
	testFollowerStartingTerm = 15
	testFollowerVotedFor     = 1
	testFollowerLeaderID     = 2
	testFollowerRemoteID     = 4
)

// MakeFollowerServerState creates and initializes an instance of serverState
// for a server in follower mode
func MakeFollowerServerState() *serverState {
	dao := server_state_dao.MakeTestServerStateDao()
	dao.UpdateTerm(testFollowerStartingTerm)
	dao.UpdateVotedFor(testFollowerVotedFor)

	s := new(serverState)
	s.dao = dao
	s.log = &mockRaftLog{value: true}
	s.leaderID = testFollowerLeaderID
	return s
}

func TestFollowerMethodsThatPanic(t *testing.T) {
	// Initialize server state and follower
	f := new(followerRole)
	s := new(serverState)

	// Test finalizeElection
	finalizeElection := func() {
		f.finalizeElection(0, []requestVoteResult{}, s)
	}
	utils.AssertPanic(t, "finalizeElection", finalizeElection)

	// Test startElection
	startElection := func() {
		f.startElection([]string{}, s)
	}
	utils.AssertPanic(t, "startElection", startElection)
}

func TestFollowerAppendEntry(t *testing.T) {
	// Create server state and follower instance
	f := new(followerRole)
	s := MakeFollowerServerState()

	// Call to appendEntry should succeed
	serverTerm := s.currentTerm()
	actualterm, ok := f.appendEntry(nil, serverTerm, s.leaderID, 0, 0, s)
	if actualterm != serverTerm {
		t.Errorf("wrong term following appendEntry: expected %d, actual %d",
			serverTerm, actualterm)
	}

	if !ok {
		t.Errorf("appendEntry was not expected to fail")
	}

	// Server term different from local value, appendEntry should panic
	appendEntryInvalidTerm := func() {
		f.appendEntry(nil, serverTerm+1, s.leaderID, 0, 0, s)
	}
	utils.AssertPanic(t, "appendEntry", appendEntryInvalidTerm)

	// Leader ID different from local value, appendEntry should panic
	appendEntryInvalidLeaderID := func() {
		f.appendEntry(nil, serverTerm, s.leaderID+1, 0, 0, s)
	}
	utils.AssertPanic(t, "appendEntry", appendEntryInvalidLeaderID)
}

func TestFollowerMakeCandidate(t *testing.T) {
	// Create server state and follower instance
	f := new(followerRole)
	s := MakeFollowerServerState()

	// Store initial term for later comparison
	initialTerm := s.currentTerm()

	// Initialize time and duration
	s.lastModified = time.Now()
	to := time.Second
	success := f.makeCandidate(to, s)

	// Apart from patological cases during test execution, this should fail
	if success {
		t.Fatalf("makeCandidate was expected to fail")
	}

	if s.currentTerm() != initialTerm {
		t.Fatalf("makeCandidate should not change the current term")
	}

	if s.role != follower {
		t.Fatalf(fmt.Sprintf("invalid server role following makeCandidate: "+
			"expected: %d, actual: %d", follower, s.role))
	}

	// Ensure time elapsed exceeds timeout
	s.lastModified = time.Now().Add(-to)
	success = f.makeCandidate(to, s)

	if !success {
		t.Fatalf("makeCandidate was expected to succeed")
	}

	if s.role != candidate {
		t.Fatalf(fmt.Sprintf("invalid server role following makeCandidate: "+
			"expected: %d, actual: %d", candidate, s.role))
	}
}

func TestFollowerPrepareAppend(t *testing.T) {
	// Create server state and follower instance
	f := new(followerRole)
	s := MakeFollowerServerState()

	// Call to prepareAppend expected to succeed
	serverTerm := s.currentTerm()
	ok := f.prepareAppend(serverTerm, s.leaderID, s)
	if !ok {
		t.Fatalf("prepareAppend expected to succeed")
	}

	if s.currentTerm() != testFollowerStartingTerm {
		t.Fatalf("prepareAppend should not change the current term")
	}

	// Server term less than local term, prepareAppend expected to fail
	serverTerm = s.currentTerm() - 1
	ok = f.prepareAppend(serverTerm, testFollowerRemoteID, s)
	if ok {
		t.Fatalf("prepareAppend expected to fail")
	}

	if s.currentTerm() != testFollowerStartingTerm {
		t.Fatalf("prepareAppend should not change the current term")
	}

	// Server term exceeds local term, prepareAppend will succeed and leader
	// ID will be updated
	serverTerm = s.currentTerm() + 1
	ok = f.prepareAppend(serverTerm, testFollowerRemoteID, s)
	if !ok {
		t.Fatalf("prepareAppend expected to succeed")
	}

	if s.currentTerm() != serverTerm {
		t.Fatalf(fmt.Sprintf("invalid term after prepareAppend: "+
			"expected: %d, actual: %d", serverTerm, s.currentTerm()))
	}

	if s.leaderID != testFollowerRemoteID {
		t.Fatalf("leader ID expected to change")
	}

	// Cause a panic as result of multiple leaders present in the same term
	prepareAppend := func() {
		serverID := int64(testFollowerRemoteID - 1)
		f.prepareAppend(serverTerm, serverID, s)
	}
	utils.AssertPanic(t, "prepareAppend", prepareAppend)
}

func TestFollowerRequestVote(t *testing.T) {
	// Create server state and follower instance
	f := new(followerRole)
	s := MakeFollowerServerState()

	// Store initial term for further comparison
	initialTerm := s.currentTerm()

	// Already voted, return false
	serverTerm := initialTerm
	currentTerm, voted := f.requestVote(serverTerm, testFollowerRemoteID, s)

	if currentTerm != initialTerm {
		t.Fatalf("invalid term returned by requestVote: "+
			"expected: %d, actual: %d", initialTerm, currentTerm)
	}

	if s.currentTerm() != initialTerm {
		t.Fatalf("currentTerm not expected to change")
	}

	if voted {
		t.Fatalf("requestVote not expected to grant a vote")
	}

	if s.leaderID != testFollowerLeaderID {
		t.Fatalf("invalid leader ID following requestVote: "+
			"expected: %d, actual: %d", testFollowerLeaderID, s.leaderID)
	}

	// Server did not vote yet, grant vote and return true
	s.updateVotedFor(invalidServerID)
	currentTerm, voted = f.requestVote(serverTerm, testFollowerRemoteID, s)

	if s.currentTerm() != initialTerm && currentTerm != initialTerm {
		t.Fatalf("invalid term returned by requestVote: "+
			"expected: %d, actual: %d", initialTerm, currentTerm)
	}

	if !voted {
		t.Fatalf("requestVote expected to grant a vote")
	}

	_, votedForID := s.votedFor()
	if votedForID != testFollowerRemoteID {
		t.Fatalf("votedFor invalid following requestVote: "+
			"expected: %d, actual: %d", testFollowerLeaderID, votedForID)
	}

	if s.leaderID != testFollowerLeaderID {
		t.Fatalf("invalid leader ID following requestVote: "+
			"expected: %d, actual: %d", testFollowerLeaderID, s.leaderID)
	}

	// Server term exceeds local term, vote will be granted
	serverTerm += 1
	s.updateVotedFor(invalidServerID)
	currentTerm, voted = f.requestVote(serverTerm, testFollowerRemoteID, s)
	if s.currentTerm() != serverTerm && currentTerm != serverTerm {
		t.Fatalf("invalid term returned by requestVote: "+
			"expected: %d, actual: %d", initialTerm, currentTerm)
	}

	if !voted {
		t.Fatalf("requestVote expected to grant a vote")
	}

	_, votedForID = s.votedFor()
	if votedForID != testFollowerRemoteID {
		t.Fatalf("votedFor invalid following requestVote: "+
			"expected: %d, actual: %d", testFollowerLeaderID, votedForID)
	}

	if s.leaderID != invalidServerID {
		t.Fatalf("invalid leader ID following requestVote: "+
			"expected: %d, actual: %d", invalidServerID, s.leaderID)
	}
}
