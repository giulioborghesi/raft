package domain

import (
	"testing"
	"time"

	"github.com/giulioborghesi/raft-implementation/src/utils"
)

const (
	testCandidateActive   = false
	testCandidateVotedFor = 1
	testCandidateRemoteID = 4
	testCandidateServerID = testCandidateVotedFor
)

func TestCandidateMethodsThatPanicOrAreTrivial(t *testing.T) {
	// Initialize server state and candidate
	c := new(candidateRole)
	s := makeTestServerState(testStartingTerm, testCandidateVotedFor,
		testCandidateServerID, invalidServerID, candidate, testCandidateActive)

	// Test trivial methods
	c.sendHeartbeat(time.Second, 0, s)

	if ok := c.processAppendEntryEvent(0, 0, 0, s); ok {
		t.Fatalf(methodExpectedToFailErrFmt, "processAppendEntryEvent")
	}

	// Test methods that panic
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
		t.Fatalf(methodExpectedToSucceedErrFmt, "makeCandidate")
	}

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
		t.Fatalf(methodExpectedToFailErrFmt, "prepareAppend")
	}

	validateServerState(s, candidate, testStartingTerm, testCandidateServerID,
		invalidServerID, t)

	// Remote server equal to local term, succeed and do not change voted for
	remoteServerTerm = testStartingTerm
	if success := c.prepareAppend(remoteServerTerm,
		remoteServerID, s); !success {
		t.Fatalf(methodExpectedToSucceedErrFmt, "prepareAppend")
	}

	validateServerState(s, follower, remoteServerTerm, testCandidateServerID,
		remoteServerID, t)

	// Remote server term greater than local term, succeed
	remoteServerTerm = testStartingTerm + 1
	if success := c.prepareAppend(remoteServerTerm,
		remoteServerID, s); !success {
		t.Fatalf(methodExpectedToSucceedErrFmt, "prepareAppend")
	}

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
	_, success := c.requestVote(remoteServerTerm, remoteServerID,
		invalidTerm, invalidServerID, s)

	if success {
		t.Fatalf(methodExpectedToFailErrFmt, "requestVote")
	}

	validateServerState(s, candidate, testStartingTerm, testCandidateVotedFor,
		invalidServerID, t)

	// Server should grant its vote to the remote server and switch to follower
	remoteServerTerm = testStartingTerm + 2
	_, success = c.requestVote(remoteServerTerm, remoteServerID,
		initialTerm, invalidServerID, s)

	if !success {
		t.Fatalf(methodExpectedToSucceedErrFmt, "requestVote")
	}

	validateServerState(s, follower, remoteServerTerm, remoteServerID,
		invalidServerID, t)

	// Server should not grant its vote because remote log is not up to date
	s.role = candidate
	remoteServerTerm = remoteServerTerm + 1
	_, success = c.requestVote(remoteServerTerm, remoteServerID,
		invalidTerm, invalidServerID, s)

	if success {
		t.Fatalf(methodExpectedToFailErrFmt, "requestVote")
	}

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

	// Election term differs from current term, do nothing
	if success := c.finalizeElection(electionTerm, results, s); success {
		t.Fatalf(methodExpectedToFailErrFmt, "finalizeElection")
	}

	validateServerState(s, candidate, electionTerm+1, testCandidateServerID,
		invalidServerID, t)

	// Election term equal to current and server term, election succeeds
	if success := c.finalizeElection(electionTerm+1, results, s); !success {
		t.Fatalf(methodExpectedToSucceedErrFmt, "finalizeElection")
	}

	validateServerState(s, leader, electionTerm+1, testCandidateServerID,
		testCandidateServerID, t)
}
