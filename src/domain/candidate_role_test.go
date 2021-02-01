package domain

import (
	"fmt"
	"testing"
	"time"

	"github.com/giulioborghesi/raft-implementation/src/datasources"
	"github.com/giulioborghesi/raft-implementation/src/utils"
)

const (
	testCandidateStartingTerm = 15
	testCandidateVotedFor     = 1
	testCandidateRemoteID     = 4
	testCandidateServerID     = 2
)

// makeCandidateServerState creates and initializes an instance of serverState
// for a server in candidate mode
func makeCandidateServerState() *serverState {
	dao := datasources.MakeTestServerStateDao()
	dao.UpdateTerm(testCandidateStartingTerm)
	dao.UpdateVotedFor(testCandidateVotedFor)

	s := new(serverState)
	s.dao = dao
	s.leaderID = invalidServerID
	s.role = candidate
	s.serverID = testCandidateServerID
	return s
}

// validateServerState validates the actual server state against the expected
// server state
func validateServerState(s *serverState, newRole int, newTerm int64,
	newVotedFor int64, newLeaderID int64, t *testing.T) {
	if s.role != newRole {
		t.Fatalf("invalid role: expected: %d, actual: %d", newRole, s.role)
	}

	term, votedFor := s.votedFor()
	if term != newTerm {
		t.Fatalf("invalid term: expected: %d, actual: %d", newTerm, term)
	}

	if newVotedFor != votedFor {
		t.Fatalf("invalid voted for: expected: %d, actual: %d",
			newVotedFor, votedFor)
	}

	if s.leaderID != newLeaderID {
		t.Fatalf("invalid leader ID: expected: %d, actual: %d",
			newLeaderID, s.leaderID)
	}
}

func TestCandidateMethodsThatPanicOrAreTrivial(t *testing.T) {
	// Initialize server state and candidate
	c := new(candidateRole)
	s := new(serverState)

	// Test trivial methods
	c.sendHeartbeat(time.Second, 0, s)
	if ok := c.processAppendEntryEvent(0, 0, 0, s); ok {
		t.Fatalf("processAppendEntry expected to fail")
	}

	// Test finalizeElection
	appendEntry := func() {
		c.appendEntry(nil, 0, 0, 0, 0, 0, s)
	}
	utils.AssertPanic(t, "appendEntry", appendEntry)
}

func TestCandidateMakeCandidate(t *testing.T) {
	// Initialize server state and candidate
	c := new(candidateRole)
	s := makeCandidateServerState()

	// Store initial term
	initialTerm := int64(testCandidateStartingTerm)

	// makeCandidate will increase term and vote for itself
	success := c.makeCandidate(time.Millisecond, s)

	if !success {
		t.Fatalf("makeCandidate expected to succeed")
	}

	// Validate changes to server state
	validateServerState(s, candidate, initialTerm+1, testCandidateServerID,
		invalidServerID, t)
}

func TestCandidatePrepareAppend(t *testing.T) {
	// Initialize server state and candidate
	c := new(candidateRole)
	s := makeCandidateServerState()

	// Prepare test data
	var remoteServerTerm int64 = testCandidateStartingTerm - 3
	var remoteServerID int64 = testCandidateServerID + 1

	// Remote server term less than local term, fail
	if success := c.prepareAppend(remoteServerTerm,
		remoteServerID, s); success {
		t.Fatalf("prepareAppend expected to fail")
	}

	// Remote server equal to local term, succeed and do not change voted for
	remoteServerTerm = testCandidateStartingTerm
	if success := c.prepareAppend(remoteServerTerm,
		remoteServerID, s); !success {
		t.Fatalf("prepareAppend expected to succeed")
	}

	// Validate changes to server state
	validateServerState(s, follower, remoteServerTerm, testCandidateServerID,
		remoteServerID, t)

	// Remote server term greater than local term, succeed
	remoteServerTerm = testCandidateStartingTerm + 1
	if success := c.prepareAppend(remoteServerTerm,
		remoteServerID, s); !success {
		t.Fatalf("prepareAppend expected to succeed")
	}

	// Validate changes to server state
	validateServerState(s, follower, remoteServerTerm, invalidServerID,
		remoteServerID, t)
}

func TestCandidateRequestVote(t *testing.T) {
	// Initialize server state and candidate
	c := new(candidateRole)
	s := makeCandidateServerState()

	// Prepare test data
	var remoteServerTerm int64 = testCandidateStartingTerm - 3
	var remoteServerID int64 = testCandidateServerID + 1

	// Server should deny vote to remote server because it is not current
	currentTerm, success := c.requestVote(remoteServerTerm, remoteServerID,
		invalidTermID, invalidServerID, s)
	if success {
		t.Fatalf("requestVote expected to fail")
	}

	if currentTerm != testCandidateStartingTerm {
		t.Fatalf(fmt.Sprintf("unexpected server term: "+
			"expected: %d, actual: %d", testCandidateStartingTerm, currentTerm))
	}

	// Server should grant its vote to the remote server and switch to follower
	remoteServerTerm = testCandidateStartingTerm + 2
	currentTerm, success = c.requestVote(remoteServerTerm, remoteServerID,
		invalidTermID, invalidServerID, s)
	if !success {
		t.Fatalf("requestVote expected to succeed")
	}

	// Validate changes to server state
	validateServerState(s, follower, remoteServerTerm, remoteServerID,
		invalidServerID, t)
}

func TestCandidateFinalizeElection(t *testing.T) {
	// Initialize server state and candidate
	c := new(candidateRole)
	s := makeCandidateServerState()

	// Prepare test data
	var electionTerm int64 = testCandidateStartingTerm - 1

	// Election results
	results := []requestVoteResult{{success: true,
		serverTerm: electionTerm + 1}}

	// Election starting term differs from current term, do nothing
	if success := c.finalizeElection(electionTerm, results, s); success {
		t.Fatalf("finalizeElection expected to fail")
	}

	// One of election term exceeds election term, do nothing
	if success := c.finalizeElection(electionTerm, results, s); success {
		t.Fatalf("finalizeElection expected to fail")
	}
}
