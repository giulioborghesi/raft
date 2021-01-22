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
	s.leaderID = invalidServerID
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
		c.appendEntry(0, 0, 0, 0, s)
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
