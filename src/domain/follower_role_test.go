package domain

import (
	"fmt"
	"testing"
	"time"

	"github.com/giulioborghesi/raft-implementation/src/utils"
)

const (
	testFollowerActive   = true
	testFollowerLeaderID = 2
	testFollowerRemoteID = 4
	testFollowerServerID = 3
	testFollowerVotedFor = 1
)

func TestFollowerMethodsThatPanic(t *testing.T) {
	// Initialize server state and follower
	f := new(followerRole)
	s := makeTestServerState(testStartingTerm, testFollowerVotedFor,
		testFollowerServerID, testFollowerLeaderID, follower, testFollowerActive)

	// Test finalizeElection
	finalizeElection := func() {
		f.finalizeElection(0, []requestVoteResult{}, s)
	}
	utils.AssertPanic(t, "finalizeElection", finalizeElection)

	// Test startElection
	startElection := func() {
		f.startElection(s)
	}
	utils.AssertPanic(t, "startElection", startElection)
}

func TestFollowerAppendEntry(t *testing.T) {
	// Create server state and follower instance
	f := new(followerRole)
	s := makeTestServerState(testStartingTerm, testFollowerVotedFor,
		testFollowerServerID, testFollowerLeaderID, follower, testFollowerActive)

	// Call to appendEntry should succeed
	serverTerm := s.currentTerm()
	actualterm, ok := f.appendEntry(nil, serverTerm, s.leaderID, 0, 0, 0, s)
	if actualterm != serverTerm {
		t.Errorf("wrong term following appendEntry: expected %d, actual %d",
			serverTerm, actualterm)
	}

	if !ok {
		t.Errorf(methodExpectedToSucceedErrFmt, "appendEntry")
	}

	// Server term different from local value, appendEntry should panic
	appendEntryInvalidTerm := func() {
		f.appendEntry(nil, serverTerm+1, s.leaderID, 0, 0, 0, s)
	}
	utils.AssertPanic(t, "appendEntry", appendEntryInvalidTerm)

	// Leader ID different from local value, appendEntry should panic
	appendEntryInvalidLeaderID := func() {
		f.appendEntry(nil, serverTerm, s.leaderID+1, 0, 0, 0, s)
	}
	utils.AssertPanic(t, "appendEntry", appendEntryInvalidLeaderID)
}

func TestFollowerMakeCandidate(t *testing.T) {
	// Create server state and follower instance
	f := new(followerRole)
	s := makeTestServerState(testStartingTerm, testFollowerVotedFor,
		testFollowerServerID, testFollowerLeaderID, follower, testFollowerActive)

	// Store initial term for later comparison
	initialTerm := s.currentTerm()

	// Time elapsed since last modification does not exceed timeout, fail
	s.lastModified = time.Now()
	to := time.Minute
	success := f.makeCandidate(to, s)

	if success {
		t.Fatalf(methodExpectedToFailErrFmt, "makeCandidate")
	}

	validateServerState(s, follower, initialTerm, testFollowerVotedFor,
		testFollowerLeaderID, t)

	// Time elapsed since last modification exceeds timeout, succeed
	s.lastModified = time.Now().Add(-to)
	success = f.makeCandidate(to, s)

	if !success {
		t.Fatalf(methodExpectedToSucceedErrFmt, "makeCandidate")
	}

	validateServerState(s, candidate, initialTerm+1, testFollowerServerID,
		invalidServerID, t)
}

