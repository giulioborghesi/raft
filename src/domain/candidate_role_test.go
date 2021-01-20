package domain

import (
	"fmt"
	"testing"
	"time"

	server_state_dao "github.com/giulioborghesi/raft-implementation/src/datasources"
	"github.com/giulioborghesi/raft-implementation/src/utils"
)

const (
	testCandidateStartingTerm = 15
	testCandidateVotedFor     = 1
	testCandidateRemoteID     = 4
	testCandidateServerID     = 2
)

// MakeCandidateServerState creates and initializes an instance of serverState
// for a server in candidate mode
func MakeCandidateServerState() *serverState {
	dao := server_state_dao.MakeTestServerStateDao()
	dao.UpdateTerm(testCandidateStartingTerm)
	dao.UpdateVotedFor(testCandidateVotedFor)

	s := new(serverState)
	s.dao = dao
	s.leaderID = invalidLeaderID
	s.role = candidate
	s.serverID = testCandidateServerID
	return s
}

func TestCandidateMethodsThatPanic(t *testing.T) {
	// Initialize server state and candidate
	c := new(candidateRole)
	s := new(serverState)

	// Test finalizeElection
	appendEntry := func() {
		c.appendEntry(0, 0, s)
	}
	utils.AssertPanic(t, "appendEntry", appendEntry)
}

func TestCandidateMakeCandidate(t *testing.T) {
	// Initialize server state and candidate
	c := new(candidateRole)
	s := MakeCandidateServerState()

	// Store initial term
	initialTerm := int64(testCandidateStartingTerm)

	// makeCandidate will increase term and vote for itself
	success := c.makeCandidate(time.Millisecond, s)

	if !success {
		t.Fatalf("makeCandidate expected to succeed")
	}

	if s.currentTerm() != initialTerm+1 {
		t.Fatalf(fmt.Sprintf("invalid term: "+
			"expected: %d, actual: %d", initialTerm+1, s.currentTerm()))
	}

	_, votedFor := s.votedFor()
	if votedFor != testCandidateServerID {
		t.Fatalf(fmt.Sprintf("invalid voted for: "+
			"expected: %d, actual: %d", testCandidateServerID, votedFor))
	}

}

/*
func TestFollowerAppendEntry(t *testing.T) {
	// Create server state and follower instance
	f := new(followerRole)
	s := MakeFollowerServerState()

	// Call toppendEntry should succeed
	serverTerm := s.currentTerm()
	actualterm, ok := f.appendEntry(serverTerm, s.leaderID, s)
	if actualterm != serverTerm {
		t.Errorf("wrong term following appendEntry: expected %d, actual %d",
			serverTerm, actualterm)
	}

	if !ok {
		t.Errorf("appendEntry was not expected to fail")
	}

	// Server term different from local value, appendEntry should panic
	appendEntryInvalidTerm := func() {
		f.appendEntry(serverTerm+1, s.leaderID, s)
	}
	utils.AssertPanic(t, "appendEntry", appendEntryInvalidTerm)

	// Leader ID different from local value, appendEntry should panic
	appendEntryInvalidLeaderID := func() {
		f.appendEntry(serverTerm, s.leaderID+1, s)
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
	s.updateVotedFor(invalidLeaderID)
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
	s.updateVotedFor(invalidLeaderID)
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

	if s.leaderID != invalidLeaderID {
		t.Fatalf("invalid leader ID following requestVote: "+
			"expected: %d, actual: %d", invalidLeaderID, s.leaderID)
	}
}
*/
