package domain

import (
	"container/list"
	"fmt"
	"testing"
	"time"

	server_state_dao "github.com/giulioborghesi/raft-implementation/src/datasources"
	"github.com/giulioborghesi/raft-implementation/src/utils"
)

const (
	testLeaderStartingTerm = 15
	testLeaderVotedFor     = 1
	testLeaderLeaderID     = 2
	testLeaderRemoteID     = 4
)

// MakeLeaderServerState creates and initializes an instance of serverState
// for a server in leader mode
func MakeLeaderServerState() *serverState {
	dao := server_state_dao.MakeTestServerStateDao()
	dao.UpdateTerm(testLeaderStartingTerm)
	dao.UpdateVotedFor(testLeaderVotedFor)

	s := new(serverState)
	s.dao = dao
	s.leaderID = testLeaderLeaderID
	s.role = leader
	s.serverID = s.leaderID
	return s
}
func TestLeaderMethodsThatPanic(t *testing.T) {
	// Initialize server state and leader
	l := new(leaderRole)
	s := new(serverState)
	l.appendResults = new(list.List)

	// Test appendEntry
	appendEntry := func() {
		l.appendEntry(0, 0, s)
	}
	utils.AssertPanic(t, "appendEntry", appendEntry)

	// Test finalizeElection
	finalizeElection := func() {
		l.finalizeElection(0, []requestVoteResult{}, s)
	}
	utils.AssertPanic(t, "finalizeElection", finalizeElection)

	// Test startElection
	startElection := func() {
		l.startElection([]string{}, s)
	}
	utils.AssertPanic(t, "startElection", startElection)
}

func TestLeaderMakeCandidate(t *testing.T) {
	// Create server state and leader instance
	l := new(leaderRole)
	s := MakeLeaderServerState()
	l.appendResults = new(list.List)

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
	l.appendResults = new(list.List)

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
	l.appendResults = new(list.List)

	// Store initial term
	initialTerm := s.currentTerm()

	// Initialize remote server term
	remoteServerTerm := int64(initialTerm - 1)
	remoteServerID := int64(s.leaderID + 1)

	// Server term less than local term, requestVote should return false
	currentTerm, success := l.requestVote(remoteServerTerm,
		remoteServerID, s)

	if currentTerm != initialTerm && s.currentTerm() != initialTerm {
		t.Fatalf(fmt.Sprintf("Invalid term returned by requestVote: "+
			"expected: %d, actual: %d", currentTerm, s.currentTerm()))
	}

	if success {
		t.Fatalf("requestVote expected to return false")
	}

	// requestVote should change server role to follower and return true
	remoteServerTerm = testLeaderStartingTerm + 1
	currentTerm, success = l.requestVote(remoteServerTerm,
		remoteServerID, s)

	if currentTerm == initialTerm {
		t.Fatalf("requestVote expected to modify server term")
	}

	if currentTerm != s.currentTerm() {
		t.Fatalf("server term not correctly updated by requestVote")
	}

	if s.role != follower {
		t.Fatalf("server did not transition to follower role")
	}

	if s.leaderID != invalidLeaderID {
		t.Fatalf("invalid leader ID: expected: %d, actual: %d",
			invalidLeaderID, s.leaderID)
	}

	if !success {
		t.Fatalf("requestVote expected to return true")
	}
}
