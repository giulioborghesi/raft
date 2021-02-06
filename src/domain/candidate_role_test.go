package domain

import (
	"fmt"
	"testing"
	"time"

	"github.com/giulioborghesi/raft-implementation/src/utils"
)

const (
	testCandidateActive   = false
	testCandidateVotedFor = 1
	testCandidateRemoteID = 4
	testCandidateServerID = 2
)

func TestCandidateMethodsThatPanicOrAreTrivial(t *testing.T) {
	// Initialize server state and candidate
	c := new(candidateRole)
	s := makeTestServerState(testStartingTerm, testCandidateVotedFor,
		testCandidateServerID, invalidServerID, candidate, testCandidateActive)

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
	s := makeTestServerState(testStartingTerm, testCandidateVotedFor,
		testCandidateServerID, invalidServerID, candidate, testCandidateActive)

	// Store initial term
	initialTerm := int64(testStartingTerm)

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
	s := makeTestServerState(testStartingTerm, testCandidateVotedFor,
		testCandidateServerID, invalidServerID, candidate, testCandidateActive)

	// Prepare test data
	var remoteServerTerm int64 = testStartingTerm - 3
	var remoteServerID int64 = testCandidateServerID + 1

	// Remote server term less than local term, fail
	if success := c.prepareAppend(remoteServerTerm,
		remoteServerID, s); success {
		t.Fatalf("prepareAppend expected to fail")
	}

	// Remote server equal to local term, succeed and do not change voted for
	remoteServerTerm = testStartingTerm
	if success := c.prepareAppend(remoteServerTerm,
		remoteServerID, s); !success {
		t.Fatalf("prepareAppend expected to succeed")
	}

	// Validate changes to server state
	validateServerState(s, follower, remoteServerTerm, testCandidateServerID,
		remoteServerID, t)

	// Remote server term greater than local term, succeed
	remoteServerTerm = testStartingTerm + 1
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
	s := makeTestServerState(testStartingTerm, testCandidateVotedFor,
		testCandidateServerID, invalidServerID, candidate, testCandidateActive)

	// Prepare test data
	var remoteServerTerm int64 = testStartingTerm - 3
	var remoteServerID int64 = testCandidateServerID + 1

	// Server should deny vote to remote server because it is not current
	currentTerm, success := c.requestVote(remoteServerTerm, remoteServerID,
		invalidTerm, invalidServerID, s)
	if success {
		t.Fatalf("requestVote expected to fail")
	}

	if currentTerm != testStartingTerm {
		t.Fatalf(fmt.Sprintf("unexpected server term: "+
			"expected: %d, actual: %d", testStartingTerm, currentTerm))
	}

	// Server should grant its vote to the remote server and switch to follower
	remoteServerTerm = testStartingTerm + 2
	currentTerm, success = c.requestVote(remoteServerTerm, remoteServerID,
		initialTerm, invalidServerID, s)
	if !success {
		t.Fatalf("requestVote expected to succeed")
	}

	// Validate changes to server state
	validateServerState(s, follower, remoteServerTerm, remoteServerID,
		invalidServerID, t)

	// Server should not grant its vote because remote log is not up to date
	s.role = candidate
	remoteServerTerm = remoteServerTerm + 1
	currentTerm, success = c.requestVote(remoteServerTerm, remoteServerID,
		invalidTerm, invalidServerID, s)
	if success {
		t.Fatalf("requestVote expected to fail")
	}

	// Validate changes to server state
	validateServerState(s, follower, remoteServerTerm, invalidServerID,
		invalidServerID, t)
}

func TestCandidateFinalizeElection(t *testing.T) {
	// Initialize server state and candidate
	c := new(candidateRole)
	s := makeTestServerState(testStartingTerm, testCandidateVotedFor,
		testCandidateServerID, invalidServerID, candidate, testCandidateActive)

	// Prepare test data
	var electionTerm int64 = testStartingTerm - 1

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