func TestFollowerPrepareAppend(t *testing.T) {
	// Create server state and follower instance
	f := new(followerRole)
	s := makeTestServerState(testStartingTerm, testFollowerVotedFor,
		testFollowerServerID, testFollowerLeaderID, follower, testFollowerActive)

	// Call to prepareAppend expected to succeed
	initialTerm := s.currentTerm()
	serverTerm := initialTerm
	ok := f.prepareAppend(serverTerm, testFollowerLeaderID, s)
	if !ok {
		t.Fatalf(methodExpectedToSucceedErrFmt, "prepareAppend")
	}

	validateServerState(s, follower, initialTerm, testFollowerVotedFor,
		testFollowerLeaderID, t)

	// Server term less than local term, prepareAppend expected to fail
	serverTerm = initialTerm - 1
	ok = f.prepareAppend(serverTerm, testFollowerRemoteID, s)
	if ok {
		t.Fatalf(methodExpectedToFailErrFmt, "prepareAppend")
	}

	if s.currentTerm() != testStartingTerm {
		t.Fatalf("prepareAppend should not change the current term")
	}

	// Server term exceeds local term, prepareAppend will succeed and leader
	// ID will be updated
	serverTerm = s.currentTerm() + 1
	ok = f.prepareAppend(serverTerm, testFollowerRemoteID, s)
	if !ok {
		t.Fatalf(methodExpectedToSucceedErrFmt, "prepareAppend")
	}

	if s.currentTerm() != serverTerm {
		t.Fatalf(fmt.Sprintf("invalid term after prepareAppend: "+
			"expected: %d, actual: %d", serverTerm, s.currentTerm()))
	}

	if s.leaderID != testFollowerRemoteID {
		t.Fatalf("leader ID expected to change")
	}
}

func TestFollowerProcessAppendEntryEvent(t *testing.T) {
	// Create server state and follower instance
	f := new(followerRole)
	s := makeTestServerState(testStartingTerm, testFollowerVotedFor,
		testFollowerServerID, testFollowerLeaderID, follower, testFollowerActive)

	// processAppendEntryEvent should fail under all circumstances
	ok := f.processAppendEntryEvent(0, 0, 0, s)
	if ok {
		t.Fatalf(methodExpectedToFailErrFmt, "processAppendEntryEvent")
	}

}

func TestFollowerRequestVote(t *testing.T) {
	// Create server state and follower instance
	f := new(followerRole)
	s := makeTestServerState(testStartingTerm, testFollowerVotedFor,
		testFollowerServerID, testFollowerLeaderID, follower, testFollowerActive)

	// Store initial term for further comparison
	initialTerm := s.currentTerm()

	// Already voted, return false
	serverTerm := initialTerm
	currentTerm, voted := f.requestVote(serverTerm, testFollowerRemoteID,
		invalidTerm, invalidLogEntryIndex, s)

	if voted {
		t.Fatalf(methodExpectedToFailErrFmt, "requestVote")
	}

	validateServerState(s, follower, initialTerm, testFollowerVotedFor,
		testFollowerLeaderID, t)

	// Server did not vote yet, grant vote and return true
	s.updateVotedFor(invalidServerID)
	currentTerm, voted = f.requestVote(serverTerm, testFollowerRemoteID,
		initialTerm, invalidLogEntryIndex, s)

	if !voted {
		t.Fatalf(methodExpectedToSucceedErrFmt, "requestVote")
	}

	validateServerState(s, follower, currentTerm, testFollowerRemoteID,
		testFollowerLeaderID, t)

	// Server term exceeds local term, vote will be granted
	serverTerm += 1
	s.updateVotedFor(invalidServerID)
	currentTerm, voted = f.requestVote(serverTerm, testFollowerRemoteID,
		initialTerm, invalidLogEntryIndex, s)

	if !voted {
		t.Fatalf(methodExpectedToSucceedErrFmt, "requestVote")
	}

	validateServerState(s, follower, currentTerm, testFollowerRemoteID,
		invalidServerID, t)

	// Server term exceeds local term but log is not current
	serverTerm = currentTerm + 1
	s.updateVotedFor(invalidServerID)
	currentTerm, voted = f.requestVote(serverTerm, testFollowerRemoteID,
		invalidTerm, invalidLogEntryIndex, s)

	if voted {
		t.Fatalf(methodExpectedToFailErrFmt, "requestVote")
	}

	validateServerState(s, follower, currentTerm, invalidServerID,
		invalidServerID, t)
}
