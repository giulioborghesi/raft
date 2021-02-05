package domain

import (
	"fmt"
	"testing"
	"time"

	"github.com/giulioborghesi/raft-implementation/src/datasources"
	"github.com/giulioborghesi/raft-implementation/src/utils"
)

const (
	testLeaderStartingTerm = 15
	testLeaderLeaderID     = 2
	testLeaderRemoteID     = 4
)

// MakeLeaderServerState creates and initializes an instance of serverState
// for a server in leader mode
func MakeLeaderServerState() *serverState {
	dao := datasources.MakeTestServerStateDao()
	dao.UpdateTerm(testLeaderStartingTerm)
	dao.UpdateVotedFor(testLeaderLeaderID)

	s := new(serverState)
	s.dao = dao
	s.leaderID = testLeaderLeaderID
	s.log = &mockRaftLog{}
	s.role = leader
	s.serverID = s.leaderID
	return s
}

func TestEncodeDecodeKey(t *testing.T) {
	var entryTerm int64 = 6
	var entryIndex int64 = 15

	// Encode and decode log entry info
	s := encodeEntry(entryTerm, entryIndex)
	actualEntryTerm, actualEntryIndex := decodeEntry(s)

	if actualEntryTerm != entryTerm {
		t.Fatalf("invalid entry term: expected: %d, actual: %d",
			entryTerm, actualEntryTerm)
	}

	if actualEntryIndex != entryIndex {
		t.Fatalf("invalid entry index: expected: %d, actual: %d",
			entryTerm, actualEntryIndex)
	}
}

func TestLeaderMethodsThatPanic(t *testing.T) {
	// Initialize server state and leader
	l := new(leaderRole)
	s := new(serverState)

	// Test appendEntry
	appendEntry := func() {
		l.appendEntry(nil, 0, 0, 0, 0, 0, s)
	}
	utils.AssertPanic(t, "appendEntry", appendEntry)

	// Test finalizeElection
	finalizeElection := func() {
		l.finalizeElection(0, []requestVoteResult{}, s)
	}
	utils.AssertPanic(t, "finalizeElection", finalizeElection)

	// Test startElection
	startElection := func() {
		l.startElection(s)
	}
	utils.AssertPanic(t, "startElection", startElection)
}

func TestLeaderMakeCandidate(t *testing.T) {
	// Create server state and leader instance
	l := new(leaderRole)
	s := MakeLeaderServerState()

	// makeCandidate always returns false
	success := l.makeCandidate(time.Millisecond, s)
	if success {
		t.Fatalf("makeCandidate did not return false")
	}
}

func TestLeaderPrepareAppend(t *testing.T) {
	// Create server state and leader instance
	l := new(leaderRole)
	s := MakeLeaderServerState()

	// Store initial term
	initialTerm := s.currentTerm()

	// Initialize remote server term
	remoteServerTerm := int64(initialTerm - 1)
	remoteServerID := int64(s.leaderID + 1)

	// Server term less than local term, prepareAppend should return false
	success := l.prepareAppend(remoteServerTerm, remoteServerID, s)

	if s.currentTerm() != initialTerm {
		t.Fatalf("prepareAppend should not modify server term")
	}

	if success {
		t.Fatalf("prepareAppend expected to return false")
	}

	// prepareAppend should change server role to follower and return true
	remoteServerTerm = testLeaderStartingTerm + 1
	success = l.prepareAppend(remoteServerTerm, remoteServerID, s)

	if s.currentTerm() == initialTerm {
		t.Fatalf("prepareAppend expected to modify server term")
	}

	if s.currentTerm() != remoteServerTerm {
		t.Fatalf("invalid term following prepareAppend: "+
			"expected: %d, actual: %d", remoteServerTerm, s.currentTerm())
	}

	if s.role != follower {
		t.Fatalf("server did not transition to follower role")
	}

	if s.leaderID != remoteServerID {
		t.Fatalf("invalid leader ID: expected: %d, actual: %d",
			remoteServerID, s.leaderID)
	}

	if !success {
		t.Fatalf("prepareAppend expected to return true")
	}

	// prepareAppend should panic when two leaders in same term are detected
	prepareAppend := func() {
		remoteServerTerm = s.currentTerm()
		l.prepareAppend(remoteServerTerm, remoteServerID, s)
	}
	utils.AssertPanic(t, "prepareAppend", prepareAppend)
}

func TestLeaderRequestVote(t *testing.T) {
	// Create server state and leader instance
	l := new(leaderRole)
	s := MakeLeaderServerState()

	// Store initial term
	initialTerm := s.currentTerm()

	// Initialize remote server term
	remoteServerTerm := int64(initialTerm - 1)
	remoteServerID := int64(s.leaderID + 1)

	// Server term less than local term, requestVote should return false
	currentTerm, success := l.requestVote(remoteServerTerm,
		remoteServerID, invalidTerm, invalidLogEntryIndex, s)

	if currentTerm != initialTerm && s.currentTerm() != initialTerm {
		t.Fatalf(fmt.Sprintf("Invalid term returned by requestVote: "+
			"expected: %d, actual: %d", currentTerm, s.currentTerm()))
	}

	if success {
		t.Fatalf("requestVote expected to return false")
	}

	// Validate changes to server state
	validateServerState(s, leader, currentTerm, testLeaderLeaderID,
		testLeaderLeaderID, t)

	// requestVote should change server role to follower and return true
	remoteServerTerm = testLeaderStartingTerm + 1
	currentTerm, success = l.requestVote(remoteServerTerm,
		remoteServerID, initialTerm, invalidLogEntryIndex, s)

	if !success {
		t.Fatalf("requestVote expected to return true")
	}

	if currentTerm != remoteServerTerm {
		t.Fatalf("invalid server term: expected: %d, actual: %d",
			remoteServerTerm, currentTerm)
	}

	// Validate changes to server state
	validateServerState(s, follower, currentTerm, remoteServerID,
		invalidServerID, t)
}
